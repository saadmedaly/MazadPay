package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/services"
	"go.uber.org/zap"
)

type PaymentMethodHandler struct {
	svc    services.PaymentMethodService
	logger *zap.Logger
}

func NewPaymentMethodHandler(svc services.PaymentMethodService, logger *zap.Logger) *PaymentMethodHandler {
	return &PaymentMethodHandler{svc: svc, logger: logger}
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

	return OK(c, fiber.Map{"message": "Payment method status toggled"})
}
