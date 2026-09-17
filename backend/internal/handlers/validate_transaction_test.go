package handlers

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/middleware"
	"github.com/mazadpay/backend/internal/services"
	"go.uber.org/zap"
)

// Customer Request #36: admin transaction review (approve/reject),
// extended to cover withdrawals with an optional attachment URL and a
// server-side guard that rejection always requires a note. Wired through
// the REAL jwtMiddleware + AdminOnly middleware chain, with a fake
// AdminService recording calls -- mirrors admin_add_balance_test.go's
// pattern.

const testValidateTxnJWTSecret = "test-secret-customer-36-validate-txn"

type fakeAdminServiceForValidateTxn struct {
	services.AdminService
	calls []validateTxnCall
}

type validateTxnCall struct {
	transactionID uuid.UUID
	approve       bool
	notes         string
	adminID       uuid.UUID
	attachmentURL string
}

func (f *fakeAdminServiceForValidateTxn) ValidateTransaction(ctx context.Context, transactionID uuid.UUID, approve bool, notes string, adminID uuid.UUID, attachmentURL string) error {
	f.calls = append(f.calls, validateTxnCall{transactionID: transactionID, approve: approve, notes: notes, adminID: adminID, attachmentURL: attachmentURL})
	return nil
}

var _ services.AdminService = (*fakeAdminServiceForValidateTxn)(nil)

func buildValidateTxnTestApp(t *testing.T, adminSvc services.AdminService) *fiber.App {
	t.Helper()
	app := fiber.New()
	logger := zap.NewNop()
	adminHandler := NewAdminHandler(adminSvc, nil, logger)

	jwtMW := middleware.JWT(testValidateTxnJWTSecret, logger, nil)
	adminMW := middleware.AdminOnly(logger)

	admin := app.Group("/v1/api/admin/transactions", jwtMW, adminMW)
	admin.Put("/:id/validate", adminHandler.ValidateTransaction)
	return app
}

func signValidateTxnTestJWT(t *testing.T, role string) string {
	t.Helper()
	claims := services.JWTClaims{
		UserID: uuid.New().String(),
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(testValidateTxnJWTSecret))
	if err != nil {
		t.Fatalf("failed to sign test JWT: %v", err)
	}
	return signed
}

func validateTxnRequest(t *testing.T, transactionID string, token string, body string) *http.Request {
	t.Helper()
	req := httptest.NewRequest("PUT", "/v1/api/admin/transactions/"+transactionID+"/validate", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return req
}

func TestValidateTransaction_Unauthenticated_Rejected(t *testing.T) {
	fakeSvc := &fakeAdminServiceForValidateTxn{}
	app := buildValidateTxnTestApp(t, fakeSvc)

	req := validateTxnRequest(t, uuid.New().String(), "", `{"approve": true, "notes": ""}`)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected 401 for missing token, got %d", resp.StatusCode)
	}
	if len(fakeSvc.calls) != 0 {
		t.Fatalf("expected ValidateTransaction to never be called for an unauthenticated request, got %d calls", len(fakeSvc.calls))
	}
}

func TestValidateTransaction_NonAdmin_Rejected(t *testing.T) {
	fakeSvc := &fakeAdminServiceForValidateTxn{}
	app := buildValidateTxnTestApp(t, fakeSvc)
	token := signValidateTxnTestJWT(t, "user")

	req := validateTxnRequest(t, uuid.New().String(), token, `{"approve": true, "notes": ""}`)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("expected 403 for non-admin role, got %d", resp.StatusCode)
	}
	if len(fakeSvc.calls) != 0 {
		t.Fatalf("expected ValidateTransaction to never be called for a non-admin request, got %d calls", len(fakeSvc.calls))
	}
}

func TestValidateTransaction_ApproveWithNoNote_Allowed(t *testing.T) {
	fakeSvc := &fakeAdminServiceForValidateTxn{}
	app := buildValidateTxnTestApp(t, fakeSvc)
	token := signValidateTxnTestJWT(t, "admin")

	transactionID := uuid.New()
	req := validateTxnRequest(t, transactionID.String(), token, `{"approve": true, "notes": ""}`)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 for approve with no note, got %d", resp.StatusCode)
	}
	if len(fakeSvc.calls) != 1 {
		t.Fatalf("expected exactly 1 ValidateTransaction call, got %d", len(fakeSvc.calls))
	}
	if !fakeSvc.calls[0].approve {
		t.Fatalf("expected approve=true")
	}
}

func TestValidateTransaction_ApproveWithNote_Allowed(t *testing.T) {
	fakeSvc := &fakeAdminServiceForValidateTxn{}
	app := buildValidateTxnTestApp(t, fakeSvc)
	token := signValidateTxnTestJWT(t, "admin")

	req := validateTxnRequest(t, uuid.New().String(), token, `{"approve": true, "notes": "verified"}`)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 for approve with note, got %d", resp.StatusCode)
	}
	if fakeSvc.calls[0].notes != "verified" {
		t.Fatalf("expected notes 'verified', got %q", fakeSvc.calls[0].notes)
	}
}

// Customer #36: rejection without a note must be blocked with 400, never
// reaching AdminService.ValidateTransaction at all.
func TestValidateTransaction_RejectWithoutNote_Rejected(t *testing.T) {
	fakeSvc := &fakeAdminServiceForValidateTxn{}
	app := buildValidateTxnTestApp(t, fakeSvc)
	token := signValidateTxnTestJWT(t, "admin")

	for _, body := range []string{`{"approve": false, "notes": ""}`, `{"approve": false, "notes": "   "}`, `{"approve": false}`} {
		req := validateTxnRequest(t, uuid.New().String(), token, body)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusBadRequest {
			t.Fatalf("expected 400 for body %s, got %d", body, resp.StatusCode)
		}
	}
	if len(fakeSvc.calls) != 0 {
		t.Fatalf("expected ValidateTransaction to never be called for a note-less rejection, got %d calls", len(fakeSvc.calls))
	}
}

func TestValidateTransaction_RejectWithNote_Allowed(t *testing.T) {
	fakeSvc := &fakeAdminServiceForValidateTxn{}
	app := buildValidateTxnTestApp(t, fakeSvc)
	token := signValidateTxnTestJWT(t, "admin")

	req := validateTxnRequest(t, uuid.New().String(), token, `{"approve": false, "notes": "invalid beneficiary"}`)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 for reject with note, got %d", resp.StatusCode)
	}
	if len(fakeSvc.calls) != 1 || fakeSvc.calls[0].approve {
		t.Fatalf("expected exactly 1 reject call reaching the service")
	}
}

// The optional attachment_url must pass through to the service call intact.
func TestValidateTransaction_AttachmentURLPassedThrough(t *testing.T) {
	fakeSvc := &fakeAdminServiceForValidateTxn{}
	app := buildValidateTxnTestApp(t, fakeSvc)
	token := signValidateTxnTestJWT(t, "admin")

	req := validateTxnRequest(t, uuid.New().String(), token, `{"approve": true, "notes": "", "attachment_url": "https://r2.example.com/transaction-review/x.png"}`)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if fakeSvc.calls[0].attachmentURL != "https://r2.example.com/transaction-review/x.png" {
		t.Fatalf("expected attachment_url to pass through, got %q", fakeSvc.calls[0].attachmentURL)
	}
}
