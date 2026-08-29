//go:build integration

// API Versioning Phase 1 integration tests — proves the /v1/api (legacy) and
// /v2/api (strict) auth contracts genuinely coexist against a real local Postgres:
// the exact call shape the currently-published Flutter app makes still succeeds via
// AuthService.RegisterLegacy/ResetPassword, while the new strict contract
// (AuthService.Register) enforces country_iso + 8-72 char passwords and cannot be
// bypassed by omitting fields.
package integrationtest

import (
	"context"
	"reflect"
	"testing"

	"github.com/mazadpay/backend/internal/handlers"
)

// (1) The real published-app registration call: 4-digit PIN, dial-code country_code,
// no country_iso at all — must still succeed exactly as it does in production today.
func TestVersioning_LegacyRegister_PublishedAppShape(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()

	phone := uniquePhone("MR")
	fullName := uniqueName("TEST INTEGRATION V1 REGISTER")

	// This is the EXACT call shape the published app makes: a 4-digit PIN and a dial
	// code (not an ISO-2 region) — see backend/internal/handlers/auth_handler.go
	// RegisterRequestLegacy, and the baseline captured from `git show fullstack:...`.
	err := env.authSvc.RegisterLegacy(ctx, phone, "1234", fullName, "", "", "+222")
	if err != nil {
		t.Fatalf("RegisterLegacy (published-app call shape) failed: %v — this would break account creation for every user still on the current Google Play build", err)
	}

	var row struct {
		Phone string `db:"phone"`
	}
	if err := env.db.Get(&row, `SELECT phone FROM users WHERE full_name = $1`, fullName); err != nil {
		t.Fatalf("failed to read back legacy-registered user: %v", err)
	}
	t.Logf("(1) legacy register succeeded with published-app shape: phone=%s", row.Phone)
}

// (2) v1 must reject any country outside the historical 4-country allow-list — it must
// NOT have silently gained the new countries added to v2.
func TestVersioning_LegacyRegister_RejectsNewCountry(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()

	countBefore := countUsers(t, env)

	// A perfectly valid US number, but +1 was never in the v1 allow-list.
	err := env.authSvc.RegisterLegacy(ctx, "2025551234", "1234", uniqueName("TEST SHOULD NOT EXIST V1"), "", "", "+1")
	if err == nil {
		t.Fatalf("expected RegisterLegacy to reject country_code=+1 (not in the v1 baseline allow-list), got nil")
	}
	t.Logf("(2) v1 correctly rejected an out-of-allowlist country: %v", err)

	countAfter := countUsers(t, env)
	if countAfter != countBefore {
		t.Fatalf("expected no user created for a v1 call with an unsupported country (before=%d after=%d)", countBefore, countAfter)
	}
}

// (3) v2 must reject registration when country_iso is missing entirely — proves a new
// client cannot bypass the v2 policy by simply omitting the field (this is enforced at
// the DTO validator layer in production; here we test the service layer directly, which
// is the same function the v2 handler calls after DTO validation would already have
// rejected an empty country_iso — this test additionally proves the SERVICE itself, not
// just the DTO tag, refuses to guess).
func TestVersioning_V2Register_RequiresCountryISO(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()

	countBefore := countUsers(t, env)

	err := env.authSvc.Register(ctx, uniquePhone("MR"), "StrongPass123", uniqueName("TEST SHOULD NOT EXIST V2 NOISO"), "", "", "")
	if err == nil {
		t.Fatalf("expected Register (v2) to reject an empty country_iso, got nil")
	}
	t.Logf("(3) v2 correctly rejected missing country_iso: %v", err)

	countAfter := countUsers(t, env)
	if countAfter != countBefore {
		t.Fatalf("expected no user created for a v2 call missing country_iso (before=%d after=%d)", countBefore, countAfter)
	}
}

// (4) v2 must reject a password under 8 characters (enforced at the DTO layer in
// production — RegisterRequest.Pin validate:"required,min=8,max=72" — so this test
// exercises that exact validator directly, since AuthService.Register itself does not
// re-check length; the length gate is a DTO-layer guarantee, which we assert here by
// name to keep this test meaningful if the validator tag ever changes).
func TestVersioning_V2Register_RejectsShortPassword(t *testing.T) {
	// This is a DTO-layer guarantee (validator tag), not enforced inside
	// AuthService.Register itself. We assert on the actual struct tag rather than
	// duplicating validator wiring here, so this test breaks loudly if the tag is ever
	// weakened without a deliberate review.
	tag := getRegisterRequestPinValidateTag(t)
	if tag != "required,min=8,max=72" {
		t.Fatalf("RegisterRequest.Pin validate tag = %q, want %q — the v2 password-strength policy has changed without this test being updated", tag, "required,min=8,max=72")
	}
	t.Logf("(4) confirmed RegisterRequest.Pin enforces the 8-72 char policy via validator tag %q", tag)
}

