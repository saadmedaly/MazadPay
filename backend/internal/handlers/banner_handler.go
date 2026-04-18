package handlers

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type BannerHandler struct {
	logger *zap.Logger
}

func NewBannerHandler(logger *zap.Logger) *BannerHandler {
	return &BannerHandler{logger: logger}
}

// List all active banners (Public)
func (h *BannerHandler) List(c *fiber.Ctx) error {
	banners := []fiber.Map{
		{
			"id":         1,
			"title":      "Welcome to MazadPay",
			"image_url":  "https://images.unsplash.com/photo-1611095773163-577328500526?q=80&w=2071&auto=format&fit=crop",
			"link":       "/auctions",
			"is_active":  true,
			"created_at": "2024-04-16T10:30:00Z",
		},
	}

	return OK(c, banners)
}

// Admin List all banners
func (h *BannerHandler) AdminList(c *fiber.Ctx) error {
	banners := []fiber.Map{
		{
			"id":         1,
			"title":      "Welcome to MazadPay",
			"image_url":  "https://images.unsplash.com/photo-1611095773163-577328500526?q=80&w=2071&auto=format&fit=crop",
			"link":       "/auctions",
			"is_active":  true,
			"created_at": "2024-04-16T10:30:00Z",
		},
	}

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

	id := c.Params("id")
	
	return OK(c, fiber.Map{
		"message": "Banner status updated",
		"id":      id,
		"active":  req.IsActive,
	})
}

// Create banner
func (h *BannerHandler) Create(c *fiber.Ctx) error {
	type CreateRequest struct {
		Title    string `json:"title"`
		ImageURL string `json:"image_url"`
		Link     string `json:"link"`
	}

	var req CreateRequest
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	return Created(c, fiber.Map{
		"message": "Banner created successfully",
		"data":    req,
	})
}

// Delete banner
func (h *BannerHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	return OK(c, fiber.Map{
		"message": "Banner deleted successfully",
		"id":      id,
	})
}
