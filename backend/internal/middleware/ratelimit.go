package middleware

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RateLimitSearchQuery applique un rate limit IP plus strict uniquement quand le
// paramètre de requête donné (ex: "q") est non vide — utilisé pour limiter la
// recherche sur GET /auctions sans pénaliser le listing simple (Public Endpoints /
// Scraping Protection). fail-open : ce n'est pas une route sensible (financière/auth).
func RateLimitSearchQuery(rdb *redis.Client, queryParam string, windowSeconds, maxAttempts int, logger *zap.Logger) fiber.Handler {
	inner := RateLimit(rdb, windowSeconds, maxAttempts, logger, false)
	return func(c *fiber.Ctx) error {
		if c.Query(queryParam) == "" {
			return c.Next()
		}
		return inner(c)
	}
}

// CacheControl pose un en-tête Cache-Control public sur des endpoints publics peu
// changeants (catégories, locations, banners, sponsors, contenu statique) pour réduire
// la charge Backend — voir Public Endpoints / Scraping Protection. À ne jamais poser
// sur les mazads/enchères actifs (voir routes.go) pour ne pas retarder l'affichage des
// mises à jour de prix/statut.
func CacheControl(maxAgeSeconds int) fiber.Handler {
	header := fmt.Sprintf("public, max-age=%d", maxAgeSeconds)
	return func(c *fiber.Ctx) error {
		c.Set("Cache-Control", header)
		return c.Next()
	}
}

// failClosedResponse est renvoyée quand Redis est indisponible sur un chemin
// sensible (auth/OTP) : on refuse la requête plutôt que de laisser passer sans
// contrôle (durcissement sécurité — voir Auth/OTP Rate Limiting Hardening).
func failClosedResponse(c *fiber.Ctx) error {
	return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
		"success": false,
		"error": fiber.Map{
			"code":    "rate_limit_unavailable",
			"message": "الخدمة مشغولة حالياً، يرجى المحاولة لاحقاً",
		},
	})
}

// RateLimit crée un middleware de rate limiting basé sur Redis, avec clé IP+path.
// windowSeconds : durée de la fenêtre en secondes
// maxAttempts : nombre maximal de tentatives par fenêtre
// failClosed : si true, une erreur Redis fait échouer la requête (503) au lieu de
// la laisser passer — à utiliser uniquement sur les routes auth/OTP sensibles.
func RateLimit(rdb *redis.Client, windowSeconds, maxAttempts int, logger *zap.Logger, failClosed bool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Clé basée sur l'IP de l'utilisateur
		ip := c.IP()
		path := c.Path()
		key := fmt.Sprintf("ratelimit:%s:%s", ip, path)

		// Vérifier le compteur
		count, err := rdb.Incr(c.Context(), key).Result()
		if err != nil {
			logger.Error("RateLimit Redis error", zap.Error(err))
			if failClosed {
				return failClosedResponse(c)
			}
			// En cas d'erreur Redis, laisser passer mais logger (routes non sensibles uniquement)
			return c.Next()
		}

		// Définir l'expiration la première fois
		if count == 1 {
			rdb.Expire(c.Context(), key, time.Duration(windowSeconds)*time.Second)
		}

		// Vérifier si dépassé
		if count > int64(maxAttempts) {
			logger.Warn("Rate limit exceeded",
				zap.String("ip", ip),
				zap.String("path", path),
				zap.Int64("attempts", count),
			)
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "rate_limited",
					"message": "محاولات كثيرة جداً، يرجى المحاولة لاحقاً",
				},
			})
		}

		return c.Next()
	}
}

// RateLimitByUser crée un rate limit basé sur l'utilisateur authentifié (user_id posé
// dans c.Locals par JWT()), pour protéger les routes wallet/bids sensibles (Wallet +
// Bids Rate Limiting). extraParam, si non vide, est le nom d'un paramètre de route
// (ex: "id" pour l'auction_id ou le transaction_id) dont la valeur est ajoutée à la clé
// — permet un compteur par (user_id, auction_id) plutôt qu'un compteur global par
// utilisateur. failClosed=true : erreur Redis ou utilisateur non authentifié (ne
// devrait pas arriver après jwtMiddleware, mais défense en profondeur) => 503, jamais
// un passage sans contrôle sur ces routes financières/sensibles.
func RateLimitByUser(rdb *redis.Client, windowSeconds, maxAttempts int, logger *zap.Logger, failClosed bool, extraParam string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		uid, ok := c.Locals("user_id").(uuid.UUID)
		if !ok {
			if failClosed {
				return failClosedResponse(c)
			}
			return c.Next()
		}

		path := c.Path()
		key := fmt.Sprintf("ratelimit:user:%s:%s", uid.String(), path)
		if extraParam != "" {
			if extraVal := c.Params(extraParam); extraVal != "" {
				key = fmt.Sprintf("%s:%s", key, extraVal)
			}
		}

		count, err := rdb.Incr(c.Context(), key).Result()
		if err != nil {
			logger.Error("RateLimitByUser Redis error", zap.Error(err))
			if failClosed {
				return failClosedResponse(c)
			}
			return c.Next()
		}

		if count == 1 {
			rdb.Expire(c.Context(), key, time.Duration(windowSeconds)*time.Second)
		}

		if count > int64(maxAttempts) {
			logger.Warn("Rate limit exceeded for user",
				zap.String("user_id", uid.String()),
				zap.String("path", path),
				zap.Int64("attempts", count),
			)
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "rate_limited",
					"message": "محاولات كثيرة جداً، يرجى المحاولة لاحقاً",
				},
			})
		}

		return c.Next()
	}
}

// RateLimitByPhone crée un rate limit basé sur le numéro de téléphone (pour OTP/Login).
// failClosed : si true (routes auth/OTP sensibles), une erreur Redis ou un corps de
// requête illisible fait échouer la requête (503) au lieu de la laisser passer sans
// contrôle — durcissement sécurité (voir Auth/OTP Rate Limiting Hardening).
func RateLimitByPhone(rdb *redis.Client, windowSeconds, maxAttempts int, logger *zap.Logger, failClosed bool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extraire le téléphone de la requête
		var reqBody map[string]interface{}
		if err := c.BodyParser(&reqBody); err != nil {
			if failClosed {
				return failClosedResponse(c)
			}
			return c.Next()
		}

		phone, ok := reqBody["phone"].(string)
		if !ok || phone == "" {
			if failClosed {
				return failClosedResponse(c)
			}
			return c.Next()
		}

		path := c.Path()
		key := fmt.Sprintf("ratelimit:phone:%s:%s", phone, path)

		// Vérifier le compteur
		count, err := rdb.Incr(c.Context(), key).Result()
		if err != nil {
			logger.Error("RateLimit Redis error", zap.Error(err))
			if failClosed {
				return failClosedResponse(c)
			}
			return c.Next()
		}

		// Définir l'expiration la première fois
		if count == 1 {
			rdb.Expire(c.Context(), key, time.Duration(windowSeconds)*time.Second)
		}

		// Vérifier si dépassé
		if count > int64(maxAttempts) {
			logger.Warn("Rate limit exceeded for phone",
				zap.String("phone", phone),
				zap.String("path", path),
				zap.Int64("attempts", count),
			)
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "rate_limited",
					"message": "محاولات كثيرة جداً، يرجى المحاولة لاحقاً",
				},
			})
		}

		return c.Next()
	}
}
