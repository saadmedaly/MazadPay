//go:build integration

package integrationtest

import (
	"context"
	"testing"

	"github.com/google/uuid"
	apperr "github.com/mazadpay/backend/internal/errors"
	"github.com/shopspring/decimal"
)

// MAZADPAY -- admin wallet controls: admin-only direct wallet-balance debit
// (the mirror of admin_add_balance_test.go's AdminAddBalance coverage) plus
// admin-only wallet disable/re-enable. AdminDeductBalance derives the target
// user_id from the anchor transaction row itself (never trusts a
// client-supplied user_id), debits the wallet and writes a completed
// 'admin_debit' ledger row atomically, and is protected against a
// double-click/retry by uq_admin_debit_reference (migration 000058) --
// exact same shape as admin_credit's uq_admin_credit_reference (000054).

func TestAdminDeductBalance_DeductsExactAmountAndWritesLedgerRow(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST ADMIN DEBIT USER")
	admin := createTestAdmin(t, env, "TEST ADMIN DEBIT ADMIN")
	creditWallet(t, env, user.ID, decimal.NewFromInt(1000))

	walletSvc := newWalletSvc(env)
	adminSvc := newTestAdminService(t, env)

	// Anchor transaction: any real transaction belonging to the target user.
	anchor, err := walletSvc.RequestWithdraw(ctx, user.ID, decimal.NewFromInt(50), "bankily", "")
	if err != nil {
		t.Fatalf("failed to create anchor transaction: %v", err)
	}

	before, err := env.walletRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to read wallet before debit: %v", err)
	}

	ledgerTx, err := adminSvc.AdminDeductBalance(ctx, anchor.ID, decimal.NewFromInt(300), "خصم إداري - مبلغ خاطئ", admin.ID)
	if err != nil {
		t.Fatalf("AdminDeductBalance failed: %v", err)
	}
	if ledgerTx.UserID != user.ID {
		t.Fatalf("expected ledger transaction to target user %s, got %s", user.ID, ledgerTx.UserID)
	}
	if !ledgerTx.Amount.Equal(decimal.NewFromInt(300)) {
		t.Fatalf("expected ledger amount 300, got %s", ledgerTx.Amount)
	}
	if ledgerTx.Type != "admin_debit" {
		t.Fatalf("expected type 'admin_debit', got %q", ledgerTx.Type)
	}
	if ledgerTx.Status != "completed" {
		t.Fatalf("expected status 'completed', got %q", ledgerTx.Status)
	}

	after, err := env.walletRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to read wallet after debit: %v", err)
	}
	expected := before.Balance.Sub(decimal.NewFromInt(300))
	if !after.Balance.Equal(expected) {
		t.Fatalf("expected balance to decrease by exactly 300 (from %s to %s), got %s", before.Balance, expected, after.Balance)
	}
	if !after.FrozenAmount.Equal(before.FrozenAmount) {
		t.Fatalf("CRITICAL: frozen_amount changed (%s -> %s) from an admin debit, which must only touch balance", before.FrozenAmount, after.FrozenAmount)
	}

	var ledgerCount int
	if err := env.db.GetContext(ctx, &ledgerCount,
		`SELECT COUNT(*) FROM transactions WHERE id = $1 AND type = 'admin_debit' AND status = 'completed'`, ledgerTx.ID); err != nil {
		t.Fatalf("failed to query ledger row: %v", err)
	}
	if ledgerCount != 1 {
		t.Fatalf("expected exactly 1 admin_debit ledger row, got %d", ledgerCount)
	}
}

