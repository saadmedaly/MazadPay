package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	apperr "github.com/mazadpay/backend/internal/errors"
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
	// AdminUpdateAuctionRequest applies an admin edit. insurancePolicy is a
	// pointer specifically so "omitted from the request body" (nil) can be
	// distinguished from "explicitly set to a value" -- a plain string field
	// on updates can't express that distinction (client feedback: omitting
	// insurance_policy from an update payload must PRESERVE the existing
	// value, never silently reset a "not_required" request back to
	// "required"). nil => leave the existing policy untouched; non-nil =>
	// must be "required" or "not_required" (validated by the handler).
	AdminUpdateAuctionRequest(ctx context.Context, id uuid.UUID, updates *models.AuctionRequest, insurancePolicy *string) error
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
	userRepo            repository.UserRepository
	auditSvc            AuditService
	notificationService NotificationService
	logger              *zap.Logger
	globalHub           GlobalHub
}

func NewRequestService(repo repository.RequestRepository, auctionRepo repository.AuctionRepository, contentRepo repository.ContentRepository, userRepo repository.UserRepository, auditSvc AuditService, notificationService NotificationService, logger *zap.Logger, globalHub GlobalHub) RequestService {
	return &requestService{
		repo:                repo,
		auctionRepo:         auctionRepo,
		contentRepo:         contentRepo,
		userRepo:            userRepo,
		auditSvc:            auditSvc,
		notificationService: notificationService,
		logger:              logger,
		globalHub:           globalHub,
	}
}

