//go:build integration

package integrationtest

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	apperr "github.com/mazadpay/backend/internal/errors"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/services"
	"github.com/shopspring/decimal"
)

// Customer Request #23: integration tests against the REAL canonical
// finalization path (auctionService.FinalizeExpiredAuction /
// CloseExpiredAuctions), not the finalizeAuctionAsWinner test helper (which
// calls repository.SetWinner directly and therefore never exercises the
// production race/finalization logic this round replaced). These run
// against a real local Postgres (see integration_test.go's build-tag doc
// comment for exact invocation), proving the DB-level race guard
// (WHERE status='active' ... RETURNING id) actually provides exactly-once
// semantics under real concurrent transactions.

// createExpiredTestAuction is createTestAuctionInsured's twin, except
// end_time is stamped in the past so FindExpiredActive/FindEndedSince (and
// therefore FinalizeExpiredAuction/CloseExpiredAuctions) will actually pick
// it up as a real expired-active candidate.
func createExpiredTestAuction(t *testing.T, env *testEnv, sellerID uuid.UUID) *models.Auction {
	t.Helper()
	ctx := context.Background()
	marketISO, currencyCode := "MR", "MRU"
	lotNumber := "TEST-EXP-" + uuid.New().String()[:8]
	a := &models.Auction{
		ID:               uuid.New(),
		SellerID:         sellerID,
		CategoryID:       mkCategoryID(),
		TitleAr:          "مزاد اختبار انتهاء " + uuid.New().String()[:6],
		LotNumber:        &lotNumber,
		StartPrice:       decimal.NewFromInt(100),
		CurrentPrice:     decimal.NewFromInt(100),
		MinIncrement:     decimal.NewFromInt(10),
		InsuranceAmount:  decimal.NewFromInt(20),
		InsurancePolicy:  "not_required",
		ReservePrice:     decimal.NewFromInt(100),
		StartTime:        time.Now().Add(-2 * time.Hour),
		EndTime:          time.Now().Add(48 * time.Hour), // created non-expired, then back-dated below
		Status:           "active",
		MarketCountryISO: &marketISO,
		CurrencyCode:     &currencyCode,
	}
	if err := env.auctionRepo.Create(ctx, nil, a); err != nil {
		t.Fatalf("failed to create expired fixture auction: %v", err)
	}
	// Back-date end_time directly -- Create's own validation would reject an
	// already-past end_time, so the auction is created valid, then expired.
	if _, err := env.db.ExecContext(ctx, `UPDATE auctions SET end_time = $1 WHERE id = $2`, time.Now().Add(-1*time.Minute), a.ID); err != nil {
		t.Fatalf("failed to back-date fixture auction end_time: %v", err)
	}
	if err := env.db.GetContext(ctx, a, `SELECT * FROM auctions WHERE id = $1`, a.ID); err != nil {
		t.Fatalf("failed to read back expired fixture auction: %v", err)
	}
	return a
}

func countAuctionWonNotifications(t *testing.T, env *testEnv, userID, auctionID uuid.UUID) int {
	t.Helper()
	var count int
	err := env.db.Get(&count, `
		SELECT COUNT(*) FROM notifications
		WHERE user_id = $1 AND type = 'auction_won'
		AND data->>'auctionId' = $2
	`, userID, auctionID.String())
	if err != nil {
		t.Fatalf("failed to count auction_won notifications: %v", err)
	}
	return count
}

func countAuctionEndedNotifications(t *testing.T, env *testEnv, userID, auctionID uuid.UUID) int {
	t.Helper()
	var count int
	err := env.db.Get(&count, `
		SELECT COUNT(*) FROM notifications
		WHERE user_id = $1 AND type = 'auction_ended'
		AND data->>'auctionId' = $2
	`, userID, auctionID.String())
	if err != nil {
		t.Fatalf("failed to count auction_ended notifications: %v", err)
	}
	return count
}

// TestFinalizeExpiredAuction_RealPath_WinnerPersistedCorrectly (Section 21):
// the actual production finalization path, not a SetWinner shortcut.
func TestFinalizeExpiredAuction_RealPath_WinnerPersistedCorrectly(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()

	seller := createTestUser(t, env, "TEST FINALIZE SELLER")
	bidderA := createTestUser(t, env, "TEST FINALIZE BIDDER A")
	bidderB := createTestUser(t, env, "TEST FINALIZE BIDDER B")
	creditWallet(t, env, bidderA.ID, decimal.NewFromInt(1000))
	creditWallet(t, env, bidderB.ID, decimal.NewFromInt(1000))

	auction := createExpiredTestAuction(t, env, seller.ID)
	// Bids must be placed while the auction is still active from PlaceBid's
	// own perspective -- but since we already back-dated end_time, use a
	// second, not-yet-expired auction for placing bids, then back-date it
	// too. Simpler: temporarily extend end_time, bid, then re-expire.
	if _, err := env.db.ExecContext(ctx, `UPDATE auctions SET end_time = $1 WHERE id = $2`, time.Now().Add(1*time.Hour), auction.ID); err != nil {
		t.Fatalf("failed to temporarily un-expire auction for bidding: %v", err)
	}

	if _, err := env.bidSvc.PlaceBid(ctx, auction.ID, bidderA.ID, decimal.NewFromInt(150)); err != nil {
		t.Fatalf("bidderA PlaceBid failed: %v", err)
	}
	topBid, err := env.bidSvc.PlaceBid(ctx, auction.ID, bidderB.ID, decimal.NewFromInt(200))
	if err != nil {
		t.Fatalf("bidderB PlaceBid failed: %v", err)
	}

	// Re-expire now that bidding is done.
	if _, err := env.db.ExecContext(ctx, `UPDATE auctions SET end_time = $1 WHERE id = $2`, time.Now().Add(-1*time.Minute), auction.ID); err != nil {
		t.Fatalf("failed to re-expire auction: %v", err)
	}

	if err := env.auctSvc.FinalizeExpiredAuction(ctx, auction.ID); err != nil {
		t.Fatalf("FinalizeExpiredAuction failed: %v", err)
	}

	var final models.Auction
	if err := env.db.Get(&final, `SELECT * FROM auctions WHERE id = $1`, auction.ID); err != nil {
		t.Fatalf("failed to read back finalized auction: %v", err)
	}

	t.Run("status = ended", func(t *testing.T) {
		if final.Status != "ended" {
			t.Fatalf("expected status 'ended', got %q", final.Status)
		}
	})
	t.Run("winner_id = highest bidder", func(t *testing.T) {
		if final.WinnerID == nil || *final.WinnerID != bidderB.ID {
			t.Fatalf("expected winner_id %s, got %v", bidderB.ID, final.WinnerID)
		}
	})
	t.Run("winning_bid_id = highest bid", func(t *testing.T) {
		if final.WinningBidID == nil || *final.WinningBidID != topBid.ID {
			t.Fatalf("expected winning_bid_id %s, got %v", topBid.ID, final.WinningBidID)
		}
	})
	t.Run("current_price remains the winning amount", func(t *testing.T) {
		if !final.CurrentPrice.Equal(decimal.NewFromInt(200)) {
			t.Fatalf("expected current_price 200, got %s", final.CurrentPrice.String())
		}
	})

	t.Run("winner's My Winnings contains the auction", func(t *testing.T) {
		winnings, err := env.userSvc.ListMyWinnings(ctx, bidderB.ID)
		if err != nil {
			t.Fatalf("ListMyWinnings failed: %v", err)
		}
		found := false
		for _, w := range winnings {
			if w.ID == auction.ID {
				found = true
			}
		}
		if !found {
			t.Fatalf("expected bidderB's My Winnings to contain auction %s", auction.ID)
		}
	})
	t.Run("losing bidder's My Winnings does not contain the auction", func(t *testing.T) {
		winnings, err := env.userSvc.ListMyWinnings(ctx, bidderA.ID)
		if err != nil {
			t.Fatalf("ListMyWinnings failed: %v", err)
		}
		for _, w := range winnings {
			if w.ID == auction.ID {
				t.Fatalf("SECURITY/CORRECTNESS REGRESSION: losing bidderA's My Winnings contained the auction")
			}
		}
	})
	t.Run("exactly one auction_won notification for the winner", func(t *testing.T) {
		count := countAuctionWonNotifications(t, env, bidderB.ID, auction.ID)
		if count != 1 {
			t.Fatalf("expected exactly 1 auction_won notification, got %d", count)
		}
	})
	t.Run("no auction_won notification for the losing bidder", func(t *testing.T) {
		count := countAuctionWonNotifications(t, env, bidderA.ID, auction.ID)
		if count != 0 {
			t.Fatalf("expected 0 auction_won notifications for the losing bidder, got %d", count)
		}
	})
}

