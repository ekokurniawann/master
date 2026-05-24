package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"

	"backend-skripsi/internal/app"
	"backend-skripsi/internal/config"
	"backend-skripsi/internal/repository"
	"backend-skripsi/internal/security"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Server struct {
	logger      *slog.Logger
	db          *gorm.DB
	rdb         *redis.Client
	handlers    *handlers
	jwtProvider *security.JWTProvider
	cacheRepo   repository.CacheRepository
}

func New(res *app.Resources) *Server {
	srv := &Server{
		logger: res.Logger,
		db:     res.DB,
		rdb:    res.Redis,
	}

	srv.handlers = srv.initDependencies()

	return srv
}

func (srv *Server) Start() error {
	cfg := config.Get()
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)

	httpSrv := &http.Server{
		Addr:           addr,
		Handler:        srv.routes(),
		ReadTimeout:    cfg.Server.ReadTimeout,
		WriteTimeout:   cfg.Server.WriteTimeout,
		IdleTimeout:    cfg.Server.IdleTimeout,
		MaxHeaderBytes: cfg.Server.MaxHeaderBytes,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		srv.logger.Info("server listening", slog.String("addr", addr))
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("server.server.Start: %w", err)

	case <-ctx.Done():
		srv.logger.Info("shutting down server gracefully")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
		defer cancel()

		if err := httpSrv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("server.server.Start: failed to shutdown: %w", err)
		}

		srv.logger.Info("server stopped cleanly")
		return nil
	}
}
