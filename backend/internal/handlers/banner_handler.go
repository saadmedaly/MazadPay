package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/services"
	"go.uber.org/zap"
)

type BannerHandler struct {
	service services.ContentService
	logger  *zap.Logger
}

func NewBannerHandler(svc services.ContentService, logger *zap.Logger) *BannerHandler {
	return &BannerHandler{
		service: svc,
		logger:  logger,
	}
}

// List all active banners (Public)
func (h *BannerHandler) List(c *fiber.Ctx) error {
	banners, err := h.service.GetBanners(c.Context(), true)
	if err != nil {
		return MapError(c, h.logger, err)
	}
	return OK(c, banners)
}

// Admin List all banners
func (h *BannerHandler) AdminList(c *fiber.Ctx) error {
	h.logger.Debug("AdminList: fetching all banners")
	banners, err := h.service.GetBanners(c.Context(), false)
	if err != nil {
		h.logger.Error("AdminList: failed to fetch banners", zap.Error(err))
		return MapError(c, h.logger, err)
	}
	h.logger.Debug("AdminList: found banners", zap.Int("count", len(banners)))
	return OK(c, banners)
}

// Update banner status
func (h *BannerHandler) Toggle(c *fiber.Ctx) error {
	type ToggleRequest struct {
		IsActive bool `json:"is_active"`
	}

	var req ToggleRequest
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	id, _ := strconv.Atoi(c.Params("id"))
	if err := h.service.ToggleBanner(c.Context(), id, req.IsActive); err != nil {
		return MapError(c, h.logger, err)
	}

	return OK(c, fiber.Map{"message": "Banner status updated"})
}

// Create banner (Admin only)
func (h *BannerHandler) Create(c *fiber.Ctx) error {
	var banner models.Banner
	if err := c.BodyParser(&banner); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	// Validation
	if banner.ImageURL == "" {
		return BadRequest(c, "image_url is required")
	}

	if err := h.service.CreateBanner(c.Context(), &banner); err != nil {
		return MapError(c, h.logger, err)
	}

	return Created(c, banner)
}

func (h *BannerHandler) Request(c *fiber.Ctx) error {
	var banner models.Banner
	if err := c.BodyParser(&banner); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	if err := h.service.RequestBanner(c.Context(), &banner); err != nil {
		return MapError(c, h.logger, err)
	}

	return OK(c, fiber.Map{"message": "Banner request submitted successfully"})
}

// Delete banner
func (h *BannerHandler) Delete(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	if err := h.service.DeleteBanner(c.Context(), id); err != nil {
		return MapError(c, h.logger, err)
	}
	return OK(c, fiber.Map{"message": "Banner deleted successfully"})
}

// Update banner
func (h *BannerHandler) Update(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	var banner models.Banner
	if err := c.BodyParser(&banner); err != nil {
		return BadRequest(c, "Invalid request body")
	}
	banner.ID = id

	if err := h.service.UpdateBanner(c.Context(), &banner); err != nil {
		return MapError(c, h.logger, err)
	}

	return OK(c, banner)
}
