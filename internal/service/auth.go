package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"backend-skripsi/internal/entity"
	"backend-skripsi/internal/handler/dto"
	"backend-skripsi/internal/handler/middleware"
	"backend-skripsi/internal/mailer"
	"backend-skripsi/internal/repository"
	"backend-skripsi/internal/security"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo    *repository.UserRepository
	cacheRepo   repository.CacheRepository
	authLogRepo *repository.AuthLogRepository
	mailClient  mailer.Mailer
	jwtProvider *security.JWTProvider
}

func NewAuthService(
	userRepo *repository.UserRepository,
	cacheRepo repository.CacheRepository,
	authLogRepo *repository.AuthLogRepository,
	mailClient mailer.Mailer,
	jwtProvider *security.JWTProvider,
) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		cacheRepo:   cacheRepo,
		authLogRepo: authLogRepo,
		mailClient:  mailClient,
		jwtProvider: jwtProvider,
	}
}

func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest) (string, error) {
	ipAddress := middleware.GetClientIPFromContext(ctx)
	userAgent := middleware.GetUserAgentFromContext(ctx)

	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, entity.ErrUserNotFound) {
			errMsg := "Email tidak ditemukan"
			_ = s.authLogRepo.Create(ctx, &entity.AuthLog{
				Event:        "LOGIN_FAILED",
				Status:       "FAILED",
				IPAddress:    ipAddress,
				UserAgent:    userAgent,
				ErrorMessage: &errMsg,
			})
			return "", fmt.Errorf("service.auth.Login: %w", entity.ErrInvalidCredentials)
		}
		return "", fmt.Errorf("service.auth.Login: %w", err)
	}

	if user.IsVerified == entity.UserUnverified {
		errMsg := "Akun belum diverifikasi"
		_ = s.authLogRepo.Create(ctx, &entity.AuthLog{
			UserID:       &user.ID,
			Event:        "LOGIN_FAILED",
			Status:       "FAILED",
			IPAddress:    ipAddress,
			UserAgent:    userAgent,
			ErrorMessage: &errMsg,
		})
		return "", fmt.Errorf("service.auth.Login: %w", entity.ErrUserNotVerified)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		errMsg := "Password salah"
		_ = s.authLogRepo.Create(ctx, &entity.AuthLog{
			UserID:       &user.ID,
			Event:        "LOGIN_FAILED",
			Status:       "FAILED",
			IPAddress:    ipAddress,
			UserAgent:    userAgent,
			ErrorMessage: &errMsg,
		})
		return "", fmt.Errorf("service.auth.Login: %w", entity.ErrInvalidCredentials)
	}

	token, err := s.jwtProvider.GenerateToken(user.ID, user.Email, user.Role.Name)
	if err != nil {
		return "", fmt.Errorf("service.auth.Login: %w", err)
	}
	_ = s.authLogRepo.Create(ctx, &entity.AuthLog{
		UserID:    &user.ID,
		Event:     "LOGIN_SUCCESS",
		Status:    "SUCCESS",
		IPAddress: ipAddress,
		UserAgent: userAgent,
	})

	return token, nil
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

func (s *AuthService) Logout(ctx context.Context, tokenString string) error {
	ipAddress := middleware.GetClientIPFromContext(ctx)
	userAgent := middleware.GetUserAgentFromContext(ctx)

	claims, err := s.jwtProvider.ValidateToken(tokenString)
	if err != nil {
		if errors.Is(err, entity.ErrTokenExpired) {
			slog.Info("logout requested for already expired token, skipping redis write")
			return nil
		}
		return fmt.Errorf("service.auth.Logout: token tidak valid: %w", err)
	}

	expirationTime := claims.ExpiresAt.Time
	remainingTTL := time.Until(expirationTime)

	if remainingTTL <= 0 {
		return nil
	}

	err = s.cacheRepo.BlacklistToken(ctx, tokenString, remainingTTL)
	if err != nil {
		return fmt.Errorf("service.auth.Logout: %w", err)
	}
	_ = s.authLogRepo.Create(ctx, &entity.AuthLog{
		UserID:    &claims.UserID,
		Event:     "LOGOUT",
		Status:    "SUCCESS",
		IPAddress: ipAddress,
		UserAgent: userAgent,
	})

	slog.Info("user logout successfully, token blacklisted dynamic",
		slog.String("user_id", claims.UserID.String()),
		slog.Float64("remaining_ttl_minutes", remainingTTL.Minutes()),
	)
	return nil
}

func (s *AuthService) GetProfile(ctx context.Context, userID uuid.UUID) (dto.UserMeResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return dto.UserMeResponse{}, fmt.Errorf("service.auth.GetProfile: %w", err)
	}

	response := dto.UserMeResponse{
		Email:       user.Email,
		FullName:    user.FullName,
		PhoneNumber: user.PhoneNumber,
		Address:     user.Address,
		Province:    user.Province,
		City:        user.City,
		PostalCode:  user.PostalCode,
	}

	return response, nil
}

func (s *AuthService) ForgotPassword(ctx context.Context, req dto.ForgotPasswordRequest) error {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, entity.ErrUserNotFound) {
			return fmt.Errorf("service.auth.ForgotPassword: %w", entity.ErrUserNotFound)
		}
		return fmt.Errorf("service.auth.ForgotPassword: %w", err)
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return fmt.Errorf("service.auth.ForgotPassword: failed to generate token: %w", err)
	}
	token := hex.EncodeToString(tokenBytes)

	ttl := 15 * time.Minute
	if err := s.cacheRepo.SetPasswordResetToken(ctx, user.Email, token, ttl); err != nil {
		return fmt.Errorf("service.auth.ForgotPassword: failed to save token to cache: %w", err)
	}

	emailCtx := context.Background()
	go func() {
		err := s.mailClient.SendPasswordReset(emailCtx, user.Email, user.FullName, token)
		if err != nil {
			slog.Error("async forgot password email sender error",
				slog.String("email", user.Email),
				slog.String("err", err.Error()),
			)
		} else {
			slog.Info("async forgot password email sent successfully", slog.String("email", user.Email))
		}
	}()

	return nil
}

func (s *AuthService) ResetPassword(ctx context.Context, req dto.ResetPasswordRequest) error {
	cachedToken, err := s.cacheRepo.GetPasswordResetToken(ctx, req.Email)
	if err != nil {
		return fmt.Errorf("service.auth.ResetPassword: %w", err)
	}

	if cachedToken == "" {
		return fmt.Errorf("service.auth.ResetPassword: %w", entity.ErrVerificationTokenExpired)
	}

	if cachedToken != req.Token {
		return fmt.Errorf("service.auth.ResetPassword: %w", entity.ErrInvalidVerificationToken)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("service.auth.ResetPassword: failed to hash new password: %w", err)
	}

	if err := s.userRepo.UpdatePassword(ctx, req.Email, string(hashedPassword)); err != nil {
		return fmt.Errorf("service.auth.ResetPassword: %w", err)
	}

	if err := s.cacheRepo.DeletePasswordResetToken(ctx, req.Email); err != nil {
		slog.Error("failed to delete password reset token from cache after success",
			slog.String("email", req.Email),
			slog.String("err", err.Error()),
		)
	}

	slog.Info("user password updated successfully via forgot password mechanism", slog.String("email", req.Email))
	return nil
}
