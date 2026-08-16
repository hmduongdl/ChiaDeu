// Package auth — lưu trữ phiên refresh token để rotate và thu hồi phía server.
// Bảng sessions lưu hash của refresh token cùng jti; mỗi lần refresh, phiên cũ bị
// thu hồi và một phiên mới được cấp.
package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrSessionNotFound = errors.New("phiên không tồn tại")

// Session là một phiên đăng nhập gắn với một refresh token cụ thể.
type Session struct {
	JTI              string
	UserID           string
	RefreshTokenHash string
	CreatedAt        time.Time
	ExpiredAt        time.Time
	RevokedAt        *time.Time
}

// SessionStore giới hạn các thao tác lưu trữ phiên.
type SessionStore interface {
	CreateSession(ctx context.Context, session Session) error
	FindSessionByJTI(ctx context.Context, jti string) (Session, error)
	RevokeSession(ctx context.Context, jti string) error
}

type PostgresSessionStore struct {
	pool *pgxpool.Pool
}

func NewPostgresSessionStore(pool *pgxpool.Pool) *PostgresSessionStore {
	return &PostgresSessionStore{pool: pool}
}

func (s *PostgresSessionStore) CreateSession(ctx context.Context, session Session) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO sessions (id, user_id, refresh_token_hash, expired_at)
		VALUES ($1, $2, $3, $4)`,
		session.JTI, session.UserID, session.RefreshTokenHash, session.ExpiredAt)
	return err
}

func (s *PostgresSessionStore) FindSessionByJTI(ctx context.Context, jti string) (Session, error) {
	var session Session
	var revokedAt pgtype.Timestamptz
	err := s.pool.QueryRow(ctx, `
		SELECT id, user_id, refresh_token_hash, created_at, expired_at, revoked_at
		FROM sessions
		WHERE id = $1`, jti,
	).Scan(&session.JTI, &session.UserID, &session.RefreshTokenHash, &session.CreatedAt, &session.ExpiredAt, &revokedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Session{}, ErrSessionNotFound
	}
	if err != nil {
		return Session{}, err
	}
	if revokedAt.Valid {
		session.RevokedAt = &revokedAt.Time
	}
	return session, nil
}

func (s *PostgresSessionStore) RevokeSession(ctx context.Context, jti string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE sessions SET revoked_at = now() WHERE id = $1`, jti)
	return err
}

// HashRefreshToken băm refresh token để chỉ lưu hash, không lưu token thô.
func HashRefreshToken(rawToken string) string {
	sum := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(sum[:])
}