// TestFinalizeExpiredAuction_ConcurrentCallers_ExactlyOneWinner (Section 22):
// simulates the exact real production race between AuctionScheduler and
// CloseExpiredAuctions by calling FinalizeExpiredAuction concurrently from
// two goroutines for the SAME auction, proving the DB-level guard (not
// Redis, not an application mutex) is what makes this safe.
func TestFinalizeExpiredAuction_ConcurrentCallers_ExactlyOneWinner(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()

	seller := createTestUser(t, env, "TEST RACE SELLER")
	bidder := createTestUser(t, env, "TEST RACE BIDDER")
	creditWallet(t, env, bidder.ID, decimal.NewFromInt(1000))

	auction := createExpiredTestAuction(t, env, seller.ID)
	if _, err := env.db.ExecContext(ctx, `UPDATE auctions SET end_time = $1 WHERE id = $2`, time.Now().Add(1*time.Hour), auction.ID); err != nil {
		t.Fatalf("failed to temporarily un-expire auction for bidding: %v", err)
	}
	topBid, err := env.bidSvc.PlaceBid(ctx, auction.ID, bidder.ID, decimal.NewFromInt(150))
	if err != nil {
		t.Fatalf("PlaceBid failed: %v", err)
	}
	if _, err := env.db.ExecContext(ctx, `UPDATE auctions SET end_time = $1 WHERE id = $2`, time.Now().Add(-1*time.Minute), auction.ID); err != nil {
		t.Fatalf("failed to re-expire auction: %v", err)
	}

	// Two concurrent finalization attempts for the SAME auction ID --
	// mirrors AuctionScheduler.checkEndedAuctions and
	// auctionService.CloseExpiredAuctions both discovering and finalizing
	// the same candidate in the same real-world tick window.
	var wg sync.WaitGroup
	errs := make([]error, 2)
	wg.Add(2)
	go func() {
		defer wg.Done()
		errs[0] = env.auctSvc.FinalizeExpiredAuction(context.Background(), auction.ID)
	}()
	go func() {
		defer wg.Done()
		errs[1] = env.auctSvc.FinalizeExpiredAuction(context.Background(), auction.ID)
	}()
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("concurrent FinalizeExpiredAuction call %d returned an error (expected a safe no-op, never an error/panic): %v", i, err)
		}
	}

	var final models.Auction
	if err := env.db.Get(&final, `SELECT * FROM auctions WHERE id = $1`, auction.ID); err != nil {
		t.Fatalf("failed to read back finalized auction: %v", err)
	}

	t.Run("exactly one active->ended transition, winner set", func(t *testing.T) {
		if final.Status != "ended" {
			t.Fatalf("expected status 'ended', got %q", final.Status)
		}
		if final.WinnerID == nil || *final.WinnerID != bidder.ID {
			t.Fatalf("expected winner_id %s, got %v", bidder.ID, final.WinnerID)
		}
		if final.WinningBidID == nil || *final.WinningBidID != topBid.ID {
			t.Fatalf("expected winning_bid_id %s, got %v", topBid.ID, final.WinningBidID)
		}
	})

	t.Run("exactly one My Winnings row, not duplicated", func(t *testing.T) {
		winnings, err := env.userSvc.ListMyWinnings(ctx, bidder.ID)
		if err != nil {
			t.Fatalf("ListMyWinnings failed: %v", err)
		}
		count := 0
		for _, w := range winnings {
			if w.ID == auction.ID {
				count++
			}
		}
		if count != 1 {
			t.Fatalf("expected exactly 1 winnings row for this auction, got %d", count)
		}
	})

	t.Run("exactly one auction_won notification despite two concurrent finalize calls", func(t *testing.T) {
		count := countAuctionWonNotifications(t, env, bidder.ID, auction.ID)
		if count != 1 {
			t.Fatalf("DUPLICATE NOTIFICATION REGRESSION: expected exactly 1 auction_won notification, got %d -- only the DB-race winner should ever send notifications", count)
		}
	})

	t.Run("exactly one seller auction_ended notification despite two concurrent finalize calls", func(t *testing.T) {
		count := countAuctionEndedNotifications(t, env, seller.ID, auction.ID)
		if count != 1 {
			t.Fatalf("DUPLICATE NOTIFICATION REGRESSION: expected exactly 1 auction_ended notification to the seller, got %d", count)
		}
	})
}

// TestFinalizeExpiredAuction_NoBids_NoWinner (Section 23).
func TestFinalizeExpiredAuction_NoBids_NoWinner(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()

	seller := createTestUser(t, env, "TEST NOBID SELLER")
	auction := createExpiredTestAuction(t, env, seller.ID)

	if err := env.auctSvc.FinalizeExpiredAuction(ctx, auction.ID); err != nil {
		t.Fatalf("FinalizeExpiredAuction failed: %v", err)
	}

	var final models.Auction
	if err := env.db.Get(&final, `SELECT * FROM auctions WHERE id = $1`, auction.ID); err != nil {
		t.Fatalf("failed to read back finalized auction: %v", err)
	}

	if final.Status != "ended" {
		t.Fatalf("expected status 'ended', got %q", final.Status)
	}
	if final.WinnerID != nil {
		t.Fatalf("expected winner_id NULL for a no-bid auction, got %v", *final.WinnerID)
	}
	if final.WinningBidID != nil {
		t.Fatalf("expected winning_bid_id NULL for a no-bid auction, got %v", *final.WinningBidID)
	}
}

