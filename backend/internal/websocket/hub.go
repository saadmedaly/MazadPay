package ws

import (
    "encoding/json"
    "sync"

    "github.com/google/uuid"
    "go.uber.org/zap"
)

// WSEvent — tous les événements WebSocket envoyés aux clients
type WSEvent struct {
    Type    string      `json:"type"`    // bid_placed | timer_tick | auction_ended | auction_won
    Payload interface{} `json:"payload"`
}

type BidPlacedPayload struct {
    AuctionID    string  `json:"auction_id"`
    NewPrice     float64 `json:"new_price"`
    BidderMasked string  `json:"bidder_phone"` // "####4709"
    BidCount     int     `json:"bid_count"`
    SecondsLeft  int64   `json:"seconds_left"`
}

type TimerTickPayload struct {
    AuctionID   string `json:"auction_id"`
    SecondsLeft int64  `json:"seconds_left"`
}

type AuctionEndedPayload struct {
    AuctionID   string  `json:"auction_id"`
    FinalPrice  float64 `json:"final_price"`
    WinnerPhone string  `json:"winner_phone"` // masqué
}

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
func (h *Hub) Broadcast(auctionID uuid.UUID, event WSEvent) {
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
func (h *Hub) BroadcastToUser(auctionID uuid.UUID, userID string, event WSEvent) {
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
