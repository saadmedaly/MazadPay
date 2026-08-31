package handlers

import (
	"context"
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/repository"
	"github.com/mazadpay/backend/internal/services"
	ws "github.com/mazadpay/backend/internal/websocket"
	"go.uber.org/zap"
)

type WSHandler struct {
	hub         *ws.Hub
	authSvc     services.AuthService
	auctionRepo repository.AuctionRepository
	userRepo    repository.UserRepository
	logger      *zap.Logger
}

func NewWSHandler(hub *ws.Hub, authSvc services.AuthService, auctionRepo repository.AuctionRepository, userRepo repository.UserRepository, logger *zap.Logger) *WSHandler {
	return &WSHandler{hub: hub, authSvc: authSvc, auctionRepo: auctionRepo, userRepo: userRepo, logger: logger}
}


// UpgradeMiddleware vérifie que la requête peut être upgradée en WebSocket
func (h *WSHandler) UpgradeMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	}
}

// HandleAuction — point d'entrée WebSocket : GET /ws/auction/:id?token=JWT
// Authentification JWT via query param pour les clients mobiles
func (h *WSHandler) HandleAuction(c *websocket.Conn) {
	auctionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		h.logger.Error("Invalid auction ID", zap.Error(err))
		c.WriteMessage(websocket.TextMessage, []byte(`{"error": "invalid auction id"}`))
		c.Close()
		return
	}

	// JWT auth via query param
	token := c.Query("token", "")
	if token == "" {
		h.logger.Warn("Missing JWT token")
		c.WriteMessage(websocket.TextMessage, []byte(`{"error": "missing token"}`))
		c.Close()
		return
	}

	// Validate JWT and extract userID
	userID, err := h.validateJWT(token)
	if err != nil {
		h.logger.Error("Invalid JWT token", zap.Error(err))
		c.WriteMessage(websocket.TextMessage, []byte(`{"error": "invalid token"}`))
		c.Close()
		return
	}

	// Country-scoped market isolation (migration 000046, V1) -- see
	// AuthorizeSubscription's doc comment for the full rationale.
	if ok := h.AuthorizeSubscription(context.Background(), auctionID, userID); !ok {
		h.logger.Info("WebSocket cross-market or invalid subscription rejected",
			zap.String("auction_id", auctionID.String()), zap.String("user_id", userID))
		c.WriteMessage(websocket.TextMessage, []byte(`{"error": "not found"}`))
		c.Close()
		return
	}

	h.logger.Info("WebSocket client connected",
		zap.String("auction_id", auctionID.String()),
		zap.String("user_id", userID))

	client := ws.NewClient(c, userID, h.logger)
	h.hub.Register(auctionID, client)

	// Envoyer l'état initial de l'enchère
	h.sendInitialState(c, auctionID)

	defer h.hub.Unregister(auctionID, client)

	go client.WritePump()
	client.ReadPump() // bloquant jusqu'à déconnexion
}

// AuthorizeSubscription enforces country-scoped market isolation (migration 000046, V1)
// for a WebSocket auction subscription: the subscriber must belong to the same market
// as the auction, checked by COUNTRY equality only -- never by currency (SN/CI both use
// XOF but are separate markets, see bid_service.go PlaceBid for the same rule enforced
// at bid time). Both the auction and the subscriber's account are loaded from trusted
// server-side sources (auctionRepo, userRepo keyed by the JWT-derived userID) -- nothing
// here is client-supplied. Returns false on any lookup failure (not-found auction,
// malformed/unknown user id) or on a market mismatch -- callers must reject the
// connection silently (no "wrong market" detail) on false, consistent with the REST
// detail/bid-history endpoints. Exported (not just inlined in HandleAuction) so it can
// be exercised directly by tests without a live WebSocket connection.
func (h *WSHandler) AuthorizeSubscription(ctx context.Context, auctionID uuid.UUID, userIDStr string) bool {
	auction, err := h.auctionRepo.FindByID(ctx, auctionID)
	if err != nil {
		return false
	}
	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return false
	}
	user, err := h.userRepo.FindByID(ctx, uid)
	if err != nil {
		return false
	}
	return user.EffectiveAccountCountryISO() == auction.EffectiveMarketCountryISO()
}

// validateJWT extrait et valide le userID depuis le token JWT
func (h *WSHandler) validateJWT(tokenString string) (string, error) {
	claims, err := h.authSvc.ValidateJWT(tokenString)
	if err != nil {
		return "", err
	}
	return claims.UserID, nil
}


// sendInitialState envoie l'état actuel de l'enchère au client
func (h *WSHandler) sendInitialState(c *websocket.Conn, auctionID uuid.UUID) {
	initialState := models.WSEvent{
		Type: "initial_state",
		Payload: map[string]interface{}{
			"auction_id": auctionID.String(),
			"message":    "Connected to auction room",
		},
	}

	payload, _ := json.Marshal(initialState)
	c.WriteMessage(websocket.TextMessage, payload)
}
