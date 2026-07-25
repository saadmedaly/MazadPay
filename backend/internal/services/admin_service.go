package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	apperr "github.com/mazadpay/backend/internal/errors"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/repository"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type AdminService interface {
	GetDashboardStats(ctx context.Context) (map[string]interface{}, error)
	GetRevenueChartData(ctx context.Context) ([]map[string]interface{}, error)
	GetRecentActivity(ctx context.Context) ([]map[string]interface{}, error)
	ListUsers(ctx context.Context, page, perPage int, query string) ([]models.User, int, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	BlockUser(ctx context.Context, id uuid.UUID, block bool, adminID uuid.UUID) error
	DeleteUser(ctx context.Context, id uuid.UUID, adminID uuid.UUID) error
	ListAuctions(ctx context.Context, page, perPage int, status, query string, sellerID *uuid.UUID) ([]models.Auction, int, error)
	ValidateAuction(ctx context.Context, id uuid.UUID, approve bool, reason string, adminID uuid.UUID) error
	UpdateAuction(ctx context.Context, id uuid.UUID, input UpdateAuctionInput) error
	DeleteAuction(ctx context.Context, id uuid.UUID) error
	ListTransactions(ctx context.Context, page, perPage int, status string, userID *uuid.UUID) ([]models.Transaction, int, error)
	GetTransactionByID(ctx context.Context, id uuid.UUID) (*models.Transaction, error)
	ValidateTransaction(ctx context.Context, id uuid.UUID, approve bool, notes string, adminID uuid.UUID) error
	ListReports(ctx context.Context, page, perPage int, status string, reportType string) ([]models.Report, int, error)
	ReviewReport(ctx context.Context, id uuid.UUID, status, notes string, adminID uuid.UUID) error
	DeleteReport(ctx context.Context, id uuid.UUID, adminID uuid.UUID) error

	// KYC
	ListKYC(ctx context.Context, status string) ([]models.KYCVerification, error)
	ReviewKYC(ctx context.Context, userID uuid.UUID, status, notes string, adminID uuid.UUID) error

	// CMS (Banners are already handled separately but can be here too)
	// FAQ/Tutorials CRUD is better in ContentService but Admin can call it

	// Categories & Locations
	CreateCategory(ctx context.Context, c *models.Category, adminID uuid.UUID) error
	UpdateCategory(ctx context.Context, c *models.Category, adminID uuid.UUID) error
	DeleteCategory(ctx context.Context, id int, adminID uuid.UUID) error
	CreateLocation(ctx context.Context, l *models.Location) error
	UpdateLocation(ctx context.Context, l *models.Location) error
	DeleteLocation(ctx context.Context, id int) error

	// Admin Management
	CreateAdmin(ctx context.Context, phone, pin, fullName, email string) error
	GenerateAdminInvitation(ctx context.Context, createdBy uuid.UUID, targetPhone string) (string, error)
	RegisterAdminWithToken(ctx context.Context, token, phone, pin, fullName, email string) error

	// Blocked Phones
	ListBlockedPhones(ctx context.Context) ([]map[string]interface{}, error)
	BlockPhone(ctx context.Context, phone, reason string, blockedBy uuid.UUID) error
	UnblockPhone(ctx context.Context, phone string, unblockedBy uuid.UUID) error

	// Countries
	GetCountries(ctx context.Context) ([]models.Country, error)
	CreateCountry(ctx context.Context, code, countryCode, nameAr, nameFr, nameEn, flagEmoji string) error
	UpdateCountry(ctx context.Context, id int, code, countryCode, nameAr, nameFr, nameEn, flagEmoji string, isActive *bool) error
	DeleteCountry(ctx context.Context, id int) error

	// Settings
	ListSettings(ctx context.Context) ([]models.SystemSettings, error)
	UpdateSetting(ctx context.Context, key, value, settingType string, userID uuid.UUID) error

	// Payment Methods (from migration 000031)
	ListPaymentMethods(ctx context.Context) ([]models.PaymentMethod, error)
	CreatePaymentMethod(ctx context.Context, code, nameAr, nameFr string, nameEn, logoURL *string, isActive *bool, countryID *int, adminID uuid.UUID) error
	UpdatePaymentMethod(ctx context.Context, id int, code, nameAr, nameFr string, nameEn, logoURL *string, isActive *bool, countryID *int, adminID uuid.UUID) error
	DeletePaymentMethod(ctx context.Context, id int, adminID uuid.UUID) error

	// Auction Car Details (from migration 000031)
	GetAuctionCarDetails(ctx context.Context, auctionID uuid.UUID) (*models.AuctionCarDetails, error)
	UpdateAuctionCarDetails(ctx context.Context, auctionID uuid.UUID, manufacturer, model *string, year, mileage *int, fuelType, transmission, color, engineSize, VIN *string) error

	// Auction Boost (from migration 000031)
	ListAuctionBoosts(ctx context.Context) ([]models.AuctionBoost, error)
	CreateAuctionBoost(ctx context.Context, auctionID uuid.UUID, boostType string, startAt, endAt time.Time, amount *decimal.Decimal) error
	DeleteAuctionBoost(ctx context.Context, id uuid.UUID) error

	// Delivery Drivers (from migration 000031)
	ListDeliveryDrivers(ctx context.Context) ([]models.DeliveryDriver, error)
	CreateDeliveryDriver(ctx context.Context, userID *uuid.UUID, vehicleType, vehiclePlate, vehicleColor, licenseNumber *string, isAvailable *bool) error
	UpdateDeliveryDriver(ctx context.Context, id uuid.UUID, vehicleType, vehiclePlate, vehicleColor, licenseNumber *string, isAvailable *bool) error
	DeleteDeliveryDriver(ctx context.Context, id uuid.UUID) error

	// User Settings (from migration 000031)
	GetUserSettings(ctx context.Context, userID uuid.UUID) (*models.UserSettings, error)
	UpdateUserSettings(ctx context.Context, userID uuid.UUID, currency, theme, language *string, notificationsEmail, notificationsPush, notificationsSMS, twoFactorEnabled *bool) error

	// Bid Auto Bid (from migration 000031)
	ListAutoBids(ctx context.Context) ([]models.BidAutoBid, error)
	UpdateAutoBid(ctx context.Context, id uuid.UUID, isActive *bool) error

	// Admin image upload (bypasses seller ownership check)
	AdminAddAuctionImages(ctx context.Context, auctionID uuid.UUID, urls []string) error
}

