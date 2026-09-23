//go:build integration

package integrationtest

import (
	"context"
	"testing"

	apperr "github.com/mazadpay/backend/internal/errors"
	"github.com/shopspring/decimal"
)

// Financial Audit finding (live-reproduced on Validation): AdminAddBalance
// (admin_credit) previously never checked or updated the anchor
// transaction's own status/type/amount -- it read only anchor.UserID. If the
// anchor was itself a still-pending deposit, that deposit remained fully
// approvable afterward via the normal review flow
// (AdminService.ValidateTransaction -> TransactionRepository.UpdateStatus),
// which credits the wallet by the DEPOSIT'S OWN stored amount a second time,
// independent of whatever amount the admin_credit used. Observed live: a
// 5000 MRU deposit's admin_credit was issued for only 500 MRU, leaving the
// original 5000 MRU deposit transaction sitting in 'pending', fully
// approvable, and worth 10x the admin_credit amount if later approved.
//
// The fix: TransactionRepository.LockAndCloseIfOpenDepositOrWithdraw is
// called inside AdminAddBalance/AdminDeductBalance's own dbtx, BEFORE the
// wallet is locked (matching UpdateStatus's own lock order to avoid a
// deadlock), and atomically marks a still-open (pending/pending_review)
// deposit/withdraw anchor 'completed' in the same commit as the wallet
// mutation -- so UpdateStatus's own tx.Status != "completed" guard then
// blocks any later independent approval from moving money again. A
// standalone admin_credit/admin_debit against any OTHER anchor type, or an
// anchor that's already terminal, is left completely untouched (this is
// deliberately preserved as an independent manual-adjustment mechanism, not
// merged into "deposit approval").

// TestAdminAddBalance_AgainstPendingDeposit_ThenApprove_NoDoubleCreit is the
// direct reproduction of the exact sequence observed live on Validation.
func TestAdminAddBalance_AgainstPendingDeposit_ThenApprove_NoDoubleCreit(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST DOUBLE CREDIT USER")
	admin := createTestAdmin(t, env, "TEST DOUBLE CREDIT ADMIN")

	walletSvc := newWalletSvc(env)
	adminSvc := newTestAdminService(t, env)
	txRepo := newTestTxRepo(env)

	// Step 2: create a pending deposit for 5000 (mirrors the Validation
	// observation exactly).
	deposit, err := walletSvc.InitiateDeposit(ctx, user.ID, decimal.NewFromInt(5000), "bankily", "", "", nil)
	if err != nil {
		t.Fatalf("InitiateDeposit failed: %v", err)
	}
	if deposit.Status == "completed" {
		t.Fatalf("test setup invariant violated: deposit must start non-terminal, got %q", deposit.Status)
	}

	// Step 3: admin issues an admin_credit against this SAME deposit's id as
	// anchor, but for a DIFFERENT (mismatched) amount -- 500, exactly as
	// observed live.
	if _, err := adminSvc.AdminAddBalance(ctx, deposit.ID, decimal.NewFromInt(500), "manual top-up (mismatched amount, as observed live)", admin.ID); err != nil {
		t.Fatalf("AdminAddBalance failed: %v", err)
	}

	// Step 4: confirm wallet credited once, by the admin_credit's own amount.
	afterCredit, err := env.walletRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to read wallet after admin_credit: %v", err)
	}
	if !afterCredit.Balance.Equal(decimal.NewFromInt(500)) {
		t.Fatalf("expected balance 500 after the single admin_credit, got %s", afterCredit.Balance)
	}

	// Step 5: THE FIX -- confirm the original deposit's own status was
	// closed out by the admin_credit, so it is no longer independently
	// approvable.
	fetchedDeposit, err := txRepo.GetByID(ctx, deposit.ID)
	if err != nil {
		t.Fatalf("failed to re-fetch deposit: %v", err)
	}
	if fetchedDeposit.Status != "completed" {
		t.Fatalf("REGRESSION: expected the anchor deposit to be closed ('completed') after admin_credit against it, got %q -- this is the exact gap that allowed double-crediting", fetchedDeposit.Status)
	}

	// Step 6/7: attempt normal approval of the original deposit -- this MUST
	// now be a safe no-op (UpdateStatus's own tx.Status != "completed" guard
	// blocks it), never bringing the wallet to 2x anything.
	if err := adminSvc.ValidateTransaction(ctx, deposit.ID, true, "", admin.ID, ""); err != nil {
		t.Fatalf("ValidateTransaction on an already-closed deposit should be a safe no-op, not an error: %v", err)
	}

	afterApproval, err := env.walletRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to read wallet after the redundant approval attempt: %v", err)
	}
	if !afterApproval.Balance.Equal(decimal.NewFromInt(500)) {
		t.Fatalf("CRITICAL DOUBLE-CREDIT: balance changed from 500 to %s after approving an already-admin_credit'd deposit -- the wallet must never reach anywhere near 5500", afterApproval.Balance)
	}
}

