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


