package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/middleware"
	"github.com/mazadpay/backend/internal/services"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type WalletHandler struct {
	svc    services.WalletService
	logger *zap.Logger
}

func NewWalletHandler(svc services.WalletService, logger *zap.Logger) *WalletHandler {
	return &WalletHandler{svc: svc, logger: logger}
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
		Amount  float64 `json:"amount"`
		Gateway string  `json:"gateway"`
	}
	var req Request
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c)
	}
	amount := decimal.NewFromFloat(req.Amount)
	tx, err := h.svc.InitiateDeposit(c.Context(), userID, amount, req.Gateway)
	if err != nil {
		return InternalError(c, "Failed to initiate deposit")
	}
	return OK(c, tx)
}

func (h *WalletHandler) UploadReceipt(c *fiber.Ctx) error {
	type Request struct {
		ReceiptURL string `json:"receipt_url"`
	}
	var req Request
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	txID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid transaction ID")
	}

	if err := h.svc.UploadReceipt(c.Context(), txID, req.ReceiptURL); err != nil {
		return InternalError(c, "Failed to upload receipt")
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
