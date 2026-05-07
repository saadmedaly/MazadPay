package middleware

import (
	"sync/atomic"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// ActiveUsersCount est une variable globale (thread-safe) mise à jour par un job CRON
var ActiveUsersCount int64

// UpdateActiveUsersCount met à jour le nombre d'utilisateurs actifs
func UpdateActiveUsersCount(count int64) {
	atomic.StoreInt64(&ActiveUsersCount, count)
}

// AdaptiveRateLimit ajuste la limite de requêtes en fonction de la charge serveur
func AdaptiveRateLimit() fiber.Handler {
	return func(c *fiber.Ctx) error {
		activeUsers := atomic.LoadInt64(&ActiveUsersCount)
		var maxReq int

		switch {
		case activeUsers < 500:
			// Zone verte : Free Tier confortable
			maxReq = 60 // 60 req/min par IP
		case activeUsers < 2000:
			// Zone orange : Surveiller, réduire les appels
			maxReq = 30 // 30 req/min par IP (plus strict)
		default:
			// Zone rouge : Neon et Fly.io sous pression (Mode dégradé)
			maxReq = 15 // 15 req/min
		}

		// Appliquer le limiter dynamique
		return limiter.New(limiter.Config{
			Max:        maxReq,
			Expiration: 1 * time.Minute,
			KeyGenerator: func(c *fiber.Ctx) string {
				// Utiliser l'ID utilisateur s'il existe, sinon l'IP
				userID := c.Get("X-User-ID")
				if userID != "" {
					return userID
				}
				return c.IP()
			},
			LimitReached: func(c *fiber.Ctx) error {
				return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
					"error":       "rate_limit_exceeded",
					"retry_after": 60,
					"message":     "Trop de requêtes en période de forte affluence, réessayez dans 1 minute",
				})
			},
		})(c)
	}
}
