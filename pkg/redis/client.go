package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/diabolusgx/guess-the-number-go/internal/config"
	"github.com/diabolusgx/guess-the-number-go/pkg/logger"
	"github.com/redis/go-redis/v9"
)

func NewClient(cfg *config.Configuration, logger *logger.Logger) (*redis.Client, error) {
	logger.Info("Initializing Redis client...")

	redisAddr := fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port)

	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: cfg.Redis.PoolSize,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		logger.Errorf("Failed to connect to Redis: %v", err)
		return nil, fmt.Errorf("redis connection test failed: %w", err)
	}

	logger.Info("Redis client initialized successfully")
	return client, nil
}
