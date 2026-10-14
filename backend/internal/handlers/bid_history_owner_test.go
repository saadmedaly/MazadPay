package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/repository"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// fakeBidServiceForHistory, fakeAuctionRepoForBidHistory, and
// fakeUserRepoForBidHistory let a test drive BidHandler.History through the
// real Fiber HTTP cycle without a database, reusing the same
// embed-real-interface-as-nil-and-override pattern established elsewhere in
// this package.
type fakeBidServiceForHistory struct {
	history []models.BidHistoryEntry
}

func (f *fakeBidServiceForHistory) PlaceBid(ctx context.Context, auctionID, userID uuid.UUID, amount decimal.Decimal) (*models.Bid, error) {
	return nil, nil
}

func (f *fakeBidServiceForHistory) GetHistory(ctx context.Context, auctionID uuid.UUID) ([]models.BidHistoryEntry, error) {
	return f.history, nil
}

type fakeAuctionRepoForBidHistory struct {
	repository.AuctionRepository
	auction *models.Auction
}

func (f *fakeAuctionRepoForBidHistory) FindByID(ctx context.Context, id uuid.UUID) (*models.Auction, error) {
	return f.auction, nil
}

type fakeUserRepoForBidHistory struct {
	repository.UserRepository
	users map[uuid.UUID]*models.User
}

func (f *fakeUserRepoForBidHistory) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	u, ok := f.users[id]
	if !ok {
		return nil, context.DeadlineExceeded
	}
	return u, nil
}

// TestBidHistory_TNOwner_LegacyNullMarketAuction_Allowed is the regression
// test for the bid-history counterpart of the legacy-owner bug (client
// feedback, round 5): GET /auctions/:id/bids enforced country-scoped market
// isolation with no owner exemption -- the third occurrence of the same bug
// already fixed for GET /auctions/:id and the WebSocket subscription. A
// legacy auction predating migration 000046 has market_country_iso = NULL,
// which falls back to "MR"; a TN owner's own bid history request for their
// own active auction was therefore 404'd, confirmed live against Staging
// (detail + WebSocket already worked for this exact auction/user, this
// endpoint was the remaining gap in the repeated live logs).
func TestBidHistory_TNOwner_LegacyNullMarketAuction_Allowed(t *testing.T) {
	sellerID := uuid.New()
	auctionID := uuid.New()
	tn := "TN"

	auction := &models.Auction{
		ID:       auctionID,
		SellerID: sellerID,
		Status:   "active",
		// MarketCountryISO/CurrencyCode intentionally nil -- legacy row.
	}

	svc := &fakeBidServiceForHistory{history: []models.BidHistoryEntry{}}
	auctionRepo := &fakeAuctionRepoForBidHistory{auction: auction}
	userRepo := &fakeUserRepoForBidHistory{users: map[uuid.UUID]*models.User{
		sellerID: {ID: sellerID, Role: "user", AccountCountryISO: &tn},
	}}
	h := NewBidHandler(svc, auctionRepo, userRepo, zap.NewNop())

	app := fiber.New()
	app.Get("/v1/api/auctions/:id/bids", func(c *fiber.Ctx) error {
		c.Locals("user_id", sellerID)
		return h.History(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/api/auctions/"+auctionID.String()+"/bids", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected the TN owner to retrieve their own legacy null-market auction's bid history (HTTP 200), got %d", resp.StatusCode)
	}
}

// TestBidHistory_TNNonOwner_LegacyNullMarketAuction_StillBlocked proves the
// fix is narrow: a DIFFERENT TN user (not the seller) requesting bid history
// for the same legacy NULL-market auction must still be blocked by the
// existing MR-fallback market policy.
func TestBidHistory_TNNonOwner_LegacyNullMarketAuction_StillBlocked(t *testing.T) {
	sellerID := uuid.New()
	otherTNUserID := uuid.New()
	auctionID := uuid.New()
	tn := "TN"

	auction := &models.Auction{
		ID:       auctionID,
		SellerID: sellerID,
		Status:   "active",
		// MarketCountryISO/CurrencyCode intentionally nil, resolves to "MR".
	}

	svc := &fakeBidServiceForHistory{history: []models.BidHistoryEntry{}}
	auctionRepo := &fakeAuctionRepoForBidHistory{auction: auction}
	userRepo := &fakeUserRepoForBidHistory{users: map[uuid.UUID]*models.User{
		otherTNUserID: {ID: otherTNUserID, Role: "user", AccountCountryISO: &tn},
	}}
	h := NewBidHandler(svc, auctionRepo, userRepo, zap.NewNop())

	app := fiber.New()
	app.Get("/v1/api/auctions/:id/bids", func(c *fiber.Ctx) error {
		c.Locals("user_id", otherTNUserID)
		return h.History(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/api/auctions/"+auctionID.String()+"/bids", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("SECURITY REGRESSION: expected a non-owner TN caller to still be blocked from a legacy null-market auction's bid history, got %d", resp.StatusCode)
	}
}

// TestBidHistory_Anonymous_LegacyNullMarketAuction_StillAllowedByMRFallback
// proves anonymous behavior is unchanged: an anonymous caller resolves to
// DefaultAccountCountryISO ("MR"), which already equals a legacy NULL-market
// auction's own MR fallback -- so anonymous access to a legacy auction's bid
// history was never blocked and remains unaffected by this fix.
func TestBidHistory_Anonymous_LegacyNullMarketAuction_StillAllowedByMRFallback(t *testing.T) {
	sellerID := uuid.New()
	auctionID := uuid.New()

	auction := &models.Auction{
		ID:       auctionID,
		SellerID: sellerID,
		Status:   "active",
		// MarketCountryISO/CurrencyCode intentionally nil, resolves to "MR",
		// matching the anonymous caller's own DefaultAccountCountryISO fallback.
	}

	svc := &fakeBidServiceForHistory{history: []models.BidHistoryEntry{}}
	auctionRepo := &fakeAuctionRepoForBidHistory{auction: auction}
	userRepo := &fakeUserRepoForBidHistory{users: map[uuid.UUID]*models.User{}}
	h := NewBidHandler(svc, auctionRepo, userRepo, zap.NewNop())

	app := fiber.New()
	app.Get("/v1/api/auctions/:id/bids", h.History) // no c.Locals("user_id", ...)

	req := httptest.NewRequest(http.MethodGet, "/v1/api/auctions/"+auctionID.String()+"/bids", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected anonymous access to a legacy MR-fallback auction's bid history to remain unaffected (HTTP 200), got %d", resp.StatusCode)
	}
}
