//go:build integration

package integrationtest

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

// Financial Audit Phase 3 (money conservation invariants): rather than
// re-proving each individual mechanism in isolation (already covered
// extensively elsewhere -- see winner_insurance_refund_test.go,
// auction_finalization_test.go, admin_credit_deposit_double_credit_test.go),
// this file ties the FULL lifecycle together in one end-to-end scenario and
// asserts the conservation invariant no single-feature test checks: at every
// step, (balance + frozen_amount) for each participant plus the sum of all
// active holds must account for exactly the money that entered the system,
// with nothing created or destroyed.

// TestMoneyConservation_DepositBidLoseWin_FullLifecycle walks one seller and
// two bidders through: deposit -> approval -> bid -> outbid -> finalization,
// and checks wallet+hold state is exactly what the sequence of operations
// implies at each checkpoint -- no drift, no orphaned funds, no double
// application.
func TestMoneyConservation_DepositBidLoseWin_FullLifecycle(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()

	seller := createTestUser(t, env, "TEST CONSERVATION SELLER")
	loser := createTestUser(t, env, "TEST CONSERVATION LOSER")
	winner := createTestUser(t, env, "TEST CONSERVATION WINNER")

	walletSvc := newWalletSvc(env)
	adminSvc := newTestAdminService(t, env)

	// Checkpoint 1: both bidders deposit and get approved -- exactly their
	// deposited amount must land, nothing more, nothing less.
	admin := createTestAdmin(t, env, "TEST CONSERVATION ADMIN")
	loserDeposit, err := walletSvc.InitiateDeposit(ctx, loser.ID, decimal.NewFromInt(1000), "bankily", "", "", nil)
	if err != nil {
		t.Fatalf("loser InitiateDeposit failed: %v", err)
	}
	winnerDeposit, err := walletSvc.InitiateDeposit(ctx, winner.ID, decimal.NewFromInt(1000), "bankily", "", "", nil)
	if err != nil {
		t.Fatalf("winner InitiateDeposit failed: %v", err)
	}
	if err := adminSvc.ValidateTransaction(ctx, loserDeposit.ID, true, "", admin.ID, ""); err != nil {
		t.Fatalf("loser deposit approval failed: %v", err)
	}
	if err := adminSvc.ValidateTransaction(ctx, winnerDeposit.ID, true, "", admin.ID, ""); err != nil {
		t.Fatalf("winner deposit approval failed: %v", err)
	}

	loserWallet, err := env.walletRepo.GetByUserID(ctx, loser.ID)
	if err != nil {
		t.Fatalf("failed to read loser wallet: %v", err)
	}
	winnerWallet, err := env.walletRepo.GetByUserID(ctx, winner.ID)
	if err != nil {
		t.Fatalf("failed to read winner wallet: %v", err)
	}
	if !loserWallet.Balance.Equal(decimal.NewFromInt(1000)) || !loserWallet.FrozenAmount.IsZero() {
		t.Fatalf("checkpoint 1: loser wallet must be exactly (1000, 0), got (%s, %s)", loserWallet.Balance, loserWallet.FrozenAmount)
	}
	if !winnerWallet.Balance.Equal(decimal.NewFromInt(1000)) || !winnerWallet.FrozenAmount.IsZero() {
		t.Fatalf("checkpoint 1: winner wallet must be exactly (1000, 0), got (%s, %s)", winnerWallet.Balance, winnerWallet.FrozenAmount)
	}

	// Checkpoint 2: both bid (insurance required, amount=20 per
	// createExpiredInsuranceRequiredAuction's fixture). Each bid freezes
	// exactly the insurance amount -- balance and frozen must move by
	// exactly 20 each, never more, and the sum balance+frozen per user must
	// stay exactly 1000 (nothing created or destroyed by freezing).
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

	loserAfterBid, err := env.walletRepo.GetByUserID(ctx, loser.ID)
	if err != nil {
		t.Fatalf("failed to read loser wallet after bid: %v", err)
	}
	winnerAfterBid, err := env.walletRepo.GetByUserID(ctx, winner.ID)
	if err != nil {
		t.Fatalf("failed to read winner wallet after bid: %v", err)
	}
	insuranceAmount := auction.InsuranceAmount
	if !loserAfterBid.Balance.Add(loserAfterBid.FrozenAmount).Equal(decimal.NewFromInt(1000)) {
		t.Fatalf("checkpoint 2 CONSERVATION VIOLATION: loser's balance+frozen (%s+%s=%s) must remain exactly 1000 after freezing insurance",
			loserAfterBid.Balance, loserAfterBid.FrozenAmount, loserAfterBid.Balance.Add(loserAfterBid.FrozenAmount))
	}
	if !loserAfterBid.FrozenAmount.Equal(insuranceAmount) {
		t.Fatalf("checkpoint 2: expected loser frozen_amount to equal the auction's insurance_amount (%s), got %s", insuranceAmount, loserAfterBid.FrozenAmount)
	}
	if !winnerAfterBid.Balance.Add(winnerAfterBid.FrozenAmount).Equal(decimal.NewFromInt(1000)) {
		t.Fatalf("checkpoint 2 CONSERVATION VIOLATION: winner's balance+frozen must remain exactly 1000, got %s+%s", winnerAfterBid.Balance, winnerAfterBid.FrozenAmount)
	}

	// Checkpoint 3: finalize. Loser's hold must release (balance+frozen
	// returns to exactly 1000, frozen back to 0). Winner's hold stays
	// active (documented gap: no automatic seller settlement exists yet --
	// this is NOT a bug in the tested code, it's the current, deliberately
	// unimplemented state confirmed by
	// TestRefundWinnerInsurance_NonWinnerBidderHoldAlreadyReleased) --
	// winner's balance+frozen must STILL sum to exactly 1000 (the freeze is
	// a hold, not a loss, until/unless a real settlement mechanism is
	// eventually implemented).
	if _, err := env.db.ExecContext(ctx, `UPDATE auctions SET end_time = $1 WHERE id = $2`, time.Now().Add(-1*time.Minute), auction.ID); err != nil {
		t.Fatalf("failed to re-expire auction: %v", err)
	}
	if err := env.auctSvc.FinalizeExpiredAuction(ctx, auction.ID); err != nil {
		t.Fatalf("FinalizeExpiredAuction failed: %v", err)
	}

	loserFinal, err := env.walletRepo.GetByUserID(ctx, loser.ID)
	if err != nil {
		t.Fatalf("failed to read final loser wallet: %v", err)
	}
	winnerFinal, err := env.walletRepo.GetByUserID(ctx, winner.ID)
	if err != nil {
		t.Fatalf("failed to read final winner wallet: %v", err)
	}

	if !loserFinal.Balance.Equal(decimal.NewFromInt(1000)) || !loserFinal.FrozenAmount.IsZero() {
		t.Fatalf("checkpoint 3 CONSERVATION VIOLATION: loser's hold release must restore exactly (1000, 0), got (%s, %s)", loserFinal.Balance, loserFinal.FrozenAmount)
	}
	if !winnerFinal.Balance.Add(winnerFinal.FrozenAmount).Equal(decimal.NewFromInt(1000)) {
		t.Fatalf("checkpoint 3 CONSERVATION VIOLATION: winner's balance+frozen must still sum to exactly 1000 post-finalization (hold retained, not lost), got %s+%s", winnerFinal.Balance, winnerFinal.FrozenAmount)
	}
	if !winnerFinal.FrozenAmount.Equal(insuranceAmount) {
		t.Fatalf("checkpoint 3: winner's frozen_amount should remain exactly the insurance amount (%s) -- documented gap, no settlement path exists -- got %s", insuranceAmount, winnerFinal.FrozenAmount)
	}

	if status := holdStatus(t, env, loser.ID, auction.ID); status != "released" {
		t.Fatalf("expected loser's hold released, got %q", status)
	}
	if status := holdStatus(t, env, winner.ID, auction.ID); status != "active" {
		t.Fatalf("expected winner's hold to remain active (documented settlement gap), got %q", status)
	}
}

