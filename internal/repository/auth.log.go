package repository

import (
	"context"

	"backend-skripsi/internal/entity"

	"gorm.io/gorm"
)

type AuthLogRepository struct {
	db *gorm.DB
}

func NewAuthLogRepository(db *gorm.DB) *AuthLogRepository {
	return &AuthLogRepository{
		db: db,
	}
}

func (r *AuthLogRepository) Create(ctx context.Context, log *entity.AuthLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}
