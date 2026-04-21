package handlers

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/services"
	ws "github.com/mazadpay/backend/internal/websocket"
	"go.uber.org/zap"
)

type AdminWSHandler struct {
	hub       *ws.AdminHub
	jwtSecret string
	logger    *zap.Logger
}

func NewAdminWSHandler(hub *ws.AdminHub, jwtSecret string, logger *zap.Logger) *AdminWSHandler {
	return &AdminWSHandler{hub: hub, jwtSecret: jwtSecret, logger: logger}
}

// UpgradeMiddleware vérifie que la requête peut être upgradée en WebSocket
func (h *AdminWSHandler) UpgradeMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	}
}

// HandleAdmin gère les connexions WebSocket admin : GET /ws/admin
func (h *AdminWSHandler) HandleAdmin(c *websocket.Conn) {
	// Récupérer le token JWT depuis le query param
	tokenStr := c.Query("token", "")

	var adminUUID uuid.UUID
	var adminRole string
	var isValid bool

	if tokenStr != "" {
		// Valider le token JWT
		token, err := jwt.ParseWithClaims(tokenStr, &services.JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.ErrUnauthorized
			}
			return []byte(h.jwtSecret), nil
		})

		if err == nil && token.Valid {
			claims := token.Claims.(*services.JWTClaims)
			// Vérifier que c'est un admin
			if strings.ToLower(claims.Role) == "admin" || strings.ToLower(claims.Role) == "super_admin" {
				parsedUID, err := uuid.Parse(claims.UserID)
				if err == nil {
					adminUUID = parsedUID
					adminRole = claims.Role
					isValid = true
				}
			}
		}
	}

	if !isValid {
		h.logger.Warn("Admin WebSocket: Invalid or missing JWT token")
		c.WriteMessage(websocket.TextMessage, []byte(`{"error": "unauthorized", "message": "Valid admin JWT token required"}`))
		c.Close()
		return
	}

	client := ws.NewClient(c, adminUUID.String(), h.logger)
	client.Role = adminRole
	h.hub.Register(client)
	defer h.hub.Unregister(client)

	h.logger.Info("Admin WebSocket client connected",
		zap.String("admin_id", adminUUID.String()),
		zap.String("role", adminRole))

	go client.WritePump()
	client.ReadPump() // bloquant jusqu'à déconnexion
}

// HandleAdminStats retourne les statistiques des connexions admin
func (h *AdminWSHandler) HandleAdminStats(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"connected_admins": h.hub.ConnectedCount(),
	})
}