// createActiveBuyNowTestAuction is createTestAuctionInsured's twin for BuyNow
// scenarios: a real, currently-active (not expired) auction with a
// buy_now_price set.
func createActiveBuyNowTestAuction(t *testing.T, env *testEnv, sellerID uuid.UUID, buyNowPrice decimal.Decimal) *models.Auction {
	t.Helper()
	ctx := context.Background()
	lotNumber := "TEST-BUYNOW-" + uuid.New().String()[:8]
	a := &models.Auction{
		ID:              uuid.New(),
		SellerID:        sellerID,
		CategoryID:      mkCategoryID(),
		TitleAr:         "مزاد اختبار شراء فوري " + uuid.New().String()[:6],
		LotNumber:       &lotNumber,
		StartPrice:      decimal.NewFromInt(100),
		CurrentPrice:    decimal.NewFromInt(100),
		MinIncrement:    decimal.NewFromInt(10),
		InsuranceAmount: decimal.NewFromInt(20),
		InsurancePolicy: "not_required",
		ReservePrice:    decimal.NewFromInt(100),
		StartTime:       time.Now().Add(-1 * time.Hour),
		EndTime:         time.Now().Add(48 * time.Hour),
		Status:          "active",
		BuyNowPrice:     &buyNowPrice,
	}
	if err := env.auctionRepo.Create(ctx, nil, a); err != nil {
		t.Fatalf("failed to create buy-now fixture auction: %v", err)
	}
	return a
}

// TestCancelAuction_NoWinnerCreated (Bug J, Section 24 -- was previously RED
// due to auctionRepo.Update silently dropping status/rejection_reason; now
// GREEN via TryCancelAuctionAtomically).
func TestCancelAuction_NoWinnerCreated(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()

	seller := createTestUser(t, env, "TEST CANCEL SELLER")
	bidder := createTestUser(t, env, "TEST CANCEL BIDDER")
	creditWallet(t, env, bidder.ID, decimal.NewFromInt(1000))

	auction := createTestAuctionInsured(t, env, seller.ID, "MR", "MRU")
	if _, err := env.bidSvc.PlaceBid(ctx, auction.ID, bidder.ID, decimal.NewFromInt(150)); err != nil {
		t.Fatalf("PlaceBid failed: %v", err)
	}

	if err := env.auctSvc.CancelAuction(ctx, auction.ID, seller.ID, "test cancellation"); err != nil {
		t.Fatalf("CancelAuction failed: %v", err)
	}

	var final models.Auction
	if err := env.db.Get(&final, `SELECT * FROM auctions WHERE id = $1`, auction.ID); err != nil {
		t.Fatalf("failed to read back cancelled auction: %v", err)
	}
	t.Run("status persists as canceled", func(t *testing.T) {
		if final.Status != "canceled" {
			t.Fatalf("expected status 'canceled', got %q", final.Status)
		}
	})
	t.Run("rejection_reason persists", func(t *testing.T) {
		if final.RejectionReason == nil || *final.RejectionReason != "test cancellation" {
			t.Fatalf("expected rejection_reason 'test cancellation', got %v", final.RejectionReason)
		}
	})
	t.Run("no winner created", func(t *testing.T) {
		if final.WinnerID != nil {
			t.Fatalf("expected no winner on a cancelled auction, got %v", *final.WinnerID)
		}
	})
	t.Run("no auction_won sent, no My Winnings entry", func(t *testing.T) {
		winnings, err := env.userSvc.ListMyWinnings(ctx, bidder.ID)
		if err != nil {
			t.Fatalf("ListMyWinnings failed: %v", err)
		}
		for _, w := range winnings {
			if w.ID == auction.ID {
				t.Fatalf("REGRESSION: a cancelled auction must never appear in a bidder's My Winnings")
			}
		}
		if countAuctionWonNotifications(t, env, bidder.ID, auction.ID) != 0 {
			t.Fatalf("REGRESSION: a cancelled auction must never send auction_won")
		}
	})
	t.Run("repeated cancel does not repeat terminal side effects (safe conflict, not a crash)", func(t *testing.T) {
		err := env.auctSvc.CancelAuction(ctx, auction.ID, seller.ID, "second cancel attempt")
		if err == nil {
			t.Fatalf("expected the second cancel of an already-cancelled auction to fail, got nil error")
		}
		var reread models.Auction
		if err := env.db.Get(&reread, `SELECT * FROM auctions WHERE id = $1`, auction.ID); err != nil {
			t.Fatalf("failed to read back auction: %v", err)
		}
		if reread.RejectionReason == nil || *reread.RejectionReason != "test cancellation" {
			t.Fatalf("expected the ORIGINAL rejection_reason to remain unchanged after a rejected repeat cancel, got %v", reread.RejectionReason)
		}
	})
}

