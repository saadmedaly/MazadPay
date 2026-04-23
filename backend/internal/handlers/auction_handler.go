package handlers

import (
	"encoding/json"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/middleware"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/repository"
	"github.com/mazadpay/backend/internal/services"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type AuctionHandler struct {
	service  services.AuctionService
	logger   *zap.Logger
	validate *validator.Validate
}

func NewAuctionHandler(svc services.AuctionService, logger *zap.Logger) *AuctionHandler {
	return &AuctionHandler{
		service:  svc,
		logger:   logger,
		validate: validator.New(),
	}
}

func (h *AuctionHandler) List(c *fiber.Ctx) error {
	f := repository.AuctionFilters{
		Status: c.Query("status", "active"),
		Query:  c.Query("q"),
	}
	if catID := c.QueryInt("category_id", 0); catID > 0 {
		f.CategoryID = catID
	}

	auctions, err := h.service.List(c.Context(), f)
	if err != nil {
		return MapError(c, h.logger, err)
	}

	return OK(c, auctions)
}

func (h *AuctionHandler) GetByID(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid auction ID")
	}

	auction, images, err := h.service.GetByID(c.Context(), id)
	if err != nil {
		return NotFound(c, "Auction")
	}

	return OK(c, fiber.Map{
		"auction": auction,
		"images":  images,
	})
}

type CreateAuctionRequest struct {
	// Accept category as string (mobile sends string) - will be mapped to ID
	// OR category_id as int (web admin sends int directly)
	Category        string                 `json:"category"`        // Optional: category name for mobile
	SubCategory     string                 `json:"sub_category"`    // Optional: subcategory name for mobile
	CategoryID      int                    `json:"category_id"`     // Required: category ID (used by web)
	SubCategoryID   *int                   `json:"sub_category_id"` // Optional: subcategory ID
	LocationID      *int                   `json:"location_id"`
	TitleAr         string                 `json:"title_ar"         validate:"required,min=3,max=200"`
	TitleFr         string                 `json:"title_fr"`
	TitleEn         string                 `json:"title_en"`
	DescriptionAr   string                 `json:"description_ar"`
	DescriptionFr   string                 `json:"description_fr"`
	DescriptionEn   string                 `json:"description_en"`
	StartPrice      float64                `json:"start_price"      validate:"required,gt=0"`
	MinIncrement    float64                `json:"min_increment"`
	InsuranceAmount float64                `json:"insurance_amount"`
	StartTime       string                 `json:"start_time"`
	EndTime         string                 `json:"end_time"         validate:"required"`
	BuyNowPrice     *float64               `json:"buy_now_price"`
	PhoneContact    string                 `json:"phone_contact"`
	ItemDetails     map[string]interface{} `json:"item_details"`
	Images          []string               `json:"images"`
	// New fields from migration 000024
	Condition       *string                `json:"condition" validate:"omitempty,oneof=new used refurbished damaged"`
	Brand           *string                `json:"brand"`
	VideoURL        *string                `json:"video_url"`
	// New field from migration 000032
	Quantity        int                    `json:"quantity" validate:"omitempty,min=1"` // Nombre d'items (défaut: 1)
}

