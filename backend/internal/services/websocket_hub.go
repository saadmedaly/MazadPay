package services

import (
	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/models"
)

// AuctionHub définit l'interface pour la gestion WebSocket des enchères
type AuctionHub interface {
	Broadcast(auctionID uuid.UUID, event models.WSEvent)
	BroadcastToUser(auctionID uuid.UUID, userID string, event models.WSEvent)
}

// AdminHub définit l'interface pour la gestion WebSocket des admins
type AdminHub interface {
	Broadcast(event models.AdminEvent)
	BroadcastNewRequest(requestType string, payload models.NewRequestPayload)
	BroadcastRequestUpdated(payload models.RequestUpdatedPayload)
}

// WebSocketHub définit l'interface pour la gestion WebSocket (général)
type WebSocketHub interface {
	// Gestion des clients
	RegisterClient(userID uuid.UUID, client ChatClient)
	UnregisterClient(userID uuid.UUID, client ChatClient)
	
	// Diffusion
	BroadcastToUser(userID uuid.UUID, event string, data interface{})
	BroadcastToConversation(conversationID uuid.UUID, event string, data interface{})
	BroadcastToAll(event string, data interface{})
	
	// Gestion des rooms
	JoinConversation(userID, conversationID uuid.UUID, client ChatClient)
	LeaveConversation(userID, conversationID uuid.UUID, client ChatClient)
	
	// Présence
	IsUserOnline(userID uuid.UUID) bool
	GetOnlineUsers() []uuid.UUID
}

// ChatClient représente un client WebSocket connecté
type ChatClient interface {
	GetUserID() uuid.UUID
	Send(message models.WebSocketMessage) error
	GetConversations() []uuid.UUID
}
