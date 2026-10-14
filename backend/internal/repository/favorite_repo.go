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

// ListByUserID
//
// Bug I fix (client feedback, Favorites/Home image consistency audit): this
// query used to be a bare `SELECT a.*`, so a.ImageURLs was always NULL --
// unlike FindAll/ListPaginated/findByIDInternal, none of which had this gap
// (see findByIDInternal's own doc comment for the identical bug, fixed there
// already). A favorited auction's real, persisted auction_images rows were
// simply never joined in here, so Favorites showed the Bug H neutral
// placeholder for an auction that Home displayed with its real image
// correctly. Mirrors the exact same COALESCE/string_agg subquery already
// proven by those three call sites -- purely additive, no existing column
// removed/renamed, and still a single query (no N+1: one subquery per row,
// executed by Postgres as part of the same SELECT, not a separate
// round-trip per auction).
func (r *favoriteRepo) ListByUserID(ctx context.Context, userID uuid.UUID) ([]models.Auction, error) {
	var auctions []models.Auction
	err := r.db.SelectContext(ctx, &auctions, `
		SELECT a.*,
		       COALESCE(
		           (SELECT string_agg(url, ',' ORDER BY display_order)
		            FROM auction_images
		            WHERE auction_id = a.id),
		           ''
		       ) as image_urls
		FROM auctions a
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
