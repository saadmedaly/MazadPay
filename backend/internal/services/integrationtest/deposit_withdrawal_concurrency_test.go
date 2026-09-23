//go:build integration

package integrationtest

import (
	"context"
	"testing"

	"github.com/shopspring/decimal"
)

// Financial Audit Phases 4/5: concurrent-approval coverage for the normal
// (non-admin_credit) deposit/withdrawal review path. TransactionRepository.
// UpdateStatus already guards against a double-approve via
// `SELECT ... FOR UPDATE` on the transaction row plus a
// tx.Status != "completed" check read from that SAME locked row -- these
// tests prove that guard holds under genuine concurrent Go-level requests
// against a real database, not just by reading the code.

// TestDepositApproval_ConcurrentApproveAttempts_ExactlyOneCredit fires two
// concurrent ValidateTransaction(approve=true) calls against the SAME
// pending deposit and asserts the wallet is credited exactly once.
func TestDepositApproval_ConcurrentApproveAttempts_ExactlyOneCredit(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST CONCURRENT DEPOSIT APPROVAL USER")
	admin := createTestAdmin(t, env, "TEST CONCURRENT DEPOSIT APPROVAL ADMIN")

	walletSvc := newWalletSvc(env)
	adminSvc := newTestAdminService(t, env)

	deposit, err := walletSvc.InitiateDeposit(ctx, user.ID, decimal.NewFromInt(4000), "bankily", "", "", nil)
	if err != nil {
		t.Fatalf("InitiateDeposit failed: %v", err)
	}

	errs := make(chan error, 2)
	for i := 0; i < 2; i++ {
		go func() {
			errs <- adminSvc.ValidateTransaction(ctx, deposit.ID, true, "", admin.ID, "")
		}()
	}
	for i := 0; i < 2; i++ {
		if err := <-errs; err != nil {
			t.Fatalf("unexpected error from concurrent ValidateTransaction: %v", err)
		}
	}

	wallet, err := env.walletRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to read wallet: %v", err)
	}
	if !wallet.Balance.Equal(decimal.NewFromInt(4000)) {
		t.Fatalf("CRITICAL DOUBLE-CREDIT: expected exactly one 4000 credit from two concurrent approvals, got balance %s", wallet.Balance)
	}
}

// TestWithdrawalApproval_ConcurrentApproveAttempts_ExactlyOneCapture mirrors
// the above for withdrawals: two concurrent approve calls against the same
// pending_review withdrawal must capture the frozen amount exactly once
// (frozen_amount -> 0), never twice, never negative.
func TestWithdrawalApproval_ConcurrentApproveAttempts_ExactlyOneCapture(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST CONCURRENT WITHDRAW APPROVAL USER")
	admin := createTestAdmin(t, env, "TEST CONCURRENT WITHDRAW APPROVAL ADMIN")
	creditWallet(t, env, user.ID, decimal.NewFromInt(2000))

	walletSvc := newWalletSvc(env)
	adminSvc := newTestAdminService(t, env)

	withdraw, err := walletSvc.RequestWithdraw(ctx, user.ID, decimal.NewFromInt(600), "bankily", "22247601175")
	if err != nil {
		t.Fatalf("RequestWithdraw failed: %v", err)
	}

	errs := make(chan error, 2)
	for i := 0; i < 2; i++ {
		go func() {
			errs <- adminSvc.ValidateTransaction(ctx, withdraw.ID, true, "", admin.ID, "")
		}()
	}
	for i := 0; i < 2; i++ {
		if err := <-errs; err != nil {
			t.Fatalf("unexpected error from concurrent withdrawal ValidateTransaction: %v", err)
		}
	}

	wallet, err := env.walletRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to read wallet: %v", err)
	}
	// balance was reduced by 600 at request time (freeze), never touched
	// again by approval (approval only captures frozen_amount).
	if !wallet.Balance.Equal(decimal.NewFromInt(1400)) {
		t.Fatalf("expected balance to remain 1400 (2000-600, untouched by approval), got %s", wallet.Balance)
	}
	if !wallet.FrozenAmount.IsZero() {
		t.Fatalf("CRITICAL: expected frozen_amount exactly 0 after concurrent approval capture, got %s (never negative, never left non-zero by a lost update)", wallet.FrozenAmount)
	}
}

// TestConcurrentWithdrawals_SameUser_CannotOverdraw fires two concurrent
// withdrawal REQUESTS (not approvals) for amounts that individually fit but
// together exceed balance -- FreezeForWithdraw's compare-and-set
// (`WHERE balance >= amount`) must ensure only the requests that actually
// fit succeed; the wallet must never go negative.
func TestConcurrentWithdrawals_SameUser_CannotOverdraw(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST CONCURRENT OVERDRAW USER")
	creditWallet(t, env, user.ID, decimal.NewFromInt(1000))

	walletSvc := newWalletSvc(env)

	// Two concurrent withdrawal requests for 700 each -- only one can
	// possibly succeed against a 1000 balance (700+700=1400 > 1000).
	type result struct {
		err error
	}
	results := make(chan result, 2)
	for i := 0; i < 2; i++ {
		go func() {
			_, err := walletSvc.RequestWithdraw(ctx, user.ID, decimal.NewFromInt(700), "bankily", "22247601175")
			results <- result{err: err}
		}()
	}

	successCount := 0
	for i := 0; i < 2; i++ {
		r := <-results
		if r.err == nil {
			successCount++
		}
	}

	if successCount != 1 {
		t.Fatalf("expected exactly 1 of 2 concurrent 700-withdrawals against a 1000 balance to succeed, got %d", successCount)
	}

	wallet, err := env.walletRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to read wallet: %v", err)
	}
	if wallet.Balance.IsNegative() {
		t.Fatalf("CRITICAL: wallet balance went negative (%s) from concurrent withdrawal requests", wallet.Balance)
	}
	if !wallet.Balance.Equal(decimal.NewFromInt(300)) {
		t.Fatalf("expected balance exactly 300 (1000-700, from the single successful request), got %s", wallet.Balance)
	}
}
