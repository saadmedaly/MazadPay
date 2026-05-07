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
	GetAppStats(ctx context.Context) (float64, int, error)
	ListAllAppRatings(ctx context.Context, page, perPage int) ([]models.AppRating, int, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type ratingRepository struct {
	db *sqlx.DB
}

func NewRatingRepository(db *sqlx.DB) RatingRepository {
	return &ratingRepository{db: db}
}

func (r *ratingRepository) Create(ctx context.Context, rating *models.AppRating) error {
	query := `
		INSERT INTO app_ratings (id, user_id, auction_id, title, rating, comment, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.ExecContext(ctx, query,
		rating.ID,
		rating.UserID,
		rating.AuctionID,
		rating.Title,
		rating.Rating,
		rating.Comment,
		rating.CreatedAt,
	)
	return err
}

func (r *ratingRepository) FindByUser(ctx context.Context, userID uuid.UUID) ([]models.AppRating, error) {
	query := `
		SELECT id, user_id, auction_id, title, rating, comment, created_at
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
		SELECT id, user_id, auction_id, title, rating, comment, created_at
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
		SELECT id, user_id, auction_id, title, rating, comment, created_at
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

func (r *ratingRepository) GetAppStats(ctx context.Context) (float64, int, error) {
	var stats struct {
		AvgRating float64 `db:"avg_rating"`
		Total     int     `db:"total"`
	}
	query := `
		SELECT COALESCE(AVG(rating), 0) as avg_rating, COUNT(*) as total 
		FROM app_ratings 
		WHERE auction_id IS NULL
	`
	err := r.db.GetContext(ctx, &stats, query)
	return stats.AvgRating, stats.Total, err
}

func (r *ratingRepository) ListAllAppRatings(ctx context.Context, page, perPage int) ([]models.AppRating, int, error) {
	var ratings []models.AppRating
	offset := (page - 1) * perPage

	// Get total count
	var total int
	err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM app_ratings WHERE auction_id IS NULL")
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT r.id, r.user_id, r.auction_id, r.title, r.rating, r.comment, r.created_at, u.full_name as user_name
		FROM app_ratings r
		LEFT JOIN users u ON r.user_id = u.id
		WHERE r.auction_id IS NULL
		ORDER BY r.created_at DESC
		LIMIT $1 OFFSET $2
	`
	err = r.db.SelectContext(ctx, &ratings, query, perPage, offset)
	return ratings, total, err
}

func (r *ratingRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM app_ratings WHERE id = $1", id)
	return err
}
