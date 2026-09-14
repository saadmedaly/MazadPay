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
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"strings"
	"sync"
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

	authSvc services.AuthService
	reqSvc  services.RequestService
	auctSvc services.AuctionService
	bidSvc  services.BidService
	userSvc services.UserService
	// notifSvc (client feedback #16): the real, DB-backed NotificationService
	// -- used to prove SendBroadcast against a live Postgres, not the fake
	// in-memory repo in notification_broadcast_test.go.
	notifSvc    services.NotificationService
	auctHandler *handlers.AuctionHandler
	bidHandler  *handlers.BidHandler
	wsHandler   *handlers.WSHandler
	userHandler *handlers.UserHandler
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
	auctSvc := services.NewAuctionService(db, auctionRepo, bidRepo, reportRepo, notifSvc, userRepo, mediaSvc, rdb, walletRepo, auditSvc, logger, nil)
	reqSvc := services.NewRequestService(reqRepo, auctionRepo, contentRepo, userRepo, auditSvc, notifSvc, logger, nil)
	bidSvc := services.NewBidService(db, auctionRepo, bidRepo, walletRepo, userRepo, notifSvc, noopHub{})
	userSvc := services.NewUserService(userRepo, favoriteRepo, auctionRepo, kycRepo, auditSvc, rdb, logger, cfg.JWT.ExpiryHours)

	auctHandler := handlers.NewAuctionHandler(auctSvc, userRepo, logger)
	bidHandler := handlers.NewBidHandler(bidSvc, auctionRepo, userRepo, logger)
	wsHandler := handlers.NewWSHandler(ws.NewHub(logger), authSvc, auctionRepo, userRepo, logger)
	boostSvc := services.NewAuctionBoostService(db)
	boostHandler := handlers.NewAuctionBoostHandler(boostSvc, auctionRepo, userRepo, logger)
	walletSvcForAutoBid := services.NewWalletService(db, walletRepo, repository.NewTransactionRepository(db, walletRepo), nil, nil, nil, nil, logger)
	autoBidSvc := services.NewBidAutoBidService(db, bidSvc, walletSvcForAutoBid)
	autoBidHandler := handlers.NewBidAutoBidHandler(autoBidSvc, auctionRepo, userRepo, logger)
	userHandler := handlers.NewUserHandler(userSvc, logger)

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
	// Bug I fix verification: exercises UserHandler.ListFavorites at the full
	// HTTP-handler level (not just favoriteRepo.ListByUserID) so the test
	// proves the actual "images" response-field transformation the mobile
	// app depends on, not only that the DB query returns image_urls.
	app.Get("/favorites/as/:userID", fakeAuthFromParam(), userHandler.ListFavorites)
	// Bug L diagnosis: exercises UserHandler.MyWinnings at the full HTTP
	// handler level (real JSON marshaling of []models.Auction through OK()),
	// not just userSvc.ListMyWinnings directly -- proves the actual wire
	// response contract the mobile app parses, not only that the DB query
	// finds the right rows.
	app.Get("/users/me/winnings/as/:userID", fakeAuthFromParam(), userHandler.MyWinnings)

	return &testEnv{
		db: db, rdb: rdb, logger: logger,
		userRepo: userRepo, auctionRepo: auctionRepo, reqRepo: reqRepo, walletRepo: walletRepo, bidRepo: bidRepo,
		authSvc: authSvc, reqSvc: reqSvc, auctSvc: auctSvc, bidSvc: bidSvc, userSvc: userSvc, notifSvc: notifSvc,
		auctHandler: auctHandler, bidHandler: bidHandler, wsHandler: wsHandler, userHandler: userHandler, app: app,
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

func (noopHub) Broadcast(auctionID uuid.UUID, event models.WSEvent)                      {}
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

// newBannerRequest (client feedback #10, bulk review notifications): unlike
// auction requests, CreateBannerRequest has no insurance-style approval gate
// -- its only validation is EndsAt.After(StartsAt) -- so this fixture needs
// no equivalent to newAuctionRequest's insurance-setup dance before approval.
func newBannerRequest(userID uuid.UUID, titleSuffix string) *models.BannerRequest {
	now := time.Now()
	return &models.BannerRequest{
		ID:       uuid.New(),
		UserID:   userID,
		TitleAr:  "بانر اختبار " + titleSuffix,
		ImageURL: "https://example.com/banner-" + titleSuffix + ".jpg",
		StartsAt: now.Add(1 * time.Hour),
		EndsAt:   now.Add(48 * time.Hour),
		Status:   "pending",
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

	// CreateAuctionRequest unconditionally forces InsuranceAmount=0 /
	// InsurancePolicy="required" (client feedback A7 -- the user must never
	// set insurance). ReviewAuctionRequest's approval guard blocks approving
	// any "required" request with a non-positive InsuranceAmount
	// (apperr.ErrRequestInsuranceNotSet), exactly as the real admin panel's
	// approve action would be blocked. The only legitimate way to clear this
	// gate is the same one the real admin workflow uses: AdminUpdateAuctionRequest
	// with a positive InsuranceAmount, via a full-object update (its handler,
	// request_handler.go AdminUpdateAuctionRequest, parses the ENTIRE request
	// body into models.AuctionRequest and applyAuctionRequestUpdates copies
	// every editable field from it onto the existing row) -- so `updates` here
	// must carry req's own current field values, not just InsuranceAmount, or
	// this call would silently blank out the request's title/prices/dates/etc.
	insuranceUpdates := *req
	insuranceUpdates.InsuranceAmount = decimal.NewFromInt(500)
	if err := env.reqSvc.AdminUpdateAuctionRequest(ctx, req.ID, &insuranceUpdates, nil); err != nil {
		t.Fatalf("AdminUpdateAuctionRequest (setting insurance) failed: %v", err)
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

// TestBulkReviewAuctionRequests_ApproveCreatesRealAuctions_RetryIsIdempotent
// (client feedback #10, bulk review notifications): BulkReviewAuctionRequests
// now delegates each id to ReviewAuctionRequest (see request_service.go) --
// this proves that delegation end-to-end against a real DB: two pending
// requests, both given valid insurance via the same real admin workflow as
// the single-approve test above, bulk-approved together, each produces its
// own real Auction row correctly linked to its own request (by seller_id +
// title_ar, the same identification the single-approve test uses), and a
// SECOND bulk-approve call over the same two ids (simulating a retry/double
// click) creates NO additional Auction rows -- proving the pending-status
// guard inside ReviewAuctionRequest makes the bulk path idempotent.
//
// Notification dispatch itself (SendLocalizedPush) is NOT independently
// re-verified here -- it is exactly the same call ReviewAuctionRequest
// already makes for single review (unchanged by this feature), and is
// covered at the unit level by request_service_bulk_review_test.go's
// pending-status-gate and duplicate-id-deduplication tests. Adding
// notification-delivery assertions here would require either a live FCM
// service (out of scope) or new mock/spy infrastructure on NotificationService
// specifically for this test, which is more infrastructure than this
// feature's scope justifies -- DB/entity-creation correctness is
// integration-tested here; the notification call-path itself is unit-tested.
func TestBulkReviewAuctionRequests_ApproveCreatesRealAuctions_RetryIsIdempotent(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()

	seller := createTestUser(t, env, "TEST BULK APPROVE SELLER N")
	admin := createTestAdmin(t, env, "TEST BULK APPROVE ADMIN N")

	req1 := newAuctionRequest(seller.ID, "n-bulk1-"+uuid.New().String()[:6])
	if err := env.reqSvc.CreateAuctionRequest(ctx, req1); err != nil {
		t.Fatalf("CreateAuctionRequest (req1) failed: %v", err)
	}
	req2 := newAuctionRequest(seller.ID, "n-bulk2-"+uuid.New().String()[:6])
	if err := env.reqSvc.CreateAuctionRequest(ctx, req2); err != nil {
		t.Fatalf("CreateAuctionRequest (req2) failed: %v", err)
	}

	for _, r := range []*models.AuctionRequest{req1, req2} {
		insuranceUpdates := *r
		insuranceUpdates.InsuranceAmount = decimal.NewFromInt(500)
		if err := env.reqSvc.AdminUpdateAuctionRequest(ctx, r.ID, &insuranceUpdates, nil); err != nil {
			t.Fatalf("AdminUpdateAuctionRequest (setting insurance for %s) failed: %v", r.ID, err)
		}
	}

	ids := []uuid.UUID{req1.ID, req2.ID}
	if err := env.reqSvc.BulkReviewAuctionRequests(ctx, ids, "approved", "bulk approved", admin.ID); err != nil {
		t.Fatalf("BulkReviewAuctionRequests(approved) failed: %v", err)
	}

	for _, r := range []*models.AuctionRequest{req1, req2} {
		var status string
		if err := env.db.GetContext(ctx, &status, `SELECT status FROM auction_requests WHERE id = $1`, r.ID); err != nil {
			t.Fatalf("failed to read back request status for %s: %v", r.ID, err)
		}
		if status != "approved" {
			t.Fatalf("(n) expected request %s status=approved, got %q", r.ID, status)
		}

		var auction models.Auction
		if err := env.db.GetContext(ctx, &auction, `SELECT * FROM auctions WHERE seller_id = $1 AND title_ar = $2`, seller.ID, r.TitleAr); err != nil {
			t.Fatalf("(n) expected a real Auction row for bulk-approved request %s (title=%q): %v", r.ID, r.TitleAr, err)
		}
		t.Logf("(n) request %s correctly produced its own Auction row (id=%s, status=%q)", r.ID, auction.ID, auction.Status)
	}

	var auctionCountBefore int
	if err := env.db.GetContext(ctx, &auctionCountBefore, `SELECT COUNT(*) FROM auctions WHERE seller_id = $1`, seller.ID); err != nil {
		t.Fatalf("failed to count auctions before retry: %v", err)
	}
	if auctionCountBefore != 2 {
		t.Fatalf("(n) expected exactly 2 auctions after the first bulk approve, got %d", auctionCountBefore)
	}

	// Retry the same bulk approve over the same (now-approved) ids -- must be
	// a no-op: ReviewAuctionRequest's pending-status guard rejects each id
	// (already approved), BulkReviewAuctionRequests logs and skips rather
	// than erroring the whole batch, and NO new Auction rows are created.
	if err := env.reqSvc.BulkReviewAuctionRequests(ctx, ids, "approved", "bulk approved again", admin.ID); err != nil {
		t.Fatalf("(n) retry BulkReviewAuctionRequests returned an unexpected error (should silently skip already-reviewed ids): %v", err)
	}

	var auctionCountAfter int
	if err := env.db.GetContext(ctx, &auctionCountAfter, `SELECT COUNT(*) FROM auctions WHERE seller_id = $1`, seller.ID); err != nil {
		t.Fatalf("failed to count auctions after retry: %v", err)
	}
	if auctionCountAfter != 2 {
		t.Fatalf("(n) SECURITY/DATA REGRESSION: retrying the same bulk approve created additional Auction rows -- expected 2, got %d", auctionCountAfter)
	}
	t.Logf("(n) confirmed retry created zero additional Auction rows (idempotent): %d auctions before and after retry", auctionCountAfter)
}

// TestBulkReviewBannerRequests_ApproveCreatesRealBanner_RetryIsIdempotent
// (client feedback #10, bulk review notifications): same proof as
// TestBulkReviewAuctionRequests_ApproveCreatesRealAuctions_RetryIsIdempotent
// above, for banner requests. BulkReviewBannerRequests delegates each id to
// ReviewBannerRequest, which gained the pending-status guard this round
// (request_service.go) -- this is what makes the retry idempotent, since
// ReviewBannerRequest previously had no such guard at all. Practical to add:
// banner requests need no insurance-equivalent setup step, so this fixture
// is simpler than the auction one. Same scope note as above: notification
// dispatch itself is unit-tested (request_service_bulk_review_test.go), not
// re-verified here -- this proves real DB/entity-creation behavior only.
func TestBulkReviewBannerRequests_ApproveCreatesRealBanner_RetryIsIdempotent(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()

	seller := createTestUser(t, env, "TEST BULK BANNER SELLER O")
	admin := createTestAdmin(t, env, "TEST BULK BANNER ADMIN O")

	req := newBannerRequest(seller.ID, "o-bulk-"+uuid.New().String()[:6])
	if err := env.reqSvc.CreateBannerRequest(ctx, req); err != nil {
		t.Fatalf("CreateBannerRequest failed: %v", err)
	}

	ids := []uuid.UUID{req.ID}
	if err := env.reqSvc.BulkReviewBannerRequests(ctx, ids, "approved", "bulk approved", admin.ID); err != nil {
		t.Fatalf("BulkReviewBannerRequests(approved) failed: %v", err)
	}

	var status string
	if err := env.db.GetContext(ctx, &status, `SELECT status FROM banner_requests WHERE id = $1`, req.ID); err != nil {
		t.Fatalf("failed to read back banner request status: %v", err)
	}
	if status != "approved" {
		t.Fatalf("(o) expected banner request status=approved, got %q", status)
	}

	var bannerCountBefore int
	if err := env.db.GetContext(ctx, &bannerCountBefore, `SELECT COUNT(*) FROM banners WHERE title_ar = $1`, req.TitleAr); err != nil {
		t.Fatalf("failed to count banners before retry: %v", err)
	}
	if bannerCountBefore != 1 {
		t.Fatalf("(o) expected exactly 1 real Banner row after bulk approve, got %d", bannerCountBefore)
	}
	t.Logf("(o) bulk-approved banner request correctly produced its own Banner row")

	// Retry the same bulk approve over the same (now-approved) id -- must be a
	// no-op: ReviewBannerRequest's pending-status guard (added this round)
	// rejects the id (already approved), BulkReviewBannerRequests logs and
	// skips rather than erroring, and NO new Banner row is created.
	if err := env.reqSvc.BulkReviewBannerRequests(ctx, ids, "approved", "bulk approved again", admin.ID); err != nil {
		t.Fatalf("(o) retry BulkReviewBannerRequests returned an unexpected error (should silently skip already-reviewed ids): %v", err)
	}

	var bannerCountAfter int
	if err := env.db.GetContext(ctx, &bannerCountAfter, `SELECT COUNT(*) FROM banners WHERE title_ar = $1`, req.TitleAr); err != nil {
		t.Fatalf("failed to count banners after retry: %v", err)
	}
	if bannerCountAfter != 1 {
		t.Fatalf("(o) SECURITY/DATA REGRESSION: retrying the same bulk approve created additional Banner rows -- expected 1, got %d", bannerCountAfter)
	}
	t.Logf("(o) confirmed retry created zero additional Banner rows (idempotent)")
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
		CategoryID:    mkCategoryID(),
		TitleAr:       "مزاد مباشر اختبار m",
		DescriptionAr: "وصف تجريبي لمزاد تم إنشاؤه مباشرة بدون مراجعة الإدارة",
		StartPrice:    decimal.NewFromInt(50),
		MinIncrement:  decimal.NewFromInt(5),
		EndTime:       time.Now().Add(24 * time.Hour),
		Quantity:      1,
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

// ==================================================
// Client feedback #19: one successful bid per user per auction, ever --
// enforced via the auction_bid_participants guard table (migration 000050),
// claimed in the SAME transaction as the bid itself. See bid_service.go
// PlaceBid step 2c and bid_repo.go ClaimParticipation.
// ==================================================

// (19-1/2/3/4) The core business rule end-to-end: A's first bid succeeds, A's
// second bid is rejected, B can still bid, and A remains rejected even after B.
func TestPlaceBid_OneSuccessfulBidPerUserPerAuction(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST ONEBID SELLER")
	userA := createTestUser(t, env, "TEST ONEBID USER A")
	userB := createTestUser(t, env, "TEST ONEBID USER B")
	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")
	creditWallet(t, env, userA.ID, decimal.NewFromInt(1000))
	creditWallet(t, env, userB.ID, decimal.NewFromInt(1000))

	// (19-1) User A's first valid bid succeeds.
	bidA, err := env.bidSvc.PlaceBid(ctx, auction.ID, userA.ID, decimal.NewFromInt(110))
	if err != nil {
		t.Fatalf("(19-1) expected User A's first bid to succeed, got: %v", err)
	}
	t.Logf("(19-1) User A first bid succeeded: %s", bidA.ID)

	// (19-2) User A's second bid attempt (higher, otherwise fully valid) on the
	// SAME auction is rejected -- even though it would clear every other rule.
	_, err = env.bidSvc.PlaceBid(ctx, auction.ID, userA.ID, decimal.NewFromInt(150))
	if err != apperr.ErrDuplicateBidder {
		t.Fatalf("(19-2) expected ErrDuplicateBidder for User A's second bid, got: %v", err)
	}
	t.Logf("(19-2) User A's second bid correctly rejected: %v", err)

	// (19-3) User B can bid on the same auction.
	bidB, err := env.bidSvc.PlaceBid(ctx, auction.ID, userB.ID, decimal.NewFromInt(150))
	if err != nil {
		t.Fatalf("(19-3) expected User B's first bid to succeed, got: %v", err)
	}
	t.Logf("(19-3) User B first bid succeeded: %s", bidB.ID)

	// (19-4) User A remains permanently ineligible, even after being outbid by B.
	_, err = env.bidSvc.PlaceBid(ctx, auction.ID, userA.ID, decimal.NewFromInt(200))
	if err != apperr.ErrDuplicateBidder {
		t.Fatalf("(19-4) expected User A to remain rejected after User B's bid, got: %v", err)
	}
	t.Logf("(19-4) User A still rejected after User B's bid: %v", err)

	// (19-10) Exactly one guard row exists for (A, auction) despite the two
	// attempts by A.
	var guardCount int
	if err := env.db.GetContext(ctx, &guardCount,
		`SELECT COUNT(*) FROM auction_bid_participants WHERE auction_id = $1 AND user_id = $2`,
		auction.ID, userA.ID); err != nil {
		t.Fatalf("failed to count guard rows for User A: %v", err)
	}
	if guardCount != 1 {
		t.Fatalf("(19-10) expected exactly 1 guard row for (User A, auction), got %d", guardCount)
	}

	// (19-11) Existing bid history is preserved -- both successful bids are
	// still visible, nothing was deleted or merged.
	history, err := env.bidSvc.GetHistory(ctx, auction.ID)
	if err != nil {
		t.Fatalf("GetHistory failed: %v", err)
	}
	if len(history) != 2 {
		t.Fatalf("(19-11) expected 2 bids in history (A's successful first bid + B's), got %d", len(history))
	}

	// (19-12) Winner selection still works: B is the top bid. Queried
	// directly here to isolate exactly what this test is proving (the
	// NULL-scan issue this comment used to describe on bidRepo.FindTopBid's
	// bidder_name/bidder_phone columns is now fixed -- see
	// TestFindTopBid_NullBidderDetails_Safe and friends below).
	var topUserID uuid.UUID
	if err := env.db.GetContext(ctx, &topUserID,
		`SELECT user_id FROM bids WHERE auction_id = $1 ORDER BY amount DESC LIMIT 1`, auction.ID); err != nil {
		t.Fatalf("failed to query top bid: %v", err)
	}
	if topUserID != userB.ID {
		t.Fatalf("(19-12) expected User B to be the top bidder, got user %s", topUserID)
	}

	t.Logf("confirmed: one-successful-bid-per-user-per-auction rule enforced end-to-end, bid history and winner selection unaffected")
}

// (19-5) A too-low, rejected first attempt does NOT consume the user's
// eligibility -- a subsequent valid bid from the same user must still succeed.
func TestPlaceBid_TooLowFirstAttempt_DoesNotConsumeEligibility(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST ONEBID TOOLOW SELLER")
	bidder := createTestUser(t, env, "TEST ONEBID TOOLOW BIDDER")
	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")
	creditWallet(t, env, bidder.ID, decimal.NewFromInt(1000))

	// Too low: current_price=100, min_increment=10 -> 105 < 110 required.
	_, err := env.bidSvc.PlaceBid(ctx, auction.ID, bidder.ID, decimal.NewFromInt(105))
	if err != apperr.ErrBidTooLow {
		t.Fatalf("precondition failed: expected ErrBidTooLow, got: %v", err)
	}

	// The same user's valid bid must still succeed -- the failed attempt above
	// must not have claimed a guard row.
	bid, err := env.bidSvc.PlaceBid(ctx, auction.ID, bidder.ID, decimal.NewFromInt(110))
	if err != nil {
		t.Fatalf("(19-5) expected the same user's valid bid to succeed after a too-low failed attempt, got: %v", err)
	}
	t.Logf("(19-5) too-low failed attempt did not consume eligibility; subsequent valid bid succeeded: %s", bid.ID)
}

// (19-6) Self-bid (seller bidding on their own auction) does NOT consume
// eligibility -- it's rejected before the guard claim, and even if the seller
// were later a legitimate bidder on a DIFFERENT auction, this must not affect
// anything (this test only proves no guard row was created for the rejected
// attempt itself).
func TestPlaceBid_SelfBid_DoesNotConsumeEligibility(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST ONEBID SELFBID SELLER")
	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")

	_, err := env.bidSvc.PlaceBid(ctx, auction.ID, seller.ID, decimal.NewFromInt(110))
	if err != apperr.ErrSelfBid {
		t.Fatalf("(19-6) expected ErrSelfBid, got: %v", err)
	}

	var guardCount int
	if err := env.db.GetContext(ctx, &guardCount,
		`SELECT COUNT(*) FROM auction_bid_participants WHERE auction_id = $1 AND user_id = $2`,
		auction.ID, seller.ID); err != nil {
		t.Fatalf("failed to count guard rows: %v", err)
	}
	if guardCount != 0 {
		t.Fatalf("(19-6) expected no guard row after a rejected self-bid, got %d", guardCount)
	}
	t.Logf("(19-6) self-bid rejection correctly left no guard row")
}

// (19-7) A failed attempt on an already-ended auction does NOT consume
// eligibility.
func TestPlaceBid_EndedAuction_DoesNotConsumeEligibility(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST ONEBID ENDED SELLER")
	bidder := createTestUser(t, env, "TEST ONEBID ENDED BIDDER")
	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")
	creditWallet(t, env, bidder.ID, decimal.NewFromInt(1000))

	if _, err := env.db.ExecContext(ctx, `UPDATE auctions SET status = 'ended' WHERE id = $1`, auction.ID); err != nil {
		t.Fatalf("failed to mark auction ended: %v", err)
	}

	_, err := env.bidSvc.PlaceBid(ctx, auction.ID, bidder.ID, decimal.NewFromInt(110))
	if err != apperr.ErrAuctionNotActive {
		t.Fatalf("(19-7) expected ErrAuctionNotActive, got: %v", err)
	}

	var guardCount int
	if err := env.db.GetContext(ctx, &guardCount,
		`SELECT COUNT(*) FROM auction_bid_participants WHERE auction_id = $1 AND user_id = $2`,
		auction.ID, bidder.ID); err != nil {
		t.Fatalf("failed to count guard rows: %v", err)
	}
	if guardCount != 0 {
		t.Fatalf("(19-7) expected no guard row after a rejected ended-auction bid, got %d", guardCount)
	}
	t.Logf("(19-7) ended-auction rejection correctly left no guard row")
}

// (19-8) A failure AFTER the guard claim but before commit (insufficient
// insurance balance) rolls back the guard row too -- a failed bid must never
// permanently consume eligibility, proving the claim and the bid are truly
// atomic within the same transaction.
func TestPlaceBid_InsuranceFailureAfterClaim_RollsBackEligibility(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST ONEBID INSURANCE SELLER")
	bidder := createTestUser(t, env, "TEST ONEBID INSURANCE BIDDER")
	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")
	// Wallet exists but has a ZERO balance -- insurance_amount=20 (from
	// createTestAuction) so ErrInsufficientForInsurance fires AFTER the
	// guard claim (step 2c) but before the transaction commits.
	creditWallet(t, env, bidder.ID, decimal.Zero)

	_, err := env.bidSvc.PlaceBid(ctx, auction.ID, bidder.ID, decimal.NewFromInt(110))
	if err != apperr.ErrInsufficientForInsurance {
		t.Fatalf("(19-8) expected ErrInsufficientForInsurance, got: %v", err)
	}

	var guardCount int
	if err := env.db.GetContext(ctx, &guardCount,
		`SELECT COUNT(*) FROM auction_bid_participants WHERE auction_id = $1 AND user_id = $2`,
		auction.ID, bidder.ID); err != nil {
		t.Fatalf("failed to count guard rows: %v", err)
	}
	if guardCount != 0 {
		t.Fatalf("(19-8) CRITICAL: expected the guard row to be rolled back after insurance failure, got %d rows -- a failed bid must never permanently consume eligibility", guardCount)
	}

	// The same user's bid must still succeed once they have sufficient balance.
	creditWallet(t, env, bidder.ID, decimal.NewFromInt(1000))
	bid, err := env.bidSvc.PlaceBid(ctx, auction.ID, bidder.ID, decimal.NewFromInt(110))
	if err != nil {
		t.Fatalf("(19-8) expected the retried bid to succeed after crediting the wallet, got: %v", err)
	}
	t.Logf("(19-8) insurance failure correctly rolled back the guard claim; retried bid succeeded: %s", bid.ID)
}

// (19-9) Two concurrent first-bid attempts by the SAME user cannot both
// commit -- proves the DB-level UNIQUE constraint, not just the application
// check, is what actually prevents a race (goroutines racing PlaceBid
// directly against real Postgres).
func TestPlaceBid_ConcurrentFirstBids_OnlyOneCommits(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST ONEBID RACE SELLER")
	bidder := createTestUser(t, env, "TEST ONEBID RACE BIDDER")
	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")
	creditWallet(t, env, bidder.ID, decimal.NewFromInt(10000))

	const n = 8
	results := make(chan error, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(amount int64) {
			defer wg.Done()
			_, err := env.bidSvc.PlaceBid(ctx, auction.ID, bidder.ID, decimal.NewFromInt(110+amount))
			results <- err
		}(int64(i))
	}
	wg.Wait()
	close(results)

	successCount := 0
	duplicateCount := 0
	for err := range results {
		if err == nil {
			successCount++
		} else if err == apperr.ErrDuplicateBidder || err == apperr.ErrBidConflict {
			duplicateCount++
		} else {
			t.Logf("(19-9) unexpected error from concurrent attempt (acceptable if a transient conflict): %v", err)
			duplicateCount++
		}
	}

	if successCount != 1 {
		t.Fatalf("(19-9) CRITICAL: expected exactly 1 of %d concurrent first-bid attempts by the same user to succeed, got %d successes", n, successCount)
	}

	var guardCount int
	if err := env.db.GetContext(ctx, &guardCount,
		`SELECT COUNT(*) FROM auction_bid_participants WHERE auction_id = $1 AND user_id = $2`,
		auction.ID, bidder.ID); err != nil {
		t.Fatalf("failed to count guard rows: %v", err)
	}
	if guardCount != 1 {
		t.Fatalf("(19-9) CRITICAL: expected exactly 1 guard row after %d concurrent attempts, got %d", n, guardCount)
	}
	t.Logf("(19-9) %d concurrent first-bid attempts by the same user: exactly 1 succeeded, exactly 1 guard row exists", n)
}

// (19-13/14/15) Migration-time safety: historical duplicate bid rows are
// NEVER deleted or rewritten, and the backfill produces exactly one guard
// row per distinct (auction_id, user_id) pair -- proven directly against the
// migration's own SQL logic (not re-running the real migration file, since
// setupEnv's schema is already migrated; this exercises the identical
// INSERT ... SELECT DISTINCT ON logic from 000050_auction_bid_participants.up.sql
// against a fresh fixture to prove it end-to-end).
func TestAuctionBidParticipants_BackfillPreservesHistoricalDuplicates(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST ONEBID BACKFILL SELLER")
	userA := createTestUser(t, env, "TEST ONEBID BACKFILL USER A")
	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")

	// Insert 3 historical bid rows for the SAME (auction, user) pair directly
	// against bids -- simulating pre-existing duplicate data exactly like the
	// 18 real duplicate groups found in Staging, bypassing PlaceBid/the guard
	// entirely (as real historical rows predating this feature would).
	for i := 0; i < 3; i++ {
		if _, err := env.db.ExecContext(ctx, `
			INSERT INTO bids (id, auction_id, user_id, amount, is_winning)
			VALUES (gen_random_uuid(), $1, $2, $3, false)`,
			auction.ID, userA.ID, decimal.NewFromInt(int64(110+i*10))); err != nil {
			t.Fatalf("failed to insert historical duplicate bid %d: %v", i, err)
		}
	}

	var bidsBefore int
	if err := env.db.GetContext(ctx, &bidsBefore, `SELECT COUNT(*) FROM bids WHERE auction_id = $1 AND user_id = $2`, auction.ID, userA.ID); err != nil {
		t.Fatalf("failed to count bids before backfill: %v", err)
	}
	if bidsBefore != 3 {
		t.Fatalf("precondition failed: expected 3 historical bid rows, got %d", bidsBefore)
	}

	// Clear any guard row this fixture might already have (none expected --
	// bidRepo.Create was never called) and run the exact backfill INSERT from
	// the migration for just this pair, to prove it produces exactly one row.
	if _, err := env.db.ExecContext(ctx, `DELETE FROM auction_bid_participants WHERE auction_id = $1 AND user_id = $2`, auction.ID, userA.ID); err != nil {
		t.Fatalf("failed to clear pre-existing guard rows: %v", err)
	}
	if _, err := env.db.ExecContext(ctx, `
		INSERT INTO auction_bid_participants (auction_id, user_id, first_bid_id, created_at)
		SELECT DISTINCT ON (b.auction_id, b.user_id)
		    b.auction_id, b.user_id, b.id, b.created_at
		FROM bids b
		WHERE b.auction_id = $1 AND b.user_id = $2
		ORDER BY b.auction_id, b.user_id, b.created_at ASC, b.id ASC`,
		auction.ID, userA.ID); err != nil {
		t.Fatalf("backfill INSERT failed: %v", err)
	}

	var bidsAfter int
	if err := env.db.GetContext(ctx, &bidsAfter, `SELECT COUNT(*) FROM bids WHERE auction_id = $1 AND user_id = $2`, auction.ID, userA.ID); err != nil {
		t.Fatalf("failed to count bids after backfill: %v", err)
	}
	if bidsAfter != 3 {
		t.Fatalf("(19-14) CRITICAL: expected all 3 historical bid rows to survive backfill unchanged, got %d -- no historical bid may EVER be deleted or rewritten", bidsAfter)
	}

	var guardCount int
	if err := env.db.GetContext(ctx, &guardCount,
		`SELECT COUNT(*) FROM auction_bid_participants WHERE auction_id = $1 AND user_id = $2`,
		auction.ID, userA.ID); err != nil {
		t.Fatalf("failed to count guard rows: %v", err)
	}
	if guardCount != 1 {
		t.Fatalf("(19-15) expected exactly 1 guard row backfilled from 3 historical duplicate bids, got %d", guardCount)
	}

	// A NEW bid attempt by this same user on this auction must now be blocked.
	creditWallet(t, env, userA.ID, decimal.NewFromInt(1000))
	_, err := env.bidSvc.PlaceBid(ctx, auction.ID, userA.ID, decimal.NewFromInt(500))
	if err != apperr.ErrDuplicateBidder {
		t.Fatalf("(19-13) expected a new bid attempt by a user with historical duplicate bids to be blocked, got: %v", err)
	}

	t.Logf("confirmed: 3 historical duplicate bid rows preserved unchanged (%d before, %d after), backfill produced exactly 1 guard row, new bid attempt correctly blocked", bidsBefore, bidsAfter)
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

	// Same real-admin-workflow insurance setup as TestReviewAuctionRequest_ApproveCreatesPublicAuction
	// above -- see that test's comment for the full explanation.
	insuranceUpdates := *req
	insuranceUpdates.InsuranceAmount = decimal.NewFromInt(500)
	if err := env.reqSvc.AdminUpdateAuctionRequest(ctx, req.ID, &insuranceUpdates, nil); err != nil {
		t.Fatalf("AdminUpdateAuctionRequest (setting insurance) failed: %v", err)
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

	txn, err := walletSvc.InitiateDeposit(ctx, user.ID, decimal.NewFromInt(100), "bank_transfer", "bank_transfer", "", nil)
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
	txn, err := walletSvc.InitiateDeposit(ctx, user.ID, decimal.NewFromInt(50), "EUR", "USD", "", nil)
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

// === Bug I: Favorites/Home image consistency ===
//
// Client feedback: a favorited auction with a real, persisted image showed
// the neutral Bug H placeholder in Favorites while the SAME auction showed
// its real image correctly on Home. Root cause: favoriteRepo.ListByUserID's
// query never joined auction_images (a.ImageURLs was always NULL), and
// UserHandler.ListFavorites returned the raw model instead of building the
// same "images" response field AuctionHandler.List already builds via
// GetImagesArray(). These tests exercise both the repository fix (image_urls
// populated) and the handler fix (images field present in the JSON
// response), using the real Postgres-backed testEnv already established by
// this file for prior Favorites tests (see (T) above).

// httpGetFavorites calls GET /favorites/as/:userID (test-only route wired in
// setupEnv) and parses the JSON body into a slice of raw maps, mirroring
// exactly what favorites_page.dart/favorites_service.dart receive as
// response.data.
func httpGetFavorites(t *testing.T, env *testEnv, callerID uuid.UUID) []map[string]interface{} {
	t.Helper()
	req := httptest.NewRequest("GET", fmt.Sprintf("/favorites/as/%s", callerID), nil)
	resp, err := env.app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("GET /favorites/as/%s: expected 200, got %d", callerID, resp.StatusCode)
	}
	var body struct {
		Success bool                     `json:"success"`
		Data    []map[string]interface{} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode favorites response: %v", err)
	}
	if !body.Success {
		t.Fatalf("expected success=true in favorites response")
	}
	return body.Data
}

// findFavoriteByID locates one auction's entry in a parsed favorites response
// by its "id" field.
func findFavoriteByID(favorites []map[string]interface{}, auctionID uuid.UUID) map[string]interface{} {
	target := auctionID.String()
	for _, f := range favorites {
		if id, _ := f["id"].(string); id == target {
			return f
		}
	}
	return nil
}

// (1) A favorited auction with exactly one persisted image returns
// images=[URL] in the Favorites HTTP response -- the exact contract mobile's
// favorites_page.dart (auction['images']) and Auction.fromJson (Home) both
// already parse identically.
func TestFavorites_SingleImage_ReturnsImagesArray(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST FAVIMG SELLER 1")
	user := createTestUser(t, env, "TEST FAVIMG USER 1")
	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")

	imageURL := "https://pub-test.r2.dev/auctions/" + auction.ID.String() + "/photo.jpg"
	seedAuctionImage(t, env, auction.ID, imageURL, 1)

	if err := env.userSvc.AddFavorite(ctx, user.ID, auction.ID); err != nil {
		t.Fatalf("AddFavorite failed: %v", err)
	}

	favorites := httpGetFavorites(t, env, user.ID)
	entry := findFavoriteByID(favorites, auction.ID)
	if entry == nil {
		t.Fatalf("favorited auction %s not found in favorites response", auction.ID)
	}

	images, ok := entry["images"].([]interface{})
	if !ok {
		t.Fatalf("expected \"images\" to be a JSON array, got %T: %v", entry["images"], entry["images"])
	}
	if len(images) != 1 {
		t.Fatalf("expected exactly 1 image, got %d: %v", len(images), images)
	}
	if images[0] != imageURL {
		t.Fatalf("expected image URL %q, got %q", imageURL, images[0])
	}
}

// (2) A favorited auction with multiple images returns them in
// display_order -- proves the string_agg subquery's ORDER BY is preserved
// end-to-end through the handler's GetImagesArray() split.
func TestFavorites_MultipleImages_PreservesDisplayOrder(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST FAVIMG SELLER 2")
	user := createTestUser(t, env, "TEST FAVIMG USER 2")
	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")

	urlFirst := "https://pub-test.r2.dev/auctions/" + auction.ID.String() + "/first.jpg"
	urlSecond := "https://pub-test.r2.dev/auctions/" + auction.ID.String() + "/second.jpg"
	urlThird := "https://pub-test.r2.dev/auctions/" + auction.ID.String() + "/third.jpg"
	// Seeded out of order on purpose -- display_order (not insertion order)
	// must determine the returned sequence.
	seedAuctionImage(t, env, auction.ID, urlThird, 3)
	seedAuctionImage(t, env, auction.ID, urlFirst, 1)
	seedAuctionImage(t, env, auction.ID, urlSecond, 2)

	if err := env.userSvc.AddFavorite(ctx, user.ID, auction.ID); err != nil {
		t.Fatalf("AddFavorite failed: %v", err)
	}

	favorites := httpGetFavorites(t, env, user.ID)
	entry := findFavoriteByID(favorites, auction.ID)
	if entry == nil {
		t.Fatalf("favorited auction %s not found in favorites response", auction.ID)
	}
	images, ok := entry["images"].([]interface{})
	if !ok || len(images) != 3 {
		t.Fatalf("expected exactly 3 images, got %v", entry["images"])
	}
	want := []string{urlFirst, urlSecond, urlThird}
	for i, w := range want {
		if images[i] != w {
			t.Fatalf("image at position %d: expected %q, got %q (full order: %v)", i, w, images[i], images)
		}
	}
}

// (3) A favorited auction with NO persisted images returns an empty images
// array -- never omits the field, never fabricates a placeholder URL (Bug H
// contract: the mobile app's neutral fallback is what must render this
// case, never a fake stock photo).
func TestFavorites_NoImages_ReturnsEmptyArrayNeverFabricated(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST FAVIMG SELLER 3")
	user := createTestUser(t, env, "TEST FAVIMG USER 3")
	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")
	// No seedAuctionImage call -- this auction genuinely has zero images.

	if err := env.userSvc.AddFavorite(ctx, user.ID, auction.ID); err != nil {
		t.Fatalf("AddFavorite failed: %v", err)
	}

	favorites := httpGetFavorites(t, env, user.ID)
	entry := findFavoriteByID(favorites, auction.ID)
	if entry == nil {
		t.Fatalf("favorited auction %s not found in favorites response", auction.ID)
	}
	images, ok := entry["images"].([]interface{})
	if !ok {
		t.Fatalf("expected \"images\" key present as an array (possibly empty), got %T: %v", entry["images"], entry["images"])
	}
	if len(images) != 0 {
		t.Fatalf("expected 0 images for an auction with none persisted, got %d: %v", len(images), images)
	}
	// Explicit anti-Bug-H-regression check: no fabricated/placeholder URL string.
	for _, img := range images {
		if s, _ := img.(string); strings.Contains(strings.ToLower(s), "corolla") {
			t.Fatalf("CRITICAL: a fabricated placeholder image URL leaked into the Favorites response: %q", s)
		}
	}
}

// (4) The SAME auction returns the SAME first/current image URL from both
// Home's list endpoint (AuctionHandler.List, via auctionRepo.FindAll) and
// Favorites (UserHandler.ListFavorites) -- the exact consistency contract
// Bug I was filed against.
func TestFavorites_HomeConsistency_SameAuctionSameImage(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST FAVIMG SELLER 4")
	user := createTestUser(t, env, "TEST FAVIMG USER 4")
	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")

	imageURL := "https://pub-test.r2.dev/auctions/" + auction.ID.String() + "/consistent.jpg"
	seedAuctionImage(t, env, auction.ID, imageURL, 1)

	if err := env.userSvc.AddFavorite(ctx, user.ID, auction.ID); err != nil {
		t.Fatalf("AddFavorite failed: %v", err)
	}

	// Home's data source: auctionRepo.FindAll (same repository method
	// AuctionHandler.List calls, see auction_handler.go's List handler).
	homeAuctions, _, err := env.auctionRepo.FindAll(ctx, repository.AuctionFilters{Status: "active"})
	if err != nil {
		t.Fatalf("FindAll failed: %v", err)
	}
	var homeAuction *models.Auction
	for i := range homeAuctions {
		if homeAuctions[i].ID == auction.ID {
			homeAuction = &homeAuctions[i]
		}
	}
	if homeAuction == nil {
		t.Fatalf("auction %s not found via Home's FindAll", auction.ID)
	}
	homeImages := homeAuction.GetImagesArray()
	if len(homeImages) != 1 || homeImages[0] != imageURL {
		t.Fatalf("expected Home to show image %q, got %v", imageURL, homeImages)
	}

	favorites := httpGetFavorites(t, env, user.ID)
	entry := findFavoriteByID(favorites, auction.ID)
	if entry == nil {
		t.Fatalf("favorited auction %s not found in favorites response", auction.ID)
	}
	favImages, ok := entry["images"].([]interface{})
	if !ok || len(favImages) != 1 {
		t.Fatalf("expected Favorites to show exactly 1 image, got %v", entry["images"])
	}
	if favImages[0] != imageURL {
		t.Fatalf("IMAGE INCONSISTENCY (Bug I regression): Home shows %q but Favorites shows %q for the SAME auction %s", imageURL, favImages[0], auction.ID)
	}
}

// (5) Bug C preservation: the Favorites response wrapper contract (data is a
// plain JSON array, never a {"favorites": [...]} nested wrapper) still
// holds after the Bug I handler change -- only each element's shape gained
// fields, the outer response.data shape is untouched.
func TestFavorites_BugC_WrapperShapePreserved(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST FAVIMG SELLER 5")
	user := createTestUser(t, env, "TEST FAVIMG USER 5")
	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")
	if err := env.userSvc.AddFavorite(ctx, user.ID, auction.ID); err != nil {
		t.Fatalf("AddFavorite failed: %v", err)
	}

	req := httptest.NewRequest("GET", fmt.Sprintf("/favorites/as/%s", user.ID), nil)
	resp, err := env.app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	defer resp.Body.Close()

	var raw map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	data, ok := raw["data"]
	if !ok {
		t.Fatalf("expected top-level \"data\" key, response: %v", raw)
	}
	if _, isList := data.([]interface{}); !isList {
		t.Fatalf("Bug C REGRESSION: expected response.data to be a plain JSON array, got %T -- extractFavoriteAuctionIds/extractFavoriteAuctions in favorites_service.dart expect exactly this shape", data)
	}
	if _, hasNestedWrapper := raw["favorites"]; hasNestedWrapper {
		t.Fatalf("Bug C REGRESSION: response must never contain a nested \"favorites\" wrapper key")
	}
}

// (6) Unrelated favorite fields (title, price, status, seller_id, etc.)
// remain present and unchanged after the Bug I handler rewrite -- proves the
// new fiber.Map construction is a faithful mirror of AuctionHandler.List's
// fields, not an accidental narrowing of the response.
func TestFavorites_UnrelatedFieldsUnchanged(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST FAVIMG SELLER 6")
	user := createTestUser(t, env, "TEST FAVIMG USER 6")
	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")
	if err := env.userSvc.AddFavorite(ctx, user.ID, auction.ID); err != nil {
		t.Fatalf("AddFavorite failed: %v", err)
	}

	favorites := httpGetFavorites(t, env, user.ID)
	entry := findFavoriteByID(favorites, auction.ID)
	if entry == nil {
		t.Fatalf("favorited auction %s not found in favorites response", auction.ID)
	}

	for _, field := range []string{
		"id", "seller_id", "category_id", "title_ar", "start_price",
		"current_price", "status", "lot_number", "views", "bidder_count",
		"currency_code", "market_country_iso",
	} {
		if _, present := entry[field]; !present {
			t.Fatalf("expected unrelated field %q to still be present in favorites response, it was dropped", field)
		}
	}
	if entry["status"] != auction.Status {
		t.Fatalf("expected status %q, got %v", auction.Status, entry["status"])
	}
	if entry["seller_id"] != seller.ID.String() {
		t.Fatalf("expected seller_id %q, got %v", seller.ID.String(), entry["seller_id"])
	}
}

// (7) Cross-user favorites isolation remains intact after the Bug I
// rewrite: user B's favorite images/data must never appear in user A's
// favorites response.
func TestFavorites_CrossUserIsolation_ImagesNotLeaked(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST FAVIMG SELLER 7")
	userA := createTestUser(t, env, "TEST FAVIMG USER 7A")
	userB := createTestUser(t, env, "TEST FAVIMG USER 7B")
	auctionA := createTestAuction(t, env, seller.ID, "MR", "MRU")
	auctionB := createTestAuction(t, env, seller.ID, "MR", "MRU")

	imageA := "https://pub-test.r2.dev/auctions/" + auctionA.ID.String() + "/a.jpg"
	imageB := "https://pub-test.r2.dev/auctions/" + auctionB.ID.String() + "/b.jpg"
	seedAuctionImage(t, env, auctionA.ID, imageA, 1)
	seedAuctionImage(t, env, auctionB.ID, imageB, 1)

	if err := env.userSvc.AddFavorite(ctx, userA.ID, auctionA.ID); err != nil {
		t.Fatalf("AddFavorite(A) failed: %v", err)
	}
	if err := env.userSvc.AddFavorite(ctx, userB.ID, auctionB.ID); err != nil {
		t.Fatalf("AddFavorite(B) failed: %v", err)
	}

	favoritesA := httpGetFavorites(t, env, userA.ID)
	if findFavoriteByID(favoritesA, auctionB.ID) != nil {
		t.Fatalf("CRITICAL: user B's favorite auction leaked into user A's favorites response")
	}
	entryA := findFavoriteByID(favoritesA, auctionA.ID)
	if entryA == nil {
		t.Fatalf("user A's own favorite auction missing from their own favorites response")
	}
	imgsA, _ := entryA["images"].([]interface{})
	if len(imgsA) != 1 || imgsA[0] != imageA {
		t.Fatalf("expected user A's favorite to show image %q, got %v", imageA, imgsA)
	}
}

// (9) Customer #20 realtime-refetch compatibility: the Favorites HTTP
// response contract that RealtimeSyncService's auction.updated handler
// triggers a plain re-fetch against (favorites_page.dart calls
// FavoritesService().getFavoriteAuctions(), which calls this same endpoint)
// is unchanged in kind -- still a 200 with a plain array in response.data,
// now simply carrying real images. No new fields required on the mobile
// side, confirming the realtime wiring needs no changes.
func TestFavorites_Customer20RefetchContractUnchanged(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST FAVIMG SELLER 9")
	user := createTestUser(t, env, "TEST FAVIMG USER 9")
	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")
	if err := env.userSvc.AddFavorite(ctx, user.ID, auction.ID); err != nil {
		t.Fatalf("AddFavorite failed: %v", err)
	}

	// Simulate the realtime handler's refetch: calling the SAME endpoint a
	// second time (e.g. after an auction.updated event) must keep returning
	// a plain 200 + array, and must reflect an image added between calls --
	// i.e. a subsequent fetch after content changes picks up the change,
	// exactly as auction.updated -> _loadAuctions() -> a fresh GET expects.
	before := httpGetFavorites(t, env, user.ID)
	entryBefore := findFavoriteByID(before, auction.ID)
	if entryBefore == nil {
		t.Fatalf("favorite missing on first fetch")
	}
	imagesBefore, _ := entryBefore["images"].([]interface{})
	if len(imagesBefore) != 0 {
		t.Fatalf("expected no images yet, got %v", imagesBefore)
	}

	newImage := "https://pub-test.r2.dev/auctions/" + auction.ID.String() + "/added-later.jpg"
	seedAuctionImage(t, env, auction.ID, newImage, 1)

	after := httpGetFavorites(t, env, user.ID)
	entryAfter := findFavoriteByID(after, auction.ID)
	if entryAfter == nil {
		t.Fatalf("favorite missing on refetch")
	}
	imagesAfter, ok := entryAfter["images"].([]interface{})
	if !ok || len(imagesAfter) != 1 || imagesAfter[0] != newImage {
		t.Fatalf("expected refetch to reflect the newly added image %q, got %v", newImage, imagesAfter)
	}
}

// (10) An auction with genuinely no image never has a URL fabricated for it
// -- direct repository-level check that image_urls is an empty string
// (never a placeholder), independent of the handler transformation tested
// above.
func TestFavoriteRepo_NoImages_ImageURLsEmptyNeverFabricated(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST FAVIMG SELLER 10")
	user := createTestUser(t, env, "TEST FAVIMG USER 10")
	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")
	if err := env.userSvc.AddFavorite(ctx, user.ID, auction.ID); err != nil {
		t.Fatalf("AddFavorite failed: %v", err)
	}

	favoriteRepo := repository.NewFavoriteRepository(env.db)
	favorites, err := favoriteRepo.ListByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("ListByUserID failed: %v", err)
	}
	var found *models.Auction
	for i := range favorites {
		if favorites[i].ID == auction.ID {
			found = &favorites[i]
		}
	}
	if found == nil {
		t.Fatalf("favorited auction not found via ListByUserID")
	}
	if found.ImageURLs == nil {
		t.Fatalf("expected ImageURLs to be a non-nil empty string (COALESCE default), got nil")
	}
	if *found.ImageURLs != "" {
		t.Fatalf("expected empty image_urls for an auction with no images, got %q", *found.ImageURLs)
	}
	images := found.GetImagesArray()
	if len(images) != 0 {
		t.Fatalf("expected 0 images, got %v", images)
	}
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

// === Lot number >99 collision hotfix (migration 000047) ===
//
// generate_lot_number() (migration 000005) used LPAD(nextval(...)::TEXT, 2, '0'),
// and PostgreSQL's LPAD TRUNCATES an input already longer than the target width
// (confirmed: LPAD('100','2','0') = '10'). Once the sequence passed 99, every new
// lot_number silently collapsed to its first two digits and collided with an
// already-used value. Migration 000047 changes the padding width to
// GREATEST(2, LENGTH(v::TEXT)) -- a floor, not a ceiling -- so small values keep
// their existing 2-digit look while nothing is ever truncated.
//
// These tests exercise the real DB trigger directly (not application code) inside
// a transaction that is always rolled back, using setval() to place the sequence
// at exactly the value under test -- this proves the FUNCTION's behavior without
// depending on, or perturbing, this local DB's actual accumulated sequence state
// or its 197 pre-existing seed rows. Per the hotfix instructions, no production-
// like row data is altered: every insert here is undone by ROLLBACK.

// setSequenceAndInsertLotNumber places auctions_lot_number_seq so that the next
// nextval() call returns exactly `target`, inserts one fixture auction (letting
// the trigger generate lot_number), and returns the generated value -- all inside
// the caller's already-open transaction (rolled back by the caller, never
// committed).
func setSequenceAndInsertLotNumber(t *testing.T, tx *sqlx.Tx, target int64) string {
	t.Helper()
	if _, err := tx.Exec(`SELECT setval('auctions_lot_number_seq', $1, true)`, target-1); err != nil {
		t.Fatalf("failed to set sequence to %d: %v", target-1, err)
	}
	var lotNumber string
	err := tx.QueryRow(`
		INSERT INTO auctions (id, seller_id, category_id, title_ar, start_price, current_price, min_increment, insurance_amount, start_time, end_time, status)
		SELECT gen_random_uuid(), (SELECT id FROM users LIMIT 1), 1, 'lot number hotfix test', 10, 10, 1, 1, now(), now() + interval '1 day', 'pending'
		RETURNING lot_number
	`).Scan(&lotNumber)
	if err != nil {
		t.Fatalf("failed to insert fixture auction at sequence target %d: %v", target, err)
	}
	return lotNumber
}

// (lot-1) Small values (1, 9) still get the existing 2-digit zero-padded format.
// Asserted directly against the fixed padding SQL expression (rather than via a
// live INSERT) because every low sequence value in this local DB's real range
// (LOT-01..LOT-09) is already taken by pre-existing seed data -- the expression
// under test is byte-for-byte identical to what generate_lot_number() (migration
// 000047) evaluates, so this is an exact, not an approximate, proof.
func TestLotNumber_SmallValues_TwoDigitPadding(t *testing.T) {
	env := setupEnv(t)
	tx, err := env.db.Beginx()
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	var padded1, padded9 string
	if err := tx.QueryRow(`SELECT LPAD(1::TEXT, GREATEST(2, LENGTH(1::TEXT)), '0')`).Scan(&padded1); err != nil {
		t.Fatalf("failed to evaluate padding expression for 1: %v", err)
	}
	if err := tx.QueryRow(`SELECT LPAD(9::TEXT, GREATEST(2, LENGTH(9::TEXT)), '0')`).Scan(&padded9); err != nil {
		t.Fatalf("failed to evaluate padding expression for 9: %v", err)
	}
	if padded1 != "01" {
		t.Fatalf("(lot-1) expected padding(1) = '01', got %q", padded1)
	}
	if padded9 != "09" {
		t.Fatalf("(lot-1) expected padding(9) = '09', got %q", padded9)
	}
	t.Logf("(lot-1) small-value padding unchanged: 1->%s, 9->%s", padded1, padded9)
}

// (lot-2) 99 unchanged (still exactly 2 digits, no regression at the old boundary).
func TestLotNumber_99_Unchanged(t *testing.T) {
	env := setupEnv(t)
	tx, err := env.db.Beginx()
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	var padded99 string
	if err := tx.QueryRow(`SELECT LPAD(99::TEXT, GREATEST(2, LENGTH(99::TEXT)), '0')`).Scan(&padded99); err != nil {
		t.Fatalf("failed to evaluate padding expression for 99: %v", err)
	}
	if padded99 != "99" {
		t.Fatalf("(lot-2) expected padding(99) = '99', got %q", padded99)
	}
	t.Logf("(lot-2) 99 unchanged: %s", padded99)
}

// (lot-3) 100 and 101 are NOT truncated (the actual regression this hotfix fixes),
// exercised against the real trigger via a genuine INSERT, not just the SQL
// expression in isolation.
func TestLotNumber_100And101_NotTruncated(t *testing.T) {
	env := setupEnv(t)
	tx, err := env.db.Beginx()
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	got100 := setSequenceAndInsertLotNumber(t, tx, 998877100)
	got101 := setSequenceAndInsertLotNumber(t, tx, 998877101)
	if got100 != "LOT-998877100" {
		t.Fatalf("(lot-3) CRITICAL: expected LOT-998877100, got %q (truncation regression)", got100)
	}
	if got101 != "LOT-998877101" {
		t.Fatalf("(lot-3) CRITICAL: expected LOT-998877101, got %q (truncation regression)", got101)
	}
	if got100 == got101 {
		t.Fatalf("(lot-3) CRITICAL: sequential values collapsed to the same lot_number: %q", got100)
	}
	t.Logf("(lot-3) large sequence values not truncated: %s, %s", got100, got101)
}

// (lot-4) Generated lot numbers remain globally unique across a batch spanning the
// old 2-digit boundary and beyond -- no two inserts in the same batch collide.
func TestLotNumber_GeneratedValuesRemainUnique(t *testing.T) {
	env := setupEnv(t)
	tx, err := env.db.Beginx()
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	targets := []int64{998877001, 998877002, 998877098, 998877099, 998877100, 998877101, 998877999, 998878000}
	seen := map[string]bool{}
	for _, target := range targets {
		got := setSequenceAndInsertLotNumber(t, tx, target)
		if seen[got] {
			t.Fatalf("(lot-4) CRITICAL: duplicate lot_number %q generated within the same batch", got)
		}
		seen[got] = true
	}
	if len(seen) != len(targets) {
		t.Fatalf("(lot-4) expected %d unique lot numbers, got %d", len(targets), len(seen))
	}
	t.Logf("(lot-4) all %d generated lot numbers across the boundary are unique: %v", len(seen), seen)
}

// (lot-5) Explicit LotNumber assignment (the existing "TEST-"+uuid fixture pattern
// used throughout this file, e.g. createTestAuction) is unaffected by this hotfix
// -- the trigger only fires when lot_number IS NULL OR ”.
func TestLotNumber_ExplicitAssignmentUnaffected(t *testing.T) {
	env := setupEnv(t)
	seller := createTestUser(t, env, "TEST LOTNUMBER EXPLICIT SELLER")
	auction := createTestAuction(t, env, seller.ID, "MR", "MRU") // sets an explicit "TEST-"+uuid LotNumber

	if auction.LotNumber == nil || !strings.HasPrefix(*auction.LotNumber, "TEST-") {
		t.Fatalf("(lot-5) expected explicit fixture LotNumber to be preserved (TEST-... prefix), got %v", auction.LotNumber)
	}
	t.Logf("(lot-5) explicit LotNumber assignment unaffected by trigger: %s", *auction.LotNumber)
}

// finalizeAuctionAsWinner reproduces the exact persistence step
// AuctionScheduler.setAuctionWinner runs on a real 1-minute tick (see
// auction_scheduler.go): opens a transaction, calls auctionRepo.SetWinner
// (winner_id/winning_bid_id/status='ended'), commits. Used here instead of
// running the full scheduler (which polls on its own ticker and would make
// this test slow/flaky) to exercise the identical repository call the
// scheduler makes once a real auction's end_time has passed.
func finalizeAuctionAsWinner(t *testing.T, env *testEnv, auctionID, winnerID, winningBidID uuid.UUID) {
	t.Helper()
	ctx := context.Background()
	tx, err := env.db.BeginTxx(ctx, nil)
	if err != nil {
		t.Fatalf("failed to begin finalize transaction: %v", err)
	}
	if err := env.auctionRepo.SetWinner(ctx, tx, auctionID, winnerID, winningBidID); err != nil {
		tx.Rollback()
		t.Fatalf("SetWinner failed: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("failed to commit finalize transaction: %v", err)
	}
}

// Customer feedback #11 final proof: the customer's actual complaint was
// "العناصر التي فزت بها" (My Winnings) staying empty even after winning --
// not merely a cosmetic image issue. This proves the real, previously-broken
// path end-to-end: a real bid -> real winner persistence (the exact
// SetWinner call the scheduler makes) -> the SAME userSvc.ListMyWinnings
// used by GET /users/me/winnings -- confirming the winner's auction is
// actually returned, a non-winner never sees it, an still-active auction is
// absent, and a no-bid ended auction is absent (mirrors the scheduler's own
// hasWinner gate, which never calls SetWinner without a real qualifying bid).
// createTestAuctionInsured duplicates createTestAuction but stamps
// insurance_policy = 'not_required' before the insert. createTestAuction
// itself leaves insurance_policy at the Go zero value (""), which
// auctionRepo.Create binds explicitly -- bypassing the column's SQL DEFAULT
// 'required' and violating chk_auctions_insurance_policy (migration 000048)
// on a local DB that has that constraint applied. This is a pre-existing gap
// in the shared fixture helper (affects ~20+ other existing tests too --
// flagged out of scope in the prior client feedback #10 round); duplicated
// locally here rather than widening scope by editing the shared helper.
func createTestAuctionInsured(t *testing.T, env *testEnv, sellerID uuid.UUID, marketISO, currencyCode string) *models.Auction {
	t.Helper()
	ctx := context.Background()
	lotNumber := "TEST-" + uuid.New().String()[:8]
	a := &models.Auction{
		ID:               uuid.New(),
		SellerID:         sellerID,
		CategoryID:       mkCategoryID(),
		TitleAr:          "مزاد اختبار فوز " + marketISO + " " + uuid.New().String()[:6],
		LotNumber:        &lotNumber,
		StartPrice:       decimal.NewFromInt(100),
		CurrentPrice:     decimal.NewFromInt(100),
		MinIncrement:     decimal.NewFromInt(10),
		InsuranceAmount:  decimal.NewFromInt(20),
		InsurancePolicy:  "not_required",
		ReservePrice:     decimal.NewFromInt(100),
		StartTime:        time.Now().Add(-1 * time.Hour),
		EndTime:          time.Now().Add(48 * time.Hour),
		Status:           "active",
		MarketCountryISO: &marketISO,
		CurrencyCode:     &currencyCode,
	}
	if err := env.auctionRepo.Create(ctx, nil, a); err != nil {
		t.Fatalf("failed to create insured fixture auction: %v", err)
	}
	if err := env.db.GetContext(ctx, a, `SELECT * FROM auctions WHERE id = $1`, a.ID); err != nil {
		t.Fatalf("failed to read back insured fixture auction: %v", err)
	}
	return a
}

func TestListMyWinnings_RealWinnerAppears_OthersExcluded(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()

	seller := createTestUser(t, env, "TEST WINNINGS SELLER")
	winner := createTestUser(t, env, "TEST WINNINGS WINNER")
	otherUser := createTestUser(t, env, "TEST WINNINGS OTHER USER")
	creditWallet(t, env, winner.ID, decimal.NewFromInt(1000))

	// (1) Won auction: a real bid is placed while still active, then the
	// auction is finalized via the exact scheduler persistence path.
	wonAuction := createTestAuctionInsured(t, env, seller.ID, "MR", "MRU")
	bid, err := env.bidSvc.PlaceBid(ctx, wonAuction.ID, winner.ID, decimal.NewFromInt(150))
	if err != nil {
		t.Fatalf("failed to place winning bid: %v", err)
	}
	finalizeAuctionAsWinner(t, env, wonAuction.ID, winner.ID, bid.ID)

	// (2) Still-active auction (never finalized): must be absent from winnings
	// even though nothing else about it distinguishes it from the won one.
	activeAuction := createTestAuctionInsured(t, env, seller.ID, "MR", "MRU")
	if _, err := env.bidSvc.PlaceBid(ctx, activeAuction.ID, winner.ID, decimal.NewFromInt(150)); err != nil {
		t.Fatalf("failed to place bid on active auction: %v", err)
	}

	// (3) Ended auction won by someone else: must be absent for `winner`.
	otherWinAuction := createTestAuctionInsured(t, env, seller.ID, "MR", "MRU")
	otherBid, err := env.bidSvc.PlaceBid(ctx, otherWinAuction.ID, otherUser.ID, decimal.NewFromInt(150))
	if err != nil {
		t.Fatalf("failed to place bid for other user's win: %v", err)
	}
	finalizeAuctionAsWinner(t, env, otherWinAuction.ID, otherUser.ID, otherBid.ID)

	// (4) No-bid ended auction: mirrors the scheduler's own else-branch
	// (auction_scheduler.go) -- status becomes 'ended' but winner_id stays
	// NULL, since hasWinner requires a real qualifying bid.
	noBidAuction := createTestAuctionInsured(t, env, seller.ID, "MR", "MRU")
	if err := env.auctionRepo.UpdateStatus(ctx, noBidAuction.ID, "ended"); err != nil {
		t.Fatalf("failed to mark no-bid auction as ended: %v", err)
	}

	// This is the exact service call behind GET /users/me/winnings
	// (internal/handlers/user_handler.go MyWinnings -> h.service.ListMyWinnings).
	winnings, err := env.userSvc.ListMyWinnings(ctx, winner.ID)
	if err != nil {
		t.Fatalf("ListMyWinnings failed: %v", err)
	}

	foundWon, foundActive, foundOtherWin, foundNoBid := false, false, false, false
	for _, a := range winnings {
		switch a.ID {
		case wonAuction.ID:
			foundWon = true
		case activeAuction.ID:
			foundActive = true
		case otherWinAuction.ID:
			foundOtherWin = true
		case noBidAuction.ID:
			foundNoBid = true
		}
	}

	if !foundWon {
		t.Fatalf("CUSTOMER COMPLAINT NOT FIXED: expected the real, backend-confirmed win (auction %s) to appear in winner %s's My Winnings, but it did not", wonAuction.ID, winner.ID)
	}
	if foundActive {
		t.Fatalf("expected still-active auction %s to be excluded from My Winnings", activeAuction.ID)
	}
	if foundOtherWin {
		t.Fatalf("expected another user's win (auction %s, winner=%s) to be excluded from %s's My Winnings", otherWinAuction.ID, otherUser.ID, winner.ID)
	}
	if foundNoBid {
		t.Fatalf("expected no-bid ended auction %s to be excluded from My Winnings", noBidAuction.ID)
	}
	t.Logf("confirmed: real win appears, active/other-user-win/no-bid auctions all correctly excluded (%d total winnings returned for winner)", len(winnings))

	// Isolation the other direction: the actual winner's win must not leak
	// into a different user's My Winnings.
	otherUserWinnings, err := env.userSvc.ListMyWinnings(ctx, otherUser.ID)
	if err != nil {
		t.Fatalf("ListMyWinnings for otherUser failed: %v", err)
	}
	for _, a := range otherUserWinnings {
		if a.ID == wonAuction.ID {
			t.Fatalf("SECURITY: winner %s's win (auction %s) leaked into otherUser %s's My Winnings", winner.ID, wonAuction.ID, otherUser.ID)
		}
	}
}

// httpGetMyWinnings calls GET /users/me/winnings/as/:userID (test-only route
// wired in setupEnv) and parses the JSON body into a slice of raw maps,
// mirroring exactly what my_winnings_page.dart's _loadWinnings() receives as
// response.data -- proves the real HTTP/JSON contract (OK() envelope +
// models.Auction marshaling), not just that the service call returns the
// right rows.
func httpGetMyWinnings(t *testing.T, env *testEnv, callerID uuid.UUID) (int, []map[string]interface{}) {
	t.Helper()
	req := httptest.NewRequest("GET", fmt.Sprintf("/users/me/winnings/as/%s", callerID), nil)
	resp, err := env.app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return resp.StatusCode, nil
	}
	var body struct {
		Success bool                     `json:"success"`
		Data    []map[string]interface{} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode my-winnings response: %v", err)
	}
	if !body.Success {
		t.Fatalf("expected success=true in my-winnings response")
	}
	return resp.StatusCode, body.Data
}

// Bug L diagnosis: reproduces the exact real-device report end-to-end
// through the ACTUAL HTTP handler (UserHandler.MyWinnings -> OK() -> JSON),
// not just userSvc.ListMyWinnings directly. A single real bid on a fresh
// auction, finalized via the exact scheduler persistence path
// (finalizeAuctionAsWinner == repository.SetWinner, the same call
// FinalizeExpiredAuction makes), then asserts the winner's auction appears
// in the raw decoded JSON with the exact fields my_winnings_page.dart reads
// (id, winner_id, current_price, end_time, status) all present and correct
// -- proving there is no wire-format/marshaling defect between the DB and
// the mobile parser.
func TestMyWinningsHTTP_RealWinAppearsInJSONResponse(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()

	seller := createTestUser(t, env, "TEST WINNINGS HTTP SELLER")
	winner := createTestUser(t, env, "TEST WINNINGS HTTP WINNER")
	creditWallet(t, env, winner.ID, decimal.NewFromInt(1000))

	auction := createTestAuctionInsured(t, env, seller.ID, "MR", "MRU")
	bid, err := env.bidSvc.PlaceBid(ctx, auction.ID, winner.ID, decimal.NewFromInt(150))
	if err != nil {
		t.Fatalf("failed to place winning bid: %v", err)
	}
	finalizeAuctionAsWinner(t, env, auction.ID, winner.ID, bid.ID)

	status, data := httpGetMyWinnings(t, env, winner.ID)
	if status != 200 {
		t.Fatalf("BUG L: GET /users/me/winnings returned HTTP %d for the real winner, expected 200", status)
	}

	var found map[string]interface{}
	for _, item := range data {
		if id, _ := item["id"].(string); id == auction.ID.String() {
			found = item
			break
		}
	}
	if found == nil {
		t.Fatalf("BUG L NOT REPRODUCED HERE: real win (auction %s) absent from the actual HTTP JSON response body; raw response had %d items: %+v", auction.ID, len(data), data)
	}

	if winnerID, _ := found["winner_id"].(string); winnerID != winner.ID.String() {
		t.Fatalf("expected winner_id=%s in JSON response, got %v", winner.ID, found["winner_id"])
	}
	if status, _ := found["status"].(string); status != "ended" {
		t.Fatalf("expected status=ended in JSON response, got %v", found["status"])
	}
	if _, ok := found["current_price"]; !ok {
		t.Fatalf("expected current_price field present in JSON response, mobile reads this for the winning amount")
	}
	if _, ok := found["end_time"]; !ok {
		t.Fatalf("expected end_time field present in JSON response, mobile reads this for the win date")
	}
	t.Logf("confirmed: real HTTP JSON response for GET /users/me/winnings contains the winner's auction with all fields mobile depends on (id=%s, winner_id=%v, status=%v)", auction.ID, found["winner_id"], found["status"])
}

// Customer feedback #12 (restore Active/Ended auctions selector): the mobile
// "Auction Types" page (all_auctions_page.dart) already had its Active/Ended
// status tabs restored (client feedback A12) and correctly requests
// GET /auctions?status=ended for the Ended tab. But AuctionRepository.FindAll
// unconditionally appended "AND end_time > NOW()" to every query regardless
// of the requested status -- harmless when this line was written (predating
// SetWinner ever being called anywhere), but since an ended auction's
// end_time is necessarily in the past by definition, this silently excluded
// EVERY ended auction from a status='ended' request, making the Ended tab
// permanently empty. This proves both fixes: status='ended' now returns real
// ended auctions, and the new total return value reflects the true count
// (not just the current page's length).
func TestFindAllAuctions_EndedStatusReturnsRealEndedAuctions(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()

	seller := createTestUser(t, env, "TEST A12 SELLER")
	winner := createTestUser(t, env, "TEST A12 WINNER")

	activeAuction := createTestAuctionInsured(t, env, seller.ID, "MR", "MRU")
	endedAuction := createTestAuctionInsured(t, env, seller.ID, "MR", "MRU")
	bid, err := env.bidSvc.PlaceBid(ctx, endedAuction.ID, winner.ID, decimal.NewFromInt(150))
	if err != nil {
		t.Fatalf("failed to place bid on soon-to-end auction: %v", err)
	}
	finalizeAuctionAsWinner(t, env, endedAuction.ID, winner.ID, bid.ID)

	// Before the fix, this returned zero results and total=0 no matter how
	// many auctions were actually ended -- "AND end_time > NOW()" excluded
	// endedAuction's row unconditionally, since its end_time is in the past.
	endedResults, endedTotal, err := env.auctSvc.List(ctx, repository.AuctionFilters{
		Status:           "ended",
		MarketCountryISO: "MR",
		Page:             1,
		PerPage:          25,
	})
	if err != nil {
		t.Fatalf("List(status=ended) failed: %v", err)
	}

	foundEnded := false
	for _, a := range endedResults {
		if a.ID == endedAuction.ID {
			foundEnded = true
		}
		if a.ID == activeAuction.ID {
			t.Fatalf("CUSTOMER COMPLAINT NOT FIXED: still-active auction %s appeared in the Ended tab's results", activeAuction.ID)
		}
	}
	if !foundEnded {
		t.Fatalf("CUSTOMER COMPLAINT NOT FIXED: real ended auction %s did not appear in status='ended' results (Ended tab would stay permanently empty)", endedAuction.ID)
	}
	if endedTotal < 1 {
		t.Fatalf("expected a real total >= 1 for status='ended', got %d", endedTotal)
	}

	// The active tab's own filtering (end_time > NOW() still applies) and its
	// total must be unaffected by this fix.
	activeResults, activeTotal, err := env.auctSvc.List(ctx, repository.AuctionFilters{
		Status:           "active",
		MarketCountryISO: "MR",
		Page:             1,
		PerPage:          25,
	})
	if err != nil {
		t.Fatalf("List(status=active) failed: %v", err)
	}
	foundActive := false
	for _, a := range activeResults {
		if a.ID == activeAuction.ID {
			foundActive = true
		}
		if a.ID == endedAuction.ID {
			t.Fatalf("ended auction %s incorrectly appeared in the Active tab's results", endedAuction.ID)
		}
	}
	if !foundActive {
		t.Fatalf("expected still-active auction %s to appear in the Active tab's results", activeAuction.ID)
	}
	if activeTotal < 1 {
		t.Fatalf("expected a real total >= 1 for status='active', got %d", activeTotal)
	}
	t.Logf("confirmed: Ended tab total=%d (includes real ended auction), Active tab total=%d, no cross-contamination", endedTotal, activeTotal)
}

// createTestCategory inserts a temporary category with the given fee_tier
// directly (client feedback #4) -- simpler than exercising the full admin
// HTTP/auth stack just to set one field on a test fixture. Cleaned up via
// t.Cleanup so it never lingers in the shared local test DB.
func createTestCategory(t *testing.T, env *testEnv, feeTier string) int {
	t.Helper()
	ctx := context.Background()
	var id int
	err := env.db.GetContext(ctx, &id, `
		INSERT INTO categories (name_ar, name_fr, fee_tier)
		VALUES ($1, $2, $3) RETURNING id`,
		"فئة اختبار "+uuid.New().String()[:6], "Test category", feeTier)
	if err != nil {
		t.Fatalf("failed to create test category: %v", err)
	}
	t.Cleanup(func() {
		env.db.ExecContext(context.Background(), `DELETE FROM categories WHERE id = $1`, id)
	})
	return id
}

// Client feedback #4 (100/500 MRU subscription fee): proves the real,
// end-to-end path -- CreateAuctionRequest stamps SubscriptionFee from the
// request's actual category (never client-supplied, never guessed from name/
// id/icon_name), a standard category yields 100, a premium category yields
// 500, and changing a category's fee_tier afterward does NOT retroactively
// change an already-created request's stamped fee (immutable once stamped,
// same principle as CurrencyCode/MarketCountryISO).
func TestCreateAuctionRequest_SubscriptionFeeFromCategory(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST FEE TIER USER")

	standardCategoryID := createTestCategory(t, env, models.FeeTierStandard)
	premiumCategoryID := createTestCategory(t, env, models.FeeTierPremium)

	standardReq := newAuctionRequest(user.ID, "fee-standard-"+uuid.New().String()[:6])
	standardReq.CategoryID = standardCategoryID
	if err := env.reqSvc.CreateAuctionRequest(ctx, standardReq); err != nil {
		t.Fatalf("CreateAuctionRequest (standard category) failed: %v", err)
	}
	if !standardReq.SubscriptionFee.Equal(models.StandardSubscriptionFee) {
		t.Fatalf("expected standard category to stamp fee %s, got %s", models.StandardSubscriptionFee, standardReq.SubscriptionFee)
	}

	premiumReq := newAuctionRequest(user.ID, "fee-premium-"+uuid.New().String()[:6])
	premiumReq.CategoryID = premiumCategoryID
	if err := env.reqSvc.CreateAuctionRequest(ctx, premiumReq); err != nil {
		t.Fatalf("CreateAuctionRequest (premium category) failed: %v", err)
	}
	if !premiumReq.SubscriptionFee.Equal(models.PremiumSubscriptionFee) {
		t.Fatalf("expected premium category to stamp fee %s, got %s", models.PremiumSubscriptionFee, premiumReq.SubscriptionFee)
	}

	// Read back from the DB via the real repository path (not just the
	// in-memory struct) to prove the Scan-order fix actually persists and
	// reloads the value correctly.
	reloaded, err := env.reqRepo.GetAuctionRequestByID(ctx, premiumReq.ID)
	if err != nil {
		t.Fatalf("GetAuctionRequestByID failed: %v", err)
	}
	if !reloaded.SubscriptionFee.Equal(models.PremiumSubscriptionFee) {
		t.Fatalf("expected reloaded premium request to keep fee %s, got %s", models.PremiumSubscriptionFee, reloaded.SubscriptionFee)
	}

	// Changing the category's fee_tier afterward must NOT retroactively
	// change the already-stamped request.
	if _, err := env.db.ExecContext(ctx, `UPDATE categories SET fee_tier = $1 WHERE id = $2`, models.FeeTierStandard, premiumCategoryID); err != nil {
		t.Fatalf("failed to flip category fee_tier: %v", err)
	}
	reloadedAfterCategoryChange, err := env.reqRepo.GetAuctionRequestByID(ctx, premiumReq.ID)
	if err != nil {
		t.Fatalf("GetAuctionRequestByID (after category change) failed: %v", err)
	}
	if !reloadedAfterCategoryChange.SubscriptionFee.Equal(models.PremiumSubscriptionFee) {
		t.Fatalf("SECURITY/CORRECTNESS: expected already-created request to keep its stamped fee %s even after the category's fee_tier changed, got %s", models.PremiumSubscriptionFee, reloadedAfterCategoryChange.SubscriptionFee)
	}
	t.Logf("confirmed: standard=%s premium=%s stamped correctly, and an already-created request's fee survives a later category fee_tier change", standardReq.SubscriptionFee, premiumReq.SubscriptionFee)
}

// Real-device Staging bug B: AuctionService.Create (the direct
// POST /v1/api/auctions path, distinct from AuctionRequest -> admin review)
// built its models.Auction{} literal without setting InsurancePolicy at all,
// so Go's zero value ("") was sent to an INSERT that explicitly includes
// insurance_policy -- violating chk_auctions_insurance_policy (migration
// 000048's CHECK only allows 'required'/'not_required') and failing with a
// real-device HTTP 500 for every direct auction creation. Proven against a
// real Postgres because the defect only manifested at DB constraint time,
// never in application-level validation.
func TestAuctionServiceCreate_StampsInsurancePolicyRequired(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()

	// auctions_lot_number_seq is deliberately left at small values (e.g. 99,
	// 100) by the TestLotNumber_* fixtures above via a non-transactional
	// setval() -- Postgres sequences are NOT rolled back with the
	// transaction that calls setval(), so those tests' resets persist across
	// this whole test binary run regardless of test order. Advance it past
	// every existing LOT-* row before relying on the trigger's auto-generated
	// lot_number here, so this test's own real bug (missing InsurancePolicy)
	// isn't masked by an unrelated lot_number collision.
	var maxLot int64
	_ = env.db.GetContext(ctx, &maxLot, `
		SELECT COALESCE(MAX(CAST(SUBSTRING(lot_number FROM 5) AS BIGINT)), 0)
		FROM auctions WHERE lot_number LIKE 'LOT-%' AND SUBSTRING(lot_number FROM 5) ~ '^[0-9]+$'`)
	if _, err := env.db.ExecContext(ctx, `SELECT setval('auctions_lot_number_seq', $1)`, maxLot+1000); err != nil {
		t.Fatalf("failed to advance auctions_lot_number_seq past existing rows: %v", err)
	}

	seller := createTestUser(t, env, "TEST DIRECT CREATE SELLER")
	categoryID := createTestCategory(t, env, models.FeeTierStandard)

	input := services.CreateAuctionInput{
		CategoryID:    categoryID,
		TitleAr:       "مزاد اختبار الإنشاء المباشر " + uuid.New().String()[:6],
		DescriptionAr: "STAGING_TEST direct auction creation",
		StartPrice:    decimal.NewFromInt(1000),
		MinIncrement:  decimal.NewFromInt(100),
		// Multi-day duration (client feedback #15) must remain preserved by
		// this fix -- end_time is passed straight through, not truncated.
		EndTime:  time.Now().Add(72 * time.Hour),
		Quantity: 1,
	}

	auction, err := env.auctSvc.Create(ctx, seller.ID, input)
	if err != nil {
		t.Fatalf("AuctionService.Create failed (this is exactly the real-device 500: %v)", err)
	}

	if auction.InsurancePolicy != models.InsurancePolicyRequired {
		t.Fatalf("expected a directly-created auction to be stamped InsurancePolicy=%q, got %q",
			models.InsurancePolicyRequired, auction.InsurancePolicy)
	}

	// Read back via the real repository path (not just the in-memory struct)
	// to prove the value actually persisted through the CHECK constraint,
	// not merely held in the Go struct.
	reloaded, _, err := env.auctSvc.GetByID(ctx, auction.ID)
	if err != nil {
		t.Fatalf("GetByID failed to reload the created auction: %v", err)
	}
	if reloaded.InsurancePolicy != models.InsurancePolicyRequired {
		t.Fatalf("expected reloaded auction to keep InsurancePolicy=%q, got %q",
			models.InsurancePolicyRequired, reloaded.InsurancePolicy)
	}
	if !reloaded.InsuranceRequired() {
		t.Fatalf("InsuranceRequired() must be true for a freshly-created direct auction")
	}

	// Multi-day duration (client feedback #15) regression check: end_time
	// must be preserved exactly, not clamped to 24h.
	if !reloaded.EndTime.After(reloaded.StartTime.Add(47 * time.Hour)) {
		t.Fatalf("expected multi-day end_time to be preserved (~72h), got start=%v end=%v", reloaded.StartTime, reloaded.EndTime)
	}

	t.Logf("confirmed: direct auction creation stamps InsurancePolicy=%q and satisfies chk_auctions_insurance_policy; multi-day duration preserved", reloaded.InsurancePolicy)
}

// Client feedback #4: editing an existing draft/rejected request's category
// must re-stamp the fee from the NEW category -- otherwise a user could
// create a request under a "standard" category (100 MRU) then edit it to a
// "premium" category (cars/real estate) while keeping the cheaper stamped
// fee.
func TestUpdateAuctionRequest_CategoryChangeRestampsSubscriptionFee(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST FEE RESTAMP USER")

	standardCategoryID := createTestCategory(t, env, models.FeeTierStandard)
	premiumCategoryID := createTestCategory(t, env, models.FeeTierPremium)

	req := newAuctionRequest(user.ID, "fee-restamp-"+uuid.New().String()[:6])
	req.CategoryID = standardCategoryID
	req.Status = "draft"
	if err := env.reqSvc.CreateAuctionRequest(ctx, req); err != nil {
		t.Fatalf("CreateAuctionRequest failed: %v", err)
	}
	if !req.SubscriptionFee.Equal(models.StandardSubscriptionFee) {
		t.Fatalf("expected initial standard fee %s, got %s", models.StandardSubscriptionFee, req.SubscriptionFee)
	}

	updates := *req
	updates.CategoryID = premiumCategoryID
	if err := env.reqSvc.UpdateAuctionRequest(ctx, req.ID, user.ID, &updates); err != nil {
		t.Fatalf("UpdateAuctionRequest (category change) failed: %v", err)
	}

	reloaded, err := env.reqRepo.GetAuctionRequestByID(ctx, req.ID)
	if err != nil {
		t.Fatalf("GetAuctionRequestByID failed: %v", err)
	}
	if !reloaded.SubscriptionFee.Equal(models.PremiumSubscriptionFee) {
		t.Fatalf("SECURITY: expected fee to be re-stamped to premium %s after changing to a premium category, got %s (user could underpay)", models.PremiumSubscriptionFee, reloaded.SubscriptionFee)
	}
	t.Logf("confirmed: editing a request's category from standard to premium correctly re-stamps the fee to %s", reloaded.SubscriptionFee)
}

// Client feedback #4 financial-integrity round: proves the real, previously-
// missing enforcement -- WalletService.InitiateDeposit, when given a real
// auction_request_id, IGNORES the client-supplied amount and substitutes the
// request's own server-stamped subscription_fee (auction_requests.
// subscription_fee), never recalculating it from category name/id/icon_name.
// A generic wallet deposit (no auction_request_id) is completely unaffected.
func TestInitiateDeposit_AuctionSubscriptionFee_ClientAmountIgnored(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	walletSvc := newWalletSvc(env)

	user := createTestUser(t, env, "TEST FEE DEPOSIT USER")
	standardCategoryID := createTestCategory(t, env, models.FeeTierStandard)
	premiumCategoryID := createTestCategory(t, env, models.FeeTierPremium)

	standardReq := newAuctionRequest(user.ID, "deposit-standard-"+uuid.New().String()[:6])
	standardReq.CategoryID = standardCategoryID
	if err := env.reqSvc.CreateAuctionRequest(ctx, standardReq); err != nil {
		t.Fatalf("CreateAuctionRequest (standard) failed: %v", err)
	}

	premiumReq := newAuctionRequest(user.ID, "deposit-premium-"+uuid.New().String()[:6])
	premiumReq.CategoryID = premiumCategoryID
	if err := env.reqSvc.CreateAuctionRequest(ctx, premiumReq); err != nil {
		t.Fatalf("CreateAuctionRequest (premium) failed: %v", err)
	}

	t.Run("standard request: deposit amount matches the expected 100 MRU fee", func(t *testing.T) {
		txn, err := walletSvc.InitiateDeposit(ctx, user.ID, decimal.NewFromInt(100), "bankily", "bankily", "", &standardReq.ID)
		if err != nil {
			t.Fatalf("InitiateDeposit (standard, correct amount) failed: %v", err)
		}
		if !txn.Amount.Equal(models.StandardSubscriptionFee) {
			t.Fatalf("expected deposit amount %s, got %s", models.StandardSubscriptionFee, txn.Amount)
		}
	})

	t.Run("premium request: deposit amount matches the expected 500 MRU fee", func(t *testing.T) {
		txn, err := walletSvc.InitiateDeposit(ctx, user.ID, decimal.NewFromInt(500), "bankily", "bankily", "", &premiumReq.ID)
		if err != nil {
			t.Fatalf("InitiateDeposit (premium, correct amount) failed: %v", err)
		}
		if !txn.Amount.Equal(models.PremiumSubscriptionFee) {
			t.Fatalf("expected deposit amount %s, got %s", models.PremiumSubscriptionFee, txn.Amount)
		}
	})

	t.Run("premium request + client sends 100 -- cannot underpay, backend overrides to 500", func(t *testing.T) {
		// Fresh premium request so this test's deposit doesn't collide with
		// the earlier subtest's non-rejected deposit for the same reference.
		req := newAuctionRequest(user.ID, "deposit-tamper-100-"+uuid.New().String()[:6])
		req.CategoryID = premiumCategoryID
		if err := env.reqSvc.CreateAuctionRequest(ctx, req); err != nil {
			t.Fatalf("CreateAuctionRequest failed: %v", err)
		}
		txn, err := walletSvc.InitiateDeposit(ctx, user.ID, decimal.NewFromInt(100), "bankily", "bankily", "", &req.ID)
		if err != nil {
			t.Fatalf("InitiateDeposit (tampered 100) failed: %v", err)
		}
		if !txn.Amount.Equal(models.PremiumSubscriptionFee) {
			t.Fatalf("SECURITY: client sent 100 for a premium (500) request and it was NOT overridden -- got stored amount %s", txn.Amount)
		}
	})

	t.Run("premium request + client sends 1 -- cannot underpay", func(t *testing.T) {
		req := newAuctionRequest(user.ID, "deposit-tamper-1-"+uuid.New().String()[:6])
		req.CategoryID = premiumCategoryID
		if err := env.reqSvc.CreateAuctionRequest(ctx, req); err != nil {
			t.Fatalf("CreateAuctionRequest failed: %v", err)
		}
		txn, err := walletSvc.InitiateDeposit(ctx, user.ID, decimal.NewFromInt(1), "bankily", "bankily", "", &req.ID)
		if err != nil {
			t.Fatalf("InitiateDeposit (tampered 1) failed: %v", err)
		}
		if !txn.Amount.Equal(models.PremiumSubscriptionFee) {
			t.Fatalf("SECURITY: client sent 1 for a premium (500) request and it was NOT overridden -- got stored amount %s", txn.Amount)
		}
	})

	t.Run("another user's auction_request_id is rejected (ownership enforced)", func(t *testing.T) {
		otherUser := createTestUser(t, env, "TEST FEE DEPOSIT OTHER USER")
		req := newAuctionRequest(user.ID, "deposit-ownership-"+uuid.New().String()[:6])
		req.CategoryID = standardCategoryID
		if err := env.reqSvc.CreateAuctionRequest(ctx, req); err != nil {
			t.Fatalf("CreateAuctionRequest failed: %v", err)
		}
		_, err := walletSvc.InitiateDeposit(ctx, otherUser.ID, decimal.NewFromInt(100), "bankily", "bankily", "", &req.ID)
		if err != apperr.ErrForbidden {
			t.Fatalf("expected ErrForbidden when a different user references someone else's request, got %v", err)
		}
	})

	t.Run("a nonexistent auction_request_id is rejected", func(t *testing.T) {
		bogusID := uuid.New()
		_, err := walletSvc.InitiateDeposit(ctx, user.ID, decimal.NewFromInt(100), "bankily", "bankily", "", &bogusID)
		if err != apperr.ErrNotFound {
			t.Fatalf("expected ErrNotFound for a nonexistent auction_request_id, got %v", err)
		}
	})

	t.Run("a duplicate non-rejected deposit for the same request is rejected", func(t *testing.T) {
		req := newAuctionRequest(user.ID, "deposit-duplicate-"+uuid.New().String()[:6])
		req.CategoryID = standardCategoryID
		if err := env.reqSvc.CreateAuctionRequest(ctx, req); err != nil {
			t.Fatalf("CreateAuctionRequest failed: %v", err)
		}
		if _, err := walletSvc.InitiateDeposit(ctx, user.ID, decimal.NewFromInt(100), "bankily", "bankily", "", &req.ID); err != nil {
			t.Fatalf("first InitiateDeposit failed: %v", err)
		}
		_, err := walletSvc.InitiateDeposit(ctx, user.ID, decimal.NewFromInt(100), "bankily", "bankily", "", &req.ID)
		if err != apperr.ErrConflict {
			t.Fatalf("expected ErrConflict for a duplicate non-rejected deposit on the same request, got %v", err)
		}
	})

	t.Run("generic wallet deposit (no auction_request_id) is completely unaffected", func(t *testing.T) {
		txn, err := walletSvc.InitiateDeposit(ctx, user.ID, decimal.NewFromInt(250), "bankily", "bankily", "", nil)
		if err != nil {
			t.Fatalf("InitiateDeposit (generic) failed: %v", err)
		}
		if !txn.Amount.Equal(decimal.NewFromInt(250)) {
			t.Fatalf("expected generic deposit to keep the client-supplied amount 250, got %s (regression in unrelated wallet top-up flow)", txn.Amount)
		}
		if txn.Reference != nil {
			t.Fatalf("expected a generic deposit to have no reference, got %v", *txn.Reference)
		}
	})
}

// Client feedback #4 (category create/edit consistency round): proves
// AuctionRepository.CreateCategory actually persists an admin-selected
// fee_tier through a real INSERT + read-back, not just in the in-memory
// struct returned by the call. Before this round CreateCategory's INSERT
// omitted fee_tier entirely, so a category created as "premium" would
// silently persist as the column's DB DEFAULT ('standard') instead.
func TestCreateCategory_PersistsFeeTier(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()

	t.Run("omitted fee_tier persists as standard", func(t *testing.T) {
		cat := &models.Category{NameAr: "فئة اختبار بدون تحديد " + uuid.New().String()[:6], NameFr: "Test no tier"}
		if err := env.auctionRepo.CreateCategory(ctx, cat); err != nil {
			t.Fatalf("CreateCategory failed: %v", err)
		}
		t.Cleanup(func() { env.db.ExecContext(context.Background(), `DELETE FROM categories WHERE id = $1`, cat.ID) })

		if cat.FeeTier != models.FeeTierStandard {
			t.Fatalf("expected in-memory FeeTier to be 'standard' after create, got %q", cat.FeeTier)
		}
		reloaded, err := env.auctionRepo.GetCategoryByID(ctx, cat.ID)
		if err != nil {
			t.Fatalf("GetCategoryByID failed: %v", err)
		}
		if reloaded.FeeTier != models.FeeTierStandard {
			t.Fatalf("expected persisted fee_tier='standard', got %q", reloaded.FeeTier)
		}
	})

	t.Run("explicit fee_tier=standard persists as standard", func(t *testing.T) {
		cat := &models.Category{NameAr: "فئة اختبار عادية " + uuid.New().String()[:6], NameFr: "Test standard", FeeTier: models.FeeTierStandard}
		if err := env.auctionRepo.CreateCategory(ctx, cat); err != nil {
			t.Fatalf("CreateCategory failed: %v", err)
		}
		t.Cleanup(func() { env.db.ExecContext(context.Background(), `DELETE FROM categories WHERE id = $1`, cat.ID) })

		reloaded, err := env.auctionRepo.GetCategoryByID(ctx, cat.ID)
		if err != nil {
			t.Fatalf("GetCategoryByID failed: %v", err)
		}
		if reloaded.FeeTier != models.FeeTierStandard {
			t.Fatalf("expected persisted fee_tier='standard', got %q", reloaded.FeeTier)
		}
	})

	t.Run("explicit fee_tier=premium survives real DB persistence -- the actual bug this round fixes", func(t *testing.T) {
		cat := &models.Category{NameAr: "فئة اختبار مميزة " + uuid.New().String()[:6], NameFr: "Test premium", FeeTier: models.FeeTierPremium}
		if err := env.auctionRepo.CreateCategory(ctx, cat); err != nil {
			t.Fatalf("CreateCategory failed: %v", err)
		}
		t.Cleanup(func() { env.db.ExecContext(context.Background(), `DELETE FROM categories WHERE id = $1`, cat.ID) })

		if cat.FeeTier != models.FeeTierPremium {
			t.Fatalf("expected in-memory FeeTier to be 'premium' after create, got %q", cat.FeeTier)
		}
		reloaded, err := env.auctionRepo.GetCategoryByID(ctx, cat.ID)
		if err != nil {
			t.Fatalf("GetCategoryByID failed: %v", err)
		}
		if reloaded.FeeTier != models.FeeTierPremium {
			t.Fatalf("CATEGORY CREATE/EDIT INCONSISTENCY NOT FIXED: admin selected 'premium' at creation, but the DB persisted %q instead", reloaded.FeeTier)
		}
		// And the category's SubscriptionFee() helper correctly reflects it,
		// proving a request filed under this category would be charged 500.
		if !reloaded.SubscriptionFee().Equal(models.PremiumSubscriptionFee) {
			t.Fatalf("expected a persisted premium category to yield SubscriptionFee()=%s, got %s", models.PremiumSubscriptionFee, reloaded.SubscriptionFee())
		}
	})

	// Invalid fee_tier rejection is a service-layer (adminService.
	// validateFeeTierInput) concern, unreachable from this package -- covered
	// by internal/services/category_fee_tier_test.go's TestValidateFeeTierInput,
	// which proves both CreateCategory and UpdateCategory share the identical
	// accept/reject rule.
}

// Client feedback #15 (allow auction duration longer than 24 hours): the
// backend's own 24h maximum was already removed in a prior round ("client
// feedback Phase B item 15", auction_handler.go/AuctionService.Create) --
// this proves the real repository/scheduler layer never had (or reintroduced)
// any hidden 24h assumption of its own. A multi-day (7-day) auction must:
// (1) be creatable with its true end_time persisted exactly, (2) NOT be
// matched by the scheduler's FindEndedSince query while still genuinely
// active, even well past the old 24h mark, and (3) be matched once its real
// end_time has actually passed.
func TestMultiDayAuction_DurationAndSchedulerSafety(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST MULTIDAY SELLER")

	newInsuredAuction := func(start, end time.Time) *models.Auction {
		lotNumber := "TEST-" + uuid.New().String()[:8]
		marketISO := "MR"
		currencyCode := "MRU"
		a := &models.Auction{
			ID:               uuid.New(),
			SellerID:         seller.ID,
			CategoryID:       mkCategoryID(),
			TitleAr:          "مزاد اختبار مدة طويلة " + uuid.New().String()[:6],
			LotNumber:        &lotNumber,
			StartPrice:       decimal.NewFromInt(100),
			CurrentPrice:     decimal.NewFromInt(100),
			MinIncrement:     decimal.NewFromInt(10),
			InsuranceAmount:  decimal.NewFromInt(20),
			InsurancePolicy:  "not_required",
			ReservePrice:     decimal.NewFromInt(100),
			StartTime:        start,
			EndTime:          end,
			Status:           "active",
			MarketCountryISO: &marketISO,
			CurrencyCode:     &currencyCode,
		}
		if err := env.auctionRepo.Create(ctx, nil, a); err != nil {
			t.Fatalf("failed to create multi-day fixture auction: %v", err)
		}
		return a
	}

	t.Run("a 7-day auction persists its true end_time, not truncated to 24h", func(t *testing.T) {
		start := time.Now().Add(1 * time.Hour)
		end := start.Add(7 * 24 * time.Hour)
		a := newInsuredAuction(start, end)

		reloaded, err := env.auctionRepo.FindByID(ctx, a.ID)
		if err != nil {
			t.Fatalf("FindByID failed: %v", err)
		}
		gotDuration := reloaded.EndTime.Sub(reloaded.StartTime)
		wantDuration := 7 * 24 * time.Hour
		if gotDuration < wantDuration-time.Second || gotDuration > wantDuration+time.Second {
			t.Fatalf("expected persisted duration ~%s, got %s (end_time was truncated somewhere)", wantDuration, gotDuration)
		}
	})

	t.Run("a still-active 7-day auction is NOT matched by the scheduler's ended-auction query, even well past 24h from its start", func(t *testing.T) {
		start := time.Now().Add(-30 * time.Hour) // started 30h ago -- past the old 24h mark
		end := start.Add(7 * 24 * time.Hour)     // but genuinely still 5+ days from ending
		a := newInsuredAuction(start, end)

		ended, err := env.auctionRepo.FindEndedSince(ctx, time.Now().Add(-1*time.Hour))
		if err != nil {
			t.Fatalf("FindEndedSince failed: %v", err)
		}
		for _, e := range ended {
			if e.ID == a.ID {
				t.Fatalf("SCHEDULER BUG: a still-active 7-day auction (ends in ~5 more days) was incorrectly matched as ended just because it started >24h ago")
			}
		}
	})

	t.Run("an auction whose true (multi-day-scheduled) end_time has now actually passed IS matched by the scheduler", func(t *testing.T) {
		start := time.Now().Add(-73 * time.Hour) // started 73h ago
		end := start.Add(72 * time.Hour)         // a real 72h auction -- ended 1h ago
		a := newInsuredAuction(start, end)

		ended, err := env.auctionRepo.FindEndedSince(ctx, time.Now().Add(-2*time.Hour))
		if err != nil {
			t.Fatalf("FindEndedSince failed: %v", err)
		}
		found := false
		for _, e := range ended {
			if e.ID == a.ID {
				found = true
			}
		}
		if !found {
			t.Fatalf("expected a genuinely-ended 72h auction to be matched by FindEndedSince, but it was not")
		}
	})
}

// Client feedback #16 (admin global/broadcast notifications). The existing
// notification_broadcast_test.go proves SendBroadcast's logic against a fake
// in-memory repo -- this proves the SAME service method against a REAL local
// Postgres, closing the gap between "logic is correct in isolation" and "it
// actually persists end-to-end the way the customer's real deployment would
// see it". Exercises exactly the path the HTTP handler
// (NotificationHandler.SendNotification, POST /admin/notifications/send)
// calls: notifSvc.SendBroadcast -> real INSERT per active user -> real
// SELECT via notifSvc.ListNotifications (the same call GET /notifications
// makes for a user's in-app list).
func TestSendBroadcast_RealDatabase_PersistsForEveryActiveUser(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()

	userA := createTestUser(t, env, "TEST BROADCAST USER A")
	userB := createTestUser(t, env, "TEST BROADCAST USER B")
	userC := createTestUser(t, env, "TEST BROADCAST USER C")

	title := "إعلان تجريبي " + uuid.New().String()[:6]
	body := "نص الإشعار كما كتبه الأدمن بالضبط، بدون أي ترجمة أو تعديل تلقائي."

	targetUsers, sent, failed, err := env.notifSvc.SendBroadcast(ctx, title, body, "general", nil)
	if err != nil {
		t.Fatalf("SendBroadcast failed: %v", err)
	}
	if targetUsers < 3 {
		t.Fatalf("expected at least the 3 fixture users to be counted as targets, got %d", targetUsers)
	}
	// No push tokens are registered for any fixture user in this test (no
	// FCM configured in setupEnv either -- notifSvc built with fcm=nil), so
	// every SendPush call takes the "FCM not configured" early-return path
	// and reports success (err == nil) -- persistence must never depend on
	// push succeeding.
	if sent != targetUsers {
		t.Fatalf("expected all %d target users to be reported as sent (persistence never depends on FCM), got sent=%d failed=%d", targetUsers, sent, failed)
	}

	for _, u := range []*models.User{userA, userB, userC} {
		notifs, err := env.notifSvc.ListNotifications(ctx, u.ID, 50)
		if err != nil {
			t.Fatalf("ListNotifications for %s failed: %v", u.ID, err)
		}
		found := false
		for _, n := range notifs {
			if n.Title == title {
				found = true
				if n.Body == nil || *n.Body != body {
					t.Fatalf("expected persisted body to exactly match the admin's literal text, got %v", n.Body)
				}
				if n.Type != "general" {
					t.Fatalf("expected persisted type='general', got %q", n.Type)
				}
				if n.IsRead {
					t.Fatalf("expected a freshly broadcast notification to start unread")
				}
			}
		}
		if !found {
			t.Fatalf("CUSTOMER COMPLAINT: broadcast notification did not appear in user %s's in-app notification list (GET /notifications source)", u.ID)
		}
	}
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
	return services.NewWalletService(env.db, env.walletRepo, repository.NewTransactionRepository(env.db, env.walletRepo), env.reqRepo, env.userRepo, env.notifSvc, nil, env.logger)
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
		InsurancePolicy:  models.InsurancePolicyRequired,
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

// ==================================================
// Client feedback #19 hardening: bidder_name/bidder_phone are legacy
// denormalized columns on bids that are frequently NULL. models.Bid scans
// them as non-nullable Go string, so any raw `SELECT *` (or explicit
// column list without COALESCE) into models.Bid failed whenever either was
// NULL -- this was silently discarded in some callers (bid_service.go's
// prevTopBid, _ := FindTopBid(...)) but became load-bearing for
// GetBidStatus/GetUserHighestBid, which the new has_bid field depends on
// for every real bid. Fixed via COALESCE(..., '') in bid_repo.go's
// FindByAuctionID/FindTopBid/FindUserBidOnAuction and auction_repo.go's
// GetUserHighestBid -- no schema change, no model change.
// ==================================================

func TestGetUserHighestBid_NullBidderName_Safe(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST NULLBID SELLER A")
	user := createTestUser(t, env, "TEST NULLBID USER A")
	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")
	creditWallet(t, env, user.ID, decimal.NewFromInt(1000))

	if _, err := env.bidSvc.PlaceBid(ctx, auction.ID, user.ID, decimal.NewFromInt(110)); err != nil {
		t.Fatalf("expected bid to succeed, got: %v", err)
	}
	// Confirm the real row actually has NULL bidder_name (the normal case
	// for a bid placed through the service, not a contrived fixture).
	var bidderName sql.NullString
	if err := env.db.GetContext(ctx, &bidderName,
		`SELECT bidder_name FROM bids WHERE auction_id = $1 AND user_id = $2`, auction.ID, user.ID); err != nil {
		t.Fatalf("failed to read back bid: %v", err)
	}
	if bidderName.Valid {
		t.Fatalf("expected bidder_name to be NULL for this fixture, got %q", bidderName.String)
	}

	bid, err := env.auctionRepo.GetUserHighestBid(ctx, auction.ID, user.ID)
	if err != nil {
		t.Fatalf("GetUserHighestBid must not fail on NULL bidder_name, got: %v", err)
	}
	if bid.BidderName != "" {
		t.Fatalf("expected BidderName to map to empty string for NULL, got %q", bid.BidderName)
	}
	t.Logf("confirmed: GetUserHighestBid handles NULL bidder_name safely")
}

func TestGetUserHighestBid_NullBidderPhone_Safe(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST NULLBID SELLER B")
	user := createTestUser(t, env, "TEST NULLBID USER B")
	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")
	creditWallet(t, env, user.ID, decimal.NewFromInt(1000))

	if _, err := env.bidSvc.PlaceBid(ctx, auction.ID, user.ID, decimal.NewFromInt(110)); err != nil {
		t.Fatalf("expected bid to succeed, got: %v", err)
	}

	bid, err := env.auctionRepo.GetUserHighestBid(ctx, auction.ID, user.ID)
	if err != nil {
		t.Fatalf("GetUserHighestBid must not fail on NULL bidder_phone, got: %v", err)
	}
	if bid.BidderPhone != "" {
		t.Fatalf("expected BidderPhone to map to empty string for NULL, got %q", bid.BidderPhone)
	}
	t.Logf("confirmed: GetUserHighestBid handles NULL bidder_phone safely")
}

func TestGetUserHighestBid_BothBidderFieldsNull_Safe(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST NULLBID SELLER C")
	user := createTestUser(t, env, "TEST NULLBID USER C")
	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")
	creditWallet(t, env, user.ID, decimal.NewFromInt(1000))

	if _, err := env.bidSvc.PlaceBid(ctx, auction.ID, user.ID, decimal.NewFromInt(110)); err != nil {
		t.Fatalf("expected bid to succeed, got: %v", err)
	}

	bid, err := env.auctionRepo.GetUserHighestBid(ctx, auction.ID, user.ID)
	if err != nil {
		t.Fatalf("GetUserHighestBid must not fail when both bidder fields are NULL, got: %v", err)
	}
	if bid.BidderName != "" || bid.BidderPhone != "" {
		t.Fatalf("expected both BidderName and BidderPhone empty, got %q / %q", bid.BidderName, bid.BidderPhone)
	}
	t.Logf("confirmed: GetUserHighestBid handles both bidder fields NULL safely")
}

func TestGetUserHighestBid_NonNullBidderDetails_Unchanged(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST NULLBID SELLER D")
	user := createTestUser(t, env, "TEST NULLBID USER D")
	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")
	creditWallet(t, env, user.ID, decimal.NewFromInt(1000))

	if _, err := env.bidSvc.PlaceBid(ctx, auction.ID, user.ID, decimal.NewFromInt(110)); err != nil {
		t.Fatalf("expected bid to succeed, got: %v", err)
	}
	// Simulate a legacy row where bidder_name/bidder_phone WERE populated --
	// the COALESCE fix must be a no-op for non-NULL values.
	if _, err := env.db.ExecContext(ctx,
		`UPDATE bids SET bidder_name = 'Ahmed', bidder_phone = '22212345678' WHERE auction_id = $1 AND user_id = $2`,
		auction.ID, user.ID); err != nil {
		t.Fatalf("failed to stamp legacy bidder details: %v", err)
	}

	bid, err := env.auctionRepo.GetUserHighestBid(ctx, auction.ID, user.ID)
	if err != nil {
		t.Fatalf("GetUserHighestBid failed: %v", err)
	}
	if bid.BidderName != "Ahmed" || bid.BidderPhone != "22212345678" {
		t.Fatalf("expected non-NULL bidder details preserved unchanged, got %q / %q", bid.BidderName, bid.BidderPhone)
	}
	t.Logf("confirmed: non-NULL legacy bidder details pass through unchanged")
}

// (19-13) GetBidStatus for a user who has never bid: HTTP-level has_bid=false,
// not an error (the earlier, separately-fixed sql.ErrNoRows bug).
func TestGetBidStatus_NeverBid_ReturnsHasBidFalse(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST BIDSTATUS SELLER A")
	user := createTestUser(t, env, "TEST BIDSTATUS USER A")
	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")

	status, err := env.auctSvc.GetBidStatus(ctx, auction.ID, user.ID)
	if err != nil {
		t.Fatalf("(19-13) expected GetBidStatus to succeed for a never-bid user, got: %v", err)
	}
	if status["has_bid"] != false {
		t.Fatalf("(19-13) expected has_bid=false for a never-bid user, got: %v", status["has_bid"])
	}
	t.Logf("(19-13) confirmed: GetBidStatus returns has_bid=false (not an error) for a never-bid user")
}

// (19-14) GetBidStatus immediately after a real first successful bid (with
// NULL legacy bidder_name/bidder_phone, the normal case): must succeed with
// has_bid=true, proving the NULL-scan fix actually unblocks this path.
func TestGetBidStatus_AfterFirstBid_ReturnsHasBidTrue(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST BIDSTATUS SELLER B")
	user := createTestUser(t, env, "TEST BIDSTATUS USER B")
	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")
	creditWallet(t, env, user.ID, decimal.NewFromInt(1000))

	if _, err := env.bidSvc.PlaceBid(ctx, auction.ID, user.ID, decimal.NewFromInt(110)); err != nil {
		t.Fatalf("expected bid to succeed, got: %v", err)
	}

	status, err := env.auctSvc.GetBidStatus(ctx, auction.ID, user.ID)
	if err != nil {
		t.Fatalf("(19-14) expected GetBidStatus to succeed after a real first bid, got: %v", err)
	}
	if status["has_bid"] != true {
		t.Fatalf("(19-14) expected has_bid=true after a real first bid, got: %v", status["has_bid"])
	}
	t.Logf("(19-14) confirmed: GetBidStatus returns has_bid=true immediately after a real first bid")
}

// FindTopBid shares the same previously-unsafe SELECT * scan path --
// confirm it too is now safe against NULL bidder_name/bidder_phone, and
// that winner-selection behavior built on it remains correct.
func TestFindTopBid_NullBidderDetails_Safe(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST TOPBID SELLER")
	userA := createTestUser(t, env, "TEST TOPBID USER A")
	userB := createTestUser(t, env, "TEST TOPBID USER B")
	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")
	creditWallet(t, env, userA.ID, decimal.NewFromInt(1000))
	creditWallet(t, env, userB.ID, decimal.NewFromInt(1000))

	if _, err := env.bidSvc.PlaceBid(ctx, auction.ID, userA.ID, decimal.NewFromInt(110)); err != nil {
		t.Fatalf("expected User A's bid to succeed, got: %v", err)
	}
	if _, err := env.bidSvc.PlaceBid(ctx, auction.ID, userB.ID, decimal.NewFromInt(150)); err != nil {
		t.Fatalf("expected User B's bid to succeed, got: %v", err)
	}

	top, err := env.bidRepo.FindTopBid(ctx, auction.ID)
	if err != nil {
		t.Fatalf("FindTopBid must not fail on NULL bidder details, got: %v", err)
	}
	if top.UserID != userB.ID {
		t.Fatalf("expected User B to be the top bid, got user %s", top.UserID)
	}
	if top.BidderName != "" || top.BidderPhone != "" {
		t.Fatalf("expected empty bidder details for NULL columns, got %q / %q", top.BidderName, top.BidderPhone)
	}
	t.Logf("confirmed: FindTopBid handles NULL bidder details safely, winner selection unaffected")
}

var _ = fmt.Sprintf // keep fmt import if unused paths change

// ==================================================
// Bug D (client feedback): GET /v1/api/users/me/settings returned HTTP 500
// for any user without a user_settings row -- which was every user, since
// nothing auto-creates this row at registration. Root cause, traced against
// current source: userRepo.GetUserSettings correctly detected the missing
// row (sql.ErrNoRows) and returned apperr.ErrNotFound, but UserHandler.
// GetUserSettings never routed errors through MapError -- it unconditionally
// returned InternalError (500) for ANY non-nil error. Separately,
// UpdateUserSettings's repository implementation was dead code: an
// `UPDATE ... SET updated_at = now()` that ignored every field in the
// payload and never checked RowsAffected, so even a "successful" PUT never
// created the missing row.
//
// Fixed by: (1) GetUserSettings returning models.DefaultUserSettings (the
// user_settings table's own column DEFAULTs) for a missing row instead of
// an error -- a missing row is normal, not a failure; (2) replacing the
// no-op UPDATE with `INSERT ... ON CONFLICT (user_id) DO UPDATE`, mirroring
// the already-correct adminService.UpdateUserSettings pattern, which is
// concurrency-safe because user_id is the table's PRIMARY KEY.
// ==================================================

// fullUserSettingsUpdate builds a models.UserSettingsUpdate with every field
// set (no nils) -- for tests exercising full-replacement-style calls, where
// models.UserSettingsUpdate's pointer fields would otherwise require
// verbose per-field &x boilerplate at every call site.
func fullUserSettingsUpdate(currency, theme, language string, notificationsEmail, notificationsPush, notificationsSMS, twoFactorEnabled bool) models.UserSettingsUpdate {
	return models.UserSettingsUpdate{
		Currency: &currency, Theme: &theme, Language: &language,
		NotificationsEmail: &notificationsEmail, NotificationsPush: &notificationsPush,
		NotificationsSMS: &notificationsSMS, TwoFactorEnabled: &twoFactorEnabled,
	}
}

// (D-1) GET for a user with no settings row returns the canonical defaults,
// not an error.
func TestGetUserSettings_MissingRow_ReturnsDefaults(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST SETTINGS USER D1")

	settings, err := env.userSvc.GetUserSettings(ctx, user.ID)
	if err != nil {
		t.Fatalf("(D-1) expected GetUserSettings to succeed for a user with no row, got: %v", err)
	}
	want := models.DefaultUserSettings(user.ID)
	if settings.Currency != want.Currency || settings.Theme != want.Theme || settings.Language != want.Language ||
		settings.NotificationsEmail != want.NotificationsEmail || settings.NotificationsPush != want.NotificationsPush ||
		settings.NotificationsSMS != want.NotificationsSMS || settings.TwoFactorEnabled != want.TwoFactorEnabled {
		t.Fatalf("(D-1) expected canonical defaults, got: %+v", settings)
	}
	if settings.UserID != user.ID {
		t.Fatalf("(D-1) expected default settings to carry the requesting user's ID, got %s", settings.UserID)
	}
	t.Logf("(D-1) confirmed: GetUserSettings returns canonical defaults for a missing row, not an error")
}

// (D-2) GET for a user with an existing, customized row returns the stored
// values, not defaults.
func TestGetUserSettings_ExistingRow_ReturnsStoredValues(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST SETTINGS USER D2")

	update := fullUserSettingsUpdate("TND", "dark", "fr", false, false, true, true)
	if err := env.userSvc.UpdateUserSettings(ctx, user.ID, update); err != nil {
		t.Fatalf("failed to seed a customized settings row: %v", err)
	}

	settings, err := env.userSvc.GetUserSettings(ctx, user.ID)
	if err != nil {
		t.Fatalf("(D-2) expected GetUserSettings to succeed, got: %v", err)
	}
	if settings.Currency != "TND" || settings.Theme != "dark" || settings.Language != "fr" ||
		settings.NotificationsEmail != false || settings.NotificationsPush != false ||
		settings.NotificationsSMS != true || settings.TwoFactorEnabled != true {
		t.Fatalf("(D-2) expected stored (non-default) values, got: %+v", settings)
	}
	t.Logf("(D-2) confirmed: GetUserSettings returns the real stored row, not defaults, once one exists")
}

// (D-3) PUT for a user with no row creates it (upsert insert path).
func TestUpdateUserSettings_MissingRow_CreatesIt(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST SETTINGS USER D3")

	var rowCountBefore int
	if err := env.db.GetContext(ctx, &rowCountBefore, `SELECT COUNT(*) FROM user_settings WHERE user_id = $1`, user.ID); err != nil {
		t.Fatalf("failed to count settings rows before: %v", err)
	}
	if rowCountBefore != 0 {
		t.Fatalf("(D-3) expected 0 settings rows before PUT, got %d", rowCountBefore)
	}

	update := fullUserSettingsUpdate("MRU", "light", "en", true, false, false, false)
	if err := env.userSvc.UpdateUserSettings(ctx, user.ID, update); err != nil {
		t.Fatalf("(D-3) expected PUT to succeed for a user with no row, got: %v", err)
	}

	var rowCountAfter int
	if err := env.db.GetContext(ctx, &rowCountAfter, `SELECT COUNT(*) FROM user_settings WHERE user_id = $1`, user.ID); err != nil {
		t.Fatalf("failed to count settings rows after: %v", err)
	}
	if rowCountAfter != 1 {
		t.Fatalf("(D-3) expected exactly 1 settings row after PUT, got %d", rowCountAfter)
	}
	t.Logf("(D-3) confirmed: PUT for a missing row creates exactly one row")
}

// (D-4) PUT for a user with an existing row updates it in place.
func TestUpdateUserSettings_ExistingRow_UpdatesInPlace(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST SETTINGS USER D4")

	first := fullUserSettingsUpdate("MRU", "light", "ar", true, true, false, false)
	if err := env.userSvc.UpdateUserSettings(ctx, user.ID, first); err != nil {
		t.Fatalf("failed initial PUT: %v", err)
	}

	second := fullUserSettingsUpdate("EUR", "dark", "en", false, false, true, true)
	if err := env.userSvc.UpdateUserSettings(ctx, user.ID, second); err != nil {
		t.Fatalf("(D-4) expected second PUT to succeed, got: %v", err)
	}

	var rowCount int
	if err := env.db.GetContext(ctx, &rowCount, `SELECT COUNT(*) FROM user_settings WHERE user_id = $1`, user.ID); err != nil {
		t.Fatalf("failed to count settings rows: %v", err)
	}
	if rowCount != 1 {
		t.Fatalf("(D-4) expected exactly 1 settings row after two PUTs, got %d (REGRESSION: duplicate row created)", rowCount)
	}

	settings, err := env.userSvc.GetUserSettings(ctx, user.ID)
	if err != nil {
		t.Fatalf("(D-4) GetUserSettings failed: %v", err)
	}
	if settings.Currency != "EUR" || settings.Theme != "dark" {
		t.Fatalf("(D-4) expected the second PUT's values to have replaced the first, got: %+v", settings)
	}
	t.Logf("(D-4) confirmed: PUT on an existing row updates the same row, never creates a duplicate")
}

// (D-5) PUT followed by GET is consistent -- the exact values just written
// are what GET returns immediately after, end to end through the service
// layer (not just the repo).
func TestUserSettings_PutThenGet_Consistent(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST SETTINGS USER D5")

	update := fullUserSettingsUpdate("TND", "dark", "fr", false, true, true, true)
	if err := env.userSvc.UpdateUserSettings(ctx, user.ID, update); err != nil {
		t.Fatalf("(D-5) PUT failed: %v", err)
	}

	settings, err := env.userSvc.GetUserSettings(ctx, user.ID)
	if err != nil {
		t.Fatalf("(D-5) GET after PUT failed: %v", err)
	}
	if settings.Currency != *update.Currency || settings.Theme != *update.Theme || settings.Language != *update.Language ||
		settings.NotificationsEmail != *update.NotificationsEmail || settings.NotificationsPush != *update.NotificationsPush ||
		settings.NotificationsSMS != *update.NotificationsSMS || settings.TwoFactorEnabled != *update.TwoFactorEnabled {
		t.Fatalf("(D-5) expected GET to reflect exactly what PUT wrote, got: %+v", settings)
	}
	t.Logf("(D-5) confirmed: PUT then GET is consistent")
}

// (D-6) Concurrent first-time PUTs for the SAME user cannot create duplicate
// rows -- proving the fix relies on the database's ON CONFLICT (user_id),
// not a race-prone SELECT-then-INSERT.
func TestUpdateUserSettings_ConcurrentFirstPuts_NoDuplicateRows(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST SETTINGS USER D6")

	const n = 8
	var wg sync.WaitGroup
	errs := make([]error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			update := fullUserSettingsUpdate("MRU", "auto", "ar", true, true, false, false)
			errs[i] = env.userSvc.UpdateUserSettings(ctx, user.ID, update)
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("(D-6) concurrent PUT %d failed: %v", i, err)
		}
	}

	var rowCount int
	if err := env.db.GetContext(ctx, &rowCount, `SELECT COUNT(*) FROM user_settings WHERE user_id = $1`, user.ID); err != nil {
		t.Fatalf("failed to count settings rows: %v", err)
	}
	if rowCount != 1 {
		t.Fatalf("(D-6) expected exactly 1 settings row after %d concurrent PUTs, got %d", n, rowCount)
	}
	t.Logf("(D-6) confirmed: %d concurrent first-time PUTs produced exactly 1 row (ON CONFLICT is the sole concurrency authority)", n)
}

// (D-7) User isolation: one user's settings (missing-row defaults, or a
// customized row) are never visible to or overwritten by another user.
func TestUserSettings_UserIsolation(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	userA := createTestUser(t, env, "TEST SETTINGS USER D7 A")
	userB := createTestUser(t, env, "TEST SETTINGS USER D7 B")

	updateA := fullUserSettingsUpdate("TND", "dark", "fr", false, false, true, true)
	if err := env.userSvc.UpdateUserSettings(ctx, userA.ID, updateA); err != nil {
		t.Fatalf("failed to set User A's settings: %v", err)
	}

	// User B never PUT anything -- must still see defaults, not A's values.
	settingsB, err := env.userSvc.GetUserSettings(ctx, userB.ID)
	if err != nil {
		t.Fatalf("(D-7) GetUserSettings for User B failed: %v", err)
	}
	if settingsB.Currency == "TND" || settingsB.Theme == "dark" {
		t.Fatalf("(D-7) SECURITY REGRESSION: User B's settings leaked User A's values: %+v", settingsB)
	}
	want := models.DefaultUserSettings(userB.ID)
	if settingsB.Currency != want.Currency || settingsB.Theme != want.Theme {
		t.Fatalf("(D-7) expected User B (no row) to see defaults, got: %+v", settingsB)
	}

	settingsA, err := env.userSvc.GetUserSettings(ctx, userA.ID)
	if err != nil {
		t.Fatalf("(D-7) GetUserSettings for User A failed: %v", err)
	}
	if settingsA.Currency != "TND" || settingsA.Theme != "dark" {
		t.Fatalf("(D-7) expected User A's own settings to be unaffected, got: %+v", settingsA)
	}
	t.Logf("(D-7) confirmed: settings are fully isolated per user")
}

// (D-8) notifications_push default and persistence -- distinct from
// Customer #10's separate users.notifications_enabled column (see D-9),
// this proves user_settings.notifications_push specifically defaults to
// TRUE (matching the table's own column DEFAULT) and persists correctly.
func TestUserSettings_NotificationsPush_DefaultAndPersistence(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST SETTINGS USER D8")

	defaults, err := env.userSvc.GetUserSettings(ctx, user.ID)
	if err != nil {
		t.Fatalf("(D-8) GetUserSettings failed: %v", err)
	}
	if !defaults.NotificationsPush {
		t.Fatalf("(D-8) expected notifications_push to default to true, got false")
	}

	update := fullUserSettingsUpdate("MRU", "auto", "ar", true, false, false, false)
	if err := env.userSvc.UpdateUserSettings(ctx, user.ID, update); err != nil {
		t.Fatalf("(D-8) PUT failed: %v", err)
	}
	after, err := env.userSvc.GetUserSettings(ctx, user.ID)
	if err != nil {
		t.Fatalf("(D-8) GetUserSettings after PUT failed: %v", err)
	}
	if after.NotificationsPush {
		t.Fatalf("(D-8) expected notifications_push=false to persist after PUT, got true")
	}
	t.Logf("(D-8) confirmed: user_settings.notifications_push defaults to true and persists correctly")
}

// (D-9) Customer #10 regression guard: users.notifications_enabled (a
// SEPARATE column, on a separate table, controlling push-delivery
// eligibility) is completely untouched by this fix -- proving the
// user_settings upsert never reads or writes it.
func TestUserSettings_DoesNotRegressCustomer10NotificationPref(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST SETTINGS USER D9")

	// Disable Customer #10's users.notifications_enabled directly.
	if _, err := env.db.ExecContext(ctx, `UPDATE users SET notifications_enabled = false WHERE id = $1`, user.ID); err != nil {
		t.Fatalf("failed to set up users.notifications_enabled = false: %v", err)
	}

	// Exercise the user_settings GET/PUT/GET cycle this round changed.
	update := fullUserSettingsUpdate("MRU", "dark", "ar", true, true, false, false)
	if err := env.userSvc.UpdateUserSettings(ctx, user.ID, update); err != nil {
		t.Fatalf("(D-9) PUT failed: %v", err)
	}
	if _, err := env.userSvc.GetUserSettings(ctx, user.ID); err != nil {
		t.Fatalf("(D-9) GET failed: %v", err)
	}

	var notificationsEnabled bool
	if err := env.db.GetContext(ctx, &notificationsEnabled, `SELECT notifications_enabled FROM users WHERE id = $1`, user.ID); err != nil {
		t.Fatalf("failed to read back users.notifications_enabled: %v", err)
	}
	if notificationsEnabled {
		t.Fatalf("(D-9) REGRESSION: users.notifications_enabled changed from false to true -- user_settings fix must never touch Customer #10's column")
	}
	t.Logf("(D-9) confirmed: Customer #10's users.notifications_enabled is untouched by the user_settings GET/PUT fix")
}

// (D-10) Bug D FINAL HARDENING: mobile's real PUT caller
// (SettingsPage._updateSetting) sends exactly ONE field per call, e.g.
// {"notifications_push": false} -- proving PUT is genuinely PARTIAL_UPDATE
// semantics, not full replacement. This test seeds a fully customized row,
// then sends the smallest possible partial update (only notifications_push
// via a hand-built single-field JSON body decoded into
// models.UserSettingsUpdate, exactly mirroring what Fiber's BodyParser does
// for a real {"notifications_push": false} request body), and proves every
// other field survives untouched. Before the pointer-field + SQL COALESCE
// fix, pre-populating Go zero-value defaults and overlaying BodyParser on
// top would have silently reset theme/language/currency/other notification
// flags back to defaults on every single toggle tap in the real app.
func TestUserSettings_PartialUpdatePreservesUnspecifiedFields(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST SETTINGS USER D10")

	seed := fullUserSettingsUpdate("MRU", "dark", "fr", false, true, true, true)
	if err := env.userSvc.UpdateUserSettings(ctx, user.ID, seed); err != nil {
		t.Fatalf("(D-10) failed to seed existing settings row: %v", err)
	}

	before, err := env.userSvc.GetUserSettings(ctx, user.ID)
	if err != nil {
		t.Fatalf("(D-10) GET before partial update failed: %v", err)
	}
	if before.Currency != "MRU" || before.Theme != "dark" || before.Language != "fr" ||
		before.NotificationsEmail != false || before.NotificationsPush != true ||
		before.NotificationsSMS != true || before.TwoFactorEnabled != true {
		t.Fatalf("(D-10) PARTIAL_UPDATE_BEFORE: seed did not persist as expected, got: %+v", before)
	}
	t.Logf("(D-10) PARTIAL_UPDATE_BEFORE: %+v", before)

	// The smallest partial update supported by the real API: decode the
	// exact JSON a real client sends for a single toggle, proving the
	// handler's actual BodyParser path (not just the repo) preserves
	// unspecified fields.
	var partial models.UserSettingsUpdate
	if err := json.Unmarshal([]byte(`{"notifications_push": false}`), &partial); err != nil {
		t.Fatalf("(D-10) failed to decode single-field JSON body: %v", err)
	}
	if partial.Currency != nil || partial.Theme != nil || partial.Language != nil ||
		partial.NotificationsEmail != nil || partial.NotificationsSMS != nil || partial.TwoFactorEnabled != nil {
		t.Fatalf("(D-10) expected every field except NotificationsPush to decode as nil (omitted), got: %+v", partial)
	}
	if partial.NotificationsPush == nil || *partial.NotificationsPush != false {
		t.Fatalf("(D-10) expected NotificationsPush to decode to a non-nil false, got: %+v", partial.NotificationsPush)
	}

	if err := env.userSvc.UpdateUserSettings(ctx, user.ID, partial); err != nil {
		t.Fatalf("(D-10) partial PUT failed: %v", err)
	}

	after, err := env.userSvc.GetUserSettings(ctx, user.ID)
	if err != nil {
		t.Fatalf("(D-10) GET after partial update failed: %v", err)
	}
	t.Logf("(D-10) PARTIAL_UPDATE_AFTER: %+v", after)

	if after.NotificationsPush != false {
		t.Fatalf("(D-10) expected notifications_push=false to be applied, got true")
	}
	if after.Currency != "MRU" || after.Theme != "dark" || after.Language != "fr" ||
		after.NotificationsEmail != false || after.NotificationsSMS != true || after.TwoFactorEnabled != true {
		t.Fatalf("(D-10) UNSPECIFIED_FIELDS_PRESERVED=NO: partial update destroyed unspecified fields, got: %+v (expected currency=MRU theme=dark language=fr notifications_email=false notifications_sms=true two_factor_enabled=true, all unchanged from before)", after)
	}

	var rowCount int
	if err := env.db.GetContext(ctx, &rowCount, `SELECT COUNT(*) FROM user_settings WHERE user_id = $1`, user.ID); err != nil {
		t.Fatalf("failed to count settings rows: %v", err)
	}
	if rowCount != 1 {
		t.Fatalf("(D-10) expected exactly 1 settings row, got %d", rowCount)
	}
	t.Logf("(D-10) confirmed: UNSPECIFIED_FIELDS_PRESERVED=YES -- partial PUT changed only notifications_push, every other field survived unchanged")
}

// ==================================================
// Bug A.1 (pre-APK error-mapping cleanup): CreateBannerRequest's
// ends_at<=starts_at business rule returned a plain errors.New(...), which
// MapError had no case for, so it fell through to an unmapped HTTP 500
// instead of 400. Fixed by returning apperr.ErrBadRequest (an existing,
// reusable "generic bad request" domain error, already mapped by
// MapError's "bad_request" case) and tightening the check from
// EndsAt.Before(StartsAt) to !EndsAt.After(StartsAt), so a zero-length
// window (ends_at == starts_at) is also correctly rejected, not just
// ends_at < starts_at.
// ==================================================

// (A1-1) A valid date range (ends_at > starts_at) is unaffected.
func TestCreateBannerRequest_A1_ValidDateRange_Unaffected(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST A1 SELLER VALID")

	req := newBannerRequest(seller.ID, "a1-valid-"+uuid.New().String()[:6])
	if err := env.reqSvc.CreateBannerRequest(ctx, req); err != nil {
		t.Fatalf("(A1-1) expected a valid date range to succeed, got: %v", err)
	}

	var count int
	if err := env.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM banner_requests WHERE id = $1`, req.ID); err != nil {
		t.Fatalf("failed to count banner_requests: %v", err)
	}
	if count != 1 {
		t.Fatalf("(A1-1) expected the valid banner request to persist, got %d rows", count)
	}
	t.Logf("(A1-1) confirmed: a valid date range still succeeds and persists")
}

// (A1-2) ends_at == starts_at (a zero-length window) is rejected as
// apperr.ErrBadRequest, and no row is persisted.
func TestCreateBannerRequest_A1_EqualDates_RejectedAsBadRequest(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST A1 SELLER EQUAL")

	req := newBannerRequest(seller.ID, "a1-equal-"+uuid.New().String()[:6])
	same := time.Now().Add(2 * time.Hour)
	req.StartsAt = same
	req.EndsAt = same

	err := env.reqSvc.CreateBannerRequest(ctx, req)
	if err != apperr.ErrBadRequest {
		t.Fatalf("(A1-2) expected apperr.ErrBadRequest for ends_at == starts_at, got: %v", err)
	}

	var count int
	if err := env.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM banner_requests WHERE id = $1`, req.ID); err != nil {
		t.Fatalf("failed to count banner_requests: %v", err)
	}
	if count != 0 {
		t.Fatalf("(A1-2) expected no row to be persisted for a rejected date range, got %d rows", count)
	}
	t.Logf("(A1-2) confirmed: ends_at == starts_at correctly rejected as ErrBadRequest, no row persisted")
}

// (A1-3) ends_at < starts_at is rejected as apperr.ErrBadRequest, and no
// row is persisted.
func TestCreateBannerRequest_A1_EndBeforeStart_RejectedAsBadRequest(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST A1 SELLER BEFORE")

	req := newBannerRequest(seller.ID, "a1-before-"+uuid.New().String()[:6])
	req.StartsAt = time.Now().Add(48 * time.Hour)
	req.EndsAt = time.Now().Add(1 * time.Hour)

	err := env.reqSvc.CreateBannerRequest(ctx, req)
	if err != apperr.ErrBadRequest {
		t.Fatalf("(A1-3) expected apperr.ErrBadRequest for ends_at < starts_at, got: %v", err)
	}

	var count int
	if err := env.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM banner_requests WHERE id = $1`, req.ID); err != nil {
		t.Fatalf("failed to count banner_requests: %v", err)
	}
	if count != 0 {
		t.Fatalf("(A1-3) expected no row to be persisted for a rejected date range, got %d rows", count)
	}
	t.Logf("(A1-3) confirmed: ends_at < starts_at correctly rejected as ErrBadRequest, no row persisted")
}

// ==================================================
// Bug G (client feedback): the Admin web's auction edit/save form always
// re-sends the auction's current image URLs in UpdateAuctionInput.Images,
// even when the admin changed nothing about images (e.g. editing only the
// title or price). AdminService.UpdateAuction previously treated ANY
// non-empty Images slice as "images changed" -- deleting the real R2
// objects and reinserting the exact same URLs into auction_images, leaving
// the DB pointing at files that no longer existed. Real-device symptom:
// an approved auction's image showed a broken/missing placeholder on My
// Auctions, Auction Details, Favorites, and Active Auctions alike, even
// though the DB row and URL looked completely valid.
//
// Fixed by comparing the incoming, ordered, non-empty URL list against the
// existing auction_images rows (also ordered by display_order) before
// doing anything: a same-images save is now a true no-op for images (no
// DB churn, no R2 deletion); only a genuine URL-set change triggers the
// replace-and-cleanup path, which still runs exactly as before.
//
// These tests exercise the real adminService.UpdateAuction against a
// local MediaService (falls back to local-storage mode without R2 creds,
// so DeleteFile calls either no-op safely or fail silently -- best-effort
// per the existing code, never fails the whole operation) to prove the
// image DB rows themselves are/aren't touched, which is what the API/
// mobile layers actually observe.
// ==================================================

func newTestAdminService(t *testing.T, env *testEnv) services.AdminService {
	t.Helper()
	cfg := config.Load()
	txRepo := repository.NewTransactionRepository(env.db, env.walletRepo)
	reportRepo := repository.NewReportRepository(env.db)
	contentRepo := repository.NewContentRepository(env.db)
	kycRepo := repository.NewKYCRepository(env.db)
	invRepo := repository.NewAdminInvitationRepository(env.db)
	settingsRepo := repository.NewSettingsRepository(env.db)
	auditRepo := repository.NewAuditRepository(env.db)
	auditSvc := services.NewAuditService(auditRepo)
	mediaSvc := services.NewMediaService(cfg, env.logger)

	return services.NewAdminService(
		env.db, env.userRepo, env.auctionRepo, env.bidRepo, txRepo, reportRepo,
		kycRepo, contentRepo, invRepo, env.reqRepo, settingsRepo,
		mediaSvc, env.notifSvc, auditSvc, env.rdb, env.logger, cfg.JWT.ExpiryHours, nil,
	)
}

func getAuctionImageURLsOrdered(t *testing.T, env *testEnv, auctionID uuid.UUID) []string {
	t.Helper()
	ctx := context.Background()
	var urls []string
	if err := env.db.SelectContext(ctx, &urls,
		`SELECT url FROM auction_images WHERE auction_id = $1 ORDER BY display_order`, auctionID); err != nil {
		t.Fatalf("failed to read auction_images: %v", err)
	}
	return urls
}

func seedAuctionImage(t *testing.T, env *testEnv, auctionID uuid.UUID, url string, order int) {
	t.Helper()
	ctx := context.Background()
	if _, err := env.db.ExecContext(ctx,
		`INSERT INTO auction_images (auction_id, url, media_type, display_order) VALUES ($1, $2, 'image', $3)`,
		auctionID, url, order); err != nil {
		t.Fatalf("failed to seed auction_images row: %v", err)
	}
}

func baseUpdateAuctionInput(a *models.Auction) services.UpdateAuctionInput {
	return services.UpdateAuctionInput{
		CategoryID:      a.CategoryID,
		TitleAr:         a.TitleAr,
		StartPrice:      a.StartPrice,
		MinIncrement:    a.MinIncrement,
		InsuranceAmount: a.InsuranceAmount,
		EndTime:         a.EndTime,
		Quantity:        1,
	}
}

// (G-1) Saving the auction with the exact same image URLs it already has
// must NOT delete the R2 object or touch the image rows.
func TestUpdateAuction_G1_SameImagesResubmitted_NoImageChange(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST BUGG SELLER G1")
	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")
	seedAuctionImage(t, env, auction.ID, "https://example.com/g1-photo.jpg", 0)

	adminSvc := newTestAdminService(t, env)

	before := getAuctionImageURLsOrdered(t, env, auction.ID)
	if len(before) != 1 || before[0] != "https://example.com/g1-photo.jpg" {
		t.Fatalf("(G-1) expected seeded image before update, got: %v", before)
	}

	input := baseUpdateAuctionInput(auction)
	input.Images = []string{"https://example.com/g1-photo.jpg"}
	if err := adminSvc.UpdateAuction(ctx, auction.ID, input); err != nil {
		t.Fatalf("(G-1) UpdateAuction failed: %v", err)
	}

	after := getAuctionImageURLsOrdered(t, env, auction.ID)
	if len(after) != 1 || after[0] != "https://example.com/g1-photo.jpg" {
		t.Fatalf("(G-1) expected the exact same image row to survive unchanged, got: %v", after)
	}
	t.Logf("(G-1) confirmed: resubmitting the same image URL is a no-op, image row untouched")
}

// (G-2) Editing an unrelated field (title) while images are re-sent
// unchanged must leave images completely untouched.
func TestUpdateAuction_G2_UnrelatedFieldEdit_ImagesUntouched(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST BUGG SELLER G2")
	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")
	seedAuctionImage(t, env, auction.ID, "https://example.com/g2-photo.jpg", 0)

	adminSvc := newTestAdminService(t, env)

	input := baseUpdateAuctionInput(auction)
	input.TitleAr = "عنوان معدل G2"
	input.Images = []string{"https://example.com/g2-photo.jpg"}
	if err := adminSvc.UpdateAuction(ctx, auction.ID, input); err != nil {
		t.Fatalf("(G-2) UpdateAuction failed: %v", err)
	}

	var newTitle string
	if err := env.db.GetContext(ctx, &newTitle, `SELECT title_ar FROM auctions WHERE id = $1`, auction.ID); err != nil {
		t.Fatalf("failed to read back title: %v", err)
	}
	if newTitle != "عنوان معدل G2" {
		t.Fatalf("(G-2) expected title to update, got: %q", newTitle)
	}

	after := getAuctionImageURLsOrdered(t, env, auction.ID)
	if len(after) != 1 || after[0] != "https://example.com/g2-photo.jpg" {
		t.Fatalf("(G-2) expected image untouched by an unrelated field edit, got: %v", after)
	}
	t.Logf("(G-2) confirmed: editing title alone leaves images completely untouched")
}

// (G-3) An actual image URL change (replacement) still triggers the
// replace-and-cleanup path correctly -- new URL persists.
func TestUpdateAuction_G3_ActualImageReplacement_StillWorks(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST BUGG SELLER G3")
	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")
	seedAuctionImage(t, env, auction.ID, "https://example.com/g3-old.jpg", 0)

	adminSvc := newTestAdminService(t, env)

	input := baseUpdateAuctionInput(auction)
	input.Images = []string{"https://example.com/g3-new.jpg"}
	if err := adminSvc.UpdateAuction(ctx, auction.ID, input); err != nil {
		t.Fatalf("(G-3) UpdateAuction failed: %v", err)
	}

	after := getAuctionImageURLsOrdered(t, env, auction.ID)
	if len(after) != 1 || after[0] != "https://example.com/g3-new.jpg" {
		t.Fatalf("(G-3) expected the new image URL to replace the old one, got: %v", after)
	}
	t.Logf("(G-3) confirmed: a genuine image URL change still replaces the image row correctly")
}

// (G-4) Removing an image (submitting fewer URLs than currently stored)
// still triggers cleanup of the removed image.
func TestUpdateAuction_G4_ImageRemoval_StillWorks(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST BUGG SELLER G4")
	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")
	seedAuctionImage(t, env, auction.ID, "https://example.com/g4-a.jpg", 0)
	seedAuctionImage(t, env, auction.ID, "https://example.com/g4-b.jpg", 1)

	adminSvc := newTestAdminService(t, env)

	input := baseUpdateAuctionInput(auction)
	input.Images = []string{"https://example.com/g4-a.jpg"}
	if err := adminSvc.UpdateAuction(ctx, auction.ID, input); err != nil {
		t.Fatalf("(G-4) UpdateAuction failed: %v", err)
	}

	after := getAuctionImageURLsOrdered(t, env, auction.ID)
	if len(after) != 1 || after[0] != "https://example.com/g4-a.jpg" {
		t.Fatalf("(G-4) expected only the retained image to remain, got: %v", after)
	}
	t.Logf("(G-4) confirmed: removing an image from the submitted list correctly removes it")
}

// (G-5) Multiple unchanged images (order-sensitive) remain untouched.
func TestUpdateAuction_G5_MultipleUnchangedImages_NoneDeleted(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST BUGG SELLER G5")
	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")
	seedAuctionImage(t, env, auction.ID, "https://example.com/g5-a.jpg", 0)
	seedAuctionImage(t, env, auction.ID, "https://example.com/g5-b.jpg", 1)
	seedAuctionImage(t, env, auction.ID, "https://example.com/g5-c.jpg", 2)

	adminSvc := newTestAdminService(t, env)

	input := baseUpdateAuctionInput(auction)
	input.Images = []string{"https://example.com/g5-a.jpg", "https://example.com/g5-b.jpg", "https://example.com/g5-c.jpg"}
	if err := adminSvc.UpdateAuction(ctx, auction.ID, input); err != nil {
		t.Fatalf("(G-5) UpdateAuction failed: %v", err)
	}

	after := getAuctionImageURLsOrdered(t, env, auction.ID)
	want := []string{"https://example.com/g5-a.jpg", "https://example.com/g5-b.jpg", "https://example.com/g5-c.jpg"}
	if len(after) != len(want) {
		t.Fatalf("(G-5) expected all 3 images to survive unchanged, got: %v", after)
	}
	for i := range want {
		if after[i] != want[i] {
			t.Fatalf("(G-5) expected order/content preserved, got: %v", after)
		}
	}
	t.Logf("(G-5) confirmed: multiple unchanged images (order-sensitive) are never deleted")
}

// (G-6) Full real Bug G flow, matching the actual mobile/admin sequence
// traced from real Staging logs: mobile creates the auction request,
// admin approves it (ReviewAuctionRequest creates the `auctions` row --
// this step does NOT itself carry any image, confirmed by reading its
// source: images are added afterward via a separate call), mobile then
// uploads the image directly to the new auction (POST
// /auctions/:id/images -> AuctionService.AddImages), and the image must
// REMAIN after a later, unrelated Admin web save -- reproducing the exact
// real-device regression end to end.
func TestUpdateAuction_G6_ApprovalCreatedAuctionImageSurvivesLaterSave(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST BUGG SELLER G6")
	admin := createTestAdmin(t, env, "TEST BUGG ADMIN G6")

	auctionReq := &models.AuctionRequest{
		ID:            uuid.New(),
		UserID:        seller.ID,
		CategoryID:    mkCategoryID(),
		TitleAr:       "مزاد اختبار G6",
		DescriptionAr: strPtr("وصف تجريبي آمن يزيد عن عشرة أحرف"),
		StartPrice:    decimal.NewFromInt(100),
		MinIncrement:  decimal.NewFromInt(10),
		StartDate:     time.Now().Add(1 * time.Hour),
		EndDate:       time.Now().Add(48 * time.Hour),
		Status:        "pending",
	}
	if err := env.reqSvc.CreateAuctionRequest(ctx, auctionReq); err != nil {
		t.Fatalf("(G-6) CreateAuctionRequest failed: %v", err)
	}

	// ReviewAuctionRequest refuses to approve until insurance is explicitly
	// set by an admin (audit V03 guard, unrelated to Bug G) -- same
	// real workflow as the other approval tests in this file.
	insuranceUpdate := *auctionReq
	insuranceUpdate.InsuranceAmount = decimal.NewFromInt(50)
	if err := env.reqSvc.AdminUpdateAuctionRequest(ctx, auctionReq.ID, &insuranceUpdate, nil); err != nil {
		t.Fatalf("(G-6) AdminUpdateAuctionRequest (setting insurance) failed: %v", err)
	}

	if err := env.reqSvc.ReviewAuctionRequest(ctx, auctionReq.ID, "approved", "approved for G-6", admin.ID); err != nil {
		t.Fatalf("(G-6) ReviewAuctionRequest(approved) failed: %v", err)
	}

	var createdAuctionID uuid.UUID
	if err := env.db.GetContext(ctx, &createdAuctionID,
		`SELECT id FROM auctions WHERE seller_id = $1 AND title_ar = $2 ORDER BY created_at DESC LIMIT 1`,
		seller.ID, "مزاد اختبار G6"); err != nil {
		t.Fatalf("(G-6) failed to find the approval-created auction: %v", err)
	}

	// Mobile's post-approval upload step (AuctionService.AddImages), the
	// real path that actually attaches the image to the new auction.
	if err := env.auctSvc.AddImages(ctx, createdAuctionID, seller.ID, []string{"https://example.com/g6-uploaded-photo.jpg"}); err != nil {
		t.Fatalf("(G-6) AddImages (mobile post-approval upload) failed: %v", err)
	}

	imagesAfterUpload := getAuctionImageURLsOrdered(t, env, createdAuctionID)
	if len(imagesAfterUpload) != 1 || imagesAfterUpload[0] != "https://example.com/g6-uploaded-photo.jpg" {
		t.Fatalf("(G-6) expected the mobile-uploaded image to attach to the approval-created auction, got: %v", imagesAfterUpload)
	}

	// Now simulate the Admin web's later, unrelated edit/save (echoing the
	// same image URL back, per its real save behavior) -- this is the
	// exact real-device Bug G reproduction.
	var createdAuction models.Auction
	if err := env.db.GetContext(ctx, &createdAuction, `SELECT * FROM auctions WHERE id = $1`, createdAuctionID); err != nil {
		t.Fatalf("failed to read back created auction: %v", err)
	}

	adminSvc := newTestAdminService(t, env)
	input := baseUpdateAuctionInput(&createdAuction)
	input.TitleAr = "مزاد اختبار G6 معدل"
	input.Images = imagesAfterUpload
	if err := adminSvc.UpdateAuction(ctx, createdAuctionID, input); err != nil {
		t.Fatalf("(G-6) UpdateAuction (later save) failed: %v", err)
	}

	imagesAfterLaterSave := getAuctionImageURLsOrdered(t, env, createdAuctionID)
	if len(imagesAfterLaterSave) != 1 || imagesAfterLaterSave[0] != "https://example.com/g6-uploaded-photo.jpg" {
		t.Fatalf("(G-6) REGRESSION: the approval-created auction's image did not survive a later unrelated admin save, got: %v", imagesAfterLaterSave)
	}
	t.Logf("(G-6) confirmed: the exact real Bug G flow (request -> approval -> mobile image upload -> later admin save) preserves the image")
}

// === Customer Request #21: deposit/withdrawal in-app notifications ===
//
// Root cause (see audit): deposit_confirmed/deposit_rejected/withdrawal_processed
// were already correctly used by adminService.ValidateTransaction, but were never
// added to the notifications table's chk_notif_type CHECK constraint after
// migration 000038 -- every INSERT attempt silently failed the constraint, and
// NotificationService.SendPush swallows repo.Create errors (logs only), so the
// failure was invisible. Separately, wallet_service.go's InitiateDeposit/
// RequestWithdraw never attempted a notification at all -- migration 000051 fixes
// the constraint, and this round adds two new, deliberately distinct submission-
// acknowledgment types (deposit_submitted/withdrawal_submitted) so a still-pending
// request is never reported using the outcome types.

// getLatestNotification returns the most recent notification row of exactly one
// type for a user, or nil if none exists. Used to assert on the outcome of a
// specific action without over-matching an unrelated notification the same
// fixture user might also have received.
func getLatestNotificationOfType(t *testing.T, env *testEnv, userID uuid.UUID, notifType string) *models.Notification {
	t.Helper()
	notifs, err := env.notifSvc.ListNotifications(context.Background(), userID, 50)
	if err != nil {
		t.Fatalf("ListNotifications failed: %v", err)
	}
	for _, n := range notifs {
		if n.Type == notifType {
			cp := n
			return &cp
		}
	}
	return nil
}

func countNotificationsOfType(t *testing.T, env *testEnv, userID uuid.UUID, notifType string) int {
	t.Helper()
	notifs, err := env.notifSvc.ListNotifications(context.Background(), userID, 50)
	if err != nil {
		t.Fatalf("ListNotifications failed: %v", err)
	}
	count := 0
	for _, n := range notifs {
		if n.Type == notifType {
			count++
		}
	}
	return count
}

// (1-4) Migration 000051: allows the three real outcome types plus preserves
// every previously-valid type. Exercised directly against the notifications
// table (bypassing the service layer) since the whole point is to test the DB
// CHECK constraint itself.
func TestNotificationMigration_AllowsWalletOutcomeTypes(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST NOTIF MIGRATION USER")

	for _, notifType := range []string{"deposit_confirmed", "deposit_rejected", "withdrawal_processed", "deposit_submitted", "withdrawal_submitted"} {
		_, err := env.db.ExecContext(ctx,
			`INSERT INTO notifications (id, user_id, type, title, body) VALUES ($1, $2, $3, $4, $5)`,
			uuid.New(), user.ID, notifType, "test title", "test body")
		if err != nil {
			t.Fatalf("(1-4) migration 000051 REGRESSION: inserting type %q was rejected: %v", notifType, err)
		}
	}
	t.Logf("(1-4) all 5 wallet notification types accepted by chk_notif_type")
}

func TestNotificationMigration_PreservesAllPreviousTypes(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST NOTIF PRESERVE USER")

	// Exact set from 000038_expand_notification_types.up.sql -- if migration
	// 000051 accidentally narrowed this set instead of only widening it, one of
	// these inserts would fail.
	previousTypes := []string{
		"bid", "win", "payment", "system", "ad", "general", "new_auction",
		"transaction", "report", "auction_sold", "new_message",
		"auction_ending_soon", "auction_approved", "auction_rejected",
		"banner_approved", "banner_rejected", "auction_pending",
		"auction_won", "auction_ended", "payment_received", "auction_reported",
	}
	for _, notifType := range previousTypes {
		_, err := env.db.ExecContext(ctx,
			`INSERT INTO notifications (id, user_id, type, title, body) VALUES ($1, $2, $3, $4, $5)`,
			uuid.New(), user.ID, notifType, "test title", "test body")
		if err != nil {
			t.Fatalf("(4) REGRESSION: previously-valid type %q was rejected after migration 000051 -- the constraint was narrowed instead of only widened: %v", notifType, err)
		}
	}
	t.Logf("(4) all %d previously-valid notification types remain accepted", len(previousTypes))
}

// (5-7) Deposit submission acknowledgment.
func TestDeposit_SubmissionCreatesAcknowledgment(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST DEPOSIT SUBMIT USER")
	walletSvc := newWalletSvc(env)

	tx, err := walletSvc.InitiateDeposit(ctx, user.ID, decimal.NewFromInt(500), "bankily", "mobile_money", "", nil)
	if err != nil {
		t.Fatalf("(5) InitiateDeposit failed: %v", err)
	}
	if tx.Status != "pending" {
		t.Fatalf("(5) expected newly-submitted deposit to be 'pending', got %q", tx.Status)
	}

	notif := getLatestNotificationOfType(t, env, user.ID, "deposit_submitted")
	if notif == nil {
		t.Fatalf("(5) expected a deposit_submitted notification, found none")
	}
	if notif.Title == "" || notif.Body == nil || *notif.Body == "" {
		t.Fatalf("(5) expected non-empty title/body, got title=%q body=%v", notif.Title, notif.Body)
	}
	// Anti-Bug-I-class-regression: submission ack must never use the outcome
	// type, even though both are triggered by "a deposit was made" in casual
	// language -- deposit_confirmed means an admin already approved it.
	if countNotificationsOfType(t, env, user.ID, "deposit_confirmed") != 0 {
		t.Fatalf("(5) CRITICAL: a still-pending deposit must never create a deposit_confirmed notification")
	}
}

func TestDeposit_SubmissionNotificationBelongsToCorrectUser(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	userA := createTestUser(t, env, "TEST DEPOSIT SCOPE USER A")
	userB := createTestUser(t, env, "TEST DEPOSIT SCOPE USER B")
	walletSvc := newWalletSvc(env)

	if _, err := walletSvc.InitiateDeposit(ctx, userA.ID, decimal.NewFromInt(300), "bankily", "mobile_money", "", nil); err != nil {
		t.Fatalf("(6) InitiateDeposit(A) failed: %v", err)
	}

	if getLatestNotificationOfType(t, env, userA.ID, "deposit_submitted") == nil {
		t.Fatalf("(6) user A should have received their own deposit_submitted notification")
	}
	if getLatestNotificationOfType(t, env, userB.ID, "deposit_submitted") != nil {
		t.Fatalf("(6) CRITICAL: user B received user A's deposit_submitted notification")
	}
}

func TestDeposit_FailedCreationCreatesNoNotification(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST DEPOSIT FAIL USER")
	walletSvc := newWalletSvc(env)

	// A zero/negative amount is rejected by InitiateDeposit's own validation
	// before any transaction row (and therefore any notification) is created.
	_, err := walletSvc.InitiateDeposit(ctx, user.ID, decimal.Zero, "bankily", "mobile_money", "", nil)
	if err == nil {
		t.Fatalf("(7) expected InitiateDeposit to reject a zero amount")
	}
	if countNotificationsOfType(t, env, user.ID, "deposit_submitted") != 0 {
		t.Fatalf("(7) CRITICAL: a failed deposit creation must never produce a false deposit_submitted notification")
	}
}

// (8-10) Deposit admin-approval outcome notifications.
func TestDeposit_ApprovalCreatesDepositConfirmed(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST DEPOSIT APPROVE USER")
	admin := createTestAdmin(t, env, "TEST DEPOSIT APPROVE ADMIN")
	walletSvc := newWalletSvc(env)
	adminSvc := newTestAdminService(t, env)

	tx, err := walletSvc.InitiateDeposit(ctx, user.ID, decimal.NewFromInt(1000), "bankily", "mobile_money", "", nil)
	if err != nil {
		t.Fatalf("(8) InitiateDeposit failed: %v", err)
	}
	if err := adminSvc.ValidateTransaction(ctx, tx.ID, true, "", admin.ID); err != nil {
		t.Fatalf("(8) ValidateTransaction(approve) failed: %v", err)
	}

	notif := getLatestNotificationOfType(t, env, user.ID, "deposit_confirmed")
	if notif == nil {
		t.Fatalf("(8) expected a deposit_confirmed notification after approval, found none -- this is the exact Customer #21 regression (chk_notif_type gap)")
	}
}

func TestDeposit_RejectionCreatesDepositRejected(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST DEPOSIT REJECT USER")
	admin := createTestAdmin(t, env, "TEST DEPOSIT REJECT ADMIN")
	walletSvc := newWalletSvc(env)
	adminSvc := newTestAdminService(t, env)

	tx, err := walletSvc.InitiateDeposit(ctx, user.ID, decimal.NewFromInt(1000), "bankily", "mobile_money", "", nil)
	if err != nil {
		t.Fatalf("(9) InitiateDeposit failed: %v", err)
	}
	if err := adminSvc.ValidateTransaction(ctx, tx.ID, false, "insufficient proof", admin.ID); err != nil {
		t.Fatalf("(9) ValidateTransaction(reject) failed: %v", err)
	}

	notif := getLatestNotificationOfType(t, env, user.ID, "deposit_rejected")
	if notif == nil {
		t.Fatalf("(9) expected a deposit_rejected notification after rejection, found none")
	}
}

func TestDeposit_RepeatedValidationDoesNotDuplicateOutcomeNotification(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST DEPOSIT DEDUPE USER")
	admin := createTestAdmin(t, env, "TEST DEPOSIT DEDUPE ADMIN")
	walletSvc := newWalletSvc(env)
	adminSvc := newTestAdminService(t, env)

	tx, err := walletSvc.InitiateDeposit(ctx, user.ID, decimal.NewFromInt(1000), "bankily", "mobile_money", "", nil)
	if err != nil {
		t.Fatalf("(10) InitiateDeposit failed: %v", err)
	}
	if err := adminSvc.ValidateTransaction(ctx, tx.ID, true, "", admin.ID); err != nil {
		t.Fatalf("(10) ValidateTransaction (first call) failed: %v", err)
	}
	// Second call on an already-terminal transaction: the existing terminal-
	// status guard in ValidateTransaction/UpdateStatus must make this a no-op,
	// not a duplicate notification.
	_ = adminSvc.ValidateTransaction(ctx, tx.ID, true, "", admin.ID)

	count := countNotificationsOfType(t, env, user.ID, "deposit_confirmed")
	if count != 1 {
		t.Fatalf("(10) expected exactly 1 deposit_confirmed notification after 2 validation calls (idempotency guard), got %d", count)
	}
}

// (11-13) Withdrawal submission acknowledgment.
func TestWithdrawal_SubmissionCreatesAcknowledgment(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST WITHDRAW SUBMIT USER")
	creditWallet(t, env, user.ID, decimal.NewFromInt(2000))
	walletSvc := newWalletSvc(env)

	tx, err := walletSvc.RequestWithdraw(ctx, user.ID, decimal.NewFromInt(500), "bankily")
	if err != nil {
		t.Fatalf("(11) RequestWithdraw failed: %v", err)
	}
	if tx.Status != "pending_review" {
		t.Fatalf("(11) expected newly-submitted withdrawal to be 'pending_review', got %q", tx.Status)
	}

	notif := getLatestNotificationOfType(t, env, user.ID, "withdrawal_submitted")
	if notif == nil {
		t.Fatalf("(11) expected a withdrawal_submitted notification, found none")
	}
	if countNotificationsOfType(t, env, user.ID, "withdrawal_processed") != 0 {
		t.Fatalf("(11) CRITICAL: a still-pending withdrawal must never create a withdrawal_processed notification")
	}
}

func TestWithdrawal_SubmissionNotificationBelongsToCorrectUser(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	userA := createTestUser(t, env, "TEST WITHDRAW SCOPE USER A")
	userB := createTestUser(t, env, "TEST WITHDRAW SCOPE USER B")
	creditWallet(t, env, userA.ID, decimal.NewFromInt(2000))
	walletSvc := newWalletSvc(env)

	if _, err := walletSvc.RequestWithdraw(ctx, userA.ID, decimal.NewFromInt(500), "bankily"); err != nil {
		t.Fatalf("(12) RequestWithdraw(A) failed: %v", err)
	}

	if getLatestNotificationOfType(t, env, userA.ID, "withdrawal_submitted") == nil {
		t.Fatalf("(12) user A should have received their own withdrawal_submitted notification")
	}
	if getLatestNotificationOfType(t, env, userB.ID, "withdrawal_submitted") != nil {
		t.Fatalf("(12) CRITICAL: user B received user A's withdrawal_submitted notification")
	}
}

func TestWithdrawal_FailedCreationCreatesNoNotification(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST WITHDRAW FAIL USER")
	// Deliberately no creditWallet call -- insufficient balance must make
	// FreezeForWithdraw fail before any transaction/notification is created.
	walletSvc := newWalletSvc(env)

	_, err := walletSvc.RequestWithdraw(ctx, user.ID, decimal.NewFromInt(999999), "bankily")
	if err == nil {
		t.Fatalf("(13) expected RequestWithdraw to fail on insufficient balance")
	}
	if countNotificationsOfType(t, env, user.ID, "withdrawal_submitted") != 0 {
		t.Fatalf("(13) CRITICAL: a failed withdrawal creation must never produce a false withdrawal_submitted notification")
	}
}

// (14-16) Withdrawal admin-approval outcome notifications.
func TestWithdrawal_CompletionCreatesWithdrawalProcessed(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST WITHDRAW COMPLETE USER")
	admin := createTestAdmin(t, env, "TEST WITHDRAW COMPLETE ADMIN")
	creditWallet(t, env, user.ID, decimal.NewFromInt(2000))
	walletSvc := newWalletSvc(env)
	adminSvc := newTestAdminService(t, env)

	tx, err := walletSvc.RequestWithdraw(ctx, user.ID, decimal.NewFromInt(800), "bankily")
	if err != nil {
		t.Fatalf("(14) RequestWithdraw failed: %v", err)
	}
	if err := adminSvc.ValidateTransaction(ctx, tx.ID, true, "", admin.ID); err != nil {
		t.Fatalf("(14) ValidateTransaction(approve) failed: %v", err)
	}

	notif := getLatestNotificationOfType(t, env, user.ID, "withdrawal_processed")
	if notif == nil {
		t.Fatalf("(14) expected a withdrawal_processed notification after completion, found none")
	}
}

func TestWithdrawal_RejectionProducesSemanticallyCorrectNotification(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST WITHDRAW REJECT USER")
	admin := createTestAdmin(t, env, "TEST WITHDRAW REJECT ADMIN")
	creditWallet(t, env, user.ID, decimal.NewFromInt(2000))
	walletSvc := newWalletSvc(env)
	adminSvc := newTestAdminService(t, env)

	tx, err := walletSvc.RequestWithdraw(ctx, user.ID, decimal.NewFromInt(800), "bankily")
	if err != nil {
		t.Fatalf("(15) RequestWithdraw failed: %v", err)
	}
	if err := adminSvc.ValidateTransaction(ctx, tx.ID, false, "suspicious activity", admin.ID); err != nil {
		t.Fatalf("(15) ValidateTransaction(reject) failed: %v", err)
	}

	// Client feedback #10 (already in place): rejection reuses withdrawal_processed
	// (not a separate withdrawal_rejected type mobile doesn't recognize),
	// distinguished by the {status}/{reason} params baked into the notification
	// body at send time -- so the correct assertion here is that the type fired
	// is still withdrawal_processed, not that a distinct type exists.
	notif := getLatestNotificationOfType(t, env, user.ID, "withdrawal_processed")
	if notif == nil {
		t.Fatalf("(15) expected a withdrawal_processed notification after rejection (reuses the single withdrawal outcome type by design), found none")
	}
}

func TestWithdrawal_RepeatedValidationDoesNotDuplicateNotification(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST WITHDRAW DEDUPE USER")
	admin := createTestAdmin(t, env, "TEST WITHDRAW DEDUPE ADMIN")
	creditWallet(t, env, user.ID, decimal.NewFromInt(2000))
	walletSvc := newWalletSvc(env)
	adminSvc := newTestAdminService(t, env)

	tx, err := walletSvc.RequestWithdraw(ctx, user.ID, decimal.NewFromInt(800), "bankily")
	if err != nil {
		t.Fatalf("(16) RequestWithdraw failed: %v", err)
	}
	if err := adminSvc.ValidateTransaction(ctx, tx.ID, true, "", admin.ID); err != nil {
		t.Fatalf("(16) ValidateTransaction (first call) failed: %v", err)
	}
	_ = adminSvc.ValidateTransaction(ctx, tx.ID, true, "", admin.ID)

	count := countNotificationsOfType(t, env, user.ID, "withdrawal_processed")
	if count != 1 {
		t.Fatalf("(16) expected exactly 1 withdrawal_processed notification after 2 validation calls, got %d", count)
	}
}

// (17-19) Notifications API surface.
func TestNotificationsAPI_ReturnsDepositNotification(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST NOTIF API DEPOSIT USER")
	walletSvc := newWalletSvc(env)

	if _, err := walletSvc.InitiateDeposit(ctx, user.ID, decimal.NewFromInt(400), "bankily", "mobile_money", "", nil); err != nil {
		t.Fatalf("(17) InitiateDeposit failed: %v", err)
	}

	notifs, err := env.notifSvc.ListNotifications(ctx, user.ID, 50)
	if err != nil {
		t.Fatalf("(17) ListNotifications failed: %v", err)
	}
	found := false
	for _, n := range notifs {
		if n.Type == "deposit_submitted" {
			found = true
		}
	}
	if !found {
		t.Fatalf("(17) expected GET /notifications (ListNotifications) to return the deposit_submitted row")
	}
}

func TestNotificationsAPI_ReturnsWithdrawalNotification(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST NOTIF API WITHDRAW USER")
	creditWallet(t, env, user.ID, decimal.NewFromInt(2000))
	walletSvc := newWalletSvc(env)

	if _, err := walletSvc.RequestWithdraw(ctx, user.ID, decimal.NewFromInt(500), "bankily"); err != nil {
		t.Fatalf("(18) RequestWithdraw failed: %v", err)
	}

	notifs, err := env.notifSvc.ListNotifications(ctx, user.ID, 50)
	if err != nil {
		t.Fatalf("(18) ListNotifications failed: %v", err)
	}
	found := false
	for _, n := range notifs {
		if n.Type == "withdrawal_submitted" {
			found = true
		}
	}
	if !found {
		t.Fatalf("(18) expected GET /notifications (ListNotifications) to return the withdrawal_submitted row")
	}
}

func TestNotificationsAPI_UnrelatedUserCannotSeeWalletNotification(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	userA := createTestUser(t, env, "TEST NOTIF API ISOLATION USER A")
	userB := createTestUser(t, env, "TEST NOTIF API ISOLATION USER B")
	walletSvc := newWalletSvc(env)

	if _, err := walletSvc.InitiateDeposit(ctx, userA.ID, decimal.NewFromInt(400), "bankily", "mobile_money", "", nil); err != nil {
		t.Fatalf("(19) InitiateDeposit(A) failed: %v", err)
	}

	notifsB, err := env.notifSvc.ListNotifications(ctx, userB.ID, 50)
	if err != nil {
		t.Fatalf("(19) ListNotifications(B) failed: %v", err)
	}
	for _, n := range notifsB {
		if n.Type == "deposit_submitted" {
			t.Fatalf("(19) CRITICAL: user B's notification list contains user A's deposit_submitted row")
		}
	}
}

// (20-24) Non-regression: existing notification types/paths remain unaffected
// by the migration 000051 constraint widening.
func TestNotificationRegression_AuctionApprovedRemainsValid(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST NOTIF REGRESSION AUCTION APPROVED")
	if err := env.notifSvc.SendLocalizedPush(ctx, user.ID, "auction_approved", "en", map[string]string{"auctionTitle": "Test"}, nil); err != nil {
		t.Fatalf("(20) REGRESSION: auction_approved notification failed after migration 000051: %v", err)
	}
	if getLatestNotificationOfType(t, env, user.ID, "auction_approved") == nil {
		t.Fatalf("(20) REGRESSION: auction_approved notification row not created")
	}
}

func TestNotificationRegression_AuctionWonRemainsValid(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST NOTIF REGRESSION AUCTION WON")
	if err := env.notifSvc.SendLocalizedPush(ctx, user.ID, "auction_won", "en", map[string]string{"auctionTitle": "Test", "finalPrice": "100", "currency": "MRU"}, nil); err != nil {
		t.Fatalf("(21) REGRESSION: auction_won notification failed after migration 000051: %v", err)
	}
	if getLatestNotificationOfType(t, env, user.ID, "auction_won") == nil {
		t.Fatalf("(21) REGRESSION: auction_won notification row not created")
	}
}

func TestNotificationRegression_BidOutbidRemainsValid(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST NOTIF REGRESSION BID OUTBID")
	if err := env.notifSvc.SendLocalizedPush(ctx, user.ID, "bid_outbid", "en", map[string]string{"auctionTitle": "Test", "newPrice": "100", "currency": "MRU"}, nil); err != nil {
		t.Fatalf("(22) REGRESSION: bid_outbid notification failed after migration 000051: %v", err)
	}
	if getLatestNotificationOfType(t, env, user.ID, "bid_outbid") == nil {
		t.Fatalf("(22) REGRESSION: bid_outbid notification row not created")
	}
}

func TestNotificationRegression_AdminBroadcastRemainsValid(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST NOTIF REGRESSION BROADCAST")
	if err := env.notifSvc.SendPush(ctx, user.ID, "Broadcast title", "Broadcast body", "system", nil); err != nil {
		t.Fatalf("(23) REGRESSION: admin broadcast (system type) notification failed after migration 000051: %v", err)
	}
	if getLatestNotificationOfType(t, env, user.ID, "system") == nil {
		t.Fatalf("(23) REGRESSION: system (broadcast) notification row not created")
	}
}

func TestNotificationRegression_APIContractUnchanged(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	user := createTestUser(t, env, "TEST NOTIF REGRESSION API CONTRACT")
	if err := env.notifSvc.SendPush(ctx, user.ID, "Test", "Test body", "system", nil); err != nil {
		t.Fatalf("(24) SendPush failed: %v", err)
	}
	notifs, err := env.notifSvc.ListNotifications(ctx, user.ID, 50)
	if err != nil {
		t.Fatalf("(24) ListNotifications failed: %v", err)
	}
	if len(notifs) == 0 {
		t.Fatalf("(24) REGRESSION: ListNotifications returned no rows for a user with a known notification")
	}
	// Bug C-class contract check: ListNotifications must return a plain slice
	// (not a wrapper), matching handlers.OK(c, notifications)'s existing
	// bare-array response.data contract that notifications_api.dart parses.
	var _ []models.Notification = notifs
}

func strPtr(s string) *string { return &s }