// emitRequestUpdated broadcasts a Customer #20 hardening-round private event
// (request.updated) to ONLY the request's owner, over the same GlobalHub
// used for auction/content events -- BroadcastToUser, never Broadcast or
// BroadcastAuctionEvent, since a request's review outcome must never be
// visible to any user other than its owner (unlike FAQ/banner/category,
// which are public, or auction events, which are market-scoped but still
// visible to every user in that market). Called only after the triggering
// review's transaction has already committed. Coexists with, and does not
// replace, the existing SendLocalizedPush notification: FCM tells the user
// "something happened" even if the app is closed; this WS event tells an
// already-OPEN Requests page to refetch right now, without the user having
// to act on the notification at all.
func (s *requestService) emitRequestUpdated(ownerID uuid.UUID, requestID uuid.UUID) {
	if s.globalHub == nil {
		return
	}
	s.globalHub.BroadcastToUser(ownerID.String(), models.GlobalWSEvent{
		Type:       models.EventRequestUpdated,
		EntityType: "request",
		EntityID:   requestID.String(),
		UpdatedAt:  time.Now().UTC().Format(time.RFC3339),
	})
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

	// Country-scoped market (migration 000046, V1) : market_country_iso/currency_code
	// sont TOUJOURS dérivés côté serveur du compte du demandeur authentifié, jamais
	// acceptés depuis le client -- même politique de confiance que UserID (voir
	// request_handler.go). Les valeurs éventuellement déjà présentes sur req (ex.
	// injectées par erreur par un appelant) sont écrasées ici sans exception.
	if err := s.stampRequestMarket(ctx, req); err != nil {
		return err
	}

	// Client feedback A7 : insurance_amount n'est plus un champ du formulaire
	// utilisateur (seul le staff/admin le définit pendant la revue, via
	// AdminUpdateAuctionRequest). Écrasé à zéro ici, sans exception, pour qu'un
	// appelant qui enverrait tout de même insurance_amount dans le corps JSON
	// (intentionnellement ou via un client obsolète) ne puisse jamais le fixer de
	// façon autoritaire -- même politique que MarketCountryISO/CurrencyCode
	// ci-dessus. ReviewAuctionRequest refuse ensuite toute approbation tant qu'un
	// admin n'a pas explicitement défini une valeur > 0.
	req.InsuranceAmount = decimal.Zero
	// Insurance policy (migration 000048) : même politique de confiance --
	// seul l'admin peut choisir "not_required" via AdminUpdateAuctionRequest,
	// jamais l'utilisateur à la création. Écrasé explicitement à "required"
	// ici (plutôt que de compter uniquement sur le DEFAULT 'required' de la
	// colonne) pour que req.InsurancePolicy reflète la vraie valeur persistée
	// immédiatement, y compris pour tout appelant qui inspecterait req après
	// la création.
	req.InsurancePolicy = models.InsurancePolicyRequired

	// SubscriptionFee (migration 000049, client feedback #4): stamped
	// server-side from the request's category, same trust policy as
	// InsuranceAmount/MarketCountryISO/CurrencyCode above -- never accepted
	// from the client, and never derived from category name/id/icon_name
	// (none of which are stable). Falls back to the standard fee if the
	// category lookup fails, since CategoryID's own existence was already
	// validated by the DB foreign key on insert below -- a lookup error here
	// would be transient (e.g. DB hiccup), not a sign of premium pricing, so
	// failing open to the cheaper fee is safe: it never overcharges, and an
	// admin reviewing the request can still see/correct it before approval.
	category, err := s.auctionRepo.GetCategoryByID(ctx, req.CategoryID)
	if err != nil {
		s.logger.Warn("CreateAuctionRequest: failed to load category for subscription fee, defaulting to standard", zap.Int("category_id", req.CategoryID), zap.Error(err))
		req.SubscriptionFee = models.StandardSubscriptionFee
	} else {
		req.SubscriptionFee = category.SubscriptionFee()
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

// stampRequestMarket dérive market_country_iso/currency_code depuis le compte du
// demandeur authentifié (req.UserID, déjà assigné par le handler avant cet appel --
// voir request_handler.go) et les écrit sur req, en écrasant toute valeur déjà
// présente. Ne fait jamais confiance au client pour ces champs (même politique que
// UserID) : migration 000046, règle "COUNTRY is the market boundary, NOT currency".
func (s *requestService) stampRequestMarket(ctx context.Context, req *models.AuctionRequest) error {
	user, err := s.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		return err
	}
	marketISO := user.EffectiveAccountCountryISO()
	req.MarketCountryISO = &marketISO

	currencyCode := models.DefaultCurrencyCode
	if country, err := s.auctionRepo.GetCountryByCode(ctx, marketISO); err == nil && country.CurrencyCode != nil && *country.CurrencyCode != "" {
		currencyCode = *country.CurrencyCode
	}
	req.CurrencyCode = &currencyCode

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

	// Client feedback A7 follow-up, refined by the insurance_policy design
	// (migration 000048) : le formulaire utilisateur ne collecte plus
	// insurance_amount (l'utilisateur ne doit jamais le définir de façon
	// autoritaire -- seul le staff/admin le fait pendant la revue, via
	// AdminUpdateAuctionRequest). Sans cette garde, approuver une demande dont
	// insurance_amount est resté à sa valeur zéro par défaut créerait un
	// auction "active" sur lequel bid_service.go (ErrInsuranceNotSet, audit
	// V03) bloquerait ensuite TOUS les enchérisseurs -- un auction publié mais
	// structurellement inutilisable. Vérifié ici, indépendamment de toute
	// validation frontend admin, avant de créer l'auction.
	//
	// req.InsuranceRequired() défaut à true (politique "required") pour toute
	// valeur vide/inattendue -- donc CETTE garde reste pleinement active pour
	// toute demande legacy ou n'ayant jamais eu de politique explicite. Elle ne
	// s'efface QUE si un admin a explicitement mis insurance_policy =
	// "not_required" via AdminUpdateAuctionRequest (lequel force aussi
	// insurance_amount à zéro dans ce cas -- voir sa canonicalisation) :
	// c'est le seul chemin légitime pour créer un auction "active" sans
	// assurance, jamais un état par défaut ou accidentel.
	if status == "approved" && req.InsuranceRequired() && !req.InsuranceAmount.GreaterThan(decimal.Zero) {
		return apperr.ErrRequestInsuranceNotSet
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
			// InsurancePolicy (migration 000048): carried forward verbatim from
			// the approved request, same pattern as InsuranceAmount itself --
			// the policy decided during admin review is the one the live
			// auction (and therefore PlaceBid) must honor.
			InsurancePolicy: req.InsurancePolicy,
			ReservePrice:    reservePrice,
			BuyNowPrice:     req.BuyNowPrice,
			StartTime:       req.StartDate,
			EndTime:         req.EndDate,
			Status:          "active",
			Views:           0,
			BidderCount:     0,
			CreatedAt:       time.Now(),
			// MarketCountryISO/CurrencyCode (migration 000046): preserved verbatim
			// from the approved request, stamped once at request-creation time --
			// never re-derived from the seller's current account here, so an
			// auction's market/currency never silently drifts if the seller's
			// account_country_iso changes after submission.
			MarketCountryISO: req.MarketCountryISO,
			CurrencyCode:     req.CurrencyCode,
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

	s.emitRequestUpdated(req.UserID, id)

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
// applyAuctionRequestUpdates deliberately does NOT copy InsuranceAmount --
// client feedback A7: the user must never be able to authoritatively set
// insurance (whether creating or editing/resubmitting a request), only the
// admin can, via AdminUpdateAuctionRequest, which applies it separately below
// after this shared helper runs. This keeps UpdateAuctionRequest (the
// owner-only edit/resubmit path) safe from a malicious/stale payload
// regardless of what insurance_amount value it contains.
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
	categoryChanged := existing.CategoryID != updates.CategoryID

	applyAuctionRequestUpdates(existing, updates)
	sanitizeDescriptions(existing)

	// SubscriptionFee (client feedback #4): re-stamped only when the category
	// actually changed -- an edit to price/description/dates must not
	// silently recompute (and potentially change) an already-correct fee,
	// but a category change (e.g. "phones" -> "cars") must not leave a stale
	// fee from the old category on the request.
	if categoryChanged {
		category, catErr := s.auctionRepo.GetCategoryByID(ctx, existing.CategoryID)
		if catErr != nil {
			s.logger.Warn("UpdateAuctionRequest: failed to load new category for subscription fee, defaulting to standard", zap.Int("category_id", existing.CategoryID), zap.Error(catErr))
			existing.SubscriptionFee = models.StandardSubscriptionFee
		} else {
			existing.SubscriptionFee = category.SubscriptionFee()
		}
	}

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

func (s *requestService) AdminUpdateAuctionRequest(ctx context.Context, id uuid.UUID, updates *models.AuctionRequest, insurancePolicy *string) error {
	existing, err := s.repo.GetAuctionRequestByID(ctx, id)
	if err != nil {
		return err
	}

	// Un admin peut éditer à n'importe quel statut (autorité de modération complète) —
	// ni vérification de propriété, ni restriction de statut, contrairement à
	// UpdateAuctionRequest. Le statut lui-même n'est pas modifié ici : ReviewAuctionRequest
	// reste le seul chemin pour approuver/rejeter.
	applyAuctionRequestUpdates(existing, updates)
	// Client feedback A7 : contrairement à applyAuctionRequestUpdates (partagée avec le
	// chemin utilisateur, qui ne touche jamais insurance_amount), l'admin EST autorisé à
	// définir le montant de l'assurance ici -- c'est le seul chemin légitime pour le
	// faire avant que ReviewAuctionRequest n'exige une valeur > 0 pour approuver.
	existing.InsuranceAmount = updates.InsuranceAmount

	// Insurance policy (migration 000048) : insurancePolicy est un pointeur
	// pour distinguer "absent du payload" (nil -> ne pas toucher existing.
	// InsurancePolicy, qui vient d'être relu depuis la DB par GetAuctionRequestByID
	// ci-dessus et porte donc déjà la valeur actuelle) de "explicitement fourni".
	// Sans cette distinction, un appel qui omettrait insurance_policy remettrait
	// silencieusement une demande "not_required" à "required" -- exigence client
	// explicite : "omitted must preserve".
	if insurancePolicy != nil {
		existing.InsurancePolicy = *insurancePolicy
	}

	// Canonicalisation (client feedback, exigence #3) : un état "not_required +
	// montant positif" ne doit jamais pouvoir être persisté -- si la politique
	// est (ou devient) "not_required", le montant est forcé à zéro ici, sans
	// exception, indépendamment de ce que l'admin a pu saisir dans le champ
	// montant (ex. un montant laissé d'un ancien état "required").
	if !existing.InsuranceRequired() {
		existing.InsuranceAmount = decimal.Zero
	}

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

// BulkReviewAuctionRequests reviews multiple auction requests in one call.
//
// Client feedback #10 (bulk review notifications) audit found a much larger
// pre-existing gap than missing notifications: this method used to call ONLY
// BulkUpdateAuctionRequestStatus, a single blind UPDATE with no pending-status
// guard and no auction creation -- unlike ReviewAuctionRequest (the
// single-review path), which creates the actual Auction row inside a
// transaction when status="approved". That meant bulk-approving a request
// flipped its DB status to "approved" but the item never went live as an
// auction, and re-approving an already-reviewed request was silently
// possible (no ErrRequestAlreadyReviewed guard). Sending an "approved!" push
// notification on top of that gap would have actively misled sellers about
// an item that was never actually published.
//
// Fixed by delegating each id to the existing, already-correct
// ReviewAuctionRequest -- reusing its transaction safety, pending-status
// guard, insurance guard, real Auction creation, audit log, and
// SendLocalizedPush call verbatim, rather than duplicating that logic here.
// This also automatically prevents duplicate auction creation and duplicate
// notifications: ReviewAuctionRequest's own req.Status != "pending" guard
// means retrying the same bulk call, or passing a duplicate id twice in one
// call, only ever processes (and notifies for) each request once -- the
// second attempt hits ErrRequestAlreadyReviewed and is skipped, not retried
// or re-notified. One request's failure (already reviewed, insurance not
// set, not found, DB error) does not stop the rest of the batch from being
// processed, matching this method's pre-existing "continue on per-item audit
// failure" tolerance.
func (s *requestService) BulkReviewAuctionRequests(ctx context.Context, ids []uuid.UUID, status, notes string, reviewedBy uuid.UUID) error {
	if status != "approved" && status != "rejected" {
		return ErrInvalidStatus
	}

	seen := make(map[uuid.UUID]bool, len(ids))
	for _, id := range ids {
		if seen[id] {
			continue
		}
		seen[id] = true

		if err := s.ReviewAuctionRequest(ctx, id, status, notes, reviewedBy); err != nil {
			if s.logger != nil {
				s.logger.Warn("BulkReviewAuctionRequests: skipping request that could not be reviewed", zap.String("request_id", id.String()), zap.Error(err))
			}
			continue
		}

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
	// Business validation (client feedback: Bug A.1) -- ends_at must be
	// strictly after starts_at (a zero-length window is not a valid ad
	// display period), returning apperr.ErrBadRequest (an existing,
	// reusable "generic bad request" domain error -- see MapError's
	// "bad_request" case) so this surfaces as HTTP 400, not an unmapped
	// plain error falling through to a 500.
	if !req.EndsAt.After(req.StartsAt) {
		return apperr.ErrBadRequest
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

	// Client feedback #10 (bulk review notifications) audit: unlike
	// ReviewAuctionRequest, this method had no pending-status guard --
	// re-reviewing an already-approved request would create a SECOND Banner
	// row and send a second "approved" notification every time it was
	// called. Now that BulkReviewBannerRequests delegates each id here (same
	// fix as BulkReviewAuctionRequests), this guard is what makes both the
	// single and bulk paths idempotent: retrying the same review, or passing
	// a duplicate id twice in one bulk call, is a no-op on the second
	// attempt rather than a duplicate creation/notification.
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

	s.emitRequestUpdated(req.UserID, id)

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

// BulkReviewBannerRequests reviews multiple banner requests in one call.
//
// Same client feedback #10 fix as BulkReviewAuctionRequests above -- this
// used to call ONLY BulkUpdateBannerRequestStatus (a blind UPDATE, no
// pending-status guard, no Banner creation), unlike ReviewBannerRequest
// (single-review), which creates the actual Banner row on approval. Fixed by
// delegating each id to ReviewBannerRequest, for the identical reasons: real
// banner creation, the pending-status guard (preventing duplicate
// creation/notification on retry or a duplicate id in the same call), and
// correct SendLocalizedPush usage, all reused verbatim rather than
// duplicated here.
func (s *requestService) BulkReviewBannerRequests(ctx context.Context, ids []uuid.UUID, status, notes string, reviewedBy uuid.UUID) error {
	if status != "approved" && status != "rejected" {
		return ErrInvalidStatus
	}

	seen := make(map[uuid.UUID]bool, len(ids))
	for _, id := range ids {
		if seen[id] {
			continue
		}
		seen[id] = true

		if err := s.ReviewBannerRequest(ctx, id, status, notes, reviewedBy); err != nil {
			if s.logger != nil {
				s.logger.Warn("BulkReviewBannerRequests: skipping request that could not be reviewed", zap.String("request_id", id.String()), zap.Error(err))
			}
			continue
		}

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
