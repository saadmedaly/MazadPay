package ws

import (
	"encoding/json"
	"sync"

	"github.com/mazadpay/backend/internal/models"
	"go.uber.org/zap"
)

// GlobalHub is Customer Request #20's authenticated mobile-wide channel: it
// exists for list/content invalidation events (auction.created, faq.updated,
// banner.updated, category.updated, ...) that must reach a connected mobile
// client even when the user is not currently inside a specific auction's
// /ws/auction/:id room. It deliberately mirrors AdminHub's flat
// "clients map + mutex, broadcast to all" shape rather than Hub's per-auction
// room shape -- there is no per-entity room to join here, only a per-market
// filter applied at broadcast time (see BroadcastAuctionEvent) so a global
// auction event never reaches a client whose account market differs from the
// auction's market, mirroring the isolation already enforced by
// WSHandler.AuthorizeSubscription for the per-auction hub.
type GlobalHub struct {
	clients map[*Client]bool
	mu      sync.RWMutex
	logger  *zap.Logger
}

func NewGlobalHub(logger *zap.Logger) *GlobalHub {
	return &GlobalHub{
		clients: make(map[*Client]bool),
		logger:  logger,
	}
}

func (h *GlobalHub) Register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[client] = true
	h.logger.Info("Global WS client connected")
}

func (h *GlobalHub) Unregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, client)
	client.conn.Close()
	h.logger.Info("Global WS client disconnected")
}

// send is the shared non-blocking write used by every Broadcast* method: a
// full client buffer disconnects that one client rather than blocking (or
// dropping) the broadcast for everyone else, matching Hub.Broadcast/
// AdminHub.Broadcast's existing backpressure behavior.
func (h *GlobalHub) send(client *Client, payload []byte) {
	select {
	case client.send <- payload:
	default:
		close(client.send)
		delete(h.clients, client)
	}
}

// Broadcast sends event to every connected global client, with no market or
// user filtering. Reserved for content that is identical for every user
// regardless of market (FAQ, banners, categories) -- never for auction
// events, which must go through BroadcastAuctionEvent instead.
func (h *GlobalHub) Broadcast(event models.GlobalWSEvent) {
	h.mu.Lock()
	defer h.mu.Unlock()

	payload, err := json.Marshal(event)
	if err != nil {
		h.logger.Error("Failed to marshal global WS event", zap.Error(err))
		return
	}
	for client := range h.clients {
		h.send(client, payload)
	}
}

// BroadcastAuctionEvent sends an auction-scoped event only to clients whose
// account market matches marketCountryISO (the auction's own
// EffectiveMarketCountryISO()) -- the global-channel equivalent of the
// country-scoped market isolation WSHandler.AuthorizeSubscription already
// enforces per-auction-room. An empty client.market (should not happen once
// HandleGlobal stamps it from the authenticated user, but defensively
// checked) never matches and is silently skipped rather than treated as a
// wildcard.
func (h *GlobalHub) BroadcastAuctionEvent(marketCountryISO string, event models.GlobalWSEvent) {
	h.mu.Lock()
	defer h.mu.Unlock()

	payload, err := json.Marshal(event)
	if err != nil {
		h.logger.Error("Failed to marshal global WS auction event", zap.Error(err))
		return
	}
	for client := range h.clients {
		if client.market != "" && client.market == marketCountryISO {
			h.send(client, payload)
		}
	}
}

// BroadcastToUser sends event only to the connected global client(s)
// belonging to userID -- for owner-only events (e.g. a future
// request-status change) that must never be visible to any other user.
func (h *GlobalHub) BroadcastToUser(userID string, event models.GlobalWSEvent) {
	h.mu.Lock()
	defer h.mu.Unlock()

	payload, err := json.Marshal(event)
	if err != nil {
		h.logger.Error("Failed to marshal global WS user event", zap.Error(err))
		return
	}
	for client := range h.clients {
		if client.userID == userID {
			h.send(client, payload)
		}
	}
}

// ConnectedCount returns the number of active global connections.
func (h *GlobalHub) ConnectedCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