// TestCancelAuction_VsExpiration_Race (Bug J, Section 4 / Section 10-17).
// Runs a real concurrent race between CancelAuction and
// FinalizeExpiredAuction for the same auction. Exactly one terminal
// transition may win; never both, never a fake cancel event with the DB
// left 'ended', never status 'canceled' with a winner attached.
func TestCancelAuction_VsExpiration_Race(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()

	seller := createTestUser(t, env, "TEST CANCELRACE SELLER")
	bidder := createTestUser(t, env, "TEST CANCELRACE BIDDER")
	creditWallet(t, env, bidder.ID, decimal.NewFromInt(1000))

	auction := createExpiredTestAuction(t, env, seller.ID)
	if _, err := env.db.ExecContext(ctx, `UPDATE auctions SET end_time = $1 WHERE id = $2`, time.Now().Add(1*time.Hour), auction.ID); err != nil {
		t.Fatalf("failed to temporarily un-expire auction for bidding: %v", err)
	}
	if _, err := env.bidSvc.PlaceBid(ctx, auction.ID, bidder.ID, decimal.NewFromInt(150)); err != nil {
		t.Fatalf("PlaceBid failed: %v", err)
	}
	if _, err := env.db.ExecContext(ctx, `UPDATE auctions SET end_time = $1 WHERE id = $2`, time.Now().Add(-1*time.Minute), auction.ID); err != nil {
		t.Fatalf("failed to re-expire auction: %v", err)
	}

	var wg sync.WaitGroup
	var cancelErr, finalizeErr error
	wg.Add(2)
	go func() {
		defer wg.Done()
		cancelErr = env.auctSvc.CancelAuction(context.Background(), auction.ID, seller.ID, "race test cancel")
	}()
	go func() {
		defer wg.Done()
		finalizeErr = env.auctSvc.FinalizeExpiredAuction(context.Background(), auction.ID)
	}()
	wg.Wait()

	// FinalizeExpiredAuction is designed to never return an error for a lost
	// race (won=false is a safe no-op) -- only CancelAuction may legitimately
	// report ErrConflict if it lost.
	if finalizeErr != nil {
		t.Fatalf("FinalizeExpiredAuction returned an unexpected error: %v", finalizeErr)
	}

	var final models.Auction
	if err := env.db.Get(&final, `SELECT * FROM auctions WHERE id = $1`, auction.ID); err != nil {
		t.Fatalf("failed to read back auction: %v", err)
	}

	if final.Status != "canceled" && final.Status != "ended" {
		t.Fatalf("expected exactly one of 'canceled' or 'ended', got %q", final.Status)
	}

	if final.Status == "canceled" {
		if final.WinnerID != nil {
			t.Fatalf("INVALID STATE: status is 'canceled' but a winner is attached (%v)", *final.WinnerID)
		}
		if cancelErr != nil {
			t.Fatalf("status is 'canceled' but CancelAuction reported an error: %v", cancelErr)
		}
	} else {
		// Expiration won: cancel must have safely failed (ErrConflict), never
		// silently succeeded while the DB shows 'ended'.
		if cancelErr == nil {
			t.Fatalf("INVALID STATE: status is 'ended' but CancelAuction reported success (no fake-success-with-wrong-DB-state allowed)")
		}
		if final.WinnerID == nil {
			t.Fatalf("expected a winner on the 'ended' auction (real bid was placed)")
		}
	}

	// Whichever path won, no duplicate/conflicting notifications for either
	// event type should exist beyond what a single real transition sends.
	wonCount := countAuctionWonNotifications(t, env, bidder.ID, auction.ID)
	if wonCount > 1 {
		t.Fatalf("expected at most 1 auction_won notification total, got %d", wonCount)
	}
	if final.Status == "canceled" && wonCount != 0 {
		t.Fatalf("expected 0 auction_won notifications when cancel won the race, got %d", wonCount)
	}
}

// TestRelistAuction_WinnerStateCleared (Bug J, Section 25 -- was previously
// RED; now GREEN via TryRelistAuctionAtomically, which additionally clears
// winning_bid_id and payment_deadline together with winner_id).
func TestRelistAuction_WinnerStateCleared(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()

	seller := createTestUser(t, env, "TEST RELIST SELLER")
	bidder := createTestUser(t, env, "TEST RELIST BIDDER")
	creditWallet(t, env, bidder.ID, decimal.NewFromInt(1000))

	auction := createExpiredTestAuction(t, env, seller.ID)
	if _, err := env.db.ExecContext(ctx, `UPDATE auctions SET end_time = $1 WHERE id = $2`, time.Now().Add(1*time.Hour), auction.ID); err != nil {
		t.Fatalf("failed to temporarily un-expire auction for bidding: %v", err)
	}
	if _, err := env.bidSvc.PlaceBid(ctx, auction.ID, bidder.ID, decimal.NewFromInt(150)); err != nil {
		t.Fatalf("PlaceBid failed: %v", err)
	}
	if _, err := env.db.ExecContext(ctx, `UPDATE auctions SET end_time = $1 WHERE id = $2`, time.Now().Add(-1*time.Minute), auction.ID); err != nil {
		t.Fatalf("failed to re-expire auction: %v", err)
	}
	if err := env.auctSvc.FinalizeExpiredAuction(ctx, auction.ID); err != nil {
		t.Fatalf("FinalizeExpiredAuction failed: %v", err)
	}

	// Confirm the win actually registered (status/winner/winning_bid_id/
	// payment_deadline all set) before relisting, so "cleared after relist"
	// is a meaningful assertion, not a vacuous one.
	var beforeRelist models.Auction
	if err := env.db.Get(&beforeRelist, `SELECT * FROM auctions WHERE id = $1`, auction.ID); err != nil {
		t.Fatalf("failed to read back auction before relist: %v", err)
	}
	if beforeRelist.WinnerID == nil || beforeRelist.WinningBidID == nil || beforeRelist.PaymentDeadline == nil {
		t.Fatalf("expected a full winner-finalization state (winner_id, winning_bid_id, payment_deadline all set) before relisting, got winner_id=%v winning_bid_id=%v payment_deadline=%v", beforeRelist.WinnerID, beforeRelist.WinningBidID, beforeRelist.PaymentDeadline)
	}
	winningsBefore, err := env.userSvc.ListMyWinnings(ctx, bidder.ID)
	if err != nil {
		t.Fatalf("ListMyWinnings failed: %v", err)
	}
	foundBefore := false
	for _, w := range winningsBefore {
		if w.ID == auction.ID {
			foundBefore = true
		}
	}
	if !foundBefore {
		t.Fatalf("expected the auction to be a real win before relisting")
	}

	newEndTime := time.Now().Add(72 * time.Hour)
	if err := env.auctSvc.RelistAuction(ctx, auction.ID, seller.ID, newEndTime); err != nil {
		t.Fatalf("RelistAuction failed: %v", err)
	}

	var final models.Auction
	if err := env.db.Get(&final, `SELECT * FROM auctions WHERE id = $1`, auction.ID); err != nil {
		t.Fatalf("failed to read back relisted auction: %v", err)
	}

	t.Run("status persists as pending", func(t *testing.T) {
		if final.Status != "pending" {
			t.Fatalf("expected status 'pending', got %q", final.Status)
		}
	})
	t.Run("winner_id cleared", func(t *testing.T) {
		if final.WinnerID != nil {
			t.Fatalf("expected winner_id cleared after relist, got %v", *final.WinnerID)
		}
	})
	t.Run("winning_bid_id cleared", func(t *testing.T) {
		if final.WinningBidID != nil {
			t.Fatalf("expected winning_bid_id cleared after relist, got %v", *final.WinningBidID)
		}
	})
	t.Run("payment_deadline cleared", func(t *testing.T) {
		if final.PaymentDeadline != nil {
			t.Fatalf("expected payment_deadline cleared after relist, got %v", *final.PaymentDeadline)
		}
	})
	t.Run("current_price reset to start_price", func(t *testing.T) {
		if !final.CurrentPrice.Equal(final.StartPrice) {
			t.Fatalf("expected current_price reset to start_price (%s), got %s", final.StartPrice.String(), final.CurrentPrice.String())
		}
	})
	t.Run("end_time updated", func(t *testing.T) {
		if final.EndTime.Unix() != newEndTime.Unix() {
			t.Fatalf("expected end_time updated to %v, got %v", newEndTime, final.EndTime)
		}
	})
	t.Run("old winner no longer sees it in My Winnings", func(t *testing.T) {
		winningsAfter, err := env.userSvc.ListMyWinnings(ctx, bidder.ID)
		if err != nil {
			t.Fatalf("ListMyWinnings failed: %v", err)
		}
		for _, w := range winningsAfter {
			if w.ID == auction.ID {
				t.Fatalf("REGRESSION: a relisted auction must no longer appear in the previous winner's My Winnings")
			}
		}
	})
	t.Run("old bid history remains (relist never deletes bids)", func(t *testing.T) {
		var bidCount int
		if err := env.db.Get(&bidCount, `SELECT COUNT(*) FROM bids WHERE auction_id = $1`, auction.ID); err != nil {
			t.Fatalf("failed to count bids: %v", err)
		}
		if bidCount == 0 {
			t.Fatalf("expected the old bid to remain as a historical record after relist, found 0 bids")
		}
	})
	t.Run("repeated/invalid relist is safe (a pending auction cannot be relisted again)", func(t *testing.T) {
		err := env.auctSvc.RelistAuction(ctx, auction.ID, seller.ID, time.Now().Add(96*time.Hour))
		if err == nil {
			t.Fatalf("expected relisting an already-'pending' auction to fail, got nil error")
		}
	})
}

