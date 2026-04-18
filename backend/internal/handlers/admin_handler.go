package handlers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/middleware"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/services"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type AdminHandler struct {
	svc    services.AdminService
	logger *zap.Logger
}

func NewAdminHandler(svc services.AdminService, logger *zap.Logger) *AdminHandler {
	return &AdminHandler{svc: svc, logger: logger}
}

// Dashboard stats
func (h *AdminHandler) DashboardStats(c *fiber.Ctx) error {
	stats, err := h.svc.GetDashboardStats(c.Context())
	if err != nil {
		return InternalError(c, "Failed to get stats: "+err.Error())
	}

	return OK(c, stats)
}

// Revenue chart
func (h *AdminHandler) RevenueChart(c *fiber.Ctx) error {
	data, err := h.svc.GetRevenueChartData(c.Context())
	if err != nil {
		return InternalError(c, "Failed to get chart data: "+err.Error())
	}

	return OK(c, data)
}

// Activity feed
func (h *AdminHandler) ActivityFeed(c *fiber.Ctx) error {
	activities, err := h.svc.GetRecentActivity(c.Context())
	if err != nil {
		return InternalError(c, "Failed to get activity feed: "+err.Error())
	}

	return OK(c, activities)
}

// CreateAdmin creates a new admin user
func (h *AdminHandler) CreateAdmin(c *fiber.Ctx) error {
	type Request struct {
		Phone    string `json:"phone"     validate:"required"`
		Pin      string `json:"pin"       validate:"required,len=4"`
		FullName string `json:"full_name" validate:"required"`
		Email    string `json:"email"     validate:"required,email"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	if err := h.svc.CreateAdmin(c.Context(), req.Phone, req.Pin, req.FullName, req.Email); err != nil {
		return InternalError(c, "Failed to create admin: "+err.Error())
	}

	return OK(c, fiber.Map{"message": "Admin user created successfully"})
}

// GenerateInvitation generates a new admin invitation link
func (h *AdminHandler) GenerateInvitation(c *fiber.Ctx) error {
	adminID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c)
	}
	token, err := h.svc.GenerateAdminInvitation(c.Context(), adminID)
	if err != nil {
		return InternalError(c, "Failed to generate invitation: "+err.Error())
	}

	return OK(c, fiber.Map{"token": token})
}

// RegisterWithInvitation registers a new admin using an invitation token
func (h *AdminHandler) RegisterWithInvitation(c *fiber.Ctx) error {
	type Request struct {
		Token    string `json:"token"     validate:"required"`
		Phone    string `json:"phone"     validate:"required"`
		Pin      string `json:"pin"       validate:"required,len=4"`
		FullName string `json:"full_name" validate:"required"`
		Email    string `json:"email"     validate:"required,email"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	if err := h.svc.RegisterAdminWithToken(c.Context(), req.Token, req.Phone, req.Pin, req.FullName, req.Email); err != nil {
		return InternalError(c, "Failed to register with invitation: "+err.Error())
	}

	return OK(c, fiber.Map{"message": "Admin registered successfully"})
}

// List users
func (h *AdminHandler) ListUsers(c *fiber.Ctx) error {
	q := c.Query("q")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "25"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 25
	}

	users, total, err := h.svc.ListUsers(c.Context(), page, perPage, q)
	if err != nil {
		return MapError(c, h.logger, err)
	}

	return PaginatedOK(c, users, fiber.Map{
		"page":     page,
		"per_page": perPage,
		"total":    total,
	})
}

// Get user by ID
func (h *AdminHandler) GetUserByID(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid user ID")
	}

	user, err := h.svc.GetUserByID(c.Context(), id)
	if err != nil {
		return NotFound(c, "User not found")
	}

	return OK(c, user)
}

// Get user auctions
func (h *AdminHandler) GetUserAuctions(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid user ID")
	}

	auctions, total, err := h.svc.ListAuctions(c.Context(), 1, 100, "", "", &id)
	if err != nil {
		return MapError(c, h.logger, err)
	}

	return PaginatedOK(c, auctions, fiber.Map{
		"page":     1,
		"per_page": 100,
		"total":    total,
	})
}

// Get user transactions
func (h *AdminHandler) GetUserTransactions(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid user ID")
	}

	transactions, total, err := h.svc.ListTransactions(c.Context(), 1, 100, "", &id)
	if err != nil {
		return MapError(c, h.logger, err)
	}

	return PaginatedOK(c, transactions, fiber.Map{
		"page":     1,
		"per_page": 100,
		"total":    total,
	})
}

// Block/Unblock user
func (h *AdminHandler) BlockUser(c *fiber.Ctx) error {
	type BlockRequest struct {
		Block bool `json:"block"`
	}

	var req BlockRequest
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid user ID")
	}

	if err := h.svc.BlockUser(c.Context(), id, req.Block); err != nil {
		return InternalError(c, "Failed to update user status")
	}

	status := "unblocked"
	if req.Block {
		status = "blocked"
	}

	return OK(c, fiber.Map{
		"message": "User " + status,
		"user_id": id.String(),
		"status":  status,
	})
}

