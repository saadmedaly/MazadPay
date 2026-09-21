package handlers

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/middleware"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/services"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// Customer Request #35: end-to-end auth tests for
// POST /v1/api/admin/transactions/:id/add-balance, wired through the REAL
// jwtMiddleware + AdminOnly middleware chain (identical to production
// routes.go registration), with a fake AdminService recording calls --
// proves unauthenticated/non-admin requests never reach AdminAddBalance at
// all. Mirrors the exact pattern established for
// winner_insurance_refund_test.go.

const testAddBalanceJWTSecret = "test-secret-customer-35-add-balance"

type fakeAdminServiceForAddBalance struct {
	services.AdminService
	calls []addBalanceCall
	err   error
}

type addBalanceCall struct {
	transactionID uuid.UUID
	amount        decimal.Decimal
	notes         string
	adminID       uuid.UUID
}

func (f *fakeAdminServiceForAddBalance) AdminAddBalance(ctx context.Context, transactionID uuid.UUID, amount decimal.Decimal, notes string, adminID uuid.UUID) (*models.Transaction, error) {
	f.calls = append(f.calls, addBalanceCall{transactionID: transactionID, amount: amount, notes: notes, adminID: adminID})
	if f.err != nil {
		return nil, f.err
	}
	return &models.Transaction{ID: uuid.New(), Amount: amount}, nil
}

var _ services.AdminService = (*fakeAdminServiceForAddBalance)(nil)

func buildAddBalanceTestApp(t *testing.T, adminSvc services.AdminService) *fiber.App {
	t.Helper()
	app := fiber.New()
	logger := zap.NewNop()
	adminHandler := NewAdminHandler(adminSvc, nil, logger)

	jwtMW := middleware.JWT(testAddBalanceJWTSecret, logger, nil)
	adminMW := middleware.AdminOnly(logger)

	admin := app.Group("/v1/api/admin/transactions", jwtMW, adminMW)
	admin.Post("/:id/add-balance", adminHandler.AdminAddBalance)
	return app
}

func signAddBalanceTestJWT(t *testing.T, role string) string {
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
	signed, err := token.SignedString([]byte(testAddBalanceJWTSecret))
	if err != nil {
		t.Fatalf("failed to sign test JWT: %v", err)
	}
	return signed
}

func addBalanceRequest(t *testing.T, transactionID string, token string, body string) *http.Request {
	t.Helper()
	req := httptest.NewRequest("POST", "/v1/api/admin/transactions/"+transactionID+"/add-balance", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return req
}

func TestAdminAddBalance_Unauthenticated_Rejected(t *testing.T) {
	fakeSvc := &fakeAdminServiceForAddBalance{}
	app := buildAddBalanceTestApp(t, fakeSvc)

	req := addBalanceRequest(t, uuid.New().String(), "", `{"amount": 100}`)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected 401 for missing token, got %d", resp.StatusCode)
	}
	if len(fakeSvc.calls) != 0 {
		t.Fatalf("expected AdminAddBalance to never be called for an unauthenticated request, got %d calls", len(fakeSvc.calls))
	}
}

func TestAdminAddBalance_NonAdmin_Rejected(t *testing.T) {
	fakeSvc := &fakeAdminServiceForAddBalance{}
	app := buildAddBalanceTestApp(t, fakeSvc)
	token := signAddBalanceTestJWT(t, "user")

	req := addBalanceRequest(t, uuid.New().String(), token, `{"amount": 100}`)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("expected 403 for non-admin role, got %d", resp.StatusCode)
	}
	if len(fakeSvc.calls) != 0 {
		t.Fatalf("expected AdminAddBalance to never be called for a non-admin request, got %d calls", len(fakeSvc.calls))
	}
}

func TestAdminAddBalance_Admin_Allowed(t *testing.T) {
	fakeSvc := &fakeAdminServiceForAddBalance{}
	app := buildAddBalanceTestApp(t, fakeSvc)
	token := signAddBalanceTestJWT(t, "admin")

	transactionID := uuid.New()
	req := addBalanceRequest(t, transactionID.String(), token, `{"amount": 250}`)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 for admin, got %d", resp.StatusCode)
	}
	if len(fakeSvc.calls) != 1 {
		t.Fatalf("expected exactly 1 AdminAddBalance call, got %d", len(fakeSvc.calls))
	}
	if fakeSvc.calls[0].transactionID != transactionID {
		t.Fatalf("expected transaction id %s, got %s", transactionID, fakeSvc.calls[0].transactionID)
	}
	if !fakeSvc.calls[0].amount.Equal(decimal.NewFromInt(250)) {
		t.Fatalf("expected amount 250, got %s", fakeSvc.calls[0].amount)
	}
}

