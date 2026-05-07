package middleware

import (
	"context"
	"sync/atomic"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	MaxDbCallsPerUserPerHour_FreeTier = 200 // Zone verte
	MaxDbCallsPerUserPerHour_WarnTier = 100 // Zone orange
)

// CheckDbQuota protège les 191.9h Compute/mois de Neon contre les abus individuels
func CheckDbQuota(pool *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Supposons que userID a été injecté par le middleware d'authentification
		userIDVal := c.Locals("userID")
		if userIDVal == nil {
			// Si pas connecté, on passe (ou on peut refuser selon l'architecture)
			return c.Next()
		}
		
		var userID int64
		switch v := userIDVal.(type) {
		case int64:
			userID = v
		case float64:
			userID = int64(v)
		case int:
			userID = int64(v)
		default:
			return c.Next()
		}

		var count int
		err := pool.QueryRow(c.Context(), `
			SELECT calls_count FROM user_api_quota
			WHERE user_id = $1
			  AND window_start > NOW() - INTERVAL '1 hour'
		`, userID).Scan(&count)

		// On ignore l'erreur NoRows car l'utilisateur n'a peut-être pas encore de quota créé
		if err != nil && err.Error() != "no rows in result set" {
			// En cas d'erreur DB autre que "pas de ligne", on laisse passer pour ne pas bloquer l'app
			return c.Next()
		}

		// Choix de la limite dynamique selon la charge serveur globale
		activeUsers := atomic.LoadInt64(&ActiveUsersCount)
		limit := MaxDbCallsPerUserPerHour_FreeTier
		if activeUsers > 500 {
			limit = MaxDbCallsPerUserPerHour_WarnTier
		}

		// Rejet si quota dépassé
		if count >= limit {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "quota_exceeded",
				"message": "Quota d'opérations horaires atteint pour maintenir la gratuité du service, réessayez dans 1h",
				"quota":   limit,
				"used":    count,
			})
		}

		// Incrémenter le compteur (UPSERT asynchrone idéalement, mais on le fait inline ici pour la simplicité)
		go func(uid int64) {
			_, _ = pool.Exec(context.Background(), `
				INSERT INTO user_api_quota (user_id, calls_count, window_start)
				VALUES ($1, 1, NOW())
				ON CONFLICT (user_id) DO UPDATE
				  SET calls_count = CASE
					  WHEN user_api_quota.window_start < NOW() - INTERVAL '1 hour'
					  THEN 1
					  ELSE user_api_quota.calls_count + 1
					END,
					window_start = CASE
					  WHEN user_api_quota.window_start < NOW() - INTERVAL '1 hour'
					  THEN NOW()
					  ELSE user_api_quota.window_start
					END
			`, uid)
		}(userID)

		return c.Next()
	}
}
