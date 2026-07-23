package handlers

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/middleware"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/services"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type AdminHandler struct {
	svc       services.AdminService
	reportSvc services.ReportService
	logger    *zap.Logger
	rdb       *redis.Client
}

func NewAdminHandler(svc services.AdminService, reportSvc services.ReportService, logger *zap.Logger, rdb ...*redis.Client) *AdminHandler {
	h := &AdminHandler{svc: svc, reportSvc: reportSvc, logger: logger}
	if len(rdb) > 0 {
		h.rdb = rdb[0]
	}
	return h
}

func (h *AdminHandler) invalidateCategoriesCache(ctx context.Context) {
	if h.rdb != nil {
		h.rdb.Del(ctx, "categories")
	}
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
	page, perPage = ClampPagination(page, perPage)

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

	adminID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c, "User not authenticated")
	}

	if err := h.svc.BlockUser(c.Context(), id, req.Block, adminID); err != nil {
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

func (h *AdminHandler) DeleteUser(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid user ID")
	}

	adminID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c, "User not authenticated")
	}

	if err := h.svc.DeleteUser(c.Context(), id, adminID); err != nil {
		fmt.Printf("Error deleting user %s: %v\n", id, err)
		return InternalError(c, fmt.Sprintf("Failed to delete user: %v", err))
	}

	return OK(c, fiber.Map{
		"message": "User deleted successfully",
		"user_id": id.String(),
	})
}

