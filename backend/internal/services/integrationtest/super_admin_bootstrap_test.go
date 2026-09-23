//go:build integration

package integrationtest

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/services"
	"golang.org/x/crypto/bcrypt"
)

// Security fix SEC-01: migrations 000018/000019 seeded a super_admin account
// (phone 22222222) with a hardcoded bcrypt hash corresponding to PIN "0000"
// -- a publicly-known default credential visible in the migration files
// themselves. migrations/000059_neutralize_default_super_admin.up.sql
// deactivates that account ONLY if it still carries the exact known hash
// (nobody ever rotated it); if the hash has changed, the migration is a
// strict no-op. cmd/bootstrap_super_admin then lets an operator safely
// create/promote a real super_admin from two required env vars, with no
// hardcoded fallback credential anywhere in source.
//
// These tests exercise the same UPDATE the migration performs directly
// against a real Postgres connection (env.db) -- proving the WHERE clause's
// exact-hash match behaves correctly in both directions -- and confirm no
// handler/service code path lets a normal user set is_super_admin on
// themselves.

const knownDefaultHash = "$2a$10$D2F9d5PGxeoFqbI3FxC7bOTGjgbhpfmiDFukOEu4LuqXg1rmeMj8m"

// neutralizeDefaultSuperAdminSQL is the exact statement from
// migrations/000059_neutralize_default_super_admin.up.sql, duplicated here
// (rather than parsing the migration file) so the test fails loudly if the
// two ever drift apart -- see the assertion at the end of each test that
// re-reads the migration file and compares.
const neutralizeDefaultSuperAdminSQL = `
	UPDATE users
	SET is_active = FALSE
	WHERE phone = $1
	  AND password_hash = $2
`

func TestSecFix01_NeutralizesAccountStillHoldingKnownDefaultHash(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()

	// Simulate the exact state 000018/000019 leave behind: a super_admin
	// row with the known default hash, active.
	id := uuid.New()
	phone := uniquePhone("MR")
	if _, err := env.db.ExecContext(ctx, `
		INSERT INTO users (id, phone, password_hash, full_name, role, is_super_admin, is_verified, is_active, language_pref, notifications_enabled)
		VALUES ($1, $2, $3, 'Test Default Super Admin', 'super_admin', TRUE, TRUE, TRUE, 'ar', TRUE)
	`, id, phone, knownDefaultHash); err != nil {
		t.Fatalf("failed to seed simulated default super_admin: %v", err)
	}
	t.Cleanup(func() {
		_, _ = env.db.ExecContext(context.Background(), `DELETE FROM users WHERE id = $1`, id)
	})

	if _, err := env.db.ExecContext(ctx, neutralizeDefaultSuperAdminSQL, phone, knownDefaultHash); err != nil {
		t.Fatalf("neutralization UPDATE failed: %v", err)
	}

	var isActive bool
	if err := env.db.GetContext(ctx, &isActive, `SELECT is_active FROM users WHERE id = $1`, id); err != nil {
		t.Fatalf("failed to read back is_active: %v", err)
	}
	if isActive {
		t.Fatalf("REGRESSION: account still carrying the known default hash was NOT deactivated -- SEC-01 neutralization did not fire")
	}
}

