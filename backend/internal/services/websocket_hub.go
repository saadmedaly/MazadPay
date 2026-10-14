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

// GlobalHub -- Customer Request #20's mobile-wide invalidation channel
// interface (see internal/websocket/global_hub.go for the concrete
// implementation). Interfaced the same way as AuctionHub/AdminHub above so
// adminService/auctionService/contentService stay unit-testable with a
// no-op double, without needing a real *ws.GlobalHub.
type GlobalHub interface {
	Broadcast(event models.GlobalWSEvent)
	BroadcastAuctionEvent(marketCountryISO string, event models.GlobalWSEvent)
	BroadcastToUser(userID string, event models.GlobalWSEvent)
}


