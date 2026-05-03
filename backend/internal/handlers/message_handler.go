package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/middleware"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/services"
	"go.uber.org/zap"
)

type MessageHandler struct {
	chatSvc services.ChatService
	logger  *zap.Logger
}

func NewMessageHandler(chatSvc services.ChatService, logger *zap.Logger) *MessageHandler {
	return &MessageHandler{
		chatSvc: chatSvc,
		logger:  logger,
	}
}

func (h *MessageHandler) SendMessage(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c)
	}

	conversationID, err := uuid.Parse(c.Params("conversation_id"))
	if err != nil {
		return BadRequest(c, "Invalid conversation ID")
	}

	var req models.SendMessageRequest
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	// Validation
	if req.Type == "" {
		return BadRequest(c, "Message type is required")
	}

	// Pour les messages texte, le contenu est requis
	if req.Type == "text" && (req.Content == nil || *req.Content == "") {
		return BadRequest(c, "Content is required for text messages")
	}

	// Pour les fichiers, l'URL est requise
	if (req.Type == "audio" || req.Type == "video" || req.Type == "image" || req.Type == "file") &&
		(req.FileURL == nil || *req.FileURL == "") {
		return BadRequest(c, "File URL is required for media messages")
	}

	message, err := h.chatSvc.SendMessage(c.Context(), conversationID, userID, &req)
	if err != nil {
		return MapError(c, h.logger, err)
	}

	return Created(c, message)
}

// GetMessages récupère les messages d'une conversation
func (h *MessageHandler) GetMessages(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c)
	}

	conversationID, err := uuid.Parse(c.Params("conversation_id"))
	if err != nil {
		return BadRequest(c, "Invalid conversation ID")
	}

	limit := c.QueryInt("limit", 50)
	offset := c.QueryInt("offset", 0)

	messages, err := h.chatSvc.GetMessages(c.Context(), conversationID, userID, limit, offset)
	if err != nil {
		return MapError(c, h.logger, err)
	}

	// Marquer les messages comme lus en arrière-plan avec timeout
	if len(messages) > 0 {
		lastMessageID := messages[0].ID
		// Utiliser un context avec timeout pour éviter les goroutines orphelines
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			h.chatSvc.MarkMessagesAsRead(ctx, conversationID, userID, &lastMessageID)
		}()
	}

	return OK(c, messages)
}

// EditMessage modifie un message
func (h *MessageHandler) EditMessage(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c)
	}

	messageID, err := uuid.Parse(c.Params("message_id"))
	if err != nil {
		return BadRequest(c, "Invalid message ID")
	}

	var req struct {
		Content string `json:"content"`
	}

	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	if req.Content == "" {
		return BadRequest(c, "Content is required")
	}

	message, err := h.chatSvc.EditMessage(c.Context(), messageID, userID, req.Content)
	if err != nil {
		return MapError(c, h.logger, err)
	}

	return OK(c, message)
}

// DeleteMessage supprime un message
func (h *MessageHandler) DeleteMessage(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c)
	}

	messageID, err := uuid.Parse(c.Params("message_id"))
	if err != nil {
		return BadRequest(c, "Invalid message ID")
	}

	if err := h.chatSvc.DeleteMessage(c.Context(), messageID, userID); err != nil {
		return MapError(c, h.logger, err)
	}

	return OK(c, fiber.Map{"message": "Message deleted"})
}

// MarkAsRead marque les messages comme lus
func (h *MessageHandler) MarkAsRead(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c)
	}

	conversationID, err := uuid.Parse(c.Params("conversation_id"))
	if err != nil {
		return BadRequest(c, "Invalid conversation ID")
	}

	var req models.MarkReadRequest
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	if err := h.chatSvc.MarkMessagesAsRead(c.Context(), conversationID, userID, req.MessageID); err != nil {
		return MapError(c, h.logger, err)
	}

	return OK(c, fiber.Map{"message": "Messages marked as read"})
}