// List auctions
func (h *AdminHandler) ListAuctions(c *fiber.Ctx) error {
	status := c.Query("status")
	q := c.Query("q")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "25"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 25
	}

	auctions, total, err := h.svc.ListAuctions(c.Context(), page, perPage, status, q, nil)
	if err != nil {
		return MapError(c, h.logger, err)
	}

	return PaginatedOK(c, auctions, fiber.Map{
		"page":     page,
		"per_page": perPage,
		"total":    total,
	})
}

// Validate/Approve auction
func (h *AdminHandler) ValidateAuction(c *fiber.Ctx) error {
	type ValidateRequest struct {
		Approve bool   `json:"approve"`
		Reason  string `json:"reason"`
	}

	var req ValidateRequest
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid auction ID")
	}

	if err := h.svc.ValidateAuction(c.Context(), id, req.Approve, req.Reason); err != nil {
		return InternalError(c, "Failed to validate auction")
	}

	status := "rejected"
	if req.Approve {
		status = "approved"
	}

	return OK(c, fiber.Map{
		"message":    "Auction " + status,
		"auction_id": id.String(),
		"status":     status,
	})
}

// Update auction (Admin view)
func (h *AdminHandler) UpdateAuction(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid auction ID")
	}

	type UpdateRequest struct {
		CategoryID      int                    `json:"category_id"`
		LocationID      *int                   `json:"location_id"`
		TitleAr         string                 `json:"title_ar"    validate:"required,min=3,max=200"`
		TitleFr         string                 `json:"title_fr"`
		TitleEn         string                 `json:"title_en"`
		DescriptionAr   string                 `json:"description_ar"`
		DescriptionFr   string                 `json:"description_fr"`
		DescriptionEn   string                 `json:"description_en"`
		StartPrice      float64                `json:"start_price"`
		MinIncrement    float64                `json:"min_increment"`
		InsuranceAmount float64                `json:"insurance_amount"`
		StartTime       *string                `json:"start_time"`
		EndTime         string                 `json:"end_time"`
		PhoneContact    string                 `json:"phone_contact"`
		BuyNowPrice     *float64               `json:"buy_now_price"`
		ItemDetails     map[string]interface{} `json:"item_details"`
		Images          []string               `json:"images"`
	}

	var req UpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		return BadRequest(c, "Invalid end_time format, use RFC3339")
	}

	var startTime *time.Time
	if req.StartTime != nil && *req.StartTime != "" {
		st, err := time.Parse(time.RFC3339, *req.StartTime)
		if err != nil {
			return BadRequest(c, "Invalid start_time format, use RFC3339")
		}
		startTime = &st
	}

	var buyNow *decimal.Decimal
	if req.BuyNowPrice != nil {
		b := decimal.NewFromFloat(*req.BuyNowPrice)
		buyNow = &b
	}

	if err := h.svc.UpdateAuction(c.Context(), id, services.UpdateAuctionInput{
		CategoryID:      req.CategoryID,
		LocationID:      req.LocationID,
		TitleAr:         req.TitleAr,
		TitleFr:         req.TitleFr,
		TitleEn:         req.TitleEn,
		DescriptionAr:   req.DescriptionAr,
		DescriptionFr:   req.DescriptionFr,
		DescriptionEn:   req.DescriptionEn,
		StartPrice:      decimal.NewFromFloat(req.StartPrice),
		MinIncrement:    decimal.NewFromFloat(req.MinIncrement),
		InsuranceAmount: decimal.NewFromFloat(req.InsuranceAmount),
		StartTime:       startTime,
		EndTime:         endTime,
		PhoneContact:    req.PhoneContact,
		BuyNowPrice:     buyNow,
		Images:          req.Images,
		ItemDetails:     req.ItemDetails,
	}); err != nil {
		return MapError(c, h.logger, err)
	}

	return OK(c, fiber.Map{"message": "Auction updated", "id": id.String()})
}

// Delete auction (Admin view)
func (h *AdminHandler) DeleteAuction(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid auction ID")
	}

	if err := h.svc.DeleteAuction(c.Context(), id); err != nil {
		return MapError(c, h.logger, err)
	}

	return OK(c, fiber.Map{"message": "Auction deleted", "id": id.String()})
}


// List all transactions (Admin view)
func (h *AdminHandler) ListTransactions(c *fiber.Ctx) error {
	status := c.Query("status")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "25"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 25
	}

	txs, total, err := h.svc.ListTransactions(c.Context(), page, perPage, status, nil)
	if err != nil {
		return MapError(c, h.logger, err)
	}

	return PaginatedOK(c, txs, fiber.Map{
		"page":     page,
		"per_page": perPage,
		"total":    total,
	})
}

