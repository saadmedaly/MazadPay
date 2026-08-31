//go:build integration

// Package integrationtest holds opt-in integration tests that run against a real local
// Postgres + Redis (see backend/docker-compose.yml — mazadpay_postgres on :5433,
// mazadpay_redis on :6380). These never run as part of the fast `go test ./...` suite;
// invoke explicitly with:
//
//	DB_HOST=localhost DB_PORT=5433 DB_USER=mazadpay DB_PASSWORD=mazadpay_secret DB_NAME=mazadpay DB_SSLMODE=disable \
//	REDIS_URL=redis://localhost:6380/0 JWT_SECRET=test-secret-key-for-integration-tests \
//	go test -tags integration ./internal/services/integrationtest/... -v
package integrationtest

import (
	"context"
	"fmt"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/mazadpay/backend/internal/config"
	"github.com/mazadpay/backend/internal/database"
	apperr "github.com/mazadpay/backend/internal/errors"
	"github.com/mazadpay/backend/internal/handlers"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/repository"
	"github.com/mazadpay/backend/internal/services"
	ws "github.com/mazadpay/backend/internal/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// testEnv bundles the real, DB/Redis-backed services under test, wired exactly the way
// cmd/server/main.go / internal/routes/routes.go construct them.
type testEnv struct {
	db     *sqlx.DB
	rdb    *redis.Client
	logger *zap.Logger

	userRepo    repository.UserRepository
	auctionRepo repository.AuctionRepository
	reqRepo     repository.RequestRepository
	walletRepo  repository.WalletRepository
	bidRepo     repository.BidRepository

	authSvc     services.AuthService
	reqSvc      services.RequestService
	auctSvc     services.AuctionService
	bidSvc      services.BidService
	userSvc     services.UserService
	auctHandler *handlers.AuctionHandler
	bidHandler  *handlers.BidHandler
	wsHandler   *handlers.WSHandler
	app         *fiber.App
}

func setupEnv(t *testing.T) *testEnv {
	t.Helper()

	cfg := config.Load()
	if cfg.JWT.Secret == "" {
		cfg.JWT.Secret = "test-secret-key-for-integration-tests"
	}

	logger := zap.NewNop()

	db, err := database.NewPostgres(cfg, logger)
	if err != nil {
		t.Fatalf("failed to connect to postgres: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	rdb, err := database.NewRedis(cfg, logger)
	if err != nil {
		t.Fatalf("failed to connect to redis: %v", err)
	}
	t.Cleanup(func() { rdb.Close() })

	userRepo := repository.NewUserRepository(db)
	auctionRepo := repository.NewAuctionRepository(db)
	reportRepo := repository.NewReportRepository(db)
	contentRepo := repository.NewContentRepository(db)
	reqRepo := repository.NewRequestRepository(db)
	notifRepo := repository.NewNotificationRepository(db)
	auditRepo := repository.NewAuditRepository(db)
	walletRepo := repository.NewWalletRepository(db)
	bidRepo := repository.NewBidRepository(db)
	favoriteRepo := repository.NewFavoriteRepository(db)
	kycRepo := repository.NewKYCRepository(db)

	notifSvc := services.NewNotificationService(notifRepo, userRepo, "", "", logger, nil)
	auditSvc := services.NewAuditService(auditRepo)
	mediaSvc := services.NewMediaService(cfg, logger)

	authSvc := services.NewAuthService(userRepo, cfg.JWT.Secret, cfg.JWT.ExpiryHours, "development", "", nil, 4, cfg.Redis.OTPTTLMinutes, rdb, logger)
	auctSvc := services.NewAuctionService(db, auctionRepo, reportRepo, notifSvc, userRepo, mediaSvc, rdb, walletRepo, auditSvc, logger)
	reqSvc := services.NewRequestService(reqRepo, auctionRepo, contentRepo, userRepo, auditSvc, notifSvc, logger)
	bidSvc := services.NewBidService(db, auctionRepo, bidRepo, walletRepo, userRepo, notifSvc, noopHub{})
	userSvc := services.NewUserService(userRepo, favoriteRepo, auctionRepo, kycRepo, auditSvc, rdb, logger, cfg.JWT.ExpiryHours)

	auctHandler := handlers.NewAuctionHandler(auctSvc, userRepo, logger)
	bidHandler := handlers.NewBidHandler(bidSvc, auctionRepo, userRepo, logger)
	wsHandler := handlers.NewWSHandler(ws.NewHub(logger), authSvc, auctionRepo, userRepo, logger)
	boostSvc := services.NewAuctionBoostService(db)
	boostHandler := handlers.NewAuctionBoostHandler(boostSvc, auctionRepo, userRepo, logger)
	walletSvcForAutoBid := services.NewWalletService(db, walletRepo, repository.NewTransactionRepository(db, walletRepo), nil, nil, logger)
	autoBidSvc := services.NewBidAutoBidService(db, bidSvc, walletSvcForAutoBid)
	autoBidHandler := handlers.NewBidAutoBidHandler(autoBidSvc, auctionRepo, userRepo, logger)

	// Minimal fiber app for HTTP-level tests of handler-layer logic (e.g. GetByID's /
	// History's market-isolation checks, which live in the handler layer, not the
	// service) -- fakeAuth sets c.Locals("user_id") directly instead of running real
	// JWT middleware, since these tests already hold a concrete userID from fixture
	// setup.
	app := fiber.New()
	app.Get("/auctions/:id", fakeAuth(nil), auctHandler.GetByID)
	app.Get("/auctions/:id/as/:userID", fakeAuthFromParam(), auctHandler.GetByID)
	app.Get("/auctions/:id/bids", fakeAuth(nil), bidHandler.History)
	app.Get("/auctions/:id/bids/as/:userID", fakeAuthFromParam(), bidHandler.History)
	app.Get("/auctions/:id/seller-contact/as/:userID", fakeAuthFromParam(), auctHandler.GetSellerContact)
	app.Get("/auctions/:id/boosts/as/:userID", fakeAuthFromParam(), boostHandler.GetAuctionBoosts)
	app.Post("/auctions/:id/boost/as/:userID", fakeAuthFromParam(), boostHandler.CreateBoost)
	app.Post("/auctions/:id/auto-bid/as/:userID", fakeAuthFromParam(), autoBidHandler.CreateAutoBid)

	return &testEnv{
		db: db, rdb: rdb, logger: logger,
		userRepo: userRepo, auctionRepo: auctionRepo, reqRepo: reqRepo, walletRepo: walletRepo, bidRepo: bidRepo,
		authSvc: authSvc, reqSvc: reqSvc, auctSvc: auctSvc, bidSvc: bidSvc, userSvc: userSvc,
		auctHandler: auctHandler, bidHandler: bidHandler, wsHandler: wsHandler, app: app,
	}
}

// fakeAuth is a stand-in for middleware.JWT in tests: sets the caller's user_id local
// directly (test already holds a concrete userID from fixture setup, no need to mint and
// parse a real JWT). Passing nil means "unauthenticated" (anonymous request).
func fakeAuth(userID *uuid.UUID) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if userID != nil {
			c.Locals("user_id", *userID)
		}
		return c.Next()
	}
}

// fakeAuthFromParam reads the caller's user_id from a :userID URL param (used by the
// /auctions/:id/as/:userID test-only route below) so a single route can be exercised as
// different callers without registering a new route per fixture user.
func fakeAuthFromParam() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if raw := c.Params("userID"); raw != "" {
			if uid, err := uuid.Parse(raw); err == nil {
				c.Locals("user_id", uid)
			}
		}
		return c.Next()
	}
}

// noopHub satisfies services.AuctionHub for tests that never need real WebSocket
// broadcasts (bidSvc.PlaceBid calls Broadcast unconditionally on success).
type noopHub struct{}

func (noopHub) Broadcast(auctionID uuid.UUID, event models.WSEvent)                        {}
func (noopHub) BroadcastToUser(auctionID uuid.UUID, userID string, event models.WSEvent) {}

// uniquePhone returns a national number that is genuinely VALID for the given region
// (libphonenumber validates it against that region's numbering plan — see
// services.NormalizeE164), made unique per test run so repeated runs against the same
// long-lived local DB never collide on the phone_e164 unique index.
//
// The random part is deliberately confined to the subscriber-number digits only: the
// region-defining prefix (MR/TN leading digit, NANP area code + exchange code) is fixed
// to a known-valid value, because a randomly-generated prefix would frequently produce a
// number libphonenumber correctly rejects, making tests flaky for reasons unrelated to
// what they're actually testing.
func uniquePhone(region string) string {
	switch region {
	case "MR":
		// MR: 8 digits, must start with 2, 3 or 4. Fix "22" then 6 random digits.
		return "22" + randomDigits(6)
	case "TN":
		// TN: 8 digits, mobile prefixes include 2x/5x/9x. Fix "20" then 6 random digits.
		return "20" + randomDigits(6)
	case "US":
		// NANP: 10 digits = area code (202 = Washington DC) + exchange code (must not
		// start with 0 or 1, so fix "555") + 4 random subscriber digits.
		return "202555" + randomDigits(4)
	case "CA":
		// NANP: 416 = Toronto, same exchange-code rule as above.
		return "416555" + randomDigits(4)
	}
	return randomDigits(8)
}

// randomDigits returns n cryptographically-insignificant but well-distributed decimal
// digits derived from a fresh UUID — enough uniqueness for test fixtures.
func randomDigits(n int) string {
	out := make([]byte, 0, n)
	for len(out) < n {
		for _, c := range uuid.New().String() {
			var d byte
			switch {
			case c >= '0' && c <= '9':
				d = byte(c - '0')
			case c >= 'a' && c <= 'f':
				d = byte(c-'a') + 10
			default:
				continue // skip dashes
			}
			out = append(out, '0'+(d%10))
			if len(out) == n {
				break
			}
		}
	}
	return string(out)
}

// uniqueName suffixes a fixture's full_name with a random token so repeated runs against
// the same long-lived local DB don't find multiple rows for the "same" fixture name.
func uniqueName(base string) string {
	return base + " " + uuid.New().String()[:8]
}

func mkCategoryID() int { return 1 } // "Phones" — confirmed present in the long-lived local DB

func newAuctionRequest(userID uuid.UUID, titleSuffix string) *models.AuctionRequest {
	descAr := "وصف تجريبي للمزاد رقم " + titleSuffix + " يحتوي على أكثر من عشرة أحرف"
	now := time.Now()
	return &models.AuctionRequest{
		ID:            uuid.New(),
		UserID:        userID,
		CategoryID:    mkCategoryID(),
		TitleAr:       "مزاد اختبار " + titleSuffix,
		DescriptionAr: &descAr,
		StartPrice:    decimal.NewFromInt(100),
		MinIncrement:  decimal.NewFromInt(10),
		StartDate:     now.Add(1 * time.Hour),
		EndDate:       now.Add(48 * time.Hour),
		Quantity:      1,
		Status:        "pending",
	}
}

