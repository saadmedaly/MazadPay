package config

import (
	"errors"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	App      AppConfig
	DB       DBConfig
	Redis    RedisConfig
	JWT      JWTConfig
	R2       R2Config
	Wablas   WablasConfig
	Business BusinessConfig
	Firebase FirebaseConfig
	Bankily  BankilyConfig
}

type AppConfig struct {
	Env                    string
	Port                   string
	Name                   string
	DevOTPCode             string // Code OTP de développement (ignoré en production)
	DefaultSuperAdminPhone string // Téléphone du super admin par défaut
	DefaultSuperAdminPin   string // PIN du super admin par défaut
	// CORSAllowedOrigins est la valeur BRUTE (non parsée) de CORS_ALLOWED_ORIGINS —
	// une liste d'origines séparées par des virgules, ou vide. Le parsing/validation
	// (rejet des espaces, doublons, wildcard, chemins/query, schémas non http(s)) vit
	// dans cmd/server (ResolveAllowedOrigins), pas ici, pour rester testable sans
	// dépendre du chargement de la config globale (voir cors.go).
	CORSAllowedOrigins string
}

type DBConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type RedisConfig struct {
	URL                    string
	OTPTTLMinutes          int
	RateLimitWindowSeconds int
	RateLimitMaxAttempts   int
}

type JWTConfig struct {
	Secret            string
	ExpiryHours       int
	RefreshExpiryDays int
}

type R2Config struct {
	Endpoint    string
	AccessKey   string
	SecretKey   string
	BucketMedia string
	PublicURL   string
	// BucketReceipts est un bucket R2 privé (sans accès public) dédié aux reçus de
	// paiement, séparé du bucket média public (audit de sécurité — voir Private
	// Receipts Bucket Separation). Vide par défaut : aucun fallback silencieux vers le
	// bucket public en production (voir media_service.go).
	BucketReceipts string
}

type WablasConfig struct {
	Token     string
	SecretKey string
	ServerURL string
}

type BusinessConfig struct {
	BidMinIncrement      float64
	InsuranceDefault     float64
	PaymentDeadlineHours int
}

// BankilyConfig (Bankily B-PAY automatic wallet top-up). Credentials are
// server-side only and come exclusively from the environment -- never from
// code, never sent to the mobile/admin clients. Every value Bankily has not
// yet confirmed (operationId format, decimal amounts, timeouts, maximum
// pending window) is configuration rather than a hard-coded guess.
type BankilyConfig struct {
	Enabled      bool
	BaseURL      string
	Username     string
	Password     string
	ClientID     string
	MerchantCode string // public merchant code shown to customers in the B-PAY instructions
	// HTTPTimeout bounds a single Bankily HTTP call. A timeout on /payment is
	// never treated as a failure -- the attempt moves to PENDING_VERIFICATION
	// and is resolved through /checkTransaction.
	HTTPTimeout         time.Duration
	OperationIDPrefix   string
	OperationIDDigits   int
	AllowDecimalAmounts bool
	MinAmount           float64
	MaxAmount           float64
	// ReconcileInterval: how often the worker scans unresolved attempts.
	ReconcileInterval time.Duration
	// MaxPendingAge: after this, an unresolved attempt is flagged needs_review
	// for an operator. It is never auto-failed -- money may have moved.
	MaxPendingAge time.Duration
}

// Ready reports whether the integration is switched on and fully configured.
func (b BankilyConfig) Ready() bool {
	return b.Enabled && b.BaseURL != "" && b.Username != "" && b.Password != "" && b.ClientID != ""
}

