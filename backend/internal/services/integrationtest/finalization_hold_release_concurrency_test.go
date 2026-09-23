//go:build integration

package integrationtest

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

// Financial Audit Phase 9 (finalization/settlement idempotency): the
// existing TestFinalizeExpiredAuction_ConcurrentCallers_ExactlyOneWinner
// (auction_finalization_test.go) uses a not_required-insurance fixture and
// checks only status/winner_id -- it never exercises hold release under
// concurrent finalization. This test uses an insurance-required auction
// with a losing bidder and fires concurrent FinalizeExpiredAuction calls,
// asserting the loser's hold is released exactly once (never double-
// released, never left orphaned) and wallet state is conserved.
func TestFinalizeExpiredAuction_ConcurrentCallers_HoldReleasedExactlyOnce(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()

	seller := createTestUser(t, env, "TEST FINALIZE RACE SELLER")
	loser := createTestUser(t, env, "TEST FINALIZE RACE LOSER")
	winner := createTestUser(t, env, "TEST FINALIZE RACE WINNER")
	creditWallet(t, env, loser.ID, decimal.NewFromInt(1000))
	creditWallet(t, env, winner.ID, decimal.NewFromInt(1000))

	auction := createExpiredInsuranceRequiredAuction(t, env, seller.ID)
	if _, err := env.db.ExecContext(ctx, `UPDATE auctions SET end_time = $1 WHERE id = $2`, time.Now().Add(1*time.Hour), auction.ID); err != nil {
		t.Fatalf("failed to temporarily un-expire auction for bidding: %v", err)
	}
	if _, err := env.bidSvc.PlaceBid(ctx, auction.ID, loser.ID, decimal.NewFromInt(150)); err != nil {
		t.Fatalf("loser PlaceBid failed: %v", err)
	}
	if _, err := env.bidSvc.PlaceBid(ctx, auction.ID, winner.ID, decimal.NewFromInt(200)); err != nil {
		t.Fatalf("winner PlaceBid failed: %v", err)
	}
	if _, err := env.db.ExecContext(ctx, `UPDATE auctions SET end_time = $1 WHERE id = $2`, time.Now().Add(-1*time.Minute), auction.ID); err != nil {
		t.Fatalf("failed to re-expire auction: %v", err)
	}

	var wg sync.WaitGroup
	errs := make([]error, 4)
	wg.Add(4)
	for i := 0; i < 4; i++ {
		go func(idx int) {
			defer wg.Done()
			errs[idx] = env.auctSvc.FinalizeExpiredAuction(context.Background(), auction.ID)
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("concurrent FinalizeExpiredAuction call %d returned an error (expected a safe no-op): %v", i, err)
		}
	}

	loserFinal, err := env.walletRepo.GetByUserID(ctx, loser.ID)
	if err != nil {
		t.Fatalf("failed to read loser wallet: %v", err)
	}
	if !loserFinal.Balance.Equal(decimal.NewFromInt(1000)) || !loserFinal.FrozenAmount.IsZero() {
		t.Fatalf("CRITICAL: expected loser wallet exactly (1000, 0) after 4 concurrent finalization calls, got (%s, %s) -- hold must be released exactly once, never double-released into extra balance",
			loserFinal.Balance, loserFinal.FrozenAmount)
	}

	if status := holdStatus(t, env, loser.ID, auction.ID); status != "released" {
		t.Fatalf("expected loser's hold released exactly once, got status %q", status)
	}

	var loserHoldRowCount int
	if err := env.db.GetContext(ctx, &loserHoldRowCount,
		`SELECT COUNT(*) FROM wallet_holds WHERE user_id = $1 AND auction_id = $2`, loser.ID, auction.ID); err != nil {
		t.Fatalf("failed to count loser hold rows: %v", err)
	}
	if loserHoldRowCount != 1 {
		t.Fatalf("CRITICAL: expected exactly 1 wallet_hold row for the loser (released, never duplicated by concurrent finalization), got %d", loserHoldRowCount)
	}

	winnerFinal, err := env.walletRepo.GetByUserID(ctx, winner.ID)
	if err != nil {
		t.Fatalf("failed to read winner wallet: %v", err)
	}
	if !winnerFinal.Balance.Add(winnerFinal.FrozenAmount).Equal(decimal.NewFromInt(1000)) {
		t.Fatalf("CONSERVATION VIOLATION: winner balance+frozen must remain exactly 1000 after concurrent finalization, got %s+%s",
			winnerFinal.Balance, winnerFinal.FrozenAmount)
	}
}
