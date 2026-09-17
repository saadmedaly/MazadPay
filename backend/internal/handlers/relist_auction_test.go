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
	"go.uber.org/zap"
)

// Customer Request #37: end-to-end auth tests for
// POST /v1/api/admin/auctions/:id/relist, wired through the REAL
// jwtMiddleware + AdminOnly middleware chain, with a fake AdminService
// recording calls -- proves unauthenticated/non-admin requests never reach
// RelistAuction at all. Mirrors the exact pattern established for
// winner_insurance_refund_test.go / admin_add_balance_test.go.

const testRelistAuctionJWTSecret = "test-secret-customer-37-relist-auction"

type fakeAdminServiceForRelistAuction struct {
	services.AdminService
	calls []relistAuctionCall
	err   error
}

type relistAuctionCall struct {
	auctionID uuid.UUID
	adminID   uuid.UUID
}

func (f *fakeAdminServiceForRelistAuction) RelistAuction(ctx context.Context, auctionID uuid.UUID, adminID uuid.UUID) (*models.Auction, error) {
	f.calls = append(f.calls, relistAuctionCall{auctionID: auctionID, adminID: adminID})
	if f.err != nil {
		return nil, f.err
	}
	return &models.Auction{ID: auctionID, Status: "pending"}, nil
}

var _ services.AdminService = (*fakeAdminServiceForRelistAuction)(nil)

func buildRelistAuctionTestApp(t *testing.T, adminSvc services.AdminService) *fiber.App {
	t.Helper()
	app := fiber.New()
	logger := zap.NewNop()
	adminHandler := NewAdminHandler(adminSvc, nil, logger)

	jwtMW := middleware.JWT(testRelistAuctionJWTSecret, logger, nil)
	adminMW := middleware.AdminOnly(logger)

	admin := app.Group("/v1/api/admin/auctions", jwtMW, adminMW)
	admin.Post("/:id/relist", adminHandler.RelistAuction)
	return app
}

func signRelistAuctionTestJWT(t *testing.T, role string) string {
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
	signed, err := token.SignedString([]byte(testRelistAuctionJWTSecret))
	if err != nil {
		t.Fatalf("failed to sign test JWT: %v", err)
	}
	return signed
}

func relistAuctionRequest(t *testing.T, auctionID string, token string) *http.Request {
	t.Helper()
	req := httptest.NewRequest("POST", "/v1/api/admin/auctions/"+auctionID+"/relist", nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return req
}

func TestRelistAuction_Unauthenticated_Rejected(t *testing.T) {
	fakeSvc := &fakeAdminServiceForRelistAuction{}
	app := buildRelistAuctionTestApp(t, fakeSvc)

	req := relistAuctionRequest(t, uuid.New().String(), "")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected 401 for missing token, got %d", resp.StatusCode)
	}
	if len(fakeSvc.calls) != 0 {
		t.Fatalf("expected RelistAuction to never be called for an unauthenticated request, got %d calls", len(fakeSvc.calls))
	}
}

func TestRelistAuction_NonAdmin_Rejected(t *testing.T) {
	fakeSvc := &fakeAdminServiceForRelistAuction{}
	app := buildRelistAuctionTestApp(t, fakeSvc)
	token := signRelistAuctionTestJWT(t, "user")

	req := relistAuctionRequest(t, uuid.New().String(), token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("expected 403 for non-admin role, got %d", resp.StatusCode)
	}
	if len(fakeSvc.calls) != 0 {
		t.Fatalf("expected RelistAuction to never be called for a non-admin request, got %d calls", len(fakeSvc.calls))
	}
}

func TestRelistAuction_Admin_Allowed(t *testing.T) {
	fakeSvc := &fakeAdminServiceForRelistAuction{}
	app := buildRelistAuctionTestApp(t, fakeSvc)
	token := signRelistAuctionTestJWT(t, "admin")

	auctionID := uuid.New()
	req := relistAuctionRequest(t, auctionID.String(), token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 for admin, got %d", resp.StatusCode)
	}
	if len(fakeSvc.calls) != 1 {
		t.Fatalf("expected exactly 1 RelistAuction call, got %d", len(fakeSvc.calls))
	}
	if fakeSvc.calls[0].auctionID != auctionID {
		t.Fatalf("expected auction id %s, got %s", auctionID, fakeSvc.calls[0].auctionID)
	}
}

func TestRelistAuction_SuperAdmin_Allowed(t *testing.T) {
	fakeSvc := &fakeAdminServiceForRelistAuction{}
	app := buildRelistAuctionTestApp(t, fakeSvc)
	token := signRelistAuctionTestJWT(t, "super_admin")

	req := relistAuctionRequest(t, uuid.New().String(), token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 for super_admin, got %d", resp.StatusCode)
	}
}
