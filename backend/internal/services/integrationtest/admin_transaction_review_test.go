//go:build integration

package integrationtest

import (
	"context"
	"testing"

	"github.com/mazadpay/backend/internal/repository"
	"github.com/shopspring/decimal"
)

// newTestTxRepo mirrors newWalletSvc/newTestAdminService's construction --
// testEnv has no txRepo field of its own, so each caller builds one the same
// way admin_service.go's dependencies are wired.
func newTestTxRepo(env *testEnv) repository.TransactionRepository {
	return repository.NewTransactionRepository(env.db, env.walletRepo)
}

// Customer Request #36: admin transaction review (approve/reject), extended
// to cover withdrawals explicitly, with an optional attachment URL and a
// server-side (not just client-side) guard that rejection always requires a
// note. Reuses the exact same AdminService.ValidateTransaction /
// transactionRepo.UpdateStatus path already proven for deposits -- no second
// review system. These tests exercise withdrawals specifically since
// deposit approve/reject was already covered by TestDeposit_* above.

func TestWithdrawalReview_ApproveWithNoNote(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST W-REVIEW APPROVE NO NOTE USER")
	admin := createTestAdmin(t, env, "TEST W-REVIEW APPROVE NO NOTE ADMIN")
	creditWallet(t, env, user.ID, decimal.NewFromInt(1000))
	walletSvc := newWalletSvc(env)
	adminSvc := newTestAdminService(t, env)
	txRepo := newTestTxRepo(env)

	tx, err := walletSvc.RequestWithdraw(ctx, user.ID, decimal.NewFromInt(300), "bankily", "22247601175")
	if err != nil {
		t.Fatalf("RequestWithdraw failed: %v", err)
	}

	if err := adminSvc.ValidateTransaction(ctx, tx.ID, true, "", admin.ID, ""); err != nil {
		t.Fatalf("expected approve-with-no-note to succeed, got: %v", err)
	}

	fetched, err := txRepo.GetByID(ctx, tx.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if fetched.Status != "completed" {
		t.Fatalf("expected status 'completed', got %q", fetched.Status)
	}
}

func TestWithdrawalReview_ApproveWithNote(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST W-REVIEW APPROVE NOTE USER")
	admin := createTestAdmin(t, env, "TEST W-REVIEW APPROVE NOTE ADMIN")
	creditWallet(t, env, user.ID, decimal.NewFromInt(1000))
	walletSvc := newWalletSvc(env)
	adminSvc := newTestAdminService(t, env)
	txRepo := newTestTxRepo(env)

	tx, err := walletSvc.RequestWithdraw(ctx, user.ID, decimal.NewFromInt(300), "bankily", "22247601175")
	if err != nil {
		t.Fatalf("RequestWithdraw failed: %v", err)
	}

	if err := adminSvc.ValidateTransaction(ctx, tx.ID, true, "تم التحقق من التحويل", admin.ID, ""); err != nil {
		t.Fatalf("expected approve-with-note to succeed, got: %v", err)
	}

	fetched, err := txRepo.GetByID(ctx, tx.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if fetched.AdminNotes == nil || *fetched.AdminNotes != "تم التحقق من التحويل" {
		t.Fatalf("expected admin_notes to be persisted, got %v", fetched.AdminNotes)
	}
}

// Rejection without a note: the HTTP 400 guard itself lives in
// admin_handler.go (see handlers.TestValidateTransaction_RejectWithoutNote_Rejected)
// and is proven end-to-end there. This test proves the complementary
// service-layer guarantee: a withdrawal that is never actually reviewed
// (because the malformed request never reached ValidateTransaction) still
// has fully untouched wallet state -- the request stays frozen exactly as
// RequestWithdraw left it, never silently released or captured.
func TestWithdrawalReview_UnreviewedWithdrawalWalletStateUntouched(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST W-REVIEW UNREVIEWED USER")
	creditWallet(t, env, user.ID, decimal.NewFromInt(1000))
	walletSvc := newWalletSvc(env)

	if _, err := walletSvc.RequestWithdraw(ctx, user.ID, decimal.NewFromInt(300), "bankily", "22247601175"); err != nil {
		t.Fatalf("RequestWithdraw failed: %v", err)
	}

	wallet, err := env.walletRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to read wallet: %v", err)
	}
	if !wallet.Balance.Equal(decimal.NewFromInt(700)) || !wallet.FrozenAmount.Equal(decimal.NewFromInt(300)) {
		t.Fatalf("expected balance=700/frozen=300 while pending review, got balance=%s/frozen=%s", wallet.Balance, wallet.FrozenAmount)
	}
}

func TestWithdrawalReview_RejectWithNote(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST W-REVIEW REJECT NOTE USER")
	admin := createTestAdmin(t, env, "TEST W-REVIEW REJECT NOTE ADMIN")
	creditWallet(t, env, user.ID, decimal.NewFromInt(1000))
	walletSvc := newWalletSvc(env)
	adminSvc := newTestAdminService(t, env)
	txRepo := newTestTxRepo(env)

	tx, err := walletSvc.RequestWithdraw(ctx, user.ID, decimal.NewFromInt(300), "bankily", "22247601175")
	if err != nil {
		t.Fatalf("RequestWithdraw failed: %v", err)
	}

	if err := adminSvc.ValidateTransaction(ctx, tx.ID, false, "رقم مستفيد غير صحيح", admin.ID, ""); err != nil {
		t.Fatalf("expected reject-with-note to succeed, got: %v", err)
	}

	fetched, err := txRepo.GetByID(ctx, tx.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if fetched.Status != "rejected" {
		t.Fatalf("expected status 'rejected', got %q", fetched.Status)
	}
	if fetched.AdminNotes == nil || *fetched.AdminNotes != "رقم مستفيد غير صحيح" {
		t.Fatalf("expected admin_notes to be persisted, got %v", fetched.AdminNotes)
	}
}

func TestWithdrawalReview_OptionalAttachmentPersists(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST W-REVIEW ATTACHMENT USER")
	admin := createTestAdmin(t, env, "TEST W-REVIEW ATTACHMENT ADMIN")
	creditWallet(t, env, user.ID, decimal.NewFromInt(1000))
	walletSvc := newWalletSvc(env)
	adminSvc := newTestAdminService(t, env)
	txRepo := newTestTxRepo(env)

	tx, err := walletSvc.RequestWithdraw(ctx, user.ID, decimal.NewFromInt(300), "bankily", "22247601175")
	if err != nil {
		t.Fatalf("RequestWithdraw failed: %v", err)
	}

	attachmentURL := "https://r2.example.com/transaction-review/evidence.png"
	if err := adminSvc.ValidateTransaction(ctx, tx.ID, true, "verified via screenshot", admin.ID, attachmentURL); err != nil {
		t.Fatalf("ValidateTransaction with attachment failed: %v", err)
	}

	fetched, err := txRepo.GetByID(ctx, tx.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if fetched.AdminAttachmentURL == nil || *fetched.AdminAttachmentURL != attachmentURL {
		t.Fatalf("expected admin_attachment_url to be persisted, got %v", fetched.AdminAttachmentURL)
	}
}

// No attachment provided must store NULL, not an empty string.
func TestWithdrawalReview_NoAttachmentStaysNil(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST W-REVIEW NO ATTACHMENT USER")
	admin := createTestAdmin(t, env, "TEST W-REVIEW NO ATTACHMENT ADMIN")
	creditWallet(t, env, user.ID, decimal.NewFromInt(1000))
	walletSvc := newWalletSvc(env)
	adminSvc := newTestAdminService(t, env)
	txRepo := newTestTxRepo(env)

	tx, err := walletSvc.RequestWithdraw(ctx, user.ID, decimal.NewFromInt(300), "bankily", "22247601175")
	if err != nil {
		t.Fatalf("RequestWithdraw failed: %v", err)
	}

	if err := adminSvc.ValidateTransaction(ctx, tx.ID, true, "", admin.ID, ""); err != nil {
		t.Fatalf("ValidateTransaction failed: %v", err)
	}

	fetched, err := txRepo.GetByID(ctx, tx.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if fetched.AdminAttachmentURL != nil {
		t.Fatalf("expected a nil admin_attachment_url when none was provided, got %q", *fetched.AdminAttachmentURL)
	}
}

// Wallet atomicity/frozen-balance handling (Customer #34/#31 pattern):
// approval captures the frozen amount (never touches balance a second
// time); rejection releases the frozen amount back to balance.
func TestWithdrawalReview_WalletStateCorrectAfterApprove(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST W-REVIEW WALLET APPROVE USER")
	admin := createTestAdmin(t, env, "TEST W-REVIEW WALLET APPROVE ADMIN")
	creditWallet(t, env, user.ID, decimal.NewFromInt(1000))
	walletSvc := newWalletSvc(env)
	adminSvc := newTestAdminService(t, env)

	tx, err := walletSvc.RequestWithdraw(ctx, user.ID, decimal.NewFromInt(300), "bankily", "22247601175")
	if err != nil {
		t.Fatalf("RequestWithdraw failed: %v", err)
	}

	afterRequest, err := env.walletRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to read wallet after request: %v", err)
	}
	if !afterRequest.Balance.Equal(decimal.NewFromInt(700)) || !afterRequest.FrozenAmount.Equal(decimal.NewFromInt(300)) {
		t.Fatalf("expected balance=700/frozen=300 after freeze, got balance=%s/frozen=%s", afterRequest.Balance, afterRequest.FrozenAmount)
	}

	if err := adminSvc.ValidateTransaction(ctx, tx.ID, true, "", admin.ID, ""); err != nil {
		t.Fatalf("ValidateTransaction(approve) failed: %v", err)
	}

	afterApprove, err := env.walletRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to read wallet after approve: %v", err)
	}
	if !afterApprove.Balance.Equal(decimal.NewFromInt(700)) {
		t.Fatalf("CRITICAL: balance changed on withdrawal approval (must stay 700, already debited at request time), got %s", afterApprove.Balance)
	}
	if !afterApprove.FrozenAmount.Equal(decimal.Zero) {
		t.Fatalf("expected frozen_amount to be captured to zero after approval, got %s", afterApprove.FrozenAmount)
	}
}

func TestWithdrawalReview_WalletStateCorrectAfterReject(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST W-REVIEW WALLET REJECT USER")
	admin := createTestAdmin(t, env, "TEST W-REVIEW WALLET REJECT ADMIN")
	creditWallet(t, env, user.ID, decimal.NewFromInt(1000))
	walletSvc := newWalletSvc(env)
	adminSvc := newTestAdminService(t, env)

	tx, err := walletSvc.RequestWithdraw(ctx, user.ID, decimal.NewFromInt(300), "bankily", "22247601175")
	if err != nil {
		t.Fatalf("RequestWithdraw failed: %v", err)
	}

	if err := adminSvc.ValidateTransaction(ctx, tx.ID, false, "بيانات غير مكتملة", admin.ID, ""); err != nil {
		t.Fatalf("ValidateTransaction(reject) failed: %v", err)
	}

	afterReject, err := env.walletRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to read wallet after reject: %v", err)
	}
	if !afterReject.Balance.Equal(decimal.NewFromInt(1000)) {
		t.Fatalf("expected balance restored to 1000 after rejection, got %s", afterReject.Balance)
	}
	if !afterReject.FrozenAmount.Equal(decimal.Zero) {
		t.Fatalf("expected frozen_amount released to zero after rejection, got %s", afterReject.FrozenAmount)
	}
}

// Customer #34's beneficiary_account must remain visible/unaffected by the
// review action (approve or reject) -- review only touches status,
// admin_notes, reviewed_by/at, and admin_attachment_url.
func TestWithdrawalReview_BeneficiaryAccountPreserved(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST W-REVIEW BENEFICIARY USER")
	admin := createTestAdmin(t, env, "TEST W-REVIEW BENEFICIARY ADMIN")
	creditWallet(t, env, user.ID, decimal.NewFromInt(1000))
	walletSvc := newWalletSvc(env)
	adminSvc := newTestAdminService(t, env)
	txRepo := newTestTxRepo(env)

	tx, err := walletSvc.RequestWithdraw(ctx, user.ID, decimal.NewFromInt(300), "bankily", "22247601175")
	if err != nil {
		t.Fatalf("RequestWithdraw failed: %v", err)
	}
	if tx.BeneficiaryAccount == nil || *tx.BeneficiaryAccount != "22247601175" {
		t.Fatalf("precondition failed: expected beneficiary_account on creation, got %v", tx.BeneficiaryAccount)
	}

	if err := adminSvc.ValidateTransaction(ctx, tx.ID, true, "", admin.ID, ""); err != nil {
		t.Fatalf("ValidateTransaction failed: %v", err)
	}

	fetched, err := txRepo.GetByID(ctx, tx.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if fetched.BeneficiaryAccount == nil || *fetched.BeneficiaryAccount != "22247601175" {
		t.Fatalf("expected beneficiary_account to survive review, got %v", fetched.BeneficiaryAccount)
	}
}

// Customer #21: withdrawal_processed notification must still fire after
// review, exactly as before this ticket's changes.
func TestWithdrawalReview_ProcessedNotificationStillFires(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST W-REVIEW NOTIF USER")
	admin := createTestAdmin(t, env, "TEST W-REVIEW NOTIF ADMIN")
	creditWallet(t, env, user.ID, decimal.NewFromInt(1000))
	walletSvc := newWalletSvc(env)
	adminSvc := newTestAdminService(t, env)

	tx, err := walletSvc.RequestWithdraw(ctx, user.ID, decimal.NewFromInt(300), "bankily", "22247601175")
	if err != nil {
		t.Fatalf("RequestWithdraw failed: %v", err)
	}

	if err := adminSvc.ValidateTransaction(ctx, tx.ID, true, "", admin.ID, "https://r2.example.com/transaction-review/x.png"); err != nil {
		t.Fatalf("ValidateTransaction failed: %v", err)
	}

	if getLatestNotificationOfType(t, env, user.ID, "withdrawal_processed") == nil {
		t.Fatalf("expected a withdrawal_processed notification after approval, found none")
	}
}

// A second-amount guard sanity check: invalid state never partially applies.
// (No partial state on failure) -- approving an already-completed
// transaction a second time must not double-capture frozen funds.
func TestWithdrawalReview_NoPartialStateOnDoubleApprove(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST W-REVIEW DOUBLE APPROVE USER")
	admin := createTestAdmin(t, env, "TEST W-REVIEW DOUBLE APPROVE ADMIN")
	creditWallet(t, env, user.ID, decimal.NewFromInt(1000))
	walletSvc := newWalletSvc(env)
	adminSvc := newTestAdminService(t, env)

	tx, err := walletSvc.RequestWithdraw(ctx, user.ID, decimal.NewFromInt(300), "bankily", "22247601175")
	if err != nil {
		t.Fatalf("RequestWithdraw failed: %v", err)
	}
	if err := adminSvc.ValidateTransaction(ctx, tx.ID, true, "", admin.ID, ""); err != nil {
		t.Fatalf("first ValidateTransaction(approve) failed: %v", err)
	}

	afterFirst, err := env.walletRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to read wallet: %v", err)
	}

	// Second approve call on an already-completed transaction must be a
	// safe no-op on wallet state (transactionRepo.UpdateStatus's
	// tx.Status != "completed" guard).
	if err := adminSvc.ValidateTransaction(ctx, tx.ID, true, "", admin.ID, ""); err != nil {
		t.Fatalf("second ValidateTransaction(approve) call errored unexpectedly: %v", err)
	}

	afterSecond, err := env.walletRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to read wallet: %v", err)
	}
	if !afterSecond.Balance.Equal(afterFirst.Balance) || !afterSecond.FrozenAmount.Equal(afterFirst.FrozenAmount) {
		t.Fatalf("CRITICAL: a second approve call changed wallet state (balance %s->%s, frozen %s->%s)",
			afterFirst.Balance, afterSecond.Balance, afterFirst.FrozenAmount, afterSecond.FrozenAmount)
	}
}

