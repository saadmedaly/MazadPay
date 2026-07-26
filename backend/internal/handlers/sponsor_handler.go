package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/middleware"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/services"
	"go.uber.org/zap"
)

type SponsorHandler struct {
	service  services.SponsorService
	logger   *zap.Logger
	auditSvc services.AuditService
}

func NewSponsorHandler(service services.SponsorService, logger *zap.Logger) *SponsorHandler {
	return &SponsorHandler{
		service: service,
		logger:  logger,
	}
}

// SetAuditService injecte le service d'audit après construction (Content/Settings
// audit logs).
func (h *SponsorHandler) SetAuditService(auditSvc services.AuditService) {
	h.auditSvc = auditSvc
}

func (h *SponsorHandler) logAudit(c *fiber.Ctx, action string, id uuid.UUID, detailsJSON models.JSONB, details string) {
	if h.auditSvc == nil {
		return
	}
	adminID, _ := middleware.GetUserID(c)
	if detailsJSON == nil {
		detailsJSON = models.JSONB{}
	}
	detailsJSON["sponsor_id"] = id.String()
	if auditErr := h.auditSvc.Log(c.Context(), adminID, action, "sponsor", &id, details,
		services.WithActorType("admin"),
		services.WithDetailsJSON(detailsJSON),
		services.WithIP(c.IP()),
		services.WithUserAgent(c.Get("User-Agent")),
	); auditErr != nil {
		if h.logger != nil {
			h.logger.Error("logAudit: failed to write audit log", zap.String("action", action), zap.String("sponsor_id", id.String()), zap.Error(auditErr))
		}
	}
}

func (h *SponsorHandler) ListActive(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	page, limit = ClampPagination(page, limit)

	sponsors, total, err := h.service.ListActiveSponsors(c.Context(), page, limit)
	if err != nil {
		return MapError(c, h.logger, err)
	}

	return OK(c, fiber.Map{
		"sponsors": sponsors,
		"total":    total,
		"page":     page,
		"limit":    limit,
	})
}

func (h *SponsorHandler) ListAll(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)
	page, limit = ClampPagination(page, limit)

	sponsors, total, err := h.service.ListAllSponsors(c.Context(), page, limit)
	if err != nil {
		return MapError(c, h.logger, err)
	}

	return OK(c, fiber.Map{
		"sponsors": sponsors,
		"total":    total,
		"page":     page,
		"limit":    limit,
	})
}

type CreateSponsorRequest struct {
	Name     string  `json:"name" validate:"required"`
	Phone    string  `json:"phone" validate:"required"`
	ImageURL string  `json:"image_url" validate:"required"`
	LinkURL  *string `json:"link_url"`
	IsActive bool    `json:"is_active"`
}

func (h *SponsorHandler) Create(c *fiber.Ctx) error {
	var req CreateSponsorRequest
	
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	sponsor := &models.Sponsor{
		Name:     req.Name,
		Phone:    req.Phone,
		ImageURL: req.ImageURL,
		LinkURL:  req.LinkURL,
		IsActive: req.IsActive,
	}

	if err := h.service.CreateSponsor(c.Context(), sponsor); err != nil {
		return MapError(c, h.logger, err)
	}
	h.logAudit(c, "sponsor_created", sponsor.ID, models.JSONB{"name": sponsor.Name, "is_active": sponsor.IsActive}, fmt.Sprintf("name=%s is_active=%v", sponsor.Name, sponsor.IsActive))

	return Created(c, sponsor)
}

func (h *SponsorHandler) Update(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid sponsor ID")
	}

	var req CreateSponsorRequest
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	sponsor, err := h.service.GetSponsorByID(c.Context(), id)
	if err != nil {
		return NotFound(c, "Sponsor not found")
	}

	sponsor.Name = req.Name
	sponsor.Phone = req.Phone
	sponsor.ImageURL = req.ImageURL
	sponsor.LinkURL = req.LinkURL
	sponsor.IsActive = req.IsActive

	if err := h.service.UpdateSponsor(c.Context(), sponsor); err != nil {
		return MapError(c, h.logger, err)
	}
	h.logAudit(c, "sponsor_updated", id, models.JSONB{"name": sponsor.Name, "is_active": sponsor.IsActive}, fmt.Sprintf("name=%s is_active=%v", sponsor.Name, sponsor.IsActive))

	return OK(c, sponsor)
}

func (h *SponsorHandler) ToggleStatus(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid sponsor ID")
	}

	if err := h.service.ToggleSponsorStatus(c.Context(), id); err != nil {
		return MapError(c, h.logger, err)
	}
	h.logAudit(c, "sponsor_updated", id, nil, "status toggled")

	return OK(c, fiber.Map{"message": "Sponsor status updated"})
}

func (h *SponsorHandler) Delete(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid sponsor ID")
	}

	if err := h.service.DeleteSponsor(c.Context(), id); err != nil {
		return MapError(c, h.logger, err)
	}
	h.logAudit(c, "sponsor_deleted", id, nil, "")

	return OK(c, fiber.Map{"message": "Sponsor deleted"})
}
