package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/middleware"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/services"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type BidAutoBidHandler struct {
	svc    services.BidAutoBidService
	logger *zap.Logger
}

func NewBidAutoBidHandler(svc services.BidAutoBidService, logger *zap.Logger) *BidAutoBidHandler {
	return &BidAutoBidHandler{svc: svc, logger: logger}
}

// CreateAutoBid - POST /api/auctions/:id/auto-bid
func (h *BidAutoBidHandler) CreateAutoBid(c *fiber.Ctx) error {
	auctionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid auction ID")
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c)
	}

	type Request struct {
		MaxAmount float64 `json:"max_amount" validate:"required,gt=0"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	autoBid := &models.BidAutoBid{
		UserID:    userID,
		AuctionID: auctionID,
		MaxAmount: decimal.NewFromFloat(req.MaxAmount),
		IsActive:  true,
	}

	if err := h.svc.Create(c.Context(), autoBid); err != nil {
		h.logger.Error("failed to create auto-bid", zap.Error(err))
		return InternalError(c, "Failed to create auto-bid")
	}

	return OK(c, fiber.Map{"message": "Auto-bid created successfully", "auto_bid": autoBid})
}

// GetMyAutoBids - GET /api/users/auto-bids
func (h *BidAutoBidHandler) GetMyAutoBids(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c)
	}

	autoBids, err := h.svc.ListByUser(c.Context(), userID)
	if err != nil {
		h.logger.Error("failed to list auto-bids", zap.Error(err))
		return InternalError(c, "Failed to list auto-bids")
	}

	return OK(c, fiber.Map{"user_id": userID, "auto_bids": autoBids})
}

// GetAuctionAutoBids - GET /api/auctions/:id/auto-bids (Admin only)
func (h *BidAutoBidHandler) GetAuctionAutoBids(c *fiber.Ctx) error {
	auctionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid auction ID")
	}

	autoBids, err := h.svc.ListByAuction(c.Context(), auctionID)
	if err != nil {
		h.logger.Error("failed to list auction auto-bids", zap.Error(err))
		return InternalError(c, "Failed to list auction auto-bids")
	}

	return OK(c, fiber.Map{"auction_id": auctionID, "auto_bids": autoBids})
}

// CancelAutoBid - DELETE /api/auctions/:id/auto-bid
func (h *BidAutoBidHandler) CancelAutoBid(c *fiber.Ctx) error {
	auctionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid auction ID")
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c)
	}

	// Find the user's auto bid for this auction
	autoBids, err := h.svc.ListByUser(c.Context(), userID)
	if err != nil {
		h.logger.Error("failed to find auto-bid", zap.Error(err))
		return InternalError(c, "Failed to find auto-bid")
	}

	for _, ab := range autoBids {
		if ab.AuctionID == auctionID {
			if err := h.svc.Delete(c.Context(), ab.ID); err != nil {
				h.logger.Error("failed to cancel auto-bid", zap.Error(err))
				return InternalError(c, "Failed to cancel auto-bid")
			}
			return OK(c, fiber.Map{"message": "Auto-bid cancelled", "auction_id": auctionID})
		}
	}

	return NotFound(c, "Auto-bid")
}

// UpdateAutoBid - PUT /api/auctions/:id/auto-bid
func (h *BidAutoBidHandler) UpdateAutoBid(c *fiber.Ctx) error {
	auctionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid auction ID")
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c)
	}

	type Request struct {
		MaxAmount float64 `json:"max_amount" validate:"required,gt=0"`
		IsActive  bool    `json:"is_active"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	// Find the user's auto bid for this auction
	autoBids, err := h.svc.ListByUser(c.Context(), userID)
	if err != nil {
		h.logger.Error("failed to find auto-bid", zap.Error(err))
		return InternalError(c, "Failed to find auto-bid")
	}

	for _, ab := range autoBids {
		if ab.AuctionID == auctionID {
			updated := &models.BidAutoBid{
				MaxAmount: decimal.NewFromFloat(req.MaxAmount),
				IsActive:  req.IsActive,
			}
			if err := h.svc.Update(c.Context(), ab.ID, updated); err != nil {
				h.logger.Error("failed to update auto-bid", zap.Error(err))
				return InternalError(c, "Failed to update auto-bid")
			}
			return OK(c, fiber.Map{"message": "Auto-bid updated", "auction_id": auctionID})
		}
	}

	return NotFound(c, "Auto-bid")
}
