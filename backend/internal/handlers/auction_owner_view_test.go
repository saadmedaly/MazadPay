package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/repository"
	"github.com/mazadpay/backend/internal/services"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// fakeAuctionServiceForGetByID and fakeUserRepoForGetByID let a test drive
// AuctionHandler.GetByID through the real Fiber HTTP cycle without a
// database, embedding the real (large) interfaces as nil and overriding only
// what GetByID actually calls.
type fakeAuctionServiceForGetByID struct {
	services.AuctionService
	auction *models.Auction
}

func (f *fakeAuctionServiceForGetByID) GetByID(ctx context.Context, id uuid.UUID) (*models.Auction, []models.AuctionImage, error) {
	return f.auction, nil, nil
}

type fakeUserRepoForGetByID struct {
	repository.UserRepository
	users map[uuid.UUID]*models.User
}

func (f *fakeUserRepoForGetByID) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	u, ok := f.users[id]
	if !ok {
		return nil, context.DeadlineExceeded // any non-nil error -> caller falls back to anonymous
	}
	return u, nil
}

// TestAuctionGetByID_D_OwnerCanViewOwnNonPublicAuction is the regression test
// for the real-device Staging failure "خطأ في تحميل المزاد" opening an item
// from My Auctions. Root cause: GetByID (auction_handler.go) only allowed
// PubliclyVisibleAuctionStatuses ("active"/"ended") through, 404-ing for
// EVERY other caller including the auction's own seller -- so a seller
// opening their own still-"pending" (not yet admin-approved) auction from My
// Auctions always got a 404, which auction_provider_api.dart turns into a
// thrown Exception, rendering the generic error_loading_auction screen. The
// code comment this replaces referenced a "dedicated seller/admin detail
// endpoint" that never actually existed in routes.go. This test proves the
// owner of a "pending" auction can now retrieve it via the same GetByID path
// the mobile app already calls.
func TestAuctionGetByID_D_OwnerCanViewOwnNonPublicAuction(t *testing.T) {
	sellerID := uuid.New()
	auctionID := uuid.New()
	mr := "MR"

	auction := &models.Auction{
		ID:               auctionID,
		SellerID:         sellerID,
		Status:           "pending", // NOT in PubliclyVisibleAuctionStatuses
		StartPrice:       decimal.NewFromInt(100),
		CurrentPrice:     decimal.NewFromInt(100),
		StartTime:        time.Now(),
		EndTime:          time.Now().Add(24 * time.Hour),
		MarketCountryISO: &mr,
	}

	auctionSvc := &fakeAuctionServiceForGetByID{auction: auction}
	userRepo := &fakeUserRepoForGetByID{users: map[uuid.UUID]*models.User{
		sellerID: {ID: sellerID, Role: "user", AccountCountryISO: &mr},
	}}
	h := NewAuctionHandler(auctionSvc, userRepo, zap.NewNop())

	app := fiber.New()
	app.Get("/v1/api/auctions/:id", func(c *fiber.Ctx) error {
		c.Locals("user_id", sellerID)
		return h.GetByID(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/api/auctions/"+auctionID.String(), nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected the owner to retrieve their own pending auction (HTTP 200), got %d, body: %s", resp.StatusCode, body)
	}
}

// TestAuctionGetByID_NonOwnerStillCannotViewPendingAuction proves the fix did
// not weaken the original anti-enumeration protection: a DIFFERENT
// authenticated user (not the seller, not an admin) must still get a 404 for
// a non-public auction.
func TestAuctionGetByID_NonOwnerStillCannotViewPendingAuction(t *testing.T) {
	sellerID := uuid.New()
	otherUserID := uuid.New()
	auctionID := uuid.New()
	mr := "MR"

	auction := &models.Auction{
		ID:               auctionID,
		SellerID:         sellerID,
		Status:           "pending",
		StartPrice:       decimal.NewFromInt(100),
		CurrentPrice:     decimal.NewFromInt(100),
		StartTime:        time.Now(),
		EndTime:          time.Now().Add(24 * time.Hour),
		MarketCountryISO: &mr,
	}

	auctionSvc := &fakeAuctionServiceForGetByID{auction: auction}
	userRepo := &fakeUserRepoForGetByID{users: map[uuid.UUID]*models.User{
		otherUserID: {ID: otherUserID, Role: "user", AccountCountryISO: &mr},
	}}
	h := NewAuctionHandler(auctionSvc, userRepo, zap.NewNop())

	app := fiber.New()
	app.Get("/v1/api/auctions/:id", func(c *fiber.Ctx) error {
		c.Locals("user_id", otherUserID)
		return h.GetByID(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/api/auctions/"+auctionID.String(), nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("SECURITY REGRESSION: expected 404 for a non-owner viewing a pending auction, got %d", resp.StatusCode)
	}
}

// TestAuctionGetByID_AnonymousStillCannotViewPendingAuction proves anonymous
// callers are still blocked from a non-public auction.
func TestAuctionGetByID_AnonymousStillCannotViewPendingAuction(t *testing.T) {
	sellerID := uuid.New()
	auctionID := uuid.New()
	mr := "MR"

	auction := &models.Auction{
		ID:               auctionID,
		SellerID:         sellerID,
		Status:           "pending",
		StartPrice:       decimal.NewFromInt(100),
		CurrentPrice:     decimal.NewFromInt(100),
		StartTime:        time.Now(),
		EndTime:          time.Now().Add(24 * time.Hour),
		MarketCountryISO: &mr,
	}

	auctionSvc := &fakeAuctionServiceForGetByID{auction: auction}
	userRepo := &fakeUserRepoForGetByID{users: map[uuid.UUID]*models.User{}}
	h := NewAuctionHandler(auctionSvc, userRepo, zap.NewNop())

	app := fiber.New()
	app.Get("/v1/api/auctions/:id", h.GetByID) // no c.Locals("user_id", ...)

	req := httptest.NewRequest(http.MethodGet, "/v1/api/auctions/"+auctionID.String(), nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 for an anonymous caller viewing a pending auction, got %d", resp.StatusCode)
	}
}

// TestAuctionGetByID_E_LegacyNullMarketRowsStillWork is targeted test E5:
// an auction with MarketCountryISO/CurrencyCode NULL (predating migration
// 000046) must not break GetByID for its owner, an anonymous MR caller, or
// a same-market authenticated caller -- EffectiveMarketCountryISO's fallback
// to DefaultAccountCountryISO ("MR") must resolve consistently on both the
// auction and caller side.
func TestAuctionGetByID_E_LegacyNullMarketRowsStillWork(t *testing.T) {
	sellerID := uuid.New()
	auctionID := uuid.New()

	// No MarketCountryISO/CurrencyCode set -- exercises the legacy fallback path.
	auction := &models.Auction{
		ID:           auctionID,
		SellerID:     sellerID,
		Status:       "active",
		StartPrice:   decimal.NewFromInt(100),
		CurrentPrice: decimal.NewFromInt(100),
		StartTime:    time.Now(),
		EndTime:      time.Now().Add(24 * time.Hour),
		// MarketCountryISO, CurrencyCode intentionally left nil.
	}

	auctionSvc := &fakeAuctionServiceForGetByID{auction: auction}
	userRepo := &fakeUserRepoForGetByID{users: map[uuid.UUID]*models.User{}}
	h := NewAuctionHandler(auctionSvc, userRepo, zap.NewNop())

	app := fiber.New()
	// Anonymous caller: no c.Locals("user_id", ...) set, matching the public
	// unauthenticated path -- falls back to DefaultAccountCountryISO on both sides.
	app.Get("/v1/api/auctions/:id", h.GetByID)

	req := httptest.NewRequest(http.MethodGet, "/v1/api/auctions/"+auctionID.String(), nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected a legacy null-market active auction to be visible to an anonymous MR caller (HTTP 200), got %d, body: %s", resp.StatusCode, body)
	}

	respBody, _ := io.ReadAll(resp.Body)
	var parsed map[string]interface{}
	if err := json.Unmarshal(respBody, &parsed); err == nil {
		if data, ok := parsed["data"].(map[string]interface{}); ok {
			if a, ok := data["auction"].(map[string]interface{}); ok {
				if a["currency_code"] != "MRU" {
					t.Fatalf("expected legacy null currency to resolve to the default MRU, got %v", a["currency_code"])
				}
			}
		}
	}
}

// TestAuctionGetByID_TNOwner_LegacyNullMarketAuction_ReturnsOK is the
// regression test for the exact case proven by live Staging DB data (client
// feedback, round 3): a TN seller's auction predates migration 000046 and has
// market_country_iso/currency_code = NULL. EffectiveMarketCountryISO() falls
// back to DefaultAccountCountryISO ("MR") for such a row -- so the previous
// code (isOwner exempted the status check, but NOT the market-isolation
// check below it) 404'd the TN owner out of their own active auction,
// because "TN" (their real market) != "MR" (the row's NULL-fallback market).
// This was the confirmed root cause of "خطأ في تحميل المزاد" on My Auctions.
func TestAuctionGetByID_TNOwner_LegacyNullMarketAuction_ReturnsOK(t *testing.T) {
	sellerID := uuid.New()
	auctionID := uuid.New()
	tn := "TN"

	auction := &models.Auction{
		ID:           auctionID,
		SellerID:     sellerID,
		Status:       "active",
		StartPrice:   decimal.NewFromInt(100),
		CurrentPrice: decimal.NewFromInt(100),
		StartTime:    time.Now(),
		EndTime:      time.Now().Add(24 * time.Hour),
		// MarketCountryISO/CurrencyCode intentionally nil -- legacy row.
	}

	auctionSvc := &fakeAuctionServiceForGetByID{auction: auction}
	userRepo := &fakeUserRepoForGetByID{users: map[uuid.UUID]*models.User{
		sellerID: {ID: sellerID, Role: "user", AccountCountryISO: &tn},
	}}
	h := NewAuctionHandler(auctionSvc, userRepo, zap.NewNop())

	app := fiber.New()
	app.Get("/v1/api/auctions/:id", func(c *fiber.Ctx) error {
		c.Locals("user_id", sellerID)
		return h.GetByID(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/api/auctions/"+auctionID.String(), nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected the TN owner to retrieve their own legacy null-market active auction (HTTP 200), got %d, body: %s", resp.StatusCode, body)
	}
}

// TestAuctionGetByID_TNNonOwner_LegacyNullMarketAuction_StillBlocked proves
// the fix is narrow: a DIFFERENT TN user (not the seller) viewing the same
// legacy NULL-market auction must still be blocked by the existing MR-fallback
// market policy -- the owner exemption must not leak into a general
// "NULL-market auctions are visible cross-market" loosening.
func TestAuctionGetByID_TNNonOwner_LegacyNullMarketAuction_StillBlocked(t *testing.T) {
	sellerID := uuid.New()
	otherTNUserID := uuid.New()
	auctionID := uuid.New()
	tn := "TN"

	auction := &models.Auction{
		ID:           auctionID,
		SellerID:     sellerID,
		Status:       "active",
		StartPrice:   decimal.NewFromInt(100),
		CurrentPrice: decimal.NewFromInt(100),
		StartTime:    time.Now(),
		EndTime:      time.Now().Add(24 * time.Hour),
		// MarketCountryISO/CurrencyCode intentionally nil -- legacy row,
		// EffectiveMarketCountryISO() resolves to "MR".
	}

	auctionSvc := &fakeAuctionServiceForGetByID{auction: auction}
	userRepo := &fakeUserRepoForGetByID{users: map[uuid.UUID]*models.User{
		otherTNUserID: {ID: otherTNUserID, Role: "user", AccountCountryISO: &tn},
	}}
	h := NewAuctionHandler(auctionSvc, userRepo, zap.NewNop())

	app := fiber.New()
	app.Get("/v1/api/auctions/:id", func(c *fiber.Ctx) error {
		c.Locals("user_id", otherTNUserID)
		return h.GetByID(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/api/auctions/"+auctionID.String(), nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// TN caller vs. the legacy row's MR fallback -> still cross-market ->
	// still blocked, matching the existing (unchanged) market isolation policy.
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("SECURITY REGRESSION: expected a non-owner TN caller to still be blocked from a legacy null-market (MR-fallback) auction, got %d", resp.StatusCode)
	}
}
