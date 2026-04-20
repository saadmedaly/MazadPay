package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberLogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/mazadpay/backend/internal/config"
	"github.com/mazadpay/backend/internal/database"
	"github.com/mazadpay/backend/internal/repository"
	"github.com/mazadpay/backend/internal/routes"
	"go.uber.org/zap"
)

func main() {
	println("--- Starting MazadPay Backend ---")

	cfg := config.Load()

	// Valider la configuration critique
	if err := cfg.Validate(); err != nil {
		println("Configuration error:", err.Error())
		os.Exit(1)
	}

	var logger *zap.Logger
	var err error
	if cfg.App.Env == "development" {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}
	if err != nil {
		println("Failed to initialize logger:", err.Error())
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Connecting to PostgreSQL...", zap.String("host", cfg.DB.Host))
	db, err := database.NewPostgres(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to connect to PostgreSQL", zap.Error(err))
	}
	defer db.Close()

	logger.Info("Connecting to Redis...", zap.String("url", cfg.Redis.URL))
	rdb, err := database.NewRedis(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	defer rdb.Close()

	app := fiber.New(fiber.Config{
		AppName:      cfg.App.Name,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}

			// Ne pas loguer les erreurs bruyantes du client (404, 405, 408) comme des ERROR
			if code == fiber.StatusNotFound || code == fiber.StatusMethodNotAllowed || code == fiber.StatusRequestTimeout {
				logger.Debug("Client request info",
					zap.String("path", c.Path()),
					zap.String("method", c.Method()),
					zap.Int("status", code),
					zap.Error(err),
				)
			} else {
				logger.Error("Request error",
					zap.String("path", c.Path()),
					zap.String("method", c.Method()),
					zap.Int("status", code),
					zap.Error(err),
				)
			}

			return c.Status(code).JSON(fiber.Map{
				"success": false,
				"error":   fiber.Map{"code": "error", "message": err.Error()},
			})
		},
	})

	app.Use(recover.New())
	app.Use(fiberLogger.New(fiberLogger.Config{
		Format: "[${time}] ${status} - ${latency} ${method} ${path}\n",
	}))
	app.Use(cors.New())

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"status": "ok"}})
	})

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "MazadPay API root"})
	})

	auctionSvc, notifSvc := routes.Setup(app, db, rdb, cfg, logger)

	// --- Seed Default Super Admin ---
	if cfg.App.DefaultSuperAdminPhone != "" && cfg.App.DefaultSuperAdminPin != "" {
		userRepo := repository.NewUserRepository(db)
		if err := userRepo.SeedDefaultSuperAdmin(context.Background(), 
			cfg.App.DefaultSuperAdminPhone, 
			cfg.App.DefaultSuperAdminPin,
			"Super Admin",
			"admin@mazadpay.com",
		); err != nil {
			logger.Error("Failed to seed default super admin", zap.Error(err))
		} else {
			logger.Info("Default super admin seeded successfully")
		}
	}

	// --- Background Tasks ---
	go func() {
		auctionTicker := time.NewTicker(30 * time.Second)
		cleanupTicker := time.NewTicker(1 * time.Hour)
		defer auctionTicker.Stop()
		defer cleanupTicker.Stop()

		for {
			select {
			case <-auctionTicker.C:
				logger.Debug("Running background: CloseExpiredAuctions")
				if err := auctionSvc.CloseExpiredAuctions(context.Background()); err != nil {
					logger.Error("Failed to close expired auctions", zap.Error(err))
				}
			case <-cleanupTicker.C:
				logger.Info("Running background: CleanupOldNotifications")
				if err := notifSvc.CleanupOldNotifications(context.Background()); err != nil {
					logger.Error("Failed to cleanup old notifications", zap.Error(err))
				}
			}
		}
	}()

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

	// Graceful shutdown timeout adapté selon l'environnement
	shutdownTimeout := 30 * time.Second
	if cfg.App.Env == "development" {
		shutdownTimeout = 10 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		// "context deadline exceeded" est normal = timeout atteint, fermeture forcée
		if err == context.DeadlineExceeded {
			logger.Info("Graceful shutdown timeout reached, forcing closure", zap.Duration("timeout", shutdownTimeout))
		} else {
			logger.Error("Error during shutdown", zap.Error(err))
		}
	}
	logger.Info("Server stopped")
}
