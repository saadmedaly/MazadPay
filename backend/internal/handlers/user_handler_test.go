package handlers

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/services"
	"go.uber.org/zap"
)

// fakeUserService is a minimal services.UserService fake used only to prove
// UpdateUserSettings's handler-level theme validation (client feedback:
// Bug D.1) -- it never touches a real database. Embedding the real
// interface means any method this test doesn't need panics if called,
// which is exactly what proves an invalid theme never reaches the service
// layer at all.
type fakeUserService struct {
	services.UserService
	updateSettingsCalled bool
	updateSettingsErr    error
}

func (f *fakeUserService) UpdateUserSettings(ctx context.Context, userID uuid.UUID, settings models.UserSettingsUpdate) error {
	f.updateSettingsCalled = true
	return f.updateSettingsErr
}

// (D1-1) theme=light is accepted and reaches the service layer.
func TestUpdateUserSettings_D1_ThemeLight_Accepted(t *testing.T) {
	fakeSvc := &fakeUserService{}
	h := NewUserHandler(fakeSvc, zap.NewNop())

	app := fiber.New()
	userID := uuid.New()
	app.Put("/v1/api/users/me/settings", func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		return h.UpdateUserSettings(c)
	})

	req := httptest.NewRequest(http.MethodPut, "/v1/api/users/me/settings", bytes.NewReader([]byte(`{"theme":"light"}`)))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("(D1-1) expected HTTP 200 for theme=light, got HTTP %d", resp.StatusCode)
	}
	if !fakeSvc.updateSettingsCalled {
		t.Fatal("(D1-1) expected a valid theme to reach the service layer")
	}
}

// (D1-2) theme=dark is accepted.
func TestUpdateUserSettings_D1_ThemeDark_Accepted(t *testing.T) {
	fakeSvc := &fakeUserService{}
	h := NewUserHandler(fakeSvc, zap.NewNop())

	app := fiber.New()
	userID := uuid.New()
	app.Put("/v1/api/users/me/settings", func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		return h.UpdateUserSettings(c)
	})

	req := httptest.NewRequest(http.MethodPut, "/v1/api/users/me/settings", bytes.NewReader([]byte(`{"theme":"dark"}`)))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("(D1-2) expected HTTP 200 for theme=dark, got HTTP %d", resp.StatusCode)
	}
	if !fakeSvc.updateSettingsCalled {
		t.Fatal("(D1-2) expected a valid theme to reach the service layer")
	}
}

// (D1-3) theme=auto is accepted.
func TestUpdateUserSettings_D1_ThemeAuto_Accepted(t *testing.T) {
	fakeSvc := &fakeUserService{}
	h := NewUserHandler(fakeSvc, zap.NewNop())

	app := fiber.New()
	userID := uuid.New()
	app.Put("/v1/api/users/me/settings", func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		return h.UpdateUserSettings(c)
	})

	req := httptest.NewRequest(http.MethodPut, "/v1/api/users/me/settings", bytes.NewReader([]byte(`{"theme":"auto"}`)))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("(D1-3) expected HTTP 200 for theme=auto, got HTTP %d", resp.StatusCode)
	}
	if !fakeSvc.updateSettingsCalled {
		t.Fatal("(D1-3) expected a valid theme to reach the service layer")
	}
}

// (D1-4) theme=invalid is rejected with HTTP 400, and never reaches the
// service layer (so it can never reach the DB / hit chk_theme / 500).
func TestUpdateUserSettings_D1_ThemeInvalid_RejectedAsBadRequest(t *testing.T) {
	fakeSvc := &fakeUserService{}
	h := NewUserHandler(fakeSvc, zap.NewNop())

	app := fiber.New()
	userID := uuid.New()
	app.Put("/v1/api/users/me/settings", func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		return h.UpdateUserSettings(c)
	})

	req := httptest.NewRequest(http.MethodPut, "/v1/api/users/me/settings", bytes.NewReader([]byte(`{"theme":"invalid"}`)))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("(D1-4) expected HTTP 400 for theme=invalid, got HTTP %d", resp.StatusCode)
	}
	if fakeSvc.updateSettingsCalled {
		t.Fatal("(D1-4) SECURITY/CORRECTNESS REGRESSION: an invalid theme must never reach the service/DB layer")
	}
}

// (D1-5) Omitting theme entirely (a partial update touching only another
// field) is not validated and still reaches the service layer -- proving
// the fix does not break Bug D's partial-update semantics.
func TestUpdateUserSettings_D1_ThemeOmitted_PartialUpdateStillAllowed(t *testing.T) {
	fakeSvc := &fakeUserService{}
	h := NewUserHandler(fakeSvc, zap.NewNop())

	app := fiber.New()
	userID := uuid.New()
	app.Put("/v1/api/users/me/settings", func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		return h.UpdateUserSettings(c)
	})

	req := httptest.NewRequest(http.MethodPut, "/v1/api/users/me/settings", bytes.NewReader([]byte(`{"notifications_push":false}`)))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("(D1-5) expected HTTP 200 when theme is omitted, got HTTP %d", resp.StatusCode)
	}
	if !fakeSvc.updateSettingsCalled {
		t.Fatal("(D1-5) expected a partial update omitting theme to still reach the service layer")
	}
}
