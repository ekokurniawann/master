package middleware

import (
	"context"
	"net"
	"net/http"

	"github.com/google/uuid"
)

type ctxKey string

const (
	CtxUserIDKey    ctxKey = "userID"
	CtxTokenKey     ctxKey = "rawToken"
	CtxClientIPKey  ctxKey = "clientIP"
	CtxUserAgentKey ctxKey = "userAgent"
)

func ExtractClientInfo(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}

		userAgent := r.Header.Get("User-Agent")
		if userAgent == "" {
			userAgent = "Unknown Device"
		}

		ctx := context.WithValue(r.Context(), CtxClientIPKey, ip)
		ctx = context.WithValue(ctx, CtxUserAgentKey, userAgent)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(CtxUserIDKey).(uuid.UUID)
	return userID, ok
}

func GetTokenFromContext(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(CtxTokenKey).(string)
	return token, ok
}

func GetClientIPFromContext(ctx context.Context) string {
	if ip, ok := ctx.Value(CtxClientIPKey).(string); ok {
		return ip
	}
	return "0.0.0.0"
}

func GetUserAgentFromContext(ctx context.Context) string {
	if ua, ok := ctx.Value(CtxUserAgentKey).(string); ok {
		return ua
	}
	return "Unknown"
}
