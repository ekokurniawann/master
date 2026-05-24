package repository

import (
	"context"
	"errors"
	"fmt"

	"backend-skripsi/internal/entity"

	"github.com/google/uuid"
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

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User

	// Eager Loading
	err := r.db.WithContext(ctx).
		Preload("Role").
		Where("email = ?", email).
		First(&user).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("repository.user.FindByEmail: %w", entity.ErrUserNotFound)
		}

		return nil, fmt.Errorf("repository.user.FindByEmail: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	var user entity.User

	err := r.db.WithContext(ctx).
		Select("id", "email", "full_name", "phone_number", "address", "province", "city", "postal_code").
		Where("id = ?", id).
		First(&user).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("repository.user.FindByID: %w", entity.ErrUserNotFound)
		}
		return nil, fmt.Errorf("repository.user.FindByID: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) UpdatePassword(ctx context.Context, email string, hashedPassword string) error {
	err := r.db.WithContext(ctx).
		Model(&entity.User{}).
		Where("email = ?", email).
		Update("password", hashedPassword).
		Error

	if err != nil {
		return fmt.Errorf("repository.user.UpdatePassword: %w", err)
	}

	return nil
}