// Focused amounts requested: 50, 100, 300.
func TestAdminDeductBalance_VariousAmounts(t *testing.T) {
	amounts := []decimal.Decimal{decimal.NewFromInt(50), decimal.NewFromInt(100), decimal.NewFromInt(300)}
	for _, amt := range amounts {
		env := setupEnv(t)
		ctx := context.Background()
		user := createTestUser(t, env, "TEST ADMIN DEBIT AMT USER")
		admin := createTestAdmin(t, env, "TEST ADMIN DEBIT AMT ADMIN")
		creditWallet(t, env, user.ID, decimal.NewFromInt(1000))

		walletSvc := newWalletSvc(env)
		adminSvc := newTestAdminService(t, env)

		anchor, err := walletSvc.RequestWithdraw(ctx, user.ID, decimal.NewFromInt(10), "bankily", "")
		if err != nil {
			t.Fatalf("failed to create anchor transaction: %v", err)
		}

		before, err := env.walletRepo.GetByUserID(ctx, user.ID)
		if err != nil {
			t.Fatalf("failed to read wallet: %v", err)
		}

		ledgerTx, err := adminSvc.AdminDeductBalance(ctx, anchor.ID, amt, "", admin.ID)
		if err != nil {
			t.Fatalf("AdminDeductBalance(%s) failed: %v", amt, err)
		}
		if !ledgerTx.Amount.Equal(amt) {
			t.Fatalf("expected ledger amount %s, got %s", amt, ledgerTx.Amount)
		}

		after, err := env.walletRepo.GetByUserID(ctx, user.ID)
		if err != nil {
			t.Fatalf("failed to read wallet: %v", err)
		}
		if !after.Balance.Equal(before.Balance.Sub(amt)) {
			t.Fatalf("expected balance %s after deducting %s from %s, got %s", before.Balance.Sub(amt), amt, before.Balance, after.Balance)
		}
	}
}

func TestAdminDeductBalance_InvalidAmountRejected(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST ADMIN DEBIT BAD AMOUNT USER")
	admin := createTestAdmin(t, env, "TEST ADMIN DEBIT BAD AMOUNT ADMIN")
	creditWallet(t, env, user.ID, decimal.NewFromInt(100))

	walletSvc := newWalletSvc(env)
	adminSvc := newTestAdminService(t, env)

	anchor, err := walletSvc.RequestWithdraw(ctx, user.ID, decimal.NewFromInt(10), "bankily", "")
	if err != nil {
		t.Fatalf("failed to create anchor transaction: %v", err)
	}

	before, err := env.walletRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to read wallet: %v", err)
	}

	for _, amount := range []decimal.Decimal{decimal.Zero, decimal.NewFromInt(-100)} {
		if _, err := adminSvc.AdminDeductBalance(ctx, anchor.ID, amount, "", admin.ID); err != apperr.ErrBadRequest {
			t.Fatalf("expected ErrBadRequest for amount %s, got %v", amount, err)
		}
	}

	after, err := env.walletRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to read wallet: %v", err)
	}
	if !after.Balance.Equal(before.Balance) {
		t.Fatalf("CRITICAL: balance changed (%s -> %s) despite a rejected invalid amount", before.Balance, after.Balance)
	}
}

// A deduct amount greater than the wallet's current balance must be
// rejected atomically -- balance must never go negative.
func TestAdminDeductBalance_InsufficientBalanceRejected(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST ADMIN DEBIT INSUFFICIENT USER")
	admin := createTestAdmin(t, env, "TEST ADMIN DEBIT INSUFFICIENT ADMIN")
	creditWallet(t, env, user.ID, decimal.NewFromInt(100))

	walletSvc := newWalletSvc(env)
	adminSvc := newTestAdminService(t, env)

	anchor, err := walletSvc.RequestWithdraw(ctx, user.ID, decimal.NewFromInt(10), "bankily", "")
	if err != nil {
		t.Fatalf("failed to create anchor transaction: %v", err)
	}

	before, err := env.walletRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to read wallet: %v", err)
	}

	// before.Balance is 90 (100 - 10 frozen for the withdrawal request);
	// attempt to deduct far more than that.
	if _, err := adminSvc.AdminDeductBalance(ctx, anchor.ID, decimal.NewFromInt(100000), "", admin.ID); err != apperr.ErrInsufficientBalanceForDebit {
		t.Fatalf("expected ErrInsufficientBalanceForDebit for an over-large deduction, got %v", err)
	}

	after, err := env.walletRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to read wallet: %v", err)
	}
	if !after.Balance.Equal(before.Balance) {
		t.Fatalf("CRITICAL: balance changed (%s -> %s) despite a rejected insufficient-balance deduction", before.Balance, after.Balance)
	}
	if after.Balance.IsNegative() {
		t.Fatalf("CRITICAL: balance went negative: %s", after.Balance)
	}
}

