//go:build integration

// Seller edit moderation lock (App Store review hardening):
//   - pending auction: seller edits allowed, stays pending
//   - active auction, no bids: a real change sends it back to pending
//     (hidden from public until an admin re-approves); an identical save is a no-op
//   - active auction with bids: edits and image additions are rejected (409)
//   - ended (incl. sold)/canceled/rejected: seller edits and image additions are rejected
//   - the "has bids?" check and the write happen under the auction row lock,
//     so a concurrent bid can never leave a pending auction with bids.
package integrationtest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	apperr "github.com/mazadpay/backend/internal/errors"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/services"
	"github.com/shopspring/decimal"
)

// editInputFrom builds the CreateAuctionInput the app would send for an
// unchanged save of a (the app resends every field on each save).
func editInputFrom(a *models.Auction) services.CreateAuctionInput {
	deref := func(s *string) string {
		if s == nil {
			return ""
		}
		return *s
	}
	return services.CreateAuctionInput{
		CategoryID:      a.CategoryID,
		SubCategoryID:   a.SubCategoryID,
		LocationID:      a.LocationID,
		TitleAr:         a.TitleAr,
		TitleFr:         deref(a.TitleFr),
		TitleEn:         deref(a.TitleEn),
		DescriptionAr:   deref(a.DescriptionAr),
		DescriptionFr:   deref(a.DescriptionFr),
		DescriptionEn:   deref(a.DescriptionEn),
		StartPrice:      a.StartPrice,
		MinIncrement:    a.MinIncrement,
		InsuranceAmount: a.InsuranceAmount,
		EndTime:         a.EndTime,
		PhoneContact:    deref(a.PhoneContact),
		ItemDetails:     a.ItemDetails,
		BuyNowPrice:     a.BuyNowPrice,
		Images:          []string{},
		Quantity:        1,
	}
}

func reloadAuction(t *testing.T, env *testEnv, id uuid.UUID) *models.Auction {
	t.Helper()
	var a models.Auction
	if err := env.db.Get(&a, `SELECT * FROM auctions WHERE id = $1`, id); err != nil {
		t.Fatalf("failed to reload auction: %v", err)
	}
	return &a
}

func setAuctionStatus(t *testing.T, env *testEnv, id uuid.UUID, status string) {
	t.Helper()
	if _, err := env.db.Exec(`UPDATE auctions SET status = $1 WHERE id = $2`, status, id); err != nil {
		t.Fatalf("failed to set auction status: %v", err)
	}
}

func countAuctionImages(t *testing.T, env *testEnv, id uuid.UUID) int {
	t.Helper()
	var n int
	if err := env.db.Get(&n, `SELECT COUNT(*) FROM auction_images WHERE auction_id = $1`, id); err != nil {
		t.Fatalf("failed to count images: %v", err)
	}
	return n
}

func countResubmissionAudits(t *testing.T, env *testEnv, id uuid.UUID) (int, string) {
	t.Helper()
	var rows []struct {
		Details *string `db:"details"`
	}
	if err := env.db.Select(&rows, `SELECT details FROM audit_logs WHERE action = 'auction_resubmitted' AND entity_id = $1`, id); err != nil {
		t.Fatalf("failed to read audit logs: %v", err)
	}
	if len(rows) == 0 || rows[0].Details == nil {
		return len(rows), ""
	}
	return len(rows), *rows[0].Details
}

// placeFundedBid gives bidder enough balance and places one valid bid.
func placeFundedBid(t *testing.T, env *testEnv, auctionID, bidderID uuid.UUID, amount int64) {
	t.Helper()
	creditWallet(t, env, bidderID, decimal.NewFromInt(1000))
	if _, err := env.bidSvc.PlaceBid(context.Background(), auctionID, bidderID, decimal.NewFromInt(amount)); err != nil {
		t.Fatalf("fixture bid failed: %v", err)
	}
}

