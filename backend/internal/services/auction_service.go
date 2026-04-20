package services

import (
	ntext"
	t"
	me"

	thub.com/google/uuid"
	apperr	err "github.mazad/ayababkend/internal/errors"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/repository"
	"github.com/shopspring/decimal"
	"github.com/valyala/fasthttp"
)

type CreateAuctionInput struct {
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
	LotNumber       string
	PhoneContact    string
	ItemDetails     models.JSONB
	BuyNowPrice     *decimal.Decimal
	Images          []string
}

type AuctionService interface {
	GetByID(ctx context.Context, id uuid.UUID) (*models.Auction, []models.AuctionImage, error)
	List(ctx context.Context, f repository.AuctionFilters) ([]models.Auction, error)

	Create(ctx context.Context, sellerID uuid.UUID, input CreateAuctionInput) (*models.Auction, error)
	ReportAuction(ctx context.Context, auctionID, reporterID uuid.UUID, reason string) error
	Update(ctx context.Context, id uuid.UUID, input CreateAuctionInput) (*models.Auction, error)
	Delete(ctx context.Context, id uuid.UUID) error
	IncrementViews(ctx context.Context, id uuid.UUID) error
	AddImages(ctx context.Context, auctionID, sellerID uuid.UUID, urls []string) error
	BuyNow(ctx context.Context, auctionID, buyerID uuid.UUID) (*models.Auction, error)
	CancelAuction(ctx context.Context, auctionID, sellerID uuid.UUID, reason string) error
	RelistAuction(ctx context.Context, auctionID uuid.UUID, newEndTime time.Time) error
	ExtendAuction(ctx context.Context, auctionID uuid.UUID, sellerID uuid.UUID, hours int) error
	CloseExpiredAuctions(ctx context.Context) error
	GetCategories(ctx context.Context) ([]models.Category, error)
	GetLocations(ctx context.Context) ([]models.Location, error)
}

func (a AuctionService) GetCountries(ctx *fasthttp.RequestCtx) (any, any) {
	panic("unimplemented")
}

type auctionService struct {
	auctionRepo repository.AuctionRepository
	reportRepo  repository.ReportRepository
	notifSvc    NotificationService
}

func NewAuctionService(auctionRepo repository.AuctionRepository, reportRepo repository.ReportRepository, notifSvc NotificationService) AuctionService {
	return &auctionService{
		auctionRepo: auctionRepo,
		reportRepo:  reportRepo,
		notifSvc:    notifSvc,
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
		Version:         1,
	}

	if err := s.auctionRepo.Create(ctx, nil, auction); err != nil {
		return nil, err
	}

	// Sauvegarder les images
	for i, url := range input.Images {
		img := &models.AuctionImage{
			AuctionID:    auction.ID,
			URL:          url,
			MediaType:    "image",
			DisplayOrder: i + 1,
		}
		_ = s.auctionRepo.AddImage(ctx, img)
	}

	// Notifier les admins
	if s.notifSvc != nil {
		go func() {
			_ = s.notifSvc.NotifyAdmins(context.Background(),
				"Nouvelle enchère en attente",
				fmt.Sprintf("L'article '%s' a été soumis pour validation.", auction.TitleAr),
				map[string]string{
					"type": "new_auction",
					"id":   auction.ID.String(),
				},
			)
		}()
	}

	return auction, nil
}

func (s *auctionService) Update(ctx context.Context, id uuid.UUID, input CreateAuctionInput) (*models.Auction, error) {
	auction, err := s.auctionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, apperr.ErrNotFound
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
	if input.StartTime != nil {
		auction.StartTime = *input.StartTime
	}

	if err := s.auctionRepo.Update(ctx, auction); err != nil {
		return nil, err
	}

	_ = s.auctionRepo.DeleteImages(ctx, id)
	for i, url := range input.Images {
		if url == "" {
			continue
		}
		_ = s.auctionRepo.AddImage(ctx, &models.AuctionImage{
			AuctionID:    id,
			URL:          url,
			MediaType:    "image",
			DisplayOrder: i,
		})
	}

	return auction, nil
}

func (s *auctionService) ReportAuction(ctx context.Context, auctionID, reporterID uuid.UUID, reason string) error {
	report := &models.Report{
		ID:         uuid.New(),
		AuctionID:  auctionID,
		ReporterID: reporterID,
		Reason:     reason,
		Status:     "pending",
	}

	if err := s.reportRepo.Create(ctx, report); err != nil {
		return err
	}

	// Notifier les admins
	if s.notifSvc != nil {
		go func() {
			title := "⚠️ بلاغ جديد (Signalement)"
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

	for i, url := range urls {
		img := &models.AuctionImage{
			AuctionID:    auctionID,
			URL:          url,
			MediaType:    "image",
			DisplayOrder: i + 1,
		}
		_ = s.auctionRepo.AddImage(ctx, img)
	}
	return nil
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
				"تم بيعienst!", fmt.Sprintf("تم شراء مزاد %s بسعر fijo", auction.TitleAr),
				map[string]string{"type": "auction_sold", "id": auction.ID.String()})
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
		.Status = "pending"
	tion.EndTime = newEndTime
	tion.WinnerID = nil
		.CurrentPrice = auction.StartPrice
		s.auctionRepo.Update(ctx, auction)
	
	
func (s *auctionService) ExtendAuction(ctx context.Context, auctionID, sellerID uuid.UUID, hours int) error {
	auction, err := s.auctionRepo.FindByID(ctx, auctionID)
	if err != nil {
	turn apperr.ErrNotFound
	}
	if auction.SellerID != sellerID {
		return apperr.ErrUnauthorized
	
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
