package handlers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/middleware"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/services"
	"go.uber.org/zap"
)

type RequestHandler struct {
	svc      services.RequestService
	logger   *zap.Logger
	validate *validator.Validate
}

func NewRequestHandler(svc services.RequestService, logger *zap.Logger) *RequestHandler {
	return &RequestHandler{svc: svc, logger: logger, validate: newDecimalAwareValidator()}
}

type ReviewRequest struct {
	Status string `json:"status" validate:"required,oneof=approved rejected"`
	Notes  string `json:"notes"`
}

// requiresNotesWhenRejected renvoie une erreur claire si le statut soumis est "rejected"
// et qu'aucun motif n'est fourni — notes n'est validable par une simple tag struct
// puisqu'elle n'est requise que conditionnellement au statut soumis dans le même corps
// de requête.
func requiresNotesWhenRejected(status, notes string) error {
	if status == "rejected" && strings.TrimSpace(notes) == "" {
		return fmt.Errorf("notes is required when status is rejected")
	}
	return nil
}

// Create Auction Request (public endpoint for users)
func (h *RequestHandler) CreateAuctionRequest(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c, "User not authenticated")
	}

	var req models.AuctionRequest
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	// UserID is always the authenticated caller, never client-supplied -- assigned
	// here, before validation, so any user_id present in the request body (whether
	// benign or a spoofing attempt naming a different user) is unconditionally
	// discarded and cannot reach the service/repository layer.
	req.ID = uuid.New()
	req.UserID = userID

	if err := h.validate.Struct(req); err != nil {
		return BadRequest(c, err.Error())
	}

	// Set default quantity if not provided
	if req.Quantity == 0 {
		req.Quantity = 1
	}

	if err := h.svc.CreateAuctionRequest(c.Context(), &req); err != nil {
		return MapError(c, h.logger, err)
	}

	return OK(c, fiber.Map{"message": "Auction request submitted successfully", "id": req.ID})
}

// Create Banner Request (public endpoint for users)
func (h *RequestHandler) CreateBannerRequest(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c, "User not authenticated")
	}

	var req models.BannerRequest
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}
	if err := h.validate.Struct(req); err != nil {
		return BadRequest(c, err.Error())
	}

	req.ID = uuid.New()
	req.UserID = userID

	if err := h.svc.CreateBannerRequest(c.Context(), &req); err != nil {
		return MapError(c, h.logger, err)
	}

	return OK(c, fiber.Map{"message": "Banner request submitted successfully", "id": req.ID})
}

// Upload Banner Request Image (public endpoint for users)
func (h *RequestHandler) UploadBannerRequestImage(c *fiber.Ctx) error {
	mediaSvc, ok := c.Locals("mediaService").(services.MediaService)
	if !ok {
		return InternalError(c, "Media service not found")
	}

	fileHeader, err := c.FormFile("image")
	if err != nil {
		return BadRequest(c, "Image is required")
	}
	if fileHeader.Size > 5*1024*1024 {
		return BadRequest(c, "Image too large (max 5MB)")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return InternalError(c, "Failed to open image")
	}
	defer file.Close()

	// Folder structure: banner-requests/{userID}
	userID, _ := middleware.GetUserID(c)
	folder := fmt.Sprintf("banner-requests/%s", userID.String())

	url, err := mediaSvc.UploadFile(c.Context(), file, fileHeader, folder)
	if err != nil {
		return MapError(c, h.logger, err)
	}

	return OK(c, fiber.Map{"url": url})
}

