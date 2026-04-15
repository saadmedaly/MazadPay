---
name: mazadpay-go-backend
description: Skill pour développer le backend Go (Fiber) de MazadPay — plateforme d'enchères en ligne pour le marché mauritanien. Utiliser ce skill pour générer des handlers, services, repositories, middlewares, WebSocket hubs, migrations SQL, ou tout fichier backend lié au projet MazadPay. Couvre l'architecture complète : Auth/OTP, Auctions, Bids, Wallets, WebSocket realtime, Notifications FCM, Upload Cloudflare R2, Cron jobs.
---

# MazadPay — Go Backend Skill

Ce skill guide le développement du backend Go de MazadPay en suivant les best practices production.
Chaque génération de code doit être fonctionnelle, testable et prête pour un déploiement réel.

---

## Stack Technique Imposée

| Composant | Technologie | Version cible |
|:---|:---|:---|
| Framework HTTP | **Fiber v2** | >= 2.52 |
| ORM / SQL | **sqlx** + requêtes SQL brutes | — |
| Migrations | **golang-migrate** | — |
| Auth | **golang-jwt/jwt v5** | — |
| WebSocket | **gorilla/websocket** via Fiber adapter | — |
| Cache | **go-redis/redis v9** | — |
| Validation | **go-playground/validator v10** | — |
| Logging | **uber-go/zap** | — |
| Config | **godotenv** + struct config | — |
| Cron | **robfig/cron v3** | — |
| UUID | **google/uuid** | — |
| Crypto | **golang.org/x/crypto** (bcrypt) | — |
| Storage | **minio-go/v7** | — |
| HTTP Client | **net/http** standard | — |

---

## Structure de Dossiers Imposée

```
backend/
├── cmd/
│   └── server/
│       └── main.go                 # Point d'entrée : init config, DB, Redis, routes, server
├── internal/
│   ├── config/
│   │   └── config.go               # Chargement .env → struct Config
│   ├── database/
│   │   ├── postgres.go             # Pool sqlx, ping, retry
│   │   └── redis.go                # Client Redis, ping
│   ├── models/                     # Structs Go miroir des tables SQL
│   │   ├── user.go
│   │   ├── auction.go
│   │   ├── bid.go
│   │   ├── wallet.go
│   │   ├── transaction.go
│   │   ├── notification.go
│   │   └── ...
│   ├── repository/                 # Couche SQL pure (pas de logique métier)
│   │   ├── user_repo.go
│   │   ├── auction_repo.go
│   │   ├── bid_repo.go
│   │   ├── wallet_repo.go
│   │   └── ...
│   ├── services/                   # Logique métier (orchestration repos + règles)
│   │   ├── auth_service.go
│   │   ├── auction_service.go
│   │   ├── bid_service.go
│   │   ├── wallet_service.go
│   │   └── ...
│   ├── handlers/                   # Fiber handlers (HTTP layer uniquement)
│   │   ├── auth_handler.go
│   │   ├── auction_handler.go
│   │   ├── bid_handler.go
│   │   ├── wallet_handler.go
│   │   └── ...
│   ├── middleware/
│   │   ├── auth.go                 # JWT middleware
│   │   ├── rate_limit.go           # Redis-based rate limiting
│   │   └── logger.go               # Request logging zap
│   ├── websocket/
│   │   ├── hub.go                  # Hub central (rooms par auction_id)
│   │   ├── client.go               # Connexion WebSocket individuelle
│   │   └── events.go               # Types d'events (bid_placed, timer_tick, auction_ended)
│   ├── cron/
│   │   └── jobs.go                 # Clôture auto enchères, nettoyage OTP
│   └── routes/
│       └── routes.go               # Enregistrement de toutes les routes Fiber
├── migrations/
│   ├── 000001_init.up.sql          # Schéma complet
│   └── 000001_init.down.sql        # Rollback
├── docker-compose.yml              # PostgreSQL + Redis
├── .env.example
├── go.mod
└── go.sum
```

