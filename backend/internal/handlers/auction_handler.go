package handlers

import (
    "encoding/json"
    "time"

    "github.com/go-playground/validator/v10"
    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"
    "github.com/mazadpay/backend/internal/middleware"
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
    CategoryID      int                    `json:"category_id"      validate:"required"`
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

    input := services.CreateAuctionInput{
        CategoryID:      req.CategoryID,
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
		"message":    "Purchase completed successfully",
		"auction":   auction,
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

