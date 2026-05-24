package middleware

import (
	"net/http"

	"backend-skripsi/internal/config"
	"backend-skripsi/internal/repository"
	"backend-skripsi/internal/response"
)

func RedisRateLimiter(cacheRepo repository.RateLimitCacheRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			cfg := config.Get()

			ipAddress := GetClientIPFromContext(ctx)

			limit := cfg.HTTP.RateLimit.Limit
			window := cfg.HTTP.RateLimit.Window

			isLimited, err := cacheRepo.IsRateLimited(ctx, ipAddress, limit, window)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			if isLimited {
				response.WriteTooManyRequests(w, "Too Many Requests")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