// editApp exposes the real AuctionHandler.Update / List / GetByID behind test auth.
func editApp(env *testEnv) *fiber.App {
	app := fiber.New()
	app.Put("/auctions/:id/as/:userID", fakeAuthFromParam(), env.auctHandler.Update)
	app.Get("/auctions", fakeAuth(nil), env.auctHandler.List)
	app.Get("/auctions/:id", fakeAuth(nil), env.auctHandler.GetByID)
	app.Get("/auctions/:id/as/:userID", fakeAuthFromParam(), env.auctHandler.GetByID)
	return app
}

func doJSON(t *testing.T, app *fiber.App, method, path string, body interface{}) (int, map[string]interface{}) {
	t.Helper()
	var r io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		r = strings.NewReader(string(b))
	}
	req := httptest.NewRequest(method, path, r)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	defer resp.Body.Close()
	var out map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&out)
	return resp.StatusCode, out
}

func publicListContains(t *testing.T, app *fiber.App, a *models.Auction) bool {
	t.Helper()
	code, out := doJSON(t, app, "GET", "/auctions?limit=100&q="+url.QueryEscape(a.TitleAr), nil)
	if code != 200 {
		t.Fatalf("public list returned %d", code)
	}
	items, _ := out["data"].([]interface{})
	for _, it := range items {
		if m, ok := it.(map[string]interface{}); ok && m["id"] == a.ID.String() {
			return true
		}
	}
	return false
}

func TestSellerEdit_PendingAuction_EditStaysPending(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "Seller Pending Edit")
	a := createTestAuction(t, env, seller.ID, "MR", "MRU")
	setAuctionStatus(t, env, a.ID, "pending")
	a = reloadAuction(t, env, a.ID)

	in := editInputFrom(a)
	in.TitleAr = a.TitleAr + " معدّل"
	if err := env.auctSvc.UpdateAuction(ctx, a.ID, seller.ID, in); err != nil {
		t.Fatalf("pending edit should succeed: %v", err)
	}
	got := reloadAuction(t, env, a.ID)
	if got.Status != "pending" || got.TitleAr != in.TitleAr {
		t.Fatalf("expected pending with new title, got status=%s title=%s", got.Status, got.TitleAr)
	}
	if n, _ := countResubmissionAudits(t, env, a.ID); n != 0 {
		t.Fatalf("a pending edit is not a resubmission, got %d audit rows", n)
	}
}

func TestSellerEdit_ActiveNoBids_MaterialEditGoesPendingAndHidden(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	app := editApp(env)
	seller := createTestUser(t, env, "Seller Active Edit")
	a := createTestAuction(t, env, seller.ID, "MR", "MRU")

	if !publicListContains(t, app, a) {
		t.Fatalf("fixture: active auction should be publicly listed before the edit")
	}

	in := editInputFrom(a)
	in.StartPrice = a.StartPrice.Add(decimal.NewFromInt(50))
	if err := env.auctSvc.UpdateAuction(ctx, a.ID, seller.ID, in); err != nil {
		t.Fatalf("active/no-bids edit should succeed: %v", err)
	}
	got := reloadAuction(t, env, a.ID)
	if got.Status != "pending" {
		t.Fatalf("expected active -> pending after material edit, got %s", got.Status)
	}
	if !got.StartPrice.Equal(in.StartPrice) {
		t.Fatalf("edit not persisted: start_price=%s", got.StartPrice)
	}

	// Public list and public detail hide it; the owner still sees it.
	if publicListContains(t, app, got) {
		t.Fatalf("pending resubmission must not appear in the public list")
	}
	if code, _ := doJSON(t, app, "GET", "/auctions/"+a.ID.String(), nil); code != 404 {
		t.Fatalf("public detail of a pending resubmission should be 404, got %d", code)
	}
	if code, _ := doJSON(t, app, "GET", fmt.Sprintf("/auctions/%s/as/%s", a.ID, seller.ID), nil); code != 200 {
		t.Fatalf("owner should still see their pending listing, got %d", code)
	}

	// Audit: one auction_resubmitted row naming the field, not its value.
	n, details := countResubmissionAudits(t, env, a.ID)
	if n != 1 || !strings.Contains(details, "start_price") {
		t.Fatalf("expected one auction_resubmitted audit naming start_price, got n=%d details=%q", n, details)
	}
	if strings.Contains(details, in.StartPrice.String()) {
		t.Fatalf("audit must not contain edited values, got %q", details)
	}
}

