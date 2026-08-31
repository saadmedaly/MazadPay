package handlers

import (
	"encoding/json"
	"fmt"
	"mime/multipart"
	"strings"
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
	userRepo repository.UserRepository
	logger   *zap.Logger
	validate *validator.Validate
}

func NewAuctionHandler(svc services.AuctionService, userRepo repository.UserRepository, logger *zap.Logger) *AuctionHandler {
	return &AuctionHandler{
		service:  svc,
		userRepo: userRepo,
		logger:   logger,
		validate: validator.New(),
	}
}

func (h *AuctionHandler) List(c *fiber.Ctx) error {
	searchQuery := c.Query("q")
	// Refuse une recherche anormalement longue plutôt que de la tronquer silencieusement
	// (Public Endpoints / Scraping Protection — évite un ILIKE '%...%' sur une chaîne
	// arbitrairement grande).
	if len(searchQuery) > 100 {
		return BadRequest(c, "Search query is too long (max 100 characters)")
	}

	// This endpoint has no auth middleware (see routes.go) — never let an
	// anonymous caller pass an arbitrary ?status= value to enumerate
	// pending/rejected/canceled auctions in bulk (International Auth /
	// Product Review Phase). Only the publicly-safe statuses are honored;
	// anything else silently falls back to "active".
	requestedStatus := c.Query("status", "active")
	if !services.PubliclyVisibleAuctionStatuses[requestedStatus] {
		requestedStatus = "active"
	}

	f := repository.AuctionFilters{
		Status: requestedStatus,
		Query:  searchQuery,
		// L'app mobile envoie déjà page/limit (voir mobile/lib/services/auction_api.dart) ;
		// on les relaie tels quels, le plafond de 100 est appliqué dans auction_repo.go.
		Page:    c.QueryInt("page", 1),
		PerPage: c.QueryInt("limit", 25),
	}
	if catID := c.QueryInt("category_id", 0); catID > 0 {
		f.CategoryID = catID
	}
	// Country-scoped market (migration 000046, V1): the public listing must never
	// silently expose auctions from another market. Authenticated caller -> their
	// own account market; anonymous caller -> DefaultAccountCountryISO ('MR'),
	// matching the pre-migration app's implicit MR-only behavior, per explicit
	// product decision (do not expose e.g. future TN/MA auctions to old MR-only
	// clients just because this endpoint has no auth requirement).
	f.MarketCountryISO = models.DefaultAccountCountryISO
	// Pass user ID so expired auctions are hidden from non-owners
	if userID, err := middleware.GetUserID(c); err == nil {
		f.UserID = &userID
		if user, err := h.userRepo.FindByID(c.Context(), userID); err == nil {
			f.MarketCountryISO = user.EffectiveAccountCountryISO()
		}
	}

	auctions, err := h.service.List(c.Context(), f)
	if err != nil {
		return MapError(c, h.logger, err)
	}

	// Transform auctions to include images array instead of comma-separated string
	var response []fiber.Map
	for _, auction := range auctions {
		// Client feedback Phase B item 15: end_time is no longer capped for
		// display -- auctions can legitimately run longer than 24h now.
		response = append(response, fiber.Map{
			"id":               auction.ID,
			"seller_id":        auction.SellerID,
			"category_id":      auction.CategoryID,
			"sub_category_id":  auction.SubCategoryID,
			"location_id":      auction.LocationID,
			"title_ar":         auction.TitleAr,
			"title_fr":         auction.TitleFr,
			"title_en":         auction.TitleEn,
			"description_ar":   auction.DescriptionAr,
			"description_fr":   auction.DescriptionFr,
			"description_en":   auction.DescriptionEn,
			"start_price":      auction.StartPrice,
			"current_price":    auction.CurrentPrice,
			"min_increment":    auction.MinIncrement,
			"insurance_amount": auction.InsuranceAmount,
			"reserve_price":    auction.ReservePrice,
			"start_time":       auction.StartTime,
			"end_time":         auction.EndTime,
			"status":           auction.Status,
			"lot_number":       auction.LotNumber,
			"views":            auction.Views,
			"bidder_count":     auction.BidderCount,
			"winner_id":        auction.WinnerID,
			"winning_bid_id":   auction.WinningBidID,
			"payment_deadline": auction.PaymentDeadline,
			"is_featured":      auction.IsFeatured,
			"featured_until":   auction.FeaturedUntil,
			"rejection_reason": auction.RejectionReason,
			"item_details":     auction.ItemDetails,
			"buy_now_price":    auction.BuyNowPrice,
			"condition":        auction.Condition,
			"brand":            auction.Brand,
			"is_verified":      auction.IsVerified,
			"video_url":        auction.VideoURL,
			"quantity":         auction.Quantity,
			"category":         auction.CategoryNameAr,
			"city":             auction.CityNameAr,
			"images":           auction.GetImagesArray(),
			"created_at":       auction.CreatedAt,
			"currency_code":      auction.EffectiveCurrencyCode(),
			"market_country_iso": auction.EffectiveMarketCountryISO(),
		})
	}

	return OK(c, response)
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

	// Real-device Staging diagnosis (My Auctions "خطأ في تحميل المزاد"): resolve
	// the caller once, up front, so both the status check and the market check
	// below can recognize an owner/admin -- previously this endpoint's caller
	// lookup only existed inside the market-isolation check further down, so a
	// seller viewing their OWN pending/draft/rejected auction (e.g. right after
	// creating it, before admin approval) always 404'd here, even though the
	// mobile app's AuctionDetailsPage has no other endpoint to call and was
	// linked to from My Auctions regardless of status. The referenced "dedicated
	// seller/admin detail endpoint" in the comment this replaces never actually
	// existed anywhere in routes.go.
	var callerID uuid.UUID
	var callerUser *models.User
	callerMarket := models.DefaultAccountCountryISO
	if uid, err := middleware.GetUserID(c); err == nil {
		callerID = uid
		if user, err := h.userRepo.FindByID(c.Context(), uid); err == nil {
			callerUser = user
			callerMarket = user.EffectiveAccountCountryISO()
		}
	}
	isOwner := callerUser != nil && auction.SellerID == callerID
	isAdmin := callerUser != nil && (callerUser.Role == "admin" || callerUser.IsSuperAdmin)

	// This endpoint has no auth middleware (see routes.go) — an anonymous
	// caller must never see a pending/rejected/canceled auction by guessing
	// or reusing its ID (International Auth / Product Review Phase). The
	// owner of the auction, or an admin, may view it regardless of status.
	if !services.PubliclyVisibleAuctionStatuses[auction.Status] && !isOwner && !isAdmin {
		return NotFound(c, "Auction")
	}

	// Country-scoped market isolation (migration 000046, V1): a cross-market caller
	// must never see auction detail by ID, even though bidding is independently
	// blocked -- 404, not a "wrong market" disclosure, matching this endpoint's
	// existing convention for any other inaccessible auction (see status check
	// above). Anonymous caller -> DefaultAccountCountryISO ('MR'), consistent with
	// List()'s public-listing default. The owner/admin exemption above already
	// covers status; market isolation still applies to everyone (an owner's own
	// auction is always in their own market by construction, so this never
	// actually blocks a legitimate owner in practice).
	if callerMarket != auction.EffectiveMarketCountryISO() && !isAdmin {
		return NotFound(c, "Auction")
	}

	// Client feedback Phase B item 15: end_time is no longer capped for
	// display -- auctions can legitimately run longer than 24h now.
	return OK(c, fiber.Map{
		"auction": fiber.Map{
			"id":               auction.ID,
			"seller_id":        auction.SellerID,
			"category_id":      auction.CategoryID,
			"sub_category_id":  auction.SubCategoryID,
			"location_id":      auction.LocationID,
			"title_ar":         auction.TitleAr,
			"title_fr":         auction.TitleFr,
			"title_en":         auction.TitleEn,
			"description_ar":   auction.DescriptionAr,
			"description_fr":   auction.DescriptionFr,
			"description_en":   auction.DescriptionEn,
			"start_price":      auction.StartPrice,
			"current_price":    auction.CurrentPrice,
			"min_increment":    auction.MinIncrement,
			"insurance_amount": auction.InsuranceAmount,
			"reserve_price":    auction.ReservePrice,
			"start_time":       auction.StartTime,
			"end_time":         auction.EndTime,
			"status":           auction.Status,
			"lot_number":       auction.LotNumber,
			"views":            auction.Views,
			"bidder_count":     auction.BidderCount,
			"winner_id":        auction.WinnerID,
			"winning_bid_id":   auction.WinningBidID,
			"payment_deadline": auction.PaymentDeadline,
			"is_featured":      auction.IsFeatured,
			"featured_until":   auction.FeaturedUntil,
			"rejection_reason": auction.RejectionReason,
			"item_details":     auction.ItemDetails,
			"buy_now_price":    auction.BuyNowPrice,
			"condition":        auction.Condition,
			"brand":            auction.Brand,
			"is_verified":      auction.IsVerified,
			"video_url":        auction.VideoURL,
			"quantity":         auction.Quantity,
			"category":         auction.CategoryNameAr,
			"city":             auction.CityNameAr,
			"image_urls":       auction.GetImagesArray(),
			"currency_code":      auction.EffectiveCurrencyCode(),
			"market_country_iso": auction.EffectiveMarketCountryISO(),
		},
		"images": images,
	})
}

