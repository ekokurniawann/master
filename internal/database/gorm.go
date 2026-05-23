package database

import (
	"context"
	"fmt"
	"time"

	"backend-skripsi/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewPostgres() (*gorm.DB, error) {
	cfg := config.Get()

	if cfg.Database.DSN == "" {
		return nil, fmt.Errorf("database.gorm.NewPostgres: database dsn is required")
	}

	db, err := gorm.Open(
		postgres.New(postgres.Config{
			DSN: cfg.Database.DSN,
		}),
		&gorm.Config{
			SkipDefaultTransaction: true,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("database.gorm.NewPostgres: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("database.gorm.NewPostgres: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.Database.ConnMaxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("database.gorm.NewPostgres: %w", err)
	}

	return db, nil
}
