package middleware

import (
 	"errors"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/services"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func JWT(jwtSecret string, logger *zap.Logger, rdb *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		tokenStr := ""

		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
		} else {
 			tokenStr = c.Query("token")
		}

		if tokenStr == "" {
			logger.Warn("Auth failed: Missing token", zap.String("path", c.Path()))
			return c.Status(401).JSON(fiber.Map{
				"success": false,
				"error":   fiber.Map{"code": "unauthorized", "message": "Missing token"},
			})
		}

		// Vérifier si le token est dans la blacklist Redis
		if rdb != nil {
			blacklisted, err := rdb.Exists(c.Context(), fmt.Sprintf("blacklist:%s", tokenStr)).Result()
			if err == nil && blacklisted > 0 {
				logger.Warn("Auth failed: Token blacklisted", zap.String("path", c.Path()))
				return c.Status(401).JSON(fiber.Map{
					"success": false,
					"error":   fiber.Map{"code": "unauthorized", "message": "Token has been invalidated (logged out)"},
				})
			}
		}

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
		c.Locals("is_super_admin", claims.IsSuperAdmin)

		return c.Next()
	}
}

func AdminOnly(logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, ok := c.Locals("user_role").(string)
		if !ok || (strings.ToLower(role) != "admin" && strings.ToLower(role) != "super_admin") {
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

func SuperAdminOnly(logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		isSuperAdmin, ok := c.Locals("is_super_admin").(bool)
		if !ok || !isSuperAdmin {
			logger.Warn("Access denied: Super Admin required",
				zap.String("path", c.Path()),
			)
			return c.Status(403).JSON(fiber.Map{
				"success": false,
				"error":   fiber.Map{"code": "forbidden", "message": "Super Admin access required"},
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