// A second AdminDeductBalance call against the SAME anchor transaction must
// be rejected cleanly, not double-debit the wallet.
func TestAdminDeductBalance_DuplicateRequestSameAnchorRejected(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST ADMIN DEBIT DUP USER")
	admin := createTestAdmin(t, env, "TEST ADMIN DEBIT DUP ADMIN")
	creditWallet(t, env, user.ID, decimal.NewFromInt(1000))

	walletSvc := newWalletSvc(env)
	adminSvc := newTestAdminService(t, env)

	anchor, err := walletSvc.RequestWithdraw(ctx, user.ID, decimal.NewFromInt(10), "bankily", "")
	if err != nil {
		t.Fatalf("failed to create anchor transaction: %v", err)
	}

	if _, err := adminSvc.AdminDeductBalance(ctx, anchor.ID, decimal.NewFromInt(200), "", admin.ID); err != nil {
		t.Fatalf("first AdminDeductBalance failed: %v", err)
	}

	afterFirst, err := env.walletRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to read wallet after first debit: %v", err)
	}

	if _, err := adminSvc.AdminDeductBalance(ctx, anchor.ID, decimal.NewFromInt(200), "", admin.ID); err != apperr.ErrDuplicateAdminDebit {
		t.Fatalf("expected ErrDuplicateAdminDebit on the second attempt for the same anchor transaction, got %v", err)
	}

	afterSecond, err := env.walletRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to read wallet after rejected second attempt: %v", err)
	}
	if !afterSecond.Balance.Equal(afterFirst.Balance) {
		t.Fatalf("CRITICAL double-debit: balance changed from %s to %s after a rejected duplicate request", afterFirst.Balance, afterSecond.Balance)
	}

	var ledgerCount int
	if err := env.db.GetContext(ctx, &ledgerCount,
		`SELECT COUNT(*) FROM transactions WHERE user_id = $1 AND type = 'admin_debit'`, user.ID); err != nil {
		t.Fatalf("failed to query ledger rows: %v", err)
	}
	if ledgerCount != 1 {
		t.Fatalf("expected exactly 1 admin_debit ledger row after a rejected duplicate, got %d", ledgerCount)
	}
}

func TestAdminDeductBalance_NonExistentAnchorTransactionRejected(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	admin := createTestAdmin(t, env, "TEST ADMIN DEBIT MISSING ANCHOR ADMIN")
	adminSvc := newTestAdminService(t, env)

	_, err := adminSvc.AdminDeductBalance(ctx, uuid.New(), decimal.NewFromInt(100), "", admin.ID)
	if err == nil {
		t.Fatalf("expected an error for a non-existent anchor transaction id")
	}
}

// ==================================================
// AdminSetWalletDisabled / wallet-disabled spend enforcement
// ==================================================

