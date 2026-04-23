package services

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/mazadpay/backend/internal/models"
	"github.com/shopspring/decimal"
)

type BidAutoBidService interface {
	List(ctx context.Context) ([]models.BidAutoBid, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]models.BidAutoBid, error)
	ListByAuction(ctx context.Context, auctionID uuid.UUID) ([]models.BidAutoBid, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.BidAutoBid, error)
	Create(ctx context.Context, autoBid *models.BidAutoBid) error
	Update(ctx context.Context, id uuid.UUID, autoBid *models.BidAutoBid) error
	Delete(ctx context.Context, id uuid.UUID) error
	ToggleActive(ctx context.Context, id uuid.UUID) error
	ProcessAutoBids(ctx context.Context, auctionID uuid.UUID, currentPrice decimal.Decimal) error
}

type bidAutoBidService struct {
	db        *sqlx.DB
	bidSvc    BidService
	walletSvc WalletService
}

func NewBidAutoBidService(db *sqlx.DB, bidSvc BidService, walletSvc WalletService) BidAutoBidService {
	return &bidAutoBidService{
		db:        db,
		bidSvc:    bidSvc,
		walletSvc: walletSvc,
	}
}

func (s *bidAutoBidService) List(ctx context.Context) ([]models.BidAutoBid, error) {
	var autoBids []models.BidAutoBid
	err := s.db.SelectContext(ctx, &autoBids, `
		SELECT id, user_id, auction_id, max_amount, current_bid_amount, is_active, created_at
		FROM bid_auto_bids
		ORDER BY created_at DESC
	`)
	return autoBids, err
}

func (s *bidAutoBidService) ListByUser(ctx context.Context, userID uuid.UUID) ([]models.BidAutoBid, error) {
	var autoBids []models.BidAutoBid
	err := s.db.SelectContext(ctx, &autoBids, `
		SELECT id, user_id, auction_id, max_amount, current_bid_amount, is_active, created_at
		FROM bid_auto_bids
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	return autoBids, err
}

func (s *bidAutoBidService) ListByAuction(ctx context.Context, auctionID uuid.UUID) ([]models.BidAutoBid, error) {
	var autoBids []models.BidAutoBid
	err := s.db.SelectContext(ctx, &autoBids, `
		SELECT id, user_id, auction_id, max_amount, current_bid_amount, is_active, created_at
		FROM bid_auto_bids
		WHERE auction_id = $1 AND is_active = true
		ORDER BY max_amount DESC
	`, auctionID)
	return autoBids, err
}

func (s *bidAutoBidService) GetByID(ctx context.Context, id uuid.UUID) (*models.BidAutoBid, error) {
	var autoBid models.BidAutoBid
	err := s.db.GetContext(ctx, &autoBid, `
		SELECT id, user_id, auction_id, max_amount, current_bid_amount, is_active, created_at
		FROM bid_auto_bids WHERE id = $1
	`, id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("auto bid not found")
	}
	return &autoBid, err
}

func (s *bidAutoBidService) Create(ctx context.Context, autoBid *models.BidAutoBid) error {
	// Validate max amount
	if autoBid.MaxAmount.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("max amount must be greater than zero")
	}
	
	// Set default values
	autoBid.IsActive = true
	autoBid.CurrentBidAmount = nil
	
	query := `
		INSERT INTO bid_auto_bids 
		(user_id, auction_id, max_amount, current_bid_amount, is_active)
		VALUES (:user_id, :auction_id, :max_amount, :current_bid_amount, :is_active)
		RETURNING id, created_at
	`
	rows, err := s.db.NamedQueryContext(ctx, query, autoBid)
	if err != nil {
		return err
	}
	defer rows.Close()
	
	if rows.Next() {
		return rows.Scan(&autoBid.ID, &autoBid.CreatedAt)
	}
	return nil
}

func (s *bidAutoBidService) Update(ctx context.Context, id uuid.UUID, autoBid *models.BidAutoBid) error {
	query := `
		UPDATE bid_auto_bids
		SET max_amount = :max_amount, is_active = :is_active
		WHERE id = :id
	`
	autoBid.ID = id
	_, err := s.db.NamedExecContext(ctx, query, autoBid)
	return err
}

func (s *bidAutoBidService) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM bid_auto_bids WHERE id = $1`, id)
	return err
}

func (s *bidAutoBidService) ToggleActive(ctx context.Context, id uuid.UUID) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE bid_auto_bids 
		SET is_active = NOT is_active
		WHERE id = $1
	`, id)
	return err
}

// ProcessAutoBids processes all active auto bids for an auction when a new bid is placed
func (s *bidAutoBidService) ProcessAutoBids(ctx context.Context, auctionID uuid.UUID, currentPrice decimal.Decimal) error {
	// Get all active auto bids for this auction, ordered by max amount (highest first)
	autoBids, err := s.ListByAuction(ctx, auctionID)
	if err != nil {
		return err
	}
	
	// Find the highest auto bid that can outbid the current price
	for _, autoBid := range autoBids {
		// Calculate the next bid amount (current price + min increment)
		// For simplicity, using 1% increment or fixed amount
		minIncrement := currentPrice.Mul(decimal.NewFromFloat(0.01))
		nextBidAmount := currentPrice.Add(minIncrement)
		
		// Ensure we don't exceed max_amount
		if nextBidAmount.GreaterThan(autoBid.MaxAmount) {
			nextBidAmount = autoBid.MaxAmount
		}
		
		// Only place bid if it's higher than current and within budget
		if nextBidAmount.GreaterThan(currentPrice) && nextBidAmount.LessThanOrEqual(autoBid.MaxAmount) {
			// TODO: Place the actual bid through bidSvc
			// This would need to be implemented with proper transaction handling
			
			// Update the current_bid_amount
			_, err := s.db.ExecContext(ctx, `
				UPDATE bid_auto_bids 
				SET current_bid_amount = $1
				WHERE id = $2
			`, nextBidAmount, autoBid.ID)
			if err != nil {
				return err
			}
			
			// Update current price for next iteration
			currentPrice = nextBidAmount
		}
	}
	
	return nil
}
