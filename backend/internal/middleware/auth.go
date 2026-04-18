package middleware

import (
	"errors"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/services"
	"go.uber.org/zap"
)

func JWT(jwtSecret string, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			logger.Warn("Auth failed: Missing or malformed token", zap.String("path", c.Path()))
			return c.Status(401).JSON(fiber.Map{
				"success": false,
				"error":   fiber.Map{"code": "unauthorized", "message": "Missing token"},
			})
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.ParseWithClaims(tokenStr, &services.JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.ErrUnauthorized
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			logger.Warn("Auth failed: Invalid or expired token", zap.Error(err), zap.String("path", c.Path()))
			return c.Status(401).JSON(fiber.Map{
				"success": false,
				"error":   fiber.Map{"code": "unauthorized", "message": "Invalid or expired token"},
			})
		}

		claims := token.Claims.(*services.JWTClaims)
		
		uid, err := uuid.Parse(claims.UserID)
		if err != nil {
			logger.Error("Auth failed: Invalid UUID in token", zap.Error(err))
			return c.Status(401).JSON(fiber.Map{
				"success": false,
				"error":   fiber.Map{"code": "unauthorized", "message": "Invalid token content"},
			})
		}

		c.Locals("user_id", uid)
		c.Locals("user_role", claims.Role)

		return c.Next()
	}
}

func AdminOnly(logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, ok := c.Locals("user_role").(string)
		if !ok || strings.ToLower(role) != "admin" {
			logger.Warn("Access denied: Admin role required", 
				zap.String("path", c.Path()), 
				zap.String("user_role", role),
			)
			return c.Status(403).JSON(fiber.Map{
				"success": false,
				"error":   fiber.Map{"code": "forbidden", "message": "Admin access required"},
			})
		}
		return c.Next()
	}
}

// GetUserID extrait l'UUID de l'utilisateur depuis le contexte Fiber
func GetUserID(c *fiber.Ctx) (uuid.UUID, error) {
	uid, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		// Tentative avec l'autre clé au cas où
		uid, ok = c.Locals("userID").(uuid.UUID)
	}
	
	if !ok {
		return uuid.Nil, errors.New("user_id not found in context or invalid type")
	}
	return uid, nil
}
