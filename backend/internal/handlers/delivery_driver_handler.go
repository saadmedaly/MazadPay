package handlers

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	apperr "github.com/mazadpay/backend/internal/errors"
	"github.com/mazadpay/backend/internal/middleware"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/services"
	"go.uber.org/zap"
)

type DeliveryDriverHandler struct {
	svc      services.DeliveryDriverService
	logger   *zap.Logger
	auditSvc services.AuditService
}

func NewDeliveryDriverHandler(svc services.DeliveryDriverService, logger *zap.Logger) *DeliveryDriverHandler {
	return &DeliveryDriverHandler{svc: svc, logger: logger}
}

// SetAuditService injecte le service d'audit après construction (Delivery Drivers
// Phase 3), même convention que BannerHandler/PaymentMethodHandler/etc. — évite de
// changer la signature de NewDeliveryDriverHandler et son appel dans routes.go.
func (h *DeliveryDriverHandler) SetAuditService(auditSvc services.AuditService) {
	h.auditSvc = auditSvc
}

// logDriverAudit journalise une action admin sur un delivery_driver (Delivery
// Drivers Phase 3). Ne journalise jamais license_number/vehicle_plate/coordonnées
// GPS — uniquement driver_id, user_id, et vehicle_type (jugé non sensible).
func (h *DeliveryDriverHandler) logDriverAudit(c *fiber.Ctx, action string, driverID uuid.UUID, userID *uuid.UUID, vehicleType *string) {
	if h.auditSvc == nil {
		return
	}
	adminID, _ := middleware.GetUserID(c)
	detailsJSON := models.JSONB{"driver_id": driverID.String()}
	if userID != nil {
		detailsJSON["user_id"] = userID.String()
	}
	if vehicleType != nil {
		detailsJSON["vehicle_type"] = *vehicleType
	}
	if auditErr := h.auditSvc.Log(c.Context(), adminID, action, "delivery_driver", &driverID,
		fmt.Sprintf("driver_id=%s", driverID.String()),
		services.WithActorType("admin"),
		services.WithDetailsJSON(detailsJSON),
		services.WithIP(c.IP()),
		services.WithUserAgent(c.Get("User-Agent")),
	); auditErr != nil {
		if h.logger != nil {
			h.logger.Error("logDriverAudit: failed to write audit log", zap.String("action", action), zap.String("driver_id", driverID.String()), zap.Error(auditErr))
		}
	}
}

// validateDriverFields applique les règles Delivery Drivers Phase 5 sur
// vehicle_type/vehicle_plate/vehicle_color/license_number, communes à Register et
// Update. Retourne les valeurs nettoyées (trim) et une erreur descriptive sinon.
// Ne valide jamais current_location_lat/lng (hors périmètre — self-service
// uniquement) ni ne pré-vérifie l'unicité de vehicle_plate/license_number : le
// schéma delivery_drivers (migration 000031) n'a pas de contrainte UNIQUE sur ces
// colonnes, donc des doublons restent possibles côté DB tant qu'une migration
// dédiée ne les interdit pas — non ajoutée dans cette phase.
func validateDriverFields(vehicleType, vehiclePlate, vehicleColor, licenseNumber string) (string, string, string, string, error) {
	vehicleType = strings.TrimSpace(vehicleType)
	vehiclePlate = strings.TrimSpace(vehiclePlate)
	vehicleColor = strings.TrimSpace(vehicleColor)
	licenseNumber = strings.TrimSpace(licenseNumber)

	if vehicleType == "" {
		return "", "", "", "", fmt.Errorf("vehicle_type is required")
	}
	if len(vehicleType) > 50 {
		return "", "", "", "", fmt.Errorf("vehicle_type must be at most 50 characters")
	}
	if vehiclePlate == "" {
		return "", "", "", "", fmt.Errorf("vehicle_plate is required")
	}
	if len(vehiclePlate) > 20 {
		return "", "", "", "", fmt.Errorf("vehicle_plate must be at most 20 characters")
	}
	if len(vehicleColor) > 50 {
		return "", "", "", "", fmt.Errorf("vehicle_color must be at most 50 characters")
	}
	if licenseNumber == "" {
		return "", "", "", "", fmt.Errorf("license_number is required")
	}
	if len(licenseNumber) > 50 {
		return "", "", "", "", fmt.Errorf("license_number must be at most 50 characters")
	}
	return vehicleType, vehiclePlate, vehicleColor, licenseNumber, nil
}