// TestAdminAddBalance_MismatchedAmountVsAnchorDeposit_StillOnlyCreditsOnce
// makes explicit that the fix does not attempt to enforce amount equality
// between the admin_credit and its anchor deposit (that would be a much
// larger behavior change to a mechanism this codebase treats as an
// independent manual-adjustment tool, not literally "deposit approval") --
// what it DOES guarantee is that whatever amount is credited happens
// exactly once, because the anchor can never be independently approved
// afterward for its own (possibly different) amount.
func TestAdminAddBalance_MismatchedAmountVsAnchorDeposit_StillOnlyCreditsOnce(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST MISMATCH AMOUNT USER")
	admin := createTestAdmin(t, env, "TEST MISMATCH AMOUNT ADMIN")

	walletSvc := newWalletSvc(env)
	adminSvc := newTestAdminService(t, env)

	deposit, err := walletSvc.InitiateDeposit(ctx, user.ID, decimal.NewFromInt(10000), "bankily", "", "", nil)
	if err != nil {
		t.Fatalf("InitiateDeposit failed: %v", err)
	}

	if _, err := adminSvc.AdminAddBalance(ctx, deposit.ID, decimal.NewFromInt(1), "deliberately tiny mismatch", admin.ID); err != nil {
		t.Fatalf("AdminAddBalance failed: %v", err)
	}

	if err := adminSvc.ValidateTransaction(ctx, deposit.ID, true, "", admin.ID, ""); err != nil {
		t.Fatalf("ValidateTransaction on the closed deposit should be a safe no-op: %v", err)
	}

	wallet, err := env.walletRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to read wallet: %v", err)
	}
	if !wallet.Balance.Equal(decimal.NewFromInt(1)) {
		t.Fatalf("expected exactly the admin_credit's own amount (1), got %s -- the 10000 deposit amount must never separately land in the wallet", wallet.Balance)
	}
}

// TestAdminAddBalance_ApprovalFirst_ThenAdminCredit_StandardDuplicateGuardStillApplies
// covers the reverse ordering: approve the deposit normally first, THEN
// attempt an admin_credit against the same (now-completed) transaction id.
// This must be rejected the same way a duplicate admin_credit always was
// (ErrDuplicateAdminCredit is NOT expected here since no admin_credit ran
// yet for this reference -- the wallet mutation itself must simply not
// double-apply the deposit's own value). AdminAddBalance has no inherent
// reason to refuse crediting an unrelated amount against an already-
// completed anchor (that's still a valid standalone manual credit), so this
// test asserts the DEPOSIT's own amount was applied exactly once, and the
// subsequent admin_credit is additive on top -- never itself a duplicate of
// the deposit.
func TestAdminAddBalance_ApprovalFirst_ThenAdminCredit_NoDoubleCredit(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST APPROVAL FIRST USER")
	admin := createTestAdmin(t, env, "TEST APPROVAL FIRST ADMIN")

	walletSvc := newWalletSvc(env)
	adminSvc := newTestAdminService(t, env)

	deposit, err := walletSvc.InitiateDeposit(ctx, user.ID, decimal.NewFromInt(2000), "bankily", "", "", nil)
	if err != nil {
		t.Fatalf("InitiateDeposit failed: %v", err)
	}

	if err := adminSvc.ValidateTransaction(ctx, deposit.ID, true, "", admin.ID, ""); err != nil {
		t.Fatalf("ValidateTransaction (normal approval) failed: %v", err)
	}

	afterApproval, err := env.walletRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to read wallet after approval: %v", err)
	}
	if !afterApproval.Balance.Equal(decimal.NewFromInt(2000)) {
		t.Fatalf("expected balance 2000 after normal deposit approval, got %s", afterApproval.Balance)
	}

	// A LATER admin_credit against the now-completed deposit id is a
	// standalone additive credit (the anchor's own status is untouched by
	// LockAndCloseIfOpenDepositOrWithdraw since it's already 'completed') --
	// this must simply add the admin_credit's own amount, not re-apply the
	// deposit's 2000.
	if _, err := adminSvc.AdminAddBalance(ctx, deposit.ID, decimal.NewFromInt(300), "unrelated bonus, deposit already settled", admin.ID); err != nil {
		t.Fatalf("AdminAddBalance against an already-completed anchor should succeed as a standalone credit: %v", err)
	}

	final, err := env.walletRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to read final wallet: %v", err)
	}
	if !final.Balance.Equal(decimal.NewFromInt(2300)) {
		t.Fatalf("expected exactly 2000 (deposit) + 300 (standalone admin_credit) = 2300, got %s", final.Balance)
	}
}