// TestBuyNow_WinnerStateConsistency (Bug J, Section 26 -- was previously RED;
// now GREEN via TryBuyNowAtomically).
func TestBuyNow_WinnerStateConsistency(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()

	seller := createTestUser(t, env, "TEST BUYNOW SELLER")
	buyer := createTestUser(t, env, "TEST BUYNOW BUYER")
	otherUser := createTestUser(t, env, "TEST BUYNOW OTHER")
	creditWallet(t, env, buyer.ID, decimal.NewFromInt(5000))

	buyNowPrice := decimal.NewFromInt(500)
	a := createActiveBuyNowTestAuction(t, env, seller.ID, buyNowPrice)

	if _, err := env.auctSvc.BuyNow(ctx, a.ID, buyer.ID); err != nil {
		t.Fatalf("BuyNow failed: %v", err)
	}

	var final models.Auction
	if err := env.db.Get(&final, `SELECT * FROM auctions WHERE id = $1`, a.ID); err != nil {
		t.Fatalf("failed to read back auction: %v", err)
	}

	t.Run("status persists as ended", func(t *testing.T) {
		if final.Status != "ended" {
			t.Fatalf("expected status 'ended', got %q", final.Status)
		}
	})
	t.Run("winner_id persists as the buyer", func(t *testing.T) {
		if final.WinnerID == nil || *final.WinnerID != buyer.ID {
			t.Fatalf("expected winner_id %s, got %v", buyer.ID, final.WinnerID)
		}
	})
	t.Run("current_price persists as buy_now_price", func(t *testing.T) {
		if !final.CurrentPrice.Equal(buyNowPrice) {
			t.Fatalf("expected current_price %s, got %s", buyNowPrice.String(), final.CurrentPrice.String())
		}
	})
	t.Run("winning_bid_id remains null (no real bid row for an instant purchase)", func(t *testing.T) {
		if final.WinningBidID != nil {
			t.Fatalf("expected winning_bid_id NULL for a BuyNow purchase, got %v", *final.WinningBidID)
		}
	})
	t.Run("buyer sees the auction in My Winnings", func(t *testing.T) {
		winnings, err := env.userSvc.ListMyWinnings(ctx, buyer.ID)
		if err != nil {
			t.Fatalf("ListMyWinnings failed: %v", err)
		}
		found := false
		for _, w := range winnings {
			if w.ID == a.ID {
				found = true
			}
		}
		if !found {
			t.Fatalf("expected buyer's My Winnings to contain the BuyNow auction")
		}
	})
	t.Run("an unrelated user does not see it", func(t *testing.T) {
		winnings, err := env.userSvc.ListMyWinnings(ctx, otherUser.ID)
		if err != nil {
			t.Fatalf("ListMyWinnings failed: %v", err)
		}
		for _, w := range winnings {
			if w.ID == a.ID {
				t.Fatalf("SECURITY/CORRECTNESS REGRESSION: an unrelated user's My Winnings contained a BuyNow auction they did not buy")
			}
		}
	})
}

// TestBuyNow_DoubleBuyNow_ExactlyOneWinner (Bug J, Section 12).
func TestBuyNow_DoubleBuyNow_ExactlyOneWinner(t *testing.T) {
	env := setupEnv(t)

	seller := createTestUser(t, env, "TEST DBLBUYNOW SELLER")
	buyerA := createTestUser(t, env, "TEST DBLBUYNOW BUYER A")
	buyerB := createTestUser(t, env, "TEST DBLBUYNOW BUYER B")
	creditWallet(t, env, buyerA.ID, decimal.NewFromInt(5000))
	creditWallet(t, env, buyerB.ID, decimal.NewFromInt(5000))

	buyNowPrice := decimal.NewFromInt(500)
	a := createActiveBuyNowTestAuction(t, env, seller.ID, buyNowPrice)

	var wg sync.WaitGroup
	errs := make([]error, 2)
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, errs[0] = env.auctSvc.BuyNow(context.Background(), a.ID, buyerA.ID)
	}()
	go func() {
		defer wg.Done()
		_, errs[1] = env.auctSvc.BuyNow(context.Background(), a.ID, buyerB.ID)
	}()
	wg.Wait()

	successCount := 0
	for _, err := range errs {
		if err == nil {
			successCount++
		}
	}
	if successCount != 1 {
		t.Fatalf("expected exactly 1 successful BuyNow call out of 2 concurrent attempts, got %d", successCount)
	}

	var final models.Auction
	if err := env.db.Get(&final, `SELECT * FROM auctions WHERE id = $1`, a.ID); err != nil {
		t.Fatalf("failed to read back auction: %v", err)
	}
	if final.WinnerID == nil {
		t.Fatalf("expected a winner to be set")
	}
	if *final.WinnerID != buyerA.ID && *final.WinnerID != buyerB.ID {
		t.Fatalf("winner_id %v does not match either concurrent buyer", *final.WinnerID)
	}
	if final.Status != "ended" {
		t.Fatalf("expected status 'ended', got %q", final.Status)
	}
	if !final.CurrentPrice.Equal(buyNowPrice) {
		t.Fatalf("expected current_price %s, got %s", buyNowPrice.String(), final.CurrentPrice.String())
	}
}

