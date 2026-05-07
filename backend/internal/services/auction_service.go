package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	apperr "github.com/mazadpay/backend/internal/errors"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/repository"
	"github.com/shopspring/decimal"
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
	List(ctx context.Context, f repository.AuctionFilters) ([]models.Auction, error)

	Create(ctx context.Context, sellerID uuid.UUID, input CreateAuctionInput) (*models.Auction, error)
	ReportAuction(ctx context.Context, auctionID, reporterID uuid.UUID, reason string) error
	Update(ctx context.Context, id uuid.UUID, input CreateAuctionInput) (*models.Auction, error)
	UpdateAuction(ctx context.Context, id uuid.UUID, userID uuid.UUID, input CreateAuctionInput) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteAuction(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	GetBidStatus(ctx context.Context, auctionID, userID uuid.UUID) (map[string]interface{}, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error)
	IncrementViews(ctx context.Context, id uuid.UUID) error
	AddImages(ctx context.Context, auctionID, sellerID uuid.UUID, urls []string) error
	BuyNow(ctx context.Context, auctionID, buyerID uuid.UUID) (*models.Auction, error)
	CancelAuction(ctx context.Context, auctionID, sellerID uuid.UUID, reason string) error
	RelistAuction(ctx context.Context, auctionID uuid.UUID, newEndTime time.Time) error
	ExtendAuction(ctx context.Context, auctionID uuid.UUID, sellerID uuid.UUID, hours int) error
	CloseExpiredAuctions(ctx context.Context) error
	GetCategories(ctx context.Context) ([]models.Category, error)
	GetLocations(ctx context.Context) ([]models.Location, error)
	GetCountries(ctx context.Context) ([]models.Country, error)
	GetLocationsByCountry(ctx context.Context, countryID int) ([]models.Location, error)
}

type auctionService struct {
	db          *sqlx.DB
	auctionRepo repository.AuctionRepository
	reportRepo  repository.ReportRepository
	notifSvc    NotificationService
	userRepo    repository.UserRepository
	mediaSvc    MediaService
}

