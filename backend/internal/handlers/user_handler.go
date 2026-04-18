package handlers

import (
    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"
    "github.com/mazadpay/backend/internal/models"
    "github.com/mazadpay/backend/internal/services"
    "go.uber.org/zap"
)

type UserHandler struct {
    service services.UserService
    logger  *zap.Logger
}

func NewUserHandler(svc services.UserService, logger *zap.Logger) *UserHandler {
    return &UserHandler{service: svc, logger: logger}
}


func (h *UserHandler) GetMe(c *fiber.Ctx) error {
	userID, err := uuid.Parse(GetUserID(c))
	if err != nil {
		return Unauthorized(c)
	}

	user, err := h.service.GetProfile(c.Context(), userID)
	if err != nil {
		return MapError(c, h.logger, err)
	}

	return OK(c, fiber.Map{"data": user})
}

func (h *UserHandler) UpdateProfile(c *fiber.Ctx) error {
	type Request struct {
		FullName string `json:"full_name"`
		Email    string `json:"email"`
		City     string `json:"city"`
	}
	var req Request
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	userID := c.Locals("userID").(uuid.UUID)
	if err := h.service.UpdateProfile(c.Context(), userID, req.FullName, req.Email, req.City); err != nil {
		return InternalError(c, "Failed to update profile")
	}
	return OK(c, fiber.Map{"message": "Profile updated"})
}

func (h *UserHandler) UpdateAvatar(c *fiber.Ctx) error {
	type Request struct {
		URL string `json:"url"`
	}
	var req Request
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	userID := c.Locals("userID").(uuid.UUID)
	if err := h.service.UpdateAvatar(c.Context(), userID, req.URL); err != nil {
		return InternalError(c, "Failed to update avatar")
	}
	return OK(c, fiber.Map{"message": "Avatar updated"})
}

func (h *UserHandler) AddFavorite(c *fiber.Ctx) error {
	auctionID, err := uuid.Parse(c.Params("auction_id"))
	if err != nil {
		return BadRequest(c, "Invalid auction ID")
	}

	userID := c.Locals("userID").(uuid.UUID)
	if err := h.service.AddFavorite(c.Context(), userID, auctionID); err != nil {
		return InternalError(c, "Failed to add favorite")
	}
	return OK(c, fiber.Map{"message": "Auction added to favorites"})
}

func (h *UserHandler) RemoveFavorite(c *fiber.Ctx) error {
	auctionID, err := uuid.Parse(c.Params("auction_id"))
	if err != nil {
		return BadRequest(c, "Invalid auction ID")
	}

	userID := c.Locals("userID").(uuid.UUID)
	if err := h.service.RemoveFavorite(c.Context(), userID, auctionID); err != nil {
		return InternalError(c, "Failed to remove favorite")
	}
	return OK(c, fiber.Map{"message": "Auction removed from favorites"})
}

func (h *UserHandler) ListFavorites(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)
	auctions, err := h.service.ListFavorites(c.Context(), userID)
	if err != nil {
		return InternalError(c, "Failed to get favorites")
	}
	return OK(c, fiber.Map{"data": auctions})
}

func (h *UserHandler) MyAuctions(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)
	auctions, err := h.service.ListMyAuctions(c.Context(), userID)
	if err != nil {
		return InternalError(c, "Failed to get my auctions")
	}
	return OK(c, fiber.Map{"data": auctions})
}

func (h *UserHandler) MyBids(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)
	auctions, err := h.service.ListMyBids(c.Context(), userID)
	if err != nil {
		return InternalError(c, "Failed to get my bids")
	}
	return OK(c, fiber.Map{"data": auctions})
}

func (h *UserHandler) MyWinnings(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)
	auctions, err := h.service.ListMyWinnings(c.Context(), userID)
	if err != nil {
		return InternalError(c, "Failed to get my winnings")
	}
	return OK(c, fiber.Map{"data": auctions})
}

func (h *UserHandler) SubmitKYC(c *fiber.Ctx) error {
	var kyc models.KYCVerification
	if err := c.BodyParser(&kyc); err != nil {
		return BadRequest(c, "Invalid request body")
	}
	kyc.UserID = c.Locals("userID").(uuid.UUID)
	if err := h.service.SubmitKYC(c.Context(), &kyc); err != nil {
		return InternalError(c, "Failed to submit KYC")
	}
	return OK(c, fiber.Map{"message": "KYC submitted successfully"})
}

func (h *UserHandler) GetKYCStatus(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)
	kyc, err := h.service.GetKYCStatus(c.Context(), userID)
	if err != nil {
		return NotFound(c, "KYC not found")
	}
	return OK(c, fiber.Map{"data": kyc})
}