// TestBuyNow_VsExpiration_Race (Bug J, Section 13).
func TestBuyNow_VsExpiration_Race(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()

	seller := createTestUser(t, env, "TEST BUYNOWRACE SELLER")
	buyer := createTestUser(t, env, "TEST BUYNOWRACE BUYER")
	bidder := createTestUser(t, env, "TEST BUYNOWRACE BIDDER")
	creditWallet(t, env, buyer.ID, decimal.NewFromInt(5000))
	creditWallet(t, env, bidder.ID, decimal.NewFromInt(1000))

	buyNowPrice := decimal.NewFromInt(500)
	a := createActiveBuyNowTestAuction(t, env, seller.ID, buyNowPrice)
	if _, err := env.bidSvc.PlaceBid(ctx, a.ID, bidder.ID, decimal.NewFromInt(150)); err != nil {
		t.Fatalf("PlaceBid failed: %v", err)
	}
	// Expire it right after bidding, so both BuyNow and FinalizeExpiredAuction
	// have a real, legitimate action to race over.
	if _, err := env.db.ExecContext(ctx, `UPDATE auctions SET end_time = $1 WHERE id = $2`, time.Now().Add(-1*time.Minute), a.ID); err != nil {
		t.Fatalf("failed to expire auction: %v", err)
	}

	var wg sync.WaitGroup
	var buyNowErr, finalizeErr error
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, buyNowErr = env.auctSvc.BuyNow(context.Background(), a.ID, buyer.ID)
	}()
	go func() {
		defer wg.Done()
		finalizeErr = env.auctSvc.FinalizeExpiredAuction(context.Background(), a.ID)
	}()
	wg.Wait()

	if finalizeErr != nil {
		t.Fatalf("FinalizeExpiredAuction returned an unexpected error: %v", finalizeErr)
	}

	var final models.Auction
	if err := env.db.Get(&final, `SELECT * FROM auctions WHERE id = $1`, a.ID); err != nil {
		t.Fatalf("failed to read back auction: %v", err)
	}
	if final.Status != "ended" {
		t.Fatalf("expected status 'ended' regardless of which side won, got %q", final.Status)
	}
	if final.WinnerID == nil {
		t.Fatalf("expected exactly one winner to be set")
	}

	if buyNowErr == nil {
		// BuyNow won the race.
		if *final.WinnerID != buyer.ID {
			t.Fatalf("BuyNow reported success but winner_id is not the buyer: %v", *final.WinnerID)
		}
		if !final.CurrentPrice.Equal(buyNowPrice) {
			t.Fatalf("expected current_price %s when BuyNow won, got %s", buyNowPrice.String(), final.CurrentPrice.String())
		}
	} else {
		// Expiration won -- the highest bidder is authoritative, BuyNow must
		// have safely failed (ErrConflict), never silently overwritten it.
		if *final.WinnerID != bidder.ID {
			t.Fatalf("expiration finalization won but winner_id is not the highest bidder: %v", *final.WinnerID)
		}
	}
}

// TestCancelAuction_VsBuyNow_Race (Bug J, Section 14).
func TestCancelAuction_VsBuyNow_Race(t *testing.T) {
	env := setupEnv(t)

	seller := createTestUser(t, env, "TEST CANCELBUYNOW SELLER")
	buyer := createTestUser(t, env, "TEST CANCELBUYNOW BUYER")
	creditWallet(t, env, buyer.ID, decimal.NewFromInt(5000))

	buyNowPrice := decimal.NewFromInt(500)
	a := createActiveBuyNowTestAuction(t, env, seller.ID, buyNowPrice)

	var wg sync.WaitGroup
	var cancelErr, buyNowErr error
	wg.Add(2)
	go func() {
		defer wg.Done()
		cancelErr = env.auctSvc.CancelAuction(context.Background(), a.ID, seller.ID, "race test cancel vs buynow")
	}()
	go func() {
		defer wg.Done()
		_, buyNowErr = env.auctSvc.BuyNow(context.Background(), a.ID, buyer.ID)
	}()
	wg.Wait()

	var final models.Auction
	if err := env.db.Get(&final, `SELECT * FROM auctions WHERE id = $1`, a.ID); err != nil {
		t.Fatalf("failed to read back auction: %v", err)
	}

	if final.Status != "canceled" && final.Status != "ended" {
		t.Fatalf("expected exactly one of 'canceled' or 'ended', got %q", final.Status)
	}

	if final.Status == "canceled" {
		if final.WinnerID != nil {
			t.Fatalf("INVALID STATE: status 'canceled' with a winner attached (%v)", *final.WinnerID)
		}
		if cancelErr != nil {
			t.Fatalf("status is 'canceled' but CancelAuction reported an error: %v", cancelErr)
		}
		if buyNowErr == nil {
			t.Fatalf("status is 'canceled' but BuyNow reported success (must not both succeed)")
		}
	} else {
		if final.WinnerID == nil || *final.WinnerID != buyer.ID {
			t.Fatalf("expected winner_id %s when BuyNow won, got %v", buyer.ID, final.WinnerID)
		}
		if buyNowErr != nil {
			t.Fatalf("status is 'ended' with the buyer as winner, but BuyNow reported an error: %v", buyNowErr)
		}
		if cancelErr == nil {
			t.Fatalf("status is 'ended' but CancelAuction reported success (must not both succeed)")
		}
	}
}

// TestGenericUpdate_PreservesWinnerAndStatus (Bug J, Section 20-22 --
// confirms the generic auctionRepo.Update, deliberately left UNCHANGED this
// round, cannot erase lifecycle state precisely because every real caller
// loads the full row via FindByID first and only mutates specific edit-form
// fields (see the audit's GENERIC_UPDATE_FULL_ENTITY_SEMANTICS finding)).
func TestGenericUpdate_PreservesWinnerAndStatus(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()

	seller := createTestUser(t, env, "TEST GENUPDATE SELLER")
	buyer := createTestUser(t, env, "TEST GENUPDATE BUYER")
	creditWallet(t, env, buyer.ID, decimal.NewFromInt(5000))

	buyNowPrice := decimal.NewFromInt(500)
	a := createActiveBuyNowTestAuction(t, env, seller.ID, buyNowPrice)
	if _, err := env.auctSvc.BuyNow(ctx, a.ID, buyer.ID); err != nil {
		t.Fatalf("BuyNow failed: %v", err)
	}

	t.Run("admin metadata edit after a win preserves winner_id/status", func(t *testing.T) {
		adminSvc := newTestAdminService(t, env)
		newTitle := "Admin-edited title after win " + uuid.New().String()[:6]
		if err := adminSvc.UpdateAuction(ctx, a.ID, services.UpdateAuctionInput{TitleAr: newTitle}); err != nil {
			t.Fatalf("AdminService.UpdateAuction failed: %v", err)
		}
		var reread models.Auction
		if err := env.db.Get(&reread, `SELECT * FROM auctions WHERE id = $1`, a.ID); err != nil {
			t.Fatalf("failed to read back auction: %v", err)
		}
		if reread.TitleAr != newTitle {
			t.Fatalf("expected the admin's title edit to persist, got %q", reread.TitleAr)
		}
		if reread.WinnerID == nil || *reread.WinnerID != buyer.ID {
			t.Fatalf("REGRESSION: admin metadata edit erased winner_id, got %v", reread.WinnerID)
		}
		if reread.Status != "ended" {
			t.Fatalf("REGRESSION: admin metadata edit erased lifecycle status, got %q", reread.Status)
		}
	})
}

