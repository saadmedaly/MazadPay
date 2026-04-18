package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sort"
	"time"
	"golang.org/x/crypto/bcrypt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/repository"
	"github.com/shopspring/decimal"
)

type AdminService interface {
	GetDashboardStats(ctx context.Context) (map[string]interface{}, error)
	GetRevenueChartData(ctx context.Context) ([]map[string]interface{}, error)
	GetRecentActivity(ctx context.Context) ([]map[string]interface{}, error)
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

	// Admin Management
	CreateAdmin(ctx context.Context, phone, pin, fullName, email string) error
	GenerateAdminInvitation(ctx context.Context, createdBy uuid.UUID) (string, error)
	RegisterAdminWithToken(ctx context.Context, token, phone, pin, fullName, email string) error

	// Blocked Phones
	ListBlockedPhones(ctx context.Context) ([]map[string]interface{}, error)
	BlockPhone(ctx context.Context, phone, reason string, blockedBy uuid.UUID) error
	UnblockPhone(ctx context.Context, phone string) error

	// Settings
	ListSettings(ctx context.Context) ([]models.SystemSettings, error)
	UpdateSetting(ctx context.Context, key, value, settingType string, userID uuid.UUID) error
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
	Images          []string
}

type adminService struct {
	db          *sqlx.DB
	userRepo    repository.UserRepository
	auctionRepo repository.AuctionRepository
	bidRepo     repository.BidRepository
	txRepo      repository.TransactionRepository
	reportRepo  repository.ReportRepository
	kycRepo     repository.KYCRepository
	contentRepo repository.ContentRepository
	invRepo     repository.AdminInvitationRepository
}

func NewAdminService(
	db *sqlx.DB,
	userRepo repository.UserRepository,
	auctionRepo repository.AuctionRepository,
	bidRepo repository.BidRepository,
	txRepo repository.TransactionRepository,
	reportRepo repository.ReportRepository,
	kycRepo repository.KYCRepository,
	contentRepo repository.ContentRepository,
	invRepo repository.AdminInvitationRepository,
) AdminService {
	return &adminService{
		db:          db,
		userRepo:    userRepo,
		auctionRepo: auctionRepo,
		bidRepo:     bidRepo,
		txRepo:      txRepo,
		reportRepo:  reportRepo,
		kycRepo:     kycRepo,
		contentRepo: contentRepo,
		invRepo:     invRepo,
	}
}

var _ AdminService = (*adminService)(nil)

func (s *adminService) GetDashboardStats(ctx context.Context) (map[string]interface{}, error) {
	totalUsers, _, err := s.userRepo.GetStats(ctx)
	if err != nil {
		return nil, err
	}
	totalAuctions, activeAuctions, pendingAuctions, err := s.auctionRepo.GetStats(ctx)
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

	pendingTransactions, _ := s.txRepo.GetPendingCount(ctx)
	weekDeposits, _ := s.txRepo.GetWeeklySum(ctx)

	return map[string]interface{}{
		"total_users":          totalUsers,
		"total_auctions":       totalAuctions,
		"total_bids":           totalBids,
		"total_revenue":        totalRevenue,
		"today_revenue":        todayRevenue,
		"active_auctions":      activeAuctions,
		"pending_auctions":     pendingAuctions,
		"pending_reports":      pendingReports,
		"pending_kycs":         len(pendingKYCs),
		"pending_transactions": pendingTransactions,
		"week_deposits":        weekDeposits,
	}, nil
}

func (s *adminService) GetRevenueChartData(ctx context.Context) ([]map[string]interface{}, error) {
	return s.txRepo.GetDailyRevenueChart(ctx)
}

