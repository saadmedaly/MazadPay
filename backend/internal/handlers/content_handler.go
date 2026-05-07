package handlers

import (
	"path/filepath"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/services"
	"go.uber.org/zap"
)

type ContentHandler struct {
	svc    services.ContentService
	logger *zap.Logger
}

func NewContentHandler(svc services.ContentService, logger *zap.Logger) *ContentHandler {
	return &ContentHandler{svc: svc, logger: logger}
}

func (h *ContentHandler) FAQ(c *fiber.Ctx) error {
	items, err := h.svc.GetFAQ(c.Context())
	if err != nil {
		return MapError(c, h.logger, err)
	}
	return OK(c, items)
}

func (h *ContentHandler) AdminListFAQ(c *fiber.Ctx) error {
	items, err := h.svc.GetFAQ(c.Context())
	if err != nil {
		return MapError(c, h.logger, err)
	}
	return OK(c, items)
}

func (h *ContentHandler) CreateFAQ(c *fiber.Ctx) error {
	var item models.FAQItem
	if err := c.BodyParser(&item); err != nil {
		return BadRequest(c, "Invalid request body")
	}
	if err := h.svc.CreateFAQ(c.Context(), &item); err != nil {
		return MapError(c, h.logger, err)
	}
	return Created(c, item)
}

func (h *ContentHandler) UpdateFAQ(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	var item models.FAQItem
	if err := c.BodyParser(&item); err != nil {
		return BadRequest(c, "Invalid request body")
	}
	item.ID = id
	if err := h.svc.UpdateFAQ(c.Context(), &item); err != nil {
		return MapError(c, h.logger, err)
	}
	return OK(c, item)
}

func (h *ContentHandler) DeleteFAQ(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	if err := h.svc.DeleteFAQ(c.Context(), id); err != nil {
		return MapError(c, h.logger, err)
	}
	return OK(c, "FAQ deleted")
}

func (h *ContentHandler) Tutorials(c *fiber.Ctx) error {
	items, err := h.svc.GetTutorials(c.Context())
	if err != nil {
		return InternalError(c, "Failed to get tutorials")
	}
	return OK(c, items)
}

func (h *ContentHandler) AdminListTutorials(c *fiber.Ctx) error {
	items, err := h.svc.GetTutorials(c.Context())
	if err != nil {
		return InternalError(c, "Failed to get tutorials")
	}
	return OK(c, items)
}

func (h *ContentHandler) CreateTutorial(c *fiber.Ctx) error {
	var t models.Tutorial
	if err := c.BodyParser(&t); err != nil {
		return BadRequest(c, "Invalid request body")
	}
	if err := h.svc.CreateTutorial(c.Context(), &t); err != nil {
		return MapError(c, h.logger, err)
	}
	return Created(c, t)
}

func (h *ContentHandler) UpdateTutorial(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	var t models.Tutorial
	if err := c.BodyParser(&t); err != nil {
		return BadRequest(c, "Invalid request body")
	}
	t.ID = id
	if err := h.svc.UpdateTutorial(c.Context(), &t); err != nil {
		return MapError(c, h.logger, err)
	}
	return OK(c, t)
}

func (h *ContentHandler) DeleteTutorial(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	if err := h.svc.DeleteTutorial(c.Context(), id); err != nil {
		return MapError(c, h.logger, err)
	}
	return OK(c, "Tutorial deleted")
}

// UploadFAQImage uploads an image for FAQ to R2 (faq-tutorials/ folder)
func (h *ContentHandler) UploadFAQImage(c *fiber.Ctx) error {
	return h.uploadContentFile(c, "faq-tutorials", "image")
}

// UploadTutorialVideo uploads a video for Tutorial to R2 (faq-tutorials/ folder)
func (h *ContentHandler) UploadTutorialVideo(c *fiber.Ctx) error {
	return h.uploadContentFile(c, "faq-tutorials", "video")
}

// UploadTutorialThumbnail uploads a thumbnail for Tutorial to R2 (faq-tutorials/ folder)
func (h *ContentHandler) UploadTutorialThumbnail(c *fiber.Ctx) error {
	return h.uploadContentFile(c, "faq-tutorials", "thumbnail")
}

func (h *ContentHandler) uploadContentFile(c *fiber.Ctx, folder, fileType string) error {
	h.logger.Info("[uploadContentFile] Starting upload", zap.String("folder", folder), zap.String("type", fileType))

	// Get media service from context
	mediaSvc, ok := c.Locals("mediaService").(services.MediaService)
	if !ok {
		h.logger.Error("[uploadContentFile] Media service not available")
		return InternalError(c, "Media service not available")
	}

	// Parse multipart form
	file, err := c.FormFile("file")
	if err != nil {
		h.logger.Error("[uploadContentFile] Failed to get file", zap.Error(err))
		return BadRequest(c, "No file provided")
	}

	// Validate file size (max 100MB for videos, 10MB for images)
	maxSize := int64(10 * 1024 * 1024)
	if fileType == "video" {
		maxSize = 100 * 1024 * 1024
	}
	if file.Size > maxSize {
		h.logger.Warn("[uploadContentFile] File too large", zap.Int64("size", file.Size), zap.Int64("max", maxSize))
		return BadRequest(c, "File too large")
	}

	// Validate file type
	ext := filepath.Ext(file.Filename)
	allowedExts := map[string]map[string]bool{
		"image":     {".jpg": true, ".jpeg": true, ".png": true, ".webp": true},
		"thumbnail": {".jpg": true, ".jpeg": true, ".png": true, ".webp": true},
		"video":     {".mp4": true, ".webm": true, ".mov": true},
	}

	if typeExts, ok := allowedExts[fileType]; !ok || !typeExts[ext] {
		h.logger.Warn("[uploadContentFile] Invalid file type", zap.String("type", fileType), zap.String("ext", ext))
		return BadRequest(c, "Invalid file type for "+fileType)
	}

	// Open file
	fileReader, err := file.Open()
	if err != nil {
		h.logger.Error("[uploadContentFile] Failed to open file", zap.Error(err))
		return InternalError(c, "Failed to open file")
	}
	defer fileReader.Close()

	// Upload to R2
	url, err := mediaSvc.UploadFile(c.Context(), fileReader, file, folder)
	if err != nil {
		h.logger.Error("[uploadContentFile] R2 upload failed", zap.Error(err), zap.String("folder", folder))
		return InternalError(c, "Failed to upload file: "+err.Error())
	}

	h.logger.Info("[uploadContentFile] Upload successful", zap.String("url", url))

	return OK(c, fiber.Map{
		"message": "File uploaded successfully",
		"url":     url,
		"type":    fileType,
		"size":    file.Size,
		"name":    file.Filename,
	})
}

func (h *ContentHandler) About(c *fiber.Ctx) error {
	return OK(c, fiber.Map{
		"title":   "About MazadPay",
		"content": "MazadPay is the leading auction platform in Mauritania.",
	})
}

func (h *ContentHandler) Privacy(c *fiber.Ctx) error {
	return OK(c, fiber.Map{
		"title":   "Privacy Policy",
		"content": "We value your privacy...",
	})
}