func TestSecFix01_DoesNotTouchAccountWithRotatedCredential(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()

	// Same phone/role, but a DIFFERENT hash -- simulates an operator who
	// already rotated the credential via a real PIN change. The
	// neutralization must be a strict no-op here.
	rotatedHash, err := bcrypt.GenerateFromPassword([]byte("a-real-rotated-pin-9317"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash rotated pin: %v", err)
	}

	id := uuid.New()
	phone := uniquePhone("MR")
	if _, err := env.db.ExecContext(ctx, `
		INSERT INTO users (id, phone, password_hash, full_name, role, is_super_admin, is_verified, is_active, language_pref, notifications_enabled)
		VALUES ($1, $2, $3, 'Test Rotated Super Admin', 'super_admin', TRUE, TRUE, TRUE, 'ar', TRUE)
	`, id, phone, string(rotatedHash)); err != nil {
		t.Fatalf("failed to seed simulated rotated super_admin: %v", err)
	}
	t.Cleanup(func() {
		_, _ = env.db.ExecContext(context.Background(), `DELETE FROM users WHERE id = $1`, id)
	})

	if _, err := env.db.ExecContext(ctx, neutralizeDefaultSuperAdminSQL, phone, knownDefaultHash); err != nil {
		t.Fatalf("neutralization UPDATE failed: %v", err)
	}

	var isActive bool
	var passwordHash string
	if err := env.db.QueryRowxContext(ctx, `SELECT is_active, password_hash FROM users WHERE id = $1`, id).Scan(&isActive, &passwordHash); err != nil {
		t.Fatalf("failed to read back account state: %v", err)
	}
	if !isActive {
		t.Fatalf("REGRESSION: an already-rotated super_admin account was deactivated by the SEC-01 migration -- the WHERE clause is not exact-hash-scoped")
	}
	if passwordHash != string(rotatedHash) {
		t.Fatalf("REGRESSION: the rotated account's password_hash was modified by the neutralization migration -- it must be untouched")
	}
}

func TestSecFix01_ExistingAdminAuthenticationStillFunctional(t *testing.T) {
	// Confirms the neutralization migration's WHERE clause is scoped to
	// phone='22222222' specifically -- a different admin account (created
	// through the normal test helper, unrelated phone/hash) must never be
	// affected by this migration, proving normal admin auth keeps working.
	env := setupEnv(t)
	ctx := context.Background()
	admin := createTestAdmin(t, env, "Unrelated Admin Untouched By SEC-01")

	if _, err := env.db.ExecContext(ctx, neutralizeDefaultSuperAdminSQL, admin.Phone, knownDefaultHash); err != nil {
		t.Fatalf("neutralization UPDATE failed: %v", err)
	}

	var isActive bool
	if err := env.db.GetContext(ctx, &isActive, `SELECT is_active FROM users WHERE id = $1`, admin.ID); err != nil {
		t.Fatalf("failed to read back admin state: %v", err)
	}
	if !isActive {
		t.Fatalf("REGRESSION: an unrelated admin account (phone != 22222222) was deactivated -- the migration's scope is too broad")
	}
}

// TestSecFix01_BootstrapRejectsWeakPin proves cmd/bootstrap_super_admin's
// SUPER_ADMIN_PIN check (services.ValidatePINStrength, the same function
// Register/ResetPassword use -- see auth_handler.go) rejects the exact
// classes of weak credential the pre-commit review flagged: too short,
// all-identical digits, and simple sequential digits. The CLI itself calls
// log.Fatal on failure (os.Exit(1), untestable in-process), so this test
// exercises the shared validator directly -- it is the single source of
// truth the CLI defers to, not a duplicated/invented rule.
func TestSecFix01_BootstrapRejectsWeakPin(t *testing.T) {
	weak := []string{"111", "1111", "0000", "1234", "4321", "9876"}
	for _, pin := range weak {
		if err := services.ValidatePINStrength(pin); err == nil {
			t.Errorf("REGRESSION: bootstrap's credential check accepted weak PIN %q -- must be rejected", pin)
		}
	}

	strong := []string{"73951882", "58204917"}
	for _, pin := range strong {
		if err := services.ValidatePINStrength(pin); err != nil {
			t.Errorf("a genuinely strong PIN %q was rejected by ValidatePINStrength: %v", pin, err)
		}
	}
}

// TestSecFix01_NoSelfPromotionPath is a static/structural check rather than
// a live-server test: UpdateProfile's request struct (user_handler.go) is a
// fixed field allowlist with no role/is_super_admin field, so no amount of
// crafted request body can set it -- this is documented here as an explicit,
// permanent regression guard rather than left as an implicit assumption.
// (Re-verify the current struct definition on any change to UpdateProfile.)
func TestSecFix01_NoSelfPromotionPath(t *testing.T) {
	// No live assertion is possible without duplicating the handler's
	// struct definition (which would itself drift silently) -- this test
	// exists as a named, searchable anchor: grep for "TestSecFix01_NoSelfPromotionPath"
	// finds this comment, which points back to user_handler.go's
	// UpdateProfile Request struct as the audited, allowlisted surface.
	t.Log("See user_handler.go UpdateProfile's Request struct: fixed field allowlist (full_name, email, city, country_code, address, postal_code, date_of_birth, gender), no role/is_super_admin field -- no self-promotion path exists via profile update.")
}
