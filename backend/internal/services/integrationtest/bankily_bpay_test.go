//go:build integration

package integrationtest

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/config"
	apperr "github.com/mazadpay/backend/internal/errors"
	"github.com/mazadpay/backend/internal/handlers"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/payments/bankily"
	"github.com/mazadpay/backend/internal/payments/bankily/bankilytest"
	"github.com/mazadpay/backend/internal/repository"
	"github.com/mazadpay/backend/internal/services"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

const bpayPasscode = "PASSCODE7731"

type bpayEnv struct {
	env  *testEnv
	fake *bankilytest.Server
	svc  services.BankilyPaymentService
	user *models.User
}

func newBpayEnv(t *testing.T, clientTimeout time.Duration) *bpayEnv {
	t.Helper()
	env := setupEnv(t)
	fake := bankilytest.New()
	t.Cleanup(fake.Close)
	cfg := config.BankilyConfig{
		Enabled: true, BaseURL: fake.URL, Username: "u", Password: "p", ClientID: "ebankily",
		MerchantCode: "027922", HTTPTimeout: clientTimeout, OperationIDPrefix: "ITEST", OperationIDDigits: 13,
		MinAmount: 1, MaxAmount: 100000, ReconcileInterval: time.Second, MaxPendingAge: 72 * time.Hour,
	}
	client := bankily.NewClient(bankily.Config{BaseURL: fake.URL, Username: "u", Password: "p", ClientID: "ebankily", Timeout: clientTimeout}, nil)
	auditSvc := services.NewAuditService(repository.NewAuditRepository(env.db))
	svc := services.NewBankilyPaymentService(env.db, cfg, client, repository.NewBankilyPaymentRepository(env.db), env.walletRepo, env.userRepo, nil, auditSvc, zaptest.NewLogger(t, zaptest.Level(zap.WarnLevel)))
	user := createTestUser(t, env, "TEST BPAY USER")
	if _, err := env.walletRepo.GetByUserID(context.Background(), user.ID); err != nil { // create MRU wallet
		t.Fatal(err)
	}
	return &bpayEnv{env: env, fake: fake, svc: svc, user: user}
}

func (b *bpayEnv) pay(t *testing.T, amount int64, idem string) (*models.BankilyPayment, error) {
	t.Helper()
	return b.svc.Pay(context.Background(), b.user.ID, services.BankilyPayInput{
		Amount: decimal.NewFromInt(amount), ClientPhone: "48110011", Passcode: bpayPasscode, IdempotencyKey: idem, Language: "fr",
	})
}

func (b *bpayEnv) balance(t *testing.T) decimal.Decimal {
	bal, _ := walletBalanceAndFrozen(t, b.env, b.user.ID)
	return bal
}

func (b *bpayEnv) ledgerRows(t *testing.T, op string) int {
	var n int
	if err := b.env.db.Get(&n, `SELECT COUNT(*) FROM transactions WHERE reference = $1 AND gateway = 'bankily_bpay'`, "bankily:"+op); err != nil {
		t.Fatal(err)
	}
	return n
}