type UpdateAuctionInput struct {
	CategoryID      int
	SubCategoryID   *int
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
	Images          []string
	ItemDetails     models.JSONB
	Condition       *string
	Brand           *string
	VideoURL        *string
	Quantity        int
}

type adminService struct {
	db           *sqlx.DB
	userRepo     repository.UserRepository
	auctionRepo  repository.AuctionRepository
	bidRepo      repository.BidRepository
	txRepo       repository.TransactionRepository
	reportRepo   repository.ReportRepository
	kycRepo      repository.KYCRepository
	contentRepo  repository.ContentRepository
	invRepo      repository.AdminInvitationRepository
	reqRepo      repository.RequestRepository
	settingsRepo repository.SettingsRepository
	mediaSvc     MediaService
	notifSvc     NotificationService
	auditSvc     AuditService
	rdb          *redis.Client
	logger       *zap.Logger
	jwtExpiry    int
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
	reqRepo repository.RequestRepository,
	settingsRepo repository.SettingsRepository,
	mediaSvc MediaService,
	notifSvc NotificationService,
	auditSvc AuditService,
	rdb *redis.Client,
	logger *zap.Logger,
	jwtExpiry int,
) AdminService {
	return &adminService{
		db:           db,
		userRepo:     userRepo,
		auctionRepo:  auctionRepo,
		bidRepo:      bidRepo,
		txRepo:       txRepo,
		reportRepo:   reportRepo,
		kycRepo:      kycRepo,
		contentRepo:  contentRepo,
		invRepo:      invRepo,
		reqRepo:      reqRepo,
		settingsRepo: settingsRepo,
		mediaSvc:     mediaSvc,
		notifSvc:     notifSvc,
		auditSvc:     auditSvc,
		rdb:          rdb,
		logger:       logger,
		jwtExpiry:    jwtExpiry,
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

	pendingAuctionRequests, _ := s.reqRepo.CountPendingAuctionRequests(ctx)
	pendingBannerRequests, _ := s.reqRepo.CountPendingBannerRequests(ctx)

	return map[string]interface{}{
		"total_users":              totalUsers,
		"total_auctions":           totalAuctions,
		"total_bids":               totalBids,
		"total_revenue":            totalRevenue,
		"today_revenue":            todayRevenue,
		"active_auctions":          activeAuctions,
		"pending_auctions":         pendingAuctions,
		"pending_reports":          pendingReports,
		"pending_kycs":             len(pendingKYCs),
		"pending_transactions":     pendingTransactions,
		"week_deposits":            weekDeposits,
		"pending_auction_requests": pendingAuctionRequests,
		"pending_banner_requests":  pendingBannerRequests,
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

func (s *adminService) BlockUser(ctx context.Context, id uuid.UUID, block bool, adminID uuid.UUID) error {
	// isActive est l'inverse logique de "bloqué" : is_active=false veut dire l'utilisateur
	// est bloqué. On récupère l'état avant modification pour old_status/new_status.
	userBefore, findErr := s.userRepo.FindByID(ctx, id)

	if err := s.userRepo.UpdateStatus(ctx, id, !block); err != nil {
		return err
	}

	if block {
		// Un compte bloqué ne doit pas rester utilisable via un JWT déjà émis —
		// invalider toutes ses sessions actives (Session Security Phase 1).
		RevokeUserSessions(ctx, s.rdb, s.logger, id, s.jwtExpiry)
	}

	if findErr == nil && s.auditSvc != nil {
		action := "user_unblocked"
		newStatus := "active"
		if block {
			action = "user_blocked"
			newStatus = "blocked"
		}
		oldStatus := "active"
		if !userBefore.IsActive {
			oldStatus = "blocked"
		}
		details := fmt.Sprintf("old_status=%s new_status=%s", oldStatus, newStatus)
		// details_json structuré (Audit Schema Phase B - User/Admin only) : jamais
		// phone/email/password_hash/tokens — uniquement target_user_id et statuts.
		detailsJSON := models.JSONB{
			"target_user_id": id.String(),
			"old_status":     oldStatus,
			"new_status":     newStatus,
		}
		// IP/User-Agent non disponibles ici : BlockUser est une méthode de service
		// sans *fiber.Ctx — non étendu dans cette phase (même limitation documentée
		// pour ValidateTransaction/ValidateAuction, Phases B précédentes).
		s.auditSvc.Log(ctx, adminID, action, "user", &id, details,
			WithActorType("admin"),
			WithDetailsJSON(detailsJSON),
		)
	}
	return nil
}

func (s *adminService) DeleteUser(ctx context.Context, id uuid.UUID, adminID uuid.UUID) error {
	// Récupérer le rôle avant suppression pour le journal d'audit (aucune donnée
	// sensible : uniquement le rôle, jamais phone/email/password_hash).
	userBefore, findErr := s.userRepo.FindByID(ctx, id)

	if err := s.userRepo.Delete(ctx, id); err != nil {
		return err
	}

	if s.auditSvc != nil {
		role := "unknown"
		if findErr == nil && userBefore != nil {
			role = userBefore.Role
		}
		// details_json structuré : uniquement target_user_id et role — jamais
		// phone/email/password_hash (Audit Schema Phase B - User/Admin only).
		detailsJSON := models.JSONB{
			"target_user_id": id.String(),
			"role":           role,
		}
		s.auditSvc.Log(ctx, adminID, "user_deleted", "user", &id, fmt.Sprintf("role=%s", role),
			WithActorType("admin"),
			WithDetailsJSON(detailsJSON),
		)
	}
	return nil
}

func (s *adminService) ListAuctions(ctx context.Context, page, perPage int, status, query string, sellerID *uuid.UUID) ([]models.Auction, int, error) {
	filters := repository.AuctionFilters{
		Status:   status,
		Query:    query,
		SellerID: sellerID,
	}
	return s.auctionRepo.ListPaginated(ctx, page, perPage, filters)
}

func (s *adminService) ValidateAuction(ctx context.Context, id uuid.UUID, approve bool, reason string, adminID uuid.UUID) error {
	// Récupérer l'auction avant modification pour connaître old_status et seller_id
	// (Auction audit logs, ne bloque jamais l'opération si l'audit échoue).
	auctionBefore, findErr := s.auctionRepo.FindByID(ctx, id)

	status := "rejected"
	if approve {
		status = "active"
		// Un auction ne peut jamais devenir "active" sans caution définie (audit V03) :
		// c'est cette valeur qui protège les enchérisseurs contre les mises sans fonds.
		existing, err := s.auctionRepo.FindByID(ctx, id)
		if err != nil {
			return apperr.ErrNotFound
		}
		if !existing.InsuranceAmount.GreaterThan(decimal.Zero) {
			return apperr.ErrInsuranceNotSet
		}
	}
	if err := s.auctionRepo.UpdateStatus(ctx, id, status); err != nil {
		return err
	}

	if findErr == nil && s.auditSvc != nil {
		action := "auction_rejected"
		if approve {
			action = "auction_approved"
		}
		details := fmt.Sprintf("seller_id=%s old_status=%s new_status=%s reason=%s",
			auctionBefore.SellerID, auctionBefore.Status, status, reason)
		// details_json structuré (Audit Schema Phase B - Auction only) : uniquement
		// des champs non sensibles, jamais de détails sur les enchérisseurs.
		detailsJSON := models.JSONB{
			"auction_id": id.String(),
			"seller_id":  auctionBefore.SellerID.String(),
			"old_status": auctionBefore.Status,
			"new_status": status,
		}
		if reason != "" {
			detailsJSON["reason"] = reason
		}
		// IP/User-Agent non disponibles ici : ValidateAuction est une méthode de
		// service sans accès à *fiber.Ctx (voir même limitation documentée pour
		// ValidateTransaction, Audit Schema Phase B - Financial only).
		s.auditSvc.Log(ctx, adminID, action, "auction", &id, details,
			WithActorType("admin"),
			WithDetailsJSON(detailsJSON),
		)
	}

	// Send notification to seller
	auction, err := s.auctionRepo.FindByID(ctx, id)
	if err != nil {
		return nil
	}
	seller, err := s.userRepo.FindByID(ctx, auction.SellerID)
	if err != nil {
		return nil
	}
	language := "ar"
	if seller.LanguagePref != "" {
		language = seller.LanguagePref
	}
	notifType := "auction_approved"
	if !approve {
		notifType = "auction_rejected"
	}
	title := auction.TitleAr
	if language == "fr" && auction.TitleFr != nil && *auction.TitleFr != "" {
		title = *auction.TitleFr
	} else if language == "en" && auction.TitleEn != nil && *auction.TitleEn != "" {
		title = *auction.TitleEn
	}
	params := map[string]string{
		"auctionTitle": title,
	}
	if !approve {
		params["reason"] = reason
	}
	data := map[string]string{
		"type":      notifType,
		"auctionId": id.String(),
	}
	_ = s.notifSvc.SendLocalizedPush(ctx, auction.SellerID, notifType, language, params, data)
	return nil
}

func (s *adminService) UpdateAuction(ctx context.Context, id uuid.UUID, input UpdateAuctionInput) error {
	auction, err := s.auctionRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	var tFr, tEn *string
	var dAr, dFr, dEn *string
	if input.TitleFr != "" {
		tFr = &input.TitleFr
	}
	if input.TitleEn != "" {
		tEn = &input.TitleEn
	}
	if input.DescriptionAr != "" {
		dAr = &input.DescriptionAr
	}
	if input.DescriptionFr != "" {
		dFr = &input.DescriptionFr
	}
	if input.DescriptionEn != "" {
		dEn = &input.DescriptionEn
	}

	var phone *string
	if input.PhoneContact != "" {
		phone = &input.PhoneContact
	}

	if input.CategoryID != 0 {
		auction.CategoryID = input.CategoryID
	}
	if input.LocationID != nil {
		auction.LocationID = input.LocationID
	}
	if input.TitleAr != "" {
		auction.TitleAr = input.TitleAr
	}
	auction.TitleFr = tFr
	auction.TitleEn = tEn
	auction.DescriptionAr = dAr
	auction.DescriptionFr = dFr
	auction.DescriptionEn = dEn
	if !input.StartPrice.IsZero() {
		auction.StartPrice = input.StartPrice
	}
	if !input.MinIncrement.IsZero() {
		auction.MinIncrement = input.MinIncrement
	}
	if !input.InsuranceAmount.IsZero() {
		auction.InsuranceAmount = input.InsuranceAmount
	}
	// Only update end_time if explicitly provided (non-zero)
	if !input.EndTime.IsZero() {
		auction.EndTime = input.EndTime
	}
	auction.PhoneContact = phone
	auction.BuyNowPrice = input.BuyNowPrice
	if input.ItemDetails != nil {
		auction.ItemDetails = input.ItemDetails
	}
	if input.StartTime != nil {
		auction.StartTime = *input.StartTime
	}

	// Start transaction for atomic update
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	if err := s.auctionRepo.Update(ctx, auction); err != nil {
		return err
	}

	// Only sync images if explicitly provided in the request
	var existingImages []models.AuctionImage
	hasNewImages := false
	for _, url := range input.Images {
		if url != "" {
			hasNewImages = true
			break
		}
	}

	if hasNewImages {
		// Get existing images before deleting (for R2 cleanup)
		var imgErr error
		existingImages, imgErr = s.auctionRepo.GetImages(ctx, id)
		if imgErr != nil {
			return fmt.Errorf("failed to get existing images: %w", imgErr)
		}

		// Replace images
		if err := s.auctionRepo.DeleteImagesTx(ctx, tx, id); err != nil {
			return err
		}
		for i, url := range input.Images {
			if url == "" {
				continue
			}
			if addErr := s.auctionRepo.AddImageTx(ctx, tx, &models.AuctionImage{
				AuctionID:    id,
				URL:          url,
				MediaType:    "image",
				DisplayOrder: i,
			}); addErr != nil {
				return addErr
			}
		}
	}

	// Commit transaction first
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// After successful commit, delete old images from R2 (best effort, don't fail if error)
	if s.mediaSvc != nil {
		for _, img := range existingImages {
			if err := s.mediaSvc.DeleteFile(ctx, img.URL); err != nil {
				// Log but don't fail - the DB update succeeded
				fmt.Printf("[Admin UpdateAuction] Warning: failed to delete old image from R2: %s, error: %v\n", img.URL, err)
			}
		}
	}

	return nil
}

func (s *adminService) DeleteAuction(ctx context.Context, id uuid.UUID) error {
	// Get existing images before deleting (for R2 cleanup)
	existingImages, err := s.auctionRepo.GetImages(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get existing images: %w", err)
	}

	// Delete auction from DB
	if err := s.auctionRepo.Delete(ctx, id); err != nil {
		return err
	}

	// After successful DB deletion, delete images from R2 (best effort)
	if s.mediaSvc != nil {
		for _, img := range existingImages {
			if err := s.mediaSvc.DeleteFile(ctx, img.URL); err != nil {
				// Log but don't fail - the DB deletion succeeded
				fmt.Printf("[Admin DeleteAuction] Warning: failed to delete image from R2: %s, error: %v\n", img.URL, err)
			}
		}
	}

	return nil
}

func (s *adminService) ListTransactions(ctx context.Context, page, perPage int, status string, userID *uuid.UUID) ([]models.Transaction, int, error) {
	return s.txRepo.ListPaginated(ctx, page, perPage, status, userID)
}

func (s *adminService) GetTransactionByID(ctx context.Context, id uuid.UUID) (*models.Transaction, error) {
	return s.txRepo.FindByID(ctx, id, nil)
}

func (s *adminService) ValidateTransaction(ctx context.Context, id uuid.UUID, approve bool, notes string, adminID uuid.UUID) error {
	status := "rejected"
	if approve {
		status = "completed"
	}

	// Récupérer la transaction avant modification pour connaître old_status (audit
	// financier — Financial audit logs, ne bloque jamais l'opération si l'audit échoue).
	txBefore, findErr := s.txRepo.FindByID(ctx, id, nil)

	if err := s.txRepo.UpdateStatus(ctx, id, status, notes, adminID); err != nil {
		return err
	}

	if findErr == nil && s.auditSvc != nil {
		action := "transaction_rejected"
		if approve {
			action = "transaction_approved"
		}
		details := fmt.Sprintf(
			"user_id=%s type=%s amount=%s old_status=%s new_status=%s notes=%s",
			txBefore.UserID, txBefore.Type, txBefore.Amount.String(), txBefore.Status, status, notes,
		)
		// details_json structuré (Audit Schema Phase B - Financial only) : uniquement
		// des champs non sensibles, jamais de receipt_url/presigned_url/token.
		detailsJSON := models.JSONB{
			"transaction_id": id.String(),
			"user_id":        txBefore.UserID.String(),
			"type":           txBefore.Type,
			"amount":         txBefore.Amount.String(),
			"old_status":     txBefore.Status,
			"new_status":     status,
		}
		if notes != "" {
			detailsJSON["notes"] = notes
		}
		// IP/User-Agent non disponibles ici : ValidateTransaction est une méthode de
		// service sans accès à *fiber.Ctx, et le handler appelant (admin_handler.go)
		// ne les transmet pas actuellement. Étendre cette signature pour les faire
		// transiter serait un changement plus large que ce que demande cette phase
		// (voir rapport Phase B) — laissé pour une phase ultérieure si nécessaire.
		s.auditSvc.Log(ctx, adminID, action, "transaction", &id, details,
			WithActorType("admin"),
			WithDetailsJSON(detailsJSON),
		)
	}

	// Send notification to user
	tx, err := s.txRepo.FindByID(ctx, id, nil)
	if err != nil {
		return nil
	}
	user, err := s.userRepo.FindByID(ctx, tx.UserID)
	if err != nil {
		return nil
	}
	language := "ar"
	if user.LanguagePref != "" {
		language = user.LanguagePref
	}
	notifType := "deposit_confirmed"
	if !approve {
		notifType = "deposit_rejected"
	}
	params := map[string]string{
		"amount": tx.Amount.String(),
	}
	if !approve {
		params["reason"] = notes
	}
	data := map[string]string{
		"type":           notifType,
		"transaction_id": id.String(),
	}
	_ = s.notifSvc.SendLocalizedPush(ctx, tx.UserID, notifType, language, params, data)
	return nil
}

func (s *adminService) ListReports(ctx context.Context, page, perPage int, status string, reportType string) ([]models.Report, int, error) {
	return s.reportRepo.ListPaginated(ctx, page, perPage, status, reportType)
}

func (s *adminService) ReviewReport(ctx context.Context, id uuid.UUID, status, notes string, adminID uuid.UUID) error {
	if err := s.reportRepo.UpdateStatus(ctx, id, status, notes, adminID); err != nil {
		return err
	}
	if s.auditSvc != nil {
		detailsJSON := models.JSONB{"report_id": id.String(), "status": status}
		if notes != "" {
			detailsJSON["notes"] = notes
		}
		s.auditSvc.Log(ctx, adminID, "report_reviewed", "report", &id, fmt.Sprintf("status=%s notes=%s", status, notes),
			WithActorType("admin"),
			WithDetailsJSON(detailsJSON),
		)
	}
	return nil
}

func (s *adminService) DeleteReport(ctx context.Context, id uuid.UUID, adminID uuid.UUID) error {
	if err := s.reportRepo.Delete(ctx, id); err != nil {
		return err
	}
	if s.auditSvc != nil {
		s.auditSvc.Log(ctx, adminID, "report_deleted", "report", &id, "",
			WithActorType("admin"),
			WithDetailsJSON(models.JSONB{"report_id": id.String()}),
		)
	}
	return nil
}

func (s *adminService) ListKYC(ctx context.Context, status string) ([]models.KYCVerification, error) {
	return s.kycRepo.List(ctx, status)
}

func (s *adminService) ReviewKYC(ctx context.Context, userID uuid.UUID, status, notes string, adminID uuid.UUID) error {
	if err := s.kycRepo.UpdateStatus(ctx, userID, status, notes, adminID); err != nil {
		return err
	}

	if s.auditSvc != nil {
		action := "kyc_rejected"
		if status == "approved" {
			action = "kyc_approved"
		}
		detailsJSON := models.JSONB{"target_user_id": userID.String(), "status": status}
		if notes != "" {
			detailsJSON["notes"] = notes
		}
		s.auditSvc.Log(ctx, adminID, action, "kyc", &userID, fmt.Sprintf("notes=%s", notes),
			WithActorType("admin"),
			WithDetailsJSON(detailsJSON),
		)
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

func (s *adminService) CreateCategory(ctx context.Context, c *models.Category, adminID uuid.UUID) error {
	if err := s.auctionRepo.CreateCategory(ctx, c); err != nil {
		return err
	}
	s.logCategoryAudit(ctx, adminID, "category_created", c.ID, c.NameAr)
	return nil
}

func (s *adminService) UpdateCategory(ctx context.Context, c *models.Category, adminID uuid.UUID) error {
	if err := s.auctionRepo.UpdateCategory(ctx, c); err != nil {
		return err
	}
	s.logCategoryAudit(ctx, adminID, "category_updated", c.ID, c.NameAr)
	return nil
}

func (s *adminService) DeleteCategory(ctx context.Context, id int, adminID uuid.UUID) error {
	if err := s.auctionRepo.DeleteCategory(ctx, id); err != nil {
		return err
	}
	s.logCategoryAudit(ctx, adminID, "category_deleted", id, "")
	return nil
}

// logCategoryAudit journalise une action admin sur une catégorie (Content/Settings
// audit logs — Phase B). IP/User-Agent non disponibles ici : ces méthodes de service
// n'ont pas accès à *fiber.Ctx (même limitation documentée pour ValidateAuction).
func (s *adminService) logCategoryAudit(ctx context.Context, adminID uuid.UUID, action string, id int, nameAr string) {
	if s.auditSvc == nil {
		return
	}
	details := fmt.Sprintf("category_id=%d", id)
	if nameAr != "" {
		details += " name_ar=" + nameAr
	}
	detailsJSON := models.JSONB{"category_id": id}
	if nameAr != "" {
		detailsJSON["name_ar"] = nameAr
	}
	s.auditSvc.Log(ctx, adminID, action, "category", nil, details,
		WithActorType("admin"),
		WithDetailsJSON(detailsJSON),
		WithEntityKey(strconv.Itoa(id)),
	)
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

// normalizePhoneForInvitation applique une normalisation simple et sûre, dédiée aux
// invitations admin (Admin Authorization Phase 1C-B) — ne touche jamais à la logique
// de login/register générale. Doit être appelée à l'identique lors de la création de
// l'invitation (hash stocké) et lors de son utilisation (hash comparé), sinon la
// comparaison échouerait silencieusement pour un numéro pourtant identique.
func normalizePhoneForInvitation(phone string) string {
	p := strings.TrimSpace(phone)
	p = strings.NewReplacer(" ", "", "-", "", "(", "", ")", "").Replace(p)
	if strings.HasPrefix(p, "00") {
		p = "+" + strings.TrimPrefix(p, "00")
	}
	return p
}

// hashPhoneForInvitation calcule le SHA-256 hex du numéro normalisé — jamais le
// numéro complet n'est stocké en base pour une invitation (Admin Authorization Phase
// 1C-B) ; seul ce hash sert à la comparaison lors de RegisterAdminWithToken.
func hashPhoneForInvitation(normalizedPhone string) string {
	sum := sha256.Sum256([]byte(normalizedPhone))
	return hex.EncodeToString(sum[:])
}

// maskPhoneForInvitation retourne une version affichable/auditable du numéro (les 4
// derniers chiffres seulement), jamais le numéro complet.
func maskPhoneForInvitation(phone string) string {
	if len(phone) < 4 {
		return "####"
	}
	return "####" + phone[len(phone)-4:]
}

// CreateAdmin n'est actuellement rattachée à aucune route (code non exposé côté API) —
// l'audit log est ajouté ici par cohérence avec RegisterAdminWithToken/CreateAdmin, sans
// modifier le comportement ni relier la fonction à un endpoint (Admin Authorization
// Phase 1A). Jamais de PIN/password_hash/token journalisé.
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

	if err := s.userRepo.Create(ctx, user); err != nil {
		return err
	}

	if s.auditSvc != nil {
		s.auditSvc.Log(ctx, user.ID, "admin_created", "user", &user.ID,
			fmt.Sprintf("user_id=%s mode=created", user.ID),
			WithActorType("admin"),
			WithDetailsJSON(models.JSONB{
				"user_id": user.ID.String(),
				"mode":    "created",
			}),
		)
	}

	return nil
}

// GenerateAdminInvitation exige désormais un targetPhone (Admin Authorization Phase
// 1C-B) : le numéro complet n'est jamais stocké, seuls son hash SHA-256 (comparaison)
// et sa version masquée (affichage/audit) le sont.
func (s *adminService) GenerateAdminInvitation(ctx context.Context, createdBy uuid.UUID, targetPhone string) (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)

	normalizedPhone := normalizePhoneForInvitation(targetPhone)
	targetPhoneHash := hashPhoneForInvitation(normalizedPhone)
	targetPhoneMasked := maskPhoneForInvitation(normalizedPhone)

	inv := &models.AdminInvitation{
		ID:                uuid.New(),
		Token:             token,
		CreatedBy:         createdBy,
		ExpiresAt:         time.Now().Add(24 * time.Hour), // 24h validity
		TargetPhoneHash:   &targetPhoneHash,
		TargetPhoneMasked: &targetPhoneMasked,
	}

	if err := s.invRepo.Create(ctx, inv); err != nil {
		return "", err
	}

	if s.auditSvc != nil {
		// Ne jamais journaliser le token ni le numéro complet ni son hash (audit de
		// sécurité — Admin Authorization Phase 1A/1C-B), uniquement l'identifiant de
		// l'invitation et le numéro masqué.
		s.auditSvc.Log(ctx, createdBy, "admin_invitation_created", "admin_invitation", &inv.ID,
			fmt.Sprintf("invitation_id=%s expires_at=%s target_phone_masked=%s", inv.ID, inv.ExpiresAt.Format(time.RFC3339), targetPhoneMasked),
			WithActorType("admin"),
			WithDetailsJSON(models.JSONB{
				"invitation_id":       inv.ID.String(),
				"created_by":          createdBy.String(),
				"expires_at":          inv.ExpiresAt.Format(time.RFC3339),
				"target_phone_masked": targetPhoneMasked,
			}),
		)
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

	// Invitation liée à un numéro cible (Admin Authorization Phase 1C-B) : une
	// invitation sans target_phone_hash (créée avant ce durcissement) est refusée —
	// comportement intentionnel, pas un bug (voir plan Phase 1C). Le message reste
	// générique, identique à celui utilisé plus bas en cas de non-correspondance,
	// pour ne jamais révéler si l'invitation avait ou non un numéro cible.
	if inv.TargetPhoneHash == nil || *inv.TargetPhoneHash == "" {
		return fmt.Errorf("invitation invalide")
	}
	normalizedInputPhone := normalizePhoneForInvitation(phone)
	computedHash := hashPhoneForInvitation(normalizedInputPhone)
	if subtle.ConstantTimeCompare([]byte(computedHash), []byte(*inv.TargetPhoneHash)) != 1 {
		// Message générique : ne jamais révéler le numéro cible attendu (empêche
		// l'énumération, cohérent avec le durcissement OTP Security Phase 1B).
		return fmt.Errorf("invitation invalide")
	}

	// Check if user already exists
	existingUser, _ := s.userRepo.FindByPhone(ctx, phone)

	hash, err := bcrypt.GenerateFromPassword([]byte(pin), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	var userID uuid.UUID
	mode := "created"

	if existingUser != nil {
		// Promote existing user to admin
		if err := s.userRepo.PromoteToAdmin(ctx, existingUser.ID, fullName, email, string(hash)); err != nil {
			return err
		}
		userID = existingUser.ID
		mode = "promoted"
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
		userID = user.ID
	}

	if s.auditSvc != nil {
		// Endpoint public (register-admin) sans JWT admin — l'acteur n'est ni un admin
		// authentifié ni un processus système au sens strict, mais l'utilisateur/
		// invité qui a consommé le token. actor_type="user" reflète cela ; actor_id
		// pointe vers le compte concerné lui-même (aucune autre identité disponible
		// côté serveur à ce stade). Jamais de token/password_hash/PIN journalisé —
		// uniquement le téléphone masqué (Admin Authorization Phase 1A).
		action := "admin_registered_with_invitation"
		if mode == "promoted" {
			action = "user_promoted_to_admin_with_invitation"
		}
		maskedPhone := "####"
		if len(phone) >= 4 {
			maskedPhone = "####" + phone[len(phone)-4:]
		}
		s.auditSvc.Log(ctx, userID, action, "user", &userID,
			fmt.Sprintf("user_id=%s invitation_id=%s mode=%s", userID, inv.ID, mode),
			WithActorType("user"),
			WithDetailsJSON(models.JSONB{
				"user_id":       userID.String(),
				"invitation_id": inv.ID.String(),
				"invited_by":    inv.CreatedBy.String(),
				"phone_masked":  maskedPhone,
				"mode":          mode,
			}),
		)
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
	phoneMasked := maskPhoneForInvitation(phone)

	// Vérifie d'abord si un compte actif existe pour ce numéro, avec une variante
	// stricte qui distingue "non enregistré" d'une vraie panne DB (Blocked Phone
	// Phase 1C) : FindByPhone (générique) fusionnerait les deux en ErrNotFound, ce
	// qui masquerait une panne DB réelle et donnerait un faux sentiment de succès.
	existingUser, findErr := s.userRepo.FindByPhoneForRevocation(ctx, phone)
	if findErr != nil && !errors.Is(findErr, apperr.ErrNotFound) {
		return findErr
	}

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO blocked_phones (phone, reason, blocked_by)
		VALUES ($1, $2, $3)
		ON CONFLICT (phone) DO UPDATE SET reason = $2
	`, phone, reason, blockedBy)
	if err != nil {
		return err
	}

	// Un numéro bloqué correspondant à un compte existant ne doit pas rester
	// utilisable via un JWT déjà émis avant le blocage — invalider ses sessions
	// actives (Blocked Phone Phase 1C), comme pour BlockUser (Session Security
	// Phase 1). Si aucun compte actif n'a ce numéro (ErrNotFound ci-dessus), ce
	// n'est pas une erreur : le blocage reste enregistré pour empêcher une
	// inscription/utilisation future.
	revokedSessions := false
	if existingUser != nil {
		if revokeErr := RevokeUserSessionsErr(ctx, s.rdb, existingUser.ID, s.jwtExpiry); revokeErr != nil {
			if s.logger != nil {
				s.logger.Error("BlockPhone: failed to revoke existing sessions",
					zap.String("phone_masked", phoneMasked),
					zap.Error(revokeErr),
				)
			}
			return fmt.Errorf("phone blocked but failed to revoke active sessions: %w", revokeErr)
		}
		revokedSessions = true
	}

	if s.auditSvc != nil {
		// Ne jamais journaliser le numéro complet (audit de sécurité — Blocked Phone
		// Phase 1B), uniquement sa version masquée (4 derniers chiffres). Cette
		// insertion ne fixe jamais expires_at (colonne existante mais non utilisée
		// par ce chemin), donc le blocage est toujours permanent ici.
		details := map[string]interface{}{
			"phone_masked":     phoneMasked,
			"permanent":        true,
			"revoked_sessions": revokedSessions,
		}
		if reason != "" {
			details["reason"] = reason
		}
		s.auditSvc.Log(ctx, blockedBy, "phone_blocked", "blocked_phone", nil,
			fmt.Sprintf("phone_masked=%s permanent=true revoked_sessions=%t", phoneMasked, revokedSessions),
			WithActorType("admin"),
			WithDetailsJSON(details),
		)
	}

	return nil
}

func (s *adminService) UnblockPhone(ctx context.Context, phone string, unblockedBy uuid.UUID) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM blocked_phones WHERE phone = $1`, phone)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no blocked phone found")
	}

	if s.auditSvc != nil {
		// Ne jamais journaliser le numéro complet (audit de sécurité — Blocked Phone
		// Phase 1B), uniquement sa version masquée (4 derniers chiffres).
		phoneMasked := maskPhoneForInvitation(phone)
		s.auditSvc.Log(ctx, unblockedBy, "phone_unblocked", "blocked_phone", nil,
			fmt.Sprintf("phone_masked=%s", phoneMasked),
			WithActorType("admin"),
			WithDetailsJSON(map[string]interface{}{
				"phone_masked": phoneMasked,
			}),
		)
	}

	return nil
}

func (s *adminService) ListSettings(ctx context.Context) ([]models.SystemSettings, error) {
	return s.settingsRepo.List(ctx)
}

func (s *adminService) UpdateSetting(ctx context.Context, key, value, settingType string, userID uuid.UUID) error {
	return s.settingsRepo.Set(ctx, key, value, settingType)
}

// Payment Methods implementations
func (s *adminService) ListPaymentMethods(ctx context.Context) ([]models.PaymentMethod, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, code, name_ar, name_fr, name_en, logo_url, is_active, country_id, created_at FROM payment_methods`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var methods []models.PaymentMethod
	for rows.Next() {
		var m models.PaymentMethod
		if err := rows.Scan(&m.ID, &m.Code, &m.NameAr, &m.NameFr, &m.NameEn, &m.LogoURL, &m.IsActive, &m.CountryID, &m.CreatedAt); err != nil {
			continue
		}
		methods = append(methods, m)
	}
	return methods, nil
}

func (s *adminService) CreatePaymentMethod(ctx context.Context, code, nameAr, nameFr string, nameEn, logoURL *string, isActive *bool, countryID *int, adminID uuid.UUID) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO payment_methods (code, name_ar, name_fr, name_en, logo_url, is_active, country_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, code, nameAr, nameFr, nameEn, logoURL, isActive, countryID)
	if err != nil {
		return err
	}
	if s.auditSvc != nil {
		// IP/User-Agent non disponibles ici : méthode de service sans *fiber.Ctx.
		s.auditSvc.Log(ctx, adminID, "payment_method_created", "payment_method", nil, fmt.Sprintf("code=%s name_ar=%s", code, nameAr),
			WithActorType("admin"),
			WithDetailsJSON(models.JSONB{"code": code, "name_ar": nameAr}),
		)
	}
	return nil
}

func (s *adminService) UpdatePaymentMethod(ctx context.Context, id int, code, nameAr, nameFr string, nameEn, logoURL *string, isActive *bool, countryID *int, adminID uuid.UUID) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE payment_methods SET code = $1, name_ar = $2, name_fr = $3, name_en = $4, logo_url = $5, is_active = $6, country_id = $7
		WHERE id = $8
	`, code, nameAr, nameFr, nameEn, logoURL, isActive, countryID, id)
	if err != nil {
		return err
	}
	if s.auditSvc != nil {
		s.auditSvc.Log(ctx, adminID, "payment_method_updated", "payment_method", nil, fmt.Sprintf("payment_method_id=%d code=%s name_ar=%s", id, code, nameAr),
			WithActorType("admin"),
			WithDetailsJSON(models.JSONB{"payment_method_id": id, "code": code, "name_ar": nameAr}),
			WithEntityKey(strconv.Itoa(id)),
		)
	}
	return nil
}

func (s *adminService) DeletePaymentMethod(ctx context.Context, id int, adminID uuid.UUID) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM payment_methods WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if s.auditSvc != nil {
		s.auditSvc.Log(ctx, adminID, "payment_method_deleted", "payment_method", nil, fmt.Sprintf("payment_method_id=%d", id),
			WithActorType("admin"),
			WithDetailsJSON(models.JSONB{"payment_method_id": id}),
			WithEntityKey(strconv.Itoa(id)),
		)
	}
	return nil
}

// Auction Car Details implementations
func (s *adminService) GetAuctionCarDetails(ctx context.Context, auctionID uuid.UUID) (*models.AuctionCarDetails, error) {
	var details models.AuctionCarDetails
	err := s.db.QueryRowContext(ctx, `
		SELECT id, auction_id, manufacturer, model, year, mileage, fuel_type, transmission, color, engine_size, vin, created_at
		FROM auction_car_details WHERE auction_id = $1
	`, auctionID).Scan(&details.ID, &details.AuctionID, &details.Manufacturer, &details.Model, &details.Year, &details.Mileage,
		&details.FuelType, &details.Transmission, &details.Color, &details.EngineSize, &details.VIN, &details.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &details, nil
}

func (s *adminService) UpdateAuctionCarDetails(ctx context.Context, auctionID uuid.UUID, manufacturer, model *string, year, mileage *int, fuelType, transmission, color, engineSize, VIN *string) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO auction_car_details (auction_id, manufacturer, model, year, mileage, fuel_type, transmission, color, engine_size, vin)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (auction_id) DO UPDATE SET manufacturer = $2, model = $3, year = $4, mileage = $5, fuel_type = $6, transmission = $7, color = $8, engine_size = $9, vin = $10
	`, auctionID, manufacturer, model, year, mileage, fuelType, transmission, color, engineSize, VIN)
	return err
}

// Auction Boost implementations
func (s *adminService) ListAuctionBoosts(ctx context.Context) ([]models.AuctionBoost, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, auction_id, boost_type, start_at, end_at, amount, status, created_at FROM auction_boosts`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var boosts []models.AuctionBoost
	for rows.Next() {
		var b models.AuctionBoost
		if err := rows.Scan(&b.ID, &b.AuctionID, &b.BoostType, &b.StartAt, &b.EndAt, &b.Amount, &b.Status, &b.CreatedAt); err != nil {
			continue
		}
		boosts = append(boosts, b)
	}
	return boosts, nil
}

func (s *adminService) CreateAuctionBoost(ctx context.Context, auctionID uuid.UUID, boostType string, startAt, endAt time.Time, amount *decimal.Decimal) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO auction_boosts (auction_id, boost_type, start_at, end_at, amount, status)
		VALUES ($1, $2, $3, $4, $5, 'active')
	`, auctionID, boostType, startAt, endAt, amount)
	return err
}

func (s *adminService) DeleteAuctionBoost(ctx context.Context, id uuid.UUID) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM auction_boosts WHERE id = $1`, id)
	return err
}

// Delivery Drivers implementations
func (s *adminService) ListDeliveryDrivers(ctx context.Context) ([]models.DeliveryDriver, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, user_id, vehicle_type, vehicle_plate, vehicle_color, license_number, rating, total_deliveries, is_available, current_location_lat, current_location_lng, created_at FROM delivery_drivers`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var drivers []models.DeliveryDriver
	for rows.Next() {
		var d models.DeliveryDriver
		if err := rows.Scan(&d.ID, &d.UserID, &d.VehicleType, &d.VehiclePlate, &d.VehicleColor, &d.LicenseNumber, &d.Rating, &d.TotalDeliveries, &d.IsAvailable, &d.CurrentLocationLat, &d.CurrentLocationLng, &d.CreatedAt); err != nil {
			continue
		}
		drivers = append(drivers, d)
	}
	return drivers, nil
}

func (s *adminService) CreateDeliveryDriver(ctx context.Context, userID *uuid.UUID, vehicleType, vehiclePlate, vehicleColor, licenseNumber *string, isAvailable *bool) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO delivery_drivers (user_id, vehicle_type, vehicle_plate, vehicle_color, license_number, is_available)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, userID, vehicleType, vehiclePlate, vehicleColor, licenseNumber, isAvailable)
	return err
}

func (s *adminService) UpdateDeliveryDriver(ctx context.Context, id uuid.UUID, vehicleType, vehiclePlate, vehicleColor, licenseNumber *string, isAvailable *bool) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE delivery_drivers SET vehicle_type = $1, vehicle_plate = $2, vehicle_color = $3, license_number = $4, is_available = $5
		WHERE id = $6
	`, vehicleType, vehiclePlate, vehicleColor, licenseNumber, isAvailable, id)
	return err
}

