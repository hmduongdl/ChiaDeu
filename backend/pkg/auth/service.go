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
	"time"
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
	store        UserStore
	sessionStore SessionStore
	tokens       *TokenManager
}

func NewService(store UserStore) *Service {
	return &Service{store: store}
}

// NewServiceWithSessions bật chế độ phiên có trạng thái: refresh token được lưu
// trong bảng sessions, rotate trên mỗi lần refresh và có thể thu hồi.
func NewServiceWithSessions(store UserStore, tokens *TokenManager, sessions SessionStore) *Service {
	return &Service{store: store, sessionStore: sessions, tokens: tokens}
}

// HasSessions cho biết service có lưu phiên phía server hay không.
func (s *Service) HasSessions() bool {
	return s.sessionStore != nil
}

// StartSession cấp cặp token và lưu phiên nếu được cấu hình.
func (s *Service) StartSession(ctx context.Context, userID string) (TokenPair, error) {
	if !s.HasSessions() {
		if s.tokens == nil {
			return TokenPair{}, ErrSessionsUnavailable
		}
		return s.tokens.CreatePair(userID)
	}
	pair, err := s.tokens.CreatePair(userID)
	if err != nil {
		return TokenPair{}, err
	}
	session := Session{
		JTI:              pair.RefreshJTI,
		UserID:           userID,
		RefreshTokenHash: HashRefreshToken(pair.RefreshToken),
		CreatedAt:        time.Now(),
		ExpiredAt:        time.Now().Add(RefreshTokenDuration),
	}
	if err := s.sessionStore.CreateSession(ctx, session); err != nil {
		return TokenPair{}, err
	}
	return pair, nil
}

// RotateSession xác thực refresh token và thu hồi phiên cũ, cấp phiên mới.
func (s *Service) RotateSession(ctx context.Context, rawRefresh string) (TokenPair, error) {
	if !s.HasSessions() {
		return TokenPair{}, ErrSessionsUnavailable
	}
	userID, jti, err := s.tokens.VerifyRefreshDetails(rawRefresh)
	if err != nil {
		return TokenPair{}, ErrInvalidToken
	}
	session, err := s.sessionStore.FindSessionByJTI(ctx, jti)
	if errors.Is(err, ErrSessionNotFound) {
		return TokenPair{}, ErrInvalidToken
	}
	if err != nil {
		return TokenPair{}, err
	}
	if session.RevokedAt != nil {
		return TokenPair{}, ErrSessionRevoked
	}
	if !session.ExpiredAt.After(time.Now()) {
		return TokenPair{}, ErrSessionExpired
	}
	if session.RefreshTokenHash != HashRefreshToken(rawRefresh) {
		return TokenPair{}, ErrInvalidToken
	}
	if session.UserID != userID {
		return TokenPair{}, ErrInvalidToken
	}
	if _, err := s.store.FindUserByID(ctx, userID); err != nil {
		return TokenPair{}, ErrInvalidToken
	}
	if err := s.sessionStore.RevokeSession(ctx, jti); err != nil {
		return TokenPair{}, err
	}
	return s.StartSession(ctx, userID)
}

// RevokeSession thu hồi phiên của refresh token (dùng cho logout); bỏ qua nếu
// token không hợp lệ để đăng xuất luôn thành công.
func (s *Service) RevokeSession(ctx context.Context, rawRefresh string) error {
	if !s.HasSessions() {
		return nil
	}
	_, jti, err := s.tokens.VerifyRefreshDetails(rawRefresh)
	if err != nil {
		return nil
	}
	return s.sessionStore.RevokeSession(ctx, jti)
}

// ErrSessionsUnavailable báo service chưa được cấu hình phiên có trạng thái.
var ErrSessionsUnavailable = errors.New("session store chưa được cấu hình")

var (
	// ErrSessionRevoked báo phiên đã bị thu hồi phía server.
	ErrSessionRevoked = errors.New("phiên đã bị thu hồi")
	// ErrSessionExpired báo phiên đã hết hạn.
	ErrSessionExpired = errors.New("phiên đã hết hạn")
)

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
