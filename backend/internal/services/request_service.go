package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/repository"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

var ErrInvalidStatus = errors.New("invalid status")

// ErrRequestAlreadyReviewed est retournée quand une demande n'est plus "pending"
// (Requests Phase 1) : empêche de ré-approuver/ré-rejeter une demande déjà
// traitée, ce qui aurait pu créer un auction/banner en double.
var ErrRequestAlreadyReviewed = errors.New("request_already_reviewed")

type RequestService interface {
	CreateAuctionRequest(ctx context.Context, req *models.AuctionRequest) error
	GetAuctionRequests(ctx context.Context, status string, userID *uuid.UUID, dateFrom, dateTo *time.Time, categoryID, locationID *int, minPrice, maxPrice *float64, sortBy, sortOrder string, page, perPage int) ([]models.AuctionRequest, int, error)
	GetAuctionRequestByID(ctx context.Context, id uuid.UUID) (*models.AuctionRequest, error)
	GetUserAuctionRequests(ctx context.Context, userID uuid.UUID, status string, page, perPage int) ([]models.AuctionRequest, int, error)
	ReviewAuctionRequest(ctx context.Context, id uuid.UUID, status, notes string, reviewedBy uuid.UUID) error
	DeleteAuctionRequest(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error
	BulkReviewAuctionRequests(ctx context.Context, ids []uuid.UUID, status, notes string, reviewedBy uuid.UUID) error
	BulkDeleteAuctionRequests(ctx context.Context, ids []uuid.UUID, deletedBy uuid.UUID) error

	// Banner Requests
	CreateBannerRequest(ctx context.Context, req *models.BannerRequest) error
	GetBannerRequests(ctx context.Context, status string, userID *uuid.UUID, dateFrom, dateTo *time.Time, sortBy, sortOrder string, page, perPage int) ([]models.BannerRequest, int, error)
	GetBannerRequestByID(ctx context.Context, id uuid.UUID) (*models.BannerRequest, error)
	GetUserBannerRequests(ctx context.Context, userID uuid.UUID, status string, page, perPage int) ([]models.BannerRequest, int, error)
	ReviewBannerRequest(ctx context.Context, id uuid.UUID, status, notes string, reviewedBy uuid.UUID) error
	DeleteBannerRequest(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error
	BulkReviewBannerRequests(ctx context.Context, ids []uuid.UUID, status, notes string, reviewedBy uuid.UUID) error
	BulkDeleteBannerRequests(ctx context.Context, ids []uuid.UUID, deletedBy uuid.UUID) error
}

type requestService struct {
	repo                repository.RequestRepository
	auctionRepo         repository.AuctionRepository
	contentRepo         repository.ContentRepository
	auditSvc            AuditService
	notificationService NotificationService
	logger              *zap.Logger
}

func NewRequestService(repo repository.RequestRepository, auctionRepo repository.AuctionRepository, contentRepo repository.ContentRepository, auditSvc AuditService, notificationService NotificationService, logger *zap.Logger) RequestService {
	return &requestService{
		repo:                repo,
		auctionRepo:         auctionRepo,
		contentRepo:         contentRepo,
		auditSvc:            auditSvc,
		notificationService: notificationService,
		logger:              logger,
	}
}

// Auction Requests
func (s *requestService) CreateAuctionRequest(ctx context.Context, req *models.AuctionRequest) error {
	// Business validation
	if req.EndDate.Before(req.StartDate) {
		return errors.New("end_date must be after start_date")
	}
	if req.ReservePrice != nil && req.ReservePrice.LessThan(req.StartPrice) {
		return errors.New("reserve_price must be greater than or equal to start_price")
	}
	if req.BuyNowPrice != nil && req.BuyNowPrice.LessThan(req.StartPrice) {
		return errors.New("buy_now_price must be greater than or equal to start_price")
	}

	req.Status = "pending"
	if err := s.repo.CreateAuctionRequest(ctx, req); err != nil {
		return err
	}

	// Notify admins via WebSocket
	if s.notificationService != nil {
		s.notificationService.NotifyNewAuctionRequest(
			req.ID.String(),
			req.UserID.String(),
			"", // user name will be fetched by frontend
			req.TitleAr,
		)
	}

	return nil
}

func (s *requestService) GetAuctionRequests(ctx context.Context, status string, userID *uuid.UUID, dateFrom, dateTo *time.Time, categoryID, locationID *int, minPrice, maxPrice *float64, sortBy, sortOrder string, page, perPage int) ([]models.AuctionRequest, int, error) {
	return s.repo.GetAuctionRequests(ctx, status, userID, dateFrom, dateTo, categoryID, locationID, minPrice, maxPrice, sortBy, sortOrder, page, perPage)
}

func (s *requestService) GetAuctionRequestByID(ctx context.Context, id uuid.UUID) (*models.AuctionRequest, error) {
	return s.repo.GetAuctionRequestByID(ctx, id)
}

func (s *requestService) GetUserAuctionRequests(ctx context.Context, userID uuid.UUID, status string, page, perPage int) ([]models.AuctionRequest, int, error) {
	return s.repo.GetUserAuctionRequests(ctx, userID, status, page, perPage)
}

func (s *requestService) ReviewAuctionRequest(ctx context.Context, id uuid.UUID, status, notes string, reviewedBy uuid.UUID) error {
	if status != "approved" && status != "rejected" {
		return ErrInvalidStatus
	}

	// Get the request details first
	req, err := s.repo.GetAuctionRequestByID(ctx, id)
	if err != nil {
		return err
	}

	// Refuse de re-réviser une demande déjà traitée (Requests Phase 1) : sans ce
	// garde, ré-approuver une demande déjà "approved" créerait un second auction en
	// double, et le statut ne peut légitimement transiter qu'une seule fois depuis
	// "pending".
	if req.Status != "pending" {
		return ErrRequestAlreadyReviewed
	}

	// Begin transaction
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update the request status within transaction
	if err := s.repo.UpdateAuctionRequestStatusTx(ctx, tx, id, status, notes, reviewedBy); err != nil {
		return err
	}

	// If approved, create the actual auction within the same transaction
	var auctionID uuid.UUID
	if status == "approved" {
		reservePrice := decimal.Decimal{}
		if req.ReservePrice != nil {
			reservePrice = *req.ReservePrice
		} else {
			reservePrice = req.StartPrice
		}

		auction := &models.Auction{
			ID:              uuid.New(),
			SellerID:        req.UserID,
			CategoryID:      req.CategoryID,
			LocationID:      req.LocationID,
			TitleAr:         req.TitleAr,
			TitleFr:         req.TitleFr,
			TitleEn:         req.TitleEn,
			DescriptionAr:   req.DescriptionAr,
			DescriptionFr:   req.DescriptionFr,
			DescriptionEn:   req.DescriptionEn,
			StartPrice:      req.StartPrice,
			CurrentPrice:    req.StartPrice,
			MinIncrement:    req.MinIncrement,
			InsuranceAmount: req.InsuranceAmount,
			ReservePrice:    reservePrice,
			BuyNowPrice:     req.BuyNowPrice,
			StartTime:       req.StartDate,
			EndTime:         req.EndDate,
			Status:          "active",
			Views:           0,
			BidderCount:     0,
			CreatedAt:       time.Now(),
		}

		if err := s.auctionRepo.Create(ctx, tx, auction); err != nil {
			return err
		}
		auctionID = auction.ID
	}

	// Commit transaction (Requests Phase 1 — manquait entièrement auparavant : sans
	// cet appel, defer tx.Rollback() annulait silencieusement la mise à jour de
	// statut et la création de l'auction, malgré une réponse de succès au client).
	if err := tx.Commit(); err != nil {
		return err
	}

	// Log audit
	if auditErr := s.auditSvc.Log(ctx, reviewedBy, fmt.Sprintf("auction_request_reviewed_%s", status), "auction_request", &id,
		fmt.Sprintf("Status changed to %s. Notes: %s", status, notes)); auditErr != nil {
		if s.logger != nil {
			s.logger.Error("ReviewAuctionRequest: failed to write audit log", zap.String("request_id", id.String()), zap.Error(auditErr))
		}
	}

	// Send localized notification (outside transaction)
	if status == "approved" {
		if req.User != nil {
			language := "ar"
			if req.User.LanguagePref != "" {
				language = req.User.LanguagePref
			}
			params := map[string]string{
				"auctionTitle": req.TitleAr,
			}
			data := map[string]string{
				"request_id": req.ID.String(),
				"auction_id": auctionID.String(),
			}
			s.notificationService.SendLocalizedPush(ctx, req.UserID, "auction_approved", language, params, data)
		}
	} else {
		if req.User != nil {
			language := "ar"
			if req.User.LanguagePref != "" {
				language = req.User.LanguagePref
			}
			params := map[string]string{
				"auctionTitle": req.TitleAr,
				"reason":       notes,
			}
			data := map[string]string{
				"request_id": req.ID.String(),
			}
			s.notificationService.SendLocalizedPush(ctx, req.UserID, "auction_rejected", language, params, data)
		}
	}

	return nil
}

func (s *requestService) DeleteAuctionRequest(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error {
	if err := s.repo.DeleteAuctionRequest(ctx, id); err != nil {
		return err
	}

	// Log audit
	if auditErr := s.auditSvc.Log(ctx, deletedBy, "auction_request_deleted", "auction_request", &id, "Auction request deleted"); auditErr != nil {
		if s.logger != nil {
			s.logger.Error("DeleteAuctionRequest: failed to write audit log", zap.String("request_id", id.String()), zap.Error(auditErr))
		}
	}

	return nil
}

func (s *requestService) BulkReviewAuctionRequests(ctx context.Context, ids []uuid.UUID, status, notes string, reviewedBy uuid.UUID) error {
	if status != "approved" && status != "rejected" {
		return ErrInvalidStatus
	}
	if err := s.repo.BulkUpdateAuctionRequestStatus(ctx, ids, status, notes, reviewedBy); err != nil {
		return err
	}

	// Log audit for each request
	for _, id := range ids {
		if auditErr := s.auditSvc.Log(ctx, reviewedBy, fmt.Sprintf("auction_requests_bulk_reviewed_%s", status), "auction_request", &id,
			fmt.Sprintf("Bulk status changed to %s. Notes: %s", status, notes)); auditErr != nil {
			if s.logger != nil {
				s.logger.Error("BulkReviewAuctionRequests: failed to write audit log", zap.String("request_id", id.String()), zap.Error(auditErr))
			}
		}
	}

	return nil
}

func (s *requestService) BulkDeleteAuctionRequests(ctx context.Context, ids []uuid.UUID, deletedBy uuid.UUID) error {
	if err := s.repo.BulkDeleteAuctionRequests(ctx, ids); err != nil {
		return err
	}

	// Log audit for each request
	for _, id := range ids {
		if auditErr := s.auditSvc.Log(ctx, deletedBy, "auction_requests_bulk_deleted", "auction_request", &id, "Bulk deleted auction request"); auditErr != nil {
			if s.logger != nil {
				s.logger.Error("BulkDeleteAuctionRequests: failed to write audit log", zap.String("request_id", id.String()), zap.Error(auditErr))
			}
		}
	}

	return nil
}

// Banner Requests
func (s *requestService) CreateBannerRequest(ctx context.Context, req *models.BannerRequest) error {
	// Business validation
	if req.EndsAt.Before(req.StartsAt) {
		return errors.New("ends_at must be after starts_at")
	}

	req.Status = "pending"
	if err := s.repo.CreateBannerRequest(ctx, req); err != nil {
		return err
	}

	// Notify admins via WebSocket
	if s.notificationService != nil {
		s.notificationService.NotifyNewBannerRequest(
			req.ID.String(),
			req.UserID.String(),
			"", // user name will be fetched by frontend
			req.TitleAr,
		)
	}

	return nil
}

func (s *requestService) GetBannerRequests(ctx context.Context, status string, userID *uuid.UUID, dateFrom, dateTo *time.Time, sortBy, sortOrder string, page, perPage int) ([]models.BannerRequest, int, error) {
	return s.repo.GetBannerRequests(ctx, status, userID, dateFrom, dateTo, sortBy, sortOrder, page, perPage)
}

func (s *requestService) GetBannerRequestByID(ctx context.Context, id uuid.UUID) (*models.BannerRequest, error) {
	return s.repo.GetBannerRequestByID(ctx, id)
}

func (s *requestService) GetUserBannerRequests(ctx context.Context, userID uuid.UUID, status string, page, perPage int) ([]models.BannerRequest, int, error) {
	return s.repo.GetUserBannerRequests(ctx, userID, status, page, perPage)
}

func (s *requestService) ReviewBannerRequest(ctx context.Context, id uuid.UUID, status, notes string, reviewedBy uuid.UUID) error {
	if status != "approved" && status != "rejected" {
		return ErrInvalidStatus
	}

	// Get the request details first
	req, err := s.repo.GetBannerRequestByID(ctx, id)
	if err != nil {
		return err
	}

	// Begin transaction
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update the request status within transaction
	if err := s.repo.UpdateBannerRequestStatusTx(ctx, tx, id, status, notes, reviewedBy); err != nil {
		return err
	}

	// If approved, create the actual banner within the same transaction
	var bannerID int
	if status == "approved" {
		titleFr := ""
		titleEn := ""
		targetURL := ""
		if req.TitleFr != nil {
			titleFr = *req.TitleFr
		}
		if req.TitleEn != nil {
			titleEn = *req.TitleEn
		}
		if req.TargetURL != nil {
			targetURL = *req.TargetURL
		}

		banner := &models.Banner{
			TitleAr:      req.TitleAr,
			TitleFr:      titleFr,
			TitleEn:      titleEn,
			ImageURL:     req.ImageURL,
			TargetURL:    targetURL,
			IsActive:     true,
			StartsAt:     &req.StartsAt,
			EndsAt:       &req.EndsAt,
			DisplayOrder: 0,
		}

		if err := s.contentRepo.CreateBannerTx(ctx, tx, banner); err != nil {
			return err
		}
		bannerID = banner.ID
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return err
	}

	// Log audit
	if auditErr := s.auditSvc.Log(ctx, reviewedBy, fmt.Sprintf("banner_request_reviewed_%s", status), "banner_request", &id,
		fmt.Sprintf("Status changed to %s. Notes: %s", status, notes)); auditErr != nil {
		if s.logger != nil {
			s.logger.Error("ReviewBannerRequest: failed to write audit log", zap.String("request_id", id.String()), zap.Error(auditErr))
		}
	}

	// Send localized notification (outside transaction)
	if status == "approved" {
		language := "ar"
		if req.User != nil && req.User.LanguagePref != "" {
			language = req.User.LanguagePref
		}
		params := map[string]string{
			"bannerTitle": req.TitleAr,
		}
		data := map[string]string{
			"request_id": req.ID.String(),
			"banner_id":  fmt.Sprintf("%d", bannerID),
		}
		s.notificationService.SendLocalizedPush(ctx, req.UserID, "banner_approved", language, params, data)
	} else {
		language := "ar"
		if req.User != nil && req.User.LanguagePref != "" {
			language = req.User.LanguagePref
		}
		params := map[string]string{
			"bannerTitle": req.TitleAr,
			"reason":      notes,
		}
		data := map[string]string{
			"request_id": req.ID.String(),
		}
		s.notificationService.SendLocalizedPush(ctx, req.UserID, "banner_rejected", language, params, data)
	}

	return nil
}

func (s *requestService) DeleteBannerRequest(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error {
	if err := s.repo.DeleteBannerRequest(ctx, id); err != nil {
		return err
	}

	// Log audit
	if auditErr := s.auditSvc.Log(ctx, deletedBy, "banner_request_deleted", "banner_request", &id, "Banner request deleted"); auditErr != nil {
		if s.logger != nil {
			s.logger.Error("DeleteBannerRequest: failed to write audit log", zap.String("request_id", id.String()), zap.Error(auditErr))
		}
	}

	return nil
}

func (s *requestService) BulkReviewBannerRequests(ctx context.Context, ids []uuid.UUID, status, notes string, reviewedBy uuid.UUID) error {
	if status != "approved" && status != "rejected" {
		return ErrInvalidStatus
	}
	if err := s.repo.BulkUpdateBannerRequestStatus(ctx, ids, status, notes, reviewedBy); err != nil {
		return err
	}

	// Log audit for each request
	for _, id := range ids {
		if auditErr := s.auditSvc.Log(ctx, reviewedBy, fmt.Sprintf("banner_requests_bulk_reviewed_%s", status), "banner_request", &id,
			fmt.Sprintf("Bulk status changed to %s. Notes: %s", status, notes)); auditErr != nil {
			if s.logger != nil {
				s.logger.Error("BulkReviewBannerRequests: failed to write audit log", zap.String("request_id", id.String()), zap.Error(auditErr))
			}
		}
	}

	return nil
}

func (s *requestService) BulkDeleteBannerRequests(ctx context.Context, ids []uuid.UUID, deletedBy uuid.UUID) error {
	if err := s.repo.BulkDeleteBannerRequests(ctx, ids); err != nil {
		return err
	}

	// Log audit for each request
	for _, id := range ids {
		if auditErr := s.auditSvc.Log(ctx, deletedBy, "banner_requests_bulk_deleted", "banner_request", &id, "Bulk deleted banner request"); auditErr != nil {
			if s.logger != nil {
				s.logger.Error("BulkDeleteBannerRequests: failed to write audit log", zap.String("request_id", id.String()), zap.Error(auditErr))
			}
		}
	}

	return nil
}
