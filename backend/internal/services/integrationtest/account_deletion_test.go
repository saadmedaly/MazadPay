//go:build integration

// Self-service account deletion (App Store guideline 5.1.1(v)): DELETE /users/me
// through the REAL JWT middleware and UserHandler.DeleteMe against Postgres +
// Redis. Proves personal data is cleared, the phone becomes reusable, sessions
// die, financial/auction history survives, and every open obligation blocks
// deletion without touching money.
package integrationtest

import (
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/config"
	"github.com/mazadpay/backend/internal/middleware"
	"github.com/mazadpay/backend/internal/models"
	"github.com/shopspring/decimal"
)

const deletionTestPin = "StrongPass123"

type deletionUser struct {
	ID      uuid.UUID
	Phone   string // as typed at registration/login
	DBPhone string // as stored in users.phone
	Token   string
}

// deletionApp wires /users/me exactly like routes.go: real JWT middleware
// (signature, blacklist, user-level revocation) in front of the real handler.
func deletionApp(env *testEnv) *fiber.App {
	cfg := config.Load()
	if cfg.JWT.Secret == "" {
		cfg.JWT.Secret = "test-secret-key-for-integration-tests"
	}
	app := fiber.New()
	auth := middleware.JWT(cfg.JWT.Secret, env.logger, env.rdb)
	app.Get("/users/me", auth, env.userHandler.GetMe)
	app.Delete("/users/me", auth, env.userHandler.DeleteMe)
	return app
}

// registerAndLogin creates a real MR account and returns a real JWT for it.
func registerAndLogin(t *testing.T, env *testEnv) deletionUser {
	t.Helper()
	ctx := context.Background()
	phone := uniquePhone("MR")
	if err := env.authSvc.Register(ctx, phone, deletionTestPin, uniqueName("Deletion"), "deletion-"+uuid.New().String()[:8]+"@example.com", "Nouakchott", "MR"); err != nil {
		t.Fatalf("register: %v", err)
	}
	token, user, err := env.authSvc.Login(ctx, phone, "MR", deletionTestPin)
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	// Wallet row (balance 0) like a real user who opened the wallet screen.
	if _, err := env.walletRepo.GetByUserID(ctx, user.ID); err != nil {
		t.Fatalf("wallet: %v", err)
	}
	return deletionUser{ID: user.ID, Phone: phone, DBPhone: user.Phone, Token: token}
}