// List auctions
func (h *AdminHandler) ListAuctions(c *fiber.Ctx) error {
	status := c.Query("status")
	q := c.Query("q")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "25"))
	page, perPage = ClampPagination(page, perPage)

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

	adminID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c, "User not authenticated")
	}

	if err := h.svc.ValidateAuction(c.Context(), id, req.Approve, req.Reason, adminID); err != nil {
		return MapError(c, h.logger, err)
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
		CategoryID      *int                   `json:"category_id"`
		SubCategoryID   *int                   `json:"sub_category_id"`
		LocationID      *int                   `json:"location_id"`
		TitleAr         *string                `json:"title_ar"    validate:"omitempty,min=3,max=200"`
		TitleFr         *string                `json:"title_fr"`
		TitleEn         *string                `json:"title_en"`
		DescriptionAr   *string                `json:"description_ar"`
		DescriptionFr   *string                `json:"description_fr"`
		DescriptionEn   *string                `json:"description_en"`
		StartPrice      *float64               `json:"start_price"`
		MinIncrement    *float64               `json:"min_increment"`
		InsuranceAmount *float64               `json:"insurance_amount"`
		StartTime       *string                `json:"start_time"`
		EndTime         string                 `json:"end_time"`
		PhoneContact    *string                `json:"phone_contact"`
		BuyNowPrice     *float64               `json:"buy_now_price"`
		ItemDetails     map[string]interface{} `json:"item_details"`
		Images          []string               `json:"images"`
		Condition       *string                `json:"condition" validate:"omitempty,oneof=new used refurbished damaged"`
		Brand           *string                `json:"brand"`
		VideoURL        *string                `json:"video_url"`
		Quantity        *int                   `json:"quantity" validate:"omitempty,min=1"`
	}

	var req UpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	// end_time is optional — zero value means "keep existing" in service
	var endTime time.Time
	if req.EndTime != "" {
		var parseErr error
		endTime, parseErr = time.Parse(time.RFC3339, req.EndTime)
		if parseErr != nil {
			return BadRequest(c, "Invalid end_time format, use RFC3339")
		}
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

	var itemDetails models.JSONB
	if req.ItemDetails != nil {
		itemDetails = models.JSONB(req.ItemDetails)
	}

	// Set defaults
	titleAr := ""
	if req.TitleAr != nil {
		titleAr = *req.TitleAr
	}
	titleFr := ""
	if req.TitleFr != nil {
		titleFr = *req.TitleFr
	}
	titleEn := ""
	if req.TitleEn != nil {
		titleEn = *req.TitleEn
	}
	descriptionAr := ""
	if req.DescriptionAr != nil {
		descriptionAr = *req.DescriptionAr
	}
	descriptionFr := ""
	if req.DescriptionFr != nil {
		descriptionFr = *req.DescriptionFr
	}
	descriptionEn := ""
	if req.DescriptionEn != nil {
		descriptionEn = *req.DescriptionEn
	}
	phoneContact := ""
	if req.PhoneContact != nil {
		phoneContact = *req.PhoneContact
	}
	startPrice := decimal.Zero
	if req.StartPrice != nil {
		startPrice = decimal.NewFromFloat(*req.StartPrice)
	}
	minIncrement := decimal.Zero
	if req.MinIncrement != nil {
		minIncrement = decimal.NewFromFloat(*req.MinIncrement)
	}
	insuranceAmount := decimal.Zero
	if req.InsuranceAmount != nil {
		insuranceAmount = decimal.NewFromFloat(*req.InsuranceAmount)
	}
	quantity := 1
	if req.Quantity != nil {
		quantity = *req.Quantity
	}

	// Get category_id from request (required field)
	categoryID := 0
	if req.CategoryID != nil {
		categoryID = *req.CategoryID
	}
	if categoryID == 0 {
		return BadRequest(c, "category_id is required")
	}

	if err := h.svc.UpdateAuction(c.Context(), id, services.UpdateAuctionInput{
		CategoryID:      categoryID,
		SubCategoryID:   req.SubCategoryID,
		LocationID:      req.LocationID,
		TitleAr:         titleAr,
		TitleFr:         titleFr,
		TitleEn:         titleEn,
		DescriptionAr:   descriptionAr,
		DescriptionFr:   descriptionFr,
		DescriptionEn:   descriptionEn,
		StartPrice:      startPrice,
		MinIncrement:    minIncrement,
		InsuranceAmount: insuranceAmount,
		StartTime:       startTime,
		EndTime:         endTime,
		PhoneContact:    phoneContact,
		BuyNowPrice:     buyNow,
		Images:          req.Images,
		ItemDetails:     itemDetails,
		Condition:       req.Condition,
		Brand:           req.Brand,
		VideoURL:        req.VideoURL,
		Quantity:        quantity,
	}); err != nil {
		h.logger.Error("UpdateAuction failed", zap.Error(err), zap.String("auction_id", id.String()))
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
	page, perPage = ClampPagination(page, perPage)

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
	page, perPage = ClampPagination(page, perPage)

	reportType := c.Query("type")
	reports, total, err := h.svc.ListReports(c.Context(), page, perPage, status, reportType)
	if err != nil {
		return MapError(c, h.logger, err)
	}

	return PaginatedOK(c, reports, fiber.Map{
		"page":     page,
		"per_page": perPage,
		"total":    total,
	})
}

// GetTransaction - GET /admin/transactions/:id
func (h *AdminHandler) GetTransaction(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid transaction ID")
	}
	txn, err := h.svc.GetTransactionByID(c.Context(), id)
	if err != nil {
		return NotFound(c, "Transaction")
	}
	return OK(c, txn)
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
}

func (h *AdminHandler) DeleteReport(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid report ID")
	}

	if err := h.svc.DeleteReport(c.Context(), id); err != nil {
		return MapError(c, h.logger, err)
	}

	return OK(c, fiber.Map{"message": "Report deleted", "id": id.String()})
}

// KYC management
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
	h.invalidateCategoriesCache(c.Context())
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
	h.invalidateCategoriesCache(c.Context())
	return OK(c, cat)
}

