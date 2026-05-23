package main

import (
	"fmt"
	"log/slog"
	"os"

	"backend-skripsi/internal/app"
	"backend-skripsi/internal/server"
)

// @title           Backend Skripsi API
// @version         1.0
// @description     Dokumentasi API untuk Aplikasi FORTISFIT
// @termsOfService  http://swagger.io/terms/

// @contact.name   Eko Kurniawan
// @contact.email  ekokurniawaann@gmail.com

// @BasePath        /api/v1
func main() {
	res, cleanup, err := app.Bootstrap()
	if err != nil {
		slog.Error("application bootstrap failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer cleanup()

	srv := server.New(res)

	if err := srv.Start(); err != nil {
		wrappedServerErr := fmt.Errorf("main.main: server failed to start: %w", err)
		res.Logger.Error("application forced to shutdown", slog.String("error", wrappedServerErr.Error()))
		os.Exit(1)
	}
}
