package middleware

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"go.uber.org/zap"
)

// SetupMiddlewares initialise tous les middlewares optimisés pour Fly.io (économies bande passante, CPU, RAM)
func SetupMiddlewares(app *fiber.App, logger *zap.Logger) {

	// 1. COMPRESSION (Priorité #1 — économise ~70% de bande passante)
	app.Use(compress.New(compress.Config{
		// LevelBestSpeed est le compromis idéal: compression très correcte sans surcharger le vCPU partagé
		Level: compress.LevelBestSpeed,
	}))

	// 5. ETAG (Évite de re-télécharger des contenus JSON/HTML identiques)
	app.Use(etag.New())

	// 2. CACHE HEADERS (Mise en cache stricte des assets statiques)
	app.Use(func(c *fiber.Ctx) error {
		path := c.Path()
		
		// Les réponses API sont dynamiques et ne doivent jamais être mises en cache par le navigateur
		if strings.HasPrefix(path, "/api/") {
			c.Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
			c.Set("Pragma", "no-cache")
			c.Set("Expires", "0")
		} else if strings.HasPrefix(path, "/static/") || strings.HasPrefix(path, "/public/") {
			// Les assets statiques peuvent vivre 24h en cache
			c.Set("Cache-Control", "public, max-age=86400")
		}
		
		return c.Next()
	})

	// 3. RATE LIMITER (Protège Neon DB et évite le scraping abusif de bande passante)
	app.Use(limiter.New(limiter.Config{
		Max:        30, // 30 requêtes par IP
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			logger.Warn("Rate limit exceeded", zap.String("ip", c.IP()))
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Trop de requêtes, veuillez ralentir.",
			})
		},
	}))

	// 4. BODY LIMIT (Évite les attaques par gros payload JSON)
	app.Use(func(c *fiber.Ctx) error {
		// Tolérance pour les formulaires multipart (upload d'images vers R2 qui ont leur propre vérification de taille)
		contentType := c.Get("Content-Type")
		if strings.HasPrefix(strings.ToLower(contentType), "multipart/form-data") {
			return c.Next()
		}

		// 50KB max pour toute requête JSON entrante
		if c.Request().Header.ContentLength() > 50*1024 {
			logger.Warn("Payload too large rejected", 
				zap.Int("size", c.Request().Header.ContentLength()), 
				zap.String("ip", c.IP()),
			)
			return c.Status(fiber.StatusRequestEntityTooLarge).JSON(fiber.Map{
				"error": "payload too large",
			})
		}
		return c.Next()
	})

	// 6. LOGGER MINIMALISTE (Logue SEULEMENT les erreurs 5xx pour économiser les I/O Disque et le CPU)
	app.Use(func(c *fiber.Ctx) error {
		start := time.Now()

		// Passe au middleware suivant
		err := c.Next()

		status := c.Response().StatusCode()

		// Gestion des erreurs Fiber retournées directement
		if err != nil {
			if e, ok := err.(*fiber.Error); ok {
				status = e.Code
			} else {
				status = fiber.StatusInternalServerError
			}
		}

		// On ne loggue que si c'est une erreur serveur (>= 500)
		if status >= 500 {
			latency := time.Since(start)
			logger.Error("Erreur Serveur (5xx)",
				zap.Int("status", status),
				zap.String("method", c.Method()),
				zap.String("path", c.Path()),
				zap.Duration("latency", latency),
				zap.String("ip", c.IP()),
				zap.Error(err),
			)
		}

		return err
	})
}
