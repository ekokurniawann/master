package repository

import (
	"context"
	"errors"
	"fmt"

	"backend-skripsi/internal/entity"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) Create(ctx context.Context, user *entity.User) error {
	err := r.db.WithContext(ctx).Create(user).Error
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf("repository.user.Create: %w (db_detail: %s)", entity.ErrEmailAlreadyExists, pgErr.Message)
		}

		return fmt.Errorf("repository.user.Create: %w", err)
	}
	return nil
}

func (r *UserRepository) UpdateVerificationStatus(ctx context.Context, email string, isVerified bool) error {
	err := r.db.WithContext(ctx).
		Model(&entity.User{}).
		Where("email = ?", email).
		Update("is_verified", isVerified).
		Error

	if err != nil {
		return fmt.Errorf("repository.user.UpdateVerificationStatus: %w", err)
	}

	return nil
}