// TestMoneyConservation_RejectedDepositLeavesNoResidue is Phase 3's
// "rejected operations leave no financial residue" invariant, isolated: a
// deposit that's rejected must leave the wallet completely untouched --
// zero balance change, not even a transient one.
func TestMoneyConservation_RejectedDepositLeavesNoResidue(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST CONSERVATION REJECT USER")
	admin := createTestAdmin(t, env, "TEST CONSERVATION REJECT ADMIN")

	walletSvc := newWalletSvc(env)
	adminSvc := newTestAdminService(t, env)

	deposit, err := walletSvc.InitiateDeposit(ctx, user.ID, decimal.NewFromInt(2500), "bankily", "", "", nil)
	if err != nil {
		t.Fatalf("InitiateDeposit failed: %v", err)
	}

	if err := adminSvc.ValidateTransaction(ctx, deposit.ID, false, "invalid receipt", admin.ID, ""); err != nil {
		t.Fatalf("ValidateTransaction (reject) failed: %v", err)
	}

	wallet, err := env.walletRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to read wallet: %v", err)
	}
	if !wallet.Balance.IsZero() || !wallet.FrozenAmount.IsZero() {
		t.Fatalf("CONSERVATION VIOLATION: a rejected deposit left residue -- balance=%s frozen=%s, expected (0, 0)", wallet.Balance, wallet.FrozenAmount)
	}
}

// TestMoneyConservation_RejectedWithdrawalReleasesExactAmount is the
// withdrawal-side "rejected operations leave no residue" invariant: a
// rejected withdrawal must release EXACTLY the frozen amount back to
// balance -- no more, no less, and frozen_amount must return to exactly 0.
func TestMoneyConservation_RejectedWithdrawalReleasesExactAmount(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST CONSERVATION W-REJECT USER")
	admin := createTestAdmin(t, env, "TEST CONSERVATION W-REJECT ADMIN")
	creditWallet(t, env, user.ID, decimal.NewFromInt(800))

	walletSvc := newWalletSvc(env)
	adminSvc := newTestAdminService(t, env)

	withdraw, err := walletSvc.RequestWithdraw(ctx, user.ID, decimal.NewFromInt(300), "bankily", "22247601175")
	if err != nil {
		t.Fatalf("RequestWithdraw failed: %v", err)
	}

	afterFreeze, err := env.walletRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to read wallet after freeze: %v", err)
	}
	if !afterFreeze.Balance.Equal(decimal.NewFromInt(500)) || !afterFreeze.FrozenAmount.Equal(decimal.NewFromInt(300)) {
		t.Fatalf("expected (500, 300) after freeze, got (%s, %s)", afterFreeze.Balance, afterFreeze.FrozenAmount)
	}

	if err := adminSvc.ValidateTransaction(ctx, withdraw.ID, false, "invalid beneficiary account", admin.ID, ""); err != nil {
		t.Fatalf("ValidateTransaction (reject) failed: %v", err)
	}

	final, err := env.walletRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to read final wallet: %v", err)
	}
	if !final.Balance.Equal(decimal.NewFromInt(800)) || !final.FrozenAmount.IsZero() {
		t.Fatalf("CONSERVATION VIOLATION: rejected withdrawal must restore exactly (800, 0), got (%s, %s)", final.Balance, final.FrozenAmount)
	}
}
