package handlers

import (
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
    Title           string                 `json:"title"            validate:"required,min=3,max=200"`
    Description     string                 `json:"description"`
    StartPrice      float64                `json:"start_price"      validate:"required,gt=0"`
    MinIncrement    float64                `json:"min_increment"    validate:"required,gt=0"`
    InsuranceAmount float64                `json:"insurance_amount" validate:"gte=0"`
    EndTime         time.Time              `json:"end_time"         validate:"required"`
    LotNumber       string                 `json:"lot_number"`
    PhoneContact    string                 `json:"phone_contact"`
    ItemDetails     map[string]interface{} `json:"item_details"`
}

func (h *AuctionHandler) Create(c *fiber.Ctx) error {
    var req CreateAuctionRequest
    if err := c.BodyParser(&req); err != nil {
        return BadRequest(c, "Invalid request body")
    }
    if err := h.validate.Struct(req); err != nil {
        return BadRequest(c, err.Error())
    }

    sellerID, _ := uuid.Parse(GetUserID(c))

    input := services.CreateAuctionInput{
        CategoryID:      req.CategoryID,
        LocationID:      req.LocationID,
        Title:           req.Title,
        Description:     req.Description,
        StartPrice:      decimal.NewFromFloat(req.StartPrice),
        MinIncrement:    decimal.NewFromFloat(req.MinIncrement),
        InsuranceAmount: decimal.NewFromFloat(req.InsuranceAmount),
        EndTime:         req.EndTime,
        LotNumber:       req.LotNumber,
        PhoneContact:    req.PhoneContact,
        ItemDetails:     req.ItemDetails,
    }

    auction, err := h.service.Create(c.Context(), sellerID, input)
    if err != nil {
        return MapError(c, h.logger, err)
    }

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