func TestAdminAddBalance_SuperAdmin_Allowed(t *testing.T) {
	fakeSvc := &fakeAdminServiceForAddBalance{}
	app := buildAddBalanceTestApp(t, fakeSvc)
	token := signAddBalanceTestJWT(t, "super_admin")

	req := addBalanceRequest(t, uuid.New().String(), token, `{"amount": 100}`)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 for super_admin, got %d", resp.StatusCode)
	}
}

// MAZADPAY -- live Validation bug: the admin UI previously sent amount as a
// JSON STRING (e.g. {"amount":"200"}), the raw <input> value never converted
// to a number. Request.Amount is `float64`, and Go's encoding/json does NOT
// coerce a JSON string into a float64 field -- BodyParser failed the
// unmarshal outright, surfaced to the admin as a generic "Invalid request
// body" 400 regardless of what value was entered. Fixed on the frontend
// (web/src/api/transactions.ts: Number(payload.amount)) so the wire format
// always carries a genuine JSON number. This test proves the backend
// contract those exact requested amounts must satisfy: a real JSON number
// is accepted and reaches AdminAddBalance with the exact value, for every
// amount explicitly requested (100, 200, 500, 1200, 5000).
func TestAdminAddBalance_ExactAmounts_AcceptedAsJSONNumbers(t *testing.T) {
	amounts := []int64{100, 200, 500, 1200, 5000}
	for _, amt := range amounts {
		fakeSvc := &fakeAdminServiceForAddBalance{}
		app := buildAddBalanceTestApp(t, fakeSvc)
		token := signAddBalanceTestJWT(t, "admin")

		transactionID := uuid.New()
		body := fmt.Sprintf(`{"amount": %d}`, amt)
		req := addBalanceRequest(t, transactionID.String(), token, body)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test failed for amount %d: %v", amt, err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("expected 200 for amount %d, got %d", amt, resp.StatusCode)
		}
		if len(fakeSvc.calls) != 1 {
			t.Fatalf("expected exactly 1 AdminAddBalance call for amount %d, got %d", amt, len(fakeSvc.calls))
		}
		if !fakeSvc.calls[0].amount.Equal(decimal.NewFromInt(amt)) {
			t.Fatalf("expected amount %d to reach AdminAddBalance exactly, got %s", amt, fakeSvc.calls[0].amount)
		}
	}
}

// MAZADPAY -- live Validation bug regression guard: a JSON STRING amount
// (e.g. {"amount":"200"}, exactly what the admin UI sent before the fix)
// must be rejected at the body-parse stage, proving the exact live failure
// mode is caught by this test suite rather than silently reappearing.
func TestAdminAddBalance_StringAmount_Rejected(t *testing.T) {
	fakeSvc := &fakeAdminServiceForAddBalance{}
	app := buildAddBalanceTestApp(t, fakeSvc)
	token := signAddBalanceTestJWT(t, "admin")

	req := addBalanceRequest(t, uuid.New().String(), token, `{"amount": "200"}`)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400 for a JSON-string amount (the exact live bug), got %d", resp.StatusCode)
	}
	if len(fakeSvc.calls) != 0 {
		t.Fatalf("expected AdminAddBalance to never be called for a malformed string amount, got %d calls", len(fakeSvc.calls))
	}
}

// Zero/negative amounts must be rejected by the handler itself (defense in
// depth alongside the service-level guard) -- never reach AdminAddBalance.
func TestAdminAddBalance_InvalidAmount_Rejected(t *testing.T) {
	fakeSvc := &fakeAdminServiceForAddBalance{}
	app := buildAddBalanceTestApp(t, fakeSvc)
	token := signAddBalanceTestJWT(t, "admin")

	for _, body := range []string{`{"amount": 0}`, `{"amount": -50}`} {
		req := addBalanceRequest(t, uuid.New().String(), token, body)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusBadRequest {
			t.Fatalf("expected 400 for body %s, got %d", body, resp.StatusCode)
		}
	}
	if len(fakeSvc.calls) != 0 {
		t.Fatalf("expected AdminAddBalance to never be called for an invalid amount, got %d calls", len(fakeSvc.calls))
	}
}
