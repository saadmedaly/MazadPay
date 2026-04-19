package middleware

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RateLimit crée un middleware de rate limiting basé sur Redis
// windowSeconds : durée de la fenêtre en secondes
// maxAttempts : nombre maximal de tentatives par fenêtre
func RateLimit(rdb *redis.Client, windowSeconds, maxAttempts int, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Clé basée sur l'IP de l'utilisateur
		ip := c.IP()
		path := c.Path()
		key := fmt.Sprintf("ratelimit:%s:%s", ip, path)

		// Vérifier le compteur
		count, err := rdb.Incr(c.Context(), key).Result()
		if err != nil {
			logger.Error("RateLimit Redis error", zap.Error(err))
			// En cas d'erreur Redis, laisser passer mais logger
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
					"message": "Too many requests, please try again later",
				},
			})
		}

		return c.Next()
	}
}

// RateLimitByPhone crée un rate limit basé sur le numéro de téléphone (pour OTP/Login)
func RateLimitByPhone(rdb *redis.Client, windowSeconds, maxAttempts int, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extraire le téléphone de la requête
		var reqBody map[string]interface{}
		if err := c.BodyParser(&reqBody); err != nil {
			return c.Next()
		}

		phone, ok := reqBody["phone"].(string)
		if !ok || phone == "" {
			return c.Next()
		}

		path := c.Path()
		key := fmt.Sprintf("ratelimit:phone:%s:%s", phone, path)

		// Vérifier le compteur
		count, err := rdb.Incr(c.Context(), key).Result()
		if err != nil {
			logger.Error("RateLimit Redis error", zap.Error(err))
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
					"message": "Too many requests, please try again later",
				},
			})
		}

		return c.Next()
	}
}