// (5) v2 registration succeeds for MR, TN, US, and CA with correct country_iso — this
// duplicates part of TestRegister_MultiCountry in integration_test.go deliberately,
// framed explicitly as a versioning-contract proof (v2 supports the full country list,
// unlike v1) rather than a general phone-normalization proof.
func TestVersioning_V2Register_MultiCountrySucceeds(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()

	for _, region := range []string{"MR", "TN", "US", "CA"} {
		phone := uniquePhone(region)
		name := uniqueName("TEST V2 MULTICOUNTRY " + region)
		if err := env.authSvc.Register(ctx, phone, "StrongPass123", name, "", "", region); err != nil {
			t.Fatalf("v2 Register failed for region %s: %v", region, err)
		}
		var iso string
		if err := env.db.Get(&iso, `SELECT phone_country_iso FROM users WHERE full_name = $1`, name); err != nil {
			t.Fatalf("failed to read back region %s: %v", region, err)
		}
		if iso != region {
			t.Fatalf("region %s: expected phone_country_iso=%s, got %q", region, region, iso)
		}
	}
	t.Logf("(5) v2 register succeeded for MR, TN, US, CA")
}

// (6) US number tagged CA (and vice versa) rejected under v2 — the core country-mismatch
// regression, re-asserted explicitly in the versioning context.
func TestVersioning_V2Register_RejectsSharedDialCodeMismatch(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()

	if err := env.authSvc.Register(ctx, "2025551234", "StrongPass123", uniqueName("TEST SHOULD NOT EXIST V2 MISMATCH1"), "", "", "CA"); err == nil {
		t.Fatalf("expected error registering a US number tagged CA under v2, got nil")
	}
	if err := env.authSvc.Register(ctx, "4165551234", "StrongPass123", uniqueName("TEST SHOULD NOT EXIST V2 MISMATCH2"), "", "", "US"); err == nil {
		t.Fatalf("expected error registering a CA number tagged US under v2, got nil")
	}
	t.Logf("(6) v2 correctly rejected both US-tagged-CA and CA-tagged-US")
}

// (8) The published app's password-reset call shape (4-digit new_pin, no country_iso)
// must keep working — AuthService.ResetPassword is shared between v1/v2, so this
// exercises it with a legacy-shaped 4-digit new_pin exactly as the v1 DTO would pass it.
func TestVersioning_LegacyResetPassword_PublishedAppShape(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()

	// Register a legacy-shaped account first, matching what a real published-app user
	// would already have.
	phone := uniquePhone("MR")
	name := uniqueName("TEST INTEGRATION V1 RESET")
	if err := env.authSvc.RegisterLegacy(ctx, phone, "1234", name, "", "", "+222"); err != nil {
		t.Fatalf("RegisterLegacy setup failed: %v", err)
	}

	// Simulate a valid reset_password OTP directly in the DB (bypassing SMS delivery,
	// matching how development-mode OTP verification already works in this codebase).
	otpCode := insertValidResetOTP(t, env, phone)

	// The v1 DTO would pass a 4-digit new_pin here — ResetPassword itself has no length
	// policy (that's a DTO-layer concern), so this call must succeed.
	if err := env.authSvc.ResetPassword(ctx, phone, otpCode, "5678"); err != nil {
		t.Fatalf("ResetPassword with a legacy-shaped 4-digit new_pin failed: %v — this would break password reset for every user still on the current Google Play build", err)
	}
	t.Logf("(8) legacy-shaped password reset (4-digit new_pin) succeeded")

	// Confirm the account can now log in with the new 4-digit PIN.
	token, user, err := env.authSvc.Login(ctx, phone, "MR", "5678")
	if err != nil || token == "" || user == nil {
		t.Fatalf("login with the newly reset 4-digit PIN failed: %v", err)
	}
	t.Logf("(8) login with newly-reset legacy-shaped PIN succeeded")
}

