package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/jmoiron/sqlx"
	"github.com/mazadpay/backend/internal/config"
	"github.com/mazadpay/backend/internal/handlers"
	"github.com/mazadpay/backend/internal/middleware"
	"github.com/mazadpay/backend/internal/repository"
	"github.com/mazadpay/backend/internal/services"
	ws "github.com/mazadpay/backend/internal/websocket"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Setup enregistre toutes les routes.
func Setup(app *fiber.App, db *sqlx.DB, rdb *redis.Client, cfg *config.Config, logger *zap.Logger) {
	// Repositories
	userRepo := repository.NewUserRepository(db)
	auctionRepo := repository.NewAuctionRepository(db)
	bidRepo := repository.NewBidRepository(db)
	walletRepo := repository.NewWalletRepository(db)

	// Hub
	hub := ws.NewHub(logger)

	// Services
	authSvc := services.NewAuthService(userRepo, cfg.JWT.Secret, cfg.JWT.ExpiryHours)
	auctionSvc := services.NewAuctionService(auctionRepo)
	bidSvc := services.NewBidService(db, auctionRepo, bidRepo, walletRepo, hub)
	userSvc := services.NewUserService(userRepo)


	api := app.Group("/v1/api")

	// Handlers
	wsHandler := handlers.NewWSHandler(hub, logger)
	bidHandler := handlers.NewBidHandler(bidSvc, logger)
	userHandler := handlers.NewUserHandler(userSvc, logger)



	// WebSocket registration
	app.Use("/ws", wsHandler.UpgradeMiddleware())
	app.Get("/ws/auction/:id", websocket.New(wsHandler.HandleAuction))

	// Routes registration
	setupAuthRoutes(api, authSvc, cfg.JWT.Secret, logger)
	setupAuctionRoutes(api, auctionSvc, bidHandler, cfg.JWT.Secret, logger)
	setupUserRoutes(api, userHandler, cfg.JWT.Secret)
	setupUserRoutes(api, userHandler, cfg.JWT.Secret)

	// More routes will be added in subsequent steps
}

func setupAuthRoutes(api fiber.Router, authSvc services.AuthService, jwtSecret string, logger *zap.Logger) {
	jwtMiddleware := middleware.JWT(jwtSecret)
	h := handlers.NewAuthHandler(authSvc, logger)


	auth := api.Group("/auth")

	// Public routes
	auth.Post("/register", h.Register)
	auth.Post("/login", h.Login)
	auth.Post("/otp/send", h.SendOTP)
	auth.Post("/otp/verify", h.VerifyOTP)
	auth.Post("/reset-password", h.ResetPassword)

	// Protected routes
	auth.Post("/logout", jwtMiddleware, h.Logout)
	auth.Put("/change-password", jwtMiddleware, h.ChangePassword)
}

func setupAuctionRoutes(api fiber.Router, auctionSvc services.AuctionService, bidHandler *handlers.BidHandler, jwtSecret string, logger *zap.Logger) {
	jwtMiddleware := middleware.JWT(jwtSecret)
	h := handlers.NewAuctionHandler(auctionSvc, logger)


	// Public routes
	api.Get("/categories", h.GetCategories)
	api.Get("/locations", h.GetLocations)
	api.Get("/auctions", h.List)
	api.Get("/auctions/:id", h.GetByID)
	api.Post("/auctions/:id/view", h.IncrementView)

	// Bids (Public history)
	api.Get("/auctions/:id/bids", bidHandler.History)

	// Protected routes
	auctions := api.Group("/auctions", jwtMiddleware)
	auctions.Post("/", h.Create)

	// Bids (Protected place)
	api.Post("/auctions/:id/bids", jwtMiddleware, bidHandler.Place)

	// Seller Contact (CONCEPTION F3.7)
	api.Get("/auctions/:id/seller-contact", jwtMiddleware, h.GetSellerContact)
}

func setupUserRoutes(api fiber.Router, userHandler *handlers.UserHandler, jwtSecret string) {
	jwtMiddleware := middleware.JWT(jwtSecret)
	users := api.Group("/users", jwtMiddleware)

	users.Get("/me", userHandler.GetMe)
}

