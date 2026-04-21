package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	apperr "github.com/mazadpay/backend/internal/errors"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/repository"
)

type RatingService interface {
	CreateRating(ctx context.Context, userID, auctionID uuid.UUID, rating int, comment string) error
	GetUserRatings(ctx context.Context, userID uuid.UUID) ([]models.AppRating, error)
	GetAuctionRatings(ctx context.Context, auctionID uuid.UUID) ([]models.AppRating, error)
	GetAverageRating(ctx context.Context, userID uuid.UUID) (float64, error)
}

type ratingService struct {
	db       *sqlx.DB
	ratingRepo repository.RatingRepository
}

func NewRatingService(db *sqlx.DB, ratingRepo repository.RatingRepository) RatingService {
	return &ratingService{
		db:          db,
		ratingRepo: ratingRepo,
	}
}

func (s *ratingService) CreateRating(ctx context.Context, userID, auctionID uuid.UUID, rating int, comment string) error {
	// Validate rating (1-5 stars)
	if rating < 1 || rating > 5 {
		return apperr.ErrInvalidRating
	}

	// Check if user won the auction
	auction, err := s.db.QueryRowContext(ctx, `
		SELECT winner_id FROM auctions WHERE id = $1 AND status = 'ended'
	`, auctionID).Scan()
	if err != nil {
		return err
	}

	var winnerID *uuid.UUID
	if auction != nil {
		err := auction.Scan(&winnerID)
		if err != nil {
			return err
		}
	}

	// Only the winner can rate
	if winnerID == nil || *winnerID != userID {
		return apperr.ErrNotAuctionWinner
	}

	// Create rating
	rating := &models.AppRating{
		ID:         uuid.New(),
		UserID:     userID,
		AuctionID:  auctionID,
		Rating:     rating,
		Comment:    &comment,
		CreatedAt:  time.Now(),
	}

	return s.ratingRepo.Create(ctx, rating)
}

func (s *ratingService) GetUserRatings(ctx context.Context, userID uuid.UUID) ([]models.AppRating, error) {
	return s.ratingRepo.FindByUser(ctx, userID)
}

func (s *ratingService) GetAuctionRatings(ctx context.Context, auctionID uuid.UUID) ([]models.AppRating, error) {
	return s.ratingRepo.FindByAuction(ctx, auctionID)
}

func (s *ratingService) GetAverageRating(ctx context.Context, userID uuid.UUID) (float64, error) {
	ratings, err := s.ratingRepo.FindByUser(ctx, userID)
	if err != nil {
		return 0, err
	}

	if len(ratings) == 0 {
		return 0, nil
	}

	var total int
	for _, rating := range ratings {
		total += rating.Rating
	}

	return float64(total) / float64(len(ratings)), nil
}
