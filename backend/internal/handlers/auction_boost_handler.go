package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/middleware"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/repository"
	"github.com/mazadpay/backend/internal/services"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type AuctionBoostHandler struct {
	svc         services.AuctionBoostService
	auctionRepo repository.AuctionRepository
	userRepo    repository.UserRepository
	logger      *zap.Logger
}

func NewAuctionBoostHandler(svc services.AuctionBoostService, auctionRepo repository.AuctionRepository, userRepo repository.UserRepository, logger *zap.Logger) *AuctionBoostHandler {
	return &AuctionBoostHandler{svc: svc, auctionRepo: auctionRepo, userRepo: userRepo, logger: logger}
}

// CreateBoost - POST /api/auctions/:id/boost
func (h *AuctionBoostHandler) CreateBoost(c *fiber.Ctx) error {
	auctionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid auction ID")
	}

	// Country-scoped market isolation (migration 000046, V1): a customer write action
	// must never be able to target an auction outside the caller's own effective
	// market -- checked by COUNTRY equality only, never currency, same invariant as
	// GetAuctionBoosts/PlaceBid. Trusted sources only: JWT-derived userID, the auction
	// loaded from auctionRepo, the caller's account market from userRepo. 404, not a
	// "wrong market" disclosure, matching this handler's read-side convention.
	auction, err := h.auctionRepo.FindByID(c.Context(), auctionID)
	if err != nil {
		return NotFound(c, "Auction")
	}
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

	// Country-scoped market isolation (migration 000046, V1): this route already
	// requires JWT auth (see routes.go) -- no anonymous branch needed. 404, not a
	// "wrong market" disclosure, matching auction_handler.go's established
	// convention. Checked before any boost data is fetched/returned.
	auction, err := h.auctionRepo.FindByID(c.Context(), auctionID)
	if err != nil {
		return NotFound(c, "Auction")
	}
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
