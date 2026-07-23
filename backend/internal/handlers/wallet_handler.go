package handlers

import (
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	apperr "github.com/mazadpay/backend/internal/errors"
	"github.com/mazadpay/backend/internal/middleware"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/services"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type WalletHandler struct {
	svc      services.WalletService
	mediaSvc services.MediaService
	auditSvc services.AuditService
	logger   *zap.Logger
}

func NewWalletHandler(svc services.WalletService, logger *zap.Logger, mediaSvc ...services.MediaService) *WalletHandler {
	h := &WalletHandler{svc: svc, logger: logger}
	if len(mediaSvc) > 0 {
		h.mediaSvc = mediaSvc[0]
	}
	return h
}

// SetAuditService injecte le service d'audit après construction pour ne pas modifier
// la signature variadique de NewWalletHandler ailleurs dans le code (Financial audit
// logs — journalise uniquement la consultation d'un reçu par un admin).
func (h *WalletHandler) SetAuditService(auditSvc services.AuditService) {
	h.auditSvc = auditSvc
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

// receiptMaxSizeBytes limite la taille d'un reçu de paiement uploadé (audit de sécurité —
// ce endpoint n'avait auparavant aucune limite propre, seulement le BodyLimit global).
const receiptMaxSizeBytes = 5 * 1024 * 1024

func (h *WalletHandler) UploadReceipt(c *fiber.Ctx) error {
	txID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid transaction ID")
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c)
	}

	// Vérifie que la transaction appartient bien à l'utilisateur AVANT tout upload
	// physique vers le stockage (audit de sécurité) : auparavant le fichier était
	// uploadé vers R2/local puis seule la mise à jour SQL vérifiait l'appartenance,
	// laissant un fichier orphelin en cas de rejet.
	if _, err := h.svc.GetTransaction(c.Context(), userID, txID); err != nil {
		return NotFound(c, "Transaction")
	}

	var receiptURL string

	// Accept multipart file upload
	fileHeader, err := c.FormFile("receipt")
	if err == nil && fileHeader != nil {
		if h.mediaSvc == nil {
			return InternalError(c, "Media service not available")
		}
		if fileHeader.Size > receiptMaxSizeBytes {
			return BadRequest(c, "Receipt file too large (max 5MB)")
		}
		file, err := fileHeader.Open()
		if err != nil {
			return InternalError(c, "Failed to open receipt file")
		}
		defer file.Close()
		// Upload vers le bucket R2 privé des reçus — receiptURL ne contient plus qu'un
		// object key (ex: "receipts/<uuid>.jpg"), jamais une URL publique (audit de
		// sécurité — Private Receipts Bucket Separation).
		receiptURL, err = h.mediaSvc.UploadPrivateFile(c.Context(), file, fileHeader, "receipts")
		if err != nil {
			return BadRequest(c, "Invalid receipt file")
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

// receiptURLExpiry est la durée de validité maximale d'une URL présignée d'un reçu de
// paiement (audit de sécurité — jamais de lien permanent, jamais stocké en DB).
const receiptURLExpiry = 5 * time.Minute

// GetReceiptURL génère une URL R2 temporaire pour visualiser le reçu d'une transaction.
// Accès réservé au propriétaire de la transaction ou à un admin (audit de sécurité —
// remplace l'exposition directe de receipt_url, une URL publique permanente).
func (h *WalletHandler) GetReceiptURL(c *fiber.Ctx) error {
	txID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid transaction ID")
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c)
	}

	role, _ := c.Locals("user_role").(string)
	isAdmin := strings.EqualFold(role, "admin") || strings.EqualFold(role, "super_admin")

	var tx *models.Transaction
	if isAdmin {
		tx, err = h.svc.GetTransactionAny(c.Context(), txID)
	} else {
		tx, err = h.svc.GetTransaction(c.Context(), userID, txID)
	}
	if err != nil || tx == nil {
		return NotFound(c, "Transaction")
	}

	// Journaliser uniquement la consultation par un admin du reçu d'un autre
	// utilisateur (Financial audit logs) — jamais l'URL elle-même, jamais quand
	// l'utilisateur consulte son propre reçu (bruit inutile, pas un accès privilégié).
	if isAdmin && h.auditSvc != nil {
		detailsJSON := models.JSONB{
			"transaction_id": txID.String(),
			"user_id":        tx.UserID.String(),
		}
		h.auditSvc.Log(c.Context(), userID, "receipt_viewed_by_admin", "transaction", &txID, "Admin viewed payment receipt",
			services.WithActorType("admin"),
			services.WithDetailsJSON(detailsJSON),
			services.WithIP(c.IP()),
			services.WithUserAgent(c.Get("User-Agent")),
		)
	}
	if tx.ReceiptURL == nil || *tx.ReceiptURL == "" {
		return NotFound(c, "Receipt")
	}
	if h.mediaSvc == nil {
		return InternalError(c, "Media service not available")
	}

	// GetReceiptURL choisit automatiquement le bon bucket : privé pour les nouveaux
	// reçus (object key), ou le bucket média public (en lecture seule, temporaire) pour
	// les anciens reçus stockés avant la séparation des buckets (audit de sécurité).
	url, err := h.mediaSvc.GetReceiptURL(c.Context(), *tx.ReceiptURL, receiptURLExpiry)
	if err != nil {
		return InternalError(c, "Failed to generate receipt URL")
	}

	return OK(c, fiber.Map{
		"url":        url,
		"expires_in": int(receiptURLExpiry.Seconds()),
	})
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
	page, perPage = ClampPagination(page, perPage)

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