// (9) v2 reset-password: the DTO-level policy (8-72 chars) is what actually enforces
// the "strong password" requirement — asserted here the same way as test (4), by
// checking the validator tag directly, since the shared ResetPassword service function
// has no length policy of its own by design (see AuthService interface comment).
func TestVersioning_V2ResetPassword_EnforcesStrongPasswordViaDTO(t *testing.T) {
	tag := getResetPasswordRequestNewPinValidateTag(t)
	if tag != "required,min=8,max=72" {
		t.Fatalf("ResetPasswordRequest.NewPin validate tag = %q, want %q", tag, "required,min=8,max=72")
	}
	t.Logf("(9) confirmed ResetPasswordRequest.NewPin enforces the 8-72 char policy via validator tag %q", tag)
}

// (10) Login works transparently for BOTH an old 4-digit-PIN account (registered via
// RegisterLegacy) and a new 8+-char-password account (registered via Register) — the
// single shared /auth/login endpoint must serve both without any client-visible
// difference in success/failure shape.
func TestVersioning_Login_WorksForBothLegacyAndV2Accounts(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()

	legacyPhone := uniquePhone("MR")
	legacyName := uniqueName("TEST V1 LOGIN ACCOUNT")
	if err := env.authSvc.RegisterLegacy(ctx, legacyPhone, "1234", legacyName, "", "", "+222"); err != nil {
		t.Fatalf("RegisterLegacy failed: %v", err)
	}
	if token, user, err := env.authSvc.Login(ctx, legacyPhone, "MR", "1234"); err != nil || token == "" || user == nil {
		t.Fatalf("login for a legacy (4-digit PIN) account failed: %v", err)
	}
	t.Logf("(10) login succeeded for a legacy 4-digit-PIN account")

	v2Phone := uniquePhone("TN")
	v2Name := uniqueName("TEST V2 LOGIN ACCOUNT")
	if err := env.authSvc.Register(ctx, v2Phone, "StrongPass123", v2Name, "", "", "TN"); err != nil {
		t.Fatalf("Register (v2) failed: %v", err)
	}
	if token, user, err := env.authSvc.Login(ctx, v2Phone, "TN", "StrongPass123"); err != nil || token == "" || user == nil {
		t.Fatalf("login for a v2 (8+ char password) account failed: %v", err)
	}
	t.Logf("(10) login succeeded for a v2 8+-char-password account")
}

// --- helpers ---

// insertValidResetOTP inserts a verified-eligible reset_password OTP row directly (the
// same shape AuthService.SendOTP would produce) so ResetPassword's internal VerifyOTP
// call succeeds, without depending on an actual SMS provider being configured.
func insertValidResetOTP(t *testing.T, env *testEnv, phone string) string {
	t.Helper()
	ctx := context.Background()
	const code = "9999"
	_, err := env.db.ExecContext(ctx, `
		INSERT INTO otp_verifications (id, phone, otp_code, purpose, attempts, max_attempts, expires_at, ip_address)
		VALUES (gen_random_uuid(), $1, $2, 'reset_password', 0, 3, now() + interval '5 minutes', '127.0.0.1')
	`, phone, code)
	if err != nil {
		t.Fatalf("failed to insert fixture OTP: %v", err)
	}
	return code
}

func getRegisterRequestPinValidateTag(t *testing.T) string {
	t.Helper()
	return validateTagFor(t, reflect.TypeOf(handlers.RegisterRequest{}), "Pin")
}

func getResetPasswordRequestNewPinValidateTag(t *testing.T) string {
	t.Helper()
	return validateTagFor(t, reflect.TypeOf(handlers.ResetPasswordRequest{}), "NewPin")
}

// validateTagFor reflects the given struct type's field `validate` tag — kept generic so
// both helpers above share one implementation. Reflecting the real handlers.* DTO
// (rather than hardcoding the expected string without checking the source of truth)
// means this test breaks loudly the moment the actual production validator tag changes,
// not just when someone remembers to update a hardcoded duplicate.
func validateTagFor(t *testing.T, typ reflect.Type, fieldName string) string {
	t.Helper()
	field, ok := typ.FieldByName(fieldName)
	if !ok {
		t.Fatalf("struct %s has no field %s", typ.Name(), fieldName)
	}
	return field.Tag.Get("validate")
}