func TestSellerEdit_ActiveNoBids_IdenticalSaveStaysActive(t *testing.T) {
	env := setupEnv(t)
	seller := createTestUser(t, env, "Seller Identical Save")
	a := createTestAuction(t, env, seller.ID, "MR", "MRU")
	before := reloadAuction(t, env, a.ID)

	if err := env.auctSvc.UpdateAuction(context.Background(), a.ID, seller.ID, editInputFrom(before)); err != nil {
		t.Fatalf("identical save should succeed: %v", err)
	}
	got := reloadAuction(t, env, a.ID)
	if got.Status != "active" {
		t.Fatalf("identical save must keep the auction active, got %s", got.Status)
	}
	if got.Version != before.Version {
		t.Fatalf("identical save must not write the row (version %d -> %d)", before.Version, got.Version)
	}
}

func TestSellerEdit_ActiveWithBids_Rejected409NoChange(t *testing.T) {
	env := setupEnv(t)
	app := editApp(env)
	seller := createTestUser(t, env, "Seller With Bids")
	bidder := createTestUser(t, env, "Bidder With Hold")
	a := createTestAuction(t, env, seller.ID, "MR", "MRU")
	placeFundedBid(t, env, a.ID, bidder.ID, 110)

	before := reloadAuction(t, env, a.ID)
	balBefore, frozenBefore := walletBalanceAndFrozen(t, env, bidder.ID)
	holdBefore := holdStatus(t, env, bidder.ID, a.ID)

	// Service level: every material field is locked.
	mutations := map[string]func(*services.CreateAuctionInput){
		"title":         func(in *services.CreateAuctionInput) { in.TitleAr += " x" },
		"description":   func(in *services.CreateAuctionInput) { in.DescriptionAr = "وصف جديد مختلف تماماً" },
		"start_price":   func(in *services.CreateAuctionInput) { in.StartPrice = in.StartPrice.Add(decimal.NewFromInt(1)) },
		"min_increment": func(in *services.CreateAuctionInput) { in.MinIncrement = in.MinIncrement.Add(decimal.NewFromInt(1)) },
		"insurance":     func(in *services.CreateAuctionInput) { in.InsuranceAmount = in.InsuranceAmount.Add(decimal.NewFromInt(1)) },
		"end_time":      func(in *services.CreateAuctionInput) { in.EndTime = in.EndTime.Add(24 * time.Hour) },
		"buy_now": func(in *services.CreateAuctionInput) {
			b := decimal.NewFromInt(999)
			in.BuyNowPrice = &b
		},
		"category": func(in *services.CreateAuctionInput) { in.CategoryID = 2 },
	}
	for name, mutate := range mutations {
		in := editInputFrom(before)
		mutate(&in)
		err := env.auctSvc.UpdateAuction(context.Background(), a.ID, seller.ID, in)
		if !errors.Is(err, apperr.ErrAuctionHasBids) {
			t.Fatalf("%s edit with bids: expected ErrAuctionHasBids, got %v", name, err)
		}
	}

	// HTTP level: 409 with the auction_has_bids code.
	body := map[string]interface{}{
		"title_ar":         before.TitleAr + " http",
		"category_id":      before.CategoryID,
		"start_price":      before.StartPrice.InexactFloat64(),
		"min_increment":    before.MinIncrement.InexactFloat64(),
		"insurance_amount": before.InsuranceAmount.InexactFloat64(),
		"end_time":         before.EndTime.UTC().Format(time.RFC3339),
	}
	code, out := doJSON(t, app, "PUT", fmt.Sprintf("/auctions/%s/as/%s", a.ID, seller.ID), body)
	if code != 409 {
		t.Fatalf("expected HTTP 409, got %d (%v)", code, out)
	}
	if !strings.Contains(fmt.Sprint(out), "auction_has_bids") {
		t.Fatalf("expected auction_has_bids error code, got %v", out)
	}

	after := reloadAuction(t, env, a.ID)
	if after.Status != "active" || after.Version != before.Version || after.TitleAr != before.TitleAr ||
		!after.StartPrice.Equal(before.StartPrice) || !after.InsuranceAmount.Equal(before.InsuranceAmount) ||
		!after.EndTime.Equal(before.EndTime) {
		t.Fatalf("rejected edit must not change the auction")
	}
	balAfter, frozenAfter := walletBalanceAndFrozen(t, env, bidder.ID)
	if !balAfter.Equal(balBefore) || !frozenAfter.Equal(frozenBefore) || holdStatus(t, env, bidder.ID, a.ID) != holdBefore {
		t.Fatalf("rejected edit must not touch balances/holds: balance %s->%s frozen %s->%s",
			balBefore, balAfter, frozenBefore, frozenAfter)
	}
}

