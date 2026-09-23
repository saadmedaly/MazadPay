//go:build integration

package integrationtest

import (
	"context"
	"sync"
	"testing"

	apperr "github.com/mazadpay/backend/internal/errors"
	"github.com/shopspring/decimal"
)

// Financial Audit Phase 7 (auction holds/insurance): the existing
// TestPlaceBid_ConcurrentSameUserBids_OnlyOneCommits (integration_test.go)
// already proves exactly one bid row survives N concurrent same-user bid
// attempts, via TryClaimBidTurn's atomic UPDATE...WHERE...RETURNING. That
// test does not check wallet/hold state, only bid-row count -- this file
// closes that gap by asserting the SAME race produces exactly one active
// wallet_hold and exactly one insurance freeze, never an orphaned or
// duplicate hold, and never more than one freeze's worth of frozen_amount.
func TestPlaceBid_ConcurrentSameUserBids_ExactlyOneHoldAndFreeze(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST HOLD RACE SELLER")
	bidder := createTestUser(t, env, "TEST HOLD RACE BIDDER")
	auction := createTestAuction(t, env, seller.ID, "MR", "MRU") // InsuranceAmount=20, InsurancePolicy=required
	creditWallet(t, env, bidder.ID, decimal.NewFromInt(10000))

	const n = 8
	results := make(chan error, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(amount int64) {
			defer wg.Done()
			_, err := env.bidSvc.PlaceBid(ctx, auction.ID, bidder.ID, decimal.NewFromInt(110+amount))
			results <- err
		}(int64(i))
	}
	wg.Wait()
	close(results)

	successCount := 0
	for err := range results {
		if err == nil {
			successCount++
		} else if err != apperr.ErrDuplicateBidder && err != apperr.ErrBidConflict {
			t.Logf("unexpected error from concurrent attempt (acceptable if a transient conflict): %v", err)
		}
	}
	if successCount != 1 {
		t.Fatalf("expected exactly 1 of %d concurrent same-user bid attempts to succeed, got %d", n, successCount)
	}

	var activeHoldCount int
	if err := env.db.GetContext(ctx, &activeHoldCount,
		`SELECT COUNT(*) FROM wallet_holds WHERE user_id = $1 AND auction_id = $2 AND status = 'active'`,
		bidder.ID, auction.ID); err != nil {
		t.Fatalf("failed to count active holds: %v", err)
	}
	if activeHoldCount != 1 {
		t.Fatalf("CRITICAL: expected exactly 1 active wallet_hold after %d concurrent bid attempts, got %d -- orphaned/duplicate hold", n, activeHoldCount)
	}

	var totalHoldCount int
	if err := env.db.GetContext(ctx, &totalHoldCount,
		`SELECT COUNT(*) FROM wallet_holds WHERE user_id = $1 AND auction_id = $2`,
		bidder.ID, auction.ID); err != nil {
		t.Fatalf("failed to count total holds: %v", err)
	}
	if totalHoldCount != 1 {
		t.Fatalf("CRITICAL: expected exactly 1 wallet_hold row total (active or otherwise) for this bidder/auction, got %d -- a failed concurrent attempt must never leave a stray hold row", totalHoldCount)
	}

	wallet, err := env.walletRepo.GetByUserID(ctx, bidder.ID)
	if err != nil {
		t.Fatalf("failed to read wallet: %v", err)
	}
	if !wallet.FrozenAmount.Equal(auction.InsuranceAmount) {
		t.Fatalf("CRITICAL: expected frozen_amount to equal exactly the auction's insurance_amount (%s) after %d concurrent bid attempts, got %s -- must never freeze more than once",
			auction.InsuranceAmount, n, wallet.FrozenAmount)
	}
	if !wallet.Balance.Add(wallet.FrozenAmount).Equal(decimal.NewFromInt(10000)) {
		t.Fatalf("CONSERVATION VIOLATION: balance+frozen (%s+%s) must remain exactly 10000 after the race, got %s",
			wallet.Balance, wallet.FrozenAmount, wallet.Balance.Add(wallet.FrozenAmount))
	}
}