func (h *AuctionHandler) Create(c *fiber.Ctx) error {
	// --- Log raw body for debugging ---
	rawBody := c.Body()
	h.logger.Info("[Create Auction] Raw body received", zap.String("body", string(rawBody)))

	var req CreateAuctionRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("[Create Auction] Body parse failed", zap.Error(err))
		return BadRequest(c, "Invalid request body: "+err.Error())
	}

	// Log parsed fields
	h.logger.Info("[Create Auction] Parsed request",
		zap.String("title_ar", req.TitleAr),
		zap.Int("category_id", req.CategoryID),
		zap.Float64("start_price", req.StartPrice),
		zap.String("end_time", req.EndTime),
		zap.String("start_time", req.StartTime),
	)

	if err := h.validate.Struct(req); err != nil {
		h.logger.Error("[Create Auction] Validation failed", zap.Error(err))
		return BadRequest(c, err.Error())
	}

	// Validate that either category (string) or category_id (int) is provided
	if req.Category == "" && req.CategoryID == 0 {
		h.logger.Error("[Create Auction] Missing category - neither category name nor category_id provided")
		return BadRequest(c, "Category is required (provide either 'category' name or 'category_id')")
	}

	// Map category string to ID if string is provided
	categoryID := req.CategoryID
	if req.Category != "" && req.CategoryID == 0 {
		// Try to find category by name (simple implementation - could be improved)
		// For now, we'll need to get all categories and find by name
		categories, err := h.service.GetCategories(c.Context())
		if err != nil {
			h.logger.Error("[Create Auction] Failed to get categories for mapping", zap.Error(err))
			return BadRequest(c, "Failed to map category")
		}

		for _, cat := range categories {
			if cat.NameAr == req.Category || cat.NameFr == req.Category || cat.NameEn == req.Category {
				categoryID = cat.ID
				break
			}
		}

		if categoryID == 0 {
			return BadRequest(c, "Category not found: "+req.Category)
		}
	}

	// Map sub_category string to ID if string is provided
	var subCategoryID *int
	if req.SubCategory != "" && req.SubCategoryID == nil {
		categories, err := h.service.GetCategories(c.Context())
		if err != nil {
			h.logger.Error("[Create Auction] Failed to get categories for subcategory mapping", zap.Error(err))
			return BadRequest(c, "Failed to map subcategory")
		}

		for _, cat := range categories {
			if cat.NameAr == req.SubCategory || cat.NameFr == req.SubCategory || cat.NameEn == req.SubCategory {
				subCategoryID = &cat.ID
				break
			}
		}
	} else if req.SubCategoryID != nil {
		subCategoryID = req.SubCategoryID
	}

	// parse end_time
	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		// try without nano
		endTime, err = time.Parse("2006-01-02T15:04:05Z", req.EndTime)
		if err != nil {
			h.logger.Error("[Create Auction] end_time parse failed", zap.String("end_time", req.EndTime), zap.Error(err))
			return BadRequest(c, "Invalid end_time format (use RFC3339): "+req.EndTime)
		}
	}

	// parse start_time (optional)
	var startTimePtr *time.Time
	if req.StartTime != "" {
		st, err := time.Parse(time.RFC3339, req.StartTime)
		if err != nil {
			st, err = time.Parse("2006-01-02T15:04:05Z", req.StartTime)
			if err != nil {
				h.logger.Error("[Create Auction] start_time parse failed", zap.String("start_time", req.StartTime), zap.Error(err))
				return BadRequest(c, "Invalid start_time format (use RFC3339): "+req.StartTime)
			}
		}
		startTimePtr = &st
	}

	// auto-compute min_increment if zero
	minIncrement := req.MinIncrement
	if minIncrement <= 0 {
		minIncrement = req.StartPrice * 0.05
		if minIncrement < 100 {
			minIncrement = 100
		}
		h.logger.Info("[Create Auction] min_increment auto-computed", zap.Float64("value", minIncrement))
	}

	sellerID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c)
	}

	var buyNow *decimal.Decimal
	if req.BuyNowPrice != nil && *req.BuyNowPrice > 0 {
		b := decimal.NewFromFloat(*req.BuyNowPrice)
		buyNow = &b
	}

	// Set default quantity if not provided
	quantity := req.Quantity
	if quantity == 0 {
		quantity = 1
	}

	input := services.CreateAuctionInput{
		CategoryID:      categoryID,
		SubCategoryID:   subCategoryID,
		LocationID:      req.LocationID,
		TitleAr:         req.TitleAr,
		TitleFr:         req.TitleFr,
		TitleEn:         req.TitleEn,
		DescriptionAr:   req.DescriptionAr,
		DescriptionFr:   req.DescriptionFr,
		DescriptionEn:   req.DescriptionEn,
		StartPrice:      decimal.NewFromFloat(req.StartPrice),
		MinIncrement:    decimal.NewFromFloat(minIncrement),
		InsuranceAmount: decimal.NewFromFloat(req.InsuranceAmount),
		StartTime:       startTimePtr,
		EndTime:         endTime,
		PhoneContact:    req.PhoneContact,
		ItemDetails:     req.ItemDetails,
		BuyNowPrice:     buyNow,
		Images:          req.Images,
		Condition:       req.Condition,
		Brand:           req.Brand,
		VideoURL:        req.VideoURL,
		Quantity:        quantity,
	}

	h.logger.Info("[Create Auction] Calling service.Create",
		zap.String("seller_id", sellerID.String()),
		zap.String("end_time", endTime.String()),
	)

	auction, err := h.service.Create(c.Context(), sellerID, input)
	if err != nil {
		h.logger.Error("[Create Auction] service.Create failed", zap.Error(err))
		return MapError(c, h.logger, err)
	}

	// Log the created auction body to be sent back
	aucionJSON, _ := json.Marshal(auction)
	h.logger.Info("[Create Auction] SUCCESS", zap.String("auction", string(aucionJSON)))

	return Created(c, auction)
}

