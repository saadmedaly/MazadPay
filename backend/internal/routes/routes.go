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
	reqRepo := repository.NewRequestRepository(db)
	auditRepo := repository.NewAuditRepository(db)

	// Hub
	hub := ws.NewHub(logger)
	adminHub := ws.NewAdminHub(logger)

	// Services
	notifSvc := services.NewNotificationService(notifRepo, userRepo, cfg.Firebase.ServiceAccountPath, logger, adminHub)
	smsSvc := services.NewSMSService(cfg.Twilio.AccountSID, cfg.Twilio.AuthToken, cfg.Twilio.PhoneNumber, logger)
	authSvc := services.NewAuthService(userRepo, cfg.JWT.Secret, cfg.JWT.ExpiryHours, cfg.App.Env, cfg.App.DevOTPCode, smsSvc, 4)
	auctionSvc := services.NewAuctionService(auctionRepo, reportRepo, notifSvc, userRepo)
	bidSvc := services.NewBidService(db, auctionRepo, bidRepo, walletRepo, hub)
	userSvc := services.NewUserService(userRepo, favoriteRepo, auctionRepo, kycRepo)
	adminSvc := services.NewAdminService(db, userRepo, auctionRepo, bidRepo, txRepo, reportRepo, kycRepo, contentRepo, invRepo, reqRepo)
	walletSvc := services.NewWalletService(walletRepo, txRepo)
	contentSvc := services.NewContentService(contentRepo, notifSvc)
	reqSvc := services.NewRequestService(reqRepo, auctionRepo, contentRepo, auditRepo, notifSvc)
	
	// New services from migration 000031
	paymentMethodSvc := services.NewPaymentMethodService(db)
	auctionBoostSvc := services.NewAuctionBoostService(db)
	deliveryDriverSvc := services.NewDeliveryDriverService(db)
	bidAutoBidSvc := services.NewBidAutoBidService(db, bidSvc, walletSvc)

	api := app.Group("/v1/api")

	// Handlers
	wsHandler := handlers.NewWSHandler(hub, logger)
	adminWSHandler := handlers.NewAdminWSHandler(adminHub, cfg.JWT.Secret, logger)
	bidHandler := handlers.NewBidHandler(bidSvc, logger)
	userHandler := handlers.NewUserHandler(userSvc, logger)
	adminHandler := handlers.NewAdminHandler(adminSvc, logger)
	bannerHandler := handlers.NewBannerHandler(contentSvc, logger)
	walletHandler := handlers.NewWalletHandler(walletSvc, logger)
	reqHandler := handlers.NewRequestHandler(reqSvc, logger)
	contentHandler := handlers.NewContentHandler(contentSvc, logger)
	notifHandler := handlers.NewNotificationHandler(notifSvc, logger)
	// New handlers with services
	paymentMethodHandler := handlers.NewPaymentMethodHandler(paymentMethodSvc, logger)
	auctionBoostHandler := handlers.NewAuctionBoostHandler(auctionBoostSvc, logger)
	deliveryDriverHandler := handlers.NewDeliveryDriverHandler(deliveryDriverSvc, logger)
	bidAutoBidHandler := handlers.NewBidAutoBidHandler(bidAutoBidSvc, logger)

	// WebSocket registration
	app.Use("/ws", wsHandler.UpgradeMiddleware())
	app.Get("/ws/auction/:id", websocket.New(wsHandler.HandleAuction))

	// Admin WebSocket
	app.Use("/ws/admin", adminWSHandler.UpgradeMiddleware())
	app.Get("/ws/admin", websocket.New(adminWSHandler.HandleAdmin))

	// Public routes for countries (no auth required)
	api.Get("/countries", adminHandler.ListCountries)

	// Routes registration
	setupAuthRoutes(api, authSvc, adminHandler, rdb, cfg, logger)
	setupAuctionRoutes(api, auctionSvc, bidHandler, userHandler, cfg.JWT.Secret, logger)
	setupUserRoutes(api, userHandler, walletHandler, cfg.JWT.Secret, logger)
	setupAdminRoutes(api, adminHandler, userHandler, cfg.JWT.Secret, logger)
	setupBannerRoutes(api, bannerHandler, cfg.JWT.Secret, logger)
	setupContentRoutes(api, contentHandler, cfg.JWT.Secret, logger)
	setupNotificationRoutes(api, notifHandler, cfg.JWT.Secret, logger)
	setupRequestRoutes(api, reqHandler, cfg.JWT.Secret, logger, auditRepo, rdb, cfg)
	// New routes
	setupPaymentMethodRoutes(api, paymentMethodHandler, cfg.JWT.Secret, logger)
	setupAuctionBoostRoutes(api, auctionBoostHandler, cfg.JWT.Secret, logger)
	setupDeliveryDriverRoutes(api, deliveryDriverHandler, cfg.JWT.Secret, logger)
	setupBidAutoBidRoutes(api, bidAutoBidHandler, cfg.JWT.Secret, logger)

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

