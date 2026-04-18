package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/middleware"
	"github.com/mazadpay/backend/internal/services"
	"go.uber.org/zap"
)

type NotificationHandler struct {
	svc    services.NotificationService
	logger *zap.Logger
}

func NewNotificationHandler(svc services.NotificationService, logger *zap.Logger) *NotificationHandler {
	return &NotificationHandler{svc: svc, logger: logger}
}

type SaveTokenRequest struct {
	FCMToken string `json:"fcm_token" validate:"required"`
	DeviceID string `json:"device_id"`
	Platform string `json:"platform"` // web, android, ios
}

func (h *NotificationHandler) SaveToken(c *fiber.Ctx) error {
		userID, err := middleware.GetUserID(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
		}

		var req SaveTokenRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
		}

		if err := h.svc.SavePushToken(c.Context(), userID, req.FCMToken, req.DeviceID, req.Platform); err != nil {
			h.logger.Error("failed to save push token", zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to save token"})
		}

		return c.SendStatus(fiber.StatusOK)
}

func (h *NotificationHandler) List(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	limitStr := c.Query("limit", "20")
	limit, _ := strconv.Atoi(limitStr)

	notifications, err := h.svc.ListNotifications(c.Context(), userID, limit)
	if err != nil {
		h.logger.Error("failed to list notifications", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to list notifications"})
	}

	return c.JSON(fiber.Map{"data": notifications})
}

func (h *NotificationHandler) MarkAllAsRead(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	if err := h.svc.MarkAllAsRead(c.Context(), userID); err != nil {
		h.logger.Error("failed to mark notifications as read", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to mark as read"})
	}

	return c.SendStatus(fiber.StatusOK)
}

func (h *NotificationHandler) MarkAsRead(c *fiber.Ctx) error {
	_, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid notification id"})
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	notifUUID := uuid.Must(uuid.Parse(c.Params("id")))
	if err := h.svc.MarkAsRead(c.Context(), notifUUID, userID); err != nil {
		h.logger.Error("failed to mark notification as read", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to mark as read"})
	}

	return c.SendStatus(fiber.StatusOK)
}
