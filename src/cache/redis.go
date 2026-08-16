// Package cache provides an optional Redis client. Vidhya Service does not
// require Redis to serve any API today - it's wired up so that when the
// team adds caching/session features, connecting is a one-line config
// change (REDIS_ENABLED=true) rather than a new integration.
package cache

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"

	"vidhya-service/src/config"
	"vidhya-service/src/logger"
)

// Connect returns a Redis client if caching is enabled and reachable.
// It returns (nil, nil) when disabled so callers can treat cache as
// best-effort and simply skip it - the rest of the service keeps working
// against Postgres alone.
func Connect(ctx context.Context, cfg config.RedisConfig) (*redis.Client, error) {
	if !cfg.Enabled {
		logger.Log.Info("", "Redis is disabled (REDIS_ENABLED=false) - skipping connection")
		return nil, nil
	}

	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Address,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("cache.Connect: ping redis at %s: %w", cfg.Address, err)
	}

	logger.Log.Info("", "Connected to Redis at %s", cfg.Address)
	return client, nil
}