func (h *AdminHandler) DeleteCategory(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	if err := h.svc.DeleteCategory(c.Context(), id); err != nil {
		return InternalError(c, "Failed to delete category: "+err.Error())
	}
	h.invalidateCategoriesCache(c.Context())
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

func (h *AdminHandler) ListBlockedPhones(c *fiber.Ctx) error {
	phones, err := h.svc.ListBlockedPhones(c.Context())
	if err != nil {
		return InternalError(c, "Failed to list blocked phones")
	}
	return OK(c, phones)
}

func (h *AdminHandler) BlockPhone(c *fiber.Ctx) error {
	type Request struct {
		Phone  string `json:"phone"`
		Reason string `json:"reason"`
	}
	var req Request
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}
	adminID, _ := middleware.GetUserID(c)
	if err := h.svc.BlockPhone(c.Context(), req.Phone, req.Reason, adminID); err != nil {
		return InternalError(c, "Failed to block phone")
	}
	return OK(c, fiber.Map{"message": "Phone blocked successfully"})
}

func (h *AdminHandler) UnblockPhone(c *fiber.Ctx) error {
	phone := c.Params("phone")
	// Decode URL encoded phone number
	decodedPhone, err := url.QueryUnescape(phone)
	if err != nil {
		h.logger.Error("Failed to decode phone parameter", zap.String("phone", phone), zap.Error(err))
		return BadRequest(c, "Invalid phone parameter")
	}

	h.logger.Info("Attempting to unblock phone", zap.String("phone", phone), zap.String("decodedPhone", decodedPhone))

	if err := h.svc.UnblockPhone(c.Context(), decodedPhone); err != nil {
		h.logger.Error("Failed to unblock phone", zap.String("phone", phone), zap.String("decodedPhone", decodedPhone), zap.Error(err))
		return InternalError(c, "Failed to unblock phone")
	}

	h.logger.Info("Successfully unblocked phone", zap.String("phone", phone), zap.String("decodedPhone", decodedPhone))
	return OK(c, fiber.Map{"message": "Phone unblocked successfully"})
}

func (h *AdminHandler) ListSettings(c *fiber.Ctx) error {
	settings, err := h.svc.ListSettings(c.Context())
	if err != nil {
		return InternalError(c, "Failed to list settings")
	}
	return OK(c, settings)
}

func (h *AdminHandler) UpdateSetting(c *fiber.Ctx) error {
	type Request struct {
		Value string `json:"value"`
		Type  string `json:"type"`
	}
	var req Request
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}
	key := c.Params("key")
	adminID, _ := middleware.GetUserID(c)
	if err := h.svc.UpdateSetting(c.Context(), key, req.Value, req.Type, adminID); err != nil {
		return InternalError(c, "Failed to update setting")
	}
	return OK(c, fiber.Map{"message": "Setting updated successfully"})
}

// Countries management
func (h *AdminHandler) ListCountries(c *fiber.Ctx) error {
	countries, err := h.svc.GetCountries(c.Context())
	if err != nil {
		return MapError(c, h.logger, err)
	}

	if countries == nil {
		countries = []models.Country{}
	}

	return OK(c, countries)
}

