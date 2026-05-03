package ws

import (
	"encoding/json"
	"sync"

	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/models"
	"go.uber.org/zap"
)

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
func (h *AdminHub) Broadcast(event models.AdminEvent) {
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
func (h *AdminHub) BroadcastNewRequest(requestType string, payload models.NewRequestPayload) {
	h.Broadcast(models.AdminEvent{
		Type:    "new_request",
		Payload: payload,
	})
}

// BroadcastRequestUpdated notifie tous les admins d'une mise à jour de demande
func (h *AdminHub) BroadcastRequestUpdated(payload models.RequestUpdatedPayload) {
	h.Broadcast(models.AdminEvent{
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
