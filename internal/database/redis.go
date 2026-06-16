package database

import (
	"context"
	"fmt"
	"time"

	"github.com/collegeassess/backend/configs"
	"github.com/redis/go-redis/v9"
)

// NewRedis returns a connected Redis client, or (nil, nil) when caching is disabled.
func NewRedis(cfg *configs.Config) (*redis.Client, error) {
	if !cfg.Redis.Enabled {
		return nil, nil
	}
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("connect redis: %w", err)
	}
	return client, nil
}