func (h *AuctionHandler) IncrementView(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid auction ID")
	}
	_ = h.service.IncrementViews(c.Context(), id)
	return OK(c, fiber.Map{"message": "View counted"})
}

func (h *AuctionHandler) GetCategories(c *fiber.Ctx) error {
	categories, err := h.service.GetCategories(c.Context())
	if err != nil {
		return MapError(c, h.logger, err)
	}

	return OK(c, categories)
}

func (h *AuctionHandler) GetLocations(c *fiber.Ctx) error {
	locations, err := h.service.GetLocations(c.Context())
	if err != nil {
		return MapError(c, h.logger, err)
	}

	return OK(c, locations)
}

func (h *AuctionHandler) GetCountries(c *fiber.Ctx) error {
	countries, err := h.service.GetCountries(c.Context())
	if err != nil {
		return MapError(c, h.logger, err)
	}

	if countries == nil {
		countries = []models.Country{}
	}

	return OK(c, countries)
}

func (h *AuctionHandler) GetLocationsByCountry(c *fiber.Ctx) error {
	countryID, err := c.ParamsInt("countryId")
	if err != nil {
		return BadRequest(c, "Invalid country ID")
	}

	locations, err := h.service.GetLocationsByCountry(c.Context(), countryID)
	if err != nil {
		return MapError(c, h.logger, err)
	}

	return OK(c, locations)
}

func (h *AuctionHandler) GetReportReasons(c *fiber.Ctx) error {
	reasons := []fiber.Map{
		{"id": "spam", "label_ar": "إعلان مزيف أو سبام", "label_fr": "Annoncespam ou frauduleuse", "label_en": "Fake or spam ad"},
		{"id": "prohibited", "label_ar": "سلعة محظورة", "label_fr": "Article interdit", "label_en": "Prohibited item"},
		{"id": "wrong_category", "label_ar": "فئة خاطئة", "label_fr": "Mauvaise catégorie", "label_en": "Wrong category"},
		{"id": "inappropriate", "label_ar": "محتوى غير لائق", "label_fr": "Contenu inapproprié", "label_en": "Inappropriate content"},
		{"id": "price_mismatch", "label_ar": "السعر غير حقيقي", "label_fr": "Prix erroné", "label_en": "Misleading price"},
		{"id": "other", "label_ar": "أخرى", "label_fr": "Autre", "label_en": "Other"},
	}
	return OK(c, reasons)
}

func (h *AuctionHandler) GetSellerContact(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid auction ID")
	}

	auction, _, err := h.service.GetByID(c.Context(), id)
	if err != nil {
		return MapError(c, h.logger, err)
	}

	phone := ""
	if auction.PhoneContact != nil {
		phone = *auction.PhoneContact
	}

	// Masquage basic pour l'instant (####xxxx)
	if len(phone) > 4 {
		phone = "####" + phone[len(phone)-4:]
	} else {
		phone = "####"
	}

	return OK(c, fiber.Map{"phone": phone})
}

func (h *AuctionHandler) Report(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid auction ID")
	}

	type ReportRequest struct {
		Reason string `json:"reason" validate:"required,min=5"`
	}

	var req ReportRequest
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	reporterID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c)
	}

	if err := h.service.ReportAuction(c.Context(), id, reporterID, req.Reason); err != nil {
		return MapError(c, h.logger, err)
	}

	return OK(c, fiber.Map{"message": "Report submitted successfully"})
}

func (h *AuctionHandler) AddImages(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid auction ID")
	}

	type ImageRequest struct {
		URLs []string `json:"urls" validate:"required,min=1,max=10"`
	}

	var req ImageRequest
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	if len(req.URLs) == 0 {
		return BadRequest(c, "At least one image URL required")
	}

	sellerID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c)
	}

	if err := h.service.AddImages(c.Context(), id, sellerID, req.URLs); err != nil {
		return MapError(c, h.logger, err)
	}

	return OK(c, fiber.Map{"message": "Images added successfully"})
}

func (h *AuctionHandler) BuyNow(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid auction ID")
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c)
	}

	auction, err := h.service.BuyNow(c.Context(), id, userID)
	if err != nil {
		return BadRequest(c, err.Error())
	}

	return OK(c, fiber.Map{
		"message":     "Purchase completed successfully",
		"auction":     auction,
		"final_price": auction.CurrentPrice,
	})
}

