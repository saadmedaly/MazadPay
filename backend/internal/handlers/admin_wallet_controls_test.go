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
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/services"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// MAZADPAY -- admin wallet controls: end-to-end auth tests for
// POST /v1/api/admin/transactions/:id/deduct-balance and
// PUT /v1/api/admin/users/:id/wallet-disabled, wired through the REAL
// jwtMiddleware + AdminOnly middleware chain, mirroring
// admin_add_balance_test.go's exact pattern.

const testWalletControlsJWTSecret = "test-secret-mazadpay-wallet-controls"

type fakeAdminServiceForWalletControls struct {
	services.AdminService
	deductCalls    []deductBalanceCall
	disabledCalls  []setDisabledCall
	deductErr      error
	setDisabledErr error
}

type deductBalanceCall struct {
	transactionID uuid.UUID
	amount        decimal.Decimal
	notes         string
	adminID       uuid.UUID
}

type setDisabledCall struct {
	userID   uuid.UUID
	disabled bool
	reason   string
	adminID  uuid.UUID
}

func (f *fakeAdminServiceForWalletControls) AdminDeductBalance(ctx context.Context, transactionID uuid.UUID, amount decimal.Decimal, notes string, adminID uuid.UUID) (*models.Transaction, error) {
	f.deductCalls = append(f.deductCalls, deductBalanceCall{transactionID: transactionID, amount: amount, notes: notes, adminID: adminID})
	if f.deductErr != nil {
		return nil, f.deductErr
	}
	return &models.Transaction{ID: uuid.New(), Amount: amount}, nil
}

func (f *fakeAdminServiceForWalletControls) AdminSetWalletDisabled(ctx context.Context, userID uuid.UUID, disabled bool, reason string, adminID uuid.UUID) error {
	f.disabledCalls = append(f.disabledCalls, setDisabledCall{userID: userID, disabled: disabled, reason: reason, adminID: adminID})
	return f.setDisabledErr
}

var _ services.AdminService = (*fakeAdminServiceForWalletControls)(nil)

func buildWalletControlsTestApp(t *testing.T, adminSvc services.AdminService) *fiber.App {
	t.Helper()
	app := fiber.New()
	logger := zap.NewNop()
	adminHandler := NewAdminHandler(adminSvc, nil, logger)

	jwtMW := middleware.JWT(testWalletControlsJWTSecret, logger, nil)
	adminMW := middleware.AdminOnly(logger)

	admin := app.Group("/v1/api/admin", jwtMW, adminMW)
	admin.Post("/transactions/:id/deduct-balance", adminHandler.AdminDeductBalance)
	admin.Put("/users/:id/wallet-disabled", adminHandler.AdminSetWalletDisabled)
	return app
}

func signWalletControlsTestJWT(t *testing.T, role string) string {
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
	signed, err := token.SignedString([]byte(testWalletControlsJWTSecret))
	if err != nil {
		t.Fatalf("failed to sign test JWT: %v", err)
	}
	return signed
}

func deductBalanceRequest(t *testing.T, transactionID string, token string, body string) *http.Request {
	t.Helper()
	req := httptest.NewRequest("POST", "/v1/api/admin/transactions/"+transactionID+"/deduct-balance", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return req
}

func setWalletDisabledRequest(t *testing.T, userID string, token string, body string) *http.Request {
	t.Helper()
	req := httptest.NewRequest("PUT", "/v1/api/admin/users/"+userID+"/wallet-disabled", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return req
}

func TestAdminDeductBalance_Unauthenticated_Rejected(t *testing.T) {
	fakeSvc := &fakeAdminServiceForWalletControls{}
	app := buildWalletControlsTestApp(t, fakeSvc)

	req := deductBalanceRequest(t, uuid.New().String(), "", `{"amount": 100}`)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected 401 for missing token, got %d", resp.StatusCode)
	}
	if len(fakeSvc.deductCalls) != 0 {
		t.Fatalf("expected AdminDeductBalance to never be called for an unauthenticated request, got %d calls", len(fakeSvc.deductCalls))
	}
}

func TestAdminDeductBalance_NonAdmin_Rejected(t *testing.T) {
	fakeSvc := &fakeAdminServiceForWalletControls{}
	app := buildWalletControlsTestApp(t, fakeSvc)
	token := signWalletControlsTestJWT(t, "user")

	req := deductBalanceRequest(t, uuid.New().String(), token, `{"amount": 100}`)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("expected 403 for non-admin role, got %d", resp.StatusCode)
	}
	if len(fakeSvc.deductCalls) != 0 {
		t.Fatalf("expected AdminDeductBalance to never be called for a non-admin request, got %d calls", len(fakeSvc.deductCalls))
	}
}