// TestCloseExpiredAuctions_DelegatesToCanonicalFinalizer proves the
// production cron entry point (CloseExpiredAuctions, the 30-second ticker)
// itself -- not just FinalizeExpiredAuction called directly -- correctly
// persists winner_id end to end.
func TestCloseExpiredAuctions_DelegatesToCanonicalFinalizer(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()

	seller := createTestUser(t, env, "TEST CLOSEEXP SELLER")
	bidder := createTestUser(t, env, "TEST CLOSEEXP BIDDER")
	creditWallet(t, env, bidder.ID, decimal.NewFromInt(1000))

	auction := createExpiredTestAuction(t, env, seller.ID)
	if _, err := env.db.ExecContext(ctx, `UPDATE auctions SET end_time = $1 WHERE id = $2`, time.Now().Add(1*time.Hour), auction.ID); err != nil {
		t.Fatalf("failed to temporarily un-expire auction for bidding: %v", err)
	}
	topBid, err := env.bidSvc.PlaceBid(ctx, auction.ID, bidder.ID, decimal.NewFromInt(150))
	if err != nil {
		t.Fatalf("PlaceBid failed: %v", err)
	}
	if _, err := env.db.ExecContext(ctx, `UPDATE auctions SET end_time = $1 WHERE id = $2`, time.Now().Add(-1*time.Minute), auction.ID); err != nil {
		t.Fatalf("failed to re-expire auction: %v", err)
	}

	if err := env.auctSvc.CloseExpiredAuctions(ctx); err != nil {
		t.Fatalf("CloseExpiredAuctions failed: %v", err)
	}

	var final models.Auction
	if err := env.db.Get(&final, `SELECT * FROM auctions WHERE id = $1`, auction.ID); err != nil {
		t.Fatalf("failed to read back auction: %v", err)
	}
	if final.WinnerID == nil || *final.WinnerID != bidder.ID {
		t.Fatalf("THE ORIGINAL CLIENT BUG: expected CloseExpiredAuctions (the actual 30-second production cron path) to persist winner_id %s via the canonical finalizer, got %v", bidder.ID, final.WinnerID)
	}
	if final.WinningBidID == nil || *final.WinningBidID != topBid.ID {
		t.Fatalf("expected winning_bid_id %s, got %v", topBid.ID, final.WinningBidID)
	}
}

// Bug K defense-in-depth (Section 14 of the diagnosis brief): proves the
// backend's own expiry guard in BidService.PlaceBid (bid_service.go,
// `if time.Now().After(auction.EndTime) { return apperr.ErrAuctionEnded }`,
// checked inside the SAME transaction as loading the auction row) actually
// rejects a bid placed after end_time has passed but BEFORE the scheduler
// has run to flip status to 'ended' -- the exact window the mobile-side CTA
// fix (auction_details_page.dart canBid) is a client-side mirror of, not a
// replacement for. This guard already existed prior to this round; this
// test is new, confirming behavior that was previously unverified.
func TestPlaceBid_RejectedAfterEndTime_BeforeSchedulerRuns(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()

	seller := createTestUser(t, env, "TEST BID-AFTER-EXPIRY SELLER")
	bidder := createTestUser(t, env, "TEST BID-AFTER-EXPIRY BIDDER")
	creditWallet(t, env, bidder.ID, decimal.NewFromInt(1000))

	// Auction still reports status='active' in the DB (scheduler has not
	// run yet) but its end_time has already passed -- the exact race window
	// this guard protects.
	auction := createExpiredTestAuction(t, env, seller.ID)

	var bidCountBefore int
	if err := env.db.GetContext(ctx, &bidCountBefore, `SELECT COUNT(*) FROM bids WHERE auction_id = $1`, auction.ID); err != nil {
		t.Fatalf("failed to count bids before attempt: %v", err)
	}

	_, err := env.bidSvc.PlaceBid(ctx, auction.ID, bidder.ID, decimal.NewFromInt(150))

	t.Run("bid rejected", func(t *testing.T) {
		if err != apperr.ErrAuctionEnded {
			t.Fatalf("expected apperr.ErrAuctionEnded, got %v", err)
		}
	})

	var final models.Auction
	if err := env.db.Get(&final, `SELECT * FROM auctions WHERE id = $1`, auction.ID); err != nil {
		t.Fatalf("failed to read back auction: %v", err)
	}

	t.Run("no bid row created", func(t *testing.T) {
		var bidCountAfter int
		if err := env.db.GetContext(ctx, &bidCountAfter, `SELECT COUNT(*) FROM bids WHERE auction_id = $1`, auction.ID); err != nil {
			t.Fatalf("failed to count bids after attempt: %v", err)
		}
		if bidCountAfter != bidCountBefore {
			t.Fatalf("expected bid count to remain %d, got %d -- a bid row was created despite rejection", bidCountBefore, bidCountAfter)
		}
	})
	t.Run("current_price unchanged", func(t *testing.T) {
		if !final.CurrentPrice.Equal(auction.CurrentPrice) {
			t.Fatalf("expected current_price to remain %s, got %s", auction.CurrentPrice.String(), final.CurrentPrice.String())
		}
	})
	t.Run("bidder_count unchanged", func(t *testing.T) {
		if final.BidderCount != auction.BidderCount {
			t.Fatalf("expected bidder_count to remain %d, got %d", auction.BidderCount, final.BidderCount)
		}
	})
	t.Run("no auction_won or bid notification created for this attempt", func(t *testing.T) {
		count := countAuctionWonNotifications(t, env, bidder.ID, auction.ID)
		if count != 0 {
			t.Fatalf("expected 0 notifications from a rejected bid attempt, got %d", count)
		}
	})
}

