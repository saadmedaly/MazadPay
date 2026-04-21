package repository

import (
	"context"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/mazadpay/backend/internal/models"
)

type RatingRepository interface {
	Create(ctx context.Context, rating *models.AppRating) error
	FindByUser(ctx context.Context, userID uuid.UUID) ([]models.AppRating, error)
	FindByAuction(ctx context.Context, auctionID uuid.UUID) ([]models.AppRating, error)
	FindByUserAndAuction(ctx context.Context, userID, auctionID uuid.UUID) (*models.AppRating, error)
}

type ratingRepository struct {
	db *sqlx.DB
}

func NewRatingRepository(db *sqlx.DB) RatingRepository {
	return &ratingRepository{db: db}
}

func (r *ratingRepository) Create(ctx context.Context, rating *models.AppRating) error {
	query := `
		INSERT INTO app_ratings (id, user_id, auction_id, rating, comment, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.ExecContext(ctx, query,
		rating.ID,
		rating.UserID,
		rating.AuctionID,
		rating.Rating,
		rating.Comment,
		rating.CreatedAt,
	)
	return err
}

func (r *ratingRepository) FindByUser(ctx context.Context, userID uuid.UUID) ([]models.AppRating, error) {
	query := `
		SELECT id, user_id, auction_id, rating, comment, created_at
		FROM app_ratings
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	var ratings []models.AppRating
	err := r.db.SelectContext(ctx, &ratings, query, userID)
	return ratings, err
}

func (r *ratingRepository) FindByAuction(ctx context.Context, auctionID uuid.UUID) ([]models.AppRating, error) {
	query := `
		SELECT id, user_id, auction_id, rating, comment, created_at
		FROM app_ratings
		WHERE auction_id = $1
		ORDER BY created_at DESC
	`
	var ratings []models.AppRating
	err := r.db.SelectContext(ctx, &ratings, query, auctionID)
	return ratings, err
}

func (r *ratingRepository) FindByUserAndAuction(ctx context.Context, userID, auctionID uuid.UUID) (*models.AppRating, error) {
	query := `
		SELECT id, user_id, auction_id, rating, comment, created_at
		FROM app_ratings
		WHERE user_id = $1 AND auction_id = $2
		ORDER BY created_at DESC
		LIMIT 1
	`
	var rating models.AppRating
	err := r.db.GetContext(ctx, &rating, query, userID, auctionID)
	if err != nil {
		return nil, err
	}
	return &rating, nil
}
