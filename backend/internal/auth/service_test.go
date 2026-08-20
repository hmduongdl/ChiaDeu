// Package auth — unit test cho AuthService.
// File này kiểm thử:
//   - Đăng ký hash mật khẩu và chuẩn hóa email
//   - Xác thực từ chối sai mật khẩu
//   - Validate input đăng ký (thiếu tên, email sai, mật khẩu ngắn)
package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type memoryUserStore struct {
	user User
}

func (s *memoryUserStore) CreateUser(_ context.Context, name, email, passwordHash string) (User, error) {
	if s.user.Email == email {
		return User{}, ErrEmailExists
	}
	s.user = User{ID: "user-123", Name: name, Email: email, PasswordHash: passwordHash, CreatedAt: time.Now()}
	return s.user, nil
}

func (s *memoryUserStore) FindUserByEmail(_ context.Context, email string) (User, error) {
	if s.user.Email != email {
		return User{}, ErrUserNotFound
	}
	return s.user, nil
}

func (s *memoryUserStore) FindUserByID(_ context.Context, id string) (User, error) {
	if s.user.ID != id {
		return User{}, ErrUserNotFound
	}
	return s.user, nil
}

func TestServiceRegisterHashesPasswordAndNormalizesEmail(t *testing.T) {
	store := &memoryUserStore{}
	service := NewService(store)
	user, err := service.Register(context.Background(), " Test User ", " TEST@Example.COM ", "password123")
	if err != nil {
		t.Fatalf("register user: %v", err)
	}
	if user.Email != "test@example.com" || user.Name != "Test User" {
		t.Fatalf("unexpected normalized user: %+v", user)
	}
	if store.user.PasswordHash == "password123" {
		t.Fatal("password was stored in plaintext")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(store.user.PasswordHash), []byte("password123")); err != nil {
		t.Fatalf("stored bcrypt hash does not match: %v", err)
	}
}

func TestServiceAuthenticateRejectsWrongPassword(t *testing.T) {
	store := &memoryUserStore{}
	service := NewService(store)
	if _, err := service.Register(context.Background(), "Test User", "test@example.com", "password123"); err != nil {
		t.Fatalf("register user: %v", err)
	}
	if _, err := service.Authenticate(context.Background(), "test@example.com", "wrong-password"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials, got %v", err)
	}
}

func TestServiceValidatesRegistration(t *testing.T) {
	service := NewService(&memoryUserStore{})
	for name, input := range map[string]struct{ name, email, password string }{
		"missing name":   {"", "test@example.com", "password123"},
		"invalid email":  {"Test", "not-an-email", "password123"},
		"short password": {"Test", "test@example.com", "short"},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := service.Register(context.Background(), input.name, input.email, input.password); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

type memorySessionStore struct {
	sessions map[string]Session
}

func (s *memorySessionStore) CreateSession(_ context.Context, session Session) error {
	if s.sessions == nil {
		s.sessions = make(map[string]Session)
	}
	s.sessions[session.JTI] = session
	return nil
}

func (s *memorySessionStore) FindSessionByJTI(_ context.Context, jti string) (Session, error) {
	session, ok := s.sessions[jti]
	if !ok {
		return Session{}, ErrSessionNotFound
	}
	return session, nil
}

func (s *memorySessionStore) RevokeSession(_ context.Context, jti string) error {
	session, ok := s.sessions[jti]
	if !ok {
		return ErrSessionNotFound
	}
	revokedAt := time.Now()
	session.RevokedAt = &revokedAt
	s.sessions[jti] = session
	return nil
}

func TestServiceRotatesAndRevokesSessions(t *testing.T) {
	tokens, err := NewTokenManager(
		"session-test-access-secret-with-at-least-thirty-two-characters",
		"session-test-refresh-secret-with-at-least-thirty-two-characters",
	)
	if err != nil {
		t.Fatalf("create token manager: %v", err)
	}
	store := &memoryUserStore{user: User{ID: "user-123", Name: "Test", Email: "test@example.com", CreatedAt: time.Now()}}
	sessions := &memorySessionStore{}
	service := NewServiceWithSessions(store, tokens, sessions)

	first, err := service.StartSession(context.Background(), "user-123")
	if err != nil {
		t.Fatalf("start session: %v", err)
	}
	if sessions.sessions[first.RefreshJTI].RefreshTokenHash == "" {
		t.Fatal("session phải lưu hash refresh token")
	}

	second, err := service.RotateSession(context.Background(), first.RefreshToken)
	if err != nil {
		t.Fatalf("rotate session: %v", err)
	}
	if second.RefreshToken == first.RefreshToken {
		t.Fatal("refresh token phải được xoay sang token mới")
	}
	if sessions.sessions[first.RefreshJTI].RevokedAt == nil {
		t.Fatal("phiên cũ phải bị thu hồi sau khi rotate")
	}

	// Token cũ đã thu hồi không thể dùng lại.
	if _, err := service.RotateSession(context.Background(), first.RefreshToken); !errors.Is(err, ErrSessionRevoked) {
		t.Fatalf("mong đợi ErrSessionRevoked, got %v", err)
	}

	// Đăng xuất thu hồi phiên hiện tại.
	if err := service.RevokeSession(context.Background(), second.RefreshToken); err != nil {
		t.Fatalf("revoke session: %v", err)
	}
	if sessions.sessions[second.RefreshJTI].RevokedAt == nil {
		t.Fatal("logout phải thu hồi phiên")
	}
}
