package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/repository"
	"github.com/microcosm-cc/bluemonday"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

var ErrInvalidStatus = errors.New("invalid status")

// ErrRejectionNotesRequired est retournée quand un admin tente de rejeter une demande
// sans fournir de motif — un rejet sans explication ne permet pas au vendeur de corriger
// et resoumettre sa demande.
var ErrRejectionNotesRequired = errors.New("rejection_notes_required")

// ErrNotRequestOwner est retournée quand un utilisateur tente de modifier une demande
// qui ne lui appartient pas (UpdateAuctionRequest, hors admin).
var ErrNotRequestOwner = errors.New("not_request_owner")

// descriptionSanitizer retire tout HTML des descriptions saisies par l'utilisateur avant
// stockage (Product Description Phase 1) — politique stricte : aucune balise autorisée.
var descriptionSanitizer = bluemonday.StrictPolicy()

// sanitizeDescriptions nettoie en place les 3 champs de description localisés d'une
// AuctionRequest.
func sanitizeDescriptions(req *models.AuctionRequest) {
	if req.DescriptionAr != nil {
		clean := descriptionSanitizer.Sanitize(*req.DescriptionAr)
		req.DescriptionAr = &clean
	}
	if req.DescriptionFr != nil {
		clean := descriptionSanitizer.Sanitize(*req.DescriptionFr)
		req.DescriptionFr = &clean
	}
	if req.DescriptionEn != nil {
		clean := descriptionSanitizer.Sanitize(*req.DescriptionEn)
		req.DescriptionEn = &clean
	}
}

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
	// UpdateAuctionRequest applique les modifications d'un vendeur à sa propre demande
	// (draft ou rejected uniquement). isAdmin=true (via AdminUpdateAuctionRequest)
	// contourne la vérification de propriété et la restriction de statut.
	UpdateAuctionRequest(ctx context.Context, id uuid.UUID, userID uuid.UUID, updates *models.AuctionRequest) error
	AdminUpdateAuctionRequest(ctx context.Context, id uuid.UUID, updates *models.AuctionRequest) error
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

	// Product Description Phase 1 : un vendeur peut sauvegarder un brouillon ("draft")
	// avant de soumettre pour revue ("pending"). Tout autre statut envoyé par l'appelant
	// est ignoré et remplacé par "pending", pour préserver le comportement historique.
	if req.Status != "draft" && req.Status != "pending" {
		req.Status = "pending"
	}

	sanitizeDescriptions(req)

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

	// Un rejet doit toujours être motivé (Product Description Phase 1) : sans ce garde,
	// le vendeur n'a aucune information pour corriger et resoumettre sa demande. Vérifié
	// ici en plus de la validation du handler pour ne jamais dépendre uniquement de
	// l'appelant HTTP.
	if status == "rejected" && strings.TrimSpace(notes) == "" {
		return ErrRejectionNotesRequired
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

// applyAuctionRequestUpdates copie les champs modifiables de updates sur existing —
// utilisé par UpdateAuctionRequest et AdminUpdateAuctionRequest pour ne jamais laisser
// l'appelant écraser des champs de gouvernance (user_id, id, timestamps, review fields)
// directement.
func applyAuctionRequestUpdates(existing, updates *models.AuctionRequest) {
	existing.CategoryID = updates.CategoryID
	existing.LocationID = updates.LocationID
	existing.TitleAr = updates.TitleAr
	existing.TitleFr = updates.TitleFr
	existing.TitleEn = updates.TitleEn
	existing.DescriptionAr = updates.DescriptionAr
	existing.DescriptionFr = updates.DescriptionFr
	existing.DescriptionEn = updates.DescriptionEn
	existing.StartPrice = updates.StartPrice
	existing.MinIncrement = updates.MinIncrement
	existing.InsuranceAmount = updates.InsuranceAmount
	existing.ReservePrice = updates.ReservePrice
	existing.BuyNowPrice = updates.BuyNowPrice
	existing.StartDate = updates.StartDate
	existing.EndDate = updates.EndDate
	existing.Images = updates.Images
	if updates.Quantity > 0 {
		existing.Quantity = updates.Quantity
	}
}

func (s *requestService) UpdateAuctionRequest(ctx context.Context, id uuid.UUID, userID uuid.UUID, updates *models.AuctionRequest) error {
	existing, err := s.repo.GetAuctionRequestByID(ctx, id)
	if err != nil {
		return err
	}

	if existing.UserID != userID {
		return ErrNotRequestOwner
	}

	// Seul un brouillon ou une demande rejetée peut être édité par son auteur — une
	// demande "pending" ou "approved" est déjà en cours/terminée de traitement admin.
	if existing.Status != "draft" && existing.Status != "rejected" {
		return ErrInvalidStatus
	}

	wasRejected := existing.Status == "rejected"

	applyAuctionRequestUpdates(existing, updates)
	sanitizeDescriptions(existing)

	if wasRejected {
		// Resoumission après rejet : retour forcé à "pending", on efface l'ancienne
		// revue (admin_notes/reviewed_by/reviewed_at) — la demande repart à zéro dans
		// la file de revue admin.
		existing.Status = "pending"
	} else {
		// Brouillon : reste "draft", ou le statut choisi par l'appelant (typiquement
		// "pending" pour soumettre) — seules "draft"/"pending" sont acceptées ici, tout
		// le reste (approved/rejected) ne peut être atteint que via ReviewAuctionRequest.
		if updates.Status == "pending" {
			existing.Status = "pending"
		} else {
			existing.Status = "draft"
		}
	}

	if err := s.repo.UpdateAuctionRequest(ctx, existing); err != nil {
		return err
	}

	if wasRejected {
		if err := s.repo.ClearAuctionRequestReview(ctx, id, "pending"); err != nil {
			return err
		}
	}

	return nil
}

func (s *requestService) AdminUpdateAuctionRequest(ctx context.Context, id uuid.UUID, updates *models.AuctionRequest) error {
	existing, err := s.repo.GetAuctionRequestByID(ctx, id)
	if err != nil {
		return err
	}

	// Un admin peut éditer à n'importe quel statut (autorité de modération complète) —
	// ni vérification de propriété, ni restriction de statut, contrairement à
	// UpdateAuctionRequest. Le statut lui-même n'est pas modifié ici : ReviewAuctionRequest
	// reste le seul chemin pour approuver/rejeter.
	applyAuctionRequestUpdates(existing, updates)
	sanitizeDescriptions(existing)

	return s.repo.UpdateAuctionRequest(ctx, existing)
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
