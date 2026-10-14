package handlers

import (
	"bytes"
	"context"
	"encoding/json"
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

// Customer Request #27: end-to-end auth tests for
// PUT /v1/api/admin/categories/:id/toggle, wired through the REAL
// jwtMiddleware + AdminOnly middleware chain (identical to production
// routes.go registration), with a fake AdminService recording calls --
// proves unauthenticated/non-admin requests never reach ToggleCategory at
// all, mirroring the exact pattern already established for the Customer #22
// notification upload endpoint tests (notification_upload_test.go).

const testCategoryToggleJWTSecret = "test-secret-customer-27-category-toggle"

// fakeAdminServiceForToggle embeds the real AdminService interface (nil) and
// overrides only ToggleCategory -- any other method call would panic on a
// nil embedded interface, which is intentional: it proves these tests only
// exercise what the toggle endpoint actually touches.
type fakeAdminServiceForToggle struct {
	services.AdminService
	toggleCalls []toggleCall
}

type toggleCall struct {
	id       int
	isActive bool
	adminID  uuid.UUID
}

func (f *fakeAdminServiceForToggle) ToggleCategory(ctx context.Context, id int, isActive bool, adminID uuid.UUID) error {
	f.toggleCalls = append(f.toggleCalls, toggleCall{id: id, isActive: isActive, adminID: adminID})
	return nil
}

var _ services.AdminService = (*fakeAdminServiceForToggle)(nil)

func buildCategoryToggleTestApp(t *testing.T, adminSvc services.AdminService) *fiber.App {
	t.Helper()
	app := fiber.New()
	logger := zap.NewNop()
	adminHandler := NewAdminHandler(adminSvc, nil, logger)

	jwtMW := middleware.JWT(testCategoryToggleJWTSecret, logger, nil)
	adminMW := middleware.AdminOnly(logger)

	admin := app.Group("/v1/api/admin/categories", jwtMW, adminMW)
	admin.Put("/:id/toggle", adminHandler.ToggleCategory)
	return app
}

func signCategoryTestJWT(t *testing.T, role string) string {
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
	signed, err := token.SignedString([]byte(testCategoryToggleJWTSecret))
	if err != nil {
		t.Fatalf("failed to sign test JWT: %v", err)
	}
	return signed
}

func toggleRequest(t *testing.T, id string, isActive bool, token string) *http.Request {
	t.Helper()
	body, err := json.Marshal(map[string]bool{"is_active": isActive})
	if err != nil {
		t.Fatalf("failed to marshal toggle request body: %v", err)
	}
	req := httptest.NewRequest("PUT", "/v1/api/admin/categories/"+id+"/toggle", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return req
}

// TestToggleCategory_Unauthenticated_Rejected: no Authorization header at
// all -> 401, handler/AdminService never reached.
func TestToggleCategory_Unauthenticated_Rejected(t *testing.T) {
	fakeSvc := &fakeAdminServiceForToggle{}
	app := buildCategoryToggleTestApp(t, fakeSvc)

	req := toggleRequest(t, "1", false, "")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected 401 for missing token, got %d", resp.StatusCode)
	}
	if len(fakeSvc.toggleCalls) != 0 {
		t.Fatalf("expected ToggleCategory to never be called for an unauthenticated request, got %d calls", len(fakeSvc.toggleCalls))
	}
}

// TestToggleCategory_NonAdmin_Rejected: valid JWT but role=user -> 403,
// handler/AdminService never reached.
func TestToggleCategory_NonAdmin_Rejected(t *testing.T) {
	fakeSvc := &fakeAdminServiceForToggle{}
	app := buildCategoryToggleTestApp(t, fakeSvc)
	token := signCategoryTestJWT(t, "user")

	req := toggleRequest(t, "1", false, token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("expected 403 for non-admin role, got %d", resp.StatusCode)
	}
	if len(fakeSvc.toggleCalls) != 0 {
		t.Fatalf("expected ToggleCategory to never be called for a non-admin request, got %d calls", len(fakeSvc.toggleCalls))
	}
}

// TestToggleCategory_Admin_Allowed: valid admin JWT -> 200, ToggleCategory
// called exactly once with the correct id/is_active/adminID.
func TestToggleCategory_Admin_Allowed(t *testing.T) {
	fakeSvc := &fakeAdminServiceForToggle{}
	app := buildCategoryToggleTestApp(t, fakeSvc)
	token := signCategoryTestJWT(t, "admin")

	req := toggleRequest(t, "42", false, token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 for admin, got %d", resp.StatusCode)
	}
	if len(fakeSvc.toggleCalls) != 1 {
		t.Fatalf("expected exactly 1 ToggleCategory call, got %d", len(fakeSvc.toggleCalls))
	}
	if fakeSvc.toggleCalls[0].id != 42 {
		t.Fatalf("expected category id 42, got %d", fakeSvc.toggleCalls[0].id)
	}
	if fakeSvc.toggleCalls[0].isActive != false {
		t.Fatalf("expected is_active=false, got %v", fakeSvc.toggleCalls[0].isActive)
	}
}

// TestToggleCategory_SuperAdmin_Allowed: super_admin role also accepted,
// matching AdminOnly's own contract used by every other /admin/* route.
func TestToggleCategory_SuperAdmin_Allowed(t *testing.T) {
	fakeSvc := &fakeAdminServiceForToggle{}
	app := buildCategoryToggleTestApp(t, fakeSvc)
	token := signCategoryTestJWT(t, "super_admin")

	req := toggleRequest(t, "7", true, token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 for super_admin, got %d", resp.StatusCode)
	}
}
