package ws

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/models"
	"go.uber.org/zap"
)

// ChatClientInterface définit l'interface pour un client de chat
type ChatClientInterface interface {
	GetUserID() uuid.UUID
	Send(message models.WebSocketMessage) error
	GetConversations() []uuid.UUID
}

// ChatServiceInterface définit l'interface minimale pour le chat service
type ChatServiceInterface interface {
	MarkMessagesAsRead(ctx context.Context, conversationID, userID uuid.UUID, messageID *uuid.UUID) error
}

// ChatHub gère les connexions WebSocket pour le chat
type ChatHub struct {
	clients    map[uuid.UUID]*ChatClient
	rooms      map[uuid.UUID]map[uuid.UUID]*ChatClient // conversationID -> userID -> client
	register   chan *ChatClient
	unregister chan *ChatClient
	broadcast  chan models.WebSocketMessage
	mu         sync.RWMutex
	logger     *zap.Logger
	chatSvc    ChatServiceInterface
}

// ChatClient représente un client WebSocket connecté
type ChatClient struct {
	hub           *ChatHub
	conn          *websocket.Conn
	send          chan models.WebSocketMessage
	userID        uuid.UUID
	conversations map[uuid.UUID]bool
	mu            sync.RWMutex
}

func NewChatHub(logger *zap.Logger, chatSvc ChatServiceInterface) *ChatHub {
	return &ChatHub{
		clients:    make(map[uuid.UUID]*ChatClient),
		rooms:      make(map[uuid.UUID]map[uuid.UUID]*ChatClient),
		register:   make(chan *ChatClient),
		unregister: make(chan *ChatClient),
		broadcast:  make(chan models.WebSocketMessage, 256),
		logger:     logger,
		chatSvc:    chatSvc,
	}
}

// SetChatService permet de définir le service de chat après création (pour éviter la dépendance circulaire)
func (h *ChatHub) SetChatService(chatSvc ChatServiceInterface) {
	h.chatSvc = chatSvc
}

// Run démarre le hub
func (h *ChatHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.userID] = client
			h.mu.Unlock()
			h.logger.Info("client registered", zap.String("user_id", client.userID.String()))

			// Notifier les autres que l'utilisateur est en ligne
			h.BroadcastToAll(models.ChatEventUserOnline, map[string]string{
				"user_id": client.userID.String(),
			})

		case client := <-h.unregister:
			h.mu.Lock()
			current, exists := h.clients[client.userID]
			if exists && current == client {
				delete(h.clients, client.userID)
				h.mu.Unlock()
				close(client.send)
			} else {
				h.mu.Unlock()
			}
			h.logger.Info("client unregistered", zap.String("user_id", client.userID.String()))

			// Retirer des rooms (éviter deadlock en copiant les IDs d'abord)
			client.mu.RLock()
			convIDs := make([]uuid.UUID, 0, len(client.conversations))
			for convID := range client.conversations {
				convIDs = append(convIDs, convID)
			}
			client.mu.RUnlock()
			for _, convID := range convIDs {
				h.LeaveConversation(client.userID, convID, client)
			}

			// Notifier les autres que l'utilisateur est hors ligne
			h.BroadcastToAll(models.ChatEventUserOffline, map[string]string{
				"user_id": client.userID.String(),
			})

		case message := <-h.broadcast:
			h.handleBroadcast(message)
		}
	}
}

// HandleWebSocket gère une connexion WebSocket
func (h *ChatHub) HandleWebSocket(userID uuid.UUID) fiber.Handler {
	return websocket.New(func(c *websocket.Conn) {
		client := &ChatClient{
			hub:           h,
			conn:          c,
			send:          make(chan models.WebSocketMessage, 256),
			userID:        userID,
			conversations: make(map[uuid.UUID]bool),
		}

		h.register <- client

		var wg sync.WaitGroup
		wg.Add(1)

		// Goroutine pour envoyer les messages au client
		go func() {
			defer wg.Done()
			client.writePump()
		}()

		// Lecture des messages du client (synchrone)
		client.readPump()

 		wg.Wait()

 		c.Close()
	})
}

// WebSocketUpgrader vérifie si on peut upgrader en WebSocket
func (h *ChatHub) WebSocketUpgrader() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	}
}

func (c *ChatClient) readPump() {
	defer func() {
		c.hub.unregister <- c
	}()

	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			c.hub.logger.Error("websocket read error", 
				zap.Error(err), 
				zap.String("user_id", c.userID.String()))
			break
		}

		// Traiter le message
		var wsMsg models.WebSocketMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			c.hub.logger.Error("invalid websocket message", zap.Error(err))
			continue
		}

		c.handleMessage(wsMsg)
	}
}

func (c *ChatClient) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteJSON(message); err != nil {
				c.hub.logger.Error("failed to write message", zap.Error(err))
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *ChatClient) handleMessage(msg models.WebSocketMessage) {
	switch msg.Event {
	case models.ChatEventConversationJoin:
		if msg.ConversationID != uuid.Nil {
			c.hub.JoinConversation(c.userID, msg.ConversationID, c)
		}

	case models.ChatEventConversationLeave:
		if msg.ConversationID != uuid.Nil {
			c.hub.LeaveConversation(c.userID, msg.ConversationID, c)
		}

	case models.ChatEventTypingStart, models.ChatEventTypingStop:
		if msg.ConversationID != uuid.Nil {
			c.hub.BroadcastToConversation(msg.ConversationID, msg.Event, msg.Data)
		}

	case models.ChatEventMessageRead:
		// Mettre à jour le statut de lecture
		if msg.ConversationID != uuid.Nil {
			data, ok := msg.Data.(map[string]interface{})
			if ok {
				ctx := context.Background()
				if msgID, ok := data["message_id"].(string); ok {
					msgUUID, _ := uuid.Parse(msgID)
					c.hub.chatSvc.MarkMessagesAsRead(ctx, msg.ConversationID, c.userID, &msgUUID)
				} else {
					c.hub.chatSvc.MarkMessagesAsRead(ctx, msg.ConversationID, c.userID, nil)
				}
			}
		}
	}
}

