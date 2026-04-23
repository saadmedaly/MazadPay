package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/mazadpay/backend/internal/models"
	"github.com/shopspring/decimal"
)

type AuctionBoostService interface {
	List(ctx context.Context) ([]models.AuctionBoost, error)
	ListByAuction(ctx context.Context, auctionID uuid.UUID) ([]models.AuctionBoost, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.AuctionBoost, error)
	Create(ctx context.Context, boost *models.AuctionBoost) error
	Cancel(ctx context.Context, id uuid.UUID) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	GetActiveBoosts(ctx context.Context) ([]models.AuctionBoost, error)
}

type auctionBoostService struct {
	db *sqlx.DB
}

func NewAuctionBoostService(db *sqlx.DB) AuctionBoostService {
	return &auctionBoostService{db: db}
}

func (s *auctionBoostService) List(ctx context.Context) ([]models.AuctionBoost, error) {
	var boosts []models.AuctionBoost
	err := s.db.SelectContext(ctx, &boosts, `
		SELECT id, auction_id, boost_type, start_at, end_at, amount, status, created_at
		FROM auction_boosts
		ORDER BY created_at DESC
	`)
	return boosts, err
}

func (s *auctionBoostService) ListByAuction(ctx context.Context, auctionID uuid.UUID) ([]models.AuctionBoost, error) {
	var boosts []models.AuctionBoost
	err := s.db.SelectContext(ctx, &boosts, `
		SELECT id, auction_id, boost_type, start_at, end_at, amount, status, created_at
		FROM auction_boosts
		WHERE auction_id = $1
		ORDER BY created_at DESC
	`, auctionID)
	return boosts, err
}

func (s *auctionBoostService) GetByID(ctx context.Context, id uuid.UUID) (*models.AuctionBoost, error) {
	var boost models.AuctionBoost
	err := s.db.GetContext(ctx, &boost, `
		SELECT id, auction_id, boost_type, start_at, end_at, amount, status, created_at
		FROM auction_boosts WHERE id = $1
	`, id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("auction boost not found")
	}
	return &boost, err
}

func (s *auctionBoostService) Create(ctx context.Context, boost *models.AuctionBoost) error {
	// Validate dates
	if boost.StartAt.After(boost.EndAt) {
		return fmt.Errorf("start date must be before end date")
	}
	
	// Set default status
	if boost.Status == "" {
		boost.Status = "active"
	}
	
	query := `
		INSERT INTO auction_boosts 
		(auction_id, boost_type, start_at, end_at, amount, status)
		VALUES (:auction_id, :boost_type, :start_at, :end_at, :amount, :status)
		RETURNING id, created_at
	`
	rows, err := s.db.NamedQueryContext(ctx, query, boost)
	if err != nil {
		return err
	}
	defer rows.Close()
	
	if rows.Next() {
		return rows.Scan(&boost.ID, &boost.CreatedAt)
	}
	return nil
}

func (s *auctionBoostService) Cancel(ctx context.Context, id uuid.UUID) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE auction_boosts 
		SET status = 'cancelled'
		WHERE id = $1
	`, id)
	return err
}

func (s *auctionBoostService) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	validStatuses := map[string]bool{
		"active":      true,
		"completed":   true,
		"cancelled":   true,
		"pending":     true,
	}
	
	if !validStatuses[status] {
		return fmt.Errorf("invalid status: %s", status)
	}
	
	_, err := s.db.ExecContext(ctx, `
		UPDATE auction_boosts 
		SET status = $1
		WHERE id = $2
	`, status, id)
	return err
}

func (s *auctionBoostService) GetActiveBoosts(ctx context.Context) ([]models.AuctionBoost, error) {
	var boosts []models.AuctionBoost
	now := time.Now()
	err := s.db.SelectContext(ctx, &boosts, `
		SELECT id, auction_id, boost_type, start_at, end_at, amount, status, created_at
		FROM auction_boosts
		WHERE status = 'active'
		  AND start_at <= $1
		  AND end_at >= $1
		ORDER BY created_at DESC
	`, now)
	return boosts, err
}

// CalculateBoostCost calculates the cost for a boost based on type and duration
func CalculateBoostCost(boostType string, days int) decimal.Decimal {
	baseCosts := map[string]float64{
		"featured": 50.0,
		"urgent":   30.0,
		"top":      100.0,
	}
	
	baseCost, ok := baseCosts[boostType]
	if !ok {
		baseCost = 20.0
	}
	
	return decimal.NewFromFloat(baseCost * float64(days))
}
