package handlers

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/mazadpay/backend/internal/middleware"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/services"
	"go.uber.org/zap"
)

type PaymentMethodHandler struct {
	svc      services.PaymentMethodService
	logger   *zap.Logger
	auditSvc services.AuditService
}

func NewPaymentMethodHandler(svc services.PaymentMethodService, logger *zap.Logger) *PaymentMethodHandler {
	return &PaymentMethodHandler{svc: svc, logger: logger}
}

// SetAuditService injecte le service d'audit après construction (Content/Settings
// audit logs).
func (h *PaymentMethodHandler) SetAuditService(auditSvc services.AuditService) {
	h.auditSvc = auditSvc
}

func (h *PaymentMethodHandler) logAudit(c *fiber.Ctx, action string, id int, detailsJSON models.JSONB, details string) {
	if h.auditSvc == nil {
		return
	}
	adminID, _ := middleware.GetUserID(c)
	if detailsJSON == nil {
		detailsJSON = models.JSONB{}
	}
	detailsJSON["payment_method_id"] = id
	if auditErr := h.auditSvc.Log(c.Context(), adminID, action, "payment_method", nil, fmt.Sprintf("payment_method_id=%d %s", id, details),
		services.WithActorType("admin"),
		services.WithDetailsJSON(detailsJSON),
		services.WithEntityKey(strconv.Itoa(id)),
		services.WithIP(c.IP()),
		services.WithUserAgent(c.Get("User-Agent")),
	); auditErr != nil {
		if h.logger != nil {
			h.logger.Error("logAudit: failed to write audit log", zap.String("action", action), zap.Int("payment_method_id", id), zap.Error(auditErr))
		}
	}
}

// ListPaymentMethods - GET /api/payment-methods
func (h *PaymentMethodHandler) ListPaymentMethods(c *fiber.Ctx) error {
	methods, err := h.svc.List(c.Context())
	if err != nil {
		h.logger.Error("failed to list payment methods", zap.Error(err))
		return InternalError(c, "Failed to list payment methods")
	}
	return OK(c, fiber.Map{"payment_methods": methods})
}

// CreatePaymentMethod - POST /api/admin/payment-methods (Admin only)
func (h *PaymentMethodHandler) CreatePaymentMethod(c *fiber.Ctx) error {
	var req models.PaymentMethod
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	if err := h.svc.Create(c.Context(), &req); err != nil {
		h.logger.Error("failed to create payment method", zap.Error(err))
		return InternalError(c, "Failed to create payment method")
	}
	h.logAudit(c, "payment_method_created", req.ID, models.JSONB{"code": req.Code, "name_ar": req.NameAr}, fmt.Sprintf("code=%s name_ar=%s", req.Code, req.NameAr))

	return OK(c, fiber.Map{"message": "Payment method created", "payment_method": req})
}

// UpdatePaymentMethod - PUT /api/admin/payment-methods/:id (Admin only)
func (h *PaymentMethodHandler) UpdatePaymentMethod(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid payment method ID")
	}

	var req models.PaymentMethod
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	if err := h.svc.Update(c.Context(), id, &req); err != nil {
		h.logger.Error("failed to update payment method", zap.Error(err))
		return InternalError(c, "Failed to update payment method")
	}
	h.logAudit(c, "payment_method_updated", id, models.JSONB{"code": req.Code, "name_ar": req.NameAr}, fmt.Sprintf("code=%s name_ar=%s", req.Code, req.NameAr))

	return OK(c, fiber.Map{"message": "Payment method updated"})
}

// DeletePaymentMethod - DELETE /api/admin/payment-methods/:id (Admin only)
func (h *PaymentMethodHandler) DeletePaymentMethod(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid payment method ID")
	}

	if err := h.svc.Delete(c.Context(), id); err != nil {
		h.logger.Error("failed to delete payment method", zap.Error(err))
		return InternalError(c, "Failed to delete payment method")
	}
	h.logAudit(c, "payment_method_deleted", id, nil, "")

	return OK(c, fiber.Map{"message": "Payment method deleted"})
}

// TogglePaymentMethodStatus - PUT /api/admin/payment-methods/:id/toggle (Admin only)
func (h *PaymentMethodHandler) TogglePaymentMethodStatus(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid payment method ID")
	}

	if err := h.svc.ToggleStatus(c.Context(), id); err != nil {
		h.logger.Error("failed to toggle payment method status", zap.Error(err))
		return InternalError(c, "Failed to toggle payment method status")
	}
	h.logAudit(c, "payment_method_updated", id, nil, "status toggled")

	return OK(c, fiber.Map{"message": "Payment method status toggled"})
}
