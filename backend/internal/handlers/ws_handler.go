package handlers

import (
    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/websocket/v2"
    "github.com/google/uuid"
    ws "github.com/mazadpay/backend/internal/websocket"
    "go.uber.org/zap"
)

type WSHandler struct {
    hub    *ws.Hub
    logger *zap.Logger
}

func NewWSHandler(hub *ws.Hub, logger *zap.Logger) *WSHandler {
    return &WSHandler{hub: hub, logger: logger}
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

// HandleAuction — point d'entrée WebSocket : GET /ws/auction/:id
func (h *WSHandler) HandleAuction(c *websocket.Conn) {
    auctionID, err := uuid.Parse(c.Params("id"))
    if err != nil {
        c.Close()
        return
    }

    // userID depuis query param (token WS)
    // En prod : valider le JWT depuis ?token=xxx
    userID := c.Query("user_id", "anonymous")

    client := ws.NewClient(c, userID, h.logger)
    h.hub.Register(auctionID, client)
    defer h.hub.Unregister(auctionID, client)

    go client.WritePump()
    client.ReadPump() // bloquant jusqu'à déconnexion
}
