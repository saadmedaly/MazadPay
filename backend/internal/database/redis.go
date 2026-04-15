package database

import (
	"context"
	"fmt"

	"github.com/mazadpay/backend/internal/config"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func NewRedis(cfg *config.Config, logger *zap.Logger) (*redis.Client, error) {
	opts, err := redis.ParseURL(cfg.Redis.URL)
	if err != nil {
		return nil, fmt.Errorf("redis parse URL: %w", err)
	}

	client := redis.NewClient(opts)

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	logger.Info("Redis connected", zap.String("url", cfg.Redis.URL))
	return client, nil
}