func TestAdminSetWalletDisabled_DisablesAndPersists(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST WALLET DISABLE USER")
	admin := createTestAdmin(t, env, "TEST WALLET DISABLE ADMIN")
	creditWallet(t, env, user.ID, decimal.NewFromInt(500))
	adminSvc := newTestAdminService(t, env)

	if err := adminSvc.AdminSetWalletDisabled(ctx, user.ID, true, "نشاط مشبوه", admin.ID); err != nil {
		t.Fatalf("AdminSetWalletDisabled(disable) failed: %v", err)
	}

	wallet, err := env.walletRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to read wallet: %v", err)
	}
	if !wallet.IsDisabled {
		t.Fatalf("expected is_disabled=true after disabling")
	}
	if wallet.DisabledReason == nil || *wallet.DisabledReason != "نشاط مشبوه" {
		t.Fatalf("expected disabled_reason to be persisted, got %v", wallet.DisabledReason)
	}
	if wallet.DisabledBy == nil || *wallet.DisabledBy != admin.ID {
		t.Fatalf("expected disabled_by=%s, got %v", admin.ID, wallet.DisabledBy)
	}
	if wallet.DisabledAt == nil {
		t.Fatalf("expected disabled_at to be set")
	}
	// Balance itself must be completely untouched by disabling -- this is a
	// pure status flag, never a money movement.
	if !wallet.Balance.Equal(decimal.NewFromInt(500)) {
		t.Fatalf("CRITICAL: balance changed by disabling the wallet, got %s", wallet.Balance)
	}
}

// The core behavior: once disabled, the user must not be able to place a
// bid requiring a NEW insurance freeze (the real spend gate).
func TestWalletDisabled_BlocksNewBidInsuranceFreeze(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST WALLET DISABLE SELLER")
	bidder := createTestUser(t, env, "TEST WALLET DISABLE BIDDER")
	admin := createTestAdmin(t, env, "TEST WALLET DISABLE BID ADMIN")
	creditWallet(t, env, bidder.ID, decimal.NewFromInt(1000))
	adminSvc := newTestAdminService(t, env)

	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")

	if err := adminSvc.AdminSetWalletDisabled(ctx, bidder.ID, true, "test", admin.ID); err != nil {
		t.Fatalf("AdminSetWalletDisabled(disable) failed: %v", err)
	}

	_, err := env.bidSvc.PlaceBid(ctx, auction.ID, bidder.ID, decimal.NewFromInt(110))
	if err != apperr.ErrWalletDisabled {
		t.Fatalf("expected ErrWalletDisabled for a bid from a disabled wallet, got %v", err)
	}
}

// The other real spend gate: withdrawal requests must also be blocked while
// disabled.
func TestWalletDisabled_BlocksNewWithdrawalRequest(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST WALLET DISABLE WITHDRAW USER")
	admin := createTestAdmin(t, env, "TEST WALLET DISABLE WITHDRAW ADMIN")
	creditWallet(t, env, user.ID, decimal.NewFromInt(1000))
	adminSvc := newTestAdminService(t, env)
	walletSvc := newWalletSvc(env)

	if err := adminSvc.AdminSetWalletDisabled(ctx, user.ID, true, "test", admin.ID); err != nil {
		t.Fatalf("AdminSetWalletDisabled(disable) failed: %v", err)
	}

	_, err := walletSvc.RequestWithdraw(ctx, user.ID, decimal.NewFromInt(100), "bankily", "")
	if err != apperr.ErrWalletDisabled {
		t.Fatalf("expected ErrWalletDisabled for a withdrawal request from a disabled wallet, got %v", err)
	}
}