func (s *adminService) DeleteDeliveryDriver(ctx context.Context, id uuid.UUID) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM delivery_drivers WHERE id = $1`, id)
	return err
}

// User Settings implementations
func (s *adminService) GetUserSettings(ctx context.Context, userID uuid.UUID) (*models.UserSettings, error) {
	var settings models.UserSettings
	err := s.db.QueryRowContext(ctx, `
		SELECT user_id, currency, theme, language, notifications_email, notifications_push, notifications_sms, two_factor_enabled, created_at, updated_at
		FROM user_settings WHERE user_id = $1
	`, userID).Scan(&settings.UserID, &settings.Currency, &settings.Theme, &settings.Language, &settings.NotificationsEmail,
		&settings.NotificationsPush, &settings.NotificationsSMS, &settings.TwoFactorEnabled, &settings.CreatedAt, &settings.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &settings, nil
}

func (s *adminService) UpdateUserSettings(ctx context.Context, userID uuid.UUID, currency, theme, language *string, notificationsEmail, notificationsPush, notificationsSMS, twoFactorEnabled *bool) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO user_settings (user_id, currency, theme, language, notifications_email, notifications_push, notifications_sms, two_factor_enabled)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (user_id) DO UPDATE SET currency = $2, theme = $3, language = $4, notifications_email = $5, notifications_push = $6, notifications_sms = $7, two_factor_enabled = $8, updated_at = now()
	`, userID, currency, theme, language, notificationsEmail, notificationsPush, notificationsSMS, twoFactorEnabled)
	return err
}

