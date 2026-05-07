package handlers

import (
	"path/filepath"
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
	banners, err := h.service.GetBanners(c.Context(), false)
	if err != nil {
		return MapError(c, h.logger, err)
	}
	if banners == nil {
		banners = []models.Banner{}
	}
	return OK(c, banners)
}

// Admin List ALL banners (including inactive) - explicit endpoint for debugging
func (h *BannerHandler) AdminListAll(c *fiber.Ctx) error {
	banners, err := h.service.GetBanners(c.Context(), false)
	if err != nil {
		h.logger.Error("AdminListAll: failed", zap.Error(err))
		return MapError(c, h.logger, err)
	}
	h.logger.Info("AdminListAll: found banners", zap.Int("count", len(banners)), zap.Any("banners", banners))
	if banners == nil {
		banners = []models.Banner{}
	}
	return OK(c, fiber.Map{"success": true, "data": banners, "total": len(banners)})
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

// UploadBannerImage uploads a banner image to R2 (banners/ folder)
func (h *BannerHandler) UploadBannerImage(c *fiber.Ctx) error {
	h.logger.Info("[UploadBannerImage] Starting banner upload")

	// Get media service from context
	mediaSvc, ok := c.Locals("mediaService").(services.MediaService)
	if !ok {
		h.logger.Error("[UploadBannerImage] Media service not available")
		return InternalError(c, "Media service not available")
	}

	// Parse multipart form
	file, err := c.FormFile("file")
	if err != nil {
		h.logger.Error("[UploadBannerImage] Failed to get file", zap.Error(err))
		return BadRequest(c, "No file provided")
	}

	// Validate file size (max 10MB for banners)
	if file.Size > 10*1024*1024 {
		h.logger.Warn("[UploadBannerImage] File too large", zap.Int64("size", file.Size))
		return BadRequest(c, "File too large (max 10MB)")
	}

	// Validate file type (only images)
	ext := filepath.Ext(file.Filename)
	allowedExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".gif": true,
	}
	if !allowedExts[ext] {
		h.logger.Warn("[UploadBannerImage] Invalid file type", zap.String("ext", ext))
		return BadRequest(c, "Invalid file type (only jpg, jpeg, png, webp, gif allowed)")
	}

	// Open file
	fileReader, err := file.Open()
	if err != nil {
		h.logger.Error("[UploadBannerImage] Failed to open file", zap.Error(err))
		return InternalError(c, "Failed to open file")
	}
	defer fileReader.Close()

	// Upload to R2 (banners/ folder)
	url, err := mediaSvc.UploadFile(c.Context(), fileReader, file, "banners")
	if err != nil {
		h.logger.Error("[UploadBannerImage] R2 upload failed", zap.Error(err))
		return InternalError(c, "Failed to upload banner: "+err.Error())
	}

	h.logger.Info("[UploadBannerImage] Upload successful", zap.String("url", url))

	return OK(c, fiber.Map{
		"message": "Banner uploaded successfully",
		"url":     url,
		"type":    "image",
		"size":    file.Size,
		"name":    file.Filename,
	})
}