// TestAdminDeductBalance_AgainstPendingWithdraw_ClosesAnchor mirrors the
// credit-side fix for the debit path: an admin_debit against a still-open
// withdraw anchor must close it out, so a later independent
// approve/reject of that withdrawal can never separately move
// balance/frozen_amount again.
func TestAdminDeductBalance_AgainstPendingWithdraw_ClosesAnchor(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST DEBIT CLOSE ANCHOR USER")
	admin := createTestAdmin(t, env, "TEST DEBIT CLOSE ANCHOR ADMIN")
	creditWallet(t, env, user.ID, decimal.NewFromInt(1000))

	walletSvc := newWalletSvc(env)
	adminSvc := newTestAdminService(t, env)
	txRepo := newTestTxRepo(env)

	withdraw, err := walletSvc.RequestWithdraw(ctx, user.ID, decimal.NewFromInt(400), "bankily", "22247601175")
	if err != nil {
		t.Fatalf("RequestWithdraw failed: %v", err)
	}

	if _, err := adminSvc.AdminDeductBalance(ctx, withdraw.ID, decimal.NewFromInt(50), "unrelated clawback", admin.ID); err != nil {
		t.Fatalf("AdminDeductBalance failed: %v", err)
	}

	fetchedWithdraw, err := txRepo.GetByID(ctx, withdraw.ID)
	if err != nil {
		t.Fatalf("failed to re-fetch withdraw: %v", err)
	}
	if fetchedWithdraw.Status != "completed" {
		t.Fatalf("expected the anchor withdraw to be closed ('completed') after admin_debit against it, got %q", fetchedWithdraw.Status)
	}

	beforeApproval, err := env.walletRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to read wallet: %v", err)
	}

	if err := adminSvc.ValidateTransaction(ctx, withdraw.ID, true, "", admin.ID, ""); err != nil {
		t.Fatalf("ValidateTransaction on the closed withdraw should be a safe no-op: %v", err)
	}

	afterApproval, err := env.walletRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to read wallet after the redundant approval attempt: %v", err)
	}
	if !afterApproval.FrozenAmount.Equal(beforeApproval.FrozenAmount) || !afterApproval.Balance.Equal(beforeApproval.Balance) {
		t.Fatalf("wallet state changed after approving an already-closed withdraw (balance %s->%s, frozen %s->%s) -- must be a no-op",
			beforeApproval.Balance, afterApproval.Balance, beforeApproval.FrozenAmount, afterApproval.FrozenAmount)
	}
}