// === (a) Register MR, TN, US, CA users with correct country_iso ===
func TestRegister_MultiCountry(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()

	cases := []struct {
		region string
	}{
		{"MR"}, {"TN"}, {"US"}, {"CA"},
	}

	var createdE164s []string
	for _, tc := range cases {
		phone := uniquePhone(tc.region)
		fullName := uniqueName("TEST INTEGRATION " + tc.region)
		err := env.authSvc.Register(ctx, phone, "StrongPass123", fullName, "", "", tc.region)
		if err != nil {
			t.Fatalf("Register(%s, region=%s) failed: %v", phone, tc.region, err)
		}

		var rows []struct {
			Phone    string `db:"phone"`
			E164     string `db:"phone_e164"`
			ISO      string `db:"phone_country_iso"`
			FullName string `db:"full_name"`
		}
		err = env.db.SelectContext(ctx, &rows, `SELECT phone, phone_e164, phone_country_iso, full_name FROM users WHERE full_name = $1`, fullName)
		if err != nil || len(rows) != 1 {
			t.Fatalf("expected exactly 1 user for region %s, got %d rows, err=%v", tc.region, len(rows), err)
		}
		r := rows[0]
		if r.E164 == "" || r.ISO != tc.region {
			t.Fatalf("region %s: expected phone_e164 populated and phone_country_iso=%s, got e164=%q iso=%q", tc.region, tc.region, r.E164, r.ISO)
		}
		t.Logf("region=%s phone=%s -> stored phone=%s phone_e164=%s phone_country_iso=%s", tc.region, phone, r.Phone, r.E164, r.ISO)
		createdE164s = append(createdE164s, r.E164)
	}

	// Verify distinct users (4 distinct E.164 values, no collisions)
	seen := map[string]bool{}
	for _, e164 := range createdE164s {
		if seen[e164] {
			t.Fatalf("duplicate phone_e164 %q across supposedly distinct users", e164)
		}
		seen[e164] = true
	}
	if len(seen) != 4 {
		t.Fatalf("expected 4 distinct users, got %d", len(seen))
	}
}

// === (b) Reject invalid number / wrong region ===
func TestRegister_RejectsInvalidOrWrongRegion(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()

	countBefore := countUsers(t, env)

	// Too-short number for MR (needs 8 digits)
	err := env.authSvc.Register(ctx, "12345", "StrongPass123", "TEST SHOULD NOT EXIST 1", "", "", "MR")
	if err == nil {
		t.Fatalf("expected error for too-short MR number, got nil")
	}
	t.Logf("too-short number correctly rejected: %v", err)

	// A real US number submitted with country_iso "CA" must be rejected (regression
	// test for the country-mismatch bug).
	usNumber := "2025551234" // valid US NANP number pattern
	err = env.authSvc.Register(ctx, usNumber, "StrongPass123", "TEST SHOULD NOT EXIST 2", "", "", "CA")
	if err == nil {
		t.Fatalf("expected error registering a US number tagged as CA region, got nil")
	}
	t.Logf("US number tagged CA correctly rejected: %v", err)

	// And the mirror case: a real CA number tagged as US.
	caNumber := "4165551234" // valid CA (Toronto) NANP number pattern
	err = env.authSvc.Register(ctx, caNumber, "StrongPass123", "TEST SHOULD NOT EXIST 3", "", "", "US")
	if err == nil {
		t.Fatalf("expected error registering a CA number tagged as US region, got nil")
	}
	t.Logf("CA number tagged US correctly rejected: %v", err)

	countAfter := countUsers(t, env)
	if countAfter != countBefore {
		t.Fatalf("expected no new users created (before=%d after=%d)", countBefore, countAfter)
	}
}

// === (c) Reject duplicate phone across different raw input formats ===
func TestRegister_RejectsDuplicateAcrossFormats(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()

	national := uniquePhone("MR") // e.g. 8-digit MR number, no prefix
	err := env.authSvc.Register(ctx, national, "StrongPass123", uniqueName("TEST DUP FIXTURE 1"), "", "", "MR")
	if err != nil {
		t.Fatalf("first registration failed unexpectedly (phone %s): %v", national, err)
	}

	countAfterFirst := countUsers(t, env)

	// Same logical number, different raw format (with country dial code prefix, spaces).
	withPrefix := "+222 " + national
	err = env.authSvc.Register(ctx, withPrefix, "StrongPass123", uniqueName("TEST DUP FIXTURE 2"), "", "", "MR")
	if err == nil {
		t.Fatalf("expected duplicate-phone error for same number in different format, got nil")
	}
	if err != apperr.ErrDuplicatePhone {
		t.Logf("note: got error %v (not exactly ErrDuplicatePhone sentinel, but still rejected)", err)
	} else {
		t.Logf("correctly rejected as ErrDuplicatePhone: %v", err)
	}

	countAfterSecond := countUsers(t, env)
	if countAfterSecond != countAfterFirst {
		t.Fatalf("expected no second user created for duplicate phone (after first=%d after second attempt=%d)", countAfterFirst, countAfterSecond)
	}
}

// === (d) Legacy MR login before AND after backfill ===
func TestLogin_LegacyUserBeforeAndAfterBackfill(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()

	rawPin := "1234"
	hash, err := bcrypt.GenerateFromPassword([]byte(rawPin), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("bcrypt hash generation failed: %v", err)
	}

	legacyPhone := uniquePhone("MR") // 8-digit MR-format raw phone, as stored pre-migration
	userID := uuid.New()

	_, err = env.db.ExecContext(ctx, `
		INSERT INTO users (id, phone, password_hash, full_name, role, is_active, is_verified, language_pref)
		VALUES ($1, $2, $3, $4, 'user', true, true, 'ar')
	`, userID, legacyPhone, string(hash), uniqueName("TEST BACKFILL FIXTURE"))
	if err != nil {
		t.Fatalf("failed to insert legacy fixture user: %v", err)
	}

	// Login BEFORE backfill: countryISO empty -> falls back to legacy FindByPhone path.
	token, user, err := env.authSvc.Login(ctx, legacyPhone, "", rawPin)
	if err != nil {
		t.Fatalf("login BEFORE backfill failed: %v", err)
	}
	if token == "" || user == nil || user.ID != userID {
		t.Fatalf("login BEFORE backfill returned unexpected result: token=%q user=%v", token, user)
	}
	t.Logf("login BEFORE backfill succeeded for user %s", userID)

	// Backfill this one row directly (mirrors cmd/backfill_phone_e164 logic).
	e164, iso, err := services.NormalizeE164(legacyPhone, "MR")
	if err != nil {
		t.Fatalf("NormalizeE164 failed for legacy fixture phone: %v", err)
	}
	res, err := env.db.ExecContext(ctx, `UPDATE users SET phone_e164 = $1, phone_country_iso = $2 WHERE id = $3 AND phone_e164 IS NULL`, e164, iso, userID)
	if err != nil {
		t.Fatalf("backfill update failed: %v", err)
	}
	if n, _ := res.RowsAffected(); n != 1 {
		t.Fatalf("expected 1 row affected by backfill update, got %d", n)
	}

	// Login AFTER backfill: same raw credentials, still no countryISO from an
	// old/un-updated client — should still work via legacy fallback OR the new E.164
	// path if countryISO were provided. We test both.
	token2, user2, err := env.authSvc.Login(ctx, legacyPhone, "", rawPin)
	if err != nil {
		t.Fatalf("login AFTER backfill (legacy path) failed: %v", err)
	}
	if token2 == "" || user2 == nil || user2.ID != userID {
		t.Fatalf("login AFTER backfill (legacy path) returned unexpected result")
	}
	t.Logf("login AFTER backfill (legacy path, no country_iso) succeeded for user %s", userID)

	token3, user3, err := env.authSvc.Login(ctx, legacyPhone, "MR", rawPin)
	if err != nil {
		t.Fatalf("login AFTER backfill (E.164 path, country_iso=MR) failed: %v", err)
	}
	if token3 == "" || user3 == nil || user3.ID != userID {
		t.Fatalf("login AFTER backfill (E.164 path) returned unexpected result")
	}
	t.Logf("login AFTER backfill (E.164 path, country_iso=MR) succeeded for user %s", userID)
}

// === (g)(h) Create auction request -> pending, not publicly visible ===
func TestAuctionRequest_CreatePendingNotPubliclyVisible(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()

	seller := createTestUser(t, env, "TEST INTEGRATION SELLER G")

	req := newAuctionRequest(seller.ID, "g-create-"+uuid.New().String()[:6])
	if err := env.reqSvc.CreateAuctionRequest(ctx, req); err != nil {
		t.Fatalf("CreateAuctionRequest failed: %v", err)
	}

	var status string
	if err := env.db.GetContext(ctx, &status, `SELECT status FROM auction_requests WHERE id = $1`, req.ID); err != nil {
		t.Fatalf("failed to read back request status: %v", err)
	}
	if status != "pending" {
		t.Fatalf("expected status=pending, got %q", status)
	}
	t.Logf("(g) auction_requests row created with status=%q", status)

	// (h): no auction row should exist yet at all for this request (not approved), so
	// there is nothing to find in the publicly-visible-filtered view.
	var auctionCount int
	if err := env.db.GetContext(ctx, &auctionCount, `SELECT COUNT(*) FROM auctions WHERE seller_id = $1 AND title_ar = $2`, seller.ID, req.TitleAr); err != nil {
		t.Fatalf("count query failed: %v", err)
	}
	if auctionCount != 0 {
		t.Fatalf("(h) expected 0 auctions for unapproved request, found %d", auctionCount)
	}
	t.Logf("(h) confirmed no auction row exists yet for pending (unapproved) request")
}

