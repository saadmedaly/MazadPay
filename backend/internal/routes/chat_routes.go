package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mazadpay/backend/internal/handlers"
	"github.com/mazadpay/backend/internal/middleware"
	"github.com/mazadpay/backend/internal/services"
	ws "github.com/mazadpay/backend/internal/websocket"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

 func setupChatRoutes(
	api fiber.Router,
	chatSvc services.ChatService,
	chatHub *ws.ChatHub,
	jwtSecret string,
	logger *zap.Logger,
	rdb *redis.Client,
) {
	jwtMiddleware := middleware.JWT(jwtSecret, logger, rdb)

	convHandler := handlers.NewConversationHandler(chatSvc, logger)
	msgHandler := handlers.NewMessageHandler(chatSvc, logger)

	// WebSocket endpoint pour le chat
	chat := api.Group("/chat")
	chat.Use("/ws", chatHub.WebSocketUpgrader())
	chat.Get("/ws", jwtMiddleware, func(c *fiber.Ctx) error {
		userID, err := middleware.GetUserID(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
		}
		handler := chatHub.HandleWebSocket(userID)
		return handler(c)
	})

	// Routes REST pour les conversations
	conversations := api.Group("/conversations", jwtMiddleware)
	conversations.Get("/", convHandler.GetConversations)
	conversations.Get("/support", middleware.AdminOnly(logger), convHandler.GetSupportConversations)
	conversations.Post("/", convHandler.CreateConversation)
	conversations.Get("/direct/:user_id", convHandler.GetDirectConversation)
	conversations.Get("/:id", convHandler.GetConversation)
	conversations.Post("/:id/join", convHandler.JoinConversation)
	conversations.Post("/:id/leave", convHandler.LeaveConversation)

	// Routes REST pour les messages
	conversations.Get("/:conversation_id/messages", msgHandler.GetMessages)
	conversations.Post("/:conversation_id/messages", msgHandler.SendMessage)
	conversations.Post("/:conversation_id/read", msgHandler.MarkAsRead)

	// Routes pour les messages individuels
	api.Put("/messages/:message_id", jwtMiddleware, msgHandler.EditMessage)
	api.Delete("/messages/:message_id", jwtMiddleware, msgHandler.DeleteMessage)
}
