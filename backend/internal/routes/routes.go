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

// Setup enregistre toutes les routes et retourne les services nécessitant des tâches de fond.
func Setup(app *fiber.App, db *sqlx.DB, rdb *redis.Client, cfg *config.Config, logger *zap.Logger) (services.AuctionService, services.NotificationService) {
	// Repositories
	userRepo := repository.NewUserRepository(db)
	auctionRepo := repository.NewAuctionRepository(db)
	bidRepo := repository.NewBidRepository(db)
	walletRepo := repository.NewWalletRepository(db)
	txRepo := repository.NewTransactionRepository(db)
	reportRepo := repository.NewReportRepository(db)
	kycRepo := repository.NewKYCRepository(db)
	contentRepo := repository.NewContentRepository(db)
	favoriteRepo := repository.NewFavoriteRepository(db)
	notifRepo := repository.NewNotificationRepository(db)
	invRepo := repository.NewAdminInvitationRepository(db)

	// Hub
	hub := ws.NewHub(logger)

	// Services
	notifSvc := services.NewNotificationService(notifRepo, userRepo, cfg.Firebase.ServiceAccountPath)
	authSvc := services.NewAuthService(userRepo, cfg.JWT.Secret, cfg.JWT.ExpiryHours)
	auctionSvc := services.NewAuctionService(auctionRepo, reportRepo, notifSvc)
	bidSvc := services.NewBidService(db, auctionRepo, bidRepo, walletRepo, hub)
	userSvc := services.NewUserService(userRepo, favoriteRepo, auctionRepo, kycRepo)
	adminSvc := services.NewAdminService(db, userRepo, auctionRepo, bidRepo, txRepo, reportRepo, kycRepo, contentRepo, invRepo)
	walletSvc := services.NewWalletService(walletRepo, txRepo)
	contentSvc := services.NewContentService(contentRepo, notifSvc)

	api := app.Group("/v1/api")

	// Handlers
	wsHandler := handlers.NewWSHandler(hub, logger)
	bidHandler := handlers.NewBidHandler(bidSvc, logger)
	userHandler := handlers.NewUserHandler(userSvc, logger)
	adminHandler := handlers.NewAdminHandler(adminSvc, logger)
	bannerHandler := handlers.NewBannerHandler(contentSvc, logger)
	walletHandler := handlers.NewWalletHandler(walletSvc, logger)
	contentHandler := handlers.NewContentHandler(contentSvc, logger)
	notifHandler := handlers.NewNotificationHandler(notifSvc, logger)

	// WebSocket registration
	app.Use("/ws", wsHandler.UpgradeMiddleware())
	app.Get("/ws/auction/:id", websocket.New(wsHandler.HandleAuction))

	// Routes registration
	setupAuthRoutes(api, authSvc, adminHandler, cfg.JWT.Secret, logger)
	setupAuctionRoutes(api, auctionSvc, bidHandler, cfg.JWT.Secret, logger)
	setupUserRoutes(api, userHandler, walletHandler, cfg.JWT.Secret, logger)
	setupAdminRoutes(api, adminHandler, cfg.JWT.Secret, logger)
	setupBannerRoutes(api, bannerHandler, cfg.JWT.Secret, logger)
	setupContentRoutes(api, contentHandler, cfg.JWT.Secret, logger)
	setupNotificationRoutes(api, notifHandler, cfg.JWT.Secret, logger)

	return auctionSvc, notifSvc
}

func setupAuthRoutes(api fiber.Router, authSvc services.AuthService, adminHandler *handlers.AdminHandler, jwtSecret string, logger *zap.Logger) {
	jwtMiddleware := middleware.JWT(jwtSecret, logger)
	h := handlers.NewAuthHandler(authSvc, logger)

	auth := api.Group("/auth")

	// Public routes
	auth.Post("/register", h.Register)
	auth.Post("/login", h.Login)
	auth.Post("/otp/send", h.SendOTP)
	auth.Post("/otp/verify", h.VerifyOTP)
	auth.Post("/reset-password", h.ResetPassword)
	auth.Post("/register-admin", adminHandler.RegisterWithInvitation)

	// Protected routes
	auth.Post("/logout", jwtMiddleware, h.Logout)
	auth.Put("/change-password", jwtMiddleware, h.ChangePassword)
}

