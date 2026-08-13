// Package auth — tầng nghiệp vụ xác thực.
// File này chứa AuthService với các chức năng:
//   - Register: đăng ký tài khoản mới (validate input, hash mật khẩu bằng bcrypt)
//   - Authenticate: xác thực email/password (dùng dummy hash để chống timing attack)
//   - CurrentUser: lấy thông tin user hiện tại từ ID
package auth

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

// Comparing against a fixed valid hash keeps unknown-email login timing close to wrong-password timing.
var dummyPasswordHash = []byte("$2a$10$7EqJtq98hPqEX7fNZaFWoO/e5p.KPdHnOAMi8rA0Z6gbQYc1vHUSa")

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

type Service struct {
	store UserStore
}

func NewService(store UserStore) *Service {
	return &Service{store: store}
}

func (s *Service) Register(ctx context.Context, name, email, password string) (User, error) {
	name = strings.TrimSpace(name)
	email = normalizeEmail(email)

	if name == "" || utf8.RuneCountInString(name) > 100 {
		return User{}, &ValidationError{Message: "name must be between 1 and 100 characters"}
	}
	if !validEmail(email) || len(email) > 100 {
		return User{}, &ValidationError{Message: "a valid email is required"}
	}
	// bcrypt only considers the first 72 password bytes, so reject longer values.
	if utf8.RuneCountInString(password) < 8 || len(password) > 72 {
		return User{}, &ValidationError{Message: "password must contain at least 8 characters and at most 72 bytes"}
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, fmt.Errorf("hash password: %w", err)
	}

	return s.store.CreateUser(ctx, name, email, string(passwordHash))
}

func (s *Service) Authenticate(ctx context.Context, email, password string) (User, error) {
	user, err := s.store.FindUserByEmail(ctx, normalizeEmail(email))
	if errors.Is(err, ErrUserNotFound) {
		_ = bcrypt.CompareHashAndPassword(dummyPasswordHash, []byte(password))
		return User{}, ErrInvalidCredentials
	}
	if err != nil {
		return User{}, err
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return User{}, ErrInvalidCredentials
	}

	user.PasswordHash = ""
	return user, nil
}

func (s *Service) CurrentUser(ctx context.Context, id string) (User, error) {
	user, err := s.store.FindUserByID(ctx, id)
	if err != nil {
		return User{}, err
	}
	user.PasswordHash = ""
	return user, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func validEmail(email string) bool {
	address, err := mail.ParseAddress(email)
	return err == nil && address.Address == email
}
