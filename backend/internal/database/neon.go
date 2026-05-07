package database

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// NeonDB wraps the pgxpool and logger for Neon database serverless interaction
type NeonDB struct {
	Pool   *pgxpool.Pool
	logger *zap.Logger
}

// ConnectWithRetry connects to Neon DB with specific free tier limits and retry logic.
// Retry x5 with exponential backoff (500ms, 1s, 2s, 4s, 8s).
// Logs "Neon cold start detected" if latency > 1s.
func ConnectWithRetry(ctx context.Context, connString string, logger *zap.Logger) (*NeonDB, error) {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	// 1. pgxpool.Config with Neon limits for Free Tier
	config.MaxConns = 10
	config.MinConns = 0
	config.MaxConnIdleTime = 4 * time.Minute
	config.MaxConnLifetime = 30 * time.Minute
	config.HealthCheckPeriod = 5 * time.Minute

	var pool *pgxpool.Pool

	// 2. Retry x5 with exponential backoff
	backoff := []time.Duration{
		500 * time.Millisecond,
		1 * time.Second,
		2 * time.Second,
		4 * time.Second,
		8 * time.Second,
	}

	// context.WithTimeout(30s) for the first connect
	connectCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	for i, d := range backoff {
		start := time.Now()

		pool, err = pgxpool.NewWithConfig(connectCtx, config)
		if err == nil {
			err = pool.Ping(connectCtx)
		}

		latency := time.Since(start)

		if err == nil {
			// Log "Neon cold start detected" if latency > 1s
			if latency > 1*time.Second {
				logger.Info("Neon cold start detected", zap.Duration("latency", latency))
			} else {
				logger.Info("PostgreSQL connected successfully", zap.Duration("latency", latency))
			}
			return &NeonDB{Pool: pool, logger: logger}, nil
		}

		logger.Warn("Failed to connect to Neon DB, retrying...",
			zap.Error(err),
			zap.Int("attempt", i+1),
			zap.Duration("delay", d),
		)

		if pool != nil {
			pool.Close()
		}

		select {
		case <-connectCtx.Done():
			return nil, fmt.Errorf("connection context timed out or cancelled: %w", connectCtx.Err())
		case <-time.After(d):
		}
	}

	return nil, fmt.Errorf("failed to connect to Neon DB after 5 attempts: %w", err)
}

// 3. Fiber Middleware that adds the pool to ctx.Locals()
func NeonPoolMiddleware(db *NeonDB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Locals("db_pool", db)
		return c.Next()
	}
}

// GetNeonDB is a helper to extract the db from fiber context (optional utility)
func GetNeonDB(c *fiber.Ctx) *NeonDB {
	if db, ok := c.Locals("db_pool").(*NeonDB); ok {
		return db
	}
	return nil
}

// 4. Wrapper ExecWithTimeout
func (db *NeonDB) ExecWithTimeout(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error) {
	// Timeout automatique 8s sur chaque requête
	execCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	start := time.Now()
	tag, err := db.Pool.Exec(execCtx, query, args...)
	latency := time.Since(start)

	// Log les requêtes > 300ms avec la requête SQL (tronquée à 200 chars)
	if latency > 300*time.Millisecond {
		logQuery := query
		if len(logQuery) > 200 {
			logQuery = logQuery[:200] + "..."
		}
		db.logger.Warn("Slow query detected",
			zap.Duration("latency", latency),
			zap.String("query", logQuery),
		)
	}

	return tag, err
}

// Close closes the underlying pool connections
func (db *NeonDB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}