func setupAuctionRoutes(api fiber.Router, auctionSvc services.AuctionService, bidHandler *handlers.BidHandler, jwtSecret string, logger *zap.Logger) {
	jwtMiddleware := middleware.JWT(jwtSecret, logger)
	h := handlers.NewAuctionHandler(auctionSvc, logger)

	// Public routes
	api.Get("/categories", h.GetCategories)
	api.Get("/locations", h.GetLocations)
	api.Get("/auctions", h.List)
	api.Get("/auctions/:id", h.GetByID)
	api.Post("/auctions/:id/view", h.IncrementView)
	api.Get("/report-reasons", h.GetReportReasons)

	// Bids (Public history)
	api.Get("/auctions/:id/bids", bidHandler.History)

	// Protected routes
	auctions := api.Group("/auctions", jwtMiddleware)
	auctions.Post("/", h.Create)
	auctions.Post("/:id/report", h.Report) // CONCEPTION Signalements
	auctions.Post("/:id/images", h.AddImages)
	auctions.Post("/:id/buy-now", h.BuyNow)
	auctions.Post("/:id/cancel", h.Cancel)
	auctions.Post("/:id/relist", h.Relist)
	auctions.Post("/:id/extend", h.Extend)

	// Bids (Protected place)
	api.Post("/auctions/:id/bids", jwtMiddleware, bidHandler.Place)

	// Seller Contact (CONCEPTION F3.7)
	api.Get("/auctions/:id/seller-contact", jwtMiddleware, h.GetSellerContact)
}

func setupUserRoutes(api fiber.Router, userHandler *handlers.UserHandler, walletHandler *handlers.WalletHandler, jwtSecret string, logger *zap.Logger) {
	jwtMiddleware := middleware.JWT(jwtSecret, logger)
	users := api.Group("/users", jwtMiddleware)

	// Profile
	users.Get("/me", userHandler.GetMe)
	users.Put("/me", userHandler.UpdateProfile)
	users.Post("/me/avatar", userHandler.UpdateAvatar)
	users.Put("/me/language", userHandler.UpdateLanguage)
	users.Put("/me/notification-prefs", userHandler.UpdateNotificationPrefs)

	// Favorites
	users.Get("/me/favorites", userHandler.ListFavorites)
	users.Post("/me/favorites/:auction_id", userHandler.AddFavorite)
	users.Delete("/me/favorites/:auction_id", userHandler.RemoveFavorite)

	// Activity
	users.Get("/me/auctions", userHandler.MyAuctions)
	users.Get("/me/bids", userHandler.MyBids)
	users.Get("/me/winnings", userHandler.MyWinnings)

	// Wallet
	users.Get("/wallet", walletHandler.GetMe)
	users.Post("/wallet/deposit", walletHandler.Deposit)
	users.Post("/wallet/transactions/:id/receipt", walletHandler.UploadReceipt)
	users.Post("/wallet/withdraw", walletHandler.Withdraw)
	users.Get("/wallet/transactions", walletHandler.Transactions)
	users.Get("/wallet/transactions/:id", walletHandler.GetTransaction)

	// KYC
	users.Get("/kyc", userHandler.GetKYCStatus)
	users.Post("/kyc", userHandler.SubmitKYC)
}

