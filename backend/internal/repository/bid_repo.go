package repository

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	apperr "github.com/mazadpay/backend/internal/errors"
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
	// ClaimParticipation (client feedback #19): atomically claims this
	// user's one-and-only bid slot on this auction, in the SAME transaction
	// as the bid itself -- see BidService.PlaceBid. Called BEFORE the bid
	// row exists (first_bid_id starts NULL; see AttachFirstBid), so a
	// repeat bidder fails fast before any wallet/insurance work. Returns
	// apperr.ErrDuplicateBidder if uq_auction_bid_participants_auction_user
	// is violated (the user already has a row for this auction); any other
	// error is returned as-is. This is the sole authoritative,
	// concurrency-safe enforcement of the one-bid-per-user rule -- a
	// raced concurrent INSERT from the same user can only ever have one
	// winner at the database level.
	ClaimParticipation(ctx context.Context, tx *sqlx.Tx, auctionID, userID uuid.UUID) error
	// AttachFirstBid records which bid claimed the participation slot, for
	// traceability only -- called right after bidRepo.Create, same
	// transaction. Never read by the eligibility check itself.
	AttachFirstBid(ctx context.Context, tx *sqlx.Tx, auctionID, userID, bidID uuid.UUID) error
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
	err := r.db.SelectContext(ctx, &bids,
		`SELECT b.id, b.auction_id, b.user_id, b.amount, b.previous_price, b.is_winning, b.created_at,
                COALESCE(b.bidder_name, u.full_name) as bidder_name, COALESCE(b.bidder_phone, u.phone) as bidder_phone, b.is_anonymous
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

func (r *bidRepo) ClaimParticipation(ctx context.Context, tx *sqlx.Tx, auctionID, userID uuid.UUID) error {
	// first_bid_id starts NULL: the bid row this claim belongs to doesn't
	// exist yet (the claim intentionally happens before bidRepo.Create, to
	// fail fast on a repeat bidder before any wallet/insurance work) -- see
	// AttachFirstBid, called right after the bid itself is inserted, still
	// inside the same transaction.
	_, err := tx.ExecContext(ctx,
		`INSERT INTO auction_bid_participants (auction_id, user_id) VALUES ($1, $2)`,
		auctionID, userID)
	if err != nil {
		if strings.Contains(err.Error(), "uq_auction_bid_participants_auction_user") {
			return apperr.ErrDuplicateBidder
		}
		return err
	}
	return nil
}

func (r *bidRepo) AttachFirstBid(ctx context.Context, tx *sqlx.Tx, auctionID, userID, bidID uuid.UUID) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE auction_bid_participants SET first_bid_id = $1 WHERE auction_id = $2 AND user_id = $3`,
		bidID, auctionID, userID)
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
