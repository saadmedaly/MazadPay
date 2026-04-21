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
	notifSvc := services.NewNotificationService(notifRepo, userRepo, cfg.Firebase.ServiceAccountPath, logger)
	smsSvc := services.NewSMSService(cfg.Twilio.AccountSID, cfg.Twilio.AuthToken, cfg.Twilio.PhoneNumber, logger)
	authSvc := services.NewAuthService(userRepo, cfg.JWT.Secret, cfg.JWT.ExpiryHours, cfg.App.Env, cfg.App.DevOTPCode, smsSvc, 4)
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

	// Public routes for countries (no auth required)
	api.Get("/countries", adminHandler.ListCountries)

	// Routes registration
	setupAuthRoutes(api, authSvc, adminHandler, rdb, cfg, logger)
	setupAuctionRoutes(api, auctionSvc, bidHandler, cfg.JWT.Secret, logger)
	setupUserRoutes(api, userHandler, walletHandler, cfg.JWT.Secret, logger)
	setupAdminRoutes(api, adminHandler, cfg.JWT.Secret, logger)
	setupBannerRoutes(api, bannerHandler, cfg.JWT.Secret, logger)
	setupContentRoutes(api, contentHandler, cfg.JWT.Secret, logger)
	setupNotificationRoutes(api, notifHandler, cfg.JWT.Secret, logger)

	return auctionSvc, notifSvc
}

// setupAuthRoutes enregistre les routes d'authentification avec rate limiting
func setupAuthRoutes(api fiber.Router, authSvc services.AuthService, adminHandler *handlers.AdminHandler, rdb *redis.Client, cfg *config.Config, logger *zap.Logger) {
	jwtMiddleware := middleware.JWT(cfg.JWT.Secret, logger)
	rateLimitMiddleware := middleware.RateLimitByPhone(rdb, cfg.Redis.RateLimitWindowSeconds, cfg.Redis.RateLimitMaxAttempts, logger)
	h := handlers.NewAuthHandler(authSvc, logger)

	auth := api.Group("/auth")

	// Public routes avec rate limiting sur phone
	auth.Post("/register", rateLimitMiddleware, h.Register)
	auth.Post("/login", rateLimitMiddleware, h.Login)
	auth.Post("/otp/send", rateLimitMiddleware, h.SendOTP)
	auth.Post("/otp/verify", rateLimitMiddleware, h.VerifyOTP)
	auth.Post("/reset-password", rateLimitMiddleware, h.ResetPassword)
	auth.Post("/register-admin", rateLimitMiddleware, adminHandler.RegisterWithInvitation)

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
	api.Get("/countries", h.GetCountries)
	api.Get("/locations/:countryId", h.GetLocationsByCountry)
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

	// Super Admin only - Delete user
	admin.Delete("/users/:id", middleware.SuperAdminOnly(logger), adminHandler.DeleteUser)

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

	// Notification management routes
	admin.Get("/notifications", adminHandler.AdminList)
	admin.Delete("/notifications/:id", adminHandler.AdminDelete)

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

	// Country management
	admin.Get("/countries", adminHandler.ListCountries)
	admin.Post("/countries", adminHandler.CreateCountry)
	admin.Put("/countries/:id", adminHandler.UpdateCountry)
	admin.Delete("/countries/:id", adminHandler.DeleteCountry)

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

	// Admin create banner (direct)
	api.Post("/banners", jwtMiddleware, adminMiddleware, h.Create)

	// Admin routes
	admin := api.Group("/admin/banners", jwtMiddleware, adminMiddleware)
	admin.Get("/", h.AdminList)
	admin.Get("/all", h.AdminListAll) // Debug - explicit all banners
	admin.Post("/", h.Create)
	admin.Put("/:id/toggle", h.Toggle)
	admin.Put("/:id", h.Update)
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
	admin.Get("/faq", h.AdminListFAQ)
	admin.Post("/faq", h.CreateFAQ)
	admin.Put("/faq/:id", h.UpdateFAQ)
	admin.Delete("/faq/:id", h.DeleteFAQ)

	// Tutorials CRUD
	admin.Get("/tutorials", h.AdminListTutorials)
	admin.Post("/tutorials", h.CreateTutorial)
	admin.Put("/tutorials/:id", h.UpdateTutorial)
	admin.Delete("/tutorials/:id", h.DeleteTutorial)
}

func setupNotificationRoutes(api fiber.Router, notifHandler *handlers.NotificationHandler, jwtSecret string, logger *zap.Logger) {
	jwtMiddleware := middleware.JWT(jwtSecret, logger)
	
	notifications := api.Group("/notifications", jwtMiddleware)
	notifications.Post("/token", notifHandler.SaveToken)
	notifications.Get("/", notifHandler.List)
	notifications.Put("/:id/read", notifHandler.MarkAsRead)
	notifications.Put("/read-all", notifHandler.MarkAllAsRead)
	
	// Admin notification management
	admin := api.Group("/admin/notifications", jwtMiddleware)
	admin.Post("/send", notifHandler.SendNotification)
	admin.Post("/broadcast", notifHandler.SendNotification)
	admin.Get("/templates", notifHandler.GetTemplates)
}
