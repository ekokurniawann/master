package server

import (
	"log/slog"
	"net/http"

	_ "backend-skripsi/docs"
	"backend-skripsi/internal/config"
	"backend-skripsi/internal/handler/middleware"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httplog/v3"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func (srv *Server) routes() http.Handler {
	r := chi.NewRouter()

	srv.middlewares(r)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message":"Welcome to FORTISFIT API"}`))
	})

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("doc.json"),
		httpSwagger.Plugins([]string{
			"SwaggerUIBundle.plugins.SearchWithInOperationsPlugin",
		}),
	))

	r.Route("/api/v1", func(v1 chi.Router) {
		v1.Group(func(auth chi.Router) {
			auth.Post("/auth/register", srv.handlers.auth.Register)
			auth.Get("/auth/verify", srv.handlers.auth.VerifyEmail)
			auth.Post("/auth/login", srv.handlers.auth.Login)
			auth.Post("/auth/forgot-password", srv.handlers.auth.ForgotPassword)
			auth.Get("/auth/reset-password", srv.handlers.auth.ResetPasswordView)
			auth.Post("/auth/reset-password", srv.handlers.auth.ResetPassword)
		})

		v1.Group(func(protected chi.Router) {
			protected.Use(middleware.JWTMiddleware(srv.jwtProvider, srv.cacheRepo))
			protected.Get("/auth/me", srv.handlers.auth.GetProfileMe)
			protected.Post("/auth/logout", srv.handlers.auth.Logout)
		})
	})

	return r
}

func (srv *Server) middlewares(r *chi.Mux) {
	cfg := config.Get()

	r.Use(chimiddleware.RealIP)

	r.Use(middleware.ExtractClientInfo)

	r.Use(chimiddleware.CleanPath)

	r.Use(httplog.RequestLogger(srv.logger, &httplog.Options{
		Level:         slog.LevelInfo,
		Schema:        httplog.SchemaECS,
		RecoverPanics: true,
		Skip: func(req *http.Request, respStatus int) bool {
			return respStatus == 404
		},
	}))

	r.Use(chimiddleware.Heartbeat(cfg.Health.Path))

	r.Use(middleware.RedisRateLimiter(srv.cacheRepo))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.HTTP.CORS.AllowedOrigins,
		AllowedMethods:   cfg.HTTP.CORS.AllowedMethods,
		AllowedHeaders:   cfg.HTTP.CORS.AllowedHeaders,
		AllowCredentials: cfg.HTTP.CORS.AllowCredentials,
		MaxAge:           cfg.HTTP.CORS.MaxAge,
	}))

	r.Use(chimiddleware.Timeout(cfg.HTTP.RequestTimeout))

	r.Use(chimiddleware.Compress(
		cfg.HTTP.CompressionLevel,
		"application/json",
	))
}
