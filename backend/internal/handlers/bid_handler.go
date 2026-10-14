package handlers

import (
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

type BidHandler struct {
    service     services.BidService
    auctionRepo repository.AuctionRepository
    userRepo    repository.UserRepository
    logger      *zap.Logger
    validate    *validator.Validate
}

func NewBidHandler(svc services.BidService, auctionRepo repository.AuctionRepository, userRepo repository.UserRepository, logger *zap.Logger) *BidHandler {
    return &BidHandler{
        service:     svc,
        auctionRepo: auctionRepo,
        userRepo:    userRepo,
        logger:      logger,
        validate:    validator.New(),
    }
}


type PlaceBidRequest struct {
    Amount float64 `json:"amount" validate:"required,gt=0"`
}

func (h *BidHandler) Place(c *fiber.Ctx) error {
    auctionID, err := uuid.Parse(c.Params("id"))
    if err != nil {
        return BadRequest(c, "Invalid auction ID")
    }

    var req PlaceBidRequest
    if err := c.BodyParser(&req); err != nil {
        return BadRequest(c, "Invalid request body")
    }
    if err := h.validate.Struct(req); err != nil {
        return BadRequest(c, err.Error())
    }

    userID, err := middleware.GetUserID(c)
    if err != nil {
        return Unauthorized(c)
    }
    amount := decimal.NewFromFloat(req.Amount)

    bid, err := h.service.PlaceBid(c.Context(), auctionID, userID, amount)
    if err != nil {
        return MapError(c, h.logger, err)
    }


    return Created(c, bid)
}

func (h *BidHandler) History(c *fiber.Ctx) error {
    auctionID, err := uuid.Parse(c.Params("id"))
    if err != nil {
        return BadRequest(c, "Invalid auction ID")
    }

    // Country-scoped market isolation (migration 000046, V1): this route has no auth
    // middleware (anonymous access is intentional, matching GET /auctions/:id), so a
    // caller may or may not carry a JWT. Authenticated -> caller's own effective
    // account market; anonymous -> DefaultAccountCountryISO ('MR'), same convention as
    // auction_handler.go List()/GetByID(). 404, not a "wrong market" disclosure --
    // matches this endpoint's own existing "auction not found" shape for a bad ID.
    auction, err := h.auctionRepo.FindByID(c.Context(), auctionID)
    if err != nil {
        return NotFound(c, "Auction")
    }
    var callerID uuid.UUID
    callerMarket := models.DefaultAccountCountryISO
    if userID, err := middleware.GetUserID(c); err == nil {
        callerID = userID
        if user, err := h.userRepo.FindByID(c.Context(), userID); err == nil {
            callerMarket = user.EffectiveAccountCountryISO()
        }
    }
    // Legacy owner fix (client feedback, round 5): mirrors the same owner
    // exemption already applied to GET /auctions/:id (auction_handler.go) and
    // the WebSocket subscription (ws_handler.go AuthorizeSubscription). A
    // legacy auction predating migration 000046 has market_country_iso/
    // currency_code = NULL, falling back to "MR" -- so the auction's own
    // owner (e.g. a TN seller) was 404'd out of their own bid history,
    // confirmed live against Staging (auction detail + WebSocket already
    // fixed for this exact case, this endpoint was the remaining gap). Every
    // other caller (non-owner, cross-market, anonymous) is unaffected.
    isOwner := callerID != uuid.Nil && auction.SellerID == callerID
    if callerMarket != auction.EffectiveMarketCountryISO() && !isOwner {
        return NotFound(c, "Auction")
    }

    bids, err := h.service.GetHistory(c.Context(), auctionID)
    if err != nil {
        return MapError(c, h.logger, err)
    }


    return OK(c, bids)
}