---

## Règles de Code — TOUJOURS Respecter

### 1. Models (structs Go)

```go
// Toujours utiliser db tags (sqlx), json tags, et validate tags
type Auction struct {
    ID              uuid.UUID       `db:"id"               json:"id"`
    SellerID        uuid.UUID       `db:"seller_id"        json:"seller_id"`
    Title           string          `db:"title"            json:"title"           validate:"required,min=3,max=200"`
    CurrentPrice    decimal.Decimal `db:"current_price"    json:"current_price"`
    Status          string          `db:"status"           json:"status"`
    EndTime         time.Time       `db:"end_time"         json:"end_time"`
    ItemDetails     JSONB           `db:"item_details"     json:"item_details,omitempty"`
    Version         int             `db:"version"          json:"version"`
    CreatedAt       time.Time       `db:"created_at"       json:"created_at"`
}

// JSONB helper type
type JSONB map[string]interface{}
func (j JSONB) Value() (driver.Value, error) { ... }
func (j *JSONB) Scan(src interface{}) error { ... }
```

### 2. Repository Pattern

```go
// Chaque repo a une interface + implémentation
type AuctionRepository interface {
    FindByID(ctx context.Context, id uuid.UUID) (*models.Auction, error)
    FindActive(ctx context.Context, filters AuctionFilters) ([]models.Auction, int, error)
    Create(ctx context.Context, tx *sqlx.Tx, auction *models.Auction) error
    UpdatePrice(ctx context.Context, tx *sqlx.Tx, id uuid.UUID, newPrice decimal.Decimal, version int) (bool, error)
}

// UpdatePrice retourne bool (true = succès optimistic lock, false = conflit de version)
// SQL: UPDATE auctions SET current_price=$1, version=version+1 WHERE id=$2 AND version=$3
```

### 3. Verrouillage Optimiste — CRITIQUE pour les Bids

```go
// Dans bid_service.go — placer une mise
func (s *BidService) PlaceBid(ctx context.Context, auctionID, userID uuid.UUID, amount decimal.Decimal) error {
    // Toujours dans une transaction SQL
    tx, err := s.db.BeginTxx(ctx, nil)
    defer tx.Rollback()
    
    // SELECT FOR UPDATE sur le wallet (évite double-spend)
    wallet, err := s.walletRepo.FindForUpdate(ctx, tx, userID)
    
    // Vérifier solde >= insurance_amount
    // Vérifier amount > current_price
    // Vérifier auction.status == "active" && auction.end_time > now()
    
    // Update avec version check
    ok, err := s.auctionRepo.UpdatePrice(ctx, tx, auctionID, amount, auction.Version)
    if !ok {
        return ErrBidConflict // Le client doit retry
    }
    
    // Insérer le bid
    // Libérer l'ancien hold du précédent meilleur enchérisseur
    // Créer un nouveau hold pour le nouvel enchérisseur
    
    tx.Commit()
    
    // Broadcaster l'event WebSocket APRÈS le commit
    s.hub.Broadcast(auctionID, BidPlacedEvent{...})
}
```

### 4. Handlers Fiber

```go
// Pattern standard pour tous les handlers
func (h *AuctionHandler) GetAuction(c *fiber.Ctx) error {
    id, err := uuid.Parse(c.Params("id"))
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("invalid_id", "Invalid auction ID"))
    }
    
    auction, err := h.service.GetByID(c.Context(), id)
    if err != nil {
        if errors.Is(err, ErrNotFound) {
            return c.Status(fiber.StatusNotFound).JSON(ErrorResponse("not_found", "Auction not found"))
        }
        return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse("server_error", "Internal error"))
    }
    
    return c.JSON(SuccessResponse(auction))
}

// Helpers de réponse — TOUJOURS utiliser ces formats
func SuccessResponse(data interface{}) fiber.Map {
    return fiber.Map{"success": true, "data": data}
}
func ErrorResponse(code, message string) fiber.Map {
    return fiber.Map{"success": false, "error": fiber.Map{"code": code, "message": message}}
}
```