// List all reports (Admin view)
func (h *AdminHandler) ListReports(c *fiber.Ctx) error {
	status := c.Query("status")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "25"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 25
	}

	reports, total, err := h.svc.ListReports(c.Context(), page, perPage, status)
	if err != nil {
		return MapError(c, h.logger, err)
	}

	return PaginatedOK(c, reports, fiber.Map{
		"page":     page,
		"per_page": perPage,
		"total":    total,
	})
}

// Validate/Approve transaction (Admin view)
func (h *AdminHandler) ValidateTransaction(c *fiber.Ctx) error {
	type ValidateRequest struct {
		Approve bool   `json:"approve"`
		Notes   string `json:"notes"`
	}

	var req ValidateRequest
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid transaction ID")
	}

	adminID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c)
	}
	if err := h.svc.ValidateTransaction(c.Context(), id, req.Approve, req.Notes, adminID); err != nil {
		return InternalError(c, "Failed to validate transaction")
	}

	status := "rejected"
	if req.Approve {
		status = "completed"
	}

	return OK(c, fiber.Map{
		"message":        "Transaction " + status,
		"transaction_id": id.String(),
		"status":         status,
	})
}

// Review/Action on report (Admin view)
func (h *AdminHandler) ReviewReport(c *fiber.Ctx) error {
	type ReviewRequest struct {
		Status string `json:"status"` // e.g., "resolved", "dismissed"
		Notes  string `json:"notes"`
	}

	var req ReviewRequest
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid report ID")
	}

	adminID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c)
	}
	if err := h.svc.ReviewReport(c.Context(), id, req.Status, req.Notes, adminID); err != nil {
		return InternalError(c, "Failed to review report")
	}

	return OK(c, fiber.Map{
		"message":   "Report reviewed",
		"report_id": id.String(),
		"status":    req.Status,
	})
}// KYC management
func (h *AdminHandler) ListKYCs(c *fiber.Ctx) error {
	status := c.Query("status")
	kycs, err := h.svc.ListKYC(c.Context(), status)
	if err != nil {
		return InternalError(c, "Failed to list KYC")
	}
	return OK(c, kycs)
}

func (h *AdminHandler) ReviewKYC(c *fiber.Ctx) error {
	type ReviewRequest struct {
		Status string `json:"status"` // approved, rejected
		Notes  string `json:"notes"`
	}
	var req ReviewRequest
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	userID, err := uuid.Parse(c.Params("user_id"))
	if err != nil {
		return BadRequest(c, "Invalid user ID")
	}

	adminID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c)
	}
	if err := h.svc.ReviewKYC(c.Context(), userID, req.Status, req.Notes, adminID); err != nil {
		return InternalError(c, "Failed to review KYC: "+err.Error())
	}

	return OK(c, fiber.Map{"message": "KYC review completed"})
}

// Category management
func (h *AdminHandler) CreateCategory(c *fiber.Ctx) error {
	var cat models.Category
	if err := c.BodyParser(&cat); err != nil {
		return BadRequest(c, "Invalid request body")
	}
	if err := h.svc.CreateCategory(c.Context(), &cat); err != nil {
		return InternalError(c, "Failed to create category: "+err.Error())
	}
	return Created(c, cat)
}

func (h *AdminHandler) UpdateCategory(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	var cat models.Category
	if err := c.BodyParser(&cat); err != nil {
		return BadRequest(c, "Invalid request body")
	}
	cat.ID = id
	if err := h.svc.UpdateCategory(c.Context(), &cat); err != nil {
		return InternalError(c, "Failed to update category: "+err.Error())
	}
	return OK(c, cat)
}

func (h *AdminHandler) DeleteCategory(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	if err := h.svc.DeleteCategory(c.Context(), id); err != nil {
		return InternalError(c, "Failed to delete category: "+err.Error())
	}
	return OK(c, fiber.Map{"message": "Category deleted"})
}

// Location management
func (h *AdminHandler) CreateLocation(c *fiber.Ctx) error {
	var loc models.Location
	if err := c.BodyParser(&loc); err != nil {
		return BadRequest(c, "Invalid request body")
	}
	if err := h.svc.CreateLocation(c.Context(), &loc); err != nil {
		return InternalError(c, "Failed to create location: "+err.Error())
	}
	return Created(c, loc)
}

func (h *AdminHandler) UpdateLocation(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	var loc models.Location
	if err := c.BodyParser(&loc); err != nil {
		return BadRequest(c, "Invalid request body")
	}
	loc.ID = id
	if err := h.svc.UpdateLocation(c.Context(), &loc); err != nil {
		return InternalError(c, "Failed to update location: "+err.Error())
	}
	return OK(c, loc)
}

func (h *AdminHandler) DeleteLocation(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	if err := h.svc.DeleteLocation(c.Context(), id); err != nil {
		return InternalError(c, "Failed to delete location: "+err.Error())
	}
	return OK(c, fiber.Map{"message": "Location deleted"})
}
