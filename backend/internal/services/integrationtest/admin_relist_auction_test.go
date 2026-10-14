//go:build integration

package integrationtest

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	apperr "github.com/mazadpay/backend/internal/errors"
	"github.com/mazadpay/backend/internal/models"
	"github.com/shopspring/decimal"
)

// Customer Request #37: admin-only relist of an ended auction, reusing the
// exact Bug J atomic transition (TryRelistAuctionAtomically) already proven
// for the seller-facing relist flow. Unlike that flow, AdminService.RelistAuction
// has no seller-ownership check and derives the new end_time itself from the
// auction's own stored (start_time, end_time) -- never trusted from a caller.

// createEndedAuctionWithPreciseDuration builds an ended auction with an
// exact, known original duration (start_time -> end_time), so tests can
// assert the relisted auction's new duration matches precisely rather than
// just "some positive span."
func createEndedAuctionWithPreciseDuration(t *testing.T, env *testEnv, sellerID uuid.UUID, duration time.Duration) *models.Auction {
	t.Helper()
	ctx := context.Background()
	marketISO, currencyCode := "MR", "MRU"
	lotNumber := "TEST-RELIST-" + uuid.New().String()[:8]
	start := time.Now().Add(-2 * duration)
	a := &models.Auction{
		ID:               uuid.New(),
		SellerID:         sellerID,
		CategoryID:       mkCategoryID(),
		TitleAr:          "مزاد اختبار إعادة النشر " + uuid.New().String()[:6],
		LotNumber:        &lotNumber,
		StartPrice:       decimal.NewFromInt(100),
		CurrentPrice:     decimal.NewFromInt(100),
		MinIncrement:     decimal.NewFromInt(10),
		InsuranceAmount:  decimal.NewFromInt(20),
		InsurancePolicy:  "not_required",
		ReservePrice:     decimal.NewFromInt(100),
		StartTime:        time.Now().Add(-2 * time.Hour), // created non-expired, then backdated below
		EndTime:          time.Now().Add(48 * time.Hour),
		Status:           "active",
		MarketCountryISO: &marketISO,
		CurrencyCode:     &currencyCode,
	}
	if err := env.auctionRepo.Create(ctx, nil, a); err != nil {
		t.Fatalf("failed to create fixture auction: %v", err)
	}
	end := start.Add(duration)
	if _, err := env.db.ExecContext(ctx,
		`UPDATE auctions SET start_time = $1, end_time = $2, status = 'ended' WHERE id = $3`,
		start, end, a.ID); err != nil {
		t.Fatalf("failed to set precise start/end/status on fixture auction: %v", err)
	}
	if err := env.db.GetContext(ctx, a, `SELECT * FROM auctions WHERE id = $1`, a.ID); err != nil {
		t.Fatalf("failed to read back fixture auction: %v", err)
	}
	return a
}

func countBidsForAuction(t *testing.T, env *testEnv, auctionID uuid.UUID) int {
	t.Helper()
	var count int
	if err := env.db.Get(&count, `SELECT COUNT(*) FROM bids WHERE auction_id = $1`, auctionID); err != nil {
		t.Fatalf("failed to count bids: %v", err)
	}
	return count
}

func TestAdminRelistAuction_EndedAuctionRelists(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST RELIST SELLER")
	admin := createTestAdmin(t, env, "TEST RELIST ADMIN")
	adminSvc := newTestAdminService(t, env)

	auction := createEndedAuctionWithPreciseDuration(t, env, seller.ID, 3*time.Hour)

	relisted, err := adminSvc.RelistAuction(ctx, auction.ID, admin.ID)
	if err != nil {
		t.Fatalf("RelistAuction failed: %v", err)
	}
	if relisted.Status != "pending" {
		t.Fatalf("expected status 'pending' after relist, got %q", relisted.Status)
	}

	var fetched models.Auction
	if err := env.db.Get(&fetched, `SELECT * FROM auctions WHERE id = $1`, auction.ID); err != nil {
		t.Fatalf("failed to read back relisted auction: %v", err)
	}
	if fetched.Status != "pending" {
		t.Fatalf("expected DB status 'pending', got %q", fetched.Status)
	}
	if !fetched.EndTime.After(time.Now()) {
		t.Fatalf("expected new end_time to be in the future, got %s", fetched.EndTime)
	}
}

func TestAdminRelistAuction_ActiveAuctionRejected(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST RELIST ACTIVE REJECT SELLER")
	admin := createTestAdmin(t, env, "TEST RELIST ACTIVE REJECT ADMIN")
	adminSvc := newTestAdminService(t, env)

	auction := createEndedAuctionWithPreciseDuration(t, env, seller.ID, 3*time.Hour)
	if _, err := env.db.ExecContext(ctx, `UPDATE auctions SET status = 'active' WHERE id = $1`, auction.ID); err != nil {
		t.Fatalf("failed to force status active: %v", err)
	}

	if _, err := adminSvc.RelistAuction(ctx, auction.ID, admin.ID); err != apperr.ErrConflict {
		t.Fatalf("expected ErrConflict for an active auction, got %v", err)
	}

	var fetched models.Auction
	if err := env.db.Get(&fetched, `SELECT * FROM auctions WHERE id = $1`, auction.ID); err != nil {
		t.Fatalf("failed to read back auction: %v", err)
	}
	if fetched.Status != "active" {
		t.Fatalf("expected status to remain 'active' after a rejected relist, got %q", fetched.Status)
	}
}