### 5. WebSocket Hub

```go
// hub.go — rooms par auction_id
type Hub struct {
    rooms map[uuid.UUID]map[*Client]bool
    mu    sync.RWMutex
    // channels: register, unregister, broadcast
}

// Events WebSocket (JSON)
type WSEvent struct {
    Type    string      `json:"type"`    // "bid_placed", "timer_tick", "auction_ended", "auction_won"
    Payload interface{} `json:"payload"`
}

// Timer tick — envoyé toutes les secondes par le Cron pour les enchères actives
// bid_placed — envoyé après commit d'un bid validé
// auction_ended — envoyé par le Cron à la clôture
// auction_won — envoyé uniquement au gagnant
```

### 6. Middleware JWT

```go
// middleware/auth.go
func JWTMiddleware(cfg *config.Config) fiber.Handler {
    return func(c *fiber.Ctx) error {
        token := extractBearerToken(c)
        claims, err := validateJWT(token, cfg.JWTSecret)
        if err != nil {
            return c.Status(401).JSON(ErrorResponse("unauthorized", "Invalid or expired token"))
        }
        c.Locals("user_id", claims.UserID)
        c.Locals("user_role", claims.Role)
        return c.Next()
    }
}

// Helper pour récupérer l'user dans les handlers
func GetUserID(c *fiber.Ctx) uuid.UUID {
    return c.Locals("user_id").(uuid.UUID)
}
```

### 7. Configuration (.env)

```env
# .env.example — tous les champs requis
APP_ENV=development
APP_PORT=8080

# PostgreSQL
DB_HOST=localhost
DB_PORT=5432
DB_USER=mazadpay
DB_PASSWORD=mazadpay_secret
DB_NAME=mazadpay
DB_SSL_MODE=disable
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5

# Redis
REDIS_URL=redis://localhost:6379/0
REDIS_OTP_TTL_MINUTES=5
REDIS_RATE_LIMIT_WINDOW_SECONDS=900
REDIS_RATE_LIMIT_MAX_ATTEMPTS=3

# JWT
JWT_SECRET=change_me_in_production_min_32_chars
JWT_EXPIRY_HOURS=72
JWT_REFRESH_EXPIRY_DAYS=30

# Cloudflare R2 / S3
R2_ENDPOINT=xxxxxxxx.r2.cloudflarestorage.com
R2_ACCESS_KEY=your_r2_access_key
R2_SECRET_KEY=your_r2_secret_key
R2_BUCKET_MEDIA=mazadpay-media
R2_PUBLIC_URL=https://pub-xxxxxx.r2.dev

# Termii SMS (à configurer plus tard)
TERMII_API_KEY=
TERMII_BASE_URL=https://api.ng.termii.com
TERMII_SENDER_ID=MazadPay

# Firebase FCM (à configurer plus tard)
FCM_SERVICE_ACCOUNT_JSON=

# App
DEFAULT_LANGUAGE=ar
INSURANCE_DEFAULT_AMOUNT=500
BID_MIN_INCREMENT=100
```

### 8. Docker Compose — Environnement de Développement

```yaml
# docker-compose.yml
version: "3.9"

services:
  postgres:
    image: postgres:15-alpine
    container_name: mazadpay_postgres
    environment:
      POSTGRES_DB: mazadpay
      POSTGRES_USER: mazadpay
      POSTGRES_PASSWORD: mazadpay_secret
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./migrations/000001_init.up.sql:/docker-entrypoint-initdb.d/init.sql
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U mazadpay"]
      interval: 5s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    container_name: mazadpay_redis
    ports:
      - "6379:6379"
    command: redis-server --save 60 1 --loglevel warning
    volumes:
      - redis_data:/data

volumes:
  postgres_data:
  redis_data:
```