func (h *AdminHandler) CreateCountry(c *fiber.Ctx) error {
	type Request struct {
		Code        string `json:"code"        validate:"required,len=2"`
		CountryCode string `json:"country_code"` // e.g., +222, +221
		NameAr      string `json:"name_ar"     validate:"required"`
		NameFr      string `json:"name_fr"     validate:"required"`
		NameEn      string `json:"name_en"     validate:"required"`
		FlagEmoji   string `json:"flag_emoji"  validate:"required"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	if err := h.svc.CreateCountry(c.Context(), req.Code, req.CountryCode, req.NameAr, req.NameFr, req.NameEn, req.FlagEmoji); err != nil {
		return InternalError(c, "Failed to create country")
	}

	return OK(c, fiber.Map{"message": "Country created successfully"})
}

func (h *AdminHandler) UpdateCountry(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id", 0)
	if err != nil {
		return BadRequest(c, "Invalid country ID")
	}

	type Request struct {
		Code        string `json:"code"        validate:"required,len=2"`
		CountryCode string `json:"country_code"` // e.g., +222, +221
		NameAr      string `json:"name_ar"     validate:"required"`
		NameFr      string `json:"name_fr"     validate:"required"`
		NameEn      string `json:"name_en"     validate:"required"`
		FlagEmoji   string `json:"flag_emoji"  validate:"required"`
		IsActive    *bool  `json:"is_active"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	if err := h.svc.UpdateCountry(c.Context(), id, req.Code, req.CountryCode, req.NameAr, req.NameFr, req.NameEn, req.FlagEmoji, req.IsActive); err != nil {
		return InternalError(c, "Failed to update country")
	}

	return OK(c, fiber.Map{"message": "Country updated successfully"})
}

func (h *AdminHandler) DeleteCountry(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id", 0)
	if err != nil {
		return BadRequest(c, "Invalid country ID")
	}

	if err := h.svc.DeleteCountry(c.Context(), id); err != nil {
		return InternalError(c, "Failed to delete country")
	}

	return OK(c, fiber.Map{"message": "Country deleted successfully"})
}

// Payment Methods management (from migration 000031)
func (h *AdminHandler) ListPaymentMethods(c *fiber.Ctx) error {
	methods, err := h.svc.ListPaymentMethods(c.Context())
	if err != nil {
		return InternalError(c, "Failed to list payment methods")
	}
	return OK(c, methods)
}

func (h *AdminHandler) CreatePaymentMethod(c *fiber.Ctx) error {
	type Request struct {
		Code      string  `json:"code"       validate:"required"`
		NameAr    string  `json:"name_ar"    validate:"required"`
		NameFr    string  `json:"name_fr"    validate:"required"`
		NameEn    *string `json:"name_en"`
		LogoURL   *string `json:"logo_url"`
		IsActive  *bool   `json:"is_active"`
		CountryID *int    `json:"country_id"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	if err := h.svc.CreatePaymentMethod(c.Context(), req.Code, req.NameAr, req.NameFr, req.NameEn, req.LogoURL, req.IsActive, req.CountryID); err != nil {
		return InternalError(c, "Failed to create payment method: "+err.Error())
	}

	return OK(c, fiber.Map{"message": "Payment method created successfully"})
}

func (h *AdminHandler) UpdatePaymentMethod(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id", 0)
	if err != nil {
		return BadRequest(c, "Invalid payment method ID")
	}

	type Request struct {
		Code      string  `json:"code"       validate:"required"`
		NameAr    string  `json:"name_ar"    validate:"required"`
		NameFr    string  `json:"name_fr"    validate:"required"`
		NameEn    *string `json:"name_en"`
		LogoURL   *string `json:"logo_url"`
		IsActive  *bool   `json:"is_active"`
		CountryID *int    `json:"country_id"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	if err := h.svc.UpdatePaymentMethod(c.Context(), id, req.Code, req.NameAr, req.NameFr, req.NameEn, req.LogoURL, req.IsActive, req.CountryID); err != nil {
		return InternalError(c, "Failed to update payment method")
	}

	return OK(c, fiber.Map{"message": "Payment method updated successfully"})
}

func (h *AdminHandler) DeletePaymentMethod(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id", 0)
	if err != nil {
		return BadRequest(c, "Invalid payment method ID")
	}

	if err := h.svc.DeletePaymentMethod(c.Context(), id); err != nil {
		return InternalError(c, "Failed to delete payment method")
	}

	return OK(c, fiber.Map{"message": "Payment method deleted successfully"})
}

// Auction Car Details management (from migration 000031)
func (h *AdminHandler) GetAuctionCarDetails(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid auction ID")
	}

	details, err := h.svc.GetAuctionCarDetails(c.Context(), id)
	if err != nil {
		return InternalError(c, "Failed to get car details")
	}

	return OK(c, details)
}

func (h *AdminHandler) UpdateAuctionCarDetails(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid auction ID")
	}

	type Request struct {
		Manufacturer *string `json:"manufacturer"`
		Model        *string `json:"model"`
		Year         *int    `json:"year"`
		Mileage      *int    `json:"mileage"`
		FuelType     *string `json:"fuel_type"`
		Transmission *string `json:"transmission"`
		Color        *string `json:"color"`
		EngineSize   *string `json:"engine_size"`
		VIN          *string `json:"vin"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	if err := h.svc.UpdateAuctionCarDetails(c.Context(), id, req.Manufacturer, req.Model, req.Year, req.Mileage, req.FuelType, req.Transmission, req.Color, req.EngineSize, req.VIN); err != nil {
		return InternalError(c, "Failed to update car details")
	}

	return OK(c, fiber.Map{"message": "Car details updated successfully"})
}

// Auction Boost management (from migration 000031)
func (h *AdminHandler) ListAuctionBoosts(c *fiber.Ctx) error {
	boosts, err := h.svc.ListAuctionBoosts(c.Context())
	if err != nil {
		return InternalError(c, "Failed to list auction boosts")
	}
	return OK(c, boosts)
}

func (h *AdminHandler) CreateAuctionBoost(c *fiber.Ctx) error {
	type Request struct {
		AuctionID uuid.UUID `json:"auction_id" validate:"required"`
		BoostType string    `json:"boost_type" validate:"required,oneof=featured urgent top"`
		StartAt   string    `json:"start_at"   validate:"required"`
		EndAt     string    `json:"end_at"     validate:"required"`
		Amount    *float64  `json:"amount"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	startAt, err := time.Parse(time.RFC3339, req.StartAt)
	if err != nil {
		return BadRequest(c, "Invalid start_at format")
	}

	endAt, err := time.Parse(time.RFC3339, req.EndAt)
	if err != nil {
		return BadRequest(c, "Invalid end_at format")
	}

	var amount *decimal.Decimal
	if req.Amount != nil {
		a := decimal.NewFromFloat(*req.Amount)
		amount = &a
	}

	if err := h.svc.CreateAuctionBoost(c.Context(), req.AuctionID, req.BoostType, startAt, endAt, amount); err != nil {
		return InternalError(c, "Failed to create auction boost")
	}

	return OK(c, fiber.Map{"message": "Auction boost created successfully"})
}

func (h *AdminHandler) DeleteAuctionBoost(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid boost ID")
	}

	if err := h.svc.DeleteAuctionBoost(c.Context(), id); err != nil {
		return InternalError(c, "Failed to delete auction boost")
	}

	return OK(c, fiber.Map{"message": "Auction boost deleted successfully"})
}