// Get Auction Requests
func (h *RequestHandler) GetAuctionRequests(c *fiber.Ctx) error {
	status := c.Query("status", "")
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 20)
	page, perPage = ClampPagination(page, perPage)
	sortBy := c.Query("sort_by", "created_at")
	sortOrder := c.Query("sort_order", "DESC")

	// Parse filter parameters
	var userID *uuid.UUID
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		id, err := uuid.Parse(userIDStr)
		if err != nil {
			return BadRequest(c, "Invalid user_id format")
		}
		userID = &id
	}

	var dateFrom, dateTo *time.Time
	if dateFromStr := c.Query("date_from"); dateFromStr != "" {
		t, err := time.Parse(time.RFC3339, dateFromStr)
		if err != nil {
			return BadRequest(c, "Invalid date_from format (use RFC3339)")
		}
		dateFrom = &t
	}
	if dateToStr := c.Query("date_to"); dateToStr != "" {
		t, err := time.Parse(time.RFC3339, dateToStr)
		if err != nil {
			return BadRequest(c, "Invalid date_to format (use RFC3339)")
		}
		dateTo = &t
	}

	// Parse category and location filters
	var categoryID, locationID *int
	if categoryIDStr := c.Query("category_id"); categoryIDStr != "" {
		id, err := strconv.Atoi(categoryIDStr)
		if err != nil {
			return BadRequest(c, "Invalid category_id format")
		}
		categoryID = &id
	}
	if locationIDStr := c.Query("location_id"); locationIDStr != "" {
		id, err := strconv.Atoi(locationIDStr)
		if err != nil {
			return BadRequest(c, "Invalid location_id format")
		}
		locationID = &id
	}

	// Parse price range filters
	var minPrice, maxPrice *float64
	if minPriceStr := c.Query("min_price"); minPriceStr != "" {
		price, err := strconv.ParseFloat(minPriceStr, 64)
		if err != nil {
			return BadRequest(c, "Invalid min_price format")
		}
		minPrice = &price
	}
	if maxPriceStr := c.Query("max_price"); maxPriceStr != "" {
		price, err := strconv.ParseFloat(maxPriceStr, 64)
		if err != nil {
			return BadRequest(c, "Invalid max_price format")
		}
		maxPrice = &price
	}

	requests, total, err := h.svc.GetAuctionRequests(c.Context(), status, userID, dateFrom, dateTo, categoryID, locationID, minPrice, maxPrice, sortBy, sortOrder, page, perPage)
	if err != nil {
		return MapError(c, h.logger, err)
	}

	return OK(c, fiber.Map{
		"data":     requests,
		"total":    total,
		"page":     page,
		"per_page": perPage,
	})
}

// Get Auction Request by ID (detail view)
func (h *RequestHandler) GetAuctionRequestByID(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid request ID")
	}

	request, err := h.svc.GetAuctionRequestByID(c.Context(), id)
	if err != nil {
		return MapError(c, h.logger, err)
	}

	return OK(c, request)
}

// Review Auction Request
func (h *RequestHandler) ReviewAuctionRequest(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return BadRequest(c, "Invalid request ID")
	}

	var req ReviewRequest
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}
	if err := h.validate.Struct(req); err != nil {
		return BadRequest(c, err.Error())
	}
	if err := requiresNotesWhenRejected(req.Status, req.Notes); err != nil {
		return BadRequest(c, err.Error())
	}

	adminID, _ := middleware.GetUserID(c)
	if err := h.svc.ReviewAuctionRequest(c.Context(), id, req.Status, req.Notes, adminID); err != nil {
		return MapError(c, h.logger, err)
	}

	return OK(c, fiber.Map{"message": "Auction request reviewed successfully"})
}

// Update Auction Request (owner: draft or rejected only)
func (h *RequestHandler) UpdateAuctionRequest(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid request ID")
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c, "User not authenticated")
	}

	var req models.AuctionRequest
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}
	if err := h.validate.Struct(req); err != nil {
		return BadRequest(c, err.Error())
	}

	if err := h.svc.UpdateAuctionRequest(c.Context(), id, userID, &req); err != nil {
		return MapError(c, h.logger, err)
	}

	return OK(c, fiber.Map{"message": "Auction request updated successfully"})
}

// Admin Update Auction Request (bypasses ownership + status restrictions)
func (h *RequestHandler) AdminUpdateAuctionRequest(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid request ID")
	}

	var req models.AuctionRequest
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}
	if err := h.validate.Struct(req); err != nil {
		return BadRequest(c, err.Error())
	}

	if err := h.svc.AdminUpdateAuctionRequest(c.Context(), id, &req); err != nil {
		return MapError(c, h.logger, err)
	}

	return OK(c, fiber.Map{"message": "Auction request updated successfully"})
}