func callDeleteMe(t *testing.T, app *fiber.App, token string) (int, string) {
	t.Helper()
	req := httptest.NewRequest("DELETE", "/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req, 10000)
	if err != nil {
		t.Fatalf("DELETE /users/me: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	var parsed struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	_ = json.Unmarshal(body, &parsed)
	return resp.StatusCode, parsed.Error.Code
}

func callGetMe(t *testing.T, app *fiber.App, token string) int {
	t.Helper()
	req := httptest.NewRequest("GET", "/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req, 10000)
	if err != nil {
		t.Fatalf("GET /users/me: %v", err)
	}
	return resp.StatusCode
}

func countRows(t *testing.T, env *testEnv, query string, args ...interface{}) int {
	t.Helper()
	var n int
	if err := env.db.Get(&n, query, args...); err != nil {
		t.Fatalf("count %q: %v", query, err)
	}
	return n
}

func TestAccountDeletion_ClearsPersonalDataAndPreservesHistory(t *testing.T) {
	env := setupEnv(t)
	app := deletionApp(env)
	ctx := context.Background()

	u := registerAndLogin(t, env)
	other := registerAndLogin(t, env)
	seller := createTestUser(t, env, "Deletion Seller")

	// Settled history the deletion must keep.
	ended := createTestAuction(t, env, seller.ID, "MR", "MRU")
	if _, err := env.db.Exec(`UPDATE auctions SET status = 'ended', winner_id = $1, payment_deadline = now() - interval '1 day' WHERE id = $2`, other.ID, ended.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := env.db.Exec(`INSERT INTO bids (auction_id, user_id, amount) VALUES ($1, $2, 110)`, ended.ID, u.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := env.db.Exec(`INSERT INTO transactions (user_id, type, amount, status) VALUES ($1, 'deposit', 500, 'completed'), ($1, 'withdraw', 500, 'completed')`, u.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := env.db.Exec(`INSERT INTO wallet_holds (user_id, auction_id, amount, status, released_at) VALUES ($1, $2, 20, 'released', now())`, u.ID, ended.ID); err != nil {
		t.Fatal(err)
	}

	// Personal, non-financial data the deletion must remove.
	if _, err := env.db.Exec(`INSERT INTO push_tokens (user_id, fcm_token, device_id, platform) VALUES ($1, $2, 'dev', 'ios')`, u.ID, "fcm-"+uuid.New().String()); err != nil {
		t.Fatal(err)
	}
	if _, err := env.db.Exec(`INSERT INTO user_settings (user_id) VALUES ($1) ON CONFLICT DO NOTHING`, u.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := env.db.Exec(`INSERT INTO user_favorites (user_id, auction_id) VALUES ($1, $2)`, u.ID, ended.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := env.db.Exec(`INSERT INTO notifications (user_id, type, title) VALUES ($1, 'general', 'hello')`, u.ID); err != nil {
		t.Fatal(err)
	}
	pendingReq := newAuctionRequest(u.ID, "deletion")
	if err := env.reqRepo.CreateAuctionRequest(ctx, pendingReq); err != nil {
		t.Fatalf("pending auction request: %v", err)
	}

	txBefore := countRows(t, env, `SELECT count(*) FROM transactions WHERE user_id = $1`, u.ID)
	bidsBefore := countRows(t, env, `SELECT count(*) FROM bids WHERE user_id = $1`, u.ID)
	holdsBefore := countRows(t, env, `SELECT count(*) FROM wallet_holds WHERE user_id = $1`, u.ID)

	// Revocation compares whole seconds with a strict <; make the login token
	// strictly older than the deletion.
	time.Sleep(1100 * time.Millisecond)

	if status, code := callDeleteMe(t, app, u.Token); status != 200 {
		t.Fatalf("DELETE /users/me = %d (%s), want 200", status, code)
	}

	var row struct {
		IsActive          bool    `db:"is_active"`
		Phone             string  `db:"phone"`
		PhoneE164         *string `db:"phone_e164"`
		PhoneCountryISO   *string `db:"phone_country_iso"`
		AccountCountryISO *string `db:"account_country_iso"`
		PasswordHash      string  `db:"password_hash"`
		FullName          *string `db:"full_name"`
		Email             *string `db:"email"`
		City              *string `db:"city"`
	}
	if err := env.db.Get(&row, `SELECT is_active, phone, phone_e164, phone_country_iso, account_country_iso, password_hash, full_name, email, city FROM users WHERE id = $1`, u.ID); err != nil {
		t.Fatalf("user row must still exist (no hard delete): %v", err)
	}
	if row.IsActive {
		t.Error("is_active still true")
	}
	if row.Phone == u.DBPhone || len(row.Phone) < 4 || row.Phone[:4] != "del_" {
		t.Errorf("phone not anonymized: %q", row.Phone)
	}
	if row.PhoneE164 != nil || row.PhoneCountryISO != nil || row.AccountCountryISO != nil {
		t.Errorf("phone_e164/country ISO not cleared: %v %v %v", row.PhoneE164, row.PhoneCountryISO, row.AccountCountryISO)
	}
	if row.Email != nil || row.City != nil {
		t.Errorf("email/city not cleared: %v %v", row.Email, row.City)
	}
	if row.FullName == nil || *row.FullName != "Deleted User" {
		t.Errorf("full_name = %v", row.FullName)
	}
	if len(row.PasswordHash) < 8 || row.PasswordHash[:8] != "deleted:" {
		t.Error("password_hash not replaced with an unusable value")
	}

	for table, want := range map[string]int{"push_tokens": 0, "user_settings": 0, "user_favorites": 0, "notifications": 0} {
		if got := countRows(t, env, `SELECT count(*) FROM `+table+` WHERE user_id = $1`, u.ID); got != want {
			t.Errorf("%s rows = %d, want %d", table, got, want)
		}
	}
	var reqStatus string
	if err := env.db.Get(&reqStatus, `SELECT status FROM auction_requests WHERE id = $1`, pendingReq.ID); err != nil || reqStatus != "rejected" {
		t.Errorf("pending auction request status = %q (%v), want rejected", reqStatus, err)
	}

	// Financial / auction history untouched.
	if got := countRows(t, env, `SELECT count(*) FROM transactions WHERE user_id = $1`, u.ID); got != txBefore {
		t.Errorf("transactions %d -> %d", txBefore, got)
	}
	if got := countRows(t, env, `SELECT count(*) FROM bids WHERE user_id = $1`, u.ID); got != bidsBefore {
		t.Errorf("bids %d -> %d", bidsBefore, got)
	}
	if got := countRows(t, env, `SELECT count(*) FROM wallet_holds WHERE user_id = $1`, u.ID); got != holdsBefore {
		t.Errorf("wallet_holds %d -> %d", holdsBefore, got)
	}
	if got := countRows(t, env, `SELECT count(*) FROM wallets WHERE user_id = $1`, u.ID); got != 1 {
		t.Errorf("wallet row count = %d, want 1", got)
	}
	if got := countRows(t, env, `SELECT count(*) FROM audit_logs WHERE action = 'user_account_deleted' AND entity_id = $1`, u.ID); got != 1 {
		t.Errorf("audit log rows = %d, want 1", got)
	}

	// Only the caller was deleted.
	var otherActive bool
	if err := env.db.Get(&otherActive, `SELECT is_active FROM users WHERE id = $1`, other.ID); err != nil || !otherActive {
		t.Errorf("other user affected: active=%v err=%v", otherActive, err)
	}
	if status := callGetMe(t, app, other.Token); status != 200 {
		t.Errorf("other user's token: GET /users/me = %d, want 200", status)
	}

	// Old session dead; deleting again with it is also refused.
	if status := callGetMe(t, app, u.Token); status != 401 {
		t.Errorf("old token: GET /users/me = %d, want 401", status)
	}
	if status, _ := callDeleteMe(t, app, u.Token); status != 401 {
		t.Errorf("old token: DELETE /users/me = %d, want 401", status)
	}

	// Cannot log in with the old credentials.
	if _, _, err := env.authSvc.Login(ctx, u.Phone, "MR", deletionTestPin); err == nil {
		t.Error("deleted user could still log in")
	}

	// Same phone can register a brand-new account.
	if err := env.authSvc.Register(ctx, u.Phone, deletionTestPin, uniqueName("Reuse"), "", "", "MR"); err != nil {
		t.Fatalf("re-register with the deleted phone: %v", err)
	}
	_, reborn, err := env.authSvc.Login(ctx, u.Phone, "MR", deletionTestPin)
	if err != nil {
		t.Fatalf("login as the new account: %v", err)
	}
	if reborn.ID == u.ID {
		t.Error("re-registration revived the deleted account instead of creating a new one")
	}
}

// Every open obligation refuses deletion with its own code and changes nothing.
func TestAccountDeletion_BlockedByOpenObligations(t *testing.T) {
	env := setupEnv(t)
	app := deletionApp(env)

	cases := []struct {
		name  string
		code  string
		setup func(t *testing.T, u deletionUser)
	}{
		{"positive balance", "account_deletion_blocked_balance", func(t *testing.T, u deletionUser) {
			creditWallet(t, env, u.ID, decimal.NewFromInt(250))
		}},
		{"active hold", "account_deletion_blocked_hold", func(t *testing.T, u deletionUser) {
			a := createTestAuction(t, env, createTestUser(t, env, "Hold Seller").ID, "MR", "MRU")
			if _, err := env.db.Exec(`INSERT INTO wallet_holds (user_id, auction_id, amount, status) VALUES ($1, $2, 20, 'active')`, u.ID, a.ID); err != nil {
				t.Fatal(err)
			}
		}},
		{"pending deposit", "account_deletion_blocked_pending_deposit", func(t *testing.T, u deletionUser) {
			if _, err := env.db.Exec(`INSERT INTO transactions (user_id, type, amount, status) VALUES ($1, 'deposit', 100, 'pending_review')`, u.ID); err != nil {
				t.Fatal(err)
			}
		}},
		{"pending withdrawal", "account_deletion_blocked_pending_withdrawal", func(t *testing.T, u deletionUser) {
			if _, err := env.db.Exec(`INSERT INTO transactions (user_id, type, amount, status) VALUES ($1, 'withdraw', 100, 'pending_review')`, u.ID); err != nil {
				t.Fatal(err)
			}
		}},
		{"leading bid on live auction", "account_deletion_blocked_leading_bid", func(t *testing.T, u deletionUser) {
			a := createTestAuction(t, env, createTestUser(t, env, "Live Seller").ID, "MR", "MRU")
			if _, err := env.db.Exec(`INSERT INTO bids (auction_id, user_id, amount) VALUES ($1, $2, 150)`, a.ID, u.ID); err != nil {
				t.Fatal(err)
			}
			if _, err := env.db.Exec(`UPDATE auctions SET last_bidder_id = $1, current_price = 150 WHERE id = $2`, u.ID, a.ID); err != nil {
				t.Fatal(err)
			}
		}},
		{"unpaid win", "account_deletion_blocked_unpaid_win", func(t *testing.T, u deletionUser) {
			a := createTestAuction(t, env, createTestUser(t, env, "Win Seller").ID, "MR", "MRU")
			if _, err := env.db.Exec(`UPDATE auctions SET status = 'ended', winner_id = $1, payment_deadline = now() + interval '48 hours' WHERE id = $2`, u.ID, a.ID); err != nil {
				t.Fatal(err)
			}
		}},
		{"live auction as seller", "account_deletion_blocked_active_listing", func(t *testing.T, u deletionUser) {
			createTestAuction(t, env, u.ID, "MR", "MRU")
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			u := registerAndLogin(t, env)
			tc.setup(t, u)

			var before struct {
				Balance decimal.Decimal `db:"balance"`
				Frozen  decimal.Decimal `db:"frozen_amount"`
			}
			if err := env.db.Get(&before, `SELECT balance, frozen_amount FROM wallets WHERE user_id = $1`, u.ID); err != nil {
				t.Fatal(err)
			}
			txBefore := countRows(t, env, `SELECT count(*) FROM transactions WHERE user_id = $1`, u.ID)

			status, code := callDeleteMe(t, app, u.Token)
			if status != 409 || code != tc.code {
				t.Fatalf("DELETE /users/me = %d %q, want 409 %q", status, code, tc.code)
			}

			var user models.User
			if err := env.db.Get(&user, `SELECT * FROM users WHERE id = $1`, u.ID); err != nil {
				t.Fatal(err)
			}
			if !user.IsActive || user.Phone != u.DBPhone || user.PhoneE164 == nil {
				t.Errorf("blocked deletion still modified the user: active=%v phone=%q", user.IsActive, user.Phone)
			}
			var after struct {
				Balance decimal.Decimal `db:"balance"`
				Frozen  decimal.Decimal `db:"frozen_amount"`
			}
			if err := env.db.Get(&after, `SELECT balance, frozen_amount FROM wallets WHERE user_id = $1`, u.ID); err != nil {
				t.Fatal(err)
			}
			if !after.Balance.Equal(before.Balance) || !after.Frozen.Equal(before.Frozen) {
				t.Errorf("wallet changed: %v/%v -> %v/%v", before.Balance, before.Frozen, after.Balance, after.Frozen)
			}
			if got := countRows(t, env, `SELECT count(*) FROM transactions WHERE user_id = $1`, u.ID); got != txBefore {
				t.Errorf("transactions %d -> %d", txBefore, got)
			}
			if s := callGetMe(t, app, u.Token); s != 200 {
				t.Errorf("session revoked by a blocked deletion: GET /users/me = %d", s)
			}
		})
	}
}
