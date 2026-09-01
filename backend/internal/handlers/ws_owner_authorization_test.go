package handlers

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/repository"
)

// fakeAuctionRepoForWS and fakeUserRepoForWS let a test drive
// WSHandler.AuthorizeSubscription directly (it's exported specifically so it
// can be tested without a live WebSocket connection, per its own doc
// comment) without a database, embedding the real (large) interfaces as nil
// and overriding only FindByID, the one method AuthorizeSubscription calls
// on each.
type fakeAuctionRepoForWS struct {
	repository.AuctionRepository
	auction *models.Auction
}

func (f *fakeAuctionRepoForWS) FindByID(ctx context.Context, id uuid.UUID) (*models.Auction, error) {
	return f.auction, nil
}

type fakeUserRepoForWS struct {
	repository.UserRepository
	users map[uuid.UUID]*models.User
}

func (f *fakeUserRepoForWS) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	u, ok := f.users[id]
	if !ok {
		return nil, context.DeadlineExceeded
	}
	return u, nil
}

// TestWSAuthorizeSubscription_TNOwner_LegacyNullMarketAuction_Allowed is the
// regression test for the WebSocket counterpart of the legacy-owner bug
// (client feedback, round 4): AuthorizeSubscription used to run only
// `user.EffectiveAccountCountryISO() == auction.EffectiveMarketCountryISO()`
// with no owner exemption -- mirroring the same bug the HTTP GetByID handler
// had before its own owner exemption. A TN seller's auction predating
// migration 000046 has market_country_iso = NULL, which
// EffectiveMarketCountryISO() falls back to "MR" for -- so the TN owner's own
// subscription to their own active auction's WebSocket was rejected with
// "WebSocket cross-market or invalid subscription rejected", confirmed live
// against Staging DB data. The owner is now exempt.
func TestWSAuthorizeSubscription_TNOwner_LegacyNullMarketAuction_Allowed(t *testing.T) {
	sellerID := uuid.New()
	auctionID := uuid.New()
	tn := "TN"

	auction := &models.Auction{
		ID:       auctionID,
		SellerID: sellerID,
		Status:   "active",
		// MarketCountryISO/CurrencyCode intentionally nil -- legacy row.
	}

	auctionRepo := &fakeAuctionRepoForWS{auction: auction}
	userRepo := &fakeUserRepoForWS{users: map[uuid.UUID]*models.User{
		sellerID: {ID: sellerID, Role: "user", AccountCountryISO: &tn},
	}}
	h := NewWSHandler(nil, nil, auctionRepo, userRepo, nil)

	if !h.AuthorizeSubscription(context.Background(), auctionID, sellerID.String()) {
		t.Fatal("expected the TN owner to be authorized for their own legacy null-market auction's WebSocket subscription")
	}
}

// TestWSAuthorizeSubscription_TNNonOwner_LegacyNullMarketAuction_Rejected
// proves the fix is narrow: a DIFFERENT TN user (not the seller) subscribing
// to the same legacy NULL-market auction must still be rejected by the
// existing MR-fallback market policy -- the owner exemption must not leak
// into a general "NULL-market auctions are subscribable cross-market"
// loosening.
func TestWSAuthorizeSubscription_TNNonOwner_LegacyNullMarketAuction_Rejected(t *testing.T) {
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

	auctionRepo := &fakeAuctionRepoForWS{auction: auction}
	userRepo := &fakeUserRepoForWS{users: map[uuid.UUID]*models.User{
		otherTNUserID: {ID: otherTNUserID, Role: "user", AccountCountryISO: &tn},
	}}
	h := NewWSHandler(nil, nil, auctionRepo, userRepo, nil)

	if h.AuthorizeSubscription(context.Background(), auctionID, otherTNUserID.String()) {
		t.Fatal("SECURITY REGRESSION: expected a non-owner TN caller to still be rejected from a legacy null-market (MR-fallback) auction's WebSocket")
	}
}

// TestWSAuthorizeSubscription_InvalidUserID_Rejected proves a missing/invalid
// caller identity (what a missing or invalid JWT resolves to earlier in
// HandleAuction, before AuthorizeSubscription is even reached) is still
// rejected -- AuthorizeSubscription itself returns false on a malformed
// userIDStr regardless of any owner logic.
func TestWSAuthorizeSubscription_InvalidUserID_Rejected(t *testing.T) {
	sellerID := uuid.New()
	auctionID := uuid.New()

	auction := &models.Auction{
		ID:       auctionID,
		SellerID: sellerID,
		Status:   "active",
	}

	auctionRepo := &fakeAuctionRepoForWS{auction: auction}
	userRepo := &fakeUserRepoForWS{users: map[uuid.UUID]*models.User{}}
	h := NewWSHandler(nil, nil, auctionRepo, userRepo, nil)

	if h.AuthorizeSubscription(context.Background(), auctionID, "not-a-valid-uuid") {
		t.Fatal("expected an invalid/malformed caller id to be rejected")
	}
}