// === (i) Reject with empty reason fails ===
func TestReviewAuctionRequest_RejectRequiresNotes(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()

	seller := createTestUser(t, env, "TEST INTEGRATION SELLER I")
	admin := createTestAdmin(t, env, "TEST INTEGRATION ADMIN I")

	req := newAuctionRequest(seller.ID, "i-reject-empty-"+uuid.New().String()[:6])
	if err := env.reqSvc.CreateAuctionRequest(ctx, req); err != nil {
		t.Fatalf("CreateAuctionRequest failed: %v", err)
	}

	err := env.reqSvc.ReviewAuctionRequest(ctx, req.ID, "rejected", "", admin.ID)
	if err == nil {
		t.Fatalf("expected error rejecting with empty notes, got nil")
	}
	if err != services.ErrRejectionNotesRequired {
		t.Fatalf("expected ErrRejectionNotesRequired, got %v", err)
	}
	t.Logf("(i) empty-reason rejection correctly returned ErrRejectionNotesRequired: %v", err)

	var status string
	if err := env.db.GetContext(ctx, &status, `SELECT status FROM auction_requests WHERE id = $1`, req.ID); err != nil {
		t.Fatalf("failed to read back status: %v", err)
	}
	if status != "pending" {
		t.Fatalf("expected status to remain pending after failed rejection, got %q", status)
	}
}

// === (j) Edit + resubmit ===
func TestUpdateAuctionRequest_EditAndResubmitAfterRejection(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()

	seller := createTestUser(t, env, "TEST INTEGRATION SELLER J")
	admin := createTestAdmin(t, env, "TEST INTEGRATION ADMIN J")

	req := newAuctionRequest(seller.ID, "j-edit-resubmit-"+uuid.New().String()[:6])
	if err := env.reqSvc.CreateAuctionRequest(ctx, req); err != nil {
		t.Fatalf("CreateAuctionRequest failed: %v", err)
	}

	if err := env.reqSvc.ReviewAuctionRequest(ctx, req.ID, "rejected", "photo required", admin.ID); err != nil {
		t.Fatalf("ReviewAuctionRequest(rejected) failed: %v", err)
	}

	updated := newAuctionRequest(seller.ID, "j-edit-resubmit-corrected")
	updated.Status = "pending"
	if err := env.reqSvc.UpdateAuctionRequest(ctx, req.ID, seller.ID, updated); err != nil {
		t.Fatalf("UpdateAuctionRequest failed: %v", err)
	}

	var row struct {
		Status     string     `db:"status"`
		AdminNotes *string    `db:"admin_notes"`
		ReviewedAt *time.Time `db:"reviewed_at"`
		ReviewedBy *uuid.UUID `db:"reviewed_by"`
	}
	if err := env.db.GetContext(ctx, &row, `SELECT status, admin_notes, reviewed_at, reviewed_by FROM auction_requests WHERE id = $1`, req.ID); err != nil {
		t.Fatalf("failed to read back row: %v", err)
	}
	if row.Status != "pending" {
		t.Fatalf("expected status=pending after resubmit, got %q", row.Status)
	}
	if row.AdminNotes != nil || row.ReviewedAt != nil || row.ReviewedBy != nil {
		t.Fatalf("expected admin_notes/reviewed_at/reviewed_by cleared after resubmit, got notes=%v at=%v by=%v", row.AdminNotes, row.ReviewedAt, row.ReviewedBy)
	}
	t.Logf("(j) resubmit correctly reset status=pending and cleared review fields")
}

// === (k) Approve -> auction visible with description ===
func TestReviewAuctionRequest_ApproveCreatesPublicAuction(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()

	seller := createTestUser(t, env, "TEST INTEGRATION SELLER K")
	admin := createTestAdmin(t, env, "TEST INTEGRATION ADMIN K")

	req := newAuctionRequest(seller.ID, "k-approve-"+uuid.New().String()[:6])
	if err := env.reqSvc.CreateAuctionRequest(ctx, req); err != nil {
		t.Fatalf("CreateAuctionRequest failed: %v", err)
	}

	if err := env.reqSvc.ReviewAuctionRequest(ctx, req.ID, "approved", "looks good", admin.ID); err != nil {
		t.Fatalf("ReviewAuctionRequest(approved) failed: %v", err)
	}

	var auction models.Auction
	if err := env.db.GetContext(ctx, &auction, `SELECT * FROM auctions WHERE seller_id = $1 AND title_ar = $2`, seller.ID, req.TitleAr); err != nil {
		t.Fatalf("expected an auction row to exist after approval: %v", err)
	}
	if auction.DescriptionAr == nil || *auction.DescriptionAr != *req.DescriptionAr {
		t.Fatalf("expected auction description_ar to carry over from request, got %v want %v", auction.DescriptionAr, req.DescriptionAr)
	}
	t.Logf("(k) auction created with status=%q description_ar carried over correctly", auction.Status)

	if !services.PubliclyVisibleAuctionStatuses[auction.Status] {
		t.Fatalf("(k) expected approved auction status %q to be in PubliclyVisibleAuctionStatuses", auction.Status)
	}
	t.Logf("(k) confirmed auction.Status=%q is in the publicly-visible-filtered set", auction.Status)
}

// === (l) User B cannot edit/review User A's request ===
func TestUpdateAuctionRequest_OwnershipEnforced(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()

	userA := createTestUser(t, env, "TEST INTEGRATION USER A L")
	userB := createTestUser(t, env, "TEST INTEGRATION USER B L")
	admin := createTestAdmin(t, env, "TEST INTEGRATION ADMIN L")

	req := newAuctionRequest(userA.ID, "l-ownership-"+uuid.New().String()[:6])
	if err := env.reqSvc.CreateAuctionRequest(ctx, req); err != nil {
		t.Fatalf("CreateAuctionRequest failed: %v", err)
	}
	// Must be rejected first (draft/rejected only editable) to exercise UpdateAuctionRequest's
	// ownership check on a state where it would otherwise succeed.
	if err := env.reqSvc.ReviewAuctionRequest(ctx, req.ID, "rejected", "needs fix", admin.ID); err != nil {
		t.Fatalf("ReviewAuctionRequest(rejected) failed: %v", err)
	}

	attempted := newAuctionRequest(userA.ID, "l-ownership-hijack-attempt")
	err := env.reqSvc.UpdateAuctionRequest(ctx, req.ID, userB.ID, attempted)
	if err == nil {
		t.Fatalf("expected ownership error when userB edits userA's request, got nil")
	}
	if err != services.ErrNotRequestOwner {
		t.Fatalf("expected ErrNotRequestOwner, got %v", err)
	}
	t.Logf("(l) userB correctly blocked from editing userA's request: %v", err)

	var titleAr string
	if err := env.db.GetContext(ctx, &titleAr, `SELECT title_ar FROM auction_requests WHERE id = $1`, req.ID); err != nil {
		t.Fatalf("failed to read back title: %v", err)
	}
	if titleAr != req.TitleAr {
		t.Fatalf("(l) request row was modified despite ownership check failing: got %q want %q", titleAr, req.TitleAr)
	}
	t.Logf("(l) confirmed request row NOT modified by unauthorized update attempt")
}

// === (m) Regular user cannot bypass review via direct auction creation ===
func TestAuctionCreate_RegularUserAlwaysPending(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()

	seller := createTestUser(t, env, "TEST INTEGRATION SELLER M")

	input := services.CreateAuctionInput{
		CategoryID:      mkCategoryID(),
		TitleAr:         "مزاد مباشر اختبار m",
		DescriptionAr:   "وصف تجريبي لمزاد تم إنشاؤه مباشرة بدون مراجعة الإدارة",
		StartPrice:      decimal.NewFromInt(50),
		MinIncrement:    decimal.NewFromInt(5),
		EndTime:         time.Now().Add(24 * time.Hour),
		Quantity:        1,
	}
	// Note: CreateAuctionInput has NO Status field at all — this alone structurally
	// proves a caller (even the non-admin-restricted POST /auctions handler) cannot set
	// an arbitrary status. Confirmed by reading internal/services/auction_service.go
	// CreateAuctionInput struct (fields: CategoryID, SubCategoryID, LocationID, titles,
	// descriptions, prices, times, LotNumber, PhoneContact, ItemDetails, BuyNowPrice,
	// Images, Condition, Brand, VideoURL, Quantity — no Status).

	auction, err := env.auctSvc.Create(ctx, seller.ID, input)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if auction.Status != "pending" {
		t.Fatalf("(m) expected auction.Status=pending regardless of input, got %q", auction.Status)
	}
	t.Logf("(m) direct auction creation by regular user resulted in status=%q (forced, cannot be bypassed)", auction.Status)

	if services.PubliclyVisibleAuctionStatuses[auction.Status] {
		t.Fatalf("(m) pending auction must NOT be publicly visible, but PubliclyVisibleAuctionStatuses says it is")
	}
	t.Logf("(m) confirmed pending auction does not appear in publicly-visible-filtered set")
}

// === (n) Admin CAN create and publish directly from dashboard ===
func TestAdminCreateAndApprove_ResultsInPublicAuction(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()

	admin := createTestAdmin(t, env, "TEST INTEGRATION ADMIN N")

	req := newAuctionRequest(admin.ID, "n-admin-publish-"+uuid.New().String()[:6])
	if err := env.reqSvc.CreateAuctionRequest(ctx, req); err != nil {
		t.Fatalf("admin-originated CreateAuctionRequest failed: %v", err)
	}

	if err := env.reqSvc.ReviewAuctionRequest(ctx, req.ID, "approved", "self-approved by admin", admin.ID); err != nil {
		t.Fatalf("ReviewAuctionRequest(approved) by admin failed: %v", err)
	}

	var auction models.Auction
	if err := env.db.GetContext(ctx, &auction, `SELECT * FROM auctions WHERE seller_id = $1 AND title_ar = $2`, admin.ID, req.TitleAr); err != nil {
		t.Fatalf("expected a live auction after admin approval: %v", err)
	}
	if !services.PubliclyVisibleAuctionStatuses[auction.Status] {
		t.Fatalf("(n) expected admin-published auction status %q to be publicly visible", auction.Status)
	}
	t.Logf("(n) admin create+approve flow resulted in publicly-visible auction, status=%q", auction.Status)
}

// === Country-scoped currency (migration 000046, Phase 1) ===
//
// Business rule under test throughout this section (per explicit product decision):
// market identity = account_country_iso equality. NEVER currency equality alone --
// SN and CI both use XOF but are separate markets (see case (E) below).

// (A) MR request creation -> correct market/currency stamped server-side.
func TestAuctionRequest_MarketCurrency_MR(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST CURRENCY SELLER MR") // MR via createTestUser

	req := newAuctionRequest(seller.ID, "currency-mr-"+uuid.New().String()[:6])
	if err := env.reqSvc.CreateAuctionRequest(ctx, req); err != nil {
		t.Fatalf("CreateAuctionRequest failed: %v", err)
	}
	if req.MarketCountryISO == nil || *req.MarketCountryISO != "MR" {
		t.Fatalf("expected market_country_iso=MR, got %v", req.MarketCountryISO)
	}
	wantCurrency := currencyOf(t, env, "MR")
	if req.CurrencyCode == nil || *req.CurrencyCode != wantCurrency {
		t.Fatalf("expected currency_code=%s, got %v", wantCurrency, req.CurrencyCode)
	}
	if wantCurrency != "MRU" {
		t.Fatalf("sanity check failed: countries.currency_code for MR must be MRU (not the stale CLDR MRO), got %s", wantCurrency)
	}
	t.Logf("(A) MR request stamped market=%s currency=%s", *req.MarketCountryISO, *req.CurrencyCode)
}