func NewAuctionService(db *sqlx.DB, auctionRepo repository.AuctionRepository, reportRepo repository.ReportRepository, notifSvc NotificationService, userRepo repository.UserRepository, mediaSvc MediaService) AuctionService {
	return &auctionService{
		db:          db,
		auctionRepo: auctionRepo,
		reportRepo:  reportRepo,
		notifSvc:    notifSvc,
		userRepo:    userRepo,
		mediaSvc:    mediaSvc,
	}
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

	// Notifier les admins via FCM avec support multi-langues
	if s.notifSvc != nil {
		go func() {
			// Récupérer le nom du vendeur
			seller, err := s.userRepo.FindByID(context.Background(), sellerID)
			userName := "Un utilisateur"
			if err == nil && seller != nil && seller.FullName != nil {
				userName = *seller.FullName
			}

			params := map[string]string{
				"userName":     userName,
				"auctionTitle": auction.TitleAr,
			}
			data := map[string]string{
				"type":      "auction_pending",
				"auctionId": auction.ID.String(),
				"sellerId":  sellerID.String(),
			}
			_ = s.notifSvc.NotifyAdminsLocalized(context.Background(), "new_auction", params, data)
		}()
	}

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

	// Get existing images before update (for R2 cleanup)
	existingImages, err := s.auctionRepo.GetImages(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing images: %w", err)
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	if err := s.auctionRepo.Update(ctx, auction); err != nil {
		return nil, err
	}

	// Synchroniser les images (toujours vider et re-remplir pour garantir la cohérence)
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

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// After successful commit, delete old images from R2 (best effort, don't fail if error)
	if s.mediaSvc != nil {
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

	// Notifier les admins
	if s.notifSvc != nil {
		go func() {
			title := " بلاغ جديد (Signalement)"
			body := fmt.Sprintf("تم الإبلاغ عن المزاد رقم %s. السبب: %s", auctionID.String()[:8], reason)
			_ = s.notifSvc.NotifyAdmins(context.Background(), title, body, map[string]string{
				"type":         "report",
				"auction_id":   auctionID.String(),
				"reference_id": report.ID.String(),
			})
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

	// Get existing images before update (for R2 cleanup)
	existingImages, err := s.auctionRepo.GetImages(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get existing images: %w", err)
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	if err := s.auctionRepo.Update(ctx, auction); err != nil {
		return err
	}

	// Mettre à jour les images
	// On vide et on re-remplit systématiquement pour refléter exactement l'état du formulaire
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

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// After successful commit, delete old images from R2 (best effort, don't fail if error)
	if s.mediaSvc != nil {
		for _, img := range existingImages {
			if err := s.mediaSvc.DeleteFile(ctx, img.URL); err != nil {
				// Log but don't fail - the DB update succeeded
				fmt.Printf("[AuctionService UpdateAuction] Warning: failed to delete old image from R2: %s, error: %v\n", img.URL, err)
			}
		}
	}

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

	return nil
}

// GetBidStatus - Statut de ma bid pour une enchère
func (s *auctionService) GetBidStatus(ctx context.Context, auctionID, userID uuid.UUID) (map[string]interface{}, error) {
	auction, err := s.auctionRepo.FindByID(ctx, auctionID)
	if err != nil {
		return nil, apperr.ErrNotFound
	}

	// Récupérer la plus haute bid de l'utilisateur
	highestBid, err := s.auctionRepo.GetUserHighestBid(ctx, auctionID, userID)
	if err != nil {
		return nil, err
	}

	status := map[string]interface{}{
		"auction_id":     auctionID,
		"is_highest_bid": false,
		"my_bid_amount":  nil,
		"current_price":  auction.CurrentPrice,
		"auction_status": auction.Status,
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

func (s *auctionService) IncrementViews(ctx context.Context, id uuid.UUID) error {
	return s.auctionRepo.IncrementViews(ctx, id)
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
	return tx.Commit()
}

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

	// Update auction as ended with winner
	auction.WinnerID = &buyerID
	auction.Status = "ended"
	auction.CurrentPrice = *auction.BuyNowPrice

	if err := s.auctionRepo.Update(ctx, auction); err != nil {
		return nil, err
	}

	// Notify seller
	if s.notifSvc != nil {
		go func() {
			_ = s.notifSvc.SendPush(context.Background(), auction.SellerID,
				"تم بيع!", fmt.Sprintf("تم شراء مزاد %s بسعر نهائي", auction.TitleAr),
				"auction_sold", map[string]string{"type": "auction_sold", "id": auction.ID.String()})
		}()
	}

	return auction, nil
}

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

	auction.Status = "canceled"
	auction.RejectionReason = &reason
	return s.auctionRepo.Update(ctx, auction)
}

func (s *auctionService) RelistAuction(ctx context.Context, auctionID uuid.UUID, newEndTime time.Time) error {
	auction, err := s.auctionRepo.FindByID(ctx, auctionID)
	if err != nil {
		return apperr.ErrNotFound
	}
	if auction.Status != "canceled" && auction.Status != "ended" {
		return fmt.Errorf("can only relist canceled or ended auctions")
	}

	auction.Status = "pending"
	auction.EndTime = newEndTime
	auction.WinnerID = nil
	auction.CurrentPrice = auction.StartPrice
	return s.auctionRepo.Update(ctx, auction)
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

	newEndTime := auction.EndTime.Add(time.Duration(hours) * time.Hour)
	auction.EndTime = newEndTime
	return s.auctionRepo.Update(ctx, auction)
}

// CloseExpiredAuctions — appelé par le Cron toutes les 30 secondes
func (s *auctionService) CloseExpiredAuctions(ctx context.Context) error {
	auctions, err := s.auctionRepo.FindExpiredActive(ctx)
	if err != nil {
		return err
	}
	for _, a := range auctions {
		_ = s.auctionRepo.UpdateStatus(ctx, a.ID, "ended")

		// Notifier le vendeur que l'enchère est terminée
		seller, err := s.userRepo.FindByID(ctx, a.SellerID)
		if err == nil && seller != nil {
			language := "ar"
			if seller.LanguagePref != "" {
				language = seller.LanguagePref
			}

			params := map[string]string{
				"auctionTitle": a.TitleAr,
				"finalPrice":   a.CurrentPrice.String(),
			}
			data := map[string]string{
				"type":       "auction_ended",
				"auctionId":  a.ID.String(),
				"finalPrice": a.CurrentPrice.String(),
			}

			// Notifier le vendeur
			_ = s.notifSvc.SendLocalizedPush(ctx, a.SellerID, "auction_ended", language, params, data)

			// Notifier le gagnant s'il y a des enchères
			if a.CurrentPrice.GreaterThan(a.StartPrice) && a.BidderCount > 0 {
				winnerID, err := s.auctionRepo.GetHighestBidder(ctx, a.ID)
				if err == nil && winnerID != uuid.Nil {
					winner, err := s.userRepo.FindByID(ctx, winnerID)
					if err == nil && winner != nil {
						winnerLang := "ar"
						if winner.LanguagePref != "" {
							winnerLang = winner.LanguagePref
						}

						winnerParams := map[string]string{
							"auctionTitle": a.TitleAr,
							"finalPrice":   a.CurrentPrice.String(),
						}
						winnerData := map[string]string{
							"type":       "auction_won",
							"auctionId":  a.ID.String(),
							"finalPrice": a.CurrentPrice.String(),
						}

						// Envoyer notification au gagnant
						_ = s.notifSvc.SendLocalizedPush(ctx, winnerID, "auction_won", winnerLang, winnerParams, winnerData)

						// Diffuser via WebSocket aux admins
						sellerName := "Vendeur"
						if seller.FullName != nil {
							sellerName = *seller.FullName
						}
						s.notifSvc.NotifyNewAuctionRequest(
							a.ID.String(),
							a.SellerID.String(),
							sellerName,
							a.TitleAr,
						)
					}
				}
			}
		}
	}
	return nil
}

func (s *auctionService) GetCategories(ctx context.Context) ([]models.Category, error) {
	return s.auctionRepo.GetCategories(ctx)
}

func (s *auctionService) GetLocations(ctx context.Context) ([]models.Location, error) {
	return s.auctionRepo.GetLocations(ctx)
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