// Re-enabling must restore normal spend behavior and clear the reason.
func TestAdminSetWalletDisabled_ReEnableRestoresSpending(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST WALLET REENABLE USER")
	admin := createTestAdmin(t, env, "TEST WALLET REENABLE ADMIN")
	creditWallet(t, env, user.ID, decimal.NewFromInt(1000))
	adminSvc := newTestAdminService(t, env)
	walletSvc := newWalletSvc(env)

	if err := adminSvc.AdminSetWalletDisabled(ctx, user.ID, true, "temporary hold", admin.ID); err != nil {
		t.Fatalf("AdminSetWalletDisabled(disable) failed: %v", err)
	}
	if _, err := walletSvc.RequestWithdraw(ctx, user.ID, decimal.NewFromInt(100), "bankily", ""); err != apperr.ErrWalletDisabled {
		t.Fatalf("expected withdrawal to be blocked while disabled, got %v", err)
	}

	if err := adminSvc.AdminSetWalletDisabled(ctx, user.ID, false, "", admin.ID); err != nil {
		t.Fatalf("AdminSetWalletDisabled(re-enable) failed: %v", err)
	}

	wallet, err := env.walletRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to read wallet: %v", err)
	}
	if wallet.IsDisabled {
		t.Fatalf("expected is_disabled=false after re-enabling")
	}
	if wallet.DisabledReason != nil || wallet.DisabledBy != nil || wallet.DisabledAt != nil {
		t.Fatalf("expected disabled_reason/disabled_by/disabled_at to be cleared on re-enable, got reason=%v by=%v at=%v", wallet.DisabledReason, wallet.DisabledBy, wallet.DisabledAt)
	}

	// Withdrawal must now succeed normally.
	tx, err := walletSvc.RequestWithdraw(ctx, user.ID, decimal.NewFromInt(100), "bankily", "")
	if err != nil {
		t.Fatalf("expected withdrawal to succeed after re-enabling, got %v", err)
	}
	if tx.Status != "pending_review" {
		t.Fatalf("expected a normal pending_review withdrawal after re-enabling, got status=%s", tx.Status)
	}
}

// Admin add/deduct-balance must remain usable on a DISABLED wallet -- an
// admin must always be able to correct a disabled wallet's balance.
func TestWalletDisabled_AdminCreditAndDebitStillWork(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST WALLET DISABLE ADMIN OVERRIDE USER")
	admin := createTestAdmin(t, env, "TEST WALLET DISABLE ADMIN OVERRIDE ADMIN")
	creditWallet(t, env, user.ID, decimal.NewFromInt(500))
	adminSvc := newTestAdminService(t, env)
	walletSvc := newWalletSvc(env)

	// Anchor transaction created BEFORE disabling (a withdrawal request would
	// itself be blocked once disabled, so it must be requested first here).
	anchor, err := walletSvc.RequestWithdraw(ctx, user.ID, decimal.NewFromInt(10), "bankily", "")
	if err != nil {
		t.Fatalf("failed to create anchor transaction: %v", err)
	}

	if err := adminSvc.AdminSetWalletDisabled(ctx, user.ID, true, "test", admin.ID); err != nil {
		t.Fatalf("AdminSetWalletDisabled(disable) failed: %v", err)
	}

	if _, err := adminSvc.AdminAddBalance(ctx, anchor.ID, decimal.NewFromInt(200), "", admin.ID); err != nil {
		t.Fatalf("expected AdminAddBalance to succeed on a disabled wallet, got %v", err)
	}

	// Reusing the SAME anchor transaction id for the debit call is legal --
	// AdminDeductBalance's idempotency reference is "admin_debit:<anchor>",
	// a distinct namespace from AdminAddBalance's "admin_credit:<anchor>". A
	// fresh anchor can't be created here since RequestWithdraw is itself
	// correctly blocked while the wallet is disabled.
	if _, err := adminSvc.AdminDeductBalance(ctx, anchor.ID, decimal.NewFromInt(50), "", admin.ID); err != nil {
		t.Fatalf("expected AdminDeductBalance to succeed on a disabled wallet, got %v", err)
	}

	wallet, err := env.walletRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to read wallet: %v", err)
	}
	// 500 initial -> 490 after freezing 10 for the withdrawal request -> +200 credit -> -50 debit = 640.
	if !wallet.Balance.Equal(decimal.NewFromInt(640)) {
		t.Fatalf("expected balance=640 after credit+debit on a disabled wallet, got %s", wallet.Balance)
	}
	if !wallet.IsDisabled {
		t.Fatalf("expected wallet to remain disabled (admin credit/debit must not implicitly re-enable it)")
	}
}