// Delivery Driver management (from migration 000031)
func (h *AdminHandler) ListDeliveryDrivers(c *fiber.Ctx) error {
	drivers, err := h.svc.ListDeliveryDrivers(c.Context())
	if err != nil {
		return InternalError(c, "Failed to list delivery drivers")
	}
	return OK(c, drivers)
}

func (h *AdminHandler) CreateDeliveryDriver(c *fiber.Ctx) error {
	type Request struct {
		UserID        *uuid.UUID `json:"user_id"`
		VehicleType   *string    `json:"vehicle_type"`
		VehiclePlate  *string    `json:"vehicle_plate"`
		VehicleColor  *string    `json:"vehicle_color"`
		LicenseNumber *string    `json:"license_number"`
		IsAvailable   *bool      `json:"is_available"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	if err := h.svc.CreateDeliveryDriver(c.Context(), req.UserID, req.VehicleType, req.VehiclePlate, req.VehicleColor, req.LicenseNumber, req.IsAvailable); err != nil {
		return InternalError(c, "Failed to create delivery driver")
	}

	return OK(c, fiber.Map{"message": "Delivery driver created successfully"})
}

func (h *AdminHandler) UpdateDeliveryDriver(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid driver ID")
	}

	type Request struct {
		VehicleType   *string `json:"vehicle_type"`
		VehiclePlate  *string `json:"vehicle_plate"`
		VehicleColor  *string `json:"vehicle_color"`
		LicenseNumber *string `json:"license_number"`
		IsAvailable   *bool   `json:"is_available"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	if err := h.svc.UpdateDeliveryDriver(c.Context(), id, req.VehicleType, req.VehiclePlate, req.VehicleColor, req.LicenseNumber, req.IsAvailable); err != nil {
		return InternalError(c, "Failed to update delivery driver")
	}

	return OK(c, fiber.Map{"message": "Delivery driver updated successfully"})
}

func (h *AdminHandler) DeleteDeliveryDriver(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid driver ID")
	}

	if err := h.svc.DeleteDeliveryDriver(c.Context(), id); err != nil {
		return InternalError(c, "Failed to delete delivery driver")
	}

	return OK(c, fiber.Map{"message": "Delivery driver deleted successfully"})
}