// (B) TN request creation -> correct market/currency (distinct from MR).
func TestAuctionRequest_MarketCurrency_TN(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUserWithCountry(t, env, "TEST CURRENCY SELLER TN", "TN")

	req := newAuctionRequest(seller.ID, "currency-tn-"+uuid.New().String()[:6])
	if err := env.reqSvc.CreateAuctionRequest(ctx, req); err != nil {
		t.Fatalf("CreateAuctionRequest failed: %v", err)
	}
	if req.MarketCountryISO == nil || *req.MarketCountryISO != "TN" {
		t.Fatalf("expected market_country_iso=TN, got %v", req.MarketCountryISO)
	}
	wantCurrency := currencyOf(t, env, "TN")
	if req.CurrencyCode == nil || *req.CurrencyCode != wantCurrency {
		t.Fatalf("expected currency_code=%s, got %v", wantCurrency, req.CurrencyCode)
	}
	if wantCurrency != "TND" {
		t.Fatalf("sanity check failed: countries.currency_code for TN must be TND, got %s", wantCurrency)
	}
	t.Logf("(B) TN request stamped market=%s currency=%s", *req.MarketCountryISO, *req.CurrencyCode)
}

// (C) Client attempting to spoof market_country_iso/currency_code in the request body
// must be ignored -- the service always overwrites with the server-derived values.
func TestAuctionRequest_MarketCurrency_ClientSpoofIgnored(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST CURRENCY SELLER SPOOF") // MR

	spoofedMarket := "FR"
	spoofedCurrency := "EUR"
	req := newAuctionRequest(seller.ID, "currency-spoof-"+uuid.New().String()[:6])
	req.MarketCountryISO = &spoofedMarket
	req.CurrencyCode = &spoofedCurrency

	if err := env.reqSvc.CreateAuctionRequest(ctx, req); err != nil {
		t.Fatalf("CreateAuctionRequest failed: %v", err)
	}
	if req.MarketCountryISO == nil || *req.MarketCountryISO != "MR" {
		t.Fatalf("client-supplied market_country_iso=FR was NOT overridden -- got %v, want MR (security regression)", req.MarketCountryISO)
	}
	if req.CurrencyCode == nil || *req.CurrencyCode == "EUR" {
		t.Fatalf("client-supplied currency_code=EUR was NOT overridden -- got %v (security regression)", req.CurrencyCode)
	}

	var row struct {
		MarketCountryISO *string `db:"market_country_iso"`
		CurrencyCode     *string `db:"currency_code"`
	}
	if err := env.db.GetContext(ctx, &row, `SELECT market_country_iso, currency_code FROM auction_requests WHERE id = $1`, req.ID); err != nil {
		t.Fatalf("failed to read back row: %v", err)
	}
	if row.MarketCountryISO == nil || *row.MarketCountryISO != "MR" {
		t.Fatalf("persisted market_country_iso is spoofed value, got %v", row.MarketCountryISO)
	}
	t.Logf("(C) client-supplied market=FR/currency=EUR correctly discarded, persisted market=%s currency=%s", *row.MarketCountryISO, *row.CurrencyCode)
}

// (D) MR bidder on MR auction: allowed.
func TestPlaceBid_SameMarket_Allowed(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST BID SELLER D")
	bidder := createTestUser(t, env, "TEST BID BIDDER D")
	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")
	creditWallet(t, env, bidder.ID, decimal.NewFromInt(1000))

	bid, err := env.bidSvc.PlaceBid(ctx, auction.ID, bidder.ID, decimal.NewFromInt(110))
	if err != nil {
		t.Fatalf("(D) expected same-market (MR->MR) bid to succeed, got error: %v", err)
	}
	t.Logf("(D) MR bidder -> MR auction bid succeeded: %s", bid.ID)
}

// (E) TN bidder on MR auction: rejected (cross-market, even though both may resolve to
// distinct currencies here -- the primary case for currency-sharing markets is (F)/SN-CI).
func TestPlaceBid_CrossMarket_Rejected(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST BID SELLER E")
	bidder := createTestUserWithCountry(t, env, "TEST BID BIDDER E", "TN")
	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")
	creditWallet(t, env, bidder.ID, decimal.NewFromInt(1000))

	_, err := env.bidSvc.PlaceBid(ctx, auction.ID, bidder.ID, decimal.NewFromInt(110))
	if err == nil {
		t.Fatalf("(E) expected TN bidder on MR auction to be rejected, got nil error")
	}
	if err != apperr.ErrCrossMarketBid {
		t.Fatalf("(E) expected ErrCrossMarketBid, got %v", err)
	}
	t.Logf("(E) TN bidder -> MR auction correctly rejected: %v", err)
}

// (F) SN bidder on CI auction: rejected DESPITE both markets sharing the same currency
// (XOF) -- this is the security/financially-critical case explicitly flagged: market
// identity must be decided by COUNTRY, never by currency equality.
func TestPlaceBid_SharedCurrencyDifferentMarket_Rejected(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()

	snCurrency := currencyOf(t, env, "SN")
	ciCurrency := currencyOf(t, env, "CI")
	if snCurrency != ciCurrency {
		t.Fatalf("test precondition failed: SN and CI must share a currency (both XOF) for this test to be meaningful, got SN=%s CI=%s", snCurrency, ciCurrency)
	}

	seller := createTestUserWithCountry(t, env, "TEST BID SELLER F CI", "CI")
	bidder := createTestUserWithCountry(t, env, "TEST BID BIDDER F SN", "SN")
	auction := createTestAuction(t, env, seller.ID, "CI", ciCurrency)
	creditWallet(t, env, bidder.ID, decimal.NewFromInt(1000))

	_, err := env.bidSvc.PlaceBid(ctx, auction.ID, bidder.ID, decimal.NewFromInt(110))
	if err == nil {
		t.Fatalf("(F) CRITICAL: expected SN bidder on CI auction to be rejected despite shared currency %s, got nil error", ciCurrency)
	}
	if err != apperr.ErrCrossMarketBid {
		t.Fatalf("(F) expected ErrCrossMarketBid, got %v", err)
	}
	t.Logf("(F) SN bidder -> CI auction correctly rejected despite shared currency %s: %v", ciCurrency, err)
}

// (G) Legacy user row (account_country_iso IS NULL, predating migration 000046)
// bidding on a legacy auction (market_country_iso IS NULL) -- both fall back to
// DefaultAccountCountryISO ('MR') and must be treated as the same market.
func TestPlaceBid_LegacyNullFallback_TreatedAsMR(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()

	seller := createTestUser(t, env, "TEST BID SELLER G LEGACY")
	bidder := createTestUser(t, env, "TEST BID BIDDER G LEGACY")

	// Force both rows' new columns back to NULL to simulate pre-migration data.
	if _, err := env.db.ExecContext(ctx, `UPDATE users SET account_country_iso = NULL WHERE id = $1`, bidder.ID); err != nil {
		t.Fatalf("failed to null out bidder account_country_iso: %v", err)
	}

	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")
	if _, err := env.db.ExecContext(ctx, `UPDATE auctions SET market_country_iso = NULL, currency_code = NULL WHERE id = $1`, auction.ID); err != nil {
		t.Fatalf("failed to null out auction market_country_iso/currency_code: %v", err)
	}

	creditWallet(t, env, bidder.ID, decimal.NewFromInt(1000))
	// Also null out the wallet's currency_code to simulate a pre-migration wallet.
	if _, err := env.db.ExecContext(ctx, `UPDATE wallets SET currency_code = NULL WHERE user_id = $1`, bidder.ID); err != nil {
		t.Fatalf("failed to null out wallet currency_code: %v", err)
	}

	_, err := env.bidSvc.PlaceBid(ctx, auction.ID, bidder.ID, decimal.NewFromInt(110))
	if err != nil {
		t.Fatalf("(G) expected legacy NULL bidder/auction (both falling back to MR/MRU) to succeed, got: %v", err)
	}
	t.Logf("(G) legacy NULL user + NULL auction correctly treated as same MR/MRU market")
}

// (H) Wallet/auction currency mismatch is rejected even within a nominally allowed
// same-market bid (defense-in-depth check in PlaceBid, see bid_service.go).
func TestPlaceBid_WalletCurrencyMismatch_Rejected(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST BID SELLER H")
	bidder := createTestUser(t, env, "TEST BID BIDDER H") // MR account, wallet will be MRU
	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")
	creditWallet(t, env, bidder.ID, decimal.NewFromInt(1000))

	// Corrupt the wallet's currency_code directly to simulate an inconsistent state that
	// should never occur in practice but must still be caught (financial-safety
	// requirement, see bid_service.go comment on this exact check).
	if _, err := env.db.ExecContext(ctx, `UPDATE wallets SET currency_code = 'EUR' WHERE user_id = $1`, bidder.ID); err != nil {
		t.Fatalf("failed to corrupt wallet currency_code: %v", err)
	}

	_, err := env.bidSvc.PlaceBid(ctx, auction.ID, bidder.ID, decimal.NewFromInt(110))
	if err == nil {
		t.Fatalf("(H) expected wallet/auction currency mismatch to be rejected, got nil error")
	}
	if err != apperr.ErrWalletCurrencyMismatch {
		t.Fatalf("(H) expected ErrWalletCurrencyMismatch, got %v", err)
	}
	t.Logf("(H) wallet currency EUR vs auction currency MRU correctly rejected: %v", err)
}

