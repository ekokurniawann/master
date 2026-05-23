package logger

import (
	"backend-skripsi/internal/config"
	"log/slog"
	"os"

	"github.com/go-chi/httplog/v3"
)

func New() *slog.Logger {
	cfg := config.Get()

	isLocalhost := cfg.App.Env == "development"
	logFormat := httplog.SchemaECS.Concise(isLocalhost)

	level := slog.LevelInfo
	if cfg.App.Debug {
		level = slog.LevelDebug
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:       level,
		ReplaceAttr: logFormat.ReplaceAttr,
	})).With(
		slog.String("app", cfg.App.Name),
		slog.String("env", cfg.App.Env),
	)
}
