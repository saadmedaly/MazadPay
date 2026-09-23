//go:build integration

package integrationtest

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/mazadpay/backend/internal/handlers"
)

// Security fix confirmation, SEC-04 (user enumeration via login errors):
// AuthHandler.Login (auth_handler.go:198-216) already unifies
// ErrInvalidPin, ErrUserNotFound, and ErrPhoneUnavailable into the exact
// same external response -- 401, code "invalid_credentials", message
// "Invalid phone or PIN" -- before this security round; this is a
// permanent regression guard proving that invariant at the real HTTP
// layer, not just by reading the source. Internal logging still
// distinguishes the causes (see auth_handler.go's own comment and
// auth_service.go's recordFailedLogin / zap fields), which this test does
// not touch.
func TestLogin_UnknownUserAndWrongPin_ProduceIdenticalExternalResponse(t *testing.T) {
	env := setupEnv(t)
	authHandler := handlers.NewAuthHandler(env.authSvc, env.logger, env.rdb)

	app := fiber.New()
	app.Post("/auth/login", authHandler.Login)

	// A real, registered user -- used for the "wrong PIN" case.
	user := createTestUser(t, env, "TEST LOGIN ENUM REAL USER")

	unknownStatus, unknownBody := doLogin(t, app, uniquePhone("MR"), "0000")
	wrongPinStatus, wrongPinBody := doLogin(t, app, user.Phone, "9999999") // wrong pin, real user

	if unknownStatus != wrongPinStatus {
		t.Fatalf("SEC-04 REGRESSION: unknown-user login returned status %d but wrong-PIN login returned %d -- these must be identical", unknownStatus, wrongPinStatus)
	}
	if unknownStatus != fiber.StatusUnauthorized {
		t.Fatalf("expected 401 for both cases, got %d", unknownStatus)
	}

	var unknownJSON, wrongPinJSON map[string]interface{}
	if err := json.Unmarshal([]byte(unknownBody), &unknownJSON); err != nil {
		t.Fatalf("failed to parse unknown-user response body: %v (body=%s)", err, unknownBody)
	}
	if err := json.Unmarshal([]byte(wrongPinBody), &wrongPinJSON); err != nil {
		t.Fatalf("failed to parse wrong-pin response body: %v (body=%s)", err, wrongPinBody)
	}

	unknownErr, _ := unknownJSON["error"].(map[string]interface{})
	wrongPinErr, _ := wrongPinJSON["error"].(map[string]interface{})
	if unknownErr == nil || wrongPinErr == nil {
		t.Fatalf("expected both responses to have an error object; unknown=%v wrongPin=%v", unknownJSON, wrongPinJSON)
	}
	if unknownErr["code"] != wrongPinErr["code"] {
		t.Fatalf("SEC-04 REGRESSION: error.code differs between unknown-user (%v) and wrong-PIN (%v) -- allows account enumeration", unknownErr["code"], wrongPinErr["code"])
	}
	if unknownErr["message"] != wrongPinErr["message"] {
		t.Fatalf("SEC-04 REGRESSION: error.message differs between unknown-user (%v) and wrong-PIN (%v) -- allows account enumeration", unknownErr["message"], wrongPinErr["message"])
	}
	if unknownErr["code"] != "invalid_credentials" {
		t.Fatalf(`expected error.code="invalid_credentials", got %v`, unknownErr["code"])
	}
}

func doLogin(t *testing.T, app *fiber.App, phone, pin string) (int, string) {
	t.Helper()
	body := `{"phone":"` + phone + `","pin":"` + pin + `"}`
	req := httptest.NewRequest("POST", "/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	return resp.StatusCode, string(b)
}