// TestAdminAddBalance_StandaloneCreditAgainstNonDepositAnchor_Unaffected
// proves the fix's no-op path: admin_credit is a legitimate independent
// manual-adjustment mechanism, and using ANY transaction (not just a
// pending deposit/withdraw) as the anchor for audit-trail purposes must
// keep working exactly as before -- this mirrors the existing
// TestAdminAddBalance_CreditsExactAmountAndWritesLedgerRow test's own anchor
// choice (a withdrawal, used purely for its user_id) to confirm the new
// LockAndCloseIfOpenDepositOrWithdraw call doesn't change behavior for an
// anchor the admin never intended to "close" via this action... except that
// per the credit-side test above, closing WOULD occur if the anchor is
// itself still pending. This test instead anchors on an already-completed
// transaction (a second, already-approved deposit) to confirm the
// close-if-open guard's no-op branch for a terminal anchor leaves it
// untouched.
func TestAdminAddBalance_StandaloneCreditAgainstAlreadyCompletedAnchor_Unaffected(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST STANDALONE CREDIT USER")
	admin := createTestAdmin(t, env, "TEST STANDALONE CREDIT ADMIN")

	walletSvc := newWalletSvc(env)
	adminSvc := newTestAdminService(t, env)
	txRepo := newTestTxRepo(env)

	deposit, err := walletSvc.InitiateDeposit(ctx, user.ID, decimal.NewFromInt(1000), "bankily", "", "", nil)
	if err != nil {
		t.Fatalf("InitiateDeposit failed: %v", err)
	}
	if err := adminSvc.ValidateTransaction(ctx, deposit.ID, true, "", admin.ID, ""); err != nil {
		t.Fatalf("ValidateTransaction failed: %v", err)
	}

	reviewedAtBefore, err := txRepo.GetByID(ctx, deposit.ID)
	if err != nil {
		t.Fatalf("failed to fetch deposit: %v", err)
	}

	if _, err := adminSvc.AdminAddBalance(ctx, deposit.ID, decimal.NewFromInt(75), "bonus, unrelated to the already-settled deposit", admin.ID); err != nil {
		t.Fatalf("AdminAddBalance against an already-completed anchor should succeed: %v", err)
	}

	reviewedAtAfter, err := txRepo.GetByID(ctx, deposit.ID)
	if err != nil {
		t.Fatalf("failed to fetch deposit: %v", err)
	}
	if reviewedAtBefore.ReviewedAt != nil && reviewedAtAfter.ReviewedAt != nil &&
		!reviewedAtBefore.ReviewedAt.Equal(*reviewedAtAfter.ReviewedAt) {
		t.Fatalf("the already-completed anchor's reviewed_at was modified by a standalone admin_credit against it -- the no-op guard should leave a terminal anchor completely untouched")
	}

	wallet, err := env.walletRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to read wallet: %v", err)
	}
	if !wallet.Balance.Equal(decimal.NewFromInt(1075)) {
		t.Fatalf("expected 1000 (approved deposit) + 75 (standalone admin_credit) = 1075, got %s", wallet.Balance)
	}
}

// TestAdminAddBalance_ConcurrentAgainstSamePendingDeposit_ExactlyOneCredit
// proves the fix is race-safe: two concurrent AdminAddBalance calls against
// the SAME pending deposit anchor must not both succeed in crediting the
// wallet -- the existing uq_admin_credit_reference unique index (both calls
// build the identical deterministic reference string
// "admin_credit:<anchorID>") rejects the second with
// ErrDuplicateAdminCredit, and the anchor-closing row lock
// (LockAndCloseIfOpenDepositOrWithdraw's SELECT ... FOR UPDATE) additionally
// serializes the two attempts rather than allowing an interleaved partial
// state.
func TestAdminAddBalance_ConcurrentAgainstSamePendingDeposit_ExactlyOneCredit(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST CONCURRENT CREDIT USER")
	admin := createTestAdmin(t, env, "TEST CONCURRENT CREDIT ADMIN")

	walletSvc := newWalletSvc(env)
	adminSvc := newTestAdminService(t, env)

	deposit, err := walletSvc.InitiateDeposit(ctx, user.ID, decimal.NewFromInt(3000), "bankily", "", "", nil)
	if err != nil {
		t.Fatalf("InitiateDeposit failed: %v", err)
	}

	type result struct {
		err error
	}
	results := make(chan result, 2)
	for i := 0; i < 2; i++ {
		go func() {
			_, err := adminSvc.AdminAddBalance(ctx, deposit.ID, decimal.NewFromInt(3000), "concurrent attempt", admin.ID)
			results <- result{err: err}
		}()
	}

	successCount := 0
	duplicateCount := 0
	for i := 0; i < 2; i++ {
		r := <-results
		if r.err == nil {
			successCount++
		} else if r.err == apperr.ErrDuplicateAdminCredit {
			duplicateCount++
		} else {
			t.Fatalf("unexpected error from concurrent AdminAddBalance: %v", r.err)
		}
	}

	if successCount != 1 {
		t.Fatalf("expected exactly 1 successful concurrent AdminAddBalance, got %d", successCount)
	}
	if duplicateCount != 1 {
		t.Fatalf("expected exactly 1 ErrDuplicateAdminCredit among the concurrent attempts, got %d", duplicateCount)
	}

	wallet, err := env.walletRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to read wallet: %v", err)
	}
	if !wallet.Balance.Equal(decimal.NewFromInt(3000)) {
		t.Fatalf("CRITICAL: expected exactly one 3000 credit from the concurrent race, got balance %s", wallet.Balance)
	}
}