// (I) Approving a request preserves/stamps the same market_country_iso/currency_code on
// the resulting auction -- never re-derived dynamically from the seller's current account.
func TestReviewAuctionRequest_ApprovePreservesMarketCurrency(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUserWithCountry(t, env, "TEST APPROVE MARKET SELLER I", "TN")
	admin := createTestAdmin(t, env, "TEST APPROVE MARKET ADMIN I")

	req := newAuctionRequest(seller.ID, "approve-market-"+uuid.New().String()[:6])
	if err := env.reqSvc.CreateAuctionRequest(ctx, req); err != nil {
		t.Fatalf("CreateAuctionRequest failed: %v", err)
	}
	if req.MarketCountryISO == nil || *req.MarketCountryISO != "TN" {
		t.Fatalf("precondition failed: request market_country_iso should be TN, got %v", req.MarketCountryISO)
	}

	// Simulate the seller's account market changing AFTER submission but BEFORE
	// approval -- the auction must still carry the ORIGINAL request market, not the
	// seller's now-current one.
	if _, err := env.db.ExecContext(ctx, `UPDATE users SET account_country_iso = 'MA' WHERE id = $1`, seller.ID); err != nil {
		t.Fatalf("failed to simulate seller account market change: %v", err)
	}

	if err := env.reqSvc.ReviewAuctionRequest(ctx, req.ID, "approved", "ok", admin.ID); err != nil {
		t.Fatalf("ReviewAuctionRequest(approved) failed: %v", err)
	}

	var auction models.Auction
	if err := env.db.GetContext(ctx, &auction, `SELECT * FROM auctions WHERE seller_id = $1 AND title_ar = $2`, seller.ID, req.TitleAr); err != nil {
		t.Fatalf("expected an auction row to exist after approval: %v", err)
	}
	if auction.MarketCountryISO == nil || *auction.MarketCountryISO != "TN" {
		t.Fatalf("(I) expected auction market_country_iso=TN (preserved from request at submission time), got %v (seller's account market was changed to MA after submission)", auction.MarketCountryISO)
	}
	t.Logf("(I) auction correctly preserved original request market=%s despite seller's account market later changing to MA", *auction.MarketCountryISO)
}

// === Phase 1.1 blocker fixes ===

// (J) TN wallet transaction (deposit) is stamped with the TN account's currency (TND),
// not left NULL and not defaulted to MRU.
func TestInitiateDeposit_TNWallet_StampsTND(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	walletSvc := newWalletSvc(env)

	user := createTestUserWithCountry(t, env, "TEST DEPOSIT TN J", "TN")

	txn, err := walletSvc.InitiateDeposit(ctx, user.ID, decimal.NewFromInt(100), "bank_transfer", "bank_transfer", "")
	if err != nil {
		t.Fatalf("InitiateDeposit failed: %v", err)
	}
	if txn.CurrencyCode == nil || *txn.CurrencyCode != "TND" {
		t.Fatalf("(J) expected deposit transaction currency_code=TND, got %v", txn.CurrencyCode)
	}

	var stored string
	if err := env.db.GetContext(ctx, &stored, `SELECT currency_code FROM transactions WHERE id = $1`, txn.ID); err != nil {
		t.Fatalf("failed to read back transaction currency_code: %v", err)
	}
	if stored != "TND" {
		t.Fatalf("(J) persisted transactions.currency_code=%q, want TND", stored)
	}
	t.Logf("(J) TN deposit correctly stamped currency_code=TND")
}

// (K) MR wallet transaction (withdrawal) is stamped with MRU.
func TestRequestWithdraw_MRWallet_StampsMRU(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	walletSvc := newWalletSvc(env)

	user := createTestUser(t, env, "TEST WITHDRAW MR K") // MR
	creditWallet(t, env, user.ID, decimal.NewFromInt(500))

	txn, err := walletSvc.RequestWithdraw(ctx, user.ID, decimal.NewFromInt(100), "bank_transfer")
	if err != nil {
		t.Fatalf("RequestWithdraw failed: %v", err)
	}
	if txn.CurrencyCode == nil || *txn.CurrencyCode != "MRU" {
		t.Fatalf("(K) expected withdraw transaction currency_code=MRU, got %v", txn.CurrencyCode)
	}
	t.Logf("(K) MR withdrawal correctly stamped currency_code=MRU")
}

// (L) Client cannot spoof transaction currency: InitiateDeposit/RequestWithdraw take no
// currency parameter at all in their service signatures -- structurally impossible to
// pass one in. This test proves the currency stamped always matches the wallet's own
// currency regardless of the raw amount/gateway/payment-method strings supplied,
// confirming there is no code path treating any caller-supplied string as a currency.
func TestTransactionCurrency_ClientCannotSpoof(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	walletSvc := newWalletSvc(env)

	user := createTestUserWithCountry(t, env, "TEST SPOOF CURRENCY L", "MA")

	// Even a gateway/payment_method string that looks like a currency code must have no
	// effect on the stamped currency_code -- these fields are never interpreted as such.
	txn, err := walletSvc.InitiateDeposit(ctx, user.ID, decimal.NewFromInt(50), "EUR", "USD", "")
	if err != nil {
		t.Fatalf("InitiateDeposit failed: %v", err)
	}
	if txn.CurrencyCode == nil || *txn.CurrencyCode != "MAD" {
		t.Fatalf("(L) expected currency_code derived from account market (MAD) regardless of gateway/payment_method strings, got %v", txn.CurrencyCode)
	}
	t.Logf("(L) gateway=EUR/payment_method=USD had no effect; correctly stamped MAD from account market")
}

// (M) Wallet/transaction monetary context cannot silently diverge: a wallet's stamped
// currency_code never changes after creation even if the owner's account_country_iso is
// later modified -- a transaction created afterward still uses the wallet's original,
// immutable currency (matching wallet_repo.go's write-once design), not the new market.
func TestWalletTransactionCurrency_CannotDiverge(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	walletSvc := newWalletSvc(env)

	user := createTestUser(t, env, "TEST DIVERGE M") // MR -> wallet currency MRU
	// Force wallet creation now, while account market is still MR.
	if _, err := env.walletRepo.GetByUserID(ctx, user.ID); err != nil {
		t.Fatalf("failed to ensure wallet exists: %v", err)
	}

	// Simulate the account market changing after the wallet already exists.
	if _, err := env.db.ExecContext(ctx, `UPDATE users SET account_country_iso = 'TN' WHERE id = $1`, user.ID); err != nil {
		t.Fatalf("failed to simulate account market change: %v", err)
	}

	var walletCurrency string
	if err := env.db.GetContext(ctx, &walletCurrency, `SELECT currency_code FROM wallets WHERE user_id = $1`, user.ID); err != nil {
		t.Fatalf("failed to read wallet currency: %v", err)
	}
	if walletCurrency != "MRU" {
		t.Fatalf("(M) expected wallet currency_code to remain MRU (immutable, write-once at creation) despite account market changing to TN, got %s", walletCurrency)
	}

	creditWallet(t, env, user.ID, decimal.NewFromInt(200))
	txn, err := walletSvc.RequestWithdraw(ctx, user.ID, decimal.NewFromInt(50), "bank_transfer")
	if err != nil {
		t.Fatalf("RequestWithdraw failed: %v", err)
	}
	if txn.CurrencyCode == nil || *txn.CurrencyCode != "MRU" {
		t.Fatalf("(M) expected new transaction to use the wallet's original immutable currency MRU (not the new account market TN's TND), got %v", txn.CurrencyCode)
	}
	t.Logf("(M) wallet currency stayed MRU after account market changed to TN; new transaction correctly used MRU, not TND")
}

// (N) Legacy NULL historical transaction (predating migration 000046) remains safely
// readable via EffectiveCurrencyCode(), falling back to MRU.
func TestTransaction_LegacyNullCurrency_ReadsSafely(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()

	user := createTestUser(t, env, "TEST LEGACY TX N")
	txID := uuid.New()
	if _, err := env.db.ExecContext(ctx, `
		INSERT INTO transactions (id, user_id, type, amount, status, currency_code)
		VALUES ($1, $2, 'deposit', 100, 'pending', NULL)
	`, txID, user.ID); err != nil {
		t.Fatalf("failed to insert legacy NULL-currency transaction fixture: %v", err)
	}

	var tx models.Transaction
	if err := env.db.GetContext(ctx, &tx, `SELECT * FROM transactions WHERE id = $1`, txID); err != nil {
		t.Fatalf("failed to read back legacy transaction: %v", err)
	}
	if tx.CurrencyCode != nil {
		t.Fatalf("(N) precondition failed: expected raw CurrencyCode nil for legacy fixture, got %v", tx.CurrencyCode)
	}
	if got := tx.EffectiveCurrencyCode(); got != "MRU" {
		t.Fatalf("(N) expected EffectiveCurrencyCode() fallback to MRU for legacy NULL transaction, got %q", got)
	}
	t.Logf("(N) legacy NULL-currency transaction reads safely via EffectiveCurrencyCode() -> MRU")
}

// (O) MR user -> MR auction detail succeeds (HTTP 200).
func TestAuctionDetail_SameMarket_Succeeds(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST DETAIL SELLER O")
	viewer := createTestUser(t, env, "TEST DETAIL VIEWER O") // MR
	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")
	if _, err := env.db.ExecContext(ctx, `UPDATE auctions SET status = 'active' WHERE id = $1`, auction.ID); err != nil {
		t.Fatalf("failed to activate fixture auction: %v", err)
	}

	status := httpGetAuctionDetail(t, env, auction.ID, &viewer.ID)
	if status != 200 {
		t.Fatalf("(O) expected 200 for MR viewer -> MR auction, got %d", status)
	}
	t.Logf("(O) MR viewer -> MR auction detail: %d", status)
}

// (P) TN user -> MR auction detail denied (404, not a disclosure).
func TestAuctionDetail_CrossMarket_Denied(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST DETAIL SELLER P")
	viewer := createTestUserWithCountry(t, env, "TEST DETAIL VIEWER P", "TN")
	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")
	if _, err := env.db.ExecContext(ctx, `UPDATE auctions SET status = 'active' WHERE id = $1`, auction.ID); err != nil {
		t.Fatalf("failed to activate fixture auction: %v", err)
	}

	status := httpGetAuctionDetail(t, env, auction.ID, &viewer.ID)
	if status != 404 {
		t.Fatalf("(P) CRITICAL: expected 404 for TN viewer -> MR auction (market isolation bypass by ID), got %d", status)
	}
	t.Logf("(P) TN viewer -> MR auction detail correctly denied: %d", status)
}

// (Q) Anonymous -> MR auction succeeds (anonymous effective market = MR).
func TestAuctionDetail_Anonymous_MRSucceeds(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST DETAIL SELLER Q")
	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")
	if _, err := env.db.ExecContext(ctx, `UPDATE auctions SET status = 'active' WHERE id = $1`, auction.ID); err != nil {
		t.Fatalf("failed to activate fixture auction: %v", err)
	}

	status := httpGetAuctionDetail(t, env, auction.ID, nil)
	if status != 200 {
		t.Fatalf("(Q) expected 200 for anonymous -> MR auction, got %d", status)
	}
	t.Logf("(Q) anonymous -> MR auction detail: %d", status)
}