func setupAuctionRoutes(api fiber.Router, auctionSvc services.AuctionService, bidHandler *handlers.BidHandler, userHandler *handlers.UserHandler, jwtSecret string, logger *zap.Logger) {
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
	auctions.Put("/:id", h.Update) // Modifier son enchère
	auctions.Delete("/:id", h.Delete) // Supprimer son enchère
	auctions.Post("/:id/report", h.Report) // CONCEPTION Signalements
	auctions.Post("/:id/images", h.AddImages)
	auctions.Post("/:id/buy-now", h.BuyNow)
	auctions.Post("/:id/cancel", h.Cancel)
	auctions.Post("/:id/relist", h.Relist)
	auctions.Post("/:id/extend", h.Extend)
	auctions.Get("/:id/bid-status", h.GetBidStatus) // Statut de ma bid
	auctions.Get("/:id/winner", h.GetWinner) // Détails du gagnant
	auctions.Post("/:id/contact", h.ContactSeller) // Contacter vendeur

	// Consolidation endpoints /my/*
	my := api.Group("/my", jwtMiddleware)
	my.Get("/auctions", userHandler.MyAuctions) // Consolidation endpoint
	my.Get("/auctions/active", userHandler.MyAuctionsActive)
	my.Get("/auctions/ended", userHandler.MyAuctionsEnded)
	my.Get("/auctions/pending", userHandler.MyAuctionsPending)
	my.Get("/bids", userHandler.MyBids) // Consolidation endpoint
	my.Get("/bids/active", userHandler.MyBidsActive)
	my.Get("/bids/won", userHandler.MyWinnings)
	my.Get("/bids/lost", userHandler.MyBidsLost)
	my.Get("/watchlist", userHandler.ListFavorites) // Enchères suivies

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

	// User Settings (new)
	users.Get("/me/settings", userHandler.GetUserSettings)
	users.Put("/me/settings", userHandler.UpdateUserSettings)

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
	users.Get("/wallet/payment-methods", walletHandler.GetPaymentMethods)

	// KYC
	users.Get("/kyc", userHandler.GetKYCStatus)
	users.Post("/kyc", userHandler.SubmitKYC)
}