func (h *AuctionHandler) Cancel(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid auction ID")
	}

	type CancelRequest struct {
		Reason string `json:"reason"`
	}
	var req CancelRequest
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c)
	}

	if err := h.service.CancelAuction(c.Context(), id, userID, req.Reason); err != nil {
		return BadRequest(c, err.Error())
	}

	return OK(c, fiber.Map{"message": "Auction canceled"})
}

func (h *AuctionHandler) Relist(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid auction ID")
	}

	type RelistRequest struct {
		EndTime string `json:"end_time"`
	}
	var req RelistRequest
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	newEndTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		return BadRequest(c, "Invalid end_time format (use RFC3339)")
	}

	if err := h.service.RelistAuction(c.Context(), id, newEndTime); err != nil {
		return BadRequest(c, err.Error())
	}

	return OK(c, fiber.Map{"message": "Auction relisted"})
}

func (h *AuctionHandler) Extend(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid auction ID")
	}

	type ExtendRequest struct {
		Hours int `json:"hours"`
	}
	var req ExtendRequest
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	if req.Hours <= 0 || req.Hours > 72 {
		return BadRequest(c, "Hours must be between 1 and 72")
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c)
	}

	if err := h.service.ExtendAuction(c.Context(), id, userID, req.Hours); err != nil {
		return BadRequest(c, err.Error())
	}

	return OK(c, fiber.Map{"message": "Auction extended"})
}

// Update - PUT /api/v1/auctions/:id - Modifier son enchère
func (h *AuctionHandler) Update(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid auction ID")
	}

	type UpdateAuctionRequest struct {
		CategoryID      *int                   `json:"category_id"`
		SubCategoryID   *int                   `json:"sub_category_id"`
		LocationID      *int                   `json:"location_id"`
		TitleAr         *string                `json:"title_ar" validate:"omitempty,min=3,max=200"`
		TitleFr         *string                `json:"title_fr"`
		TitleEn         *string                `json:"title_en"`
		DescriptionAr   *string                `json:"description_ar"`
		DescriptionFr   *string                `json:"description_fr"`
		DescriptionEn   *string                `json:"description_en"`
		StartPrice      *float64               `json:"start_price" validate:"omitempty,gt=0"`
		MinIncrement    *float64               `json:"min_increment"`
		InsuranceAmount *float64               `json:"insurance_amount"`
		EndTime         *string                `json:"end_time"`
		BuyNowPrice     *float64               `json:"buy_now_price"`
		PhoneContact    *string                `json:"phone_contact"`
		ItemDetails     map[string]interface{} `json:"item_details"`
		Condition       *string                `json:"condition" validate:"omitempty,oneof=new used refurbished damaged"`
		Brand           *string                `json:"brand"`
		VideoURL        *string                `json:"video_url"`
		Quantity        *int                   `json:"quantity" validate:"omitempty,min=1"` // Nombre d'items
	}

	var req UpdateAuctionRequest
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c)
	}

	// Parse end_time if provided
	var endTime *time.Time
	if req.EndTime != nil && *req.EndTime != "" {
		et, err := time.Parse(time.RFC3339, *req.EndTime)
		if err != nil {
			et, err = time.Parse("2006-01-02T15:04:05Z", *req.EndTime)
			if err != nil {
				return BadRequest(c, "Invalid end_time format")
			}
		}
		endTime = &et
	}

	// Convert optional fields to required CreateAuctionInput format
	var categoryID int
	if req.CategoryID != nil {
		categoryID = *req.CategoryID
	}
	var startPrice, minIncrement, insuranceAmount decimal.Decimal
	if req.StartPrice != nil {
		startPrice = decimal.NewFromFloat(*req.StartPrice)
	}
	if req.MinIncrement != nil {
		minIncrement = decimal.NewFromFloat(*req.MinIncrement)
	}
	if req.InsuranceAmount != nil {
		insuranceAmount = decimal.NewFromFloat(*req.InsuranceAmount)
	}
	var buyNowPrice *decimal.Decimal
	if req.BuyNowPrice != nil {
		b := decimal.NewFromFloat(*req.BuyNowPrice)
		buyNowPrice = &b
	}

	var titleAr, titleFr, titleEn, descriptionAr, descriptionFr, descriptionEn string
	if req.TitleAr != nil {
		titleAr = *req.TitleAr
	}
	if req.TitleFr != nil {
		titleFr = *req.TitleFr
	}
	if req.TitleEn != nil {
		titleEn = *req.TitleEn
	}
	if req.DescriptionAr != nil {
		descriptionAr = *req.DescriptionAr
	}
	if req.DescriptionFr != nil {
		descriptionFr = *req.DescriptionFr
	}
	if req.DescriptionEn != nil {
		descriptionEn = *req.DescriptionEn
	}

	var phoneContact string
	if req.PhoneContact != nil {
		phoneContact = *req.PhoneContact
	}

	var quantity int = 1 // default
	if req.Quantity != nil {
		quantity = *req.Quantity
	}

	if err := h.service.UpdateAuction(c.Context(), id, userID, services.CreateAuctionInput{
		CategoryID:      categoryID,
		SubCategoryID:   req.SubCategoryID,
		LocationID:      req.LocationID,
		TitleAr:         titleAr,
		TitleFr:         titleFr,
		TitleEn:         titleEn,
		DescriptionAr:   descriptionAr,
		DescriptionFr:   descriptionFr,
		DescriptionEn:   descriptionEn,
		StartPrice:      startPrice,
		MinIncrement:    minIncrement,
		InsuranceAmount: insuranceAmount,
		EndTime:         *endTime,
		PhoneContact:    phoneContact,
		ItemDetails:     req.ItemDetails,
		BuyNowPrice:     buyNowPrice,
		Images:          []string{},
		Condition:       req.Condition,
		Brand:           req.Brand,
		VideoURL:        req.VideoURL,
		Quantity:        quantity,
	}); err != nil {
		return BadRequest(c, err.Error())
	}

	return OK(c, fiber.Map{"message": "Auction updated successfully"})
}

