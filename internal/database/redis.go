package database

import (
	"context"
	"fmt"
	"time"

	"backend-skripsi/internal/config"

	"github.com/redis/go-redis/v9"
)

func NewRedis() (*redis.Client, error) {
	cfg := config.Get()

	if cfg.Redis.URL == "" {
		return nil, fmt.Errorf("database.redis.NewRedis: redis url is required")
	}

	opt, err := redis.ParseURL(cfg.Redis.URL)
	if err != nil {
		return nil, fmt.Errorf("database.redis.NewRedis: %w", err)
	}

	opt.PoolSize = cfg.Redis.PoolSize
	opt.MinIdleConns = cfg.Redis.MinIdleConns

	rdb := redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("database.redis.NewRedis: %w", err)
	}

	return rdb, nil
}
