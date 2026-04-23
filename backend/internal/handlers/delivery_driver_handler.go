package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/middleware"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/services"
	"go.uber.org/zap"
)

type DeliveryDriverHandler struct {
	svc    services.DeliveryDriverService
	logger *zap.Logger
}

func NewDeliveryDriverHandler(svc services.DeliveryDriverService, logger *zap.Logger) *DeliveryDriverHandler {
	return &DeliveryDriverHandler{svc: svc, logger: logger}
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

	if err := h.svc.Update(c.Context(), driverID, &req); err != nil {
		h.logger.Error("failed to update driver", zap.Error(err))
		return InternalError(c, "Failed to update driver")
	}

	return OK(c, fiber.Map{"message": "Driver updated", "driver_id": driverID})
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