// Admin Create Auction Request (admin creates on behalf of themselves, draft or pending)
func (h *RequestHandler) AdminCreateAuctionRequest(c *fiber.Ctx) error {
	adminID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c, "User not authenticated")
	}

	var req models.AuctionRequest
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	// UserID is always the authenticated admin, never client-supplied -- assigned
	// before validation for the same reason as CreateAuctionRequest above.
	req.ID = uuid.New()
	req.UserID = adminID

	if err := h.validate.Struct(req); err != nil {
		return BadRequest(c, err.Error())
	}

	if req.Quantity == 0 {
		req.Quantity = 1
	}

	if err := h.svc.CreateAuctionRequest(c.Context(), &req); err != nil {
		return MapError(c, h.logger, err)
	}

	return OK(c, fiber.Map{"message": "Auction request created successfully", "id": req.ID})
}

// Delete Auction Request
func (h *RequestHandler) DeleteAuctionRequest(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return BadRequest(c, "Invalid request ID")
	}

	adminID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c, "User not authenticated")
	}

	if err := h.svc.DeleteAuctionRequest(c.Context(), id, adminID); err != nil {
		return MapError(c, h.logger, err)
	}

	return OK(c, fiber.Map{"message": "Auction request deleted successfully"})
}

// Get Banner Requests
func (h *RequestHandler) GetBannerRequests(c *fiber.Ctx) error {
	status := c.Query("status", "")
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 20)
	page, perPage = ClampPagination(page, perPage)
	sortBy := c.Query("sort_by", "created_at")
	sortOrder := c.Query("sort_order", "DESC")

	// Parse filter parameters
	var userID *uuid.UUID
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		id, err := uuid.Parse(userIDStr)
		if err != nil {
			return BadRequest(c, "Invalid user_id format")
		}
		userID = &id
	}

	var dateFrom, dateTo *time.Time
	if dateFromStr := c.Query("date_from"); dateFromStr != "" {
		t, err := time.Parse(time.RFC3339, dateFromStr)
		if err != nil {
			return BadRequest(c, "Invalid date_from format (use RFC3339)")
		}
		dateFrom = &t
	}
	if dateToStr := c.Query("date_to"); dateToStr != "" {
		t, err := time.Parse(time.RFC3339, dateToStr)
		if err != nil {
			return BadRequest(c, "Invalid date_to format (use RFC3339)")
		}
		dateTo = &t
	}

	requests, total, err := h.svc.GetBannerRequests(c.Context(), status, userID, dateFrom, dateTo, sortBy, sortOrder, page, perPage)
	if err != nil {
		return MapError(c, h.logger, err)
	}

	return OK(c, fiber.Map{
		"data":     requests,
		"total":    total,
		"page":     page,
		"per_page": perPage,
	})
}

// Get Banner Request by ID (detail view)
func (h *RequestHandler) GetBannerRequestByID(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid request ID")
	}

	request, err := h.svc.GetBannerRequestByID(c.Context(), id)
	if err != nil {
		return MapError(c, h.logger, err)
	}

	return OK(c, request)
}

// Get User's Auction Requests
func (h *RequestHandler) GetUserAuctionRequests(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c, "User not authenticated")
	}

	status := c.Query("status", "")
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 20)
	page, perPage = ClampPagination(page, perPage)

	requests, total, err := h.svc.GetUserAuctionRequests(c.Context(), userID, status, page, perPage)
	if err != nil {
		return MapError(c, h.logger, err)
	}

	return OK(c, fiber.Map{
		"data":     requests,
		"total":    total,
		"page":     page,
		"per_page": perPage,
	})
}

