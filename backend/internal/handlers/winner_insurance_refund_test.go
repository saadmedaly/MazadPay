package handlers

import (
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

// Customer Request #31: end-to-end auth tests for
// POST /v1/api/admin/auctions/:id/refund-winner-insurance, wired through
// the REAL jwtMiddleware + AdminOnly middleware chain (identical to
// production routes.go registration), with a fake AdminService recording
// calls -- proves unauthenticated/non-admin requests never reach
// RefundWinnerInsurance at all. Mirrors the exact pattern already
// established for category_toggle_test.go.

const testWinnerRefundJWTSecret = "test-secret-customer-31-winner-refund"

type fakeAdminServiceForWinnerRefund struct {
	services.AdminService
	calls []refundCall
}

type refundCall struct {
	auctionID uuid.UUID
	adminID   uuid.UUID
}

func (f *fakeAdminServiceForWinnerRefund) RefundWinnerInsurance(ctx context.Context, auctionID uuid.UUID, adminID uuid.UUID) (*models.Transaction, error) {
	f.calls = append(f.calls, refundCall{auctionID: auctionID, adminID: adminID})
	return &models.Transaction{ID: uuid.New(), Amount: decimal.NewFromInt(100)}, nil
}

var _ services.AdminService = (*fakeAdminServiceForWinnerRefund)(nil)

func buildWinnerRefundTestApp(t *testing.T, adminSvc services.AdminService) *fiber.App {
	t.Helper()
	app := fiber.New()
	logger := zap.NewNop()
	adminHandler := NewAdminHandler(adminSvc, nil, logger)

	jwtMW := middleware.JWT(testWinnerRefundJWTSecret, logger, nil)
	adminMW := middleware.AdminOnly(logger)

	admin := app.Group("/v1/api/admin/auctions", jwtMW, adminMW)
	admin.Post("/:id/refund-winner-insurance", adminHandler.RefundWinnerInsurance)
	return app
}

func signWinnerRefundTestJWT(t *testing.T, role string) string {
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
	signed, err := token.SignedString([]byte(testWinnerRefundJWTSecret))
	if err != nil {
		t.Fatalf("failed to sign test JWT: %v", err)
	}
	return signed
}

func winnerRefundRequest(t *testing.T, auctionID string, token string) *http.Request {
	t.Helper()
	req := httptest.NewRequest("POST", "/v1/api/admin/auctions/"+auctionID+"/refund-winner-insurance", nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return req
}

func TestRefundWinnerInsurance_Unauthenticated_Rejected(t *testing.T) {
	fakeSvc := &fakeAdminServiceForWinnerRefund{}
	app := buildWinnerRefundTestApp(t, fakeSvc)

	req := winnerRefundRequest(t, uuid.New().String(), "")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected 401 for missing token, got %d", resp.StatusCode)
	}
	if len(fakeSvc.calls) != 0 {
		t.Fatalf("expected RefundWinnerInsurance to never be called for an unauthenticated request, got %d calls", len(fakeSvc.calls))
	}
}

func TestRefundWinnerInsurance_NonAdmin_Rejected(t *testing.T) {
	fakeSvc := &fakeAdminServiceForWinnerRefund{}
	app := buildWinnerRefundTestApp(t, fakeSvc)
	token := signWinnerRefundTestJWT(t, "user")

	req := winnerRefundRequest(t, uuid.New().String(), token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("expected 403 for non-admin role, got %d", resp.StatusCode)
	}
	if len(fakeSvc.calls) != 0 {
		t.Fatalf("expected RefundWinnerInsurance to never be called for a non-admin request, got %d calls", len(fakeSvc.calls))
	}
}

func TestRefundWinnerInsurance_Admin_Allowed(t *testing.T) {
	fakeSvc := &fakeAdminServiceForWinnerRefund{}
	app := buildWinnerRefundTestApp(t, fakeSvc)
	token := signWinnerRefundTestJWT(t, "admin")

	auctionID := uuid.New()
	req := winnerRefundRequest(t, auctionID.String(), token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 for admin, got %d", resp.StatusCode)
	}
	if len(fakeSvc.calls) != 1 {
		t.Fatalf("expected exactly 1 RefundWinnerInsurance call, got %d", len(fakeSvc.calls))
	}
	if fakeSvc.calls[0].auctionID != auctionID {
		t.Fatalf("expected auction id %s, got %s", auctionID, fakeSvc.calls[0].auctionID)
	}
}

func TestRefundWinnerInsurance_SuperAdmin_Allowed(t *testing.T) {
	fakeSvc := &fakeAdminServiceForWinnerRefund{}
	app := buildWinnerRefundTestApp(t, fakeSvc)
	token := signWinnerRefundTestJWT(t, "super_admin")

	req := winnerRefundRequest(t, uuid.New().String(), token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 for super_admin, got %d", resp.StatusCode)
	}
}
