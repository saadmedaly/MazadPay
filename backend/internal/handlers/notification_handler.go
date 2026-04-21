package handlers

import (
	"context"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/middleware"
	"github.com/mazadpay/backend/internal/services"
	"go.uber.org/zap"
)

type NotificationHandler struct {
	svc      services.NotificationService
	logger   *zap.Logger
	validate *validator.Validate
}

func NewNotificationHandler(svc services.NotificationService, logger *zap.Logger) *NotificationHandler {
	return &NotificationHandler{
		svc:      svc,
		logger:   logger,
		validate: validator.New(),
	}
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

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if err := h.svc.SavePushToken(c.Context(), userID, req.FCMToken, req.DeviceID, req.Platform); err != nil {
		h.logger.Error("Failed to save push token", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save push token"})
	}

	return c.JSON(fiber.Map{"message": "Push token saved successfully"})
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

// Get notification templates
func (h *NotificationHandler) getTemplate(lang, notifType string) string {
	templates := map[string]map[string]string{
		"ar": {
			"bid_placed":    "تم تقديم مزايدة جديدة على {auction_title}",
			"bid_outbid":    "لقد تم تجاوز مزايدتك في {auction_title}",
			"auction_won":   "تهانينا ! لقد فزت بمزاد {auction_title}",
			"auction_lost":   "لللأسف، لم تفز بمزاد {auction_title}",
			"payment_received": "تم استلام مبلغ الدفع بنجاح",
			"auction_started": "بدأ المزاد {auction_title}",
			"auction_ended":   "انتهى المزاد {auction_title}",
		},
		"fr": {
			"bid_placed":    "Nouvelle enchère placée sur {auction_title}",
			"bid_outbid":    "Votre enchère sur {auction_title} a été dépassée",
			"auction_won":   "Félicitations ! Vous avez remporté l'enchère {auction_title}",
			"auction_lost":   "Désolé, vous n'avez pas remporté l'enchère {auction_title}",
			"payment_received": "Paiement reçu avec succès",
			"auction_started": "L'enchère {auction_title} a commencé",
			"auction_ended":   "L'enchère {auction_title} est terminée",
		},
		"en": {
			"bid_placed":    "New bid placed on {auction_title}",
			"bid_outbid":    "Your bid on {auction_title} has been outbid",
			"auction_won":   "Congratulations! You won auction {auction_title}",
			"auction_lost":   "Sorry, you didn't win auction {auction_title}",
			"payment_received": "Payment received successfully",
			"auction_started": "Auction {auction_title} has started",
			"auction_ended":   "Auction {auction_title} has ended",
		},
	}

	if langTemplates, exists := templates[lang]; exists {
		if template, templateExists := langTemplates[notifType]; templateExists {
			return template
		}
	}
	
	// Default to Arabic
	if arTemplate, exists := templates["ar"]; exists {
		if template, templateExists := arTemplate[notifType]; templateExists {
			return template
		}
	}
	
	return notifType
}

func (h *NotificationHandler) SendNotification(c *fiber.Ctx) error {
	var req SendNotificationRequest
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	if err := h.validate.Struct(req); err != nil {
		return BadRequest(c, err.Error())
	}

	// Get user language preference (default to Arabic)
	lang := c.Get("Accept-Language", "ar")
	if len(lang) > 2 {
		lang = lang[:2] // Extract "ar", "fr", "en"
	}

	// Generate localized message
	message := req.Body
	if template, exists := map[string]string{
		"bid_placed":    h.getTemplate(lang, "bid_placed"),
		"bid_outbid":    h.getTemplate(lang, "bid_outbid"),
		"auction_won":   h.getTemplate(lang, "auction_won"),
		"auction_lost":   h.getTemplate(lang, "auction_lost"),
		"payment_received": h.getTemplate(lang, "payment_received"),
		"auction_started": h.getTemplate(lang, "auction_started"),
		"auction_ended":   h.getTemplate(lang, "auction_ended"),
	}[req.Type]; exists {
		message = template
	}

	h.logger.Info("SendNotification request", 
		zap.String("type", req.Type),
		zap.String("lang", lang),
		zap.String("title", req.Title))

	// Priority: broadcast > specific user > admins
	if req.Broadcast {
		// Send to all users
		if err := h.svc.SendBroadcast(c.Context(), req.Title, message, req.Type, req.Data); err != nil {
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
		if err := h.svc.SendPush(c.Context(), userUUID, req.Title, message, req.Data); err != nil {
			h.logger.Error("send to user failed", zap.Error(err))
			return MapError(c, h.logger, err)
		}
		return OK(c, fiber.Map{"message": "Notification sent to user", "user_id": req.UserID})
	}

	// Default: notify admins only
	if err := h.svc.NotifyAdmins(c.Context(), req.Title, message, req.Data); err != nil {
		h.logger.Error("notify admins failed", zap.Error(err))
		return MapError(c, h.logger, err)
	}

	return OK(c, fiber.Map{"message": "Notification sent to admins"})
}

func (h *NotificationHandler) GetTemplates(c *fiber.Ctx) error {
	lang := c.Query("lang", "ar")
	
	templates := map[string]map[string]string{
		"ar": {
			"bid_placed":    "تم تقديم مزايدة جديدة على {auction_title}",
			"bid_outbid":    "لقد تم تجاوز مزايدتك في {auction_title}",
			"auction_won":   "تهانينا ! لقد فزت بمزاد {auction_title}",
			"auction_lost":   "لللأسف، لم تفز بمزاد {auction_title}",
			"payment_received": "تم استلام مبلغ الدفع بنجاح",
			"auction_started": "بدأ المزاد {auction_title}",
			"auction_ended":   "انتهى المزاد {auction_title}",
		},
		"fr": {
			"bid_placed":    "Nouvelle enchère placée sur {auction_title}",
			"bid_outbid":    "Votre enchère sur {auction_title} a été dépassée",
			"auction_won":   "Félicitations ! Vous avez remporté l'enchère {auction_title}",
			"auction_lost":   "Désolé, vous n'avez pas remporté l'enchère {auction_title}",
			"payment_received": "Paiement reçu avec succès",
			"auction_started": "L'enchère {auction_title} a commencé",
			"auction_ended":   "L'enchère {auction_title} est terminée",
		},
		"en": {
			"bid_placed":    "New bid placed on {auction_title}",
			"bid_outbid":    "Your bid on {auction_title} has been outbid",
			"auction_won":   "Congratulations! You won auction {auction_title}",
			"auction_lost":   "Sorry, you didn't win auction {auction_title}",
			"payment_received": "Payment received successfully",
			"auction_started": "Auction {auction_title} has started",
			"auction_ended":   "Auction {auction_title} has ended",
		},
	}

	if langTemplates, exists := templates[lang]; exists {
		return OK(c, langTemplates)
	}

	return OK(c, templates["ar"]) // Default to Arabic
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
