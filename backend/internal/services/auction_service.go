package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	apperr "github.com/mazadpay/backend/internal/errors"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/repository"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type CreateAuctionInput struct {
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
	LotNumber       string
	PhoneContact    string
	ItemDetails     models.JSONB
	BuyNowPrice     *decimal.Decimal
	Images          []string
	Condition       *string
	Brand           *string
	VideoURL        *string
	Quantity        int
}

type AuctionService interface {
	GetByID(ctx context.Context, id uuid.UUID) (*models.Auction, []models.AuctionImage, error)
	// List's second return value is the real total matching f, ignoring
	// pagination (client feedback #12 tab counts).
	List(ctx context.Context, f repository.AuctionFilters) ([]models.Auction, int, error)

	Create(ctx context.Context, sellerID uuid.UUID, input CreateAuctionInput) (*models.Auction, error)
	ReportAuction(ctx context.Context, auctionID, reporterID uuid.UUID, reason string) error
	Update(ctx context.Context, id uuid.UUID, input CreateAuctionInput) (*models.Auction, error)
	UpdateAuction(ctx context.Context, id uuid.UUID, userID uuid.UUID, input CreateAuctionInput) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteAuction(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	GetBidStatus(ctx context.Context, auctionID, userID uuid.UUID) (map[string]interface{}, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error)
	IncrementViews(ctx context.Context, id uuid.UUID, userID *uuid.UUID) error
	AddImages(ctx context.Context, auctionID, sellerID uuid.UUID, urls []string) error
	BuyNow(ctx context.Context, auctionID, buyerID uuid.UUID) (*models.Auction, error)
	CancelAuction(ctx context.Context, auctionID, sellerID uuid.UUID, reason string) error
	RelistAuction(ctx context.Context, auctionID, sellerID uuid.UUID, newEndTime time.Time) error
	ExtendAuction(ctx context.Context, auctionID uuid.UUID, sellerID uuid.UUID, hours int) error
	CloseExpiredAuctions(ctx context.Context) error
	// FinalizeExpiredAuction (Customer #23 canonical finalizer) is the SINGLE
	// authoritative operation for transitioning one expired auction from
	// active -> ended, replacing the two previously-independent
	// implementations (AuctionScheduler.checkEndedAuctions's own winner
	// logic, and this service's own CloseExpiredAuctions winner logic).
	// Race-safe: see auctionRepo.TrySetWinnerAtomically/TryEndAuctionAtomically
	// for why concurrent callers for the same auction ID cannot both "win".
	// Notifications/realtime events are only ever sent by whichever single
	// caller's transaction actually performed the transition.
	FinalizeExpiredAuction(ctx context.Context, auctionID uuid.UUID) error
	GetCategories(ctx context.Context) ([]models.Category, error)
	GetLocations(ctx context.Context) ([]models.Location, error)
	GetCountries(ctx context.Context) ([]models.Country, error)
	GetLocationsByCountry(ctx context.Context, countryID int) ([]models.Location, error)
}

type auctionService struct {
	db          *sqlx.DB
	auctionRepo repository.AuctionRepository
	bidRepo     repository.BidRepository
	reportRepo  repository.ReportRepository
	notifSvc    NotificationService
	userRepo    repository.UserRepository
	mediaSvc    MediaService
	rdb         *redis.Client
	walletRepo  repository.WalletRepository
	auditSvc    AuditService
	logger      *zap.Logger
	globalHub   GlobalHub
}

func NewAuctionService(db *sqlx.DB, auctionRepo repository.AuctionRepository, bidRepo repository.BidRepository, reportRepo repository.ReportRepository, notifSvc NotificationService, userRepo repository.UserRepository, mediaSvc MediaService, rdb *redis.Client, walletRepo repository.WalletRepository, auditSvc AuditService, logger *zap.Logger, globalHub GlobalHub) AuctionService {
	return &auctionService{
		db:          db,
		auctionRepo: auctionRepo,
		bidRepo:     bidRepo,
		reportRepo:  reportRepo,
		notifSvc:    notifSvc,
		userRepo:    userRepo,
		mediaSvc:    mediaSvc,
		rdb:         rdb,
		walletRepo:  walletRepo,
		auditSvc:    auditSvc,
		logger:      logger,
		globalHub:   globalHub,
	}
}

// emitAuctionEvent broadcasts a Customer #20 global invalidation event for
// auction over globalHub, scoped to the auction's own market (see
// GlobalHub.BroadcastAuctionEvent). Called ONLY after the triggering DB
// write has already committed successfully -- every call site below sits
// after its transaction's tx.Commit()/repo call returns nil, never before.
// Never fails the caller: globalHub is best-effort real-time delivery, REST
// remains the source of truth (mobile catches up via reconnect/resume
// refetch even if this silently no-ops, e.g. globalHub == nil in tests).
func (s *auctionService) emitAuctionEvent(auction *models.Auction, eventType string) {
	if s.globalHub == nil || auction == nil {
		return
	}
	s.globalHub.BroadcastAuctionEvent(auction.EffectiveMarketCountryISO(), models.GlobalWSEvent{
		Type:       eventType,
		EntityType: "auction",
		EntityID:   auction.ID.String(),
		UpdatedAt:  time.Now().UTC().Format(time.RFC3339),
	})
}

// PubliclyVisibleAuctionStatuses liste les statuts qu'un visiteur anonyme
// peut consulter sur la page de détail publique (International Auth /
// Product Review Phase). Utilisé uniquement par le handler de l'endpoint
// public GET /auctions/:id (voir auction_handler.go GetByID) — GetByID au
// niveau service reste volontairement sans filtre de statut, car il est
// aussi utilisé par des flux internes légitimes qui doivent voir un auction
// "pending" (upload d'images pendant la création, contrôle de propriété,
// etc. — voir handleMultipartImages).
var PubliclyVisibleAuctionStatuses = map[string]bool{
	"active": true,
	"ended":  true,
}

// imagesUpdateIncludesNewURLs reports whether an update payload's Images
// slice actually contains at least one real (non-empty) URL. Shared by
// Update and UpdateAuction (client feedback, R2 deletion-path audit): both
// must only wipe/replace auction_images and delete the corresponding R2
// objects when the caller genuinely supplied new images -- never on a plain
// field edit (title/price/description/duration) where Images is empty or
// nil, which is the case on every real call from AuctionHandler.Update
// (PUT /auctions/:id), since that handler's own request DTO never parses an
// "images" field from the body at all. Exported behavior, not exported
// symbol -- kept unexported since it's an internal helper, but factored out
// so its exact decision logic is tested directly rather than duplicated
// inline at each call site.
func imagesUpdateIncludesNewURLs(images []string) bool {
	for _, url := range images {
		if url != "" {
			return true
		}
	}
	return false
}

func (s *auctionService) GetByID(ctx context.Context, id uuid.UUID) (*models.Auction, []models.AuctionImage, error) {
	auction, err := s.auctionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, nil, apperr.ErrNotFound
	}
	images, _ := s.auctionRepo.GetImages(ctx, id)
	return auction, images, nil
}

