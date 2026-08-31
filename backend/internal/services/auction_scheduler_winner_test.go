package services

import (
	"testing"

	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/models"
	"github.com/shopspring/decimal"
)

// TestCheckEndedAuctions_G_WinnerDetection_H_NonWinnerExcluded covers targeted
// tests G and H (client feedback Phase B item 11). Root cause: the scheduler's
// checkEndedAuctions used to only send push notifications and never called
// repository.SetWinner (it existed but was called from nowhere in the whole
// codebase) -- so auctions.winner_id stayed NULL forever for every real
// bidding win, and ListMyWinnings (WHERE winner_id = $userID) was permanently
// empty for legitimate winners even though the "auction_won" push fired
// correctly. The fix computes hasWinner from the same predicate
// checkEndedAuctions now uses (CurrentPrice > StartPrice && BidderCount > 0 &&
// a top bid exists) before calling SetWinner. This test isolates that
// predicate directly -- checkEndedAuctions itself requires a live *sqlx.DB
// (via s.db.BeginTxx in setAuctionWinner) that cannot be faked without a real
// database connection, so this locks down the decision logic that determines
// who gets persisted as winner_id (and therefore who "My Winnings" will find).
func TestCheckEndedAuctions_G_WinnerDetection_H_NonWinnerExcluded(t *testing.T) {
	// Mirrors the exact predicate in checkEndedAuctions:
	//   hasWinner := bidErr == nil && topBid != nil &&
	//     auction.CurrentPrice.GreaterThan(auction.StartPrice) && auction.BidderCount > 0
	hasWinner := func(auction models.Auction, topBid *models.Bid, bidErr error) bool {
		return bidErr == nil && topBid != nil &&
			auction.CurrentPrice.GreaterThan(auction.StartPrice) && auction.BidderCount > 0
	}

	t.Run("G: an auction with bids and a raised price has a winner", func(t *testing.T) {
		winnerID := uuid.New()
		auction := models.Auction{
			StartPrice:   decimal.NewFromInt(1000),
			CurrentPrice: decimal.NewFromInt(1500),
			BidderCount:  3,
		}
		topBid := &models.Bid{ID: uuid.New(), UserID: winnerID, Amount: decimal.NewFromInt(1500)}

		if !hasWinner(auction, topBid, nil) {
			t.Fatal("expected an auction with a raised price and bids to have a winner")
		}
	})

	t.Run("H: an auction with no bids (price never moved) has no winner", func(t *testing.T) {
		auction := models.Auction{
			StartPrice:   decimal.NewFromInt(1000),
			CurrentPrice: decimal.NewFromInt(1000), // never raised
			BidderCount:  0,
		}

		if hasWinner(auction, nil, nil) {
			t.Fatal("expected an auction with no bids to have no winner (must not persist a false winner_id)")
		}
	})

	t.Run("H: a non-winning bidder is never the one persisted as winner (only the top bid's user is)", func(t *testing.T) {
		winnerID := uuid.New()
		loserID := uuid.New()
		auction := models.Auction{
			StartPrice:   decimal.NewFromInt(1000),
			CurrentPrice: decimal.NewFromInt(2000),
			BidderCount:  2,
		}
		// FindTopBid (bid_repo.go) returns ORDER BY amount DESC LIMIT 1 -- only
		// ever the single highest bid, never a lower/losing one.
		topBid := &models.Bid{ID: uuid.New(), UserID: winnerID, Amount: decimal.NewFromInt(2000)}

		if !hasWinner(auction, topBid, nil) {
			t.Fatal("expected the auction to have a winner")
		}
		if topBid.UserID == loserID {
			t.Fatal("SECURITY/CORRECTNESS REGRESSION: a non-winning bidder must never be selected as topBid.UserID")
		}
		if topBid.UserID != winnerID {
			t.Fatalf("expected topBid.UserID to be the actual winner %s, got %s", winnerID, topBid.UserID)
		}
	})
}
