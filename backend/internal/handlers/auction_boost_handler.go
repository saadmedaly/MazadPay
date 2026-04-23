package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/services"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type AuctionBoostHandler struct {
	svc    services.AuctionBoostService
	logger *zap.Logger
}

func NewAuctionBoostHandler(svc services.AuctionBoostService, logger *zap.Logger) *AuctionBoostHandler {
	return &AuctionBoostHandler{svc: svc, logger: logger}
}

// CreateBoost - POST /api/auctions/:id/boost
func (h *AuctionBoostHandler) CreateBoost(c *fiber.Ctx) error {
	auctionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid auction ID")
	}

	type Request struct {
		BoostType string  `json:"boost_type" validate:"required,oneof=featured urgent top"`
		StartAt   string  `json:"start_at" validate:"required"`
		EndAt     string  `json:"end_at" validate:"required"`
		Amount    float64 `json:"amount"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	startAt, err := time.Parse(time.RFC3339, req.StartAt)
	if err != nil {
		return BadRequest(c, "Invalid start_at format, use RFC3339")
	}
	endAt, err := time.Parse(time.RFC3339, req.EndAt)
	if err != nil {
		return BadRequest(c, "Invalid end_at format, use RFC3339")
	}

	var amount *decimal.Decimal
	if req.Amount > 0 {
		a := decimal.NewFromFloat(req.Amount)
		amount = &a
	}

	boost := &models.AuctionBoost{
		AuctionID: auctionID,
		BoostType: req.BoostType,
		StartAt:   startAt,
		EndAt:     endAt,
		Amount:    amount,
	}

	if err := h.svc.Create(c.Context(), boost); err != nil {
		h.logger.Error("failed to create boost", zap.Error(err))
		return InternalError(c, "Failed to create boost")
	}

	return OK(c, fiber.Map{"message": "Boost created successfully", "boost": boost})
}

// GetAuctionBoosts - GET /api/auctions/:id/boosts
func (h *AuctionBoostHandler) GetAuctionBoosts(c *fiber.Ctx) error {
	auctionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid auction ID")
	}

	boosts, err := h.svc.ListByAuction(c.Context(), auctionID)
	if err != nil {
		h.logger.Error("failed to list auction boosts", zap.Error(err))
		return InternalError(c, "Failed to list auction boosts")
	}

	return OK(c, fiber.Map{"auction_id": auctionID, "boosts": boosts})
}

// CancelBoost - DELETE /api/auctions/:id/boosts/:boost_id
func (h *AuctionBoostHandler) CancelBoost(c *fiber.Ctx) error {
	boostID, err := uuid.Parse(c.Params("boost_id"))
	if err != nil {
		return BadRequest(c, "Invalid boost ID")
	}

	if err := h.svc.Cancel(c.Context(), boostID); err != nil {
		h.logger.Error("failed to cancel boost", zap.Error(err))
		return InternalError(c, "Failed to cancel boost")
	}

	return OK(c, fiber.Map{"message": "Boost cancelled", "boost_id": boostID})
}

// GetActiveBoosts - GET /api/admin/boosts/active (Admin only)
func (h *AuctionBoostHandler) GetActiveBoosts(c *fiber.Ctx) error {
	boosts, err := h.svc.GetActiveBoosts(c.Context())
	if err != nil {
		h.logger.Error("failed to get active boosts", zap.Error(err))
		return InternalError(c, "Failed to get active boosts")
	}

	return OK(c, fiber.Map{"boosts": boosts})
}

// UpdateBoostStatus - PUT /api/admin/boosts/:id/status (Admin only)
func (h *AuctionBoostHandler) UpdateBoostStatus(c *fiber.Ctx) error {
	boostID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid boost ID")
	}

	type Request struct {
		Status string `json:"status" validate:"required,oneof=active completed cancelled"`
	}
	var req Request
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	if err := h.svc.UpdateStatus(c.Context(), boostID, req.Status); err != nil {
		h.logger.Error("failed to update boost status", zap.Error(err))
		return InternalError(c, "Failed to update boost status")
	}

	return OK(c, fiber.Map{"message": "Boost status updated"})
}
