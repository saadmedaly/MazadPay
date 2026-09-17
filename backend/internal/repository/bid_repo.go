package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/mazadpay/backend/internal/models"
)

type BidRepository interface {
	Create(ctx context.Context, tx *sqlx.Tx, bid *models.Bid) error
	FindByAuction(ctx context.Context, auctionID uuid.UUID) ([]models.Bid, error)
	FindHistoryByAuction(ctx context.Context, auctionID uuid.UUID) ([]models.BidHistoryEntry, error)
	FindTopBid(ctx context.Context, auctionID uuid.UUID) (*models.Bid, error)
	FindUserBidOnAuction(ctx context.Context, userID, auctionID uuid.UUID) (*models.Bid, error)
	SetAllNotWinning(ctx context.Context, tx *sqlx.Tx, auctionID uuid.UUID) error
	FindUserActiveBids(ctx context.Context, userID uuid.UUID) ([]models.Bid, error)
	Count(ctx context.Context) (int, error)
}

type bidRepo struct{ db *sqlx.DB }

func NewBidRepository(db *sqlx.DB) BidRepository {
	return &bidRepo{db: db}
}

func (r *bidRepo) Create(ctx context.Context, tx *sqlx.Tx, bid *models.Bid) error {
	_, err := tx.NamedExecContext(ctx, `
        INSERT INTO bids (id, auction_id, user_id, amount, previous_price, is_winning)
        VALUES (:id, :auction_id, :user_id, :amount, :previous_price, :is_winning)
    `, bid)
	return err
}

func (r *bidRepo) FindByAuction(ctx context.Context, auctionID uuid.UUID) ([]models.Bid, error) {
	var bids []models.Bid
	err := r.db.SelectContext(ctx, &bids,
		`SELECT b.id, b.auction_id, b.user_id, b.amount, b.previous_price, b.is_winning, b.created_at
         FROM bids b
         WHERE b.auction_id = $1
         ORDER BY b.amount DESC, b.created_at DESC`,
		auctionID)
	return bids, err
}

func (r *bidRepo) FindHistoryByAuction(ctx context.Context, auctionID uuid.UUID) ([]models.BidHistoryEntry, error) {
	var bids []models.BidHistoryEntry
	// Bug M fix: COALESCE(b.bidder_name, u.full_name) alone can still yield
	// SQL NULL when BOTH the bid's own denormalized bidder_name and the
	// joined user's full_name are NULL (a real, reachable state -- a user
	// with no full_name set placing a bid). BidHistoryEntry.BidderName/
	// BidderPhone are non-nullable Go strings, so that NULL failed the scan
	// entirely ("converting NULL to string is unsupported"), returning an
	// error for the WHOLE history query even though real bid rows existed --
	// mobile's My Winnings/bid-history provider then silently swallowed the
	// error and rendered an empty list despite bid_count > 0. A final ''
	// fallback (matching the existing pattern in FindByAuctionID below)
	// guarantees a non-NULL string in every case; the mobile UI already has
	// its own safe empty-name fallback for this.
	err := r.db.SelectContext(ctx, &bids,
		`SELECT b.id, b.auction_id, b.user_id, b.amount, b.previous_price, b.is_winning, b.created_at,
                COALESCE(b.bidder_name, u.full_name, '') as bidder_name, COALESCE(b.bidder_phone, u.phone, '') as bidder_phone, b.is_anonymous
         FROM bids b
         LEFT JOIN users u ON u.id = b.user_id
         WHERE b.auction_id = $1
         ORDER BY b.amount DESC, b.created_at DESC`,
		auctionID)
	return bids, err
}

func (r *bidRepo) FindByAuctionID(ctx context.Context, auctionID uuid.UUID) ([]models.Bid, error) {
	var bids []models.Bid
	// bidder_name/bidder_phone are legacy denormalized columns that are
	// often NULL; models.Bid scans them as non-nullable string, so a raw
	// SELECT * fails here whenever either is NULL (client feedback #19
	// hardening: this was blocking GetBidStatus's has_bid for a real bid).
	err := r.db.SelectContext(ctx, &bids,
		`SELECT id, auction_id, user_id, amount, previous_price, is_winning,
                COALESCE(bidder_name, '') as bidder_name, COALESCE(bidder_phone, '') as bidder_phone,
                is_anonymous, created_at
         FROM bids WHERE auction_id = $1 ORDER BY amount DESC`, auctionID)
	return bids, err
}

func (r *bidRepo) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM bids")
	return count, err
}

func (r *bidRepo) FindTopBid(ctx context.Context, auctionID uuid.UUID) (*models.Bid, error) {
	var bid models.Bid
	// See FindByAuctionID: bidder_name/bidder_phone can be NULL.
	err := r.db.GetContext(ctx, &bid,
		`SELECT id, auction_id, user_id, amount, previous_price, is_winning,
                COALESCE(bidder_name, '') as bidder_name, COALESCE(bidder_phone, '') as bidder_phone,
                is_anonymous, created_at
         FROM bids WHERE auction_id = $1 ORDER BY amount DESC LIMIT 1`, auctionID)
	if err != nil {
		return nil, err
	}
	return &bid, nil
}

func (r *bidRepo) FindUserBidOnAuction(ctx context.Context, userID, auctionID uuid.UUID) (*models.Bid, error) {
	var bid models.Bid
	// See FindByAuctionID: bidder_name/bidder_phone can be NULL.
	err := r.db.GetContext(ctx, &bid,
		`SELECT id, auction_id, user_id, amount, previous_price, is_winning,
                COALESCE(bidder_name, '') as bidder_name, COALESCE(bidder_phone, '') as bidder_phone,
                is_anonymous, created_at
         FROM bids WHERE user_id = $1 AND auction_id = $2 ORDER BY amount DESC LIMIT 1`,
		userID, auctionID)
	if err != nil {
		return nil, err
	}
	return &bid, nil
}

func (r *bidRepo) SetAllNotWinning(ctx context.Context, tx *sqlx.Tx, auctionID uuid.UUID) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE bids SET is_winning = false WHERE auction_id = $1`, auctionID)
	return err
}

func (r *bidRepo) FindUserActiveBids(ctx context.Context, userID uuid.UUID) ([]models.Bid, error) {
	var bids []models.Bid
	err := r.db.SelectContext(ctx, &bids, `
        SELECT b.* FROM bids b
        JOIN auctions a ON a.id = b.auction_id
        WHERE b.user_id = $1 AND a.status IN ('active', 'pending')
        ORDER BY b.created_at DESC
    `, userID)
	return bids, err
}