// Bid Auto Bid implementations
func (s *adminService) ListAutoBids(ctx context.Context) ([]models.BidAutoBid, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, user_id, auction_id, max_amount, current_bid_amount, is_active, created_at FROM bid_auto_bids`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bids []models.BidAutoBid
	for rows.Next() {
		var b models.BidAutoBid
		if err := rows.Scan(&b.ID, &b.UserID, &b.AuctionID, &b.MaxAmount, &b.CurrentBidAmount, &b.IsActive, &b.CreatedAt); err != nil {
			continue
		}
		bids = append(bids, b)
	}
	return bids, nil
}

func (s *adminService) UpdateAutoBid(ctx context.Context, id uuid.UUID, isActive *bool) error {
	_, err := s.db.ExecContext(ctx, `UPDATE bid_auto_bids SET is_active = $1 WHERE id = $2`, isActive, id)
	return err
}

func (s *adminService) GetCountries(ctx context.Context) ([]models.Country, error) {
	return s.auctionRepo.GetCountries(ctx)
}

func (s *adminService) CreateCountry(ctx context.Context, code, countryCode, nameAr, nameFr, nameEn, flagEmoji string) error {
	country := &models.Country{
		Code:        code,
		CountryCode: nil, // NULL by default
		NameAr:      nameAr,
		NameFr:      nameFr,
		NameEn:      nameEn,
		FlagEmoji:   flagEmoji,
		IsActive:    true,
	}
	// Set CountryCode only if not empty
	if countryCode != "" {
		country.CountryCode = &countryCode
	}
	return s.auctionRepo.CreateCountry(ctx, country)
}

func (s *adminService) UpdateCountry(ctx context.Context, id int, code, countryCode, nameAr, nameFr, nameEn, flagEmoji string, isActive *bool) error {
	country := &models.Country{
		ID:          id,
		Code:        code,
		CountryCode: nil, // NULL by default
		NameAr:      nameAr,
		NameFr:      nameFr,
		NameEn:      nameEn,
		FlagEmoji:   flagEmoji,
		IsActive:    true,
	}
	// Set CountryCode only if not empty
	if countryCode != "" {
		country.CountryCode = &countryCode
	}
	if isActive != nil {
		country.IsActive = *isActive
	}

	return s.auctionRepo.UpdateCountry(ctx, country)
}

func (s *adminService) DeleteCountry(ctx context.Context, id int) error {
	return s.auctionRepo.DeleteCountry(ctx, id)
}


func (s *adminService) AdminAddAuctionImages(ctx context.Context, auctionID uuid.UUID, urls []string) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()
	for i, url := range urls {
		if url == "" {
			continue
		}
		img := &models.AuctionImage{
			AuctionID:    auctionID,
			URL:          url,
			MediaType:    "image",
			DisplayOrder: i + 1,
		}
		if err := s.auctionRepo.AddImageTx(ctx, tx, img); err != nil {
			return fmt.Errorf("failed to save image: %w", err)
		}
	}
	return tx.Commit()
}
