package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTService struct {
	secret          []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

type Claims struct {
	Email string `json:"email"`
	Role  string `json:"role"`
	jwt.RegisteredClaims
}

func NewJWTService(secret string, accessTokenTTL, refreshTokenTTL time.Duration) *JWTService {
	return &JWTService{secret: []byte(secret), accessTokenTTL: accessTokenTTL, refreshTokenTTL: refreshTokenTTL}
}

func (s *JWTService) GenerateAccessToken(_ context.Context, userID, email string, role string) (string, error) {
	now := time.Now().UTC()
	claims := Claims{
		Email: email,
		Role:  role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

func (s *JWTService) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func (s *JWTService) ValidateToken(tokenString string) (string, string, string, error) {
	claims, err := s.ParseToken(tokenString)
	if err != nil {
		return "", "", "", err
	}

	return claims.Subject, claims.Email, claims.Role, nil
}

func (s *JWTService) NewRefreshToken() (string, string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", "", err
	}

	plain := base64.RawURLEncoding.EncodeToString(raw)
	return plain, s.HashRefreshToken(plain), nil
}

func (s *JWTService) HashRefreshToken(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:])
}

func (s *JWTService) RefreshTokenTTL() time.Duration {
	return s.refreshTokenTTL
}
