//go:build integration

package integrationtest

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/handlers"
	"github.com/mazadpay/backend/internal/middleware"
	"github.com/shopspring/decimal"
)

// Financial Audit Phase 6 (Admin RBAC review): AdminAddBalance/
// AdminDeductBalance are gated by AdminOnly, not SuperAdminOnly. Reviewed
// against the codebase's own existing pattern (routes.go): SuperAdminOnly is
// reserved specifically for IRREVERSIBLE/structural admin actions (delete
// user, delete location/country, delete/toggle payment method, delete
// driver, generate an admin invitation) -- every one of those either
// destroys a row other data may reference, or grants a NEW admin identity.
// AdminOnly gates REVERSIBLE/operational actions, including every other
// wallet-review action (deposit/withdrawal approve-or-reject via
// ValidateTransaction, itself also AdminOnly). admin_credit/admin_debit fit
// this existing pattern: a mistaken credit/debit is correctable by an
// opposite entry from any other admin, it does not destroy data or mint a
// new privileged identity. Per the audit's own instruction not to change
// RBAC without evidence of a genuine gap, no RBAC change was made -- this
// file instead adds the missing HTTP-level abuse test proving the CURRENT
// boundary actually holds (previously untested: the only existing
// AdminAddBalance/AdminDeductBalance tests call the service layer directly,
// bypassing the route/middleware chain entirely).
//
// fakeAuthWithRole mirrors fakeAuth (integration_test.go) but additionally
// sets user_role/is_super_admin -- the two Locals keys AdminOnly/
// SuperAdminOnly actually check (see middleware/auth.go) -- since the real
// JWT middleware is what populates them from token claims in production,
// and these tests exercise the unmodified AdminOnly/SuperAdminOnly
// middleware itself, not a stand-in.
func fakeAuthWithRole(userID uuid.UUID, role string, isSuperAdmin bool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		c.Locals("user_role", role)
		c.Locals("is_super_admin", isSuperAdmin)
		return c.Next()
	}
}

func TestAdminWalletRBAC_NormalUserCannotCallAddBalance(t *testing.T) {
	env := setupEnv(t)
	normalUser := createTestUser(t, env, "TEST RBAC NORMAL USER ADD BALANCE")
	target := createTestUser(t, env, "TEST RBAC TARGET ADD BALANCE")

	adminSvc := newTestAdminService(t, env)
	h := handlers.NewAdminHandler(adminSvc, nil, env.logger)

	app := fiber.New()
	app.Post("/admin/transactions/:id/add-balance",
		fakeAuthWithRole(normalUser.ID, "user", false),
		middleware.AdminOnly(env.logger),
		h.AdminAddBalance)

	// No real anchor transaction is needed -- AdminOnly must deny before the
	// handler is ever reached, so any well-formed UUID in the path suffices.
	req := httptest.NewRequest("POST", "/admin/transactions/"+uuid.New().String()+"/add-balance",
		strings.NewReader(`{"amount":999999,"notes":"rbac abuse attempt"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	if resp.StatusCode != 403 {
		t.Fatalf("SECURITY REGRESSION: expected 403 for a normal user calling admin add-balance, got %d", resp.StatusCode)
	}

	wallet, err := env.walletRepo.GetByUserID(context.Background(), target.ID)
	if err != nil {
		t.Fatalf("failed to read target wallet: %v", err)
	}
	if wallet.Balance.GreaterThan(decimal.Zero) {
		t.Fatalf("CRITICAL: target wallet balance is %s, a normal-user RBAC bypass must never move money", wallet.Balance)
	}
}

func TestAdminWalletRBAC_NormalUserCannotCallDeductBalance(t *testing.T) {
	env := setupEnv(t)
	normalUser := createTestUser(t, env, "TEST RBAC NORMAL USER DEDUCT BALANCE")
	target := createTestUser(t, env, "TEST RBAC TARGET DEDUCT BALANCE")
	creditWallet(t, env, target.ID, decimal.NewFromInt(1000))

	adminSvc := newTestAdminService(t, env)
	h := handlers.NewAdminHandler(adminSvc, nil, env.logger)

	app := fiber.New()
	app.Post("/admin/transactions/:id/deduct-balance",
		fakeAuthWithRole(normalUser.ID, "user", false),
		middleware.AdminOnly(env.logger),
		h.AdminDeductBalance)

	req := httptest.NewRequest("POST", "/admin/transactions/"+uuid.New().String()+"/deduct-balance",
		strings.NewReader(`{"amount":500,"notes":"rbac abuse attempt"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	if resp.StatusCode != 403 {
		t.Fatalf("SECURITY REGRESSION: expected 403 for a normal user calling admin deduct-balance, got %d", resp.StatusCode)
	}

	wallet, err := env.walletRepo.GetByUserID(context.Background(), target.ID)
	if err != nil {
		t.Fatalf("failed to read target wallet: %v", err)
	}
	if !wallet.Balance.Equal(decimal.NewFromInt(1000)) {
		t.Fatalf("CRITICAL: target wallet balance changed to %s from a denied normal-user RBAC bypass attempt, expected unchanged 1000", wallet.Balance)
	}
}

// TestAdminWalletRBAC_RegularAdminCanCallAddBalance is the control case:
// confirms AdminOnly (the CURRENT, deliberately-not-changed gate) still
// permits a regular admin -- proving the two denial tests above are
// actually testing the role boundary, not some unrelated failure (e.g. a
// broken route).
func TestAdminWalletRBAC_RegularAdminCanCallAddBalance(t *testing.T) {
	env := setupEnv(t)
	target := createTestUser(t, env, "TEST RBAC CONTROL TARGET")

	adminSvc := newTestAdminService(t, env)
	h := handlers.NewAdminHandler(adminSvc, nil, env.logger)

	app := fiber.New()
	app.Post("/admin/transactions/:id/add-balance",
		fakeAuthWithRole(uuid.New(), "admin", false),
		middleware.AdminOnly(env.logger),
		h.AdminAddBalance)

	walletSvc := newWalletSvc(env)
	anchor, err := walletSvc.InitiateDeposit(context.Background(), target.ID, decimal.NewFromInt(50), "bankily", "", "", nil)
	if err != nil {
		t.Fatalf("failed to create anchor transaction: %v", err)
	}

	req := httptest.NewRequest("POST", "/admin/transactions/"+anchor.ID.String()+"/add-balance",
		strings.NewReader(`{"amount":250,"notes":"control case, real admin"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200 for a real admin (control case), got %d", resp.StatusCode)
	}

	wallet, err := env.walletRepo.GetByUserID(context.Background(), target.ID)
	if err != nil {
		t.Fatalf("failed to read target wallet: %v", err)
	}
	if !wallet.Balance.Equal(decimal.NewFromInt(250)) {
		t.Fatalf("expected the real admin's credit to succeed (balance 250), got %s", wallet.Balance)
	}
}