// (R) Anonymous -> TN auction denied (anonymous effective market = MR, not TN).
func TestAuctionDetail_Anonymous_TNDenied(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUserWithCountry(t, env, "TEST DETAIL SELLER R", "TN")
	auction := createTestAuction(t, env, seller.ID, "TN", "TND")
	if _, err := env.db.ExecContext(ctx, `UPDATE auctions SET status = 'active' WHERE id = $1`, auction.ID); err != nil {
		t.Fatalf("failed to activate fixture auction: %v", err)
	}

	status := httpGetAuctionDetail(t, env, auction.ID, nil)
	if status != 404 {
		t.Fatalf("(R) expected 404 for anonymous -> TN auction (old-client-compatible MR-only default), got %d", status)
	}
	t.Logf("(R) anonymous -> TN auction detail correctly denied: %d", status)
}

// (S) User cannot add a cross-market auction to favorites.
func TestAddFavorite_CrossMarket_Denied(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST FAVORITE SELLER S")
	user := createTestUserWithCountry(t, env, "TEST FAVORITE USER S", "TN")
	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")

	err := env.userSvc.AddFavorite(ctx, user.ID, auction.ID)
	if err == nil {
		t.Fatalf("(S) expected error adding cross-market (TN user, MR auction) favorite, got nil")
	}
	if err != apperr.ErrNotFound {
		t.Fatalf("(S) expected ErrNotFound, got %v", err)
	}

	var count int
	if err := env.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM user_favorites WHERE user_id = $1 AND auction_id = $2`, user.ID, auction.ID); err != nil {
		t.Fatalf("count query failed: %v", err)
	}
	if count != 0 {
		t.Fatalf("(S) expected no favorite row to be created, found %d", count)
	}
	t.Logf("(S) cross-market favorite correctly rejected: %v", err)
}

// (T) Favorites listing never returns an auction outside the caller's effective market,
// even for a stale row created before the AddFavorite guard existed (defense-in-depth).
func TestListFavorites_ExcludesCrossMarket(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	sellerMR := createTestUser(t, env, "TEST FAVLIST SELLER MR T")
	sellerTN := createTestUserWithCountry(t, env, "TEST FAVLIST SELLER TN T", "TN")
	user := createTestUser(t, env, "TEST FAVLIST USER T") // MR

	mrAuction := createTestAuction(t, env, sellerMR.ID, "MR", "MRU")
	tnAuction := createTestAuction(t, env, sellerTN.ID, "TN", "TND")

	if err := env.userSvc.AddFavorite(ctx, user.ID, mrAuction.ID); err != nil {
		t.Fatalf("AddFavorite(same-market) failed: %v", err)
	}
	// Bypass the AddFavorite guard directly at the repo level to simulate a stale
	// cross-market row predating this fix.
	if _, err := env.db.ExecContext(ctx, `INSERT INTO user_favorites (user_id, auction_id) VALUES ($1, $2)`, user.ID, tnAuction.ID); err != nil {
		t.Fatalf("failed to insert stale cross-market favorite fixture: %v", err)
	}

	favorites, err := env.userSvc.ListFavorites(ctx, user.ID)
	if err != nil {
		t.Fatalf("ListFavorites failed: %v", err)
	}
	for _, a := range favorites {
		if a.ID == tnAuction.ID {
			t.Fatalf("(T) CRITICAL: cross-market TN auction leaked into MR user's favorites list")
		}
	}
	found := false
	for _, a := range favorites {
		if a.ID == mrAuction.ID {
			found = true
		}
	}
	if !found {
		t.Fatalf("(T) expected same-market MR favorite to still be present, got %d favorites", len(favorites))
	}
	t.Logf("(T) favorites list correctly excluded stale cross-market TN favorite, kept MR favorite (%d total)", len(favorites))
}

// === Phase 1.2 final isolation fixes ===

// (bid-history A) MR authenticated viewer -> MR auction bids succeeds.
func TestBidHistory_SameMarket_Succeeds(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST BIDHIST SELLER A")
	viewer := createTestUser(t, env, "TEST BIDHIST VIEWER A") // MR
	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")
	if _, err := env.db.ExecContext(ctx, `UPDATE auctions SET status = 'active' WHERE id = $1`, auction.ID); err != nil {
		t.Fatalf("failed to activate fixture auction: %v", err)
	}

	status := httpGetBidHistory(t, env, auction.ID, &viewer.ID)
	if status != 200 {
		t.Fatalf("(bid-history A) expected 200 for MR viewer -> MR auction bids, got %d", status)
	}
	t.Logf("(bid-history A) MR viewer -> MR auction bids: %d", status)
}

// (bid-history B) TN authenticated viewer -> MR auction bids denied.
func TestBidHistory_CrossMarket_Denied(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST BIDHIST SELLER B")
	viewer := createTestUserWithCountry(t, env, "TEST BIDHIST VIEWER B", "TN")
	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")
	if _, err := env.db.ExecContext(ctx, `UPDATE auctions SET status = 'active' WHERE id = $1`, auction.ID); err != nil {
		t.Fatalf("failed to activate fixture auction: %v", err)
	}

	status := httpGetBidHistory(t, env, auction.ID, &viewer.ID)
	if status != 404 {
		t.Fatalf("(bid-history B) CRITICAL: expected 404 for TN viewer -> MR auction bids (market isolation bypass by ID), got %d", status)
	}
	t.Logf("(bid-history B) TN viewer -> MR auction bids correctly denied: %d", status)
}

// (bid-history C) Anonymous -> MR auction bids succeeds (anonymous effective market = MR).
func TestBidHistory_Anonymous_MRSucceeds(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST BIDHIST SELLER C")
	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")
	if _, err := env.db.ExecContext(ctx, `UPDATE auctions SET status = 'active' WHERE id = $1`, auction.ID); err != nil {
		t.Fatalf("failed to activate fixture auction: %v", err)
	}

	status := httpGetBidHistory(t, env, auction.ID, nil)
	if status != 200 {
		t.Fatalf("(bid-history C) expected 200 for anonymous -> MR auction bids, got %d", status)
	}
	t.Logf("(bid-history C) anonymous -> MR auction bids: %d", status)
}

// (bid-history D) Anonymous -> TN auction bids denied.
func TestBidHistory_Anonymous_TNDenied(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUserWithCountry(t, env, "TEST BIDHIST SELLER D", "TN")
	auction := createTestAuction(t, env, seller.ID, "TN", "TND")
	if _, err := env.db.ExecContext(ctx, `UPDATE auctions SET status = 'active' WHERE id = $1`, auction.ID); err != nil {
		t.Fatalf("failed to activate fixture auction: %v", err)
	}

	status := httpGetBidHistory(t, env, auction.ID, nil)
	if status != 404 {
		t.Fatalf("(bid-history D) expected 404 for anonymous -> TN auction bids, got %d", status)
	}
	t.Logf("(bid-history D) anonymous -> TN auction bids correctly denied: %d", status)
}

// (bid-history E) Knowing the auction ID alone is insufficient to bypass market
// isolation: a TN viewer who knows a real, active MR auction's ID still gets 404, and
// the response carries no distinguishing "cross market" detail (same 404 shape as an
// unknown/non-existent ID).
func TestBidHistory_IDKnowledgeInsufficientToBypass(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST BIDHIST SELLER E")
	viewer := createTestUserWithCountry(t, env, "TEST BIDHIST VIEWER E", "TN")
	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")
	if _, err := env.db.ExecContext(ctx, `UPDATE auctions SET status = 'active' WHERE id = $1`, auction.ID); err != nil {
		t.Fatalf("failed to activate fixture auction: %v", err)
	}

	crossMarketStatus := httpGetBidHistory(t, env, auction.ID, &viewer.ID)
	nonExistentStatus := httpGetBidHistory(t, env, uuid.New(), &viewer.ID)
	if crossMarketStatus != nonExistentStatus {
		t.Fatalf("(bid-history E) cross-market response (%d) distinguishable from non-existent-ID response (%d) -- disclosure risk", crossMarketStatus, nonExistentStatus)
	}
	if crossMarketStatus != 404 {
		t.Fatalf("(bid-history E) expected both to be 404, got %d", crossMarketStatus)
	}
	t.Logf("(bid-history E) real cross-market auction ID and a random non-existent ID both correctly return %d, indistinguishable", crossMarketStatus)
}

// (ws-1) WebSocket subscription: MR user -> MR auction authorized.
func TestWSAuthorizeSubscription_SameMarket_Allowed(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST WS SELLER 1")
	viewer := createTestUser(t, env, "TEST WS VIEWER 1") // MR
	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")

	if ok := env.wsHandler.AuthorizeSubscription(ctx, auction.ID, viewer.ID.String()); !ok {
		t.Fatalf("(ws-1) expected MR viewer -> MR auction WS subscription to be authorized")
	}
	t.Logf("(ws-1) MR -> MR WebSocket subscription correctly authorized")
}

// (ws-2) WebSocket subscription: TN user -> MR auction rejected.
func TestWSAuthorizeSubscription_CrossMarket_Rejected(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST WS SELLER 2")
	viewer := createTestUserWithCountry(t, env, "TEST WS VIEWER 2", "TN")
	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")

	if ok := env.wsHandler.AuthorizeSubscription(ctx, auction.ID, viewer.ID.String()); ok {
		t.Fatalf("(ws-2) expected TN viewer -> MR auction WS subscription to be rejected")
	}
	t.Logf("(ws-2) TN -> MR WebSocket subscription correctly rejected")
}

// (ws-3) SECURITY-CRITICAL: SN user -> CI auction rejected DESPITE both using XOF --
// WebSocket market isolation must be decided by country equality alone, exactly like
// PlaceBid's cross-market check, never by currency equality.
func TestWSAuthorizeSubscription_SharedCurrencyDifferentMarket_Rejected(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()

	snCurrency := currencyOf(t, env, "SN")
	ciCurrency := currencyOf(t, env, "CI")
	if snCurrency != ciCurrency {
		t.Fatalf("test precondition failed: SN and CI must share a currency for this test to be meaningful, got SN=%s CI=%s", snCurrency, ciCurrency)
	}

	seller := createTestUserWithCountry(t, env, "TEST WS SELLER 3 CI", "CI")
	subscriber := createTestUserWithCountry(t, env, "TEST WS SUBSCRIBER 3 SN", "SN")
	auction := createTestAuction(t, env, seller.ID, "CI", ciCurrency)

	if ok := env.wsHandler.AuthorizeSubscription(ctx, auction.ID, subscriber.ID.String()); ok {
		t.Fatalf("(ws-3) CRITICAL: expected SN subscriber -> CI auction WS subscription to be rejected despite shared currency %s", ciCurrency)
	}
	t.Logf("(ws-3) SN -> CI WebSocket subscription correctly rejected despite shared currency %s", ciCurrency)
}

// === Phase 1.3 final two-endpoint fixes ===
// Both /seller-contact and /boosts already require JWT auth (jwtMiddleware in
// routes.go) -- no anonymous-access branch exists or is required for either.

// (seller-contact 1) MR user -> MR auction seller contact succeeds.
func TestSellerContact_SameMarket_Succeeds(t *testing.T) {
	env := setupEnv(t)
	seller := createTestUser(t, env, "TEST SELLERCONTACT SELLER 1")
	viewer := createTestUser(t, env, "TEST SELLERCONTACT VIEWER 1") // MR
	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")

	status, body := httpGetSellerContact(t, env, auction.ID, viewer.ID)
	if status != 200 {
		t.Fatalf("(seller-contact 1) expected 200 for MR viewer -> MR auction, got %d", status)
	}
	if !strings.Contains(body, "phone") {
		t.Fatalf("(seller-contact 1) expected a phone field in the successful response, got %q", body)
	}
	t.Logf("(seller-contact 1) MR -> MR seller-contact: %d", status)
}

// (seller-contact 2) TN user -> MR auction seller contact denied, AND no seller
// contact payload (not even masked) leaks in the denial response.
func TestSellerContact_CrossMarket_DeniedNoLeak(t *testing.T) {
	env := setupEnv(t)
	seller := createTestUser(t, env, "TEST SELLERCONTACT SELLER 2")
	viewer := createTestUserWithCountry(t, env, "TEST SELLERCONTACT VIEWER 2", "TN")
	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")
	if err := setAuctionPhoneContact(t, env, auction.ID, "22345678"); err != nil {
		t.Fatalf("failed to set fixture phone_contact: %v", err)
	}

	status, body := httpGetSellerContact(t, env, auction.ID, viewer.ID)
	if status != 404 {
		t.Fatalf("(seller-contact 2) CRITICAL: expected 404 for TN viewer -> MR auction seller-contact, got %d", status)
	}
	if strings.Contains(body, "phone") || strings.Contains(body, "####") {
		t.Fatalf("(seller-contact 2) CRITICAL: denial response leaked phone/contact data: %q", body)
	}
	t.Logf("(seller-contact 2) TN -> MR seller-contact correctly denied with no data leak: %d %q", status, body)
}

// (seller-contact 3) Knowing the auction ID is insufficient to bypass isolation: the
// cross-market response is indistinguishable from a non-existent-ID response.
func TestSellerContact_IDKnowledgeInsufficientToBypass(t *testing.T) {
	env := setupEnv(t)
	seller := createTestUser(t, env, "TEST SELLERCONTACT SELLER 3")
	viewer := createTestUserWithCountry(t, env, "TEST SELLERCONTACT VIEWER 3", "TN")
	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")

	crossMarketStatus, _ := httpGetSellerContact(t, env, auction.ID, viewer.ID)
	nonExistentStatus, _ := httpGetSellerContact(t, env, uuid.New(), viewer.ID)
	if crossMarketStatus != nonExistentStatus || crossMarketStatus != 404 {
		t.Fatalf("(seller-contact 3) expected both cross-market (%d) and non-existent (%d) to be indistinguishable 404s", crossMarketStatus, nonExistentStatus)
	}
	t.Logf("(seller-contact 3) real cross-market ID and random ID both correctly return %d, indistinguishable", crossMarketStatus)
}

// (boosts 1) MR user -> MR auction boosts succeeds.
func TestAuctionBoosts_SameMarket_Succeeds(t *testing.T) {
	env := setupEnv(t)
	seller := createTestUser(t, env, "TEST BOOSTS SELLER 1")
	viewer := createTestUser(t, env, "TEST BOOSTS VIEWER 1") // MR
	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")

	status := httpGetAuctionBoosts(t, env, auction.ID, viewer.ID)
	if status != 200 {
		t.Fatalf("(boosts 1) expected 200 for MR viewer -> MR auction boosts, got %d", status)
	}
	t.Logf("(boosts 1) MR -> MR boosts: %d", status)
}

// (boosts 2) TN user -> MR auction boosts denied.
func TestAuctionBoosts_CrossMarket_Denied(t *testing.T) {
	env := setupEnv(t)
	seller := createTestUser(t, env, "TEST BOOSTS SELLER 2")
	viewer := createTestUserWithCountry(t, env, "TEST BOOSTS VIEWER 2", "TN")
	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")

	status := httpGetAuctionBoosts(t, env, auction.ID, viewer.ID)
	if status != 404 {
		t.Fatalf("(boosts 2) CRITICAL: expected 404 for TN viewer -> MR auction boosts, got %d", status)
	}
	t.Logf("(boosts 2) TN -> MR boosts correctly denied: %d", status)
}

// === Phase 1.4 final write isolation ===

// (write-1) TN user cannot create a boost for an MR auction.
func TestCreateBoost_CrossMarket_Denied(t *testing.T) {
	env := setupEnv(t)
	seller := createTestUser(t, env, "TEST CREATEBOOST SELLER 1")
	buyer := createTestUserWithCountry(t, env, "TEST CREATEBOOST BUYER 1", "TN")
	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")

	status := httpCreateBoost(t, env, auction.ID, buyer.ID)
	if status != 404 {
		t.Fatalf("(write-1) CRITICAL: expected 404 for TN user creating a boost on an MR auction, got %d", status)
	}

	count := countBoostsForAuction(t, env, auction.ID)
	if count != 0 {
		t.Fatalf("(write-1) expected no boost row to be created, found %d", count)
	}
	t.Logf("(write-1) TN -> MR CreateBoost correctly denied, no row created: %d", status)
}

// (write-2) MR user CAN create a boost for an MR auction (same-market action still works).
func TestCreateBoost_SameMarket_Succeeds(t *testing.T) {
	env := setupEnv(t)
	seller := createTestUser(t, env, "TEST CREATEBOOST SELLER 2")
	buyer := createTestUser(t, env, "TEST CREATEBOOST BUYER 2") // MR
	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")

	status := httpCreateBoost(t, env, auction.ID, buyer.ID)
	if status != 200 {
		t.Fatalf("(write-2) expected 200 for MR user creating a boost on an MR auction, got %d", status)
	}

	count := countBoostsForAuction(t, env, auction.ID)
	if count != 1 {
		t.Fatalf("(write-2) expected exactly 1 boost row to be created, found %d", count)
	}
	t.Logf("(write-2) MR -> MR CreateBoost correctly succeeded, 1 row created: %d", status)
}

// (write-3) TN user cannot create an auto-bid for an MR auction.
func TestCreateAutoBid_CrossMarket_Denied(t *testing.T) {
	env := setupEnv(t)
	seller := createTestUser(t, env, "TEST CREATEAUTOBID SELLER 3")
	buyer := createTestUserWithCountry(t, env, "TEST CREATEAUTOBID BUYER 3", "TN")
	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")

	status := httpCreateAutoBid(t, env, auction.ID, buyer.ID)
	if status != 404 {
		t.Fatalf("(write-3) CRITICAL: expected 404 for TN user creating an auto-bid on an MR auction, got %d", status)
	}

	count := countAutoBidsForAuction(t, env, auction.ID)
	if count != 0 {
		t.Fatalf("(write-3) expected no auto-bid row to be created, found %d", count)
	}
	t.Logf("(write-3) TN -> MR CreateAutoBid correctly denied, no row created: %d", status)
}

// (write-4) SN user cannot create an auto-bid for a CI auction, despite both using XOF --
// security-critical case, country equality only, never currency.
func TestCreateAutoBid_SharedCurrencyDifferentMarket_Denied(t *testing.T) {
	env := setupEnv(t)
	snCurrency := currencyOf(t, env, "SN")
	ciCurrency := currencyOf(t, env, "CI")
	if snCurrency != ciCurrency {
		t.Fatalf("test precondition failed: SN and CI must share a currency, got SN=%s CI=%s", snCurrency, ciCurrency)
	}

	seller := createTestUserWithCountry(t, env, "TEST CREATEAUTOBID SELLER 4 CI", "CI")
	buyer := createTestUserWithCountry(t, env, "TEST CREATEAUTOBID BUYER 4 SN", "SN")
	auction := createTestAuction(t, env, seller.ID, "CI", ciCurrency)

	status := httpCreateAutoBid(t, env, auction.ID, buyer.ID)
	if status != 404 {
		t.Fatalf("(write-4) CRITICAL: expected 404 for SN user creating an auto-bid on a CI auction despite shared currency %s, got %d", ciCurrency, status)
	}
	t.Logf("(write-4) SN -> CI CreateAutoBid correctly denied despite shared currency %s: %d", ciCurrency, status)
}

// (write-5) MR user CAN create an auto-bid for an MR auction (same-market action still works).
func TestCreateAutoBid_SameMarket_Succeeds(t *testing.T) {
	env := setupEnv(t)
	seller := createTestUser(t, env, "TEST CREATEAUTOBID SELLER 5")
	buyer := createTestUser(t, env, "TEST CREATEAUTOBID BUYER 5") // MR
	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")

	status := httpCreateAutoBid(t, env, auction.ID, buyer.ID)
	if status != 200 {
		t.Fatalf("(write-5) expected 200 for MR user creating an auto-bid on an MR auction, got %d", status)
	}

	count := countAutoBidsForAuction(t, env, auction.ID)
	if count != 1 {
		t.Fatalf("(write-5) expected exactly 1 auto-bid row to be created, found %d", count)
	}
	t.Logf("(write-5) MR -> MR CreateAutoBid correctly succeeded, 1 row created: %d", status)
}

// --- Phase 1.4 helpers ---

// httpCreateBoost performs a real HTTP-level POST against env.app's
// POST /auctions/:id/boost/as/:userID test route (backed by the real
// boostHandler.CreateBoost), with a minimal valid boost payload.
func httpCreateBoost(t *testing.T, env *testEnv, auctionID, callerID uuid.UUID) int {
	t.Helper()
	path := fmt.Sprintf("/auctions/%s/boost/as/%s", auctionID, callerID)
	body := fmt.Sprintf(`{"boost_type":"featured","start_at":%q,"end_at":%q}`,
		time.Now().Format(time.RFC3339), time.Now().Add(24*time.Hour).Format(time.RFC3339))
	req := httptest.NewRequest("POST", path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := env.app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	defer resp.Body.Close()
	return resp.StatusCode
}

// httpCreateAutoBid performs a real HTTP-level POST against env.app's
// POST /auctions/:id/auto-bid/as/:userID test route (backed by the real
// autoBidHandler.CreateAutoBid), with a minimal valid payload.
func httpCreateAutoBid(t *testing.T, env *testEnv, auctionID, callerID uuid.UUID) int {
	t.Helper()
	path := fmt.Sprintf("/auctions/%s/auto-bid/as/%s", auctionID, callerID)
	body := `{"max_amount":500}`
	req := httptest.NewRequest("POST", path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := env.app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	defer resp.Body.Close()
	return resp.StatusCode
}

func countBoostsForAuction(t *testing.T, env *testEnv, auctionID uuid.UUID) int {
	t.Helper()
	var n int
	if err := env.db.GetContext(context.Background(), &n, `SELECT COUNT(*) FROM auction_boosts WHERE auction_id = $1`, auctionID); err != nil {
		t.Fatalf("failed to count boosts: %v", err)
	}
	return n
}

func countAutoBidsForAuction(t *testing.T, env *testEnv, auctionID uuid.UUID) int {
	t.Helper()
	var n int
	if err := env.db.GetContext(context.Background(), &n, `SELECT COUNT(*) FROM bid_auto_bids WHERE auction_id = $1`, auctionID); err != nil {
		t.Fatalf("failed to count auto-bids: %v", err)
	}
	return n
}

// --- Phase 1.3 helpers ---

func setAuctionPhoneContact(t *testing.T, env *testEnv, auctionID uuid.UUID, phone string) error {
	t.Helper()
	_, err := env.db.ExecContext(context.Background(), `UPDATE auctions SET phone_contact = $1 WHERE id = $2`, phone, auctionID)
	return err
}

// httpGetSellerContact performs a real HTTP-level request against env.app's
// GET /auctions/:id/seller-contact/as/:userID test route (backed by the real
// auctHandler.GetSellerContact). Returns the status code and raw response body (to
// allow asserting the denial response contains no seller-contact data).
func httpGetSellerContact(t *testing.T, env *testEnv, auctionID, callerID uuid.UUID) (int, string) {
	t.Helper()
	path := fmt.Sprintf("/auctions/%s/seller-contact/as/%s", auctionID, callerID)
	req := httptest.NewRequest("GET", path, nil)
	resp, err := env.app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(bodyBytes)
}

// httpGetAuctionBoosts performs a real HTTP-level request against env.app's
// GET /auctions/:id/boosts/as/:userID test route (backed by the real
// boostHandler.GetAuctionBoosts).
func httpGetAuctionBoosts(t *testing.T, env *testEnv, auctionID, callerID uuid.UUID) int {
	t.Helper()
	path := fmt.Sprintf("/auctions/%s/boosts/as/%s", auctionID, callerID)
	req := httptest.NewRequest("GET", path, nil)
	resp, err := env.app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	defer resp.Body.Close()
	return resp.StatusCode
}

// --- Phase 1.2 helpers ---

// httpGetBidHistory performs a real HTTP-level request against env.app's
// GET /auctions/:id/bids route (backed by the real bidHandler.History, exercising the
// market-isolation check added in Phase 1.2). callerID nil means anonymous.
func httpGetBidHistory(t *testing.T, env *testEnv, auctionID uuid.UUID, callerID *uuid.UUID) int {
	t.Helper()
	var path string
	if callerID != nil {
		path = fmt.Sprintf("/auctions/%s/bids/as/%s", auctionID, callerID)
	} else {
		path = fmt.Sprintf("/auctions/%s/bids", auctionID)
	}
	req := httptest.NewRequest("GET", path, nil)
	resp, err := env.app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	defer resp.Body.Close()
	return resp.StatusCode
}

// --- Phase 1.1 helpers ---

func newWalletSvc(env *testEnv) services.WalletService {
	return services.NewWalletService(env.db, env.walletRepo, repository.NewTransactionRepository(env.db, env.walletRepo), nil, nil, env.logger)
}

// httpGetAuctionDetail performs a real HTTP-level request against env.app's
// GET /auctions/:id route (backed by the real auctHandler.GetByID, exercising the
// market-isolation check that lives in the handler layer, not the service). callerID
// nil means an anonymous (unauthenticated) request.
func httpGetAuctionDetail(t *testing.T, env *testEnv, auctionID uuid.UUID, callerID *uuid.UUID) int {
	t.Helper()
	var path string
	if callerID != nil {
		path = fmt.Sprintf("/auctions/%s/as/%s", auctionID, callerID)
	} else {
		path = fmt.Sprintf("/auctions/%s", auctionID)
	}
	req := httptest.NewRequest("GET", path, nil)
	resp, err := env.app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	defer resp.Body.Close()
	return resp.StatusCode
}

// --- helpers ---

func countUsers(t *testing.T, env *testEnv) int {
	t.Helper()
	var n int
	if err := env.db.Get(&n, `SELECT COUNT(*) FROM users`); err != nil {
		t.Fatalf("failed to count users: %v", err)
	}
	return n
}

func createTestUser(t *testing.T, env *testEnv, fullName string) *models.User {
	t.Helper()
	ctx := context.Background()
	phone := uniquePhone("MR")
	// Suffix the name so repeated runs against this long-lived local DB don't collide.
	unique := uniqueName(fullName)
	if err := env.authSvc.Register(ctx, phone, "StrongPass123", unique, "", "", "MR"); err != nil {
		t.Fatalf("failed to create fixture user %q (phone %s): %v", unique, phone, err)
	}
	var u models.User
	if err := env.db.Get(&u, `SELECT * FROM users WHERE full_name = $1`, unique); err != nil {
		t.Fatalf("failed to read back fixture user %q: %v", unique, err)
	}
	return &u
}

func createTestAdmin(t *testing.T, env *testEnv, fullName string) *models.User {
	t.Helper()
	ctx := context.Background()
	u := createTestUser(t, env, fullName)
	if _, err := env.db.ExecContext(ctx, `UPDATE users SET role = 'admin' WHERE id = $1`, u.ID); err != nil {
		t.Fatalf("failed to promote fixture user to admin: %v", err)
	}
	u.Role = "admin"
	return u
}

// createTestUserWithCountry inserts a fixture user directly (bypassing Register/
// NormalizeE164's phone-validity requirements) with an explicit account_country_iso --
// used for country-scoped-market tests (migration 000046) covering markets whose real
// dialing plans aren't otherwise exercised by uniquePhone (e.g. SN, CI).
func createTestUserWithCountry(t *testing.T, env *testEnv, fullName, countryISO string) *models.User {
	t.Helper()
	ctx := context.Background()
	hash, err := bcrypt.GenerateFromPassword([]byte("StrongPass123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("bcrypt hash generation failed: %v", err)
	}
	unique := uniqueName(fullName)
	id := uuid.New()
	phone := "TESTFIXTURE" + id.String()[:8]
	_, err = env.db.ExecContext(ctx, `
		INSERT INTO users (id, phone, password_hash, full_name, role, is_active, is_verified, language_pref, account_country_iso)
		VALUES ($1, $2, $3, $4, 'user', true, true, 'ar', $5)
	`, id, phone, string(hash), unique, countryISO)
	if err != nil {
		t.Fatalf("failed to insert fixture user with country %s: %v", countryISO, err)
	}
	var u models.User
	if err := env.db.Get(&u, `SELECT * FROM users WHERE id = $1`, id); err != nil {
		t.Fatalf("failed to read back fixture user: %v", err)
	}
	return &u
}

// currencyOf looks up countries.currency_code for a given ISO-2 code, failing the test
// if not found -- avoids hardcoding currency assumptions duplicated from the migration.
func currencyOf(t *testing.T, env *testEnv, countryISO string) string {
	t.Helper()
	var code string
	if err := env.db.Get(&code, `SELECT currency_code FROM countries WHERE code = $1`, countryISO); err != nil {
		t.Fatalf("failed to look up currency_code for country %s: %v", countryISO, err)
	}
	return code
}

// createTestAuction inserts a minimal active auction directly, with an explicit
// market_country_iso/currency_code and insurance_amount, for bidding tests that need a
// pre-existing auction rather than going through the request->approve flow.
func createTestAuction(t *testing.T, env *testEnv, sellerID uuid.UUID, marketISO, currencyCode string) *models.Auction {
	t.Helper()
	ctx := context.Background()
	// LotNumber must be unique per row (auctions.lot_number has a unique constraint) --
	// left unset this defaulted to "" for every fixture auction, which worked only by
	// accident while a given test run created at most one such fixture; Phase 1.3 added
	// enough createTestAuction call sites in a single `go test` invocation to collide.
	lotNumber := "TEST-" + uuid.New().String()[:8]
	a := &models.Auction{
		ID:               uuid.New(),
		SellerID:         sellerID,
		CategoryID:       mkCategoryID(),
		TitleAr:          "مزاد اختبار سوق " + marketISO + " " + uuid.New().String()[:6],
		LotNumber:        &lotNumber,
		StartPrice:       decimal.NewFromInt(100),
		CurrentPrice:     decimal.NewFromInt(100),
		MinIncrement:     decimal.NewFromInt(10),
		InsuranceAmount:  decimal.NewFromInt(20),
		ReservePrice:     decimal.NewFromInt(100),
		StartTime:        time.Now().Add(-1 * time.Hour),
		EndTime:          time.Now().Add(48 * time.Hour),
		Status:           "active",
		MarketCountryISO: &marketISO,
		CurrencyCode:     &currencyCode,
	}
	if err := env.auctionRepo.Create(ctx, nil, a); err != nil {
		t.Fatalf("failed to create fixture auction: %v", err)
	}
	if err := env.db.GetContext(ctx, a, `SELECT * FROM auctions WHERE id = $1`, a.ID); err != nil {
		t.Fatalf("failed to read back fixture auction: %v", err)
	}
	return a
}

// creditWallet gives a user's auto-created wallet enough balance to cover a bid's
// insurance hold, bypassing the deposit-review flow (not under test here).
func creditWallet(t *testing.T, env *testEnv, userID uuid.UUID, amount decimal.Decimal) {
	t.Helper()
	ctx := context.Background()
	// Ensure the wallet row exists (and gets its currency_code stamped) first.
	if _, err := env.walletRepo.GetByUserID(ctx, userID); err != nil {
		t.Fatalf("failed to ensure wallet exists for user %s: %v", userID, err)
	}
	if _, err := env.db.ExecContext(ctx, `UPDATE wallets SET balance = balance + $1 WHERE user_id = $2`, amount, userID); err != nil {
		t.Fatalf("failed to credit wallet for user %s: %v", userID, err)
	}
}

var _ = fmt.Sprintf // keep fmt import if unused paths change