func setupAdminRoutes(api fiber.Router, adminHandler *handlers.AdminHandler, userHandler *handlers.UserHandler, jwtSecret string, logger *zap.Logger) {
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

	// KYC management
	admin.Get("/kyc", adminHandler.ListKYCs)
	admin.Put("/kyc/:user_id", adminHandler.ReviewKYC)
	admin.Put("/users/:user_id/kyc-status", userHandler.UpdateKYCStatus)

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

	// Payment Methods management (from migration 000031)
	admin.Get("/payment-methods", adminHandler.ListPaymentMethods)
	admin.Post("/payment-methods", adminHandler.CreatePaymentMethod)
	admin.Put("/payment-methods/:id", adminHandler.UpdatePaymentMethod)
	admin.Delete("/payment-methods/:id", adminHandler.DeletePaymentMethod)

	// Auction Car Details management (from migration 000031)
	admin.Get("/auctions/:id/car-details", adminHandler.GetAuctionCarDetails)
	admin.Put("/auctions/:id/car-details", adminHandler.UpdateAuctionCarDetails)

	// Auction Boost management (from migration 000031)
	admin.Get("/auction-boosts", adminHandler.ListAuctionBoosts)
	admin.Post("/auction-boosts", adminHandler.CreateAuctionBoost)
	admin.Delete("/auction-boosts/:id", adminHandler.DeleteAuctionBoost)

	// Delivery Drivers management (from migration 000031)
	admin.Get("/delivery-drivers", adminHandler.ListDeliveryDrivers)
	admin.Post("/delivery-drivers", adminHandler.CreateDeliveryDriver)
	admin.Put("/delivery-drivers/:id", adminHandler.UpdateDeliveryDriver)
	admin.Delete("/delivery-drivers/:id", adminHandler.DeleteDeliveryDriver)

	// User Settings management (from migration 000031)
	admin.Get("/users/:id/settings", adminHandler.GetUserSettings)
	admin.Put("/users/:id/settings", adminHandler.UpdateUserSettings)

	// Bid Auto Bid management (from migration 000031)
	admin.Get("/auto-bids", adminHandler.ListAutoBids)
	admin.Put("/auto-bids/:id", adminHandler.UpdateAutoBid)
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
	notifications.Post("/push-tokens", notifHandler.SaveToken)
	notifications.Get("/", notifHandler.List)
	notifications.Put("/:id/read", notifHandler.MarkAsRead)
	notifications.Put("/read-all", notifHandler.MarkAllAsRead)
	
	// Admin notification management
	admin := api.Group("/admin/notifications", jwtMiddleware)
	admin.Get("", notifHandler.AdminList)
	admin.Post("/send", notifHandler.SendNotification)
	admin.Post("/broadcast", notifHandler.SendNotification)
	admin.Put("/:id/read", notifHandler.MarkAsRead)
	admin.Put("/read-all", notifHandler.MarkAllAsRead)
	admin.Delete("/:id", notifHandler.AdminDelete)
	admin.Get("/templates", notifHandler.GetTemplates)
}

func setupRequestRoutes(api fiber.Router, reqHandler *handlers.RequestHandler, jwtSecret string, logger *zap.Logger, auditRepo repository.AuditRepository, rdb *redis.Client, cfg *config.Config) {
	jwtMiddleware := middleware.JWT(jwtSecret, logger)
	adminMiddleware := middleware.AdminOnly(logger)

	// Rate limiting for request submissions: 5 requests per hour per user
	rateLimitSubmit := middleware.RateLimit(rdb, 3600, 5, logger)

	// Public routes for users to submit requests (protected by JWT + rate limiting)
	user := api.Group("/requests", jwtMiddleware)
	user.Post("/auctions", rateLimitSubmit, reqHandler.CreateAuctionRequest)
	user.Post("/banners", rateLimitSubmit, reqHandler.CreateBannerRequest)
	user.Get("/auctions/my", reqHandler.GetUserAuctionRequests)
	user.Get("/banners/my", reqHandler.GetUserBannerRequests)

	// Admin request management
	admin := api.Group("/admin/requests", jwtMiddleware, adminMiddleware)

	// Auction requests
	admin.Get("/auctions", reqHandler.GetAuctionRequests)
	admin.Get("/auctions/:id", reqHandler.GetAuctionRequestByID)
	admin.Put("/auctions/:id/review", reqHandler.ReviewAuctionRequest)
	admin.Delete("/auctions/:id", reqHandler.DeleteAuctionRequest)
	admin.Post("/auctions/bulk/review", reqHandler.BulkReviewAuctionRequests)
	admin.Post("/auctions/bulk/delete", reqHandler.BulkDeleteAuctionRequests)

	// Banner requests
	admin.Get("/banners", reqHandler.GetBannerRequests)
	admin.Get("/banners/:id", reqHandler.GetBannerRequestByID)
	admin.Put("/banners/:id/review", reqHandler.ReviewBannerRequest)
	admin.Delete("/banners/:id", reqHandler.DeleteBannerRequest)
	admin.Post("/banners/bulk/review", reqHandler.BulkReviewBannerRequests)
	admin.Post("/banners/bulk/delete", reqHandler.BulkDeleteBannerRequests)

	// Audit logs (admin only)
	auditHandler := handlers.NewAuditHandler(services.NewAuditService(auditRepo), logger)
	admin.Get("/audit/logs", auditHandler.GetAuditLogs)
	admin.Get("/audit/logs/:entity_type/:entity_id", auditHandler.GetAuditLogsByEntity)
}

