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

type AdminListNotificationsRequest struct {
	UserID string `query:"user_id"`
	Status string `query:"status"`
	Limit  int    `query:"limit,20"`
}

func (h *NotificationHandler) AdminList(c *fiber.Ctx) error {
	limitStr := c.Query("limit", "20")
	limit, _ := strconv.Atoi(limitStr)
	status := c.Query("status", "all")

	notifications, err := h.svc.AdminListNotifications(c.Context(), status, limit)
	if err != nil {
		h.logger.Error("failed to list notifications", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to list notifications"})
	}

	return c.JSON(fiber.Map{"data": notifications})
}

type SendNotificationRequest struct {
	UserID     string `json:"user_id"`
	Title      string `json:"title" validate:"required"`
	Body       string `json:"body" validate:"required"`
	Type       string `json:"type"`
	Data       map[string]string `json:"data"`
	Broadcast  bool   `json:"broadcast"`
}

func (h *NotificationHandler) SendNotification(c *fiber.Ctx) error {
	var req SendNotificationRequest
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	h.logger.Info("SendNotification request", zap.Any("req", req))

	// Priority: broadcast > specific user > admins
	if req.Broadcast {
		// Send to all users
		if err := h.svc.SendBroadcast(c.Context(), req.Title, req.Body, req.Type, req.Data); err != nil {
			h.logger.Error("broadcast failed", zap.Error(err))
			return MapError(c, h.logger, err)
		}
		return OK(c, fiber.Map{"message": "Notification sent to all users", "type": "broadcast"})
	}

	if req.UserID != "" {
		// Send to specific user
		userUUID, err := uuid.Parse(req.UserID)
		if err != nil {
			return BadRequest(c, "Invalid user ID")
		}
		if err := h.svc.SendPush(c.Context(), userUUID, req.Title, req.Body, req.Data); err != nil {
			h.logger.Error("send to user failed", zap.Error(err))
			return MapError(c, h.logger, err)
		}
		return OK(c, fiber.Map{"message": "Notification sent to user", "user_id": req.UserID})
	}

	// Default: notify admins only
	if err := h.svc.NotifyAdmins(c.Context(), req.Title, req.Body, req.Data); err != nil {
		h.logger.Error("notify admins failed", zap.Error(err))
		return MapError(c, h.logger, err)
	}

	return OK(c, fiber.Map{"message": "Notification sent to admins"})
}

func (h *NotificationHandler) AdminDelete(c *fiber.Ctx) error {
	notifID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return BadRequest(c, "Invalid notification ID")
	}

	if err := h.svc.DeleteNotification(c.Context(), notifID); err != nil {
		return MapError(c, h.logger, err)
	}

	return OK(c, fiber.Map{"message": "Notification deleted"})
}
