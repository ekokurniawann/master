package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"time"

	"backend-skripsi/internal/entity"
	"backend-skripsi/internal/handler/dto"
	"backend-skripsi/internal/mailer"
	"backend-skripsi/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo   *repository.UserRepository
	cacheRepo  repository.CacheRepository
	mailClient mailer.Mailer
}

func NewAuthService(
	userRepo *repository.UserRepository,
	cacheRepo repository.CacheRepository,
	mailClient mailer.Mailer,
) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		cacheRepo:  cacheRepo,
		mailClient: mailClient,
	}
}

func (s *AuthService) Register(ctx context.Context, req dto.RegisterRequest) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("service.auth.Register: failed to hash password: %w", err)
	}

	newUser := &entity.User{
		RoleID:     entity.RoleCustomerID,
		Email:      req.Email,
		Password:   string(hashedPassword),
		FullName:   req.FullName,
		IsVerified: entity.UserUnverified,
	}

	if err := s.userRepo.Create(ctx, newUser); err != nil {
		return fmt.Errorf("service.auth.Register: %w", err)
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return fmt.Errorf("service.auth.Register: failed to generate secure token: %w", err)
	}
	token := hex.EncodeToString(tokenBytes)

	ttl := 15 * time.Minute
	if err := s.cacheRepo.SetVerificationToken(ctx, newUser.Email, token, ttl); err != nil {
		return fmt.Errorf("service.auth.Register: failed to save token to cache: %w", err)
	}

	emailCtx := context.Background()
	go func() {
		err := s.mailClient.SendVerification(emailCtx, newUser.Email, newUser.FullName, token)
		if err != nil {
			slog.Error("async email sender error",
				slog.String("email", newUser.Email),
				slog.String("err", err.Error()),
			)
		} else {
			slog.Info("async email verification sent successfully", slog.String("email", newUser.Email))
		}
	}()

	return nil
}

func (s *AuthService) VerifyEmail(ctx context.Context, email string, token string) error {
	cachedToken, err := s.cacheRepo.GetVerificationToken(ctx, email)
	if err != nil {
		return fmt.Errorf("service.auth.VerifyEmail: failed to fetch token from cache: %w", err)
	}

	if cachedToken == "" {
		return fmt.Errorf("service.auth.VerifyEmail: %w", entity.ErrVerificationTokenExpired)
	}

	if cachedToken != token {
		return fmt.Errorf("service.auth.VerifyEmail: %w", entity.ErrInvalidVerificationToken)
	}

	if err := s.userRepo.UpdateVerificationStatus(ctx, email, entity.UserVerified); err != nil {
		return fmt.Errorf("service.auth.VerifyEmail: %w", err)
	}

	if err := s.cacheRepo.DeleteVerificationToken(ctx, email); err != nil {
		slog.Error("failed to delete verification token from cache after success",
			slog.String("email", email),
			slog.String("err", err.Error()),
		)
	}

	return nil
}
