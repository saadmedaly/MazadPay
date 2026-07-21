package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	apperr "github.com/mazadpay/backend/internal/errors"
	"github.com/mazadpay/backend/internal/middleware"
	"github.com/mazadpay/backend/internal/services"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type WalletHandler struct {
	svc      services.WalletService
	mediaSvc services.MediaService
	logger   *zap.Logger
}

func NewWalletHandler(svc services.WalletService, logger *zap.Logger, mediaSvc ...services.MediaService) *WalletHandler {
	h := &WalletHandler{svc: svc, logger: logger}
	if len(mediaSvc) > 0 {
		h.mediaSvc = mediaSvc[0]
	}
	return h
}

func (h *WalletHandler) GetMe(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c)
	}
	wallet, err := h.svc.GetBalance(c.Context(), userID)
	if err != nil {
		return InternalError(c, "Failed to get wallet: "+err.Error())
	}
	return OK(c, wallet)
}

func (h *WalletHandler) Deposit(c *fiber.Ctx) error {
	type Request struct {
		Amount           float64 `json:"amount" validate:"required,gt=0"`
		Gateway          string  `json:"gateway" validate:"required"`
		PaymentMethod    string  `json:"payment_method"` // masrvi, bankily, sedad, click, etc.
		ReceiptImageTemp string  `json:"receipt_image_temp"` // Image uploadée avant validation
	}
	var req Request
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}
	// Vérification explicite : les tags `validate` sur ce struct ne sont pas exécutés
	// (aucun validator.Validate n'est instancié dans WalletHandler), donc on vérifie
	// manuellement ici plutôt que de compter sur des tags décoratifs (audit V05-bis/V11).
	if req.Amount <= 0 {
		return BadRequest(c, "Amount must be greater than 0")
	}
	if req.Gateway == "" {
		return BadRequest(c, "Gateway is required")
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c)
	}
	amount := decimal.NewFromFloat(req.Amount)
	tx, err := h.svc.InitiateDeposit(c.Context(), userID, amount, req.Gateway, req.PaymentMethod, req.ReceiptImageTemp)
	if err != nil {
		return InternalError(c, "Failed to initiate deposit")
	}
	return OK(c, tx)
}

func (h *WalletHandler) UploadReceipt(c *fiber.Ctx) error {
	txID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid transaction ID")
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c)
	}

	var receiptURL string

	// Accept multipart file upload
	fileHeader, err := c.FormFile("receipt")
	if err == nil && fileHeader != nil {
		if h.mediaSvc == nil {
			return InternalError(c, "Media service not available")
		}
		file, err := fileHeader.Open()
		if err != nil {
			return InternalError(c, "Failed to open receipt file")
		}
		defer file.Close()
		receiptURL, err = h.mediaSvc.UploadFile(c.Context(), file, fileHeader, "receipts")
		if err != nil {
			return InternalError(c, "Failed to upload receipt image")
		}
	} else {
		// Fallback: JSON body with receipt_url
		type Request struct {
			ReceiptURL string `json:"receipt_url"`
		}
		var req Request
		if err := c.BodyParser(&req); err != nil || req.ReceiptURL == "" {
			return BadRequest(c, "Receipt file or receipt_url required")
		}
		receiptURL = req.ReceiptURL
	}

	if err := h.svc.UploadReceipt(c.Context(), txID, userID, receiptURL); err != nil {
		if err == apperr.ErrNotFound {
			return NotFound(c, "Transaction")
		}
		return InternalError(c, "Failed to update receipt")
	}

	return OK(c, fiber.Map{"message": "Receipt uploaded successfully"})
}

func (h *WalletHandler) Withdraw(c *fiber.Ctx) error {
	type Request struct {
		Amount  float64 `json:"amount"`
		Gateway string  `json:"gateway"`
	}
	var req Request
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}
	if req.Amount <= 0 {
		return BadRequest(c, "Amount must be greater than 0")
	}
	if req.Gateway == "" {
		return BadRequest(c, "Gateway is required")
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c)
	}
	amount := decimal.NewFromFloat(req.Amount)
	tx, err := h.svc.RequestWithdraw(c.Context(), userID, amount, req.Gateway)
	if err != nil {
		return InternalError(c, err.Error())
	}
	return OK(c, tx)
}

func (h *WalletHandler) Transactions(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c)
	}
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "20"))

	txs, total, err := h.svc.GetTransactions(c.Context(), userID, page, perPage)
	if err != nil {
		return InternalError(c, "Failed to get transactions")
	}

	return PaginatedOK(c, txs, fiber.Map{
		"page":     page,
		"per_page": perPage,
		"total":    total,
	})
}

func (h *WalletHandler) GetTransaction(c *fiber.Ctx) error {
	txID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid transaction ID")
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c)
	}

	tx, err := h.svc.GetTransaction(c.Context(), userID, txID)
	if err != nil {
		return NotFound(c, "Transaction")
	}

	return OK(c, tx)
}

// GetPaymentMethods - GET /api/v1/wallet/payment-methods
func (h *WalletHandler) GetPaymentMethods(c *fiber.Ctx) error {
	_, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c)
	}

	methods, err := h.svc.GetPaymentMethods(c.Context())
	if err != nil {
		return InternalError(c, "Failed to get payment methods")
	}
	return OK(c, methods)
}
