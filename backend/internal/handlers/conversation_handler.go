package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/middleware"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/services"
  	"go.uber.org/zap"
)

type ConversationHandler struct {
	chatSvc services.ChatService
	logger  *zap.Logger
}

func NewConversationHandler(chatSvc services.ChatService, logger *zap.Logger) *ConversationHandler {
	return &ConversationHandler{
		chatSvc: chatSvc,
		logger:  logger,
	}
}

// CreateConversation crée une nouvelle conversation
func (h *ConversationHandler) CreateConversation(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("Unauthorized"))
	}

	var req models.CreateConversationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Invalid request body"))
	}

	// Validation
	if req.Type == "" || len(req.UserIDs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Type and user_ids are required"))
	}

	conversation, err := h.chatSvc.CreateConversation(c.Context(), &req, userID)
	if err != nil {
		h.logger.Error("failed to create conversation", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusCreated).JSON(SuccessResponse(conversation))
}

// GetConversations récupère les conversations de l'utilisateur
func (h *ConversationHandler) GetConversations(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("Unauthorized"))
	}

	limit := c.QueryInt("limit", 20)
	offset := c.QueryInt("offset", 0)

	role, _ := c.Locals("user_role").(string)
	var conversations []models.UserConversation

	if role == "admin" || role == "super_admin" {
		conversations, err = h.chatSvc.GetAdminConversations(c.Context(), userID, limit, offset)
	} else {
		conversations, err = h.chatSvc.GetUserConversations(c.Context(), userID, limit, offset)
	}

	if err != nil {
		h.logger.Error("failed to get conversations", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse("Failed to get conversations"))
	}

	return c.JSON(SuccessResponse(conversations))
}

// GetConversation récupère une conversation spécifique
func (h *ConversationHandler) GetConversation(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("Unauthorized"))
	}

	conversationID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Invalid conversation ID"))
	}

	conversation, participants, err := h.chatSvc.GetConversation(c.Context(), conversationID, userID)
	if err != nil {
		h.logger.Error("failed to get conversation", zap.Error(err))
		return c.Status(fiber.StatusForbidden).JSON(ErrorResponse("Conversation not found or access denied"))
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse(fiber.Map{
		"conversation": conversation,
		"participants": participants,
	}))
}

// JoinConversation permet à un utilisateur de rejoindre une conversation
func (h *ConversationHandler) JoinConversation(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("Unauthorized"))
	}

	conversationID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Invalid conversation ID"))
	}

	if err := h.chatSvc.JoinConversation(c.Context(), conversationID, userID); err != nil {
		h.logger.Error("failed to join conversation", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.JSON(SuccessResponse("Joined conversation"))
}

// LeaveConversation permet à un utilisateur de quitter une conversation
func (h *ConversationHandler) LeaveConversation(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("Unauthorized"))
	}

	conversationID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Invalid conversation ID"))
	}

	if err := h.chatSvc.LeaveConversation(c.Context(), conversationID, userID); err != nil {
		h.logger.Error("failed to leave conversation", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.JSON(SuccessResponse("Left conversation"))
}

// GetDirectConversation récupère ou crée une conversation directe
func (h *ConversationHandler) GetDirectConversation(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("Unauthorized"))
	}

	targetUserID, err := uuid.Parse(c.Params("user_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("Invalid user ID"))
	}

	conversation, err := h.chatSvc.GetOrCreateDirectConversation(c.Context(), userID, targetUserID)
	if err != nil {
		h.logger.Error("failed to get or create direct conversation", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.JSON(SuccessResponse(conversation))
}
// GetSupportConversations récupère toutes les conversations de type support (admin uniquement)
func (h *ConversationHandler) GetSupportConversations(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 20)
	offset := c.QueryInt("offset", 0)

	conversations, err := h.chatSvc.GetSupportConversations(c.Context(), limit, offset)
	if err != nil {
		h.logger.Error("failed to get support conversations", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse("Failed to get support conversations"))
	}

	return c.JSON(SuccessResponse(conversations))
}