func TestAdminRelistAuction_PendingAuctionRejected(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST RELIST PENDING REJECT SELLER")
	admin := createTestAdmin(t, env, "TEST RELIST PENDING REJECT ADMIN")
	adminSvc := newTestAdminService(t, env)

	auction := createEndedAuctionWithPreciseDuration(t, env, seller.ID, 3*time.Hour)
	if _, err := env.db.ExecContext(ctx, `UPDATE auctions SET status = 'pending' WHERE id = $1`, auction.ID); err != nil {
		t.Fatalf("failed to force status pending: %v", err)
	}

	if _, err := adminSvc.RelistAuction(ctx, auction.ID, admin.ID); err != apperr.ErrConflict {
		t.Fatalf("expected ErrConflict for a pending auction, got %v", err)
	}
}

// The relisted auction's new duration (end_time - relist-time) must match
// the ORIGINAL duration (original end_time - original start_time) exactly.
func TestAdminRelistAuction_SameDurationPreserved(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST RELIST DURATION SELLER")
	admin := createTestAdmin(t, env, "TEST RELIST DURATION ADMIN")
	adminSvc := newTestAdminService(t, env)

	const originalDuration = 5 * time.Hour
	auction := createEndedAuctionWithPreciseDuration(t, env, seller.ID, originalDuration)

	before := time.Now()
	if _, err := adminSvc.RelistAuction(ctx, auction.ID, admin.ID); err != nil {
		t.Fatalf("RelistAuction failed: %v", err)
	}
	after := time.Now()

	var fetched models.Auction
	if err := env.db.Get(&fetched, `SELECT * FROM auctions WHERE id = $1`, auction.ID); err != nil {
		t.Fatalf("failed to read back relisted auction: %v", err)
	}

	// new_end_time should fall within [before, after] + originalDuration,
	// allowing for the tiny window the RelistAuction call itself took.
	minExpected := before.Add(originalDuration)
	maxExpected := after.Add(originalDuration)
	if fetched.EndTime.Before(minExpected.Add(-time.Second)) || fetched.EndTime.After(maxExpected.Add(time.Second)) {
		t.Fatalf("expected new end_time within [%s, %s] (now + original 5h duration), got %s", minExpected, maxExpected, fetched.EndTime)
	}

	// start_time must NOT be modified by relist.
	if !fetched.StartTime.Equal(auction.StartTime) {
		t.Fatalf("expected start_time to remain unchanged, was %s, now %s", auction.StartTime, fetched.StartTime)
	}
}

// Winner/payment terminal fields must all be cleared, and price reset to
// the original start price.
func TestAdminRelistAuction_WinnerStateCleared(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	adminSvc := newTestAdminService(t, env)
	admin := createTestAdmin(t, env, "TEST RELIST WINNER CLEAR ADMIN")

	auction, winner, _ := setupEndedAuctionWithWinnerHold(t, env)
	if auction.WinnerID == nil || *auction.WinnerID != winner.ID {
		t.Fatalf("precondition failed: expected a real winner on the fixture auction")
	}

	if _, err := adminSvc.RelistAuction(ctx, auction.ID, admin.ID); err != nil {
		t.Fatalf("RelistAuction failed: %v", err)
	}

	var fetched models.Auction
	if err := env.db.Get(&fetched, `SELECT * FROM auctions WHERE id = $1`, auction.ID); err != nil {
		t.Fatalf("failed to read back relisted auction: %v", err)
	}
	if fetched.WinnerID != nil {
		t.Fatalf("expected winner_id to be cleared, got %v", *fetched.WinnerID)
	}
	if fetched.WinningBidID != nil {
		t.Fatalf("expected winning_bid_id to be cleared, got %v", *fetched.WinningBidID)
	}
	if fetched.PaymentDeadline != nil {
		t.Fatalf("expected payment_deadline to be cleared, got %v", *fetched.PaymentDeadline)
	}
	if !fetched.CurrentPrice.Equal(fetched.StartPrice) {
		t.Fatalf("expected current_price reset to start_price (%s), got %s", fetched.StartPrice, fetched.CurrentPrice)
	}
}

// Historical bid rows must never be deleted or modified by relist.
func TestAdminRelistAuction_HistoricalBidsPreserved(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	adminSvc := newTestAdminService(t, env)
	admin := createTestAdmin(t, env, "TEST RELIST BIDS PRESERVED ADMIN")

	auction, _, _ := setupEndedAuctionWithWinnerHold(t, env)

	bidsBefore := countBidsForAuction(t, env, auction.ID)
	if bidsBefore == 0 {
		t.Fatalf("precondition failed: expected at least 1 bid on the fixture auction")
	}

	if _, err := adminSvc.RelistAuction(ctx, auction.ID, admin.ID); err != nil {
		t.Fatalf("RelistAuction failed: %v", err)
	}

	bidsAfter := countBidsForAuction(t, env, auction.ID)
	if bidsAfter != bidsBefore {
		t.Fatalf("expected bid count unchanged after relist (before=%d, after=%d)", bidsBefore, bidsAfter)
	}
}

// A lost-race guard: relisting an already-relisted (now 'pending') auction a
// second time must be rejected cleanly, not silently succeed twice.
func TestAdminRelistAuction_SecondRelistRejected(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST RELIST TWICE SELLER")
	admin := createTestAdmin(t, env, "TEST RELIST TWICE ADMIN")
	adminSvc := newTestAdminService(t, env)

	auction := createEndedAuctionWithPreciseDuration(t, env, seller.ID, 2*time.Hour)

	if _, err := adminSvc.RelistAuction(ctx, auction.ID, admin.ID); err != nil {
		t.Fatalf("first RelistAuction failed: %v", err)
	}
	if _, err := adminSvc.RelistAuction(ctx, auction.ID, admin.ID); err != apperr.ErrConflict {
		t.Fatalf("expected ErrConflict on second relist attempt (now pending, not ended), got %v", err)
	}
}
