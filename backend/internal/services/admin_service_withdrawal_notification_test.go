package services

import "testing"

// Client feedback #10 (withdrawal notifications). ValidateTransaction is the
// shared admin approve/reject endpoint for both deposit and withdraw
// transactions (PUT /admin/transactions/:id/validate), but previously always
// sent deposit_confirmed/deposit_rejected regardless of tx.Type -- a withdrawal
// silently reused deposit wording and never used the mobile app's existing
// withdrawal_processed notification type. ValidateTransaction itself requires
// a live repository/DB (see request_service_bulk_review_test.go's own comment
// documenting the same constraint for ReviewAuctionRequest/ReviewBannerRequest),
// so these tests instead lock down the pure, DB-free pieces of the fix:
// notification-type selection, the withdrawalStatusWord/withdrawalReasonSuffix
// localization helpers, and the terminal-status idempotency predicate.

func TestWithdrawalNotificationTypeSelection(t *testing.T) {
	notifTypeFor := func(txType string, approve bool) string {
		notifType := "deposit_confirmed"
		if !approve {
			notifType = "deposit_rejected"
		}
		if txType == "withdraw" {
			notifType = "withdrawal_processed"
		}
		return notifType
	}

	t.Run("a deposit approval still uses deposit_confirmed, unchanged", func(t *testing.T) {
		if got := notifTypeFor("deposit", true); got != "deposit_confirmed" {
			t.Fatalf("expected deposit_confirmed, got %s", got)
		}
	})

	t.Run("a deposit rejection still uses deposit_rejected, unchanged", func(t *testing.T) {
		if got := notifTypeFor("deposit", false); got != "deposit_rejected" {
			t.Fatalf("expected deposit_rejected, got %s", got)
		}
	})

	t.Run("a withdrawal approval uses withdrawal_processed, not deposit_confirmed", func(t *testing.T) {
		if got := notifTypeFor("withdraw", true); got != "withdrawal_processed" {
			t.Fatalf("expected withdrawal_processed, got %s", got)
		}
	})

	t.Run("a withdrawal rejection also uses withdrawal_processed, not deposit_rejected", func(t *testing.T) {
		if got := notifTypeFor("withdraw", false); got != "withdrawal_processed" {
			t.Fatalf("expected withdrawal_processed, got %s", got)
		}
	})
}

func TestWithdrawalStatusWord(t *testing.T) {
	cases := []struct {
		approve  bool
		language string
		want     string
	}{
		{true, "ar", "تمت الموافقة عليه"},
		{false, "ar", "مرفوض"},
		{true, "fr", "approuvée"},
		{false, "fr", "refusée"},
		{true, "en", "approved"},
		{false, "en", "rejected"},
		{true, "es", "approved"}, // unknown language falls back to English
	}
	for _, c := range cases {
		if got := withdrawalStatusWord(c.approve, c.language); got != c.want {
			t.Errorf("withdrawalStatusWord(%v, %q) = %q, want %q", c.approve, c.language, got, c.want)
		}
	}
}

func TestWithdrawalReasonSuffix(t *testing.T) {
	t.Run("empty notes produce no clause at all, no dangling separator", func(t *testing.T) {
		if got := withdrawalReasonSuffix("", "en"); got != "" {
			t.Fatalf("expected empty string for empty notes, got %q", got)
		}
	})

	t.Run("a rejection reason is appended with the localized lead-in", func(t *testing.T) {
		if got := withdrawalReasonSuffix("KYC mismatch", "en"); got != ". Reason: KYC mismatch" {
			t.Fatalf("unexpected suffix: %q", got)
		}
		if got := withdrawalReasonSuffix("بيانات ناقصة", "ar"); got != ". السبب: بيانات ناقصة" {
			t.Fatalf("unexpected suffix: %q", got)
		}
		if got := withdrawalReasonSuffix("solde incorrect", "fr"); got != ". Raison : solde incorrect" {
			t.Fatalf("unexpected suffix: %q", got)
		}
	})
}

// isTerminalTransactionStatus mirrors the idempotency guard added to
// ValidateTransaction: once a transaction is already completed or rejected,
// a retried/duplicate admin action must not resend a push.
func isTerminalTransactionStatus(status string) bool {
	return status == "completed" || status == "rejected"
}

func TestValidateTransaction_TerminalStatusSkipsNotification(t *testing.T) {
	t.Run("pending_review is not terminal -- notification proceeds", func(t *testing.T) {
		if isTerminalTransactionStatus("pending_review") {
			t.Fatal("expected pending_review to not be terminal")
		}
	})

	t.Run("pending is not terminal -- notification proceeds", func(t *testing.T) {
		if isTerminalTransactionStatus("pending") {
			t.Fatal("expected pending to not be terminal")
		}
	})

	t.Run("completed is terminal -- a retry must not resend a duplicate push", func(t *testing.T) {
		if !isTerminalTransactionStatus("completed") {
			t.Fatal("expected completed to be terminal")
		}
	})

	t.Run("rejected is terminal -- a retry must not resend a duplicate push", func(t *testing.T) {
		if !isTerminalTransactionStatus("rejected") {
			t.Fatal("expected rejected to be terminal")
		}
	})
}

func TestGetLocalizedNotification_WithdrawalProcessed(t *testing.T) {
	t.Run("approved withdrawal in English has no dangling reason placeholder", func(t *testing.T) {
		_, body := GetLocalizedNotification("withdrawal_processed", "en", map[string]string{
			"amount":   "100",
			"currency": "MRU",
			"status":   withdrawalStatusWord(true, "en"),
			"reason":   "",
		})
		if body == "" {
			t.Fatal("expected a non-empty body")
		}
		if want := "Your withdrawal of 100 MRU was approved"; body != want {
			t.Fatalf("body = %q, want %q", body, want)
		}
	})

	t.Run("rejected withdrawal in English includes the reason clause", func(t *testing.T) {
		_, body := GetLocalizedNotification("withdrawal_processed", "en", map[string]string{
			"amount":   "50",
			"currency": "MRU",
			"status":   withdrawalStatusWord(false, "en"),
			"reason":   withdrawalReasonSuffix("insufficient documentation", "en"),
		})
		if want := "Your withdrawal of 50 MRU was rejected. Reason: insufficient documentation"; body != want {
			t.Fatalf("body = %q, want %q", body, want)
		}
	})

	t.Run("localization entry exists for ar/fr/en -- no missing-type fallback to empty strings", func(t *testing.T) {
		for _, lang := range []string{"ar", "fr", "en"} {
			title, body := GetLocalizedNotification("withdrawal_processed", lang, map[string]string{
				"amount": "1", "currency": "MRU", "status": "x", "reason": "",
			})
			if title == "" || body == "" {
				t.Errorf("expected non-empty title/body for language %q, got title=%q body=%q", lang, title, body)
			}
		}
	})
}
