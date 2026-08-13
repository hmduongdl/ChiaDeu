// Package auth — quản lý JWT token.
// File này cung cấp TokenManager để:
//   - Tạo cặp access + refresh token (HS256, có issuer và expiry)
//   - Xác thực token (kiểm tra chữ ký, thời hạn, loại token)
//   - Phân biệt access token (15 phút) và refresh token (7 ngày)
//   - Yêu cầu mỗi secret dài ít nhất 32 ký tự và khác nhau
package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	AccessTokenDuration  = 15 * time.Minute
	RefreshTokenDuration = 7 * 24 * time.Hour

	accessTokenType  = "access"
	refreshTokenType = "refresh"
)

var ErrInvalidToken = errors.New("invalid token")

type Claims struct {
	TokenType string `json:"tokenType"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type TokenManager struct {
	accessSecret  []byte
	refreshSecret []byte
	now           func() time.Time
}

func NewTokenManager(accessSecret, refreshSecret string) (*TokenManager, error) {
	if len(accessSecret) < 32 || len(refreshSecret) < 32 {
		return nil, errors.New("JWT secrets must each contain at least 32 characters")
	}
	if accessSecret == refreshSecret {
		return nil, errors.New("JWT access and refresh secrets must be different")
	}

	return &TokenManager{
		accessSecret:  []byte(accessSecret),
		refreshSecret: []byte(refreshSecret),
		now:           time.Now,
	}, nil
}

func (m *TokenManager) CreatePair(userID string) (TokenPair, error) {
	accessToken, err := m.create(userID, accessTokenType, AccessTokenDuration, m.accessSecret)
	if err != nil {
		return TokenPair{}, err
	}
	refreshToken, err := m.create(userID, refreshTokenType, RefreshTokenDuration, m.refreshSecret)
	if err != nil {
		return TokenPair{}, err
	}
	return TokenPair{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (m *TokenManager) CreateAccessToken(userID string) (string, error) {
	return m.create(userID, accessTokenType, AccessTokenDuration, m.accessSecret)
}

func (m *TokenManager) VerifyAccessToken(rawToken string) (string, error) {
	return m.verify(rawToken, accessTokenType, m.accessSecret)
}

func (m *TokenManager) VerifyRefreshToken(rawToken string) (string, error) {
	return m.verify(rawToken, refreshTokenType, m.refreshSecret)
}

func (m *TokenManager) create(userID, tokenType string, duration time.Duration, secret []byte) (string, error) {
	now := m.now()
	claims := Claims{
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			Issuer:    "chiadeu-api",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("sign JWT: %w", err)
	}
	return token, nil
}

func (m *TokenManager) verify(rawToken, expectedType string, secret []byte) (string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(
		rawToken,
		claims,
		func(token *jwt.Token) (any, error) { return secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer("chiadeu-api"),
		jwt.WithExpirationRequired(),
	)
	if err != nil || !token.Valid || claims.TokenType != expectedType || claims.Subject == "" {
		return "", ErrInvalidToken
	}
	return claims.Subject, nil
}
