package handlers

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/repository"
	"github.com/mazadpay/backend/internal/services"
	ws "github.com/mazadpay/backend/internal/websocket"
	"go.uber.org/zap"
)

// GlobalWSHandler serves Customer Request #20's authenticated mobile-wide
// channel (GET /ws/global): list/content invalidation events for a connected
// client that is not currently inside a specific auction's /ws/auction/:id
// room (Home, Favorites, My Auctions, FAQ, Banners, Categories). Structurally
// a near-copy of WSHandler (same JWT-via-query-param auth, same
// Client/WritePump/ReadPump plumbing) but registers with GlobalHub instead
// of joining an auction room, and stamps the connecting user's market onto
// the client so GlobalHub.BroadcastAuctionEvent can filter by it.
type GlobalWSHandler struct {
	hub      *ws.GlobalHub
	authSvc  services.AuthService
	userRepo repository.UserRepository
	logger   *zap.Logger
}

func NewGlobalWSHandler(hub *ws.GlobalHub, authSvc services.AuthService, userRepo repository.UserRepository, logger *zap.Logger) *GlobalWSHandler {
	return &GlobalWSHandler{hub: hub, authSvc: authSvc, userRepo: userRepo, logger: logger}
}

// UpgradeMiddleware mirrors WSHandler.UpgradeMiddleware/AdminWSHandler.UpgradeMiddleware.
func (h *GlobalWSHandler) UpgradeMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	}
}

// HandleGlobal -- point d'entrée WebSocket : GET /ws/global?token=JWT
func (h *GlobalWSHandler) HandleGlobal(c *websocket.Conn) {
	token := c.Query("token", "")
	if token == "" {
		h.logger.Warn("Global WS: missing JWT token")
		c.WriteMessage(websocket.TextMessage, []byte(`{"error": "missing token"}`))
		c.Close()
		return
	}

	claims, err := h.authSvc.ValidateJWT(token)
	if err != nil {
		h.logger.Error("Global WS: invalid JWT token", zap.Error(err))
		c.WriteMessage(websocket.TextMessage, []byte(`{"error": "invalid token"}`))
		c.Close()
		return
	}
	userID := claims.UserID

	// Load the connecting user's own account to stamp their market onto the
	// client (GlobalHub.BroadcastAuctionEvent's filter) -- same trusted,
	// server-side lookup pattern as WSHandler.AuthorizeSubscription (never
	// client-supplied). A lookup failure closes the connection rather than
	// falling back to a wildcard/empty market, which would otherwise receive
	// every market's auction events.
	uid, err := uuid.Parse(userID)
	if err != nil {
		h.logger.Error("Global WS: invalid user id in token", zap.String("user_id", userID))
		c.WriteMessage(websocket.TextMessage, []byte(`{"error": "invalid token"}`))
		c.Close()
		return
	}
	user, err := h.userRepo.FindByID(context.Background(), uid)
	if err != nil {
		h.logger.Error("Global WS: user lookup failed", zap.String("user_id", userID), zap.Error(err))
		c.WriteMessage(websocket.TextMessage, []byte(`{"error": "unauthorized"}`))
		c.Close()
		return
	}

	h.logger.Info("Global WebSocket client connected", zap.String("user_id", userID))

	client := ws.NewClient(c, userID, h.logger)
	client.SetMarket(user.EffectiveAccountCountryISO())
	h.hub.Register(client)
	defer h.hub.Unregister(client)

	h.sendInitialState(c)

	go client.WritePump()
	client.ReadPump() // bloquant jusqu'à déconnexion
}

// sendInitialState mirrors WSHandler.sendInitialState -- confirms the
// connection to the client; carries no entity data (REST remains canonical).
func (h *GlobalWSHandler) sendInitialState(c *websocket.Conn) {
	initialState := models.GlobalWSEvent{
		Type:      "connected",
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	payload, _ := json.Marshal(initialState)
	c.WriteMessage(websocket.TextMessage, payload)
}
