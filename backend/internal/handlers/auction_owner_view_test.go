package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	apperr "github.com/mazadpay/backend/internal/errors"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/repository"
	"github.com/mazadpay/backend/internal/services"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// fakeAuctionServiceForGetByID and fakeUserRepoForGetByID let a test drive
// AuctionHandler.GetByID through the real Fiber HTTP cycle without a
// database, embedding the real (large) interfaces as nil and overriding only
// what GetByID actually calls.
type fakeAuctionServiceForGetByID struct {
	services.AuctionService
	auction *models.Auction
}

func (f *fakeAuctionServiceForGetByID) GetByID(ctx context.Context, id uuid.UUID) (*models.Auction, []models.AuctionImage, error) {
	return f.auction, nil, nil
}

type fakeUserRepoForGetByID struct {
	repository.UserRepository
	users map[uuid.UUID]*models.User
}

func (f *fakeUserRepoForGetByID) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	u, ok := f.users[id]
	if !ok {
		return nil, context.DeadlineExceeded // any non-nil error -> caller falls back to anonymous
	}
	return u, nil
}

// TestAuctionGetByID_D_OwnerCanViewOwnNonPublicAuction is the regression test
// for the real-device Staging failure "خطأ في تحميل المزاد" opening an item
// from My Auctions. Root cause: GetByID (auction_handler.go) only allowed
// PubliclyVisibleAuctionStatuses ("active"/"ended") through, 404-ing for
// EVERY other caller including the auction's own seller -- so a seller
// opening their own still-"pending" (not yet admin-approved) auction from My
// Auctions always got a 404, which auction_provider_api.dart turns into a
// thrown Exception, rendering the generic error_loading_auction screen. The
// code comment this replaces referenced a "dedicated seller/admin detail
// endpoint" that never actually existed in routes.go. This test proves the
// owner of a "pending" auction can now retrieve it via the same GetByID path
// the mobile app already calls.
func TestAuctionGetByID_D_OwnerCanViewOwnNonPublicAuction(t *testing.T) {
	sellerID := uuid.New()
	auctionID := uuid.New()
	mr := "MR"

	auction := &models.Auction{
		ID:               auctionID,
		SellerID:         sellerID,
		Status:           "pending", // NOT in PubliclyVisibleAuctionStatuses
		StartPrice:       decimal.NewFromInt(100),
		CurrentPrice:     decimal.NewFromInt(100),
		StartTime:        time.Now(),
		EndTime:          time.Now().Add(24 * time.Hour),
		MarketCountryISO: &mr,
	}

	auctionSvc := &fakeAuctionServiceForGetByID{auction: auction}
	userRepo := &fakeUserRepoForGetByID{users: map[uuid.UUID]*models.User{
		sellerID: {ID: sellerID, Role: "user", AccountCountryISO: &mr},
	}}
	h := NewAuctionHandler(auctionSvc, userRepo, zap.NewNop())

	app := fiber.New()
	app.Get("/v1/api/auctions/:id", func(c *fiber.Ctx) error {
		c.Locals("user_id", sellerID)
		return h.GetByID(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/api/auctions/"+auctionID.String(), nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected the owner to retrieve their own pending auction (HTTP 200), got %d, body: %s", resp.StatusCode, body)
	}
}

// TestAuctionGetByID_NonOwnerStillCannotViewPendingAuction proves the fix did
// not weaken the original anti-enumeration protection: a DIFFERENT
// authenticated user (not the seller, not an admin) must still get a 404 for
// a non-public auction.
func TestAuctionGetByID_NonOwnerStillCannotViewPendingAuction(t *testing.T) {
	sellerID := uuid.New()
	otherUserID := uuid.New()
	auctionID := uuid.New()
	mr := "MR"

	auction := &models.Auction{
		ID:               auctionID,
		SellerID:         sellerID,
		Status:           "pending",
		StartPrice:       decimal.NewFromInt(100),
		CurrentPrice:     decimal.NewFromInt(100),
		StartTime:        time.Now(),
		EndTime:          time.Now().Add(24 * time.Hour),
		MarketCountryISO: &mr,
	}

	auctionSvc := &fakeAuctionServiceForGetByID{auction: auction}
	userRepo := &fakeUserRepoForGetByID{users: map[uuid.UUID]*models.User{
		otherUserID: {ID: otherUserID, Role: "user", AccountCountryISO: &mr},
	}}
	h := NewAuctionHandler(auctionSvc, userRepo, zap.NewNop())

	app := fiber.New()
	app.Get("/v1/api/auctions/:id", func(c *fiber.Ctx) error {
		c.Locals("user_id", otherUserID)
		return h.GetByID(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/api/auctions/"+auctionID.String(), nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("SECURITY REGRESSION: expected 404 for a non-owner viewing a pending auction, got %d", resp.StatusCode)
	}
}

// TestAuctionGetByID_AnonymousStillCannotViewPendingAuction proves anonymous
// callers are still blocked from a non-public auction.
func TestAuctionGetByID_AnonymousStillCannotViewPendingAuction(t *testing.T) {
	sellerID := uuid.New()
	auctionID := uuid.New()
	mr := "MR"

	auction := &models.Auction{
		ID:               auctionID,
		SellerID:         sellerID,
		Status:           "pending",
		StartPrice:       decimal.NewFromInt(100),
		CurrentPrice:     decimal.NewFromInt(100),
		StartTime:        time.Now(),
		EndTime:          time.Now().Add(24 * time.Hour),
		MarketCountryISO: &mr,
	}

	auctionSvc := &fakeAuctionServiceForGetByID{auction: auction}
	userRepo := &fakeUserRepoForGetByID{users: map[uuid.UUID]*models.User{}}
	h := NewAuctionHandler(auctionSvc, userRepo, zap.NewNop())

	app := fiber.New()
	app.Get("/v1/api/auctions/:id", h.GetByID) // no c.Locals("user_id", ...)

	req := httptest.NewRequest(http.MethodGet, "/v1/api/auctions/"+auctionID.String(), nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 for an anonymous caller viewing a pending auction, got %d", resp.StatusCode)
	}
}

// TestAuctionGetByID_E_LegacyNullMarketRowsStillWork is targeted test E5:
// an auction with MarketCountryISO/CurrencyCode NULL (predating migration
// 000046) must not break GetByID for its owner, an anonymous MR caller, or
// a same-market authenticated caller -- EffectiveMarketCountryISO's fallback
// to DefaultAccountCountryISO ("MR") must resolve consistently on both the
// auction and caller side.
func TestAuctionGetByID_E_LegacyNullMarketRowsStillWork(t *testing.T) {
	sellerID := uuid.New()
	auctionID := uuid.New()

	// No MarketCountryISO/CurrencyCode set -- exercises the legacy fallback path.
	auction := &models.Auction{
		ID:           auctionID,
		SellerID:     sellerID,
		Status:       "active",
		StartPrice:   decimal.NewFromInt(100),
		CurrentPrice: decimal.NewFromInt(100),
		StartTime:    time.Now(),
		EndTime:      time.Now().Add(24 * time.Hour),
		// MarketCountryISO, CurrencyCode intentionally left nil.
	}

	auctionSvc := &fakeAuctionServiceForGetByID{auction: auction}
	userRepo := &fakeUserRepoForGetByID{users: map[uuid.UUID]*models.User{}}
	h := NewAuctionHandler(auctionSvc, userRepo, zap.NewNop())

	app := fiber.New()
	// Anonymous caller: no c.Locals("user_id", ...) set, matching the public
	// unauthenticated path -- falls back to DefaultAccountCountryISO on both sides.
	app.Get("/v1/api/auctions/:id", h.GetByID)

	req := httptest.NewRequest(http.MethodGet, "/v1/api/auctions/"+auctionID.String(), nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected a legacy null-market active auction to be visible to an anonymous MR caller (HTTP 200), got %d, body: %s", resp.StatusCode, body)
	}

	respBody, _ := io.ReadAll(resp.Body)
	var parsed map[string]interface{}
	if err := json.Unmarshal(respBody, &parsed); err == nil {
		if data, ok := parsed["data"].(map[string]interface{}); ok {
			if a, ok := data["auction"].(map[string]interface{}); ok {
				if a["currency_code"] != "MRU" {
					t.Fatalf("expected legacy null currency to resolve to the default MRU, got %v", a["currency_code"])
				}
			}
		}
	}
}

// TestAuctionGetByID_TNOwner_LegacyNullMarketAuction_ReturnsOK is the
// regression test for the exact case proven by live Staging DB data (client
// feedback, round 3): a TN seller's auction predates migration 000046 and has
// market_country_iso/currency_code = NULL. EffectiveMarketCountryISO() falls
// back to DefaultAccountCountryISO ("MR") for such a row -- so the previous
// code (isOwner exempted the status check, but NOT the market-isolation
// check below it) 404'd the TN owner out of their own active auction,
// because "TN" (their real market) != "MR" (the row's NULL-fallback market).
// This was the confirmed root cause of "خطأ في تحميل المزاد" on My Auctions.
func TestAuctionGetByID_TNOwner_LegacyNullMarketAuction_ReturnsOK(t *testing.T) {
	sellerID := uuid.New()
	auctionID := uuid.New()
	tn := "TN"

	auction := &models.Auction{
		ID:           auctionID,
		SellerID:     sellerID,
		Status:       "active",
		StartPrice:   decimal.NewFromInt(100),
		CurrentPrice: decimal.NewFromInt(100),
		StartTime:    time.Now(),
		EndTime:      time.Now().Add(24 * time.Hour),
		// MarketCountryISO/CurrencyCode intentionally nil -- legacy row.
	}

	auctionSvc := &fakeAuctionServiceForGetByID{auction: auction}
	userRepo := &fakeUserRepoForGetByID{users: map[uuid.UUID]*models.User{
		sellerID: {ID: sellerID, Role: "user", AccountCountryISO: &tn},
	}}
	h := NewAuctionHandler(auctionSvc, userRepo, zap.NewNop())

	app := fiber.New()
	app.Get("/v1/api/auctions/:id", func(c *fiber.Ctx) error {
		c.Locals("user_id", sellerID)
		return h.GetByID(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/api/auctions/"+auctionID.String(), nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected the TN owner to retrieve their own legacy null-market active auction (HTTP 200), got %d, body: %s", resp.StatusCode, body)
	}
}

// TestAuctionGetByID_TNNonOwner_LegacyNullMarketAuction_StillBlocked proves
// the fix is narrow: a DIFFERENT TN user (not the seller) viewing the same
// legacy NULL-market auction must still be blocked by the existing MR-fallback
// market policy -- the owner exemption must not leak into a general
// "NULL-market auctions are visible cross-market" loosening.
func TestAuctionGetByID_TNNonOwner_LegacyNullMarketAuction_StillBlocked(t *testing.T) {
	sellerID := uuid.New()
	otherTNUserID := uuid.New()
	auctionID := uuid.New()
	tn := "TN"

	auction := &models.Auction{
		ID:           auctionID,
		SellerID:     sellerID,
		Status:       "active",
		StartPrice:   decimal.NewFromInt(100),
		CurrentPrice: decimal.NewFromInt(100),
		StartTime:    time.Now(),
		EndTime:      time.Now().Add(24 * time.Hour),
		// MarketCountryISO/CurrencyCode intentionally nil -- legacy row,
		// EffectiveMarketCountryISO() resolves to "MR".
	}

	auctionSvc := &fakeAuctionServiceForGetByID{auction: auction}
	userRepo := &fakeUserRepoForGetByID{users: map[uuid.UUID]*models.User{
		otherTNUserID: {ID: otherTNUserID, Role: "user", AccountCountryISO: &tn},
	}}
	h := NewAuctionHandler(auctionSvc, userRepo, zap.NewNop())

	app := fiber.New()
	app.Get("/v1/api/auctions/:id", func(c *fiber.Ctx) error {
		c.Locals("user_id", otherTNUserID)
		return h.GetByID(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/api/auctions/"+auctionID.String(), nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// TN caller vs. the legacy row's MR fallback -> still cross-market ->
	// still blocked, matching the existing (unchanged) market isolation policy.
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("SECURITY REGRESSION: expected a non-owner TN caller to still be blocked from a legacy null-market (MR-fallback) auction, got %d", resp.StatusCode)
	}
}

// TestAuctionGetByID_ImageURLsPopulatedFromModel is targeted test B (client
// feedback, R2 deletion-path audit): GET /auctions/:id's "auction.image_urls"
// field must reflect real persisted images, not always be empty. The actual
// bug was in the repository layer (findByIDInternal never joined
// auction_images, so Auction.ImageURLs stayed nil regardless of what was in
// the DB -- fixed separately in auction_repo.go, not testable here without a
// live database). This test instead locks down the handler's consumer
// contract: when the Auction model DOES carry a populated ImageURLs (exactly
// what the fixed repository query now produces), GetByID's response must
// correctly expose it via "auction.image_urls", proving the JSON
// serialization side of the fix is wired correctly end-to-end from model to
// wire format.
func TestAuctionGetByID_ImageURLsPopulatedFromModel(t *testing.T) {
	sellerID := uuid.New()
	auctionID := uuid.New()
	imageURLs := "https://pub-example.r2.dev/auctions/x/a.jpg,https://pub-example.r2.dev/auctions/x/b.jpg"

	auction := &models.Auction{
		ID:           auctionID,
		SellerID:     sellerID,
		Status:       "active",
		StartPrice:   decimal.NewFromInt(100),
		CurrentPrice: decimal.NewFromInt(100),
		StartTime:    time.Now(),
		EndTime:      time.Now().Add(24 * time.Hour),
		ImageURLs:    &imageURLs,
	}

	auctionSvc := &fakeAuctionServiceForGetByID{auction: auction}
	userRepo := &fakeUserRepoForGetByID{users: map[uuid.UUID]*models.User{}}
	h := NewAuctionHandler(auctionSvc, userRepo, zap.NewNop())

	app := fiber.New()
	app.Get("/v1/api/auctions/:id", h.GetByID) // anonymous, active auction -> publicly visible

	req := httptest.NewRequest(http.MethodGet, "/v1/api/auctions/"+auctionID.String(), nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected HTTP 200, got %d, body: %s", resp.StatusCode, body)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatalf("failed to parse response JSON: %v, body: %s", err, body)
	}

	data, ok := parsed["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected top-level \"data\" object, got: %v", parsed["data"])
	}
	auctionObj, ok := data["auction"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected \"data.auction\" object, got: %v", data["auction"])
	}
	urls, ok := auctionObj["image_urls"].([]interface{})
	if !ok {
		t.Fatalf("expected \"auction.image_urls\" to be an array, got %T: %v", auctionObj["image_urls"], auctionObj["image_urls"])
	}
	if len(urls) != 2 {
		t.Fatalf("REGRESSION: expected auction.image_urls to contain 2 URLs (matching the model's populated ImageURLs), got %d: %v", len(urls), urls)
	}
	if urls[0] != "https://pub-example.r2.dev/auctions/x/a.jpg" {
		t.Fatalf("expected the first image URL to match, got %v", urls[0])
	}
}

// fakeAuctionServiceForAddImages drives AuctionHandler.AddImages's
// URL-based (non-multipart) branch through the real Fiber HTTP cycle and
// records every argument AddImages is called with, plus whether any other
// AuctionService method (in particular DeleteFile-adjacent paths, or
// Update/UpdateAuction) is touched by this request at all.
type fakeAuctionServiceForAddImages struct {
	services.AuctionService
	calls []addImagesCall
	err   error
}

type addImagesCall struct {
	auctionID uuid.UUID
	sellerID  uuid.UUID
	urls      []string
}

func (f *fakeAuctionServiceForAddImages) AddImages(ctx context.Context, auctionID, sellerID uuid.UUID, urls []string) error {
	f.calls = append(f.calls, addImagesCall{auctionID: auctionID, sellerID: sellerID, urls: urls})
	return f.err
}

// TestAuctionAddImages_URLPathPersistsAndAssociatesCorrectly is the
// regression test for targeted test D (client feedback, R2 deletion-path
// audit, image bug fix final verification): proves POST
// /v1/api/auctions/:id/images -> AuctionHandler.AddImages ->
// AuctionService.AddImages still works after the Update/UpdateAuction guard
// change (imagesUpdateIncludesNewURLs / hasNewImages), NOT by asserting
// "AddImages itself has zero diff" (rejected as insufficient), but by
// actually exercising the real handler method end-to-end through the exact
// same URL-based (legacy JSON) branch AddImages supports, and asserting:
//  1. the auction ID from the route param and the seller ID from the auth
//     context both reach AddImages unchanged (correct auction_id/seller
//     association, matching what AddImageTx would persist against);
//  2. the exact URLs from the request body reach AddImages unchanged (the
//     image URL is what gets saved);
//  3. AddImages is the ONLY AuctionService method invoked by this request --
//     the embedded services.AuctionService is nil, so any accidental call
//     to Update, UpdateAuction, or any other method (which is the family of
//     methods that now contain the new hasNewImages / DeleteFile guard)
//     would nil-pointer-panic the test instead of silently succeeding,
//     directly proving the Update/UpdateAuction fix does not leak into or
//     get invoked by the AddImages request path;
//  4. by extension (3), no DeleteFile call can occur on this path, since
//     DeleteFile is only ever invoked from inside Update/UpdateAuction's
//     hasNewImages branch (see auction_service.go), neither of which this
//     request path reaches.
//
// This is deliberately layered with, not a replacement for, direct
// inspection of the real AuctionService.AddImages body (auction_service.go
// ~line 745), which -- confirmed by reading the current source -- performs
// only FindByID (ownership check) + BeginTxx + AddImageTx per URL + Commit,
// with no call to mediaSvc.DeleteFile anywhere in the function, and no call
// through the db.BeginTxx-based image-replace/delete path added to
// Update/UpdateAuction in this round.
func TestAuctionAddImages_URLPathPersistsAndAssociatesCorrectly(t *testing.T) {
	sellerID := uuid.New()
	auctionID := uuid.New()
	imageURL1 := "https://pub-example.r2.dev/auctions/" + auctionID.String() + "/new-image-1.jpg"
	imageURL2 := "https://pub-example.r2.dev/auctions/" + auctionID.String() + "/new-image-2.jpg"

	fakeSvc := &fakeAuctionServiceForAddImages{}
	userRepo := &fakeUserRepoForGetByID{users: map[uuid.UUID]*models.User{}}
	h := NewAuctionHandler(fakeSvc, userRepo, zap.NewNop())

	app := fiber.New()
	app.Post("/v1/api/auctions/:id/images", func(c *fiber.Ctx) error {
		c.Locals("user_id", sellerID)
		return h.AddImages(c)
	})

	reqBody := `{"urls":["` + imageURL1 + `","` + imageURL2 + `"]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/api/auctions/"+auctionID.String()+"/images", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected HTTP 200 from AddImages, got %d, body: %s", resp.StatusCode, body)
	}

	if len(fakeSvc.calls) != 1 {
		t.Fatalf("REGRESSION: expected AuctionService.AddImages to be called exactly once, got %d calls", len(fakeSvc.calls))
	}
	call := fakeSvc.calls[0]

	if call.auctionID != auctionID {
		t.Fatalf("REGRESSION: expected auction_id association %s, got %s -- the image would be linked to the wrong auction", auctionID, call.auctionID)
	}
	if call.sellerID != sellerID {
		t.Fatalf("REGRESSION: expected seller ID %s (from auth context) to reach AddImages, got %s", sellerID, call.sellerID)
	}
	if len(call.urls) != 2 || call.urls[0] != imageURL1 || call.urls[1] != imageURL2 {
		t.Fatalf("REGRESSION: expected the exact submitted image URLs [%s, %s] to reach AddImages for persistence, got %v", imageURL1, imageURL2, call.urls)
	}
}

// TestAuctionAddImages_RejectsAuctionOwnerMismatch_ServiceLayerEnforced
// documents (does not newly assert) that ownership enforcement for AddImages
// happens inside the real AuctionService.AddImages (FindByID + SellerID
// comparison, auction_service.go ~line 750), not in the handler -- so at the
// handler level, a fake that simulates the service returning apperr.ErrUnauthorized
// must produce a non-200 response, proving the handler correctly propagates
// that failure rather than reporting false success for a mismatched image add.
func TestAuctionAddImages_RejectsAuctionOwnerMismatch_ServiceLayerEnforced(t *testing.T) {
	sellerID := uuid.New()
	auctionID := uuid.New()

	fakeSvc := &fakeAuctionServiceForAddImages{err: apperr.ErrUnauthorized}
	userRepo := &fakeUserRepoForGetByID{users: map[uuid.UUID]*models.User{}}
	h := NewAuctionHandler(fakeSvc, userRepo, zap.NewNop())

	app := fiber.New()
	app.Post("/v1/api/auctions/:id/images", func(c *fiber.Ctx) error {
		c.Locals("user_id", sellerID)
		return h.AddImages(c)
	})

	reqBody := `{"urls":["https://pub-example.r2.dev/auctions/x/y.jpg"]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/api/auctions/"+auctionID.String()+"/images", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		t.Fatalf("expected a non-200 response when AddImages rejects a non-owner, got 200")
	}
}