### 9. Routes — Enregistrement Complet

```go
// routes/routes.go — déclarer TOUTES les routes ici
func SetupRoutes(app *fiber.App, handlers *Handlers, authMiddleware fiber.Handler) {
    api := app.Group("/v1/api")
    
    // Public
    auth := api.Group("/auth")
    auth.Post("/register", handlers.Auth.Register)
    auth.Post("/login", handlers.Auth.Login)
    auth.Post("/otp/send", handlers.Auth.SendOTP)
    auth.Post("/otp/verify", handlers.Auth.VerifyOTP)
    auth.Post("/reset-password", handlers.Auth.ResetPassword)
    
    // Protected
    protected := api.Use(authMiddleware)
    
    protected.Post("/auth/logout", handlers.Auth.Logout)
    protected.Put("/auth/change-password", handlers.Auth.ChangePassword)
    
    // Users
    protected.Get("/users/me", handlers.User.GetMe)
    protected.Put("/users/me", handlers.User.UpdateMe)
    protected.Post("/users/me/avatar", handlers.User.UploadAvatar)
    protected.Get("/users/me/bids", handlers.User.MyBids)
    protected.Get("/users/me/favorites", handlers.User.MyFavorites)
    protected.Get("/users/me/winnings", handlers.User.MyWinnings)
    
    // Auctions
    protected.Get("/auctions", handlers.Auction.List)
    protected.Get("/auctions/search", handlers.Auction.Search)
    protected.Get("/auctions/:id", handlers.Auction.GetByID)
    protected.Post("/auctions", handlers.Auction.Create)
    protected.Post("/auctions/:id/bids", handlers.Bid.Place)
    protected.Get("/auctions/:id/bids", handlers.Bid.History)
    protected.Get("/auctions/:id/summary", handlers.Auction.Summary)
    protected.Post("/auctions/:id/view", handlers.Auction.IncrementView)
    protected.Post("/auctions/:id/report", handlers.Auction.Report)
    protected.Get("/auctions/:id/seller-contact", handlers.Auction.SellerContact)
    
    // Favorites
    protected.Post("/favorites/:auction_id", handlers.Favorite.Add)
    protected.Delete("/favorites/:auction_id", handlers.Favorite.Remove)
    
    // Wallet & Finance
    protected.Get("/wallets/me", handlers.Wallet.GetMyWallet)
    protected.Post("/transactions/deposit", handlers.Transaction.Deposit)
    protected.Post("/transactions/withdraw", handlers.Transaction.Withdraw)
    protected.Get("/transactions", handlers.Transaction.History)
    protected.Post("/transactions/:id/receipt", handlers.Transaction.UploadReceipt)
    
    // Notifications
    protected.Get("/notifications", handlers.Notification.List)
    protected.Put("/notifications/read-all", handlers.Notification.MarkAllRead)
    
    // Content (Public)
    api.Get("/banners", handlers.Content.Banners)
    api.Get("/categories", handlers.Content.Categories)
    api.Get("/faq", handlers.Content.FAQ)
    api.Get("/tutorials", handlers.Content.Tutorials)
    api.Get("/about", handlers.Content.About)
    api.Get("/privacy-policy", handlers.Content.PrivacyPolicy)
    api.Get("/config/version", handlers.Content.Version)
    
    // WebSocket
    app.Get("/ws/auction/:id", handlers.WS.HandleAuction)
    
    // Admin (role check dans le handler)
    admin := protected.Group("/admin")
    admin.Get("/transactions", handlers.Admin.ListTransactions)
    admin.Put("/transactions/:id/validate", handlers.Admin.ValidateTransaction)
    admin.Get("/auctions/pending", handlers.Admin.PendingAuctions)
    admin.Put("/auctions/:id/approve", handlers.Admin.ApproveAuction)
    admin.Put("/auctions/:id/reject", handlers.Admin.RejectAuction)
    admin.Get("/reports", handlers.Admin.ListReports)
}
```

