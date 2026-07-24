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

		// Révocation au niveau utilisateur (Session Security Phase 1) : rejette tout
		// JWT émis avant un blocage de compte ou une réinitialisation de PIN, même
		// s'il est par ailleurs valide (signature/exp corrects, non blacklisté
		// individuellement). fail-closed : si Redis est indisponible, on refuse la
		// requête plutôt que de laisser passer un token potentiellement révoqué —
		// cohérent avec le traitement fail-closed déjà appliqué aux routes
		// sensibles (Auth/OTP Rate Limiting Hardening). Risque assumé : une panne
		// Redis bloque alors tous les endpoints authentifiés, pas seulement l'auth ;
		// à surveiller en production.
		if rdb != nil {
			revokedBefore, rErr := rdb.Get(c.Context(), services.RevokedBeforeKey(uid)).Int64()
			if rErr != nil && rErr != redis.Nil {
				logger.Error("Auth failed: Redis error checking user revocation", zap.Error(rErr), zap.String("path", c.Path()))
				return c.Status(503).JSON(fiber.Map{
					"success": false,
					"error":   fiber.Map{"code": "service_unavailable", "message": "الخدمة مشغولة حالياً، يرجى المحاولة لاحقاً"},
				})
			}
			// Comparaison stricte (<, pas <=) : IssuedAt et revokedBefore sont tous deux
			// tronqués à la seconde (Unix()), donc un nouveau login effectué la même
			// seconde qu'une révocation (ex: ResetPassword suivi immédiatement d'un
			// nouveau Login avec le nouveau PIN) aurait pu avoir IssuedAt == revokedBefore
			// et être rejeté à tort avec un <=. RevokeUserSessions écrit revoked_before
			// avant que ResetPassword ne retourne, donc tout token émis par un Login
			// postérieur — même à la même seconde — est légitimement postérieur en temps
			// réel malgré l'égalité au niveau de la précision seconde ; < laisse passer
			// ce cas tout en rejetant strictement tout ce qui est antérieur.
			if rErr == nil && claims.IssuedAt != nil && claims.IssuedAt.Unix() < revokedBefore {
				logger.Warn("Auth failed: Token issued before user-level revocation", zap.String("path", c.Path()), zap.String("user_id", uid.String()))
				return c.Status(401).JSON(fiber.Map{
					"success": false,
					"error":   fiber.Map{"code": "unauthorized", "message": "Token has been invalidated (logged out)"},
				})
			}
		}

		c.Locals("user_id", uid)
		c.Locals("user_role", claims.Role)
		c.Locals("is_super_admin", claims.IsSuperAdmin)

		return c.Next()
	}
}

// OptionalJWT decodes the JWT if present and sets user_id in Locals, but never rejects the request.
// Use for endpoints that are public but benefit from knowing the authenticated user.
func OptionalJWT(jwtSecret string, rdb *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		tokenStr := ""
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
		} else {
			tokenStr = c.Query("token")
		}
		if tokenStr == "" {
			return c.Next()
		}
		if rdb != nil {
			blacklisted, _ := rdb.Exists(c.Context(), fmt.Sprintf("blacklist:%s", tokenStr)).Result()
			if blacklisted > 0 {
				return c.Next()
			}
		}
		token, err := jwt.ParseWithClaims(tokenStr, &services.JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(jwtSecret), nil
		})
		if err != nil || !token.Valid {
			return c.Next()
		}
		if claims, ok := token.Claims.(*services.JWTClaims); ok {
			if uid, err := uuid.Parse(claims.UserID); err == nil {
				c.Locals("user_id", uid)
				c.Locals("user_role", claims.Role)
			}
		}
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