// Get User's Banner Requests
func (h *RequestHandler) GetUserBannerRequests(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c, "User not authenticated")
	}

	status := c.Query("status", "")
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 20)
	page, perPage = ClampPagination(page, perPage)

	requests, total, err := h.svc.GetUserBannerRequests(c.Context(), userID, status, page, perPage)
	if err != nil {
		return MapError(c, h.logger, err)
	}

	return OK(c, fiber.Map{
		"data":     requests,
		"total":    total,
		"page":     page,
		"per_page": perPage,
	})
}

// Review Banner Request
func (h *RequestHandler) ReviewBannerRequest(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return BadRequest(c, "Invalid request ID")
	}

	var req ReviewRequest
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}
	if err := h.validate.Struct(req); err != nil {
		return BadRequest(c, err.Error())
	}

	adminID, _ := middleware.GetUserID(c)
	if err := h.svc.ReviewBannerRequest(c.Context(), id, req.Status, req.Notes, adminID); err != nil {
		return MapError(c, h.logger, err)
	}

	return OK(c, fiber.Map{"message": "Banner request reviewed successfully"})
}

// Delete Banner Request
func (h *RequestHandler) DeleteBannerRequest(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return BadRequest(c, "Invalid request ID")
	}

	adminID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c, "User not authenticated")
	}

	if err := h.svc.DeleteBannerRequest(c.Context(), id, adminID); err != nil {
		return MapError(c, h.logger, err)
	}

	return OK(c, fiber.Map{"message": "Banner request deleted successfully"})
}

func (h *RequestHandler) BulkReviewAuctionRequests(c *fiber.Ctx) error {
	var req struct {
		IDs    []uuid.UUID `json:"ids" validate:"required,min=1"`
		Status string      `json:"status" validate:"required,oneof=approved rejected"`
		Notes  string      `json:"notes"`
	}

	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	if err := h.validate.Struct(&req); err != nil {
		return BadRequest(c, err.Error())
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c, "User not authenticated")
	}

	if err := h.svc.BulkReviewAuctionRequests(c.Context(), req.IDs, req.Status, req.Notes, userID); err != nil {
		return MapError(c, h.logger, err)
	}

	return OK(c, fiber.Map{
		"updated_count": len(req.IDs),
	})
}

func (h *RequestHandler) BulkDeleteAuctionRequests(c *fiber.Ctx) error {
	var req struct {
		IDs []uuid.UUID `json:"ids" validate:"required,min=1"`
	}

	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	if err := h.validate.Struct(&req); err != nil {
		return BadRequest(c, err.Error())
	}

	adminID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c, "User not authenticated")
	}

	if err := h.svc.BulkDeleteAuctionRequests(c.Context(), req.IDs, adminID); err != nil {
		return MapError(c, h.logger, err)
	}

	return OK(c, fiber.Map{
		"deleted_count": len(req.IDs),
	})
}

func (h *RequestHandler) BulkReviewBannerRequests(c *fiber.Ctx) error {
	var req struct {
		IDs    []uuid.UUID `json:"ids" validate:"required,min=1"`
		Status string      `json:"status" validate:"required,oneof=approved rejected"`
		Notes  string      `json:"notes"`
	}

	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	if err := h.validate.Struct(&req); err != nil {
		return BadRequest(c, err.Error())
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c, "User not authenticated")
	}

	if err := h.svc.BulkReviewBannerRequests(c.Context(), req.IDs, req.Status, req.Notes, userID); err != nil {
		return MapError(c, h.logger, err)
	}

	return OK(c, fiber.Map{
		"updated_count": len(req.IDs),
	})
}

func (h *RequestHandler) BulkDeleteBannerRequests(c *fiber.Ctx) error {
	var req struct {
		IDs []uuid.UUID `json:"ids" validate:"required,min=1"`
	}

	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	if err := h.validate.Struct(&req); err != nil {
		return BadRequest(c, err.Error())
	}

	adminID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c, "User not authenticated")
	}

	if err := h.svc.BulkDeleteBannerRequests(c.Context(), req.IDs, adminID); err != nil {
		return MapError(c, h.logger, err)
	}

	return OK(c, fiber.Map{
		"deleted_count": len(req.IDs),
	})
}
