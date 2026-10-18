package database

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/mazadpay/backend/internal/config"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func NewRedis(cfg *config.Config, logger *zap.Logger) (*redis.Client, error) {
	opts, err := redis.ParseURL(cfg.Redis.URL)
	if err != nil {
		// Never wrap the parse error: url.Parse errors embed the raw URL,
		// password included.
		return nil, fmt.Errorf("redis parse URL (%s): invalid REDIS_URL", RedactRedisURL(cfg.Redis.URL))
	}

	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	logger.Info("Redis connected", zap.String("addr", RedactRedisURL(cfg.Redis.URL)))
	return client, nil
}

// RedactRedisURL returns scheme://host:port/db for logging -- never the
// username, password or query string. Unparseable input yields a fixed
// placeholder rather than any part of the raw value.
func RedactRedisURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return "<unparseable REDIS_URL>"
	}
	return (&url.URL{Scheme: u.Scheme, Host: u.Host, Path: u.Path}).String()
}
