package handlers

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/middleware"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/repository"
	"github.com/mazadpay/backend/internal/services"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// BankilyHandler exposes Bankily B-PAY automatic wallet top-up. Bankily
// credentials never leave the server: the app only sends amount, the
// customer's Bankily phone and the one-time B-PAY passcode.
type BankilyHandler struct {
	svc      services.BankilyPaymentService
	userRepo repository.UserRepository
	logger   *zap.Logger
}

func NewBankilyHandler(svc services.BankilyPaymentService, userRepo repository.UserRepository, logger *zap.Logger) *BankilyHandler {
	return &BankilyHandler{svc: svc, userRepo: userRepo, logger: logger}
}

// bankilyUserView is the customer-facing shape: phone masked, no internal
// review/error fields.
type bankilyUserView struct {
	ID                   uuid.UUID       `json:"id"`
	OperationID          string          `json:"operation_id"`
	Amount               decimal.Decimal `json:"amount"`
	CurrencyCode         string          `json:"currency_code"`
	ClientPhone          string          `json:"client_phone"`
	Status               string          `json:"status"`
	BankilyTransactionID *string         `json:"bankily_transaction_id"`
	FailureReason        *string         `json:"failure_reason"`
	CreatedAt            string          `json:"created_at"`
	CompletedAt          *string         `json:"completed_at"`
}

func toBankilyUserView(p *models.BankilyPayment) bankilyUserView {
	v := bankilyUserView{
		ID:                   p.ID,
		OperationID:          p.OperationID,
		Amount:               p.Amount,
		CurrencyCode:         p.CurrencyCode,
		ClientPhone:          models.MaskPhone(p.ClientPhone),
		Status:               p.Status,
		BankilyTransactionID: p.BankilyTransactionID,
		FailureReason:        p.FailureReason,
		CreatedAt:            p.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
	if p.CompletedAt != nil {
		s := p.CompletedAt.UTC().Format("2006-01-02T15:04:05Z")
		v.CompletedAt = &s
	}
	// SUBMITTED is an internal, sub-second step: the app only distinguishes
	// processing / pending / success / failed.
	if v.Status == models.BankilyStatusInitiated || v.Status == models.BankilyStatusSubmitted {
		v.Status = models.BankilyStatusPendingVerification
	}
	return v
}

// Config: GET /users/wallet/bankily/config
func (h *BankilyHandler) Config(c *fiber.Ctx) error {
	min, max, decimals := h.svc.Limits()
	return OK(c, fiber.Map{
		"enabled":        h.svc.Enabled(),
		"merchant_code":  h.svc.MerchantCode(),
		"min_amount":     min,
		"max_amount":     max,
		"allow_decimals": decimals,
	})
}

// Pay: POST /users/wallet/bankily/pay
func (h *BankilyHandler) Pay(c *fiber.Ctx) error {
	var req struct {
		Amount         string `json:"amount"`
		ClientPhone    string `json:"client_phone"`
		Passcode       string `json:"passcode"`
		IdempotencyKey string `json:"idempotency_key"`
	}
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}
	amount, err := decimal.NewFromString(strings.TrimSpace(req.Amount))
	if err != nil {
		return BadRequest(c, "Invalid amount")
	}
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c)
	}
	language := "ar"
	if h.userRepo != nil {
		if u, err := h.userRepo.FindByID(c.Context(), userID); err == nil && u.LanguagePref != "" {
			language = u.LanguagePref
		}
	}
	idemKey := req.IdempotencyKey
	if idemKey == "" {
		idemKey = c.Get("Idempotency-Key")
	}
	p, err := h.svc.Pay(c.Context(), userID, services.BankilyPayInput{
		Amount:         amount,
		ClientPhone:    req.ClientPhone,
		Passcode:       req.Passcode,
		IdempotencyKey: idemKey,
		Language:       language,
	})
	req.Passcode = ""
	if err != nil {
		return MapError(c, h.logger, err)
	}
	return OK(c, toBankilyUserView(p))
}

// List: GET /users/wallet/bankily/payments
func (h *BankilyHandler) List(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c)
	}
	items, err := h.svc.ListForUser(c.Context(), userID)
	if err != nil {
		return MapError(c, h.logger, err)
	}
	out := make([]bankilyUserView, 0, len(items))
	for i := range items {
		out = append(out, toBankilyUserView(&items[i]))
	}
	return OK(c, out)
}

// Get: GET /users/wallet/bankily/payments/:id
func (h *BankilyHandler) Get(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid payment ID")
	}
	p, err := h.svc.GetForUser(c.Context(), userID, id)
	if err != nil {
		return MapError(c, h.logger, err)
	}
	return OK(c, toBankilyUserView(p))
}

// Check: POST /users/wallet/bankily/payments/:id/check -- checkTransaction only.
func (h *BankilyHandler) Check(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid payment ID")
	}
	if _, err := h.svc.GetForUser(c.Context(), userID, id); err != nil {
		return MapError(c, h.logger, err)
	}
	p, err := h.svc.Recheck(c.Context(), id, "user", userID)
	if err != nil {
		return MapError(c, h.logger, err)
	}
	return OK(c, toBankilyUserView(p))
}

// AdminList: GET /admin/bankily/payments?status=&page=&per_page=
func (h *BankilyHandler) AdminList(c *fiber.Ctx) error {
	status := strings.ToUpper(c.Query("status"))
	switch status {
	case "", models.BankilyStatusInitiated, models.BankilyStatusSubmitted, models.BankilyStatusPendingVerification, models.BankilyStatusSuccess, models.BankilyStatusFailed, models.BankilyStatusNeedsReview:
	default:
		return BadRequest(c, "Invalid status")
	}
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 20)
	items, total, err := h.svc.ListAdmin(c.Context(), status, page, perPage)
	if err != nil {
		return MapError(c, h.logger, err)
	}
	return PaginatedOK(c, items, fiber.Map{"total": total, "page": page, "per_page": perPage})
}

// AdminRecheck: POST /admin/bankily/payments/:id/recheck -- calls
// checkTransaction only; it can never credit a wallet by itself.
func (h *BankilyHandler) AdminRecheck(c *fiber.Ctx) error {
	adminID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid payment ID")
	}
	p, err := h.svc.Recheck(c.Context(), id, "admin", adminID)
	if err != nil {
		return MapError(c, h.logger, err)
	}
	return OK(c, p)
}
