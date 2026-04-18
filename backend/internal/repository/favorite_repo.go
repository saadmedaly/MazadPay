package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/mazadpay/backend/internal/models"
)

type FavoriteRepository interface {
	Add(ctx context.Context, userID, auctionID uuid.UUID) error
	Remove(ctx context.Context, userID, auctionID uuid.UUID) error
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]models.Auction, error)
	IsFavorite(ctx context.Context, userID, auctionID uuid.UUID) (bool, error)
}

type favoriteRepo struct {
	db *sqlx.DB
}

func NewFavoriteRepository(db *sqlx.DB) FavoriteRepository {
	return &favoriteRepo{db: db}
}

func (r *favoriteRepo) Add(ctx context.Context, userID, auctionID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO user_favorites (user_id, auction_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, userID, auctionID)
	return err
}

func (r *favoriteRepo) Remove(ctx context.Context, userID, auctionID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM user_favorites WHERE user_id = $1 AND auction_id = $2
	`, userID, auctionID)
	return err
}

func (r *favoriteRepo) ListByUserID(ctx context.Context, userID uuid.UUID) ([]models.Auction, error) {
	var auctions []models.Auction
	err := r.db.SelectContext(ctx, &auctions, `
		SELECT a.* FROM auctions a
		JOIN user_favorites f ON a.id = f.auction_id
		WHERE f.user_id = $1
	`, userID)
	return auctions, err
}

func (r *favoriteRepo) IsFavorite(ctx context.Context, userID, auctionID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.GetContext(ctx, &exists, `
		SELECT EXISTS(SELECT 1 FROM user_favorites WHERE user_id = $1 AND auction_id = $2)
	`, userID, auctionID)
	return exists, err
}
