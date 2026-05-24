package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type VerificationCacheRepository interface {
	SetVerificationToken(ctx context.Context, email string, token string, ttl time.Duration) error
	GetVerificationToken(ctx context.Context, email string) (string, error)
	DeleteVerificationToken(ctx context.Context, email string) error
}

type BlacklistCacheRepository interface {
	BlacklistToken(ctx context.Context, tokenString string, ttl time.Duration) error
	IsTokenBlacklisted(ctx context.Context, tokenString string) (bool, error)
}

type PasswordResetCacheRepository interface {
	SetPasswordResetToken(ctx context.Context, email string, token string, ttl time.Duration) error
	GetPasswordResetToken(ctx context.Context, email string) (string, error)
	DeletePasswordResetToken(ctx context.Context, email string) error
}

type RateLimitCacheRepository interface {
	IsRateLimited(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
}

type CacheRepository interface {
	VerificationCacheRepository
	BlacklistCacheRepository
	PasswordResetCacheRepository
	RateLimitCacheRepository
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

func (r *redisCacheRepo) BlacklistToken(ctx context.Context, tokenString string, ttl time.Duration) error {
	key := "auth:blacklist:" + tokenString
	return r.rdb.Set(ctx, key, "1", ttl).Err()
}

func (r *redisCacheRepo) IsTokenBlacklisted(ctx context.Context, tokenString string) (bool, error) {
	key := "auth:blacklist:" + tokenString
	n, err := r.rdb.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *redisCacheRepo) SetPasswordResetToken(ctx context.Context, email string, token string, ttl time.Duration) error {
	key := "auth:forgot-password:" + email
	return r.rdb.Set(ctx, key, token, ttl).Err()
}

func (r *redisCacheRepo) GetPasswordResetToken(ctx context.Context, email string) (string, error) {
	key := "auth:forgot-password:" + email
	val, err := r.rdb.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", nil
	}
	return val, err
}

func (r *redisCacheRepo) DeletePasswordResetToken(ctx context.Context, email string) error {
	key := "auth:forgot-password:" + email
	return r.rdb.Del(ctx, key).Err()
}

func (r *redisCacheRepo) IsRateLimited(ctx context.Context, ip string, limit int, window time.Duration) (bool, error) {
	key := "auth:ratelimit:" + ip

	pipe := r.rdb.Pipeline()

	incr := pipe.Incr(ctx, key)
	pipe.ExpireNX(ctx, key, window)

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return false, fmt.Errorf("repository.cache.IsRateLimited: %w", err)
	}

	currentCount, err := incr.Result()
	if err != nil {
		return false, fmt.Errorf("repository.cache.IsRateLimited: %w", err)
	}

	if int(currentCount) > limit {
		return true, nil
	}

	return false, nil
}
