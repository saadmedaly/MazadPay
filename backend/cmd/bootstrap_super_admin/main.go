// bootstrap_super_admin is a manual, operator-run CLI tool (never invoked by
// the application at normal startup -- see cmd/server/main.go, which does
// not import or call this package) for safely creating or promoting a
// super_admin account, replacing the old pattern of seeding a hardcoded
// default credential via SQL migration (see security fix SEC-01,
// migrations/000059_neutralize_default_super_admin.up.sql, which deactivates
// that old default-credential account if it was never rotated).
//
// Requires two environment variables, with NO fallback/default value for
// either -- the process exits with a clear error if either is missing,
// rather than silently doing nothing or falling back to a known value:
//
//	SUPER_ADMIN_PHONE   the account's phone number (same format accepted by
//	                    normal registration/login)
//	SUPER_ADMIN_PIN     the PIN/password to set, hashed with the same
//	                    bcrypt.DefaultCost used everywhere else in this
//	                    codebase (see internal/services/auth_service.go) --
//	                    never logged, never printed, never written anywhere
//	                    except the bcrypt hash stored in the database
//
// Usage:
//
//	SUPER_ADMIN_PHONE=+22200000000 SUPER_ADMIN_PIN=<real PIN> \
//	  go run ./cmd/bootstrap_super_admin
//
// Idempotent: if an account with this phone already exists, it is promoted
// in place (role=super_admin, is_super_admin=true, is_active=true,
// is_verified=true) and its password_hash is updated to the freshly-hashed
// SUPER_ADMIN_PIN. If no such account exists, one is created. Either way,
// running this command is always an explicit, deliberate action taken by
// whoever set those two env vars -- there is no scenario where it fires
// unintentionally or with a value nobody chose.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/mazadpay/backend/internal/config"
	"github.com/mazadpay/backend/internal/services"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	phone := os.Getenv("SUPER_ADMIN_PHONE")
	pin := os.Getenv("SUPER_ADMIN_PIN")

	if phone == "" {
		log.Fatal("SUPER_ADMIN_PHONE is required and must not be empty -- refusing to bootstrap without an explicit phone number")
	}
	if pin == "" {
		log.Fatal("SUPER_ADMIN_PIN is required and must not be empty -- refusing to bootstrap without an explicit PIN (no default is ever used)")
	}

	// Reuse the same strength check applied to every other privileged
	// credential in this codebase (Register/ResetPassword, via
	// services.ValidatePINStrength in auth_handler.go) rather than inventing
	// a bootstrap-only rule: rejects PINs under 4 chars, all-identical
	// (1111), or simple sequential (1234/4321) values.
	if err := services.ValidatePINStrength(pin); err != nil {
		log.Fatalf("SUPER_ADMIN_PIN does not meet the application's minimum credential strength rules: %v", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(pin), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("failed to hash SUPER_ADMIN_PIN: %v", err)
	}
	// pin is no longer needed after this point; never referenced again.
	pin = ""

	cfg := config.Load()
	db, err := sqlx.Connect("postgres", fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s client_encoding=UTF8",
		cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.Name, cfg.DB.SSLMode,
	))
	if err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	var existingID uuid.UUID
	err = db.GetContext(ctx, &existingID, `SELECT id FROM users WHERE phone = $1`, phone)
	switch {
	case err == nil:
		// Existing account: promote in place, set the new credential.
		_, err = db.ExecContext(ctx, `
			UPDATE users
			SET password_hash = $1,
			    role = 'super_admin',
			    is_super_admin = TRUE,
			    is_active = TRUE,
			    is_verified = TRUE,
			    updated_at = NOW()
			WHERE id = $2
		`, string(hash), existingID)
		if err != nil {
			log.Fatalf("failed to promote existing account: %v", err)
		}
		fmt.Printf("Promoted existing account (id=%s) to super_admin and set the new credential.\n", existingID)
	case err.Error() == "sql: no rows in result set":
		newID := uuid.New()
		_, err = db.ExecContext(ctx, `
			INSERT INTO users (
				id, phone, password_hash, full_name, role, is_super_admin,
				is_verified, is_active, language_pref, notifications_enabled,
				terms_accepted_at, created_at, updated_at
			) VALUES (
				$1, $2, $3, 'Super Admin', 'super_admin', TRUE,
				TRUE, TRUE, 'ar', TRUE,
				NOW(), NOW(), NOW()
			)
		`, newID, phone, string(hash))
		if err != nil {
			log.Fatalf("failed to create new super_admin account: %v", err)
		}
		fmt.Printf("Created new super_admin account (id=%s).\n", newID)
	default:
		log.Fatalf("failed to look up existing account: %v", err)
	}

	fmt.Println("Bootstrap complete. The PIN was never logged or printed.")
}