// New route setup functions for additional features

func setupPaymentMethodRoutes(api fiber.Router, h *handlers.PaymentMethodHandler, jwtSecret string, logger *zap.Logger) {
	jwtMiddleware := middleware.JWT(jwtSecret, logger)
	adminMiddleware := middleware.AdminOnly(logger)

	// Public routes
	api.Get("/payment-methods", h.ListPaymentMethods)

	// Admin routes
	admin := api.Group("/admin/payment-methods", jwtMiddleware, adminMiddleware)
	admin.Post("/", h.CreatePaymentMethod)
	admin.Put("/:id", h.UpdatePaymentMethod)
	admin.Delete("/:id", h.DeletePaymentMethod)
	admin.Put("/:id/toggle", h.TogglePaymentMethodStatus)
}

func setupAuctionBoostRoutes(api fiber.Router, h *handlers.AuctionBoostHandler, jwtSecret string, logger *zap.Logger) {
	jwtMiddleware := middleware.JWT(jwtSecret, logger)
	adminMiddleware := middleware.AdminOnly(logger)

	// User routes
	api.Post("/auctions/:id/boost", jwtMiddleware, h.CreateBoost)
	api.Get("/auctions/:id/boosts", jwtMiddleware, h.GetAuctionBoosts)
	api.Delete("/auctions/:id/boosts/:boost_id", jwtMiddleware, h.CancelBoost)

	// Admin routes
	admin := api.Group("/admin/boosts", jwtMiddleware, adminMiddleware)
	admin.Get("/active", h.GetActiveBoosts)
	admin.Put("/:id/status", h.UpdateBoostStatus)
}

func setupDeliveryDriverRoutes(api fiber.Router, h *handlers.DeliveryDriverHandler, jwtSecret string, logger *zap.Logger) {
	jwtMiddleware := middleware.JWT(jwtSecret, logger)
	adminMiddleware := middleware.AdminOnly(logger)

	// Admin routes
	admin := api.Group("/admin/drivers", jwtMiddleware, adminMiddleware)
	admin.Post("/register", h.RegisterDriver)
	admin.Get("/", h.ListDrivers)
	admin.Get("/:id", h.GetDriver)
	admin.Put("/:id", h.UpdateDriver)
	admin.Get("/available", h.GetAvailableDrivers)

	// Driver routes
	driver := api.Group("/drivers", jwtMiddleware)
	driver.Put("/location", h.UpdateDriverLocation)
	driver.Put("/availability", h.ToggleAvailability)
}

func setupBidAutoBidRoutes(api fiber.Router, h *handlers.BidAutoBidHandler, jwtSecret string, logger *zap.Logger) {
	jwtMiddleware := middleware.JWT(jwtSecret, logger)
	adminMiddleware := middleware.AdminOnly(logger)

	// User routes
	api.Post("/auctions/:id/auto-bid", jwtMiddleware, h.CreateAutoBid)
	api.Get("/users/auto-bids", jwtMiddleware, h.GetMyAutoBids)
	api.Delete("/auctions/:id/auto-bid", jwtMiddleware, h.CancelAutoBid)
	api.Put("/auctions/:id/auto-bid", jwtMiddleware, h.UpdateAutoBid)

	// Admin routes
	admin := api.Group("/admin/auctions", jwtMiddleware, adminMiddleware)
	admin.Get("/:id/auto-bids", h.GetAuctionAutoBids)
}
