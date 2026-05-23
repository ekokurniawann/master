package repository

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type CacheRepository interface {
	SetVerificationToken(ctx context.Context, email string, token string, ttl time.Duration) error
	GetVerificationToken(ctx context.Context, email string) (string, error)
	DeleteVerificationToken(ctx context.Context, email string) error
}

type redisCacheRepo struct {
	rdb *redis.Client
}

func NewCacheRepository(rdb *redis.Client) CacheRepository {
	return &redisCacheRepo{
		rdb: rdb,
	}
}

func (r *redisCacheRepo) SetVerificationToken(ctx context.Context, email string, token string, ttl time.Duration) error {
	key := "auth:verify:" + email
	return r.rdb.Set(ctx, key, token, ttl).Err()
}

func (r *redisCacheRepo) GetVerificationToken(ctx context.Context, email string) (string, error) {
	key := "auth:verify:" + email
	val, err := r.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

func (r *redisCacheRepo) DeleteVerificationToken(ctx context.Context, email string) error {
	key := "auth:verify:" + email
	return r.rdb.Del(ctx, key).Err()
}