func (s *auctionService) List(ctx context.Context, f repository.AuctionFilters) ([]models.Auction, int, error) {
	return s.auctionRepo.FindAll(ctx, f)
}

func (s *auctionService) Create(ctx context.Context, sellerID uuid.UUID, input CreateAuctionInput) (*models.Auction, error) {
	// Minimum: end_time must be at least 1 minute in the future
	if input.EndTime.Before(time.Now().Add(1 * time.Minute)) {
		return nil, fmt.Errorf("end_time must be at least 1 minute in the future (got: %s)", input.EndTime.Format(time.RFC3339))
	}

	// Note: Image validation is intentionally skipped here because images are uploaded
	// separately after auction creation via the multipart endpoint POST /v1/api/auctions/:id/images

	// Vérifier que la catégorie existe
	if input.CategoryID > 0 {
		categories, err := s.auctionRepo.GetCategories(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch categories: %w", err)
		}
		categoryExists := false
		for _, c := range categories {
			if c.ID == input.CategoryID {
				categoryExists = true
				break
			}
		}
		if !categoryExists {
			return nil, fmt.Errorf("category with ID %d does not exist", input.CategoryID)
		}
	}

	// Vérifier que la location existe si fournie
	if input.LocationID != nil && *input.LocationID > 0 {
		locations, err := s.auctionRepo.GetLocations(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch locations: %w", err)
		}
		locationExists := false
		for _, l := range locations {
			if l.ID == *input.LocationID {
				locationExists = true
				break
			}
		}
		if !locationExists {
			return nil, fmt.Errorf("location with ID %d does not exist", *input.LocationID)
		}
	}

	var st time.Time
	if input.StartTime != nil {
		st = *input.StartTime
	} else {
		st = time.Now()
	}

	var lotNum *string
	if input.LotNumber != "" {
		lotNum = &input.LotNumber
	}

	var phone *string
	if input.PhoneContact != "" {
		phone = &input.PhoneContact
	}

	var tFr, tEn, dAr, dFr, dEn *string
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

	auction := &models.Auction{
		ID:              uuid.New(),
		SellerID:        sellerID,
		CategoryID:      input.CategoryID,
		SubCategoryID:   input.SubCategoryID,
		LocationID:      input.LocationID,
		TitleAr:         input.TitleAr,
		TitleFr:         tFr,
		TitleEn:         tEn,
		DescriptionAr:   dAr,
		DescriptionFr:   dFr,
		DescriptionEn:   dEn,
		StartPrice:      input.StartPrice,
		CurrentPrice:    input.StartPrice,
		MinIncrement:    input.MinIncrement,
		InsuranceAmount: input.InsuranceAmount,
		StartTime:       st,
		EndTime:         input.EndTime,
		Status:          "pending", // Admin doit valider
		LotNumber:       lotNum,
		PhoneContact:    phone,
		ItemDetails:     input.ItemDetails,
		BuyNowPrice:     input.BuyNowPrice,
		Condition:       input.Condition,
		Brand:           input.Brand,
		VideoURL:        input.VideoURL,
		Quantity:        input.Quantity,
		Version:         1,
		// InsurancePolicy (migration 000048): must be stamped explicitly --
		// the INSERT includes this column, so an unset Go zero value ("")
		// is sent as-is rather than falling back to the column's DB DEFAULT
		// 'required', which violates chk_auctions_insurance_policy. Same
		// trust policy as RequestService.CreateAuctionRequest: only an
		// admin can later flip this to 'not_required', never the seller at
		// creation time.
		InsurancePolicy: models.InsurancePolicyRequired,
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	if err := s.auctionRepo.Create(ctx, tx, auction); err != nil {
		return nil, err
	}

	// Sauvegarder les images
	for i, url := range input.Images {
		if url == "" {
			continue
		}
		img := &models.AuctionImage{
			AuctionID:    auction.ID,
			URL:          url,
			MediaType:    "image",
			DisplayOrder: i + 1,
		}
		if err := s.auctionRepo.AddImageTx(ctx, tx, img); err != nil {
			return nil, fmt.Errorf("failed to save image URL to database: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Customer #20: a newly created auction is still "pending" (not yet
	// admin-approved) so it isn't visible on Home/Active lists yet -- no
	// broadcast here. The visible-to-mobile transition happens in
	// adminService.ValidateAuction (approve -> "active"), which emits
	// EventAuctionStatusChanged instead.

	return auction, nil
}

func (s *auctionService) Update(ctx context.Context, id uuid.UUID, input CreateAuctionInput) (*models.Auction, error) {
	auction, err := s.auctionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, apperr.ErrNotFound
	}

	// Vérifier que la catégorie existe
	if input.CategoryID > 0 {
		categories, err := s.auctionRepo.GetCategories(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch categories: %w", err)
		}
		categoryExists := false
		for _, c := range categories {
			if c.ID == input.CategoryID {
				categoryExists = true
				break
			}
		}
		if !categoryExists {
			return nil, fmt.Errorf("category with ID %d does not exist", input.CategoryID)
		}
	}

	// Vérifier que la location existe si fournie
	if input.LocationID != nil && *input.LocationID > 0 {
		locations, err := s.auctionRepo.GetLocations(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch locations: %w", err)
		}
		locationExists := false
		for _, l := range locations {
			if l.ID == *input.LocationID {
				locationExists = true
				break
			}
		}
		if !locationExists {
			return nil, fmt.Errorf("location with ID %d does not exist", *input.LocationID)
		}
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

	auction.CategoryID = input.CategoryID
	auction.SubCategoryID = input.SubCategoryID
	auction.LocationID = input.LocationID
	auction.TitleAr = input.TitleAr
	auction.TitleFr = tFr
	auction.TitleEn = tEn
	auction.DescriptionAr = dAr
	auction.DescriptionFr = dFr
	auction.DescriptionEn = dEn
	auction.StartPrice = input.StartPrice
	auction.MinIncrement = input.MinIncrement
	auction.InsuranceAmount = input.InsuranceAmount
	auction.EndTime = input.EndTime
	auction.PhoneContact = phone
	auction.BuyNowPrice = input.BuyNowPrice
	auction.ItemDetails = input.ItemDetails
	auction.Condition = input.Condition
	auction.Brand = input.Brand
	auction.VideoURL = input.VideoURL
	auction.Quantity = input.Quantity
	if input.StartTime != nil {
		auction.StartTime = *input.StartTime
	}

	// Image bug fix (client feedback, R2 deletion-path audit): only touch
	// auction_images/R2 when the caller actually supplied at least one
	// non-empty URL. This handler's own DTO chain (AuctionHandler.Update,
	// PUT /auctions/:id) never even parses an "images" field from the
	// request body, so `input.Images` was ALWAYS empty on every real call --
	// yet this function used to unconditionally wipe existing auction_images
	// rows and delete the corresponding R2 objects regardless. Image
	// management belongs exclusively to the dedicated POST /auctions/:id/images
	// endpoint (AddImages); a plain field-edit update must never touch images
	// at all. Same principle already used correctly by adminService.
	// UpdateAuction's hasNewImages guard.
	hasNewImages := imagesUpdateIncludesNewURLs(input.Images)

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	if err := s.auctionRepo.Update(ctx, auction); err != nil {
		return nil, err
	}

	var existingImages []models.AuctionImage
	if hasNewImages {
		// Get existing images before update (for R2 cleanup) -- only needed
		// when we're actually about to replace them.
		existingImages, err = s.auctionRepo.GetImages(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("failed to get existing images: %w", err)
		}

		// Synchroniser les images (vider et re-remplir) -- uniquement quand
		// l'appelant a explicitement fourni au moins une nouvelle image.
		if err := s.auctionRepo.DeleteImagesTx(ctx, tx, id); err != nil {
			return nil, fmt.Errorf("failed to clear images: %w", err)
		}

		for i, url := range input.Images {
			if url == "" {
				continue
			}
			if err := s.auctionRepo.AddImageTx(ctx, tx, &models.AuctionImage{
				AuctionID:    id,
				URL:          url,
				MediaType:    "image",
				DisplayOrder: i + 1,
			}); err != nil {
				return nil, fmt.Errorf("failed to update image URL in database: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// After successful commit, delete old images from R2 (best effort, don't fail if error)
	if hasNewImages && s.mediaSvc != nil {
		for _, img := range existingImages {
			if err := s.mediaSvc.DeleteFile(ctx, img.URL); err != nil {
				// Log but don't fail - the DB update succeeded
				fmt.Printf("[AuctionService Update] Warning: failed to delete old image from R2: %s, error: %v\n", img.URL, err)
			}
		}
	}

	return auction, nil
}

func (s *auctionService) ReportAuction(ctx context.Context, auctionID, reporterID uuid.UUID, reason string) error {
	report := &models.Report{
		ID:         uuid.New(),
		AuctionID:  &auctionID,
		ReporterID: reporterID,
		Reason:     reason,
		Type:       "auction",
		Status:     "pending",
	}

	if err := s.reportRepo.Create(ctx, report); err != nil {
		return err
	}

	// Notifier les admins (localized)
	if s.notifSvc != nil {
		go func() {
			auction, err := s.auctionRepo.FindByID(context.Background(), auctionID)
			if err != nil {
				return
			}
			title := auction.TitleAr
			if auction.TitleFr != nil && *auction.TitleFr != "" {
				title = *auction.TitleFr
			}
			if auction.TitleEn != nil && *auction.TitleEn != "" {
				title = *auction.TitleEn
			}
			params := map[string]string{
				"auctionTitle": title,
				"reason":       reason,
			}
			data := map[string]string{
				"type":         "report",
				"auction_id":   auctionID.String(),
				"reference_id": report.ID.String(),
			}
			_ = s.notifSvc.NotifyAdminsLocalized(context.Background(), "auction_reported", params, data)
		}()
	}

	return nil
}

func (s *auctionService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.auctionRepo.Delete(ctx, id)
}

// UpdateAuction - Modifier son enchère avec validation du propriétaire
func (s *auctionService) UpdateAuction(ctx context.Context, id uuid.UUID, userID uuid.UUID, input CreateAuctionInput) error {
	auction, err := s.auctionRepo.FindByID(ctx, id)
	if err != nil {
		return apperr.ErrNotFound
	}

	// Vérifier que l'utilisateur est le vendeur
	if auction.SellerID != userID {
		return apperr.ErrUnauthorized
	}

	// Ne permettre la modification que si l'enchère est en statut pending ou active
	if auction.Status != "pending" && auction.Status != "active" {
		return fmt.Errorf("can only update pending or active auctions")
	}

	// Vérifier que la catégorie existe
	if input.CategoryID > 0 {
		categories, err := s.auctionRepo.GetCategories(ctx)
		if err != nil {
			return fmt.Errorf("failed to fetch categories: %w", err)
		}
		categoryExists := false
		for _, c := range categories {
			if c.ID == input.CategoryID {
				categoryExists = true
				break
			}
		}
		if !categoryExists {
			return fmt.Errorf("category with ID %d does not exist", input.CategoryID)
		}
	}

	// Vérifier que la location existe si fournie
	if input.LocationID != nil && *input.LocationID > 0 {
		locations, err := s.auctionRepo.GetLocations(ctx)
		if err != nil {
			return fmt.Errorf("failed to fetch locations: %w", err)
		}
		locationExists := false
		for _, l := range locations {
			if l.ID == *input.LocationID {
				locationExists = true
				break
			}
		}
		if !locationExists {
			return fmt.Errorf("location with ID %d does not exist", *input.LocationID)
		}
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

	auction.CategoryID = input.CategoryID
	auction.SubCategoryID = input.SubCategoryID
	auction.LocationID = input.LocationID
	auction.TitleAr = input.TitleAr
	auction.TitleFr = tFr
	auction.TitleEn = tEn
	auction.DescriptionAr = dAr
	auction.DescriptionFr = dFr
	auction.DescriptionEn = dEn
	auction.StartPrice = input.StartPrice
	auction.MinIncrement = input.MinIncrement
	auction.InsuranceAmount = input.InsuranceAmount
	auction.EndTime = input.EndTime
	auction.PhoneContact = phone
	auction.BuyNowPrice = input.BuyNowPrice
	auction.ItemDetails = input.ItemDetails
	auction.Condition = input.Condition
	auction.Brand = input.Brand
	auction.VideoURL = input.VideoURL
	if input.StartTime != nil {
		auction.StartTime = *input.StartTime
	}

	// Image bug fix (client feedback, R2 deletion-path audit): AuctionHandler.
	// Update's own request DTO (PUT /auctions/:id, the mobile app's own
	// edit-auction call) never parses an "images" field from the body at
	// all, so `input.Images` was ALWAYS empty on every real call reaching
	// this function -- yet it used to unconditionally wipe existing
	// auction_images rows and delete the corresponding R2 objects on every
	// plain field edit (title, price, description, duration...). This is
	// the confirmed root cause of an uploaded image being deleted from R2
	// while its auction_images DB row could independently survive depending
	// on call ordering relative to the upload. Only touch images at all when
	// the caller actually supplied at least one non-empty URL -- image
	// management belongs exclusively to POST /auctions/:id/images (AddImages).
	// Same principle already used correctly by adminService.UpdateAuction's
	// hasNewImages guard.
	hasNewImages := imagesUpdateIncludesNewURLs(input.Images)

	var existingImages []models.AuctionImage
	if hasNewImages {
		// Get existing images before update (for R2 cleanup) -- only needed
		// when we're actually about to replace them.
		existingImages, err = s.auctionRepo.GetImages(ctx, id)
		if err != nil {
			return fmt.Errorf("failed to get existing images: %w", err)
		}
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	if err := s.auctionRepo.Update(ctx, auction); err != nil {
		return err
	}

	if hasNewImages {
		// Mettre à jour les images (vider et re-remplir) -- uniquement quand
		// l'appelant a explicitement fourni au moins une nouvelle image.
		if err := s.auctionRepo.DeleteImagesTx(ctx, tx, id); err != nil {
			return fmt.Errorf("failed to clear old images: %w", err)
		}

		for i, url := range input.Images {
			if url == "" {
				continue
			}
			if err := s.auctionRepo.AddImageTx(ctx, tx, &models.AuctionImage{
				AuctionID:    id,
				URL:          url,
				MediaType:    "image",
				DisplayOrder: i + 1,
			}); err != nil {
				return fmt.Errorf("failed to save new image URL to database: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// After successful commit, delete old images from R2 (best effort, don't fail if error)
	if hasNewImages && s.mediaSvc != nil {
		for _, img := range existingImages {
			if err := s.mediaSvc.DeleteFile(ctx, img.URL); err != nil {
				// Log but don't fail - the DB update succeeded
				fmt.Printf("[AuctionService UpdateAuction] Warning: failed to delete old image from R2: %s, error: %v\n", img.URL, err)
			}
		}
	}

	s.emitAuctionEvent(auction, models.EventAuctionUpdated)

	return nil
}

// DeleteAuction - Supprimer son enchère avec validation du propriétaire
func (s *auctionService) DeleteAuction(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	auction, err := s.auctionRepo.FindByID(ctx, id)
	if err != nil {
		return apperr.ErrNotFound
	}

	// Vérifier que l'utilisateur est le vendeur
	if auction.SellerID != userID {
		return apperr.ErrUnauthorized
	}

	// Ne permettre la suppression que si l'enchère est en statut pending
	if auction.Status != "pending" {
		return fmt.Errorf("can only delete pending auctions")
	}

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
				fmt.Printf("[AuctionService DeleteAuction] Warning: failed to delete image from R2: %s, error: %v\n", img.URL, err)
			}
		}
	}

	// A "pending" auction was never approved/visible on any mobile list yet
	// (see Create's comment above), but a client could still be viewing its
	// own pending auction (e.g. Auction Details, My Auctions) -- emit so
	// that view catches up rather than silently keeping a now-deleted
	// auction on screen.
	s.emitAuctionEvent(auction, models.EventAuctionDeleted)

	return nil
}

// GetBidStatus - Statut de ma bid pour une enchère
func (s *auctionService) GetBidStatus(ctx context.Context, auctionID, userID uuid.UUID) (map[string]interface{}, error) {
	auction, err := s.auctionRepo.FindByID(ctx, auctionID)
	if err != nil {
		return nil, apperr.ErrNotFound
	}

	// Récupérer la plus haute bid de l'utilisateur -- sql.ErrNoRows (the
	// user has never bid on this auction) is the normal, expected state for
	// most callers of this endpoint, not a failure: GetBidStatus's own
	// contract already returns my_bid_amount=nil/is_highest_bid=false for
	// exactly this case below. Only report a genuine error for anything
	// else -- pre-existing bug, found while wiring has_bid (client feedback
	// #19): every user who had never bid previously got a 404 here instead
	// of a normal has_bid=false/my_bid_amount=nil response.
	highestBid, err := s.auctionRepo.GetUserHighestBid(ctx, auctionID, userID)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// has_bid (Customer #38, replacing the old client feedback #19 permanent
	// "ever bid" meaning): whether this user is currently blocked from
	// bidding again, i.e. whether they are the CURRENT last bidder
	// (auctions.last_bidder_id) -- the exact same atomic source of truth
	// PlaceBid/TryClaimBidTurn uses to accept or reject the next bid, so
	// this field can never drift from the real enforcement. False again
	// once someone else outbids them, unlike the old permanent semantic.
	hasBid := auction.LastBidderID != nil && *auction.LastBidderID == userID

	status := map[string]interface{}{
		"auction_id":     auctionID,
		"is_highest_bid": false,
		"my_bid_amount":  nil,
		"current_price":  auction.CurrentPrice,
		"auction_status": auction.Status,
		"has_bid":        hasBid,
	}

	if highestBid != nil {
		status["my_bid_amount"] = highestBid.Amount
		// Vérifier si c'est la plus haute bid
		if auction.CurrentPrice.Equal(highestBid.Amount) {
			status["is_highest_bid"] = true
		}
	}

	return status, nil
}

func (s *auctionService) IncrementViews(ctx context.Context, id uuid.UUID, userID *uuid.UUID) error {
	return s.auctionRepo.IncrementViews(ctx, id, userID)
}

func (s *auctionService) AddImages(ctx context.Context, auctionID, sellerID uuid.UUID, urls []string) error {
	auction, err := s.auctionRepo.FindByID(ctx, auctionID)
	if err != nil {
		return apperr.ErrNotFound
	}
	if auction.SellerID != sellerID {
		return apperr.ErrUnauthorized
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	for i, url := range urls {
		img := &models.AuctionImage{
			AuctionID:    auctionID,
			URL:          url,
			MediaType:    "image",
			DisplayOrder: i + 1,
		}
		if err := s.auctionRepo.AddImageTx(ctx, tx, img); err != nil {
			return fmt.Errorf("failed to save uploaded image URL to database: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}

	s.emitAuctionEvent(auction, models.EventAuctionUpdated)
	return nil
}

// BuyNow (Bug J) validates preconditions, then performs the winner
// transition via TryBuyNowAtomically -- the same WHERE status='active' ...
// RETURNING id race guard already proven correct for Customer #23's
// expiration finalizer, closing the BuyNow-vs-expiration and concurrent
// double-BuyNow races. If another caller (a concurrent BuyNow, or the
// expiration scheduler) already transitioned the auction out of 'active'
// between the FindByID precondition check above and this atomic write, won
// is false and this call safely fails with ErrConflict rather than silently
// overwriting -- or previously, silently no-op'ing while still reporting
// success (the exact Bug J defect).
func (s *auctionService) BuyNow(ctx context.Context, auctionID, buyerID uuid.UUID) (*models.Auction, error) {
	auction, err := s.auctionRepo.FindByID(ctx, auctionID)
	if err != nil {
		return nil, apperr.ErrNotFound
	}
	if auction.Status != "active" {
		return nil, fmt.Errorf("auction is not active")
	}
	if auction.BuyNowPrice == nil {
		return nil, fmt.Errorf("auction does not have buy now price")
	}
	if auction.SellerID == buyerID {
		return nil, fmt.Errorf("cannot buy your own auction")
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	won, err := s.auctionRepo.TryBuyNowAtomically(ctx, tx, auctionID, buyerID, *auction.BuyNowPrice)
	if err != nil {
		return nil, err
	}
	if !won {
		return nil, apperr.ErrConflict
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	auction.WinnerID = &buyerID
	auction.Status = "ended"
	auction.CurrentPrice = *auction.BuyNowPrice
	return auction, nil
}

// CancelAuction (Bug J) validates preconditions, then performs the status
// transition via TryCancelAuctionAtomically -- the WHERE status NOT IN
// ('ended','canceled') guard matches the exact pre-existing business rule
// (both 'pending' and 'active' auctions were always cancellable) and closes
// the Cancel-vs-expiration race: if FinalizeExpiredAuction already
// transitioned the auction to 'ended' between the precondition check above
// and this atomic write, won is false and Cancel now correctly fails with
// ErrConflict instead of the prior defect (reporting success, emitting a
// realtime "canceled" event, and releasing holds -- all for a DB row that
// silently never changed).
func (s *auctionService) CancelAuction(ctx context.Context, auctionID, sellerID uuid.UUID, reason string) error {
	auction, err := s.auctionRepo.FindByID(ctx, auctionID)
	if err != nil {
		return apperr.ErrNotFound
	}
	if auction.SellerID != sellerID {
		return apperr.ErrUnauthorized
	}
	if auction.Status == "ended" || auction.Status == "canceled" {
		return fmt.Errorf("cannot cancel auction in current status: %s", auction.Status)
	}

	oldStatus := auction.Status

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	won, err := s.auctionRepo.TryCancelAuctionAtomically(ctx, tx, auctionID, reason)
	if err != nil {
		return err
	}
	if !won {
		return apperr.ErrConflict
	}
	if err := tx.Commit(); err != nil {
		return err
	}

	auction.Status = "canceled"
	auction.RejectionReason = &reason

	// Emitted right after the status write commits -- the wallet-hold-release
	// step below is a separate best-effort side effect with its own early
	// returns and must not gate this event.
	s.emitAuctionEvent(auction, models.EventAuctionStatusChanged)

	if s.auditSvc != nil {
		details := fmt.Sprintf("seller_id=%s old_status=%s new_status=canceled reason=%s", sellerID, oldStatus, reason)
		detailsJSON := models.JSONB{
			"auction_id": auctionID.String(),
			"seller_id":  sellerID.String(),
			"old_status": oldStatus,
			"new_status": "canceled",
		}
		if reason != "" {
			detailsJSON["reason"] = reason
		}
		// IP/User-Agent non disponibles ici : CancelAuction est une méthode de service
		// sans *fiber.Ctx — non étendu dans cette phase (voir rapport Phase B).
		if auditErr := s.auditSvc.Log(ctx, sellerID, "auction_cancelled", "auction", &auctionID, details,
			WithActorType("user"),
			WithDetailsJSON(detailsJSON),
		); auditErr != nil {
			if s.logger != nil {
				s.logger.Error("CancelAuction: failed to write audit log", zap.String("auction_id", auctionID.String()), zap.Error(auditErr))
			}
		}
	}

	// Libère la caution (insurance_amount) de tous les enchérisseurs — l'auction est
	// annulé, personne ne doit rester avec des fonds gelés (audit de sécurité V03/V09).
	dbtx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil // l'annulation de l'auction a réussi ; on log mais ne bloque pas dessus
	}
	defer dbtx.Rollback()
	if err := s.walletRepo.ReleaseHoldsForAuction(ctx, dbtx, auctionID); err != nil {
		return nil
	}
	_ = dbtx.Commit()

	if s.auditSvc != nil {
		detailsJSON := models.JSONB{
			"auction_id":          auctionID.String(),
			"hold_release_reason": "cancelled",
		}
		if auditErr := s.auditSvc.Log(ctx, sellerID, "auction_holds_released", "auction", &auctionID, "reason=cancelled",
			WithActorType("user"),
			WithDetailsJSON(detailsJSON),
		); auditErr != nil {
			if s.logger != nil {
				s.logger.Error("CancelAuction: failed to write audit log for holds release", zap.String("auction_id", auctionID.String()), zap.Error(auditErr))
			}
		}
	}
	return nil
}

// RelistAuction (Bug J) validates preconditions, then performs the status/
// winner-reset transition via TryRelistAuctionAtomically. Clears winner_id,
// winning_bid_id, AND payment_deadline together -- winning_bid_id and
// payment_deadline are themselves finalization state tied to the SAME prior
// winner_id (set together by SetWinner/TrySetWinnerAtomically/
// TryBuyNowAtomically), so leaving either behind after clearing winner_id
// would attach a stale old-run winner reference to a freshly relisted
// auction. Historical bid ROWS are never touched (relist has never deleted/
// reset bids; that is unchanged). Guard (status IN ('canceled','ended'))
// matches the exact pre-existing precondition and additionally closes any
// theoretical Relist-vs-concurrent-lifecycle-change race.
func (s *auctionService) RelistAuction(ctx context.Context, auctionID, sellerID uuid.UUID, newEndTime time.Time) error {
	auction, err := s.auctionRepo.FindByID(ctx, auctionID)
	if err != nil {
		return apperr.ErrNotFound
	}
	if auction.SellerID != sellerID {
		return apperr.ErrUnauthorized
	}
	if auction.Status != "canceled" && auction.Status != "ended" {
		return fmt.Errorf("can only relist canceled or ended auctions")
	}

	oldStatus := auction.Status

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	won, err := s.auctionRepo.TryRelistAuctionAtomically(ctx, tx, auctionID, newEndTime, auction.StartPrice)
	if err != nil {
		return err
	}
	if !won {
		return apperr.ErrConflict
	}
	if err := tx.Commit(); err != nil {
		return err
	}

	auction.Status = "pending"
	auction.EndTime = newEndTime
	auction.WinnerID = nil
	auction.WinningBidID = nil
	auction.PaymentDeadline = nil
	auction.CurrentPrice = auction.StartPrice

	if s.auditSvc != nil {
		details := fmt.Sprintf("old_status=%s new_status=pending new_end_time=%s", oldStatus, newEndTime.Format(time.RFC3339))
		detailsJSON := models.JSONB{
			"auction_id":   auctionID.String(),
			"seller_id":    sellerID.String(),
			"old_status":   oldStatus,
			"new_status":   "pending",
			"new_end_time": newEndTime.Format(time.RFC3339),
		}
		if auditErr := s.auditSvc.Log(ctx, sellerID, "auction_relisted", "auction", &auctionID, details,
			WithActorType("user"),
			WithDetailsJSON(detailsJSON),
		); auditErr != nil {
			if s.logger != nil {
				s.logger.Error("RelistAuction: failed to write audit log", zap.String("auction_id", auctionID.String()), zap.Error(auditErr))
			}
		}
	}
	s.emitAuctionEvent(auction, models.EventAuctionStatusChanged)
	return nil
}

func (s *auctionService) ExtendAuction(ctx context.Context, auctionID, sellerID uuid.UUID, hours int) error {
	auction, err := s.auctionRepo.FindByID(ctx, auctionID)
	if err != nil {
		return apperr.ErrNotFound
	}
	if auction.SellerID != sellerID {
		return apperr.ErrUnauthorized
	}
	if auction.Status != "active" {
		return fmt.Errorf("can only extend active auctions")
	}

	oldEndTime := auction.EndTime
	newEndTime := auction.EndTime.Add(time.Duration(hours) * time.Hour)
	auction.EndTime = newEndTime
	if err := s.auctionRepo.Update(ctx, auction); err != nil {
		return err
	}

	if s.auditSvc != nil {
		details := fmt.Sprintf("old_end_time=%s new_end_time=%s", oldEndTime.Format(time.RFC3339), newEndTime.Format(time.RFC3339))
		detailsJSON := models.JSONB{
			"auction_id":    auctionID.String(),
			"seller_id":     sellerID.String(),
			"old_end_time":  oldEndTime.Format(time.RFC3339),
			"new_end_time":  newEndTime.Format(time.RFC3339),
		}
		if auditErr := s.auditSvc.Log(ctx, sellerID, "auction_extended", "auction", &auctionID, details,
			WithActorType("user"),
			WithDetailsJSON(detailsJSON),
		); auditErr != nil {
			if s.logger != nil {
				s.logger.Error("ExtendAuction: failed to write audit log", zap.String("auction_id", auctionID.String()), zap.Error(auditErr))
			}
		}
	}
	s.emitAuctionEvent(auction, models.EventAuctionUpdated)
	return nil
}

// CloseExpiredAuctions — appelé par le Cron toutes les 30 secondes.
// Customer #23: this used to contain its own independent winner-selection/
// persistence logic (never calling SetWinner), which raced against
// AuctionScheduler.checkEndedAuctions's own separate implementation --
// whichever loop's UpdateStatus("ended") landed first silently starved the
// other, since both queried on status='active'. Now a pure discovery loop:
// find candidate auction IDs, delegate the entire finalize decision (winner
// selection, persistence, notifications, realtime event) to the single
// canonical FinalizeExpiredAuction. Every auction this finds is finalized
// exactly the same way regardless of which background loop discovered it
// first, and a candidate already finalized by the other loop safely no-ops
// here (FinalizeExpiredAuction re-checks status='active' itself).
func (s *auctionService) CloseExpiredAuctions(ctx context.Context) error {
	auctions, err := s.auctionRepo.FindExpiredActive(ctx)
	if err != nil {
		return err
	}
	for _, a := range auctions {
		if err := s.FinalizeExpiredAuction(ctx, a.ID); err != nil {
			if s.logger != nil {
				s.logger.Error("CloseExpiredAuctions: FinalizeExpiredAuction failed", zap.String("auction_id", a.ID.String()), zap.Error(err))
			}
		}
	}
	return nil
}

// FinalizeExpiredAuction (Customer #23 canonical finalizer) is the single
// authoritative active -> ended transition for one auction. Safe to call
// concurrently for the same auctionID from multiple goroutines/processes:
// exactly one call actually performs the transition (see
// auctionRepo.TrySetWinnerAtomically/TryEndAuctionAtomically for the DB-level
// race guard); every other concurrent call observes won=false and returns
// nil immediately, sending no notifications and emitting no realtime event
// -- so "DB state is the primary idempotency guard", not Redis.
func (s *auctionService) FinalizeExpiredAuction(ctx context.Context, auctionID uuid.UUID) error {
	a, _, err := s.GetByID(ctx, auctionID)
	if err != nil {
		return err
	}
	if a == nil {
		return nil
	}
	// Only ever finalize an auction that is actually still active -- a
	// cheap pre-check before opening a transaction; the real race guard is
	// still the atomic WHERE status='active' inside the transaction below,
	// this is purely an optimization to skip the transaction+bid lookup
	// entirely for an auction that's obviously already settled.
	if a.Status != "active" {
		return nil
	}

	topBid, bidErr := s.bidRepo.FindTopBid(ctx, auctionID)
	hasWinner := bidErr == nil && topBid != nil && a.CurrentPrice.GreaterThan(a.StartPrice) && a.BidderCount > 0

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var won bool
	var winnerID uuid.UUID
	var winningBidID uuid.UUID
	if hasWinner {
		winnerID = topBid.UserID
		winningBidID = topBid.ID
		won, err = s.auctionRepo.TrySetWinnerAtomically(ctx, tx, auctionID, winnerID, winningBidID)
	} else {
		won, err = s.auctionRepo.TryEndAuctionAtomically(ctx, tx, auctionID)
	}
	if err != nil {
		return err
	}
	if !won {
		// Another concurrent caller (the other background loop, or a
		// duplicate discovery of the same candidate) already finalized this
		// auction first -- safe no-op, no notifications, no event.
		return nil
	}

	// Libère la caution des enchérisseurs non-gagnants (audit de sécurité V03/V09),
	// dans la MEME transaction que la transition de statut -- si l'une échoue,
	// les deux sont annulées ensemble plutôt que de laisser un état partiel
	// (auction 'ended' mais holds non libérés).
	var winnerIDPtr *uuid.UUID
	if hasWinner {
		winnerIDPtr = &winnerID
	}
	if err := s.walletRepo.ReleaseHoldsForNonWinners(ctx, tx, auctionID, winnerIDPtr); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	// Everything below only runs for the single caller that actually
	// performed the transition (won == true, transaction committed) --
	// notifications/realtime events are sent exactly once per real
	// auction-close, never per background-loop tick.
	auctionCopy := *a
	auctionCopy.Status = "ended"
	s.emitAuctionEvent(&auctionCopy, models.EventAuctionStatusChanged)

	if s.auditSvc != nil {
		detailsJSON := models.JSONB{
			"auction_id":          auctionID.String(),
			"hold_release_reason": "expired_non_winners",
		}
		if auditErr := s.auditSvc.Log(ctx, uuid.Nil, "auction_holds_released", "auction", &auctionID, "reason=expired_non_winners",
			WithActorType("system"),
			WithSystemActor(),
			WithDetailsJSON(detailsJSON),
		); auditErr != nil {
			if s.logger != nil {
				s.logger.Error("FinalizeExpiredAuction: failed to write audit log", zap.String("auction_id", auctionID.String()), zap.Error(auditErr))
			}
		}
	}

	seller, sellerErr := s.userRepo.FindByID(ctx, a.SellerID)
	if sellerErr == nil && seller != nil {
		language := "ar"
		if seller.LanguagePref != "" {
			language = seller.LanguagePref
		}
		params := map[string]string{
			"auctionTitle": a.TitleAr,
			"finalPrice":   a.CurrentPrice.String(),
			"currency":     a.EffectiveCurrencyCode(),
		}
		data := map[string]string{
			"type":       "auction_ended",
			"auctionId":  auctionID.String(),
			"finalPrice": a.CurrentPrice.String(),
		}
		_ = s.notifSvc.SendLocalizedPush(ctx, a.SellerID, "auction_ended", language, params, data)
	}

	if hasWinner {
		winner, winnerErr := s.userRepo.FindByID(ctx, winnerID)
		if winnerErr == nil && winner != nil {
			winnerLang := "ar"
			if winner.LanguagePref != "" {
				winnerLang = winner.LanguagePref
			}
			winnerParams := map[string]string{
				"auctionTitle": a.TitleAr,
				"finalPrice":   a.CurrentPrice.String(),
				"currency":     a.EffectiveCurrencyCode(),
			}
			winnerData := map[string]string{
				"type":       "auction_won",
				"auctionId":  auctionID.String(),
				"finalPrice": a.CurrentPrice.String(),
			}
			_ = s.notifSvc.SendLocalizedPush(ctx, winnerID, "auction_won", winnerLang, winnerParams, winnerData)
		}
	}

	return nil
}

func (s *auctionService) GetCategories(ctx context.Context) ([]models.Category, error) {
	cacheKey := "categories"

	// Try to get from cache
	if s.rdb != nil {
		val, err := s.rdb.Get(ctx, cacheKey).Result()
		if err == nil {
			var categories []models.Category
			if err := json.Unmarshal([]byte(val), &categories); err == nil {
				return categories, nil
			}
		}
	}

	// Get from DB
	categories, err := s.auctionRepo.GetCategories(ctx)
	if err != nil {
		return nil, err
	}

	// Save to cache (1 hour)
	if s.rdb != nil && len(categories) > 0 {
		data, _ := json.Marshal(categories)
		_ = s.rdb.Set(ctx, cacheKey, data, 1*time.Hour).Err()
	}

	return categories, nil
}

func (s *auctionService) GetLocations(ctx context.Context) ([]models.Location, error) {
	cacheKey := "locations"

	// Try to get from cache
	if s.rdb != nil {
		val, err := s.rdb.Get(ctx, cacheKey).Result()
		if err == nil {
			var locations []models.Location
			if err := json.Unmarshal([]byte(val), &locations); err == nil {
				return locations, nil
			}
		}
	}

	// Get from DB
	locations, err := s.auctionRepo.GetLocations(ctx)
	if err != nil {
		return nil, err
	}

	// Save to cache (1 hour)
	if s.rdb != nil && len(locations) > 0 {
		data, _ := json.Marshal(locations)
		_ = s.rdb.Set(ctx, cacheKey, data, 1*time.Hour).Err()
	}

	return locations, nil
}

func (s *auctionService) GetCountries(ctx context.Context) ([]models.Country, error) {
	return s.auctionRepo.GetCountries(ctx)
}

func (s *auctionService) GetLocationsByCountry(ctx context.Context, countryID int) ([]models.Location, error) {
	return s.auctionRepo.GetLocationsByCountry(ctx, countryID)
}

func (s *auctionService) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	return s.userRepo.FindByID(ctx, userID)
}