func (b *bpayEnv) reload(t *testing.T, id uuid.UUID) *models.BankilyPayment {
	p, err := b.svc.GetForUser(context.Background(), b.user.ID, id)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// reconcileOnlyMine runs the reconciler with every other test's unresolved
// rows parked, so a shared local DB never mixes fakes.
func (b *bpayEnv) reconcileOnlyMine(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	b.env.db.MustExecContext(ctx, `UPDATE bankily_payments SET next_check_at = NOW() + INTERVAL '1 day'
		WHERE user_id <> $1 AND status IN ('SUBMITTED','PENDING_VERIFICATION')`, b.user.ID)
	b.env.db.MustExecContext(ctx, `UPDATE bankily_payments SET next_check_at = NOW() - INTERVAL '1 second'
		WHERE user_id = $1 AND status IN ('SUBMITTED','PENDING_VERIFICATION')`, b.user.ID)
	if err := b.svc.ReconcileOnce(ctx); err != nil {
		t.Fatal(err)
	}
}

func uniqKey() string { return "k-" + uuid.NewString() }

// 1. TS -> exactly one wallet credit.
func TestBankily_TS_CreditsExactlyOnce(t *testing.T) {
	b := newBpayEnv(t, 5*time.Second)
	before := b.balance(t)
	p, err := b.pay(t, 120, uniqKey())
	if err != nil {
		t.Fatal(err)
	}
	if p.Status != models.BankilyStatusSuccess || p.WalletTransactionID == nil {
		t.Fatalf("expected SUCCESS with ledger row, got %s", p.Status)
	}
	if got := b.balance(t).Sub(before); !got.Equal(decimal.NewFromInt(120)) {
		t.Fatalf("balance delta %s, want 120", got)
	}
	if n := b.ledgerRows(t, p.OperationID); n != 1 {
		t.Fatalf("ledger rows %d, want 1", n)
	}
	// Re-checking a settled payment changes nothing.
	if _, err := b.svc.Recheck(context.Background(), p.ID, "user", b.user.ID); err != nil {
		t.Fatal(err)
	}
	b.reconcileOnlyMine(t)
	if got := b.balance(t).Sub(before); !got.Equal(decimal.NewFromInt(120)) || b.ledgerRows(t, p.OperationID) != 1 {
		t.Fatalf("re-check/reconcile must not credit again: delta=%s", got)
	}
	var status string
	b.env.db.Get(&status, `SELECT status FROM transactions WHERE id = $1`, *p.WalletTransactionID)
	if status != "completed" {
		t.Fatalf("ledger row must be completed (no manual approval), got %s", status)
	}
	var audits int
	b.env.db.Get(&audits, `SELECT COUNT(*) FROM audit_logs WHERE entity_type='bankily_payment' AND entity_id=$1 AND action='bankily_payment_credited'`, p.ID)
	if audits != 1 {
		t.Fatalf("expected 1 credited audit entry, got %d", audits)
	}
}

// 2. TF -> zero wallet credit.
func TestBankily_TF_NoCredit(t *testing.T) {
	b := newBpayEnv(t, 5*time.Second)
	b.fake.SetDefaultStatus("TF")
	before := b.balance(t)
	p, err := b.pay(t, 500, uniqKey())
	if err != nil {
		t.Fatal(err)
	}
	if p.Status != models.BankilyStatusFailed || p.WalletTransactionID != nil {
		t.Fatalf("expected FAILED, got %s", p.Status)
	}
	if !b.balance(t).Equal(before) || b.ledgerRows(t, p.OperationID) != 0 {
		t.Fatal("TF must not change the wallet")
	}
}

// errorCode 1 and Bankily has no record of the operation -> definite failure.
func TestBankily_PaymentRejected_NotFound_Fails(t *testing.T) {
	b := newBpayEnv(t, 5*time.Second)
	b.fake.SetPaymentHook(func(w http.ResponseWriter, _ map[string]string) bool {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"errorCode":"1","errorMessage":"passcode invalide"}`))
		return true
	})
	before := b.balance(t)
	p, err := b.pay(t, 100, uniqKey())
	if err != nil {
		t.Fatal(err)
	}
	if p.Status != models.BankilyStatusFailed || p.FailureReason == nil || *p.FailureReason != "payment_rejected" {
		t.Fatalf("expected FAILED/payment_rejected, got %s", p.Status)
	}
	if !b.balance(t).Equal(before) {
		t.Fatal("rejected payment changed balance")
	}
}

// 3. TA -> pending, zero credit; later TS resolves to exactly one credit.
func TestBankily_TA_PendingThenResolved(t *testing.T) {
	b := newBpayEnv(t, 5*time.Second)
	b.fake.SetDefaultStatus("TA")
	before := b.balance(t)
	p, err := b.pay(t, 1000, uniqKey())
	if err != nil {
		t.Fatal(err)
	}
	if p.Status != models.BankilyStatusPendingVerification || !b.balance(t).Equal(before) {
		t.Fatalf("TA must stay pending with no credit, got %s", p.Status)
	}
	b.reconcileOnlyMine(t) // still TA
	if q := b.reload(t, p.ID); q.Status != models.BankilyStatusPendingVerification || q.CheckAttempts < 2 {
		t.Fatalf("still TA: status=%s attempts=%d", q.Status, q.CheckAttempts)
	}
	b.fake.SetOp(p.OperationID, "TS")
	b.reconcileOnlyMine(t)
	if q := b.reload(t, p.ID); q.Status != models.BankilyStatusSuccess {
		t.Fatalf("expected SUCCESS after TS, got %s", q.Status)
	}
	if !b.balance(t).Sub(before).Equal(decimal.NewFromInt(1000)) || b.ledgerRows(t, p.OperationID) != 1 {
		t.Fatal("expected exactly one 1000 credit")
	}
}

// 4. Payment timeout -> PENDING_VERIFICATION -> checkTransaction, never a resend.
func TestBankily_PaymentTimeout_RecoveredByCheck(t *testing.T) {
	b := newBpayEnv(t, 300*time.Millisecond)
	b.fake.SetPaymentHook(func(w http.ResponseWriter, _ map[string]string) bool {
		time.Sleep(700 * time.Millisecond) // Bankily still processes it
		return false
	})
	before := b.balance(t)
	p, err := b.pay(t, 4500, uniqKey())
	if err != nil {
		t.Fatal(err)
	}
	if p.Status != models.BankilyStatusPendingVerification {
		t.Fatalf("timeout must be PENDING_VERIFICATION, got %s", p.Status)
	}
	if !b.balance(t).Equal(before) {
		t.Fatal("timeout must not credit")
	}
	time.Sleep(800 * time.Millisecond)
	b.fake.SetPaymentHook(nil)
	b.reconcileOnlyMine(t)
	if q := b.reload(t, p.ID); q.Status != models.BankilyStatusSuccess {
		t.Fatalf("expected SUCCESS after check, got %s", q.Status)
	}
	if b.fake.PaymentCalls.Load() != 1 {
		t.Fatalf("payment must never be resent after a timeout, calls=%d", b.fake.PaymentCalls.Load())
	}
	if !b.balance(t).Sub(before).Equal(decimal.NewFromInt(4500)) {
		t.Fatal("expected one 4500 credit")
	}
}

// 5. checkTransaction timeout -> stays pending, retried later.
func TestBankily_CheckTimeout_StaysPendingThenRetries(t *testing.T) {
	b := newBpayEnv(t, 300*time.Millisecond)
	b.fake.SetCheckHook(func(w http.ResponseWriter, _ string) bool {
		time.Sleep(600 * time.Millisecond)
		return false
	})
	before := b.balance(t)
	p, err := b.pay(t, 200, uniqKey())
	if err != nil {
		t.Fatal(err)
	}
	if p.Status != models.BankilyStatusPendingVerification || p.CheckAttempts != 1 {
		t.Fatalf("check timeout: status=%s attempts=%d", p.Status, p.CheckAttempts)
	}
	b.reconcileOnlyMine(t)
	if q := b.reload(t, p.ID); q.Status != models.BankilyStatusPendingVerification || q.CheckAttempts != 2 || !b.balance(t).Equal(before) {
		t.Fatalf("second check timeout: status=%s attempts=%d", q.Status, q.CheckAttempts)
	}
	b.fake.SetCheckHook(nil)
	b.reconcileOnlyMine(t)
	if q := b.reload(t, p.ID); q.Status != models.BankilyStatusSuccess {
		t.Fatalf("expected SUCCESS once check answers, got %s", q.Status)
	}
}

// 6. Duplicate operationId / duplicate ledger reference rejected by the DB.
func TestBankily_DuplicateOperationID_RejectedByDB(t *testing.T) {
	b := newBpayEnv(t, 5*time.Second)
	p, err := b.pay(t, 100, uniqKey())
	if err != nil {
		t.Fatal(err)
	}
	_, err = b.env.db.Exec(`INSERT INTO bankily_payments (id,user_id,operation_id,amount,client_phone,language,status)
		VALUES ($1,$2,$3,10,'48110011','FR','INITIATED')`, uuid.New(), b.user.ID, p.OperationID)
	if err == nil || !strings.Contains(err.Error(), "uq_bankily_payments_operation_id") {
		t.Fatalf("duplicate operation_id must be rejected, got %v", err)
	}
	_, err = b.env.db.Exec(`INSERT INTO transactions (id,user_id,type,amount,gateway,status,reference)
		VALUES ($1,$2,'deposit',100,'bankily_bpay','completed',$3)`, uuid.New(), b.user.ID, "bankily:"+p.OperationID)
	if err == nil || !strings.Contains(err.Error(), "uq_transactions_bankily_reference") {
		t.Fatalf("duplicate ledger reference must be rejected, got %v", err)
	}
	_, err = b.env.db.Exec(`UPDATE bankily_payments SET status='SUCCESS' WHERE id = (
		SELECT id FROM bankily_payments WHERE user_id=$1 AND status <> 'SUCCESS' LIMIT 1)`, b.user.ID)
	if err != nil && !strings.Contains(err.Error(), "chk_bankily_payments_success_has_credit") {
		t.Fatalf("unexpected error: %v", err)
	}
	// Bankily itself refuses a reused operationId.
	client := bankily.NewClient(bankily.Config{BaseURL: b.fake.URL, Username: "u", Password: "p", ClientID: "ebankily"}, nil)
	res, err := client.Payment(context.Background(), bankily.PaymentRequest{ClientPhone: "48110011", Passcode: "x", OperationID: p.OperationID, Amount: "100", Language: "FR"})
	if err != nil || res.ErrorCode == "0" {
		t.Fatalf("reused operationId must not succeed: %+v %v", res, err)
	}
}

// 7. Duplicate client submission (same idempotency key) reuses the attempt.
func TestBankily_DuplicateSubmission_SameAttempt(t *testing.T) {
	b := newBpayEnv(t, 5*time.Second)
	before := b.balance(t)
	key := uniqKey()
	p1, err := b.pay(t, 300, key)
	if err != nil {
		t.Fatal(err)
	}
	p2, err := b.pay(t, 300, key)
	if err != nil {
		t.Fatal(err)
	}
	if p1.ID != p2.ID || b.fake.PaymentCalls.Load() != 1 {
		t.Fatalf("duplicate submission must reuse the attempt: %s vs %s, calls=%d", p1.ID, p2.ID, b.fake.PaymentCalls.Load())
	}
	if _, err := b.pay(t, 999, key); !errors.Is(err, apperr.ErrBankilyIdempotencyReused) {
		t.Fatalf("same key with different amount must be refused, got %v", err)
	}
	if !b.balance(t).Sub(before).Equal(decimal.NewFromInt(300)) {
		t.Fatal("expected a single 300 credit")
	}
}

// 8. Simultaneous confirmation paths -> one credit.
func TestBankily_ConcurrentConfirmations_SingleCredit(t *testing.T) {
	b := newBpayEnv(t, 5*time.Second)
	b.fake.SetDefaultStatus("TA")
	before := b.balance(t)
	p, err := b.pay(t, 2000, uniqKey())
	if err != nil {
		t.Fatal(err)
	}
	b.fake.SetOp(p.OperationID, "TS")
	b.env.db.MustExec(`UPDATE bankily_payments SET next_check_at = NOW() + INTERVAL '1 day' WHERE user_id <> $1 AND status IN ('SUBMITTED','PENDING_VERIFICATION')`, b.user.ID)
	b.env.db.MustExec(`UPDATE bankily_payments SET next_check_at = NOW() - INTERVAL '1 second' WHERE id = $1`, p.ID)

	var wg sync.WaitGroup
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			switch i % 3 {
			case 0:
				b.svc.Recheck(context.Background(), p.ID, "user", b.user.ID)
			case 1:
				b.svc.Recheck(context.Background(), p.ID, "admin", b.user.ID)
			default:
				b.svc.ReconcileOnce(context.Background())
			}
		}(i)
	}
	wg.Wait()
	if q := b.reload(t, p.ID); q.Status != models.BankilyStatusSuccess {
		t.Fatalf("expected SUCCESS, got %s", q.Status)
	}
	if n := b.ledgerRows(t, p.OperationID); n != 1 {
		t.Fatalf("ledger rows %d, want exactly 1", n)
	}
	if got := b.balance(t).Sub(before); !got.Equal(decimal.NewFromInt(2000)) {
		t.Fatalf("balance delta %s, want exactly 2000", got)
	}
}

// 9 + 10. Expired access token / expired refresh token -> transparent renewal.
func TestBankily_ExpiredTokens_Renewed(t *testing.T) {
	b := newBpayEnv(t, 5*time.Second)
	if _, err := b.pay(t, 100, uniqKey()); err != nil {
		t.Fatal(err)
	}
	b.fake.ExpireAccessTokens()
	p, err := b.pay(t, 100, uniqKey())
	if err != nil || p.Status != models.BankilyStatusSuccess {
		t.Fatalf("expired access token: %v %v", p, err)
	}
	b.fake.ExpireAllTokens()
	p, err = b.pay(t, 100, uniqKey())
	if err != nil || p.Status != models.BankilyStatusSuccess {
		t.Fatalf("expired refresh token: %v %v", p, err)
	}
	if b.fake.PasswordGrants.Load() < 2 {
		t.Fatal("expected a fresh login after all tokens expired")
	}
}

// 11. Malformed Bankily response -> pending, no credit, then resolved.
func TestBankily_MalformedResponse_Pending(t *testing.T) {
	b := newBpayEnv(t, 5*time.Second)
	b.fake.SetPaymentHook(func(w http.ResponseWriter, body map[string]string) bool {
		b.fake.SetOp(body["operationId"], "TS")     // Bankily processed it...
		w.Write([]byte("<html>proxy error</html>")) // ...but we can't read the answer
		return true
	})
	b.fake.SetCheckHook(func(w http.ResponseWriter, _ string) bool {
		w.Write([]byte(`{"oops":`))
		return true
	})
	before := b.balance(t)
	p, err := b.pay(t, 1444, uniqKey())
	if err != nil {
		t.Fatal(err)
	}
	if p.Status != models.BankilyStatusPendingVerification || !b.balance(t).Equal(before) {
		t.Fatalf("malformed response must stay pending with no credit, got %s", p.Status)
	}
	b.fake.SetCheckHook(nil)
	b.reconcileOnlyMine(t)
	if q := b.reload(t, p.ID); q.Status != models.BankilyStatusSuccess || !b.balance(t).Sub(before).Equal(decimal.NewFromInt(1444)) {
		t.Fatalf("expected SUCCESS with one credit, got %s", q.Status)
	}
}

// 12. Bankily 5xx -> pending, no credit; TF on check -> failed.
func TestBankily_5xx_PendingThenFailed(t *testing.T) {
	b := newBpayEnv(t, 5*time.Second)
	b.fake.SetPaymentHook(func(w http.ResponseWriter, body map[string]string) bool {
		b.fake.SetOp(body["operationId"], "TF")
		w.WriteHeader(http.StatusServiceUnavailable)
		return true
	})
	before := b.balance(t)
	p, err := b.pay(t, 4700, uniqKey())
	if err != nil {
		t.Fatal(err)
	}
	// The inline check already sees TF.
	if q := b.reload(t, p.ID); q.Status != models.BankilyStatusFailed || !b.balance(t).Equal(before) {
		t.Fatalf("expected FAILED with no credit, got %s", q.Status)
	}
}

// 5xx where Bankily has no record yet: never auto-failed, stays pending.
func TestBankily_5xx_UnknownToBankily_StaysPending(t *testing.T) {
	b := newBpayEnv(t, 5*time.Second)
	b.fake.SetPaymentHook(func(w http.ResponseWriter, _ map[string]string) bool {
		w.WriteHeader(http.StatusInternalServerError)
		return true
	})
	p, err := b.pay(t, 100, uniqKey())
	if err != nil {
		t.Fatal(err)
	}
	b.reconcileOnlyMine(t)
	if q := b.reload(t, p.ID); q.Status != models.BankilyStatusPendingVerification {
		t.Fatalf("ambiguous attempt must remain pending, got %s", q.Status)
	}
}

// 13. User double-taps Pay (concurrent, same idempotency key) -> one payment.
func TestBankily_DoubleTap_OnePayment(t *testing.T) {
	b := newBpayEnv(t, 5*time.Second)
	before := b.balance(t)
	key := uniqKey()
	var wg sync.WaitGroup
	ids := make([]uuid.UUID, 6)
	for i := range ids {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if p, err := b.pay(t, 250, key); err == nil {
				ids[i] = p.ID
			}
		}(i)
	}
	wg.Wait()
	for _, id := range ids {
		if id != ids[0] {
			t.Fatalf("double tap produced different attempts: %v", ids)
		}
	}
	var attempts int
	b.env.db.Get(&attempts, `SELECT COUNT(*) FROM bankily_payments WHERE user_id = $1 AND idempotency_key = $2`, b.user.ID, key)
	if attempts != 1 || b.fake.PaymentCalls.Load() != 1 {
		t.Fatalf("attempts=%d bankily calls=%d, want 1/1", attempts, b.fake.PaymentCalls.Load())
	}
	if !b.balance(t).Sub(before).Equal(decimal.NewFromInt(250)) {
		t.Fatal("expected a single 250 credit")
	}
}

// Admin cannot approve/reject or add-balance against an automatic row.
func TestBankily_AdminManualActionsBlocked(t *testing.T) {
	b := newBpayEnv(t, 5*time.Second)
	p, err := b.pay(t, 100, uniqKey())
	if err != nil || p.WalletTransactionID == nil {
		t.Fatal(err)
	}
	admin := createTestAdmin(t, b.env, "TEST BPAY ADMIN")
	adminSvc := newTestAdminService(t, b.env)
	before := b.balance(t)
	if err := adminSvc.ValidateTransaction(context.Background(), *p.WalletTransactionID, false, "stale reject", admin.ID, ""); !errors.Is(err, apperr.ErrAutomaticTransaction) {
		t.Fatalf("reject must be refused, got %v", err)
	}
	if err := adminSvc.ValidateTransaction(context.Background(), *p.WalletTransactionID, true, "", admin.ID, ""); !errors.Is(err, apperr.ErrAutomaticTransaction) {
		t.Fatalf("approve must be refused, got %v", err)
	}
	if _, err := adminSvc.AdminAddBalance(context.Background(), *p.WalletTransactionID, decimal.NewFromInt(100), "", admin.ID); !errors.Is(err, apperr.ErrAutomaticTransaction) {
		t.Fatalf("add-balance must be refused, got %v", err)
	}
	if !b.balance(t).Equal(before) {
		t.Fatal("admin actions changed the balance")
	}
}

// Input validation and the passcode is never persisted anywhere.
func TestBankily_ValidationAndNoPasscodeStored(t *testing.T) {
	b := newBpayEnv(t, 5*time.Second)
	ctx := context.Background()
	bad := []services.BankilyPayInput{
		{Amount: decimal.NewFromInt(100), ClientPhone: "12345678", Passcode: "1234"},
		{Amount: decimal.NewFromInt(100), ClientPhone: "48110011", Passcode: ""},
		{Amount: decimal.NewFromInt(0), ClientPhone: "48110011", Passcode: "1234"},
		{Amount: decimal.RequireFromString("100.5"), ClientPhone: "48110011", Passcode: "1234"},
		{Amount: decimal.NewFromInt(200000), ClientPhone: "48110011", Passcode: "1234"},
	}
	for i, in := range bad {
		in.IdempotencyKey = uniqKey() // so each case exercises its own validation
		if _, err := b.svc.Pay(ctx, b.user.ID, in); err == nil || errors.Is(err, apperr.ErrBankilyIdempotencyRequired) {
			t.Fatalf("case %d must be rejected", i)
		}
	}
	if b.fake.PaymentCalls.Load() != 0 {
		t.Fatal("invalid input reached Bankily")
	}
	b.env.db.MustExec(`UPDATE wallets SET is_disabled = TRUE WHERE user_id = $1`, b.user.ID)
	if _, err := b.pay(t, 100, uniqKey()); !errors.Is(err, apperr.ErrWalletDisabled) {
		t.Fatalf("disabled wallet must be refused, got %v", err)
	}
	b.env.db.MustExec(`UPDATE wallets SET is_disabled = FALSE WHERE user_id = $1`, b.user.ID)
	if _, err := b.pay(t, 100, uniqKey()); err != nil {
		t.Fatal(err)
	}
	var hits int
	b.env.db.Get(&hits, `SELECT COUNT(*) FROM bankily_payments WHERE row_to_json(bankily_payments)::text LIKE '%' || $1 || '%'`, bpayPasscode)
	var auditHits int
	b.env.db.Get(&auditHits, `SELECT COUNT(*) FROM audit_logs WHERE row_to_json(audit_logs)::text LIKE '%' || $1 || '%'`, bpayPasscode)
	var phoneHits int
	b.env.db.Get(&phoneHits, `SELECT COUNT(*) FROM audit_logs WHERE entity_type='bankily_payment' AND row_to_json(audit_logs)::text LIKE '%48110011%'`)
	if hits != 0 || auditHits != 0 || phoneHits != 0 {
		t.Fatalf("passcode/phone leaked: payments=%d audit=%d auditPhone=%d", hits, auditHits, phoneHits)
	}
}

// HTTP contract: masked phone, no passcode echoed, admin list, user isolation.
func TestBankily_HTTPContract(t *testing.T) {
	b := newBpayEnv(t, 5*time.Second)
	h := handlers.NewBankilyHandler(b.svc, b.env.userRepo, b.env.logger)
	app := fiber.New()
	uid := b.user.ID
	app.Post("/pay", fakeAuth(&uid), h.Pay)
	app.Get("/payments/:id", fakeAuth(&uid), h.Get)
	app.Get("/admin/payments", h.AdminList)
	other := uuid.New()
	app.Get("/other/payments/:id", fakeAuth(&other), h.Get)

	body := `{"amount":"150","client_phone":"48110011","passcode":"` + bpayPasscode + `","idempotency_key":"` + uniqKey() + `"}`
	req := httptest.NewRequest("POST", "/pay", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, 10000)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 || strings.Contains(string(raw), bpayPasscode) || strings.Contains(string(raw), "48110011") {
		t.Fatalf("pay response %d leaks or failed: %s", resp.StatusCode, raw)
	}
	var out struct {
		Data struct {
			ID          string `json:"id"`
			Status      string `json:"status"`
			ClientPhone string `json:"client_phone"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &out); err != nil || out.Data.Status != "SUCCESS" || out.Data.ClientPhone != "48****11" {
		t.Fatalf("unexpected pay body: %s", raw)
	}
	if r, _ := app.Test(httptest.NewRequest("GET", "/other/payments/"+out.Data.ID, nil)); r.StatusCode != 404 {
		t.Fatalf("another user must not see this payment, got %d", r.StatusCode)
	}
	r, _ := app.Test(httptest.NewRequest("GET", "/admin/payments?status=SUCCESS", nil))
	adminRaw, _ := io.ReadAll(r.Body)
	if r.StatusCode != 200 || !strings.Contains(string(adminRaw), out.Data.ID) || strings.Contains(string(adminRaw), bpayPasscode) {
		t.Fatalf("admin list: %d %s", r.StatusCode, adminRaw)
	}
	if r, _ := app.Test(httptest.NewRequest("GET", "/admin/payments?status=BOGUS", nil)); r.StatusCode != 400 {
		t.Fatalf("invalid status filter must be 400, got %d", r.StatusCode)
	}
}

// errorCode 5 ("Operation ID already existe", undocumented): the attempt goes
// to NEEDS_REVIEW and no automatic path ever credits it -- even when Bankily
// later reports TS for that operationId (it may be someone else's payment).
func TestBankily_ErrorCode5_NeedsReview_NeverAutoCredits(t *testing.T) {
	b := newBpayEnv(t, 5*time.Second)
	b.fake.SetPaymentHook(func(w http.ResponseWriter, body map[string]string) bool {
		b.fake.SetOp(body["operationId"], "TS") // an existing, successful operation
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"errorCode":"5","errorMessage":"Operation ID already existe"}`))
		return true
	})
	before := b.balance(t)
	p, err := b.pay(t, 700, uniqKey())
	if err != nil {
		t.Fatal(err)
	}
	if p.Status != models.BankilyStatusNeedsReview || !p.NeedsReview {
		t.Fatalf("errorCode 5 must go to NEEDS_REVIEW, got %s needs_review=%v", p.Status, p.NeedsReview)
	}
	if p.BankilyErrorCode == nil || *p.BankilyErrorCode != "5" || p.FailureReason == nil || *p.FailureReason != "operation_id_already_exists" ||
		p.BankilyErrorMessage == nil || *p.BankilyErrorMessage != "Operation ID already existe" {
		t.Fatalf("Bankily metadata not preserved: %+v", p)
	}
	if b.fake.CheckCalls.Load() != 0 {
		t.Fatal("errorCode 5 must not continue into automatic checkTransaction resolution")
	}

	// Repeated user/admin re-checks, the reconciler, and concurrent mixes:
	// Bankily keeps answering TS, yet nothing credits.
	for i := 0; i < 3; i++ {
		if _, err := b.svc.Recheck(context.Background(), p.ID, "user", b.user.ID); err != nil {
			t.Fatal(err)
		}
		if _, err := b.svc.Recheck(context.Background(), p.ID, "admin", b.user.ID); err != nil {
			t.Fatal(err)
		}
		b.reconcileOnlyMine(t)
	}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); b.svc.Recheck(context.Background(), p.ID, "admin", b.user.ID) }()
	}
	wg.Wait()

	q := b.reload(t, p.ID)
	if q.Status != models.BankilyStatusNeedsReview || q.WalletTransactionID != nil {
		t.Fatalf("review attempt changed: status=%s", q.Status)
	}
	if q.BankilyStatus == nil || *q.BankilyStatus != "TS" || q.CheckAttempts != 14 {
		t.Fatalf("informational checks not recorded: bankily_status=%v attempts=%d", q.BankilyStatus, q.CheckAttempts)
	}
	if q.BankilyErrorCode == nil || *q.BankilyErrorCode != "5" {
		t.Fatal("re-checks must not overwrite the original errorCode 5")
	}
	if !b.balance(t).Equal(before) || b.ledgerRows(t, p.OperationID) != 0 {
		t.Fatal("errorCode 5 attempt credited the wallet")
	}
	if b.fake.PaymentCalls.Load() != 1 {
		t.Fatalf("payment must not be resent, calls=%d", b.fake.PaymentCalls.Load())
	}

	// Exactly one audit entry records the move to manual review.
	var reviewAudits int
	b.env.db.Get(&reviewAudits, `SELECT COUNT(*) FROM audit_logs WHERE entity_id=$1 AND action='bankily_payment_needs_review'`, p.ID)
	if reviewAudits != 1 {
		t.Fatalf("expected one needs_review audit entry, got %d", reviewAudits)
	}

	// Admin sees it under the review filter; no ledger row exists to approve.
	h := handlers.NewBankilyHandler(b.svc, b.env.userRepo, b.env.logger)
	app := fiber.New()
	app.Get("/admin/payments", h.AdminList)
	r, _ := app.Test(httptest.NewRequest("GET", "/admin/payments?status=NEEDS_REVIEW&per_page=100", nil))
	raw, _ := io.ReadAll(r.Body)
	if r.StatusCode != 200 || !strings.Contains(string(raw), p.ID.String()) || !strings.Contains(string(raw), `"needs_review":true`) {
		t.Fatalf("admin review list: %d %s", r.StatusCode, raw)
	}
}

// The idempotency key is mandatory: no key -> typed 400, no Bankily call,
// no attempt row.
func TestBankily_IdempotencyKeyRequired(t *testing.T) {
	b := newBpayEnv(t, 5*time.Second)
	count := func() int {
		var n int
		b.env.db.Get(&n, `SELECT COUNT(*) FROM bankily_payments WHERE user_id = $1`, b.user.ID)
		return n
	}
	if _, err := b.pay(t, 100, ""); !errors.Is(err, apperr.ErrBankilyIdempotencyRequired) {
		t.Fatalf("service: expected ErrBankilyIdempotencyRequired, got %v", err)
	}

	h := handlers.NewBankilyHandler(b.svc, b.env.userRepo, b.env.logger)
	app := fiber.New()
	uid := b.user.ID
	app.Post("/pay", fakeAuth(&uid), h.Pay)
	req := httptest.NewRequest("POST", "/pay", strings.NewReader(`{"amount":"100","client_phone":"48110011","passcode":"`+bpayPasscode+`"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, 10000)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 400 || !strings.Contains(string(raw), `"bankily_idempotency_key_required"`) {
		t.Fatalf("HTTP: expected 400 bankily_idempotency_key_required, got %d %s", resp.StatusCode, raw)
	}
	if b.fake.PaymentCalls.Load() != 0 || b.fake.PasswordGrants.Load() != 0 {
		t.Fatalf("missing key reached Bankily: payments=%d logins=%d", b.fake.PaymentCalls.Load(), b.fake.PasswordGrants.Load())
	}
	if n := count(); n != 0 {
		t.Fatalf("missing key created %d attempt rows", n)
	}

	// The header form is accepted and duplicate-key behaviour is unchanged.
	key := uniqKey()
	send := func(amount string) (int, string) {
		req := httptest.NewRequest("POST", "/pay", strings.NewReader(`{"amount":"`+amount+`","client_phone":"48110011","passcode":"`+bpayPasscode+`"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", key)
		resp, err := app.Test(req, 10000)
		if err != nil {
			t.Fatal(err)
		}
		raw, _ := io.ReadAll(resp.Body)
		return resp.StatusCode, string(raw)
	}
	if code, body := send("100"); code != 200 {
		t.Fatalf("first submit: %d %s", code, body)
	}
	if code, body := send("100"); code != 200 {
		t.Fatalf("same key + same payload: %d %s", code, body)
	}
	if code, body := send("900"); code != 409 || !strings.Contains(body, "bankily_idempotency_key_reused") {
		t.Fatalf("same key + different amount must be 409, got %d %s", code, body)
	}
	if n := count(); n != 1 || b.fake.PaymentCalls.Load() != 1 {
		t.Fatalf("expected 1 attempt / 1 Bankily call, got %d / %d", n, b.fake.PaymentCalls.Load())
	}
}

// HTTP: admin approve/reject/add-balance on a Bankily ledger row return the
// typed 409 automatic_transaction, and change neither the row nor the wallet.
func TestBankily_AdminHTTP_ManualActions409(t *testing.T) {
	b := newBpayEnv(t, 5*time.Second)
	p, err := b.pay(t, 100, uniqKey())
	if err != nil || p.WalletTransactionID == nil {
		t.Fatalf("setup payment: %v", err)
	}
	admin := createTestAdmin(t, b.env, "TEST BPAY HTTP ADMIN")
	h := handlers.NewAdminHandler(newTestAdminService(t, b.env), nil, b.env.logger)
	app := fiber.New()
	aid := admin.ID
	app.Put("/admin/transactions/:id/validate", fakeAuth(&aid), h.ValidateTransaction)
	app.Post("/admin/transactions/:id/add-balance", fakeAuth(&aid), h.AdminAddBalance)

	type rowState struct {
		Status     string     `db:"status"`
		ReviewedBy *uuid.UUID `db:"reviewed_by"`
		AdminNotes *string    `db:"admin_notes"`
	}
	var before rowState
	b.env.db.Get(&before, `SELECT status, reviewed_by, admin_notes FROM transactions WHERE id = $1`, *p.WalletTransactionID)
	balBefore := b.balance(t)

	cases := []struct{ name, method, path, body string }{
		{"approve", "PUT", "/validate", `{"approve":true,"notes":""}`},
		{"reject", "PUT", "/validate", `{"approve":false,"notes":"stale reject"}`},
		{"add-balance", "POST", "/add-balance", `{"amount":50,"notes":"x"}`},
	}
	for _, c := range cases {
		req := httptest.NewRequest(c.method, "/admin/transactions/"+p.WalletTransactionID.String()+c.path, strings.NewReader(c.body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, 10000)
		if err != nil {
			t.Fatal(err)
		}
		raw, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != 409 || !strings.Contains(string(raw), `"automatic_transaction"`) {
			t.Fatalf("%s: expected 409 automatic_transaction, got %d %s", c.name, resp.StatusCode, raw)
		}
	}

	var after rowState
	b.env.db.Get(&after, `SELECT status, reviewed_by, admin_notes FROM transactions WHERE id = $1`, *p.WalletTransactionID)
	if after.Status != "completed" || after.Status != before.Status || after.ReviewedBy != nil || after.AdminNotes != nil {
		t.Fatalf("Bankily ledger row changed: before=%+v after=%+v", before, after)
	}
	if !b.balance(t).Equal(balBefore) || b.ledgerRows(t, p.OperationID) != 1 {
		t.Fatal("wallet or ledger changed by refused admin actions")
	}
}
