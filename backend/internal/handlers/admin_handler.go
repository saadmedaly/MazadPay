package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AdminHandler struct {
	logger *zap.Logger
}

func NewAdminHandler(logger *zap.Logger) *AdminHandler {
	return &AdminHandler{logger: logger}
}

// Dashboard stats
func (h *AdminHandler) DashboardStats(c *fiber.Ctx) error {
	stats := fiber.Map{
		"total_users":         150,
		"total_auctions":      45,
		"total_bids":          320,
		"total_revenue":       15000.50,
		"today_revenue":       1200.75,
		"active_auctions":     12,
		"pending_validations": 3,
	}

	return OK(c, fiber.Map{"data": stats})
}

// Revenue chart
func (h *AdminHandler) RevenueChart(c *fiber.Ctx) error {
	chart := fiber.Map{
		"labels": []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun"},
		"data":   []float64{1200, 1900, 1500, 2500, 2200, 3000},
	}

	return OK(c, fiber.Map{"data": chart})
}

// Activity feed
func (h *AdminHandler) ActivityFeed(c *fiber.Ctx) error {
	activities := []fiber.Map{
		{
			"id":        "1",
			"type":      "new_auction",
			"message":   "New auction created",
			"timestamp": "2024-04-16T10:30:00Z",
		},
		{
			"id":        "2",
			"type":      "new_user",
			"message":   "New user registered",
			"timestamp": "2024-04-16T09:15:00Z",
		},
	}

	return OK(c, fiber.Map{"data": activities})
}

// List users
func (h *AdminHandler) ListUsers(c *fiber.Ctx) error {
	// Query params: q (search), page, per_page
	// For now, return dummy data
	users := []fiber.Map{
		{
			"id":         uuid.New().String(),
			"phone":      "+212600000001",
			"role":       "user",
			"verified":   true,
			"created_at": "2024-04-16T10:30:00Z",
		},
	}

	return OK(c, fiber.Map{
		"data": users,
		"pagination": fiber.Map{
			"page":     1,
			"per_page": 25,
			"total":    150,
		},
	})
}

// Get user by ID
func (h *AdminHandler) GetUserByID(c *fiber.Ctx) error {
	userID := c.Params("id")

	user := fiber.Map{
		"id":             userID,
		"phone":          "+212600000001",
		"role":           "user",
		"verified":       true,
		"language_pref":  "ar",
		"auctions_count": 5,
		"bids_count":     20,
		"wallet_balance": 500.50,
		"created_at":     "2024-04-16T10:30:00Z",
	}

	return OK(c, fiber.Map{"data": user})
}

// Get user auctions
func (h *AdminHandler) GetUserAuctions(c *fiber.Ctx) error {
	// userID := c.Params("id")

	auctions := []fiber.Map{
		{
			"id":         uuid.New().String(),
			"title":      "Vintage Watch",
			"status":     "active",
			"created_at": "2024-04-16T10:30:00Z",
		},
	}

	return OK(c, fiber.Map{"data": auctions})
}

// Get user transactions
func (h *AdminHandler) GetUserTransactions(c *fiber.Ctx) error {
	// userID := c.Params("id")

	transactions := []fiber.Map{
		{
			"id":     uuid.New().String(),
			"type":   "bid",
			"amount": 100.50,
			"status": "completed",
			"date":   "2024-04-16T10:30:00Z",
		},
	}

	return OK(c, fiber.Map{"data": transactions})
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

	userID := c.Params("id")
	status := "unblocked"
	if req.Block {
		status = "blocked"
	}

	return OK(c, fiber.Map{
		"message": "User " + status,
		"user_id": userID,
		"status":  status,
	})
}

// List auctions
func (h *AdminHandler) ListAuctions(c *fiber.Ctx) error {
	// Query params: status, page, per_page
	auctions := []fiber.Map{
		{
			"id":         uuid.New().String(),
			"title":      "Vintage Watch",
			"status":     "pending",
			"seller":     "+212600000001",
			"created_at": "2024-04-16T10:30:00Z",
		},
	}

	return OK(c, fiber.Map{
		"data": auctions,
		"pagination": fiber.Map{
			"page":     1,
			"per_page": 25,
			"total":    45,
		},
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

	auctionID := c.Params("id")
	status := "rejected"
	if req.Approve {
		status = "approved"
	}

	return OK(c, fiber.Map{
		"message":    "Auction " + status,
		"auction_id": auctionID,
		"status":     status,
	})
}