// User Settings management (from migration 000031)
func (h *AdminHandler) GetUserSettings(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid user ID")
	}

	settings, err := h.svc.GetUserSettings(c.Context(), id)
	if err != nil {
		return InternalError(c, "Failed to get user settings")
	}

	return OK(c, settings)
}

func (h *AdminHandler) UpdateUserSettings(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid user ID")
	}

	type Request struct {
		Currency           *string `json:"currency"`
		Theme              *string `json:"theme"`
		Language           *string `json:"language"`
		NotificationsEmail *bool   `json:"notifications_email"`
		NotificationsPush  *bool   `json:"notifications_push"`
		NotificationsSMS   *bool   `json:"notifications_sms"`
		TwoFactorEnabled   *bool   `json:"two_factor_enabled"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	if err := h.svc.UpdateUserSettings(c.Context(), id, req.Currency, req.Theme, req.Language, req.NotificationsEmail, req.NotificationsPush, req.NotificationsSMS, req.TwoFactorEnabled); err != nil {
		return InternalError(c, "Failed to update user settings")
	}

	return OK(c, fiber.Map{"message": "User settings updated successfully"})
}

// Bid Auto Bid management (from migration 000031)
func (h *AdminHandler) ListAutoBids(c *fiber.Ctx) error {
	bids, err := h.svc.ListAutoBids(c.Context())
	if err != nil {
		return InternalError(c, "Failed to list auto bids")
	}
	return OK(c, bids)
}

func (h *AdminHandler) UpdateAutoBid(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid auto bid ID")
	}

	type Request struct {
		IsActive *bool `json:"is_active"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	if err := h.svc.UpdateAutoBid(c.Context(), id, req.IsActive); err != nil {
		return InternalError(c, "Failed to update auto bid")
	}

	return OK(c, fiber.Map{"message": "Auto bid updated successfully"})
}

// ExportTransactionsCSV exports transactions to CSV
func (h *AdminHandler) ExportTransactionsCSV(c *fiber.Ctx) error {
	status := c.Query("status")
	
	// Parse dates if provided
	var startDate, endDate *time.Time
	if s := c.Query("start_date"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			startDate = &t
		}
	}
	if s := c.Query("end_date"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			endDate = &t
		}
	}

	csvData, err := h.reportSvc.ExportTransactionsCSV(c.Context(), status, startDate, endDate)
	if err != nil {
		return InternalError(c, "Failed to export transactions: "+err.Error())
	}

	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=transactions_%s.csv", time.Now().Format("2006-01-02")))
	
	return c.SendString(csvData)
}

// ExportRevenueCSV exports revenue stats to CSV
func (h *AdminHandler) ExportRevenueCSV(c *fiber.Ctx) error {
	csvData, err := h.reportSvc.ExportRevenueCSV(c.Context())
	if err != nil {
		return InternalError(c, "Failed to export revenue: "+err.Error())
	}

	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=revenue_%s.csv", time.Now().Format("2006-01-02")))
	
	return c.SendString(csvData)
}

