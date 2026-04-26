package auth

import (
	"doproj/internal/config"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenManager struct {
	secretKey []byte
	tokenTTL  int
}

func NewTokenManager(cfg *config.Config) *TokenManager {
	return &TokenManager{
		secretKey: []byte(cfg.JWTSecret),
		tokenTTL:  cfg.JWTExpirationTime,
	}
}

type Claims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

func (m *TokenManager) GenerateToken(userID uint) (string, error) {
	expirationTime := time.Now().Add(time.Duration(m.tokenTTL) * time.Hour)

	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secretKey)
}

func (m *TokenManager) ValidateToken(tokenString string) (uint, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			// Algorithm check
			return nil, errors.New("unexpected signing method")
		}
		return m.secretKey, nil
	})
	if err != nil {
		return 0, err
	}

	if !token.Valid {
		return 0, errors.New("invalid token")
	}
	return claims.UserID, nil
}