// Bug M diagnosis: reproduces the exact real-device symptom -- bid_count > 0
// on the auction summary, but the bid-history list renders empty. The
// suspected root cause is that BidHistoryEntry.BidderName/BidderPhone are
// non-nullable Go strings, but the repository query
// (COALESCE(b.bidder_name, u.full_name)) can legitimately produce SQL NULL
// when BOTH the bid row's own denormalized bidder_name/bidder_phone AND the
// joined user's full_name/phone are NULL -- sqlx's Scan then fails
// ("converting NULL to string is unsupported"), GetHistory returns an
// error, and the mobile Riverpod provider's catch-all silently swallows it
// and returns an empty list (auction_provider_api.dart AuctionHistoryApi.build,
// `catch (e) { ... return []; }`), producing exactly the observed
// contradiction: bid_count=1 but "لا توجد مزايدات حتى الآن".
func TestGetHistory_NullBidderNameAndUserFullName_DoesNotFailSilently(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()

	seller := createTestUser(t, env, "TEST BIDHISTORY SELLER")
	bidder := createTestUser(t, env, "TEST BIDHISTORY BIDDER")
	creditWallet(t, env, bidder.ID, decimal.NewFromInt(1000))

	// Null out the bidder's own full_name/phone AFTER creation (Register
	// requires a phone; simulate the real-device condition where the
	// joined user's full_name legitimately has no value) -- this mirrors
	// the exact staging DB state found for the reported auction (both
	// bids.bidder_name and users.full_name NULL for that bidder).
	if _, err := env.db.ExecContext(ctx, `UPDATE users SET full_name = NULL WHERE id = $1`, bidder.ID); err != nil {
		t.Fatalf("failed to null out fixture user full_name: %v", err)
	}

	auction := createTestAuctionInsured(t, env, seller.ID, "MR", "MRU")

	bid, err := env.bidSvc.PlaceBid(ctx, auction.ID, bidder.ID, decimal.NewFromInt(150))
	if err != nil {
		t.Fatalf("PlaceBid failed: %v", err)
	}

	// Confirm the DB state matches the real-device condition: bid row's own
	// bidder_name is NULL (never set by PlaceBid), and the joined user's
	// full_name is now also NULL.
	var bidderNameNull, fullNameNull bool
	if err := env.db.QueryRowContext(ctx, `SELECT bidder_name IS NULL FROM bids WHERE id = $1`, bid.ID).Scan(&bidderNameNull); err != nil {
		t.Fatalf("failed to check bids.bidder_name nullness: %v", err)
	}
	if err := env.db.QueryRowContext(ctx, `SELECT full_name IS NULL FROM users WHERE id = $1`, bidder.ID).Scan(&fullNameNull); err != nil {
		t.Fatalf("failed to check users.full_name nullness: %v", err)
	}
	if !bidderNameNull || !fullNameNull {
		t.Fatalf("fixture setup did not reproduce the real-device NULL condition: bids.bidder_name NULL=%v, users.full_name NULL=%v", bidderNameNull, fullNameNull)
	}

	history, err := env.bidSvc.GetHistory(ctx, auction.ID)
	if err != nil {
		t.Fatalf("BUG M REPRODUCED: GetHistory failed with a real bid present (bid_count=1) due to a NULL bidder_name/full_name scan error: %v", err)
	}
	if len(history) != 1 {
		t.Fatalf("expected exactly 1 bid history entry, got %d", len(history))
	}
	if history[0].ID != bid.ID {
		t.Fatalf("expected the real bid %s in history, got %s", bid.ID, history[0].ID)
	}
}

// Bug M: Section 13 required coverage -- ordering rule, multiple real
// bidders, and cross-auction isolation for GetHistory/FindHistoryByAuction.
func TestGetHistory_MultipleBidders_AllRenderedInCorrectOrder(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()

	seller := createTestUser(t, env, "TEST HISTORY ORDER SELLER")
	bidderA := createTestUser(t, env, "TEST HISTORY ORDER BIDDER A")
	bidderB := createTestUser(t, env, "TEST HISTORY ORDER BIDDER B")
	bidderC := createTestUser(t, env, "TEST HISTORY ORDER BIDDER C")
	creditWallet(t, env, bidderA.ID, decimal.NewFromInt(1000))
	creditWallet(t, env, bidderB.ID, decimal.NewFromInt(1000))
	creditWallet(t, env, bidderC.ID, decimal.NewFromInt(1000))

	auction := createTestAuctionInsured(t, env, seller.ID, "MR", "MRU")

	bidA, err := env.bidSvc.PlaceBid(ctx, auction.ID, bidderA.ID, decimal.NewFromInt(150))
	if err != nil {
		t.Fatalf("bidderA PlaceBid failed: %v", err)
	}
	bidB, err := env.bidSvc.PlaceBid(ctx, auction.ID, bidderB.ID, decimal.NewFromInt(200))
	if err != nil {
		t.Fatalf("bidderB PlaceBid failed: %v", err)
	}
	bidC, err := env.bidSvc.PlaceBid(ctx, auction.ID, bidderC.ID, decimal.NewFromInt(250))
	if err != nil {
		t.Fatalf("bidderC PlaceBid failed: %v", err)
	}

	history, err := env.bidSvc.GetHistory(ctx, auction.ID)
	if err != nil {
		t.Fatalf("GetHistory failed: %v", err)
	}

	t.Run("all bidders rendered", func(t *testing.T) {
		if len(history) != 3 {
			t.Fatalf("expected 3 bid history entries, got %d", len(history))
		}
	})
	t.Run("ordering rule: highest amount first (mirrors mobile's isWinner = index == 0)", func(t *testing.T) {
		if history[0].ID != bidC.ID || history[0].Amount.String() != "250" {
			t.Fatalf("expected highest bid (bidderC, 250) first, got id=%s amount=%s", history[0].ID, history[0].Amount.String())
		}
		if history[1].ID != bidB.ID {
			t.Fatalf("expected second-highest bid (bidderB, 200) second, got id=%s", history[1].ID)
		}
		if history[2].ID != bidA.ID {
			t.Fatalf("expected lowest bid (bidderA, 150) last, got id=%s", history[2].ID)
		}
	})
}

func TestGetHistory_WrongAuctionID_NotMixed(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()

	seller := createTestUser(t, env, "TEST HISTORY ISOLATION SELLER")
	bidder := createTestUser(t, env, "TEST HISTORY ISOLATION BIDDER")
	creditWallet(t, env, bidder.ID, decimal.NewFromInt(1000))

	auctionOne := createTestAuctionInsured(t, env, seller.ID, "MR", "MRU")
	auctionTwo := createTestAuctionInsured(t, env, seller.ID, "MR", "MRU")

	if _, err := env.bidSvc.PlaceBid(ctx, auctionOne.ID, bidder.ID, decimal.NewFromInt(150)); err != nil {
		t.Fatalf("PlaceBid on auctionOne failed: %v", err)
	}

	historyTwo, err := env.bidSvc.GetHistory(ctx, auctionTwo.ID)
	if err != nil {
		t.Fatalf("GetHistory for auctionTwo failed: %v", err)
	}
	if len(historyTwo) != 0 {
		t.Fatalf("expected auctionTwo (no bids) to have empty history, got %d entries -- a bid from a different auction leaked in", len(historyTwo))
	}

	historyOne, err := env.bidSvc.GetHistory(ctx, auctionOne.ID)
	if err != nil {
		t.Fatalf("GetHistory for auctionOne failed: %v", err)
	}
	if len(historyOne) != 1 {
		t.Fatalf("expected auctionOne to have exactly 1 bid, got %d", len(historyOne))
	}
}

func TestGetHistory_ZeroBids_ReturnsEmptyNotError(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()

	seller := createTestUser(t, env, "TEST HISTORY EMPTY SELLER")
	auction := createTestAuctionInsured(t, env, seller.ID, "MR", "MRU")

	history, err := env.bidSvc.GetHistory(ctx, auction.ID)
	if err != nil {
		t.Fatalf("GetHistory failed for a zero-bid auction: %v", err)
	}
	if len(history) != 0 {
		t.Fatalf("expected empty history for a zero-bid auction, got %d entries", len(history))
	}
}
