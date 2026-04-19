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
	Termii   TermiiConfig
	Business BusinessConfig
	Firebase FirebaseConfig
}

type AppConfig struct {
	Env        string
	Port       string
	Name       string
	DevOTPCode string // Code OTP de développement (ignoré en production)
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
}

type TermiiConfig struct {
	APIKey   string
	BaseURL  string
	SenderID string
}

type BusinessConfig struct {
	BidMinIncrement      float64
	InsuranceDefault     float64
	PaymentDeadlineHours int
}

type FirebaseConfig struct {
	ServiceAccountPath string
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
			RateLimitMaxAttempts:   getEnvInt("REDIS_RATE_LIMIT_MAX_ATTEMPTS", 3),
		},
		JWT: JWTConfig{
			Secret:            getEnv("JWT_SECRET", ""),
			ExpiryHours:       getEnvInt("JWT_EXPIRY_HOURS", 72),
			RefreshExpiryDays: getEnvInt("JWT_REFRESH_EXPIRY_DAYS", 30),
		},
		R2: R2Config{
			Endpoint:    getEnv("R2_ENDPOINT", "xxxxxxxx.r2.cloudflarestorage.com"),
			AccessKey:   getEnv("R2_ACCESS_KEY", "your_r2_access_key"),
			SecretKey:   getEnv("R2_SECRET_KEY", "your_r2_secret_key"),
			BucketMedia: getEnv("R2_BUCKET_MEDIA", "mazadpay-media"),
			PublicURL:   getEnv("R2_PUBLIC_URL", "https://pub-xxxxxx.r2.dev"),
		},
		Termii: TermiiConfig{
			APIKey:   getEnv("TERMII_API_KEY", ""),
			BaseURL:  getEnv("TERMII_BASE_URL", "https://api.ng.termii.com"),
			SenderID: getEnv("TERMII_SENDER_ID", "MazadPay"),
		},
		Business: BusinessConfig{
			BidMinIncrement:      getEnvFloat("BID_MIN_INCREMENT", 100),
			InsuranceDefault:     getEnvFloat("INSURANCE_DEFAULT_AMOUNT", 500),
			PaymentDeadlineHours: getEnvInt("PAYMENT_DEADLINE_HOURS", 48),
		},
		Firebase: FirebaseConfig{
			ServiceAccountPath: getEnv("FIREBASE_SERVICE_ACCOUNT_PATH", ""),
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
