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
        `SELECT b.*, u.phone as bidder_phone 
         FROM bids b
         JOIN users u ON u.id = b.user_id
         WHERE b.auction_id = $1 
         ORDER BY b.amount DESC, b.created_at DESC`,
        auctionID)
    return bids, err
}

func (r *bidRepo) FindByAuctionID(ctx context.Context, auctionID uuid.UUID) ([]models.Bid, error) {
	var bids []models.Bid
	err := r.db.SelectContext(ctx, &bids, `SELECT * FROM bids WHERE auction_id = $1 ORDER BY amount DESC`, auctionID)
	return bids, err
}

func (r *bidRepo) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM bids")
	return count, err
}

func (r *bidRepo) FindTopBid(ctx context.Context, auctionID uuid.UUID) (*models.Bid, error) {
    var bid models.Bid
    err := r.db.GetContext(ctx, &bid,
        `SELECT * FROM bids WHERE auction_id = $1 ORDER BY amount DESC LIMIT 1`, auctionID)
    if err != nil {
        return nil, err
    }
    return &bid, nil
}

func (r *bidRepo) FindUserBidOnAuction(ctx context.Context, userID, auctionID uuid.UUID) (*models.Bid, error) {
    var bid models.Bid
    err := r.db.GetContext(ctx, &bid,
        `SELECT * FROM bids WHERE user_id = $1 AND auction_id = $2 ORDER BY amount DESC LIMIT 1`,
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