func TestSellerEdit_HTTPSuccessReportsPendingStatus(t *testing.T) {
	env := setupEnv(t)
	app := editApp(env)
	seller := createTestUser(t, env, "Seller HTTP Resubmit")
	a := createTestAuction(t, env, seller.ID, "MR", "MRU")
	body := map[string]interface{}{
		"title_ar":         a.TitleAr + " جديد",
		"category_id":      a.CategoryID,
		"start_price":      a.StartPrice.InexactFloat64(),
		"min_increment":    a.MinIncrement.InexactFloat64(),
		"insurance_amount": a.InsuranceAmount.InexactFloat64(),
		"end_time":         a.EndTime.UTC().Format(time.RFC3339),
	}
	code, out := doJSON(t, app, "PUT", fmt.Sprintf("/auctions/%s/as/%s", a.ID, seller.ID), body)
	if code != 200 {
		t.Fatalf("expected 200, got %d (%v)", code, out)
	}
	data, _ := out["data"].(map[string]interface{})
	if data["status"] != "pending" {
		t.Fatalf("expected response status=pending, got %v", out)
	}
}

func TestSellerAddImages_ActiveNoBids_GoesPending(t *testing.T) {
	env := setupEnv(t)
	seller := createTestUser(t, env, "Seller Image Resubmit")
	a := createTestAuction(t, env, seller.ID, "MR", "MRU")

	if err := env.auctSvc.AddImages(context.Background(), a.ID, seller.ID, []string{"https://cdn.example.test/new.jpg"}); err != nil {
		t.Fatalf("AddImages on active/no-bids should succeed: %v", err)
	}
	if got := reloadAuction(t, env, a.ID); got.Status != "pending" {
		t.Fatalf("expected active -> pending after image change, got %s", got.Status)
	}
	if countAuctionImages(t, env, a.ID) != 1 {
		t.Fatalf("image should be saved")
	}
	if n, details := countResubmissionAudits(t, env, a.ID); n != 1 || !strings.Contains(details, "images") {
		t.Fatalf("expected one auction_resubmitted audit naming images, got n=%d %q", n, details)
	}
}

func TestSellerAddImages_PendingAllowed(t *testing.T) {
	env := setupEnv(t)
	seller := createTestUser(t, env, "Seller Image Pending")
	a := createTestAuction(t, env, seller.ID, "MR", "MRU")
	setAuctionStatus(t, env, a.ID, "pending")

	if err := env.auctSvc.AddImages(context.Background(), a.ID, seller.ID, []string{"https://cdn.example.test/p.jpg"}); err != nil {
		t.Fatalf("AddImages on pending should succeed: %v", err)
	}
	if got := reloadAuction(t, env, a.ID); got.Status != "pending" {
		t.Fatalf("pending should stay pending, got %s", got.Status)
	}
}

func TestSellerAddImages_ActiveWithBids_Rejected(t *testing.T) {
	env := setupEnv(t)
	seller := createTestUser(t, env, "Seller Image Bids")
	bidder := createTestUser(t, env, "Bidder Image Bids")
	a := createTestAuction(t, env, seller.ID, "MR", "MRU")
	placeFundedBid(t, env, a.ID, bidder.ID, 110)

	err := env.auctSvc.AddImages(context.Background(), a.ID, seller.ID, []string{"https://cdn.example.test/x.jpg"})
	if !errors.Is(err, apperr.ErrAuctionHasBids) {
		t.Fatalf("expected ErrAuctionHasBids, got %v", err)
	}
	if countAuctionImages(t, env, a.ID) != 0 {
		t.Fatalf("rejected image must not be saved")
	}
	if got := reloadAuction(t, env, a.ID); got.Status != "active" {
		t.Fatalf("status must stay active, got %s", got.Status)
	}
}

