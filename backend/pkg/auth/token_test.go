// Package auth — unit test cho TokenManager.
// File này kiểm thử:
//   - Access token và refresh token có thể verify riêng biệt
//   - Token hết hạn bị từ chối
//   - Secret quá ngắn hoặc trùng nhau bị từ chối khi khởi tạo
package auth

import (
	"errors"
	"testing"
	"time"
)

const (
	testAccessSecret  = "access-secret-that-is-longer-than-thirty-two-characters"
	testRefreshSecret = "refresh-secret-that-is-longer-than-thirty-two-characters"
)

func TestTokenManagerCreatesDistinctTypedTokens(t *testing.T) {
	manager, err := NewTokenManager(testAccessSecret, testRefreshSecret)
	if err != nil {
		t.Fatalf("create token manager: %v", err)
	}

	pair, err := manager.CreatePair("user-123")
	if err != nil {
		t.Fatalf("create token pair: %v", err)
	}
	if userID, err := manager.VerifyAccessToken(pair.AccessToken); err != nil || userID != "user-123" {
		t.Fatalf("verify access token: id=%q err=%v", userID, err)
	}
	if userID, err := manager.VerifyRefreshToken(pair.RefreshToken); err != nil || userID != "user-123" {
		t.Fatalf("verify refresh token: id=%q err=%v", userID, err)
	}
	if _, err := manager.VerifyRefreshToken(pair.AccessToken); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("access token must not be accepted as refresh token: %v", err)
	}
	if _, err := manager.VerifyAccessToken(pair.RefreshToken); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("refresh token must not be accepted as access token: %v", err)
	}
}

func TestTokenManagerRejectsExpiredToken(t *testing.T) {
	manager, err := NewTokenManager(testAccessSecret, testRefreshSecret)
	if err != nil {
		t.Fatalf("create token manager: %v", err)
	}
	manager.now = func() time.Time { return time.Now().Add(-AccessTokenDuration - time.Minute) }

	token, err := manager.CreateAccessToken("user-123")
	if err != nil {
		t.Fatalf("create expired token: %v", err)
	}
	if _, err := manager.VerifyAccessToken(token); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expired token must be rejected: %v", err)
	}
}

func TestTokenManagerRequiresIndependentStrongSecrets(t *testing.T) {
	if _, err := NewTokenManager("short", testRefreshSecret); err == nil {
		t.Fatal("expected short secret to fail")
	}
	if _, err := NewTokenManager(testAccessSecret, testAccessSecret); err == nil {
		t.Fatal("expected identical secrets to fail")
	}
}
