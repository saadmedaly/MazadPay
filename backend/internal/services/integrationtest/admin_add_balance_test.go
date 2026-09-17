//go:build integration

package integrationtest

import (
	"context"
	"testing"

	"github.com/google/uuid"
	apperr "github.com/mazadpay/backend/internal/errors"
	"github.com/shopspring/decimal"
)

// Customer Request #35: admin-only direct wallet-balance credit, reachable
// from a transaction's detail page. AdminAddBalance derives the target
// user_id from the anchor transaction row itself (never trusts a
// client-supplied user_id), credits the wallet and writes a completed
// 'admin_credit' ledger row atomically, and is protected against a
// double-click/retry by uq_admin_credit_reference (migration 000054).

func TestAdminAddBalance_CreditsExactAmountAndWritesLedgerRow(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST ADMIN CREDIT USER")
	admin := createTestAdmin(t, env, "TEST ADMIN CREDIT ADMIN")
	creditWallet(t, env, user.ID, decimal.NewFromInt(100))

	walletSvc := newWalletSvc(env)
	adminSvc := newTestAdminService(t, env)

	// Anchor transaction: any real transaction belonging to the target user
	// (a withdrawal request here) -- AdminAddBalance derives user_id from it.
	anchor, err := walletSvc.RequestWithdraw(ctx, user.ID, decimal.NewFromInt(50), "bankily", "")
	if err != nil {
		t.Fatalf("failed to create anchor transaction: %v", err)
	}

	before, err := env.walletRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to read wallet before credit: %v", err)
	}

	ledgerTx, err := adminSvc.AdminAddBalance(ctx, anchor.ID, decimal.NewFromInt(500), "top-up per client request", admin.ID)
	if err != nil {
		t.Fatalf("AdminAddBalance failed: %v", err)
	}
	if ledgerTx.UserID != user.ID {
		t.Fatalf("expected ledger transaction to target user %s, got %s", user.ID, ledgerTx.UserID)
	}
	if !ledgerTx.Amount.Equal(decimal.NewFromInt(500)) {
		t.Fatalf("expected ledger amount 500, got %s", ledgerTx.Amount)
	}
	if ledgerTx.Type != "admin_credit" {
		t.Fatalf("expected type 'admin_credit', got %q", ledgerTx.Type)
	}
	if ledgerTx.Status != "completed" {
		t.Fatalf("expected status 'completed', got %q", ledgerTx.Status)
	}

	after, err := env.walletRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to read wallet after credit: %v", err)
	}
	// Balance was already reduced by 50 from RequestWithdraw's freeze
	// (100 -> 50 balance / 50 frozen), then credited exactly +500.
	expected := before.Balance.Add(decimal.NewFromInt(500))
	if !after.Balance.Equal(expected) {
		t.Fatalf("expected balance to increase by exactly 500 (from %s to %s), got %s", before.Balance, expected, after.Balance)
	}

	var ledgerCount int
	if err := env.db.GetContext(ctx, &ledgerCount,
		`SELECT COUNT(*) FROM transactions WHERE id = $1 AND type = 'admin_credit' AND status = 'completed'`, ledgerTx.ID); err != nil {
		t.Fatalf("failed to query ledger row: %v", err)
	}
	if ledgerCount != 1 {
		t.Fatalf("expected exactly 1 admin_credit ledger row, got %d", ledgerCount)
	}
}

func TestAdminAddBalance_InvalidAmountRejected(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST ADMIN CREDIT BAD AMOUNT USER")
	admin := createTestAdmin(t, env, "TEST ADMIN CREDIT BAD AMOUNT ADMIN")
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
		if _, err := adminSvc.AdminAddBalance(ctx, anchor.ID, amount, "", admin.ID); err != apperr.ErrBadRequest {
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

// A second AdminAddBalance call against the SAME anchor transaction (e.g. an
// admin double-clicking "إضافة الرصيد", or a client retry after a dropped
// response) must be rejected cleanly, not double-credit the wallet.
func TestAdminAddBalance_DuplicateRequestSameAnchorRejected(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST ADMIN CREDIT DUP USER")
	admin := createTestAdmin(t, env, "TEST ADMIN CREDIT DUP ADMIN")
	creditWallet(t, env, user.ID, decimal.NewFromInt(100))

	walletSvc := newWalletSvc(env)
	adminSvc := newTestAdminService(t, env)

	anchor, err := walletSvc.RequestWithdraw(ctx, user.ID, decimal.NewFromInt(10), "bankily", "")
	if err != nil {
		t.Fatalf("failed to create anchor transaction: %v", err)
	}

	if _, err := adminSvc.AdminAddBalance(ctx, anchor.ID, decimal.NewFromInt(200), "", admin.ID); err != nil {
		t.Fatalf("first AdminAddBalance failed: %v", err)
	}

	afterFirst, err := env.walletRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to read wallet after first credit: %v", err)
	}

	if _, err := adminSvc.AdminAddBalance(ctx, anchor.ID, decimal.NewFromInt(200), "", admin.ID); err != apperr.ErrDuplicateAdminCredit {
		t.Fatalf("expected ErrDuplicateAdminCredit on the second attempt for the same anchor transaction, got %v", err)
	}

	afterSecond, err := env.walletRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to read wallet after rejected second attempt: %v", err)
	}
	if !afterSecond.Balance.Equal(afterFirst.Balance) {
		t.Fatalf("CRITICAL double-credit: balance changed from %s to %s after a rejected duplicate request", afterFirst.Balance, afterSecond.Balance)
	}

	var ledgerCount int
	if err := env.db.GetContext(ctx, &ledgerCount,
		`SELECT COUNT(*) FROM transactions WHERE user_id = $1 AND type = 'admin_credit'`, user.ID); err != nil {
		t.Fatalf("failed to query ledger rows: %v", err)
	}
	if ledgerCount != 1 {
		t.Fatalf("expected exactly 1 admin_credit ledger row after a rejected duplicate, got %d", ledgerCount)
	}
}

// A non-existent anchor transaction id must fail cleanly (no wallet touched,
// no ledger row created) rather than crediting an undetermined user.
func TestAdminAddBalance_NonExistentAnchorTransactionRejected(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	admin := createTestAdmin(t, env, "TEST ADMIN CREDIT MISSING ANCHOR ADMIN")
	adminSvc := newTestAdminService(t, env)

	_, err := adminSvc.AdminAddBalance(ctx, uuid.New(), decimal.NewFromInt(100), "", admin.ID)
	if err == nil {
		t.Fatalf("expected an error for a non-existent anchor transaction id")
	}
}
