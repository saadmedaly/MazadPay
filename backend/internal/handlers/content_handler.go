package handlers

import (
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
		return InternalError(c, "Failed to get FAQ")
	}
	return OK(c, items)
}

func (h *ContentHandler) CreateFAQ(c *fiber.Ctx) error {
	var item models.FAQItem
	if err := c.BodyParser(&item); err != nil {
		return BadRequest(c, "Invalid request body")
	}
	if err := h.svc.CreateFAQ(c.Context(), &item); err != nil {
		return InternalError(c, "Failed to create FAQ")
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
		return InternalError(c, "Failed to update FAQ")
	}
	return OK(c, item)
}

func (h *ContentHandler) DeleteFAQ(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	if err := h.svc.DeleteFAQ(c.Context(), id); err != nil {
		return InternalError(c, "Failed to delete FAQ")
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

func (h *ContentHandler) CreateTutorial(c *fiber.Ctx) error {
	var t models.Tutorial
	if err := c.BodyParser(&t); err != nil {
		return BadRequest(c, "Invalid request body")
	}
	if err := h.svc.CreateTutorial(c.Context(), &t); err != nil {
		return InternalError(c, "Failed to create tutorial")
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
		return InternalError(c, "Failed to update tutorial")
	}
	return OK(c, t)
}

func (h *ContentHandler) DeleteTutorial(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	if err := h.svc.DeleteTutorial(c.Context(), id); err != nil {
		return InternalError(c, "Failed to delete tutorial")
	}
	return OK(c, "Tutorial deleted")
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