func TestAdminDeductBalance_Admin_Allowed(t *testing.T) {
	fakeSvc := &fakeAdminServiceForWalletControls{}
	app := buildWalletControlsTestApp(t, fakeSvc)
	token := signWalletControlsTestJWT(t, "admin")

	transactionID := uuid.New()
	req := deductBalanceRequest(t, transactionID.String(), token, `{"amount": 300, "notes": "مبلغ خاطئ"}`)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 for admin, got %d", resp.StatusCode)
	}
	if len(fakeSvc.deductCalls) != 1 {
		t.Fatalf("expected exactly 1 AdminDeductBalance call, got %d", len(fakeSvc.deductCalls))
	}
	if fakeSvc.deductCalls[0].transactionID != transactionID {
		t.Fatalf("expected transaction id %s, got %s", transactionID, fakeSvc.deductCalls[0].transactionID)
	}
	if !fakeSvc.deductCalls[0].amount.Equal(decimal.NewFromInt(300)) {
		t.Fatalf("expected amount 300, got %s", fakeSvc.deductCalls[0].amount)
	}
}

// Zero/negative amounts must be rejected by the handler itself (defense in
// depth alongside the service-level guard) -- never reach AdminDeductBalance.
func TestAdminDeductBalance_InvalidAmount_Rejected(t *testing.T) {
	fakeSvc := &fakeAdminServiceForWalletControls{}
	app := buildWalletControlsTestApp(t, fakeSvc)
	token := signWalletControlsTestJWT(t, "admin")

	for _, body := range []string{`{"amount": 0}`, `{"amount": -50}`} {
		req := deductBalanceRequest(t, uuid.New().String(), token, body)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusBadRequest {
			t.Fatalf("expected 400 for body %s, got %d", body, resp.StatusCode)
		}
	}
	if len(fakeSvc.deductCalls) != 0 {
		t.Fatalf("expected AdminDeductBalance to never be called for an invalid amount, got %d calls", len(fakeSvc.deductCalls))
	}
}

func TestAdminSetWalletDisabled_Unauthenticated_Rejected(t *testing.T) {
	fakeSvc := &fakeAdminServiceForWalletControls{}
	app := buildWalletControlsTestApp(t, fakeSvc)

	req := setWalletDisabledRequest(t, uuid.New().String(), "", `{"disabled": true, "reason": "test"}`)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected 401 for missing token, got %d", resp.StatusCode)
	}
	if len(fakeSvc.disabledCalls) != 0 {
		t.Fatalf("expected AdminSetWalletDisabled to never be called for an unauthenticated request, got %d calls", len(fakeSvc.disabledCalls))
	}
}

func TestAdminSetWalletDisabled_NonAdmin_Rejected(t *testing.T) {
	fakeSvc := &fakeAdminServiceForWalletControls{}
	app := buildWalletControlsTestApp(t, fakeSvc)
	token := signWalletControlsTestJWT(t, "user")

	req := setWalletDisabledRequest(t, uuid.New().String(), token, `{"disabled": true}`)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("expected 403 for non-admin role, got %d", resp.StatusCode)
	}
	if len(fakeSvc.disabledCalls) != 0 {
		t.Fatalf("expected AdminSetWalletDisabled to never be called for a non-admin request, got %d calls", len(fakeSvc.disabledCalls))
	}
}

func TestAdminSetWalletDisabled_Disable_Allowed(t *testing.T) {
	fakeSvc := &fakeAdminServiceForWalletControls{}
	app := buildWalletControlsTestApp(t, fakeSvc)
	token := signWalletControlsTestJWT(t, "admin")

	userID := uuid.New()
	req := setWalletDisabledRequest(t, userID.String(), token, `{"disabled": true, "reason": "نشاط مشبوه"}`)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 for admin disabling a wallet, got %d", resp.StatusCode)
	}
	if len(fakeSvc.disabledCalls) != 1 {
		t.Fatalf("expected exactly 1 AdminSetWalletDisabled call, got %d", len(fakeSvc.disabledCalls))
	}
	call := fakeSvc.disabledCalls[0]
	if call.userID != userID || !call.disabled || call.reason != "نشاط مشبوه" {
		t.Fatalf("unexpected call params: %+v", call)
	}
}

func TestAdminSetWalletDisabled_ReEnable_Allowed(t *testing.T) {
	fakeSvc := &fakeAdminServiceForWalletControls{}
	app := buildWalletControlsTestApp(t, fakeSvc)
	token := signWalletControlsTestJWT(t, "super_admin")

	userID := uuid.New()
	req := setWalletDisabledRequest(t, userID.String(), token, `{"disabled": false}`)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 for super_admin re-enabling a wallet, got %d", resp.StatusCode)
	}
	if len(fakeSvc.disabledCalls) != 1 || fakeSvc.disabledCalls[0].disabled {
		t.Fatalf("expected exactly 1 re-enable (disabled=false) call, got %+v", fakeSvc.disabledCalls)
	}
}