type FirebaseConfig struct {
	ServiceAccountPath string
	ServiceAccountJSON string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment")
	}

	return &Config{
		App: AppConfig{
			Env:        getEnv("APP_ENV", "development"),
			Port:       getEnv("APP_PORT", "8082"),
			Name:       getEnv("APP_NAME", "MazadPay"),
			DevOTPCode: getEnv("DEV_OTP_CODE", ""),
			// Pas de fallback : si ces variables ne sont pas définies explicitement,
			// aucun super admin par défaut ne doit être créé (voir audit de sécurité V02 -
			// un PIN par défaut connu dans le code source est une porte dérobée).
			DefaultSuperAdminPhone: getEnv("DEFAULT_SUPER_ADMIN_PHONE", ""),
			DefaultSuperAdminPin:   getEnv("DEFAULT_SUPER_ADMIN_PIN", ""),
			CORSAllowedOrigins:     getEnv("CORS_ALLOWED_ORIGINS", ""),
		},
		DB: DBConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			User:            getEnv("DB_USER", "mazadpay"),
			Password:        getEnv("DB_PASSWORD", ""),
			Name:            getEnv("DB_NAME", "mazadpay"),
			SSLMode:         getEnv("DB_SSL_MODE", "disable"),
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: time.Duration(getEnvInt("DB_CONN_MAX_LIFETIME_MINUTES", 30)) * time.Minute,
		},
		Redis: RedisConfig{
			URL:                    getEnv("REDIS_URL", "redis://localhost:6379/0"),
			OTPTTLMinutes:          getEnvInt("REDIS_OTP_TTL_MINUTES", 5),
			RateLimitWindowSeconds: getEnvInt("REDIS_RATE_LIMIT_WINDOW_SECONDS", 900),
			RateLimitMaxAttempts:   getEnvInt("REDIS_RATE_LIMIT_MAX_ATTEMPTS", 6),
		},
		JWT: JWTConfig{
			Secret:            getEnv("JWT_SECRET", ""),
			ExpiryHours:       getEnvInt("JWT_EXPIRY_HOURS", 72),
			RefreshExpiryDays: getEnvInt("JWT_REFRESH_EXPIRY_DAYS", 30),
		},
		R2: R2Config{
			Endpoint:       getEnv("R2_ENDPOINT", "xxxxxxxx.r2.cloudflarestorage.com"),
			AccessKey:      getEnv("R2_ACCESS_KEY", "your_r2_access_key"),
			SecretKey:      getEnv("R2_SECRET_KEY", "your_r2_secret_key"),
			BucketMedia:    getEnv("R2_BUCKET_MEDIA", "mazad-mwdia"),
			PublicURL:      getEnv("R2_PUBLIC_URL", "https://pub-xxxxxx.r2.dev"),
			BucketReceipts: getEnv("R2_BUCKET_RECEIPTS", ""),
		},
		Wablas: WablasConfig{
			Token:     getEnv("WABLAS_TOKEN", ""),
			SecretKey: getEnv("WABLAS_SECRET_KEY", ""),
			ServerURL: getEnv("WABLAS_SERVER_URL", "https://wablas.com"),
		},
		Business: BusinessConfig{
			BidMinIncrement:      getEnvFloat("BID_MIN_INCREMENT", 100),
			InsuranceDefault:     getEnvFloat("INSURANCE_DEFAULT_AMOUNT", 500),
			PaymentDeadlineHours: getEnvInt("PAYMENT_DEADLINE_HOURS", 48),
		},
		Firebase: FirebaseConfig{
			ServiceAccountPath: getEnv("FIREBASE_SERVICE_ACCOUNT_PATH", ""),
			ServiceAccountJSON: getEnv("FIREBASE_SERVICE_ACCOUNT_JSON", ""),
		},
		Bankily: BankilyConfig{
			Enabled:             getEnvBool("BANKILY_ENABLED", false),
			BaseURL:             getEnv("BANKILY_BASE_URL", ""),
			Username:            getEnv("BANKILY_USERNAME", ""),
			Password:            getEnv("BANKILY_PASSWORD", ""),
			ClientID:            getEnv("BANKILY_CLIENT_ID", ""),
			MerchantCode:        getEnv("BANKILY_MERCHANT_CODE", ""),
			HTTPTimeout:         time.Duration(getEnvInt("BANKILY_HTTP_TIMEOUT_SECONDS", 30)) * time.Second,
			OperationIDPrefix:   getEnv("BANKILY_OPERATION_ID_PREFIX", "MZP"),
			OperationIDDigits:   getEnvInt("BANKILY_OPERATION_ID_DIGITS", 13),
			AllowDecimalAmounts: getEnvBool("BANKILY_ALLOW_DECIMAL_AMOUNTS", false),
			MinAmount:           getEnvFloat("BANKILY_MIN_AMOUNT", 100),
			MaxAmount:           getEnvFloat("BANKILY_MAX_AMOUNT", 100000),
			ReconcileInterval:   time.Duration(getEnvInt("BANKILY_RECONCILE_INTERVAL_SECONDS", 30)) * time.Second,
			MaxPendingAge:       time.Duration(getEnvInt("BANKILY_MAX_PENDING_HOURS", 72)) * time.Hour,
		},
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvFloat(key string, fallback float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}

// Validate vérifie que les valeurs critique de la config sont configurées
func (c *Config) Validate() error {
	if c.JWT.Secret == "" {
		return errors.New("JWT_SECRET must be set in environment")
	}
	if c.App.Port == "" {
		return errors.New("APP_PORT must be set")
	}
	return nil
}