func setupAdminRoutes(api fiber.Router, adminHandler *handlers.AdminHandler, jwtSecret string, logger *zap.Logger) {
	jwtMiddleware := middleware.JWT(jwtSecret, logger)
	adminMiddleware := middleware.AdminOnly(logger)

	admin := api.Group("/admin", jwtMiddleware, adminMiddleware)

	// Dashboard routes
	admin.Get("/dashboard/stats", adminHandler.DashboardStats)
	admin.Get("/dashboard/revenue-chart", adminHandler.RevenueChart)
	admin.Get("/dashboard/activity", adminHandler.ActivityFeed)

	// User management routes
	admin.Get("/users", adminHandler.ListUsers)
	admin.Get("/users/:id", adminHandler.GetUserByID)
	admin.Get("/users/:id/auctions", adminHandler.GetUserAuctions)
	admin.Get("/users/:id/transactions", adminHandler.GetUserTransactions)
	admin.Post("/invitations", adminHandler.GenerateInvitation)
	admin.Put("/users/:id/block", adminHandler.BlockUser)

	// Auction management routes
	admin.Get("/auctions", adminHandler.ListAuctions)
	admin.Put("/auctions/:id/validate", adminHandler.ValidateAuction)
	admin.Put("/auctions/:id", adminHandler.UpdateAuction)
	admin.Delete("/auctions/:id", adminHandler.DeleteAuction)

	// Additional management routes
	admin.Get("/transactions", adminHandler.ListTransactions)
	admin.Put("/transactions/:id/validate", adminHandler.ValidateTransaction)
	admin.Get("/reports", adminHandler.ListReports)
	admin.Put("/reports/:id/review", adminHandler.ReviewReport)

	// KYC management
	admin.Get("/kyc", adminHandler.ListKYCs)
	admin.Put("/kyc/:user_id", adminHandler.ReviewKYC)

	// Category management
	admin.Post("/categories", adminHandler.CreateCategory)
	admin.Put("/categories/:id", adminHandler.UpdateCategory)
	admin.Delete("/categories/:id", adminHandler.DeleteCategory)

	// Location management
	admin.Post("/locations", adminHandler.CreateLocation)
	admin.Put("/locations/:id", adminHandler.UpdateLocation)
	admin.Delete("/locations/:id", adminHandler.DeleteLocation)

	// Blocked phones management
	admin.Get("/blocked-phones", adminHandler.ListBlockedPhones)
	admin.Post("/blocked-phones", adminHandler.BlockPhone)
	admin.Delete("/blocked-phones/:phone", adminHandler.UnblockPhone)

	// Settings management
	admin.Get("/settings", adminHandler.ListSettings)
	admin.Put("/settings/:key", adminHandler.UpdateSetting)
}

func setupBannerRoutes(api fiber.Router, h *handlers.BannerHandler, jwtSecret string, logger *zap.Logger) {
	jwtMiddleware := middleware.JWT(jwtSecret, logger)
	adminMiddleware := middleware.AdminOnly(logger)

	// Public routes
	api.Get("/banners", h.List)
	
	// Protected routes
	api.Post("/banners/request", jwtMiddleware, h.Request) // CONCEPTION Demandes d'annonces

	// Admin routes
	admin := api.Group("/admin/banners", jwtMiddleware, adminMiddleware)
	admin.Get("/", h.AdminList)
	admin.Post("/", h.Create)
	admin.Put("/:id/toggle", h.Toggle)
	admin.Delete("/:id", h.Delete)
}

func setupContentRoutes(api fiber.Router, h *handlers.ContentHandler, jwtSecret string, logger *zap.Logger) {
	// Public routes
	api.Get("/faq", h.FAQ)
	api.Get("/tutorials", h.Tutorials)
	api.Get("/about", h.About)
	api.Get("/privacy-policy", h.Privacy)

	// Admin routes
	jwtMiddleware := middleware.JWT(jwtSecret, logger)
	adminMiddleware := middleware.AdminOnly(logger)

	admin := api.Group("/admin", jwtMiddleware, adminMiddleware)

	// FAQ CRUD
	admin.Post("/faq", h.CreateFAQ)
	admin.Put("/faq/:id", h.UpdateFAQ)
	admin.Delete("/faq/:id", h.DeleteFAQ)

	// Tutorials CRUD
	admin.Post("/tutorials", h.CreateTutorial)
	admin.Put("/tutorials/:id", h.UpdateTutorial)
	admin.Delete("/tutorials/:id", h.DeleteTutorial)
}

func setupNotificationRoutes(api fiber.Router, h *handlers.NotificationHandler, jwtSecret string, logger *zap.Logger) {
	jwtMiddleware := middleware.JWT(jwtSecret, logger)
	
	notif := api.Group("/notifications", jwtMiddleware)
	
	notif.Post("/push-tokens", h.SaveToken)
	notif.Get("/", h.List)
	notif.Put("/read-all", h.MarkAllAsRead)
	notif.Put("/:id/read", h.MarkAsRead)
}


