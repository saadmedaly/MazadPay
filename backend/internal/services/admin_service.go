package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/repository"
	"github.com/shopspring/decimal"
)

type AdminService interface {
	GetDashboardStats(ctx context.Context) (map[string]interface{}, error)
	ListUsers(ctx context.Context, page, perPage int, query string) ([]models.User, int, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	BlockUser(ctx context.Context, id uuid.UUID, block bool) error
	ListAuctions(ctx context.Context, page, perPage int, status, query string, sellerID *uuid.UUID) ([]models.Auction, int, error)
	ValidateAuction(ctx context.Context, id uuid.UUID, approve bool, reason string) error
	UpdateAuction(ctx context.Context, id uuid.UUID, input UpdateAuctionInput) error
	DeleteAuction(ctx context.Context, id uuid.UUID) error
	ListTransactions(ctx context.Context, page, perPage int, status string, userID *uuid.UUID) ([]models.Transaction, int, error)
	ValidateTransaction(ctx context.Context, id uuid.UUID, approve bool, notes string, adminID uuid.UUID) error
	ListReports(ctx context.Context, page, perPage int, status string) ([]models.Report, int, error)
	ReviewReport(ctx context.Context, id uuid.UUID, status, notes string, adminID uuid.UUID) error

	// KYC
	ListKYC(ctx context.Context, status string) ([]models.KYCVerification, error)
	ReviewKYC(ctx context.Context, userID uuid.UUID, status, notes string, adminID uuid.UUID) error

	// CMS (Banners are already handled separately but can be here too)
	// FAQ/Tutorials CRUD is better in ContentService but Admin can call it

	// Categories & Locations
	CreateCategory(ctx context.Context, c *models.Category) error
	UpdateCategory(ctx context.Context, c *models.Category) error
	DeleteCategory(ctx context.Context, id int) error
	CreateLocation(ctx context.Context, l *models.Location) error
	UpdateLocation(ctx context.Context, l *models.Location) error
	DeleteLocation(ctx context.Context, id int) error
}

type UpdateAuctionInput struct {
	CategoryID      int
	LocationID      *int
	TitleAr         string
	TitleFr         string
	TitleEn         string
	DescriptionAr   string
	DescriptionFr   string
	DescriptionEn   string
	StartPrice      decimal.Decimal
	MinIncrement    decimal.Decimal
	InsuranceAmount decimal.Decimal
	StartTime       *time.Time
	EndTime         time.Time
	PhoneContact    string
	BuyNowPrice     *decimal.Decimal
	ItemDetails     models.JSONB
}

type adminService struct {
	userRepo    repository.UserRepository
	auctionRepo repository.AuctionRepository
	bidRepo     repository.BidRepository
	txRepo      repository.TransactionRepository
	reportRepo  repository.ReportRepository
	kycRepo     repository.KYCRepository
	contentRepo repository.ContentRepository
}

func NewAdminService(
	userRepo repository.UserRepository,
	auctionRepo repository.AuctionRepository,
	bidRepo repository.BidRepository,
	txRepo repository.TransactionRepository,
	reportRepo repository.ReportRepository,
	kycRepo repository.KYCRepository,
	contentRepo repository.ContentRepository,
) AdminService {
	return &adminService{
		userRepo:    userRepo,
		auctionRepo: auctionRepo,
		bidRepo:     bidRepo,
		txRepo:      txRepo,
		reportRepo:  reportRepo,
		kycRepo:     kycRepo,
		contentRepo: contentRepo,
	}
}

func (s *adminService) GetDashboardStats(ctx context.Context) (map[string]interface{}, error) {
	totalUsers, _, err := s.userRepo.GetStats(ctx)
	if err != nil {
		return nil, err
	}
	totalAuctions, activeAuctions, pendingValidations, err := s.auctionRepo.GetStats(ctx)
	if err != nil {
		return nil, err
	}
	totalBids, err := s.bidRepo.Count(ctx)
	if err != nil {
		return nil, err
	}
	totalRevenue, todayRevenue, err := s.txRepo.GetStats(ctx)
	if err != nil {
		return nil, err
	}
	pendingReports, err := s.reportRepo.PendingCount(ctx)
	if err != nil {
		return nil, err
	}
	pendingKYCs, _ := s.kycRepo.List(ctx, "pending")

	return map[string]interface{}{
		"total_users":         totalUsers,
		"total_auctions":      totalAuctions,
		"total_bids":          totalBids,
		"total_revenue":       totalRevenue,
		"today_revenue":       todayRevenue,
		"active_auctions":     activeAuctions,
		"pending_validations": pendingValidations,
		"pending_reports":     pendingReports,
		"pending_kycs":        len(pendingKYCs),
	}, nil
}

func (s *adminService) ListUsers(ctx context.Context, page, perPage int, query string) ([]models.User, int, error) {
	return s.userRepo.ListPaginated(ctx, page, perPage, query)
}

func (s *adminService) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return s.userRepo.FindByID(ctx, id)
}

