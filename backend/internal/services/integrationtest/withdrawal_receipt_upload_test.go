//go:build integration

package integrationtest

import (
	"context"
	"testing"

	"github.com/shopspring/decimal"
)

// MAZADPAY -- withdrawal approval + payment receipt bug: this file covers
// the notification data-payload assertion (transaction_id, used by the
// mobile app's withdrawal_processed notification routing fix -- see
// notification_handler.dart) that admin_transaction_review_test.go's
// TestWithdrawalReview_ProcessedNotificationStillFires didn't yet assert on.
// The actual upload/magic-byte fix itself is covered at the MediaService
// level (media_service_upload_test.go) and the handler level
// (transaction_upload_test.go) -- this integration test proves the
// downstream persistence + notification contract once a real attachment URL
// (as MediaService.UploadFile would now return for a genuine screenshot) is
// submitted through the real ValidateTransaction path.

func TestWithdrawalReview_NotificationContainsTransactionID(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST W-REVIEW NOTIF TXNID USER")
	admin := createTestAdmin(t, env, "TEST W-REVIEW NOTIF TXNID ADMIN")
	creditWallet(t, env, user.ID, decimal.NewFromInt(1000))
	walletSvc := newWalletSvc(env)
	adminSvc := newTestAdminService(t, env)

	tx, err := walletSvc.RequestWithdraw(ctx, user.ID, decimal.NewFromInt(300), "bankily", "22247601175")
	if err != nil {
		t.Fatalf("RequestWithdraw failed: %v", err)
	}

	attachmentURL := "https://r2.example.com/transaction-review/evidence.png"
	if err := adminSvc.ValidateTransaction(ctx, tx.ID, true, "", admin.ID, attachmentURL); err != nil {
		t.Fatalf("ValidateTransaction failed: %v", err)
	}

	notif := getLatestNotificationOfType(t, env, user.ID, "withdrawal_processed")
	if notif == nil {
		t.Fatalf("expected a withdrawal_processed notification after approval, found none")
	}
	gotTxID, ok := notif.Data["transaction_id"].(string)
	if !ok || gotTxID != tx.ID.String() {
		t.Fatalf("expected notification data.transaction_id=%s, got %v", tx.ID.String(), notif.Data["transaction_id"])
	}
}

// End-to-end proof that an attachment URL shaped exactly like what
// MediaService.UploadFile now returns for a genuine (possibly
// extension-mismatched) screenshot flows all the way through
// ValidateTransaction into admin_attachment_url, and the withdrawal
// completes normally -- the fix doesn't just accept the upload, it doesn't
// break anything downstream of it either.
func TestWithdrawalReview_UploadedAttachmentURL_CompletesApprovalNormally(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST W-REVIEW UPLOAD E2E USER")
	admin := createTestAdmin(t, env, "TEST W-REVIEW UPLOAD E2E ADMIN")
	creditWallet(t, env, user.ID, decimal.NewFromInt(1000))
	walletSvc := newWalletSvc(env)
	adminSvc := newTestAdminService(t, env)
	txRepo := newTestTxRepo(env)

	tx, err := walletSvc.RequestWithdraw(ctx, user.ID, decimal.NewFromInt(300), "bankily", "22247601175")
	if err != nil {
		t.Fatalf("RequestWithdraw failed: %v", err)
	}

	// Shaped exactly like a real MediaService.UploadFile return value: a
	// UUID-based filename under transaction-review/, using the extension
	// derived from the REAL detected content type (see
	// imageExtensionForMimeType in media_service.go) -- not necessarily
	// whatever the client originally uploaded it as.
	uploadedURL := "https://cdn.example.com/transaction-review/3fa85f64-5717-4562-b3fc-2c963f66afa6.png"

	if err := adminSvc.ValidateTransaction(ctx, tx.ID, true, "تم التحويل", admin.ID, uploadedURL); err != nil {
		t.Fatalf("ValidateTransaction with uploaded attachment URL failed: %v", err)
	}

	fetched, err := txRepo.GetByID(ctx, tx.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if fetched.Status != "completed" {
		t.Fatalf("expected withdrawal to be completed, got status=%s", fetched.Status)
	}
	if fetched.AdminAttachmentURL == nil || *fetched.AdminAttachmentURL != uploadedURL {
		t.Fatalf("expected admin_attachment_url=%s, got %v", uploadedURL, fetched.AdminAttachmentURL)
	}

	wallet, err := env.walletRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to read wallet: %v", err)
	}
	if !wallet.Balance.Equal(decimal.NewFromInt(700)) {
		t.Fatalf("expected balance=700 (1000 - 300 withdrawn, deducted exactly once at request time), got %s", wallet.Balance)
	}
	if !wallet.FrozenAmount.Equal(decimal.Zero) {
		t.Fatalf("expected frozen_amount=0 after approval capture, got %s", wallet.FrozenAmount)
	}
}