// Delete - DELETE /api/v1/auctions/:id - Supprimer son enchère
func (h *AuctionHandler) Delete(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid auction ID")
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c)
	}

	if err := h.service.DeleteAuction(c.Context(), id, userID); err != nil {
		return BadRequest(c, err.Error())
	}

	return OK(c, fiber.Map{"message": "Auction deleted successfully"})
}

// GetBidStatus - GET /api/v1/auctions/:id/bid-status - Statut de ma bid
func (h *AuctionHandler) GetBidStatus(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid auction ID")
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c)
	}

	status, err := h.service.GetBidStatus(c.Context(), id, userID)
	if err != nil {
		return NotFound(c, "Bid status")
	}

	return OK(c, status)
}

// GetWinner - GET /api/v1/auctions/:id/winner - Détails du gagnant
func (h *AuctionHandler) GetWinner(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid auction ID")
	}

	auction, _, err := h.service.GetByID(c.Context(), id)
	if err != nil {
		return NotFound(c, "Auction")
	}

	if auction.WinnerID == nil {
		return OK(c, fiber.Map{
			"has_winner": false,
			"message":    "No winner yet",
		})
	}

	// Get winner details
	winner, err := h.service.GetUserByID(c.Context(), *auction.WinnerID)
	if err != nil {
		return InternalError(c, "Failed to get winner details")
	}

	return OK(c, fiber.Map{
		"has_winner":       true,
		"winner_id":        auction.WinnerID,
		"winner_name":      winner.FullName,
		"winner_phone":     maskPhone(winner.Phone),
		"winning_amount":   auction.CurrentPrice,
		"payment_deadline": auction.PaymentDeadline,
		"auction_status":   auction.Status,
	})
}

func maskPhone(phone string) string {
	if len(phone) <= 4 {
		return "****"
	}
	return phone[:2] + "****" + phone[len(phone)-2:]
}

// ContactSeller - POST /api/v1/auctions/:id/contact - Envoyer message au vendeur
func (h *AuctionHandler) ContactSeller(c *fiber.Ctx) error {
	type ContactRequest struct {
		Message string `json:"message" validate:"required"`
	}
	var req ContactRequest
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid auction ID")
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c)
	}

	// Get auction to find seller
	auction, _, err := h.service.GetByID(c.Context(), id)
	if err != nil {
		return NotFound(c, "Auction")
	}

	// Don't allow contacting yourself
	if auction.SellerID == userID {
		return BadRequest(c, "Cannot contact yourself")
	}

	// TODO: Implement notification to seller
	// For now, just return success
	h.logger.Info("[Contact Seller]",
		zap.String("auction_id", id.String()),
		zap.String("from_user", userID.String()),
		zap.String("to_seller", auction.SellerID.String()),
		zap.String("message", req.Message),
	)

	return OK(c, fiber.Map{
		"message": "Message sent to seller",
	})
}