type CreateAuctionRequest struct {

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
	Condition *string `json:"condition" validate:"omitempty,oneof=new used refurbished damaged"`
	Brand     *string `json:"brand"`
	VideoURL  *string `json:"video_url"`
	// New field from migration 000032
	Quantity int `json:"quantity" validate:"omitempty,min=1"` // Nombre d'items (défaut: 1)
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

	// Client feedback Phase B item 15: the 24h max-duration cap is removed --
	// multi-day auctions (48h, 72h, several days) must now be allowed. The
	// only remaining constraint is end_time > start_time, which the previous
	// code never actually checked on its own (a bad end_time <= start_time
	// would have been silently overwritten by the old cap instead of
	// rejected) -- checked explicitly here now that the cap no longer masks it.
	effectiveStart := time.Now()
	if startTimePtr != nil {
		effectiveStart = *startTimePtr
	}
	if !endTime.After(effectiveStart) {
		return BadRequest(c, "end_time must be after start_time")
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
	var userID *uuid.UUID
	if uid, err := middleware.GetUserID(c); err == nil {
		userID = &uid
	}
	_ = h.service.IncrementViews(c.Context(), id, userID)
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

	// Country-scoped market isolation (migration 000046, V1): this route already
	// requires JWT auth (see routes.go) -- no anonymous branch needed, unlike
	// GetByID/History. 404, not a "wrong market" disclosure, matching this handler's
	// established convention -- checked BEFORE the phone number is read/masked below,
	// so no seller contact data is ever computed for a denied caller.
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c, "User not authenticated")
	}
	user, err := h.userRepo.FindByID(c.Context(), userID)
	if err != nil {
		return NotFound(c, "Auction")
	}
	if user.EffectiveAccountCountryISO() != auction.EffectiveMarketCountryISO() {
		return NotFound(c, "Auction")
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
		ReasonID string `json:"reason_id" validate:"required"`
		Details  string `json:"details"`
	}

	var req ReportRequest
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	if req.ReasonID == "" {
		return BadRequest(c, "Reason ID is required")
	}

	reporterID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c)
	}

	// Combine reason_id and details for the service
	reason := req.ReasonID
	if req.Details != "" {
		reason = fmt.Sprintf("[%s] %s", req.ReasonID, req.Details)
	}

	if err := h.service.ReportAuction(c.Context(), id, reporterID, reason); err != nil {
		return MapError(c, h.logger, err)
	}

	return OK(c, fiber.Map{"message": "Report submitted successfully"})
}