// UploadAuctionImages uploads images for an auction (admin bypass — no seller ownership check)
func (h *AdminHandler) UploadAuctionImages(c *fiber.Ctx) error {
	auctionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid auction ID")
	}

	mediaSvc, ok := c.Locals("mediaService").(services.MediaService)
	if !ok {
		h.logger.Error("[AdminUploadAuctionImages] Media service not available")
		return InternalError(c, "Media service not available")
	}

	form, err := c.MultipartForm()
	if err != nil {
		h.logger.Error("[AdminUploadAuctionImages] Failed to parse multipart form", zap.Error(err))
		return BadRequest(c, "Failed to parse form data: "+err.Error())
	}

	files := form.File["images"]
	if len(files) == 0 {
		return BadRequest(c, "No images provided")
	}

	// Upload each image individually using UploadFile
	var urls []string
	for _, file := range files {
		if file.Size > 10*1024*1024 {
			return BadRequest(c, "File too large: "+file.Filename+" (max 10MB)")
		}
		f, err := file.Open()
		if err != nil {
			return InternalError(c, "Failed to open file: "+file.Filename)
		}
		folder := "auctions/" + auctionID.String()
		url, uploadErr := mediaSvc.UploadFile(c.Context(), f, file, folder)
		f.Close()
		if uploadErr != nil {
			h.logger.Error("[AdminUploadAuctionImages] Upload failed", zap.String("file", file.Filename), zap.Error(uploadErr))
			return InternalError(c, "Failed to upload: "+file.Filename+": "+uploadErr.Error())
		}
		urls = append(urls, url)
	}

	if len(urls) == 0 {
		return InternalError(c, "No images uploaded")
	}
	if err := h.svc.AdminAddAuctionImages(c.Context(), auctionID, urls); err != nil {
		h.logger.Error("[AdminUploadAuctionImages] Failed to save image URLs", zap.Error(err))
		return InternalError(c, "Failed to save image URLs: "+err.Error())
	}

	h.logger.Info("[AdminUploadAuctionImages] Success",
		zap.String("auction_id", auctionID.String()),
		zap.Int("count", len(urls)),
		zap.Strings("urls", urls))

	return OK(c, fiber.Map{
		"message": "Images uploaded successfully",
		"urls":    urls,
		"count":   len(urls),
	})
}

// UploadCategoryImage uploads a category image to R2 (categories/ folder)
func (h *AdminHandler) UploadCategoryImage(c *fiber.Ctx) error {
	h.logger.Info("[UploadCategoryImage] Starting category image upload")

	// Get media service from context
	mediaSvc, ok := c.Locals("mediaService").(services.MediaService)
	if !ok {
		h.logger.Error("[UploadCategoryImage] Media service not available")
		return InternalError(c, "Media service not available")
	}

	// Parse multipart form
	file, err := c.FormFile("file")
	if err != nil {
		h.logger.Error("[UploadCategoryImage] Failed to get file", zap.Error(err))
		return BadRequest(c, "No file provided")
	}

	// Validate file size (max 10MB)
	if file.Size > 10*1024*1024 {
		h.logger.Warn("[UploadCategoryImage] File too large", zap.Int64("size", file.Size))
		return BadRequest(c, "File too large (max 10MB)")
	}

	// Validate file type (only images)
	ext := filepath.Ext(file.Filename)
	allowedExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".gif": true,
	}
	if !allowedExts[ext] {
		h.logger.Warn("[UploadCategoryImage] Invalid file type", zap.String("ext", ext))
		return BadRequest(c, "Invalid file type (only jpg, jpeg, png, webp, gif allowed)")
	}

	// Open file
	fileReader, err := file.Open()
	if err != nil {
		h.logger.Error("[UploadCategoryImage] Failed to open file", zap.Error(err))
		return InternalError(c, "Failed to open file")
	}
	defer fileReader.Close()

	// Upload to R2 (categories/ folder)
	url, err := mediaSvc.UploadFile(c.Context(), fileReader, file, "categories")
	if err != nil {
		h.logger.Error("[UploadCategoryImage] R2 upload failed", zap.Error(err))
		return InternalError(c, "Failed to upload category image: "+err.Error())
	}

	h.logger.Info("[UploadCategoryImage] Upload successful", zap.String("url", url))

	return OK(c, fiber.Map{
		"message": "Category image uploaded successfully",
		"url":     url,
		"type":    "image",
		"size":    file.Size,
		"name":    file.Filename,
	})
}
