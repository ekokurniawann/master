package server

import (
	"log/slog"
	"net/http"

	_ "backend-skripsi/docs"
	"backend-skripsi/internal/config"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httplog/v3"
	"github.com/go-chi/httprate"
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
		})
	})

	return r
}

func (srv *Server) middlewares(r *chi.Mux) {
	cfg := config.Get()

	r.Use(middleware.RealIP)
	r.Use(middleware.CleanPath)

	r.Use(httplog.RequestLogger(srv.logger, &httplog.Options{
		Level:         slog.LevelInfo,
		Schema:        httplog.SchemaECS,
		RecoverPanics: true,
		Skip: func(req *http.Request, respStatus int) bool {
			return respStatus == 404
		},
	}))

	r.Use(middleware.Heartbeat(cfg.Health.Path))

	r.Use(httprate.LimitByIP(
		cfg.HTTP.RateLimit.Limit,
		cfg.HTTP.RateLimit.Window,
	))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.HTTP.CORS.AllowedOrigins,
		AllowedMethods:   cfg.HTTP.CORS.AllowedMethods,
		AllowedHeaders:   cfg.HTTP.CORS.AllowedHeaders,
		AllowCredentials: cfg.HTTP.CORS.AllowCredentials,
		MaxAge:           cfg.HTTP.CORS.MaxAge,
	}))

	r.Use(middleware.Timeout(cfg.HTTP.RequestTimeout))

	r.Use(middleware.Compress(
		cfg.HTTP.CompressionLevel,
		"application/json",
	))
}