func (s *adminService) GetRecentActivity(ctx context.Context) ([]map[string]interface{}, error) {
	var activities []map[string]interface{}

	// Latest 5 Auctions
	auctions, _, _ := s.auctionRepo.ListPaginated(ctx, 1, 5, repository.AuctionFilters{})
	for _, a := range auctions {
		activities = append(activities, map[string]interface{}{
			"id":          "auc_" + a.ID.String(),
			"type":        "auction",
			"description": "مزاد جديد: " + a.TitleAr,
			"created_at":  a.CreatedAt,
		})
	}

	// Latest 5 Transactions
	txs, _, _ := s.txRepo.ListPaginated(ctx, 1, 5, "", nil)
	for _, t := range txs {
		activities = append(activities, map[string]interface{}{
			"id":          "tx_" + t.ID.String(),
			"type":        "transaction",
			"description": "عملية مالية جديدة بقيمة " + t.Amount.String() + " MRU",
			"created_at":  t.CreatedAt,
		})
	}

	// Sort by created_at descending
	sort.Slice(activities, func(i, j int) bool {
		ti := activities[i]["created_at"].(time.Time)
		tj := activities[j]["created_at"].(time.Time)
		return ti.After(tj)
	})

	// Limit to 10
	if len(activities) > 10 {
		activities = activities[:10]
	}

	return activities, nil
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
	auction.ItemDetails   = input.ItemDetails
	if input.StartTime != nil {
		auction.StartTime = *input.StartTime
	}
	
	if err := s.auctionRepo.Update(ctx, auction); err != nil {
		return err
	}

	// Sync images: Delete existing and add new ones
	if err := s.auctionRepo.DeleteImages(ctx, id); err != nil {
		return err
	}
	for i, url := range input.Images {
		if url == "" { continue }
		err := s.auctionRepo.AddImage(ctx, &models.AuctionImage{
			AuctionID:    id,
			URL:          url,
			MediaType:    "image",
			DisplayOrder: i,
		})
		if err != nil {
			return err
		}
	}

	return nil
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

func (s *adminService) CreateAdmin(ctx context.Context, phone, pin, fullName, email string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(pin), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &models.User{
		ID:           uuid.New(),
		Phone:        phone,
		PasswordHash: string(hash),
		FullName:     &fullName,
		Email:        &email,
		LanguagePref: "ar",
		Role:         "admin",
		IsVerified:   true, // Admins created by admins are verified
	}

	return s.userRepo.Create(ctx, user)
}

func (s *adminService) GenerateAdminInvitation(ctx context.Context, createdBy uuid.UUID) (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)

	inv := &models.AdminInvitation{
		ID:        uuid.New(),
		Token:     token,
		CreatedBy: createdBy,
		ExpiresAt: time.Now().Add(24 * time.Hour), // 24h validity
	}

	if err := s.invRepo.Create(ctx, inv); err != nil {
		return "", err
	}

	return token, nil
}

func (s *adminService) RegisterAdminWithToken(ctx context.Context, token, phone, pin, fullName, email string) error {
	inv, err := s.invRepo.GetByToken(ctx, token)
	if err != nil {
		return err
	}
	if inv == nil {
		return fmt.Errorf("invitation non trouvée")
	}
	if inv.UsedAt != nil {
		return fmt.Errorf("cette invitation a déjà été utilisée")
	}
	if time.Now().After(inv.ExpiresAt) {
		return fmt.Errorf("cette invitation a expiré")
	}

	// Check if user already exists
	existingUser, _ := s.userRepo.FindByPhone(ctx, phone)

	hash, err := bcrypt.GenerateFromPassword([]byte(pin), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if existingUser != nil {
		// Promote existing user to admin
		if err := s.userRepo.PromoteToAdmin(ctx, existingUser.ID, fullName, email, string(hash)); err != nil {
		    return err
		}
	} else {
		// Create new admin user
		user := &models.User{
			ID:           uuid.New(),
			Phone:        phone,
			PasswordHash: string(hash),
			FullName:     &fullName,
			Email:        &email,
			LanguagePref: "ar",
			Role:         "admin",
			IsVerified:   true,
		}
		if err := s.userRepo.Create(ctx, user); err != nil {
			return err
		}
	}

	// Mark invitation as used
	return s.invRepo.MarkAsUsed(ctx, inv.ID)
}

func (s *adminService) ListBlockedPhones(ctx context.Context) ([]map[string]interface{}, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT phone, reason, blocked_at, expires_at 
		FROM blocked_phones 
		ORDER BY blocked_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var phones []map[string]interface{}
	for rows.Next() {
		var phone string
		var reason *string
		var blockedAt time.Time
		var expiresAt *time.Time
		if err := rows.Scan(&phone, &reason, &blockedAt, &expiresAt); err != nil {
			continue
		}
		phones = append(phones, map[string]interface{}{
			"phone":      phone,
			"reason":     reason,
			"blocked_at": blockedAt,
			"expires_at": expiresAt,
		})
	}
	return phones, nil
}

func (s *adminService) BlockPhone(ctx context.Context, phone, reason string, blockedBy uuid.UUID) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO blocked_phones (phone, reason, blocked_by)
		VALUES ($1, $2, $3)
		ON CONFLICT (phone) DO UPDATE SET reason = $2
	`, phone, reason, blockedBy)
	return err
}

func (s *adminService) UnblockPhone(ctx context.Context, phone string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM blocked_phones WHERE phone = $1`, phone)
	return err
}

func (s *adminService) ListSettings(ctx context.Context) ([]models.SystemSettings, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, key, value, type FROM system_settings`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var settings []models.SystemSettings
	for rows.Next() {
		var s models.SystemSettings
		if err := rows.Scan(&s.ID, &s.Key, &s.Value, &s.Type); err != nil {
			continue
		}
		settings = append(settings, s)
	}
	return settings, nil
}

func (s *adminService) UpdateSetting(ctx context.Context, key, value, settingType string, userID uuid.UUID) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO system_settings (key, value, type, updated_by, updated_at)
		VALUES ($1, $2, $3, $4, now())
		ON CONFLICT (key) DO UPDATE SET value = $2, type = $3, updated_by = $4, updated_at = now()
	`, key, value, settingType, userID)
	return err
}