func (s *adminService) BlockUser(ctx context.Context, id uuid.UUID, block bool) error {
	return s.userRepo.UpdateStatus(ctx, id, !block)
}

func (s *adminService) ListAuctions(ctx context.Context, page, perPage int, status, query string, sellerID *uuid.UUID) ([]models.Auction, int, error) {
	filters := repository.AuctionFilters{
		Status:   status,
		Query:    query,
		SellerID: sellerID,
	}
	return s.auctionRepo.ListPaginated(ctx, page, perPage, filters)
}

func (s *adminService) ValidateAuction(ctx context.Context, id uuid.UUID, approve bool, reason string) error {
	status := "rejected"
	if approve {
		status = "active"
	}
	return s.auctionRepo.UpdateStatus(ctx, id, status)
}

func (s *adminService) UpdateAuction(ctx context.Context, id uuid.UUID, input UpdateAuctionInput) error {
	auction, err := s.auctionRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	var tFr, tEn *string
	var dAr, dFr, dEn *string
	if input.TitleFr != "" { tFr = &input.TitleFr }
	if input.TitleEn != "" { tEn = &input.TitleEn }
	if input.DescriptionAr != "" { dAr = &input.DescriptionAr }
	if input.DescriptionFr != "" { dFr = &input.DescriptionFr }
	if input.DescriptionEn != "" { dEn = &input.DescriptionEn }

	var phone *string
	if input.PhoneContact != "" { phone = &input.PhoneContact }

	auction.CategoryID    = input.CategoryID
	auction.LocationID    = input.LocationID
	auction.TitleAr       = input.TitleAr
	auction.TitleFr       = tFr
	auction.TitleEn       = tEn
	auction.DescriptionAr = dAr
	auction.DescriptionFr = dFr
	auction.DescriptionEn = dEn
	auction.StartPrice    = input.StartPrice
	auction.MinIncrement  = input.MinIncrement
	auction.InsuranceAmount = input.InsuranceAmount
	auction.EndTime       = input.EndTime
	auction.PhoneContact  = phone
	auction.BuyNowPrice   = input.BuyNowPrice
	if input.StartTime != nil {
		auction.StartTime = *input.StartTime
	}
	return s.auctionRepo.Update(ctx, auction)
}

func (s *adminService) DeleteAuction(ctx context.Context, id uuid.UUID) error {
	return s.auctionRepo.Delete(ctx, id)
}

func (s *adminService) ListTransactions(ctx context.Context, page, perPage int, status string, userID *uuid.UUID) ([]models.Transaction, int, error) {
	return s.txRepo.ListPaginated(ctx, page, perPage, status, userID)
}

func (s *adminService) ValidateTransaction(ctx context.Context, id uuid.UUID, approve bool, notes string, adminID uuid.UUID) error {
	status := "rejected"
	if approve {
		status = "completed"
	}
	return s.txRepo.UpdateStatus(ctx, id, status, notes, adminID)
}

func (s *adminService) ListReports(ctx context.Context, page, perPage int, status string) ([]models.Report, int, error) {
	return s.reportRepo.ListPaginated(ctx, page, perPage, status)
}

func (s *adminService) ReviewReport(ctx context.Context, id uuid.UUID, status, notes string, adminID uuid.UUID) error {
	return s.reportRepo.UpdateStatus(ctx, id, status, notes, adminID)
}

func (s *adminService) ListKYC(ctx context.Context, status string) ([]models.KYCVerification, error) {
	return s.kycRepo.List(ctx, status)
}

func (s *adminService) ReviewKYC(ctx context.Context, userID uuid.UUID, status, notes string, adminID uuid.UUID) error {
	if err := s.kycRepo.UpdateStatus(ctx, userID, status, notes, adminID); err != nil {
		return err
	}
	// If approved, mark user as verified
	if status == "approved" {
		user, err := s.userRepo.FindByID(ctx, userID)
		if err == nil {
			return s.userRepo.SetVerified(ctx, user.Phone)
		}
	}
	return nil
}

func (s *adminService) CreateCategory(ctx context.Context, c *models.Category) error {
	return s.auctionRepo.CreateCategory(ctx, c)
}

func (s *adminService) UpdateCategory(ctx context.Context, c *models.Category) error {
	return s.auctionRepo.UpdateCategory(ctx, c)
}

func (s *adminService) DeleteCategory(ctx context.Context, id int) error {
	return s.auctionRepo.DeleteCategory(ctx, id)
}

func (s *adminService) CreateLocation(ctx context.Context, l *models.Location) error {
	return s.auctionRepo.CreateLocation(ctx, l)
}

func (s *adminService) UpdateLocation(ctx context.Context, l *models.Location) error {
	return s.auctionRepo.UpdateLocation(ctx, l)
}

func (s *adminService) DeleteLocation(ctx context.Context, id int) error {
	return s.auctionRepo.DeleteLocation(ctx, id)
}