func TestSellerEditAndImages_ClosedStatusesRejected(t *testing.T) {
	env := setupEnv(t)
	// "sold" is not a status: a sold auction is "ended" with a winner.
	for _, status := range []string{"ended", "canceled", "rejected"} {
		seller := createTestUser(t, env, "Seller Closed "+status)
		a := createTestAuction(t, env, seller.ID, "MR", "MRU")
		setAuctionStatus(t, env, a.ID, status)
		a = reloadAuction(t, env, a.ID)

		in := editInputFrom(a)
		in.TitleAr += " x"
		if err := env.auctSvc.UpdateAuction(context.Background(), a.ID, seller.ID, in); !errors.Is(err, apperr.ErrAuctionNotEditable) {
			t.Fatalf("%s: edit expected ErrAuctionNotEditable, got %v", status, err)
		}
		if err := env.auctSvc.AddImages(context.Background(), a.ID, seller.ID, []string{"https://cdn.example.test/c.jpg"}); !errors.Is(err, apperr.ErrAuctionNotEditable) {
			t.Fatalf("%s: AddImages expected ErrAuctionNotEditable, got %v", status, err)
		}
		if got := reloadAuction(t, env, a.ID); got.Status != status || got.TitleAr != a.TitleAr || countAuctionImages(t, env, a.ID) != 0 {
			t.Fatalf("%s: rejected edit changed the auction", status)
		}
	}
}

