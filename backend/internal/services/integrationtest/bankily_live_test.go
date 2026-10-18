//go:build integration && bankilylive

// Live run of the Bankily TEST scenarios (test.xlsx) through the real service
// against the Bankily TEST environment. Never part of the normal suites:
//
//	BPAY_ENV_FILE=../../../.env BPAY_ROWS_FILE=/path/rows.json BPAY_RESULTS_FILE=/path/out.json \
//	go test -tags 'integration bankilylive' -run TestBankilyLive ./internal/services/integrationtest/ -v
//
// Credentials come only from BPAY_ENV_FILE (gitignored backend/.env). The
// rows file holds the sheet's TEST customer data and stays outside the repo.
// Passcodes are never logged; phones are masked.
package integrationtest

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/mazadpay/backend/internal/config"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/payments/bankily"
	"github.com/mazadpay/backend/internal/repository"
	"github.com/mazadpay/backend/internal/services"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

type liveRow struct {
	Row              int    `json:"row"`
	Phone            string `json:"phone"`
	Amount           string `json:"amount"`
	Passcode         string `json:"passcode"`
	SheetOperationID string `json:"sheet_operation_id"`
}

type liveResult struct {
	Row              int     `json:"row"`
	Phone            string  `json:"phone_masked"`
	Amount           string  `json:"amount"`
	SheetOperationID string  `json:"sheet_operation_id"`
	UsedOperationID  string  `json:"used_operation_id"`
	ErrorCode        *string `json:"error_code"`
	ErrorMessage     *string `json:"error_message"`
	TransactionID    *string `json:"transaction_id"`
	CheckStatus      *string `json:"check_status"`
	FinalStatus      string  `json:"final_status"`
	FailureReason    *string `json:"failure_reason"`
	CheckAttempts    int     `json:"check_attempts"`
	WalletCredited   string  `json:"wallet_credited"`
	LedgerRows       int     `json:"ledger_rows"`
	ServiceError     string  `json:"service_error,omitempty"`
}

func TestBankilyLive(t *testing.T) {
	if err := godotenv.Load(os.Getenv("BPAY_ENV_FILE")); err != nil {
		t.Skip("BPAY_ENV_FILE not set/readable")
	}
	cfg := config.Load().Bankily
	if !cfg.Ready() || !strings.Contains(cfg.BaseURL, "-tst.") {
		t.Fatal("refusing to run: Bankily config not ready or not the TEST environment")
	}
	cfg.MinAmount = 1 // the sheet includes amounts below the production minimum
	cfg.OperationIDPrefix = "MZPT"

	raw, err := os.ReadFile(os.Getenv("BPAY_ROWS_FILE"))
	if err != nil {
		t.Fatal(err)
	}
	var rows []liveRow
	if err := json.Unmarshal(raw, &rows); err != nil {
		t.Fatal(err)
	}

	env := setupEnv(t)
	logger := zaptest.NewLogger(t, zaptest.Level(zap.InfoLevel))
	client := bankily.NewClient(bankily.Config{BaseURL: cfg.BaseURL, Username: cfg.Username, Password: cfg.Password, ClientID: cfg.ClientID, Timeout: cfg.HTTPTimeout}, nil)
	auditSvc := services.NewAuditService(repository.NewAuditRepository(env.db))
	svc := services.NewBankilyPaymentService(env.db, cfg, client, repository.NewBankilyPaymentRepository(env.db), env.walletRepo, env.userRepo, nil, auditSvc, logger)
	user := createTestUser(t, env, "TEST BPAY LIVE")
	ctx := context.Background()
	if _, err := env.walletRepo.GetByUserID(ctx, user.ID); err != nil {
		t.Fatal(err)
	}

	// Fresh authentication + checkTransaction format for an unknown operation.
	probe, err := client.CheckTransaction(ctx, "MZPT-PROBE-UNKNOWN-"+time.Now().Format("150405"))
	if err != nil {
		t.Fatalf("fresh authentication / checkTransaction probe failed: %v", err)
	}
	t.Logf("PROBE unknown operation -> errorCode=%q status=%q message=%q", probe.ErrorCode, probe.Status, probe.ErrorMessage)

	results := make([]liveResult, 0, len(rows))
	ids := make([]*models.BankilyPayment, len(rows))
	for i, r := range rows {
		before, _ := walletBalanceAndFrozen(t, env, user.ID)
		p, err := svc.Pay(ctx, user.ID, services.BankilyPayInput{
			Amount: decimal.RequireFromString(r.Amount), ClientPhone: r.Phone, Passcode: r.Passcode, Language: "fr",
			IdempotencyKey: "live-" + uuid.NewString(),
		})
		res := liveResult{Row: r.Row, Phone: models.MaskPhone(r.Phone), Amount: r.Amount, SheetOperationID: r.SheetOperationID}
		if err != nil {
			res.ServiceError = err.Error()
			results = append(results, res)
			continue
		}
		ids[i] = p
		after, _ := walletBalanceAndFrozen(t, env, user.ID)
		res.WalletCredited = after.Sub(before).String()
		results = append(results, res)
	}

	// Give TA/uncertain attempts a chance to settle, then re-check once.
	time.Sleep(20 * time.Second)
	for i, p := range ids {
		if p == nil {
			continue
		}
		before, _ := walletBalanceAndFrozen(t, env, user.ID)
		if q, err := svc.Recheck(ctx, p.ID, "system", user.ID); err == nil && !p.IsTerminal() {
			after, _ := walletBalanceAndFrozen(t, env, user.ID)
			if d := after.Sub(before); !d.IsZero() {
				prev, _ := decimal.NewFromString(results[i].WalletCredited)
				results[i].WalletCredited = prev.Add(d).String()
			}
			_ = q
		}
		final, _ := svc.GetForUser(ctx, user.ID, p.ID)
		results[i].UsedOperationID = final.OperationID
		results[i].ErrorCode = final.BankilyErrorCode
		results[i].ErrorMessage = final.BankilyErrorMessage
		results[i].TransactionID = final.BankilyTransactionID
		results[i].CheckStatus = final.BankilyStatus
		results[i].FinalStatus = final.Status
		results[i].FailureReason = final.FailureReason
		results[i].CheckAttempts = final.CheckAttempts
		env.db.Get(&results[i].LedgerRows, `SELECT COUNT(*) FROM transactions WHERE gateway='bankily_bpay' AND reference=$1`, "bankily:"+final.OperationID)
	}

	// Duplicate operationId against real Bankily: replay row 1's operation.
	if ids[0] != nil {
		dup, err := client.Payment(ctx, bankily.PaymentRequest{
			ClientPhone: rows[0].Phone, Passcode: rows[0].Passcode, OperationID: ids[0].OperationID, Amount: rows[0].Amount, Language: "FR",
		})
		if err != nil {
			t.Logf("DUPLICATE operationId replay -> error: %v", err)
		} else {
			t.Logf("DUPLICATE operationId replay -> errorCode=%q message=%q transactionId=%q", dup.ErrorCode, dup.ErrorMessage, dup.TransactionID)
		}
		chk, err := client.CheckTransaction(ctx, ids[0].OperationID)
		if err == nil {
			t.Logf("DUPLICATE follow-up check -> errorCode=%q status=%q transactionId=%q", chk.ErrorCode, chk.Status, chk.TransactionID)
		}
	}

	for _, r := range results {
		t.Logf("ROW %d phone=%s amount=%s -> status=%s code=%s msg=%q txId=%s check=%s credited=%s ledger=%d err=%s",
			r.Row, r.Phone, r.Amount, r.FinalStatus, deref(r.ErrorCode), deref(r.ErrorMessage), deref(r.TransactionID), deref(r.CheckStatus), r.WalletCredited, r.LedgerRows, r.ServiceError)
	}
	if out := os.Getenv("BPAY_RESULTS_FILE"); out != "" {
		b, _ := json.MarshalIndent(results, "", "  ")
		_ = os.WriteFile(out, b, 0o600)
	}
}

