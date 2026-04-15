package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/mazadpay/backend/internal/config"
	"github.com/mazadpay/backend/internal/database"
	"github.com/mazadpay/backend/internal/routes"
	"go.uber.org/zap"
)

func main() {
 	logger, _ := zap.NewProduction()
	defer logger.Sync()

 	cfg := config.Load()

	db, err := database.NewPostgres(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to connect to PostgreSQL", zap.Error(err))
	}
	defer db.Close()

	rdb, err := database.NewRedis(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	defer rdb.Close()

	app := fiber.New(fiber.Config{
		AppName:      cfg.App.Name,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			// Log l'erreur imprévue
			logger.Error("Unhandled request error",
				zap.String("path", c.Path()),
				zap.String("method", c.Method()),
				zap.Error(err),
			)
			return c.Status(500).JSON(fiber.Map{
				"success": false,
				"error":   fiber.Map{"code": "server_error", "message": "An internal error occurred"},
			})
		},
	})

	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"status": "ok"}})
	})
	routes.Setup(app, db, rdb, cfg, logger)

	// Démarrage gracieux
	go func() {
		if err := app.Listen(":" + cfg.App.Port); err != nil {
			logger.Fatal("Server failed", zap.Error(err))
		}
	}()

	logger.Info("Server started", zap.String("port", cfg.App.Port))

	// Arrêt gracieux sur signal OS
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = ctx
	app.Shutdown()
	logger.Info("Server stopped")
}
