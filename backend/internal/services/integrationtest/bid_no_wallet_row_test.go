//go:build integration

package integrationtest

import (
	"context"
	"errors"
	"testing"

	apperr "github.com/mazadpay/backend/internal/errors"
	"github.com/shopspring/decimal"
)

// TestPlaceBid_NoWalletRow_ReturnsInsufficientForInsurance covers a Security
// Gate finding: an eligible, correctly market-matched, non-owner bidder who
// has never made a deposit (so has NO row in wallets at all, as opposed to a
// funded row with a 0 balance) was getting apperr.ErrNotFound from
// WalletRepository.FindForUpdate -- which MapError renders as a misleading
// generic "Resource not found" 404, indistinguishable from the auction
// itself not existing. Live-reproduced against a real Validation auction:
// the bidder could GET the same auction fine (200), proving the auction
// unambiguously existed, yet POSTing a bid returned 404.
//
// Every pre-existing PlaceBid test in this package calls creditWallet
// first, which (via WalletRepository.GetByUserID's lazy
// INSERT ... ON CONFLICT DO NOTHING) always ensures a wallet row exists
// before the test bidder ever places a bid -- so none of them ever
// exercised the genuinely-no-row path. This test intentionally never
// credits (or otherwise touches) the bidder's wallet, reproducing the exact
// gap.
//
// Having zero wallet rows is financially identical to having a funded row
// with a 0 balance -- both mean "cannot cover the insurance freeze" -- so
// the fix (bid_service.go PlaceBid) treats a FindForUpdate ErrNotFound the
// same way the balance check two branches below already treats an
// insufficient balance: apperr.ErrInsufficientForInsurance, not
// apperr.ErrNotFound.
func TestPlaceBid_NoWalletRow_ReturnsInsufficientForInsurance(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()

	seller := createTestUser(t, env, "TEST NO WALLET ROW SELLER")
	bidder := createTestUser(t, env, "TEST NO WALLET ROW BIDDER")
	// Deliberately no creditWallet(t, env, bidder.ID, ...) call -- this
	// bidder must have zero rows in the wallets table.

	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")
	if !auction.InsuranceRequired() {
		t.Fatalf("fixture auction must require insurance for this test to be meaningful")
	}

	var walletRowCount int
	if err := env.db.GetContext(ctx, &walletRowCount, `SELECT count(*) FROM wallets WHERE user_id = $1`, bidder.ID); err != nil {
		t.Fatalf("failed to check wallet row count: %v", err)
	}
	if walletRowCount != 0 {
		t.Fatalf("test setup invariant violated: bidder already has %d wallet row(s), expected 0", walletRowCount)
	}

	bidAmount := auction.CurrentPrice.Add(auction.MinIncrement)
	_, err := env.bidSvc.PlaceBid(ctx, auction.ID, bidder.ID, bidAmount)

	if err == nil {
		t.Fatalf("expected PlaceBid to fail for a bidder with no wallet row, got success")
	}
	if errors.Is(err, apperr.ErrNotFound) {
		t.Fatalf("REGRESSION: PlaceBid returned ErrNotFound (renders as misleading \"Resource not found\" 404) for a bidder with no wallet row -- the auction demonstrably exists and is readable, this must be ErrInsufficientForInsurance instead")
	}
	if !errors.Is(err, apperr.ErrInsufficientForInsurance) {
		t.Fatalf("expected ErrInsufficientForInsurance, got: %v", err)
	}

	// Sanity: a bidder who DOES have a funded-but-empty wallet row gets the
	// exact same error -- proving the fix makes the no-row case
	// indistinguishable from the already-correct empty-row case, rather than
	// inventing a new, different error path.
	fundedBidder := createTestUser(t, env, "TEST NO WALLET ROW FUNDED-EMPTY BIDDER")
	creditWallet(t, env, fundedBidder.ID, decimal.NewFromInt(0))
	_, err2 := env.bidSvc.PlaceBid(ctx, auction.ID, fundedBidder.ID, bidAmount)
	if !errors.Is(err2, apperr.ErrInsufficientForInsurance) {
		t.Fatalf("sanity check failed: a funded-but-empty wallet should also get ErrInsufficientForInsurance, got: %v", err2)
	}
}
