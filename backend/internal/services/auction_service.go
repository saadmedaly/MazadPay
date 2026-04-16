package services

import (
    "context"
    "fmt"
    "time"

    "github.com/google/uuid"
    apperr "github.com/mazadpay/backend/internal/errors"
    "github.com/mazadpay/backend/internal/models"
    "github.com/mazadpay/backend/internal/repository"
    "github.com/shopspring/decimal"
)

type CreateAuctionInput struct {
    CategoryID      int
    LocationID      *int
    Title           string
    Description     string
    StartPrice      decimal.Decimal
    MinIncrement    decimal.Decimal
    InsuranceAmount decimal.Decimal
    EndTime         time.Time
    LotNumber       string
    PhoneContact    string
    ItemDetails     models.JSONB
    BuyNowPrice     *decimal.Decimal
}

type AuctionService interface {
    GetByID(ctx context.Context, id uuid.UUID) (*models.Auction, []models.AuctionImage, error)
    List(ctx context.Context, f repository.AuctionFilters) ([]models.Auction, error)

    Create(ctx context.Context, sellerID uuid.UUID, input CreateAuctionInput) (*models.Auction, error)
    IncrementViews(ctx context.Context, id uuid.UUID) error
    CloseExpiredAuctions(ctx context.Context) error
    GetCategories(ctx context.Context) ([]models.Category, error)
    GetLocations(ctx context.Context) ([]models.Location, error)
}

type auctionService struct {
    auctionRepo repository.AuctionRepository
}

func NewAuctionService(auctionRepo repository.AuctionRepository) AuctionService {
    return &auctionService{auctionRepo: auctionRepo}
}

func (s *auctionService) GetByID(ctx context.Context, id uuid.UUID) (*models.Auction, []models.AuctionImage, error) {
    auction, err := s.auctionRepo.FindByID(ctx, id)
    if err != nil {
        return nil, nil, apperr.ErrNotFound
    }
    images, _ := s.auctionRepo.GetImages(ctx, id)
    return auction, images, nil
}

func (s *auctionService) List(ctx context.Context, f repository.AuctionFilters) ([]models.Auction, error) {
    return s.auctionRepo.FindAll(ctx, f)
}


func (s *auctionService) Create(ctx context.Context, sellerID uuid.UUID, input CreateAuctionInput) (*models.Auction, error) {
    if input.EndTime.Before(time.Now().Add(1 * time.Hour)) {
        return nil, fmt.Errorf("%w: end_time must be at least 1 hour in the future", apperr.ErrBadRequest)
    }

    auction := &models.Auction{
        ID:              uuid.New(),
        SellerID:        sellerID,
        CategoryID:      input.CategoryID,
        LocationID:      input.LocationID,
        Title:           input.Title,
        Description:     &input.Description,
        StartPrice:      input.StartPrice,
        CurrentPrice:    input.StartPrice,
        MinIncrement:    input.MinIncrement,
        InsuranceAmount: input.InsuranceAmount,
        EndTime:         input.EndTime,
        Status:          "pending", // Admin doit valider
        LotNumber:       &input.LotNumber,
        PhoneContact:    &input.PhoneContact,
        ItemDetails:     input.ItemDetails,
        BuyNowPrice:     input.BuyNowPrice,
        Version:         1,
    }

    if err := s.auctionRepo.Create(ctx, nil, auction); err != nil {
        return nil, err
    }
    return auction, nil
}

func (s *auctionService) IncrementViews(ctx context.Context, id uuid.UUID) error {
    return s.auctionRepo.IncrementViews(ctx, id)
}

// CloseExpiredAuctions — appelé par le Cron toutes les 30 secondes
func (s *auctionService) CloseExpiredAuctions(ctx context.Context) error {
    auctions, err := s.auctionRepo.FindExpiredActive(ctx)
    if err != nil {
        return err
    }
    for _, a := range auctions {
        _ = s.auctionRepo.UpdateStatus(ctx, a.ID, "ended")
        // TODO Étape 8 : notifier le gagnant via FCM + WebSocket
    }
    return nil
}

func (s *auctionService) GetCategories(ctx context.Context) ([]models.Category, error) {
    return s.auctionRepo.GetCategories(ctx)
}

func (s *auctionService) GetLocations(ctx context.Context) ([]models.Location, error) {
    return s.auctionRepo.GetLocations(ctx)
}