// RegisterClient enregistre un client
func (h *ChatHub) RegisterClient(userID uuid.UUID, client ChatClientInterface) {
	// Convertir l'interface en *ChatClient si possible
	if c, ok := client.(*ChatClient); ok {
		h.register <- c
	}
}

// UnregisterClient désenregistre un client
func (h *ChatHub) UnregisterClient(userID uuid.UUID, client ChatClientInterface) {
	if c, ok := client.(*ChatClient); ok {
		h.unregister <- c
	}
}

// BroadcastToUser envoie un message à un utilisateur spécifique
func (h *ChatHub) BroadcastToUser(userID uuid.UUID, event string, data interface{}) {
	h.mu.RLock()
	client, ok := h.clients[userID]
	h.mu.RUnlock()

	if ok {
		msg := models.WebSocketMessage{
			Event:     event,
			Data:      data,
			Timestamp: time.Now(),
		}
		select {
		case client.send <- msg:
		default:
			close(client.send)
			h.mu.Lock()
			delete(h.clients, userID)
			h.mu.Unlock()
		}
	}
}

// BroadcastToConversation envoie un message à tous les participants d'une conversation
func (h *ChatHub) BroadcastToConversation(conversationID uuid.UUID, event string, data interface{}) {
	h.mu.RLock()
	room, ok := h.rooms[conversationID]
	h.mu.RUnlock()

	if !ok {
		return
	}

	msg := models.WebSocketMessage{
		Event:          event,
		ConversationID: conversationID,
		Data:           data,
		Timestamp:      time.Now(),
	}

	for _, client := range room {
		select {
		case client.send <- msg:
		default:
			close(client.send)
			h.mu.Lock()
			delete(h.clients, client.userID)
			h.mu.Unlock()
		}
	}
}

// BroadcastToAll envoie un message à tous les clients connectés
func (h *ChatHub) BroadcastToAll(event string, data interface{}) {
	h.mu.RLock()
	clients := make([]*ChatClient, 0, len(h.clients))
	for _, client := range h.clients {
		clients = append(clients, client)
	}
	h.mu.RUnlock()

	msg := models.WebSocketMessage{
		Event:     event,
		Data:      data,
		Timestamp: time.Now(),
	}

	for _, client := range clients {
		select {
		case client.send <- msg:
		default:
			close(client.send)
			h.mu.Lock()
			delete(h.clients, client.userID)
			h.mu.Unlock()
		}
	}
}

// JoinConversation ajoute un utilisateur à une conversation (room)
func (h *ChatHub) JoinConversation(userID, conversationID uuid.UUID, client ChatClientInterface) {
	c, ok := client.(*ChatClient)
	if !ok {
		return
	}

	h.mu.Lock()
	if _, ok := h.rooms[conversationID]; !ok {
		h.rooms[conversationID] = make(map[uuid.UUID]*ChatClient)
	}
	h.rooms[conversationID][userID] = c
	h.mu.Unlock()

	c.mu.Lock()
	c.conversations[conversationID] = true
	c.mu.Unlock()

	h.logger.Info("user joined conversation",
		zap.String("user_id", userID.String()),
		zap.String("conversation_id", conversationID.String()))
}

// LeaveConversation retire un utilisateur d'une conversation
func (h *ChatHub) LeaveConversation(userID, conversationID uuid.UUID, client ChatClientInterface) {
	h.mu.Lock()
	if room, ok := h.rooms[conversationID]; ok {
		delete(room, userID)
		if len(room) == 0 {
			delete(h.rooms, conversationID)
		}
	}
	h.mu.Unlock()

	if c, ok := client.(*ChatClient); ok {
		c.mu.Lock()
		delete(c.conversations, conversationID)
		c.mu.Unlock()
	}

	h.logger.Info("user left conversation",
		zap.String("user_id", userID.String()),
		zap.String("conversation_id", conversationID.String()))
}

// IsUserOnline vérifie si un utilisateur est en ligne
func (h *ChatHub) IsUserOnline(userID uuid.UUID) bool {
	h.mu.RLock()
	_, ok := h.clients[userID]
	h.mu.RUnlock()
	return ok
}

// GetOnlineUsers retourne la liste des utilisateurs en ligne
func (h *ChatHub) GetOnlineUsers() []uuid.UUID {
	h.mu.RLock()
	users := make([]uuid.UUID, 0, len(h.clients))
	for userID := range h.clients {
		users = append(users, userID)
	}
	h.mu.RUnlock()
	return users
}

func (h *ChatHub) handleBroadcast(message models.WebSocketMessage) {
	// Implémentation si nécessaire pour des broadcasts spécifiques
}

// GetUserID retourne l'ID de l'utilisateur
func (c *ChatClient) GetUserID() uuid.UUID {
	return c.userID
}

// Send envoie un message au client
func (c *ChatClient) Send(message models.WebSocketMessage) error {
	select {
	case c.send <- message:
		return nil
	default:
		return http.ErrAbortHandler
	}
}

// GetConversations retourne les conversations du client
func (c *ChatClient) GetConversations() []uuid.UUID {
	c.mu.RLock()
	convs := make([]uuid.UUID, 0, len(c.conversations))
	for convID := range c.conversations {
		convs = append(convs, convID)
	}
	c.mu.RUnlock()
	return convs
}
