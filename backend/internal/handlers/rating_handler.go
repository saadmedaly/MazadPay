package handlers

import (
	"github.com/google/uuid"
	"github.com/gofiber/fiber/v2"
	"github.com/mazadpay/backend/internal/middleware"
	"github.com/mazadpay/backend/internal/services"
	"go.uber.org/zap"
)

type RatingHandler struct {
	service services.RatingService
	logger  *zap.Logger
}

func NewRatingHandler(service services.RatingService, logger *zap.Logger) *RatingHandler {
	return &RatingHandler{
		service: service,
		logger:  logger,
	}
}

type CreateAppRatingRequest struct {
	Title   string `json:"title" validate:"required"`
	Rating  int    `json:"rating" validate:"required,min=1,max=5"`
	Comment string `json:"comment"`
}

func (h *RatingHandler) CreateAppRating(c *fiber.Ctx) error {
	var req CreateAppRatingRequest
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c)
	}

	if err := h.service.CreateAppRating(c.Context(), userID, req.Title, req.Rating, req.Comment); err != nil {
		return MapError(c, h.logger, err)
	}

	return OK(c, fiber.Map{"message": "Rating submitted successfully"})
}

func (h *RatingHandler) GetAppStats(c *fiber.Ctx) error {
	avgRating, total, err := h.service.GetAppStats(c.Context())
	if err != nil {
		return MapError(c, h.logger, err)
	}

	return OK(c, fiber.Map{
		"average_rating": avgRating,
		"total_ratings":  total,
	})
}

func (h *RatingHandler) ListAppRatings(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)

	ratings, total, err := h.service.ListAppRatings(c.Context(), page, limit)
	if err != nil {
		return MapError(c, h.logger, err)
	}

	return OK(c, fiber.Map{
		"ratings": ratings,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}

func (h *RatingHandler) DeleteAppRating(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid rating ID")
	}

	if err := h.service.DeleteAppRating(c.Context(), id); err != nil {
		return MapError(c, h.logger, err)
	}

	return OK(c, fiber.Map{"message": "Rating deleted successfully"})
}