func TestSellerEdit_NonOwnerRejected(t *testing.T) {
	env := setupEnv(t)
	seller := createTestUser(t, env, "Seller Owner")
	other := createTestUser(t, env, "Not The Owner")
	a := createTestAuction(t, env, seller.ID, "MR", "MRU")
	in := editInputFrom(a)
	in.TitleAr += " hijack"
	if err := env.auctSvc.UpdateAuction(context.Background(), a.ID, other.ID, in); !errors.Is(err, apperr.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}

func TestSellerEdit_ResubmittedAuctionCanBeReapproved(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	app := editApp(env)
	adminSvc := newTestAdminService(t, env)
	seller := createTestUser(t, env, "Seller Reapprove")
	admin := createTestAdmin(t, env, "Admin Reapprove")
	a := createTestAuction(t, env, seller.ID, "MR", "MRU")

	in := editInputFrom(a)
	in.TitleAr += " v2"
	if err := env.auctSvc.UpdateAuction(ctx, a.ID, seller.ID, in); err != nil {
		t.Fatalf("edit failed: %v", err)
	}
	if err := adminSvc.ValidateAuction(ctx, a.ID, true, "", admin.ID); err != nil {
		t.Fatalf("admin re-approval failed: %v", err)
	}
	got := reloadAuction(t, env, a.ID)
	if got.Status != "active" || got.TitleAr != in.TitleAr {
		t.Fatalf("expected active with edited title after re-approval, got %s %q", got.Status, got.TitleAr)
	}
	if !publicListContains(t, app, got) {
		t.Fatalf("re-approved auction should be public again")
	}
}

// A bid and a seller edit racing on the same active auction must end in one
// of two valid states: (bid accepted, edit rejected, still active) or
// (edit accepted -> pending, bid rejected). Never pending with a bid.
func TestSellerEdit_RaceWithBid_NeverPendingWithBids(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	bidWins, editWins := 0, 0
	for i := 0; i < 15; i++ {
		seller := createTestUser(t, env, fmt.Sprintf("Race Seller %d", i))
		bidder := createTestUser(t, env, fmt.Sprintf("Race Bidder %d", i))
		a := createTestAuction(t, env, seller.ID, "MR", "MRU")
		creditWallet(t, env, bidder.ID, decimal.NewFromInt(1000))
		in := editInputFrom(a)
		in.TitleAr += " race"

		var wg sync.WaitGroup
		var bidErr, editErr error
		start := make(chan struct{})
		wg.Add(2)
		go func() {
			defer wg.Done()
			<-start
			_, bidErr = env.bidSvc.PlaceBid(ctx, a.ID, bidder.ID, decimal.NewFromInt(110))
		}()
		go func() {
			defer wg.Done()
			<-start
			editErr = env.auctSvc.UpdateAuction(ctx, a.ID, seller.ID, in)
		}()
		close(start)
		wg.Wait()

		got := reloadAuction(t, env, a.ID)
		var bids int
		if err := env.db.Get(&bids, `SELECT COUNT(*) FROM bids WHERE auction_id = $1`, a.ID); err != nil {
			t.Fatalf("count bids: %v", err)
		}
		if got.Status == "pending" && bids > 0 {
			t.Fatalf("iteration %d: pending auction with %d bids (bidErr=%v editErr=%v)", i, bids, bidErr, editErr)
		}
		switch {
		case bidErr == nil && editErr == nil:
			t.Fatalf("iteration %d: both the bid and the edit succeeded", i)
		case bidErr == nil:
			bidWins++
			if !errors.Is(editErr, apperr.ErrAuctionHasBids) || got.Status != "active" || bids != 1 {
				t.Fatalf("iteration %d: bid won but state is inconsistent: status=%s bids=%d editErr=%v", i, got.Status, bids, editErr)
			}
		case editErr == nil:
			editWins++
			if got.Status != "pending" || bids != 0 {
				t.Fatalf("iteration %d: edit won but state is inconsistent: status=%s bids=%d bidErr=%v", i, got.Status, bids, bidErr)
			}
			if _, frozen := walletBalanceAndFrozen(t, env, bidder.ID); !frozen.IsZero() {
				t.Fatalf("iteration %d: rejected bid left %s frozen", i, frozen)
			}
		}
	}
	t.Logf("race outcomes: bid won %d, edit won %d", bidWins, editWins)
}

// The mobile edit screen sends only title/description (all three languages)
// and start_price. Every other field must keep its stored value.
func TestSellerEdit_AppPartialPayload_KeepsOtherFields(t *testing.T) {
	env := setupEnv(t)
	app := editApp(env)
	seller := createTestUser(t, env, "Seller App Payload")
	a := createTestAuction(t, env, seller.ID, "MR", "MRU")
	before := reloadAuction(t, env, a.ID)

	body := map[string]interface{}{
		"title_ar": "عنوان معدّل من التطبيق", "title_fr": "عنوان معدّل من التطبيق", "title_en": "عنوان معدّل من التطبيق",
		"description_ar": "وصف معدّل من التطبيق", "description_fr": "وصف معدّل من التطبيق", "description_en": "وصف معدّل من التطبيق",
		"start_price": 150.0,
	}
	code, out := doJSON(t, app, "PUT", fmt.Sprintf("/auctions/%s/as/%s", a.ID, seller.ID), body)
	if code != 200 {
		t.Fatalf("app-shaped edit should succeed, got %d (%v)", code, out)
	}
	got := reloadAuction(t, env, a.ID)
	if got.Status != "pending" || got.TitleAr != "عنوان معدّل من التطبيق" || !got.StartPrice.Equal(decimal.NewFromInt(150)) {
		t.Fatalf("edit not applied: status=%s title=%q price=%s", got.Status, got.TitleAr, got.StartPrice)
	}
	if got.CategoryID != before.CategoryID || !got.EndTime.Equal(before.EndTime) ||
		!got.InsuranceAmount.Equal(before.InsuranceAmount) || !got.MinIncrement.Equal(before.MinIncrement) ||
		!intPtrEq(got.LocationID, before.LocationID) {
		t.Fatalf("fields absent from the payload must be kept: category %d->%d end %s->%s insurance %s->%s increment %s->%s",
			before.CategoryID, got.CategoryID, before.EndTime, got.EndTime,
			before.InsuranceAmount, got.InsuranceAmount, before.MinIncrement, got.MinIncrement)
	}
}

func intPtrEq(a, b *int) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}
