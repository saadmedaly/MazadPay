package handlers

import (
    "github.com/go-playground/validator/v10"
    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"
    "github.com/mazadpay/backend/internal/middleware"
    "github.com/mazadpay/backend/internal/services"
    "github.com/shopspring/decimal"
    "go.uber.org/zap"
)

type BidHandler struct {
    service  services.BidService
    logger   *zap.Logger
    validate *validator.Validate
}

func NewBidHandler(svc services.BidService, logger *zap.Logger) *BidHandler {
    return &BidHandler{
        service:  svc,
        logger:   logger,
        validate: validator.New(),
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

    bids, err := h.service.GetHistory(c.Context(), auctionID)
    if err != nil {
        return MapError(c, h.logger, err)
    }


    return OK(c, bids)
}
