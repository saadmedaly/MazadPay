package handlers

import (
    "encoding/json"
    "time"

    "github.com/go-playground/validator/v10"
    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"
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
        return InternalError(c)
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

    sellerID, _ := uuid.Parse(GetUserID(c))

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
    cats, err := h.service.GetCategories(c.Context())
    if err != nil {
        return InternalError(c)
    }
    return OK(c, cats)
}

func (h *AuctionHandler) GetLocations(c *fiber.Ctx) error {
    locs, err := h.service.GetLocations(c.Context())
    if err != nil {
        return InternalError(c)
    }
    return OK(c, locs)
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

