package ws

import (
	"encoding/json"
	"sync"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// AdminEvent — événements envoyés aux admins en temps réel
type AdminEvent struct {
	Type    string      `json:"type"` // new_request | request_updated | request_reviewed | new_bid | auction_ended
	Payload interface{} `json:"payload"`
}

// NewRequestPayload — nouvelle demande reçue
type NewRequestPayload struct {
	RequestID   string `json:"request_id"`
	RequestType string `json:"request_type"` // auction | banner
	UserID      string `json:"user_id"`
	UserName    string `json:"user_name"`
	Title       string `json:"title"`
	CreatedAt   string `json:"created_at"`
}

// RequestUpdatedPayload — demande mise à jour
type RequestUpdatedPayload struct {
	RequestID   string `json:"request_id"`
	RequestType string `json:"request_type"`
	Status      string `json:"status"`
	UpdatedBy   string `json:"updated_by"`
	UpdatedAt   string `json:"updated_at"`
}

// AdminHub gère les connexions WebSocket des admins
type AdminHub struct {
	clients map[*Client]bool
	mu      sync.RWMutex
	logger  *zap.Logger
}

// AdminClient — client WebSocket admin
type AdminClient struct {
	*Client
	adminID uuid.UUID
}

func NewAdminHub(logger *zap.Logger) *AdminHub {
	return &AdminHub{
		clients: make(map[*Client]bool),
		logger:  logger,
	}
}

func (h *AdminHub) Register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[client] = true
	h.logger.Info("Admin connected via WebSocket")
}

func (h *AdminHub) Unregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, client)
	client.conn.Close()
	h.logger.Info("Admin disconnected from WebSocket")
}

// Broadcast envoie un événement à tous les admins connectés
func (h *AdminHub) Broadcast(event AdminEvent) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	payload, err := json.Marshal(event)
	if err != nil {
		h.logger.Error("Failed to marshal admin event", zap.Error(err))
		return
	}

	for client := range h.clients {
		select {
		case client.send <- payload:
		default:
			// Buffer plein : déconnecter le client
			close(client.send)
			delete(h.clients, client)
		}
	}
}

// BroadcastNewRequest notifie tous les admins d'une nouvelle demande
func (h *AdminHub) BroadcastNewRequest(requestType string, payload NewRequestPayload) {
	h.Broadcast(AdminEvent{
		Type:    "new_request",
		Payload: payload,
	})
}

// BroadcastRequestUpdated notifie tous les admins d'une mise à jour de demande
func (h *AdminHub) BroadcastRequestUpdated(payload RequestUpdatedPayload) {
	h.Broadcast(AdminEvent{
		Type:    "request_updated",
		Payload: payload,
	})
}

// ConnectedCount retourne le nombre d'admins connectés
func (h *AdminHub) ConnectedCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