### 10. Cron Jobs

```go
// cron/jobs.go
func StartCronJobs(scheduler *cron.Cron, auctionService services.AuctionService) {
    // Clôture automatique des enchères expirées — toutes les 30 secondes
    scheduler.AddFunc("@every 30s", func() {
        auctionService.CloseExpiredAuctions(context.Background())
    })
    
    // Nettoyage OTP expirés — toutes les heures
    scheduler.AddFunc("@every 1h", func() {
        auctionService.CleanExpiredOTPs(context.Background())
    })
    
    // Marquer les paiements post-victoire comme overdue — toutes les heures
    scheduler.AddFunc("@every 1h", func() {
        auctionService.CheckOverduePayments(context.Background())
    })
    
    scheduler.Start()
}
```

---

## Workflow de Développement — Ordre Imposé

Toujours implémenter dans cet ordre pour chaque feature :

1. **Migration SQL** → `migrations/000001_init.up.sql`
2. **Model** → `internal/models/xxx.go`
3. **Repository interface + implémentation** → `internal/repository/xxx_repo.go`
4. **Service** → `internal/services/xxx_service.go`
5. **Handler** → `internal/handlers/xxx_handler.go`
6. **Route** → enregistrée dans `internal/routes/routes.go`

---

## Gestion des Erreurs — Types Standards

```go
// internal/errors/errors.go
var (
    ErrNotFound        = errors.New("resource_not_found")
    ErrUnauthorized    = errors.New("unauthorized")
    ErrForbidden       = errors.New("forbidden")
    ErrBidTooLow       = errors.New("bid_too_low")
    ErrBidConflict     = errors.New("bid_conflict")          // Optimistic lock fail → retry
    ErrInsufficientBalance = errors.New("insufficient_balance")
    ErrAuctionEnded    = errors.New("auction_ended")
    ErrOTPExpired      = errors.New("otp_expired")
    ErrOTPInvalid      = errors.New("otp_invalid")
    ErrOTPMaxAttempts  = errors.New("otp_max_attempts")
    ErrDuplicatePhone  = errors.New("phone_already_registered")
)
```

---

## Commandes de Démarrage Rapide

```bash
# Lancer l'infra (DB + Redis)
docker-compose up -d

# Vérifier que tout est up
docker-compose ps

# Vérifier que tout est up
docker-compose ps

# Initialiser le projet Go
cd backend
go mod init github.com/mazadpay/backend
go mod tidy

# Lancer les migrations
docker exec -i mazadpay_postgres psql -U mazadpay -d mazadpay < migrations/000001_init.up.sql

# Lancer le serveur en dev (avec hot-reload via air)
go install github.com/air-verse/air@latest
air

# Build production
go build -o ./bin/mazadpay ./cmd/server/main.go
```

---

## Règles Absolues — NE JAMAIS Violer

1. **Toute opération financière** (bid, deposit, withdraw, hold) DOIT être dans une transaction SQL (`BEGIN/COMMIT`) avec `SELECT FOR UPDATE` sur le wallet.
2. **Jamais de logique métier dans les handlers** — uniquement dans les services.
3. **Jamais de SQL dans les services** — uniquement dans les repositories.
4. **Broadcaster WebSocket APRÈS le COMMIT SQL** — jamais avant.
5. **Toujours masquer les numéros de téléphone** côté serveur avant de les renvoyer au client (format `####XXXX`).
6. **Valider l'input** avec `go-playground/validator` dans tous les handlers POST/PUT.
7. **Verrouillage optimiste** sur `auctions.version` pour chaque update de prix.
8. **Rate limiting OTP** : 3 tentatives max, blocage 15 min via Redis.
9. **Rôle admin** : vérifier `c.Locals("user_role") == "admin"` dans chaque handler admin.
10. **Logs structurés** avec `zap.Logger` — pas de `fmt.Println` en production.