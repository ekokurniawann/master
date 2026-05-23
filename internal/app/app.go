package app

import (
	"fmt"
	"log"
	"log/slog"

	"backend-skripsi/internal/config"
	"backend-skripsi/internal/database"
	"backend-skripsi/internal/logger"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Resources struct {
	Logger *slog.Logger
	DB     *gorm.DB
	Redis  *redis.Client
}

func Bootstrap() (*Resources, func(), error) {
	if err := godotenv.Load(); err != nil {
		log.Println("info: .env file not found, using system environment variables")
	}

	if err := config.Load("config.yml"); err != nil {
		return nil, nil, fmt.Errorf("app.app.Bootstrap: failed to load config: %w", err)
	}

	logInstance := logger.New()
	slog.SetDefault(logInstance)

	db, err := database.NewPostgres()
	if err != nil {
		return nil, nil, fmt.Errorf("app.app.Bootstrap: database connections failed: %w", err)
	}

	rdb, err := database.NewRedis()
	if err != nil {
		if sqlDB, sqlErr := db.DB(); sqlErr == nil {
			_ = sqlDB.Close()
		}
		return nil, nil, fmt.Errorf("app.app.Bootstrap: redis connection failed: %w", err)
	}

	cleanup := func() {
		if errClose := rdb.Close(); errClose != nil {
			logInstance.Error("failed to close redis connection", slog.String("error", fmt.Errorf("app.app.cleanup: %w", errClose).Error()))
		} else {
			logInstance.Info("redis connection closed successfully")
		}

		if sqlDB, errDB := db.DB(); errDB == nil {
			if errClose := sqlDB.Close(); errClose != nil {
				logInstance.Error("failed to close database connection", slog.String("error", fmt.Errorf("app.app.cleanup: %w", errClose).Error()))
			} else {
				logInstance.Info("database connection closed successfully")
			}
		}
	}

	res := &Resources{
		Logger: logInstance,
		DB:     db,
		Redis:  rdb,
	}

	return res, cleanup, nil
}
