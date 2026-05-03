package ws

import (
    "encoding/json"
    "sync"

    "github.com/google/uuid"
    "github.com/mazadpay/backend/internal/models"
    "go.uber.org/zap"
)

// Hub gère les rooms WebSocket (une room par enchère)
type Hub struct {
    rooms  map[uuid.UUID]map[*Client]bool
    mu     sync.RWMutex
    logger *zap.Logger
}

func NewHub(logger *zap.Logger) *Hub {
    return &Hub{
        rooms:  make(map[uuid.UUID]map[*Client]bool),
        logger: logger,
    }
}

func (h *Hub) Register(auctionID uuid.UUID, client *Client) {
    h.mu.Lock()
    defer h.mu.Unlock()
    if h.rooms[auctionID] == nil {
        h.rooms[auctionID] = make(map[*Client]bool)
    }
    h.rooms[auctionID][client] = true
    h.logger.Info("WS client joined", zap.String("auction", auctionID.String()))
}

func (h *Hub) Unregister(auctionID uuid.UUID, client *Client) {
    h.mu.Lock()
    defer h.mu.Unlock()
    if room, ok := h.rooms[auctionID]; ok {
        delete(room, client)
        if len(room) == 0 {
            delete(h.rooms, auctionID)
        }
    }
    client.conn.Close()
}

// Broadcast envoie un événement à tous les clients d'une room
func (h *Hub) Broadcast(auctionID uuid.UUID, event models.WSEvent) {
    h.mu.RLock()
    defer h.mu.RUnlock()

    payload, err := json.Marshal(event)
    if err != nil {
        return
    }

    room, ok := h.rooms[auctionID]
    if !ok {
        return
    }

    for client := range room {
        select {
        case client.send <- payload:
        default:
            // Buffer plein : déconnecter le client silencieusement
            close(client.send)
            delete(room, client)
        }
    }
}

// BroadcastToUser envoie un événement uniquement à un utilisateur spécifique (ex: auction_won)
func (h *Hub) BroadcastToUser(auctionID uuid.UUID, userID string, event models.WSEvent) {
    h.mu.RLock()
    defer h.mu.RUnlock()

    payload, _ := json.Marshal(event)
    room, ok := h.rooms[auctionID]
    if !ok {
        return
    }
    for client := range room {
        if client.userID == userID {
            select {
            case client.send <- payload:
            default:
            }
        }
    }
}

// ConnectedCount retourne le nombre de connexions actives pour une room
func (h *Hub) ConnectedCount(auctionID uuid.UUID) int {
    h.mu.RLock()
    defer h.mu.RUnlock()
    return len(h.rooms[auctionID])
}