func deref(s *string) string {
	if s == nil {
		return "-"
	}
	return *s
}

// TestBankilyLiveReconcile runs the reconciler once against Bankily TEST for
// every unresolved attempt left by TestBankilyLive and logs the outcome.
func TestBankilyLiveReconcile(t *testing.T) {
	if err := godotenv.Load(os.Getenv("BPAY_ENV_FILE")); err != nil {
		t.Skip("BPAY_ENV_FILE not set/readable")
	}
	cfg := config.Load().Bankily
	if !cfg.Ready() || !strings.Contains(cfg.BaseURL, "-tst.") {
		t.Fatal("refusing to run: Bankily config not ready or not the TEST environment")
	}
	env := setupEnv(t)
	client := bankily.NewClient(bankily.Config{BaseURL: cfg.BaseURL, Username: cfg.Username, Password: cfg.Password, ClientID: cfg.ClientID, Timeout: cfg.HTTPTimeout}, nil)
	svc := services.NewBankilyPaymentService(env.db, cfg, client, repository.NewBankilyPaymentRepository(env.db), env.walletRepo, env.userRepo, nil,
		services.NewAuditService(repository.NewAuditRepository(env.db)), zaptest.NewLogger(t, zaptest.Level(zap.InfoLevel)))
	ctx := context.Background()
	env.db.MustExecContext(ctx, `UPDATE bankily_payments SET next_check_at = NOW() - INTERVAL '1 second'
		WHERE operation_id LIKE 'MZPT%' AND status IN ('SUBMITTED','PENDING_VERIFICATION')`)
	if err := svc.ReconcileOnce(ctx); err != nil {
		t.Fatal(err)
	}
	var rows []models.BankilyPayment
	if err := env.db.SelectContext(ctx, &rows, `SELECT * FROM bankily_payments WHERE operation_id LIKE 'MZPT%' ORDER BY created_at`); err != nil {
		t.Fatal(err)
	}
	for _, p := range rows {
		var ledger int
		env.db.Get(&ledger, `SELECT COUNT(*) FROM transactions WHERE gateway='bankily_bpay' AND reference=$1`, "bankily:"+p.OperationID)
		t.Logf("RECONCILE %s amount=%s phone=%s -> status=%s check=%s attempts=%d ledger=%d",
			p.OperationID, p.Amount.String(), models.MaskPhone(p.ClientPhone), p.Status, deref(p.BankilyStatus), p.CheckAttempts, ledger)
	}
}
