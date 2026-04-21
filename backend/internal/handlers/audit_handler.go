package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/services"
	"go.uber.org/zap"
)

type AuditHandler struct {
	svc    services.AuditService
	logger *zap.Logger
}

func NewAuditHandler(svc services.AuditService, logger *zap.Logger) *AuditHandler {
	return &AuditHandler{svc: svc, logger: logger}
}

func (h *AuditHandler) GetAuditLogs(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 50)

	logs, total, err := h.svc.List(c.Context(), page, perPage)
	if err != nil {
		return MapError(c, h.logger, err)
	}

	return OK(c, fiber.Map{
		"data":  logs,
		"total": total,
		"page":  page,
		"per_page": perPage,
	})
}

func (h *AuditHandler) GetAuditLogsByEntity(c *fiber.Ctx) error {
	entityType := c.Params("entity_type")
	entityIDParam := c.Params("entity_id")
	entityID, err := uuid.Parse(entityIDParam)
	if err != nil {
		return BadRequest(c, "Invalid entity ID")
	}

	logs, err := h.svc.GetByEntity(c.Context(), entityType, entityID)
	if err != nil {
		return MapError(c, h.logger, err)
	}

	return OK(c, fiber.Map{
		"data": logs,
	})
}
