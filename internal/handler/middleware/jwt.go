package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"backend-skripsi/internal/entity"
	"backend-skripsi/internal/repository"
	"backend-skripsi/internal/response"
	"backend-skripsi/internal/security"
)

func JWTMiddleware(jwtProvider *security.JWTProvider, cacheRepo repository.BlacklistCacheRepository) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.WriteError(w, http.StatusUnauthorized, "token autentikasi tidak ditemukan", nil)
				return
			}

			fields := strings.Fields(authHeader)
			if len(fields) < 2 || strings.ToLower(fields[0]) != "bearer" {
				response.WriteError(w, http.StatusUnauthorized, "format token tidak valid, gunakan Bearer", nil)
				return
			}

			tokenString := fields[1]

			isBlacklisted, err := cacheRepo.IsTokenBlacklisted(r.Context(), tokenString)
			if err != nil {
				response.WriteInternalServerError(w)
				return
			}
			if isBlacklisted {
				response.WriteError(w, http.StatusUnauthorized, entity.ErrTokenInvalid.Error(), nil)
				return
			}

			claims, err := jwtProvider.ValidateToken(tokenString)
			if err != nil {
				if errors.Is(err, entity.ErrTokenExpired) {
					response.WriteError(w, http.StatusUnauthorized, entity.ErrTokenExpired.Error(), nil)
					return
				}

				response.WriteError(w, http.StatusUnauthorized, entity.ErrTokenInvalid.Error(), nil)
				return
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, CtxUserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, CtxTokenKey, tokenString)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
