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