func (h *AuctionHandler) AddImages(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid auction ID")
	}

	sellerID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c)
	}

	// Check if multipart form data (file upload) or JSON (URLs)
	contentType := strings.TrimSpace(c.Get("Content-Type"))
	h.logger.Info("[AddImages] Request received", zap.String("content_type", contentType), zap.String("auction_id", id.String()))

	isMultipart := len(c.Request().Header.MultipartFormBoundary()) > 0 ||
		strings.HasPrefix(strings.ToLower(contentType), "multipart/form-data")

	if isMultipart {
		return h.handleMultipartImages(c, id, sellerID)
	}

	// Fallback to URL-based images (legacy support)
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

	if err := h.service.AddImages(c.Context(), id, sellerID, req.URLs); err != nil {
		return MapError(c, h.logger, err)
	}

	return OK(c, fiber.Map{"message": "Images added successfully"})
}

func (h *AuctionHandler) handleMultipartImages(c *fiber.Ctx, auctionID, sellerID uuid.UUID) error {
	h.logger.Info("[handleMultipartImages] Starting image upload process",
		zap.String("auction_id", auctionID.String()),
		zap.String("seller_id", sellerID.String()))

	// Pre-validate auction ownership before uploading to R2
	auction, _, err := h.service.GetByID(c.Context(), auctionID)
	if err != nil {
		h.logger.Error("[handleMultipartImages] Auction not found",
			zap.Error(err),
			zap.String("auction_id", auctionID.String()))
		return NotFound(c, "Auction")
	}
	if auction.SellerID != sellerID {
		h.logger.Warn("[handleMultipartImages] Unauthorized - user does not own auction",
			zap.String("auction_id", auctionID.String()),
			zap.String("seller_id", sellerID.String()),
			zap.String("actual_seller", auction.SellerID.String()))
		return Forbidden(c)
	}

	h.logger.Info("[handleMultipartImages] Auction ownership validated",
		zap.String("auction_id", auctionID.String()))

	// Get the media service from the context (set up in routes)
	mediaSvc, ok := c.Locals("mediaService").(services.MediaService)
	if !ok {
		h.logger.Error("[handleMultipartImages] Media service not available in context")
		return InternalError(c, "Media service not available")
	}

	// Parse multipart form (max 10 files, max 10MB each)
	form, err := c.MultipartForm()
	if err != nil {
		h.logger.Error("[handleMultipartImages] Failed to parse multipart form",
			zap.Error(err))
		return BadRequest(c, "Failed to parse form data: "+err.Error())
	}

	files := form.File["images"]
	if len(files) == 0 {
		h.logger.Warn("[handleMultipartImages] No images provided in form data")
		return BadRequest(c, "No images provided")
	}

	if len(files) > 10 {
		h.logger.Warn("[handleMultipartImages] Too many images",
			zap.Int("count", len(files)))
		return BadRequest(c, "Maximum 10 images allowed")
	}

	h.logger.Info("[handleMultipartImages] Processing files",
		zap.Int("file_count", len(files)))

	var fileReaders []multipart.File
	var headers []*multipart.FileHeader

	// Ensure all file handles are closed even if upload fails
	defer func() {
		for _, f := range fileReaders {
			if f != nil {
				f.Close()
			}
		}
	}()

	for i, file := range files {
		// Validate file size (max 10MB)
		if file.Size > 10*1024*1024 {
			h.logger.Warn("[handleMultipartImages] File too large",
				zap.String("filename", file.Filename),
				zap.Int64("size", file.Size))
			return BadRequest(c, "File too large: "+file.Filename+" (max 10MB)")
		}

		// Validate file type
		allowedTypes := map[string]bool{
			"image/jpeg": true,
			"image/jpg":  true,
			"image/png":  true,
			"image/webp": true,
		}
		contentType := file.Header.Get("Content-Type")
		if !allowedTypes[contentType] {
			h.logger.Warn("[handleMultipartImages] Invalid file type",
				zap.String("filename", file.Filename),
				zap.String("content_type", contentType))
			return BadRequest(c, "Invalid file type: "+file.Filename+" (only JPEG, PNG, WebP allowed)")
		}

		f, err := file.Open()
		if err != nil {
			h.logger.Error("[handleMultipartImages] Failed to open file",
				zap.Error(err),
				zap.String("filename", file.Filename))
			return InternalError(c, "Failed to open file: "+file.Filename)
		}
		fileReaders = append(fileReaders, f)
		headers = append(headers, file)
		h.logger.Info("[handleMultipartImages] File validated and opened",
			zap.Int("index", i+1),
			zap.String("filename", file.Filename),
			zap.Int64("size", file.Size))
	}

	// Upload to R2
	h.logger.Info("[handleMultipartImages] Starting R2 upload",
		zap.String("auction_id", auctionID.String()),
		zap.Int("file_count", len(fileReaders)))

	urls, err := mediaSvc.UploadAuctionImages(c.Context(), fileReaders, headers, auctionID)
	if err != nil {
		h.logger.Error("[handleMultipartImages] R2 upload failed - detailed error",
			zap.Error(err),
			zap.String("auction_id", auctionID.String()),
			zap.String("error_type", "CLOUDFLARE_R2_ERROR"),
			zap.Int("attempted_files", len(fileReaders)))
		return InternalError(c, "Failed to upload images to Cloudflare R2: "+err.Error())
	}

	h.logger.Info("[handleMultipartImages] R2 upload successful",
		zap.String("auction_id", auctionID.String()),
		zap.Int("uploaded_count", len(urls)))

	// Save URLs to database
	if err := h.service.AddImages(c.Context(), auctionID, sellerID, urls); err != nil {
		h.logger.Error("[handleMultipartImages] Failed to save image URLs to database - cleaning up R2",
			zap.Error(err),
			zap.String("auction_id", auctionID.String()),
			zap.Strings("urls", urls))

		// Clean up uploaded images from R2 to prevent orphaned files
		for _, url := range urls {
			if delErr := mediaSvc.DeleteFile(c.Context(), url); delErr != nil {
				h.logger.Warn("[handleMultipartImages] Failed to cleanup R2 image after DB error",
					zap.String("url", url),
					zap.Error(delErr))
			} else {
				h.logger.Info("[handleMultipartImages] Cleaned up R2 image after DB error",
					zap.String("url", url))
			}
		}

		return MapError(c, h.logger, err)
	}

	h.logger.Info("[handleMultipartImages] Images saved to database successfully",
		zap.String("auction_id", auctionID.String()),
		zap.Int("count", len(urls)))

	return OK(c, fiber.Map{
		"message": "Images uploaded successfully",
		"urls":    urls,
		"count":   len(urls),
	})
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

	userID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c)
	}

	if err := h.service.RelistAuction(c.Context(), id, userID, newEndTime); err != nil {
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
		"currency_code":    auction.EffectiveCurrencyCode(),
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

