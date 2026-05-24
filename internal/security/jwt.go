package security

import (
	"errors"
	"fmt"
	"time"

	"backend-skripsi/internal/config"
	"backend-skripsi/internal/entity"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTClaims struct {
	UserID   uuid.UUID `json:"user_id"`
	Email    string    `json:"email"`
	RoleName string    `json:"role_name"`
	jwt.RegisteredClaims
}

type JWTProvider struct {
	secretKey []byte
	issuer    string
	audience  string
}

func NewJWTProvider() *JWTProvider {
	cfg := config.Get().JWT

	return &JWTProvider{
		secretKey: []byte(cfg.Secret),
		issuer:    cfg.Issuer,
		audience:  cfg.Audience,
	}
}

func (j *JWTProvider) GenerateToken(userID uuid.UUID, email, roleName string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)

	claims := JWTClaims{
		UserID:   userID,
		Email:    email,
		RoleName: roleName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    j.issuer,
			Audience:  jwt.ClaimStrings{j.audience},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(j.secretKey)
	if err != nil {
		return "", fmt.Errorf("security.jwt.GenerateToken: failed to sign token: %w", err)
	}

	return signedToken, nil
}

func (j *JWTProvider) ValidateToken(tokenStr string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("security.jwt.ValidateToken: %w (got: %v)", entity.ErrTokenUnexpectedMethod, t.Header["alg"])
		}
		return j.secretKey, nil
	}, jwt.WithIssuer(j.issuer), jwt.WithAudience(j.audience))

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, fmt.Errorf("security.jwt.ValidateToken: %w", entity.ErrTokenExpired)
		}
		return nil, fmt.Errorf("security.jwt.ValidateToken: %w", entity.ErrTokenInvalid)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("security.jwt.ValidateToken: %w", entity.ErrTokenInvalid)
	}

	return claims, nil
}
