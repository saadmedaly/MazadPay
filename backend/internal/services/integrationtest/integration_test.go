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
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/mazadpay/backend/internal/config"
	"github.com/mazadpay/backend/internal/database"
	apperr "github.com/mazadpay/backend/internal/errors"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/repository"
	"github.com/mazadpay/backend/internal/services"
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

	authSvc services.AuthService
	reqSvc  services.RequestService
	auctSvc services.AuctionService
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

	notifSvc := services.NewNotificationService(notifRepo, userRepo, "", "", logger, nil)
	auditSvc := services.NewAuditService(auditRepo)
	mediaSvc := services.NewMediaService(cfg, logger)

	authSvc := services.NewAuthService(userRepo, cfg.JWT.Secret, cfg.JWT.ExpiryHours, "development", "", nil, 4, cfg.Redis.OTPTTLMinutes, rdb, logger)
	auctSvc := services.NewAuctionService(db, auctionRepo, reportRepo, notifSvc, userRepo, mediaSvc, rdb, nil, auditSvc, logger)
	reqSvc := services.NewRequestService(reqRepo, auctionRepo, contentRepo, auditSvc, notifSvc, logger)

	return &testEnv{
		db: db, rdb: rdb, logger: logger,
		userRepo: userRepo, auctionRepo: auctionRepo, reqRepo: reqRepo,
		authSvc: authSvc, reqSvc: reqSvc, auctSvc: auctSvc,
	}
}

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

var _ = fmt.Sprintf // keep fmt import if unused paths change