// RegisterDriver - POST /api/admin/drivers/register (Admin only)
func (h *DeliveryDriverHandler) RegisterDriver(c *fiber.Ctx) error {
	type Request struct {
		UserID        uuid.UUID `json:"user_id" validate:"required"`
		VehicleType   string    `json:"vehicle_type" validate:"required"`
		VehiclePlate  string    `json:"vehicle_plate" validate:"required"`
		VehicleColor  string    `json:"vehicle_color"`
		LicenseNumber string    `json:"license_number" validate:"required"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	if req.UserID == uuid.Nil {
		return BadRequest(c, "user_id is required")
	}

	vehicleType, vehiclePlate, vehicleColor, licenseNumber, err := validateDriverFields(req.VehicleType, req.VehiclePlate, req.VehicleColor, req.LicenseNumber)
	if err != nil {
		return BadRequest(c, err.Error())
	}
	req.VehicleType = vehicleType
	req.VehiclePlate = vehiclePlate
	req.VehicleColor = vehicleColor
	req.LicenseNumber = licenseNumber

	driver := &models.DeliveryDriver{
		UserID:         &req.UserID,
		VehicleType:    &req.VehicleType,
		VehiclePlate:   &req.VehiclePlate,
		VehicleColor:   &req.VehicleColor,
		LicenseNumber:  &req.LicenseNumber,
		IsAvailable:    true,
		TotalDeliveries: 0,
	}

	if err := h.svc.Create(c.Context(), driver); err != nil {
		h.logger.Error("failed to register driver", zap.Error(err))
		return InternalError(c, "Failed to register driver")
	}

	h.logDriverAudit(c, "delivery_driver_registered", driver.ID, driver.UserID, driver.VehicleType)

	return OK(c, fiber.Map{"message": "Driver registered successfully", "driver": driver})
}

// ListDrivers - GET /api/admin/drivers (Admin only)
func (h *DeliveryDriverHandler) ListDrivers(c *fiber.Ctx) error {
	drivers, err := h.svc.List(c.Context())
	if err != nil {
		h.logger.Error("failed to list drivers", zap.Error(err))
		return InternalError(c, "Failed to list drivers")
	}
	return OK(c, fiber.Map{"drivers": drivers})
}

// GetDriver - GET /api/admin/drivers/:id (Admin only)
func (h *DeliveryDriverHandler) GetDriver(c *fiber.Ctx) error {
	driverID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid driver ID")
	}

	driver, err := h.svc.GetByID(c.Context(), driverID)
	if err != nil {
		h.logger.Error("failed to get driver", zap.Error(err))
		return NotFound(c, "Driver")
	}

	return OK(c, fiber.Map{"driver": driver})
}

// UpdateDriver - PUT /api/admin/drivers/:id (Admin only)
func (h *DeliveryDriverHandler) UpdateDriver(c *fiber.Ctx) error {
	driverID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid driver ID")
	}

	var req models.DeliveryDriver
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	var vehicleType, vehiclePlate, vehicleColor, licenseNumber string
	if req.VehicleType != nil {
		vehicleType = *req.VehicleType
	}
	if req.VehiclePlate != nil {
		vehiclePlate = *req.VehiclePlate
	}
	if req.VehicleColor != nil {
		vehicleColor = *req.VehicleColor
	}
	if req.LicenseNumber != nil {
		licenseNumber = *req.LicenseNumber
	}
	vehicleType, vehiclePlate, vehicleColor, licenseNumber, err = validateDriverFields(vehicleType, vehiclePlate, vehicleColor, licenseNumber)
	if err != nil {
		return BadRequest(c, err.Error())
	}
	req.VehicleType = &vehicleType
	req.VehiclePlate = &vehiclePlate
	req.VehicleColor = &vehicleColor
	req.LicenseNumber = &licenseNumber

	if err := h.svc.Update(c.Context(), driverID, &req); err != nil {
		h.logger.Error("failed to update driver", zap.Error(err))
		return InternalError(c, "Failed to update driver")
	}

	h.logDriverAudit(c, "delivery_driver_updated", driverID, req.UserID, req.VehicleType)

	return OK(c, fiber.Map{"message": "Driver updated", "driver_id": driverID})
}

// DeleteDriver - DELETE /api/admin/drivers/:id (Super Admin only)
// Delivery Drivers Phase 4 : Delete() vérifie désormais RowsAffected et renvoie
// apperr.ErrNotFound si l'id n'existe pas, au lieu d'un succès silencieux.
func (h *DeliveryDriverHandler) DeleteDriver(c *fiber.Ctx) error {
	driverID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid driver ID")
	}

	if err := h.svc.Delete(c.Context(), driverID); err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			return NotFound(c, "Driver")
		}
		h.logger.Error("failed to delete driver", zap.Error(err))
		return InternalError(c, "Failed to delete driver")
	}

	h.logDriverAudit(c, "delivery_driver_deleted", driverID, nil, nil)

	return OK(c, fiber.Map{"message": "Driver deleted", "driver_id": driverID})
}

// UpdateDriverLocation - PUT /api/drivers/location (Driver only)
func (h *DeliveryDriverHandler) UpdateDriverLocation(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c)
	}

	type Request struct {
		Lat float64 `json:"lat" validate:"required"`
		Lng float64 `json:"lng" validate:"required"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	driver, err := h.svc.GetByUserID(c.Context(), userID)
	if err != nil {
		h.logger.Error("failed to find driver for user", zap.Error(err))
		return NotFound(c, "Driver")
	}

	if err := h.svc.UpdateLocation(c.Context(), driver.ID, req.Lat, req.Lng); err != nil {
		h.logger.Error("failed to update driver location", zap.Error(err))
		return InternalError(c, "Failed to update location")
	}

	return OK(c, fiber.Map{"message": "Location updated"})
}

// ToggleAvailability - PUT /api/drivers/availability (Driver only)
func (h *DeliveryDriverHandler) ToggleAvailability(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c)
	}

	type Request struct {
		Available bool `json:"available" validate:"required"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	driver, err := h.svc.GetByUserID(c.Context(), userID)
	if err != nil {
		h.logger.Error("failed to find driver for user", zap.Error(err))
		return NotFound(c, "Driver")
	}

	if err := h.svc.UpdateAvailability(c.Context(), driver.ID, req.Available); err != nil {
		h.logger.Error("failed to update availability", zap.Error(err))
		return InternalError(c, "Failed to update availability")
	}

	return OK(c, fiber.Map{"message": "Availability updated", "available": req.Available})
}

// GetAvailableDrivers - GET /api/admin/drivers/available (Admin only)
func (h *DeliveryDriverHandler) GetAvailableDrivers(c *fiber.Ctx) error {
	drivers, err := h.svc.List(c.Context())
	if err != nil {
		h.logger.Error("failed to list available drivers", zap.Error(err))
		return InternalError(c, "Failed to list available drivers")
	}

	// Filter only available drivers
	var available []models.DeliveryDriver
	for _, d := range drivers {
		if d.IsAvailable {
			available = append(available, d)
		}
	}

	return OK(c, fiber.Map{"drivers": available})
}
