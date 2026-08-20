// Package auth — tầng truy xuất dữ liệu người dùng.
// File này định nghĩa:
//   - UserStore interface:抽象 các thao tác CRUD với bảng users
//   - PostgresUserStore: triển khai UserStore dùng pgxpool
//   - Các lỗi sentinel: ErrEmailExists, ErrUserNotFound
//   - Xử lý unique constraint violation (mã 23505) từ PostgreSQL
package auth

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrEmailExists  = errors.New("email already exists")
	ErrUserNotFound = errors.New("user not found")
)

type UserStore interface {
	CreateUser(ctx context.Context, name, email, passwordHash string) (User, error)
	FindUserByEmail(ctx context.Context, email string) (User, error)
	FindUserByID(ctx context.Context, id string) (User, error)
}

type PostgresUserStore struct {
	pool *pgxpool.Pool
}

func NewPostgresUserStore(pool *pgxpool.Pool) *PostgresUserStore {
	return &PostgresUserStore{pool: pool}
}

func (s *PostgresUserStore) CreateUser(ctx context.Context, name, email, passwordHash string) (User, error) {
	const query = `
		INSERT INTO users (name, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, name, COALESCE(email, ''), COALESCE(phone, ''), COALESCE(avatar_url, ''), created_at`

	var user User
	err := s.pool.QueryRow(ctx, query, name, email, passwordHash).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Phone,
		&user.AvatarURL,
		&user.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return User{}, ErrEmailExists
		}
		return User{}, err
	}

	return user, nil
}

func (s *PostgresUserStore) FindUserByEmail(ctx context.Context, email string) (User, error) {
	return s.findUser(ctx, `
		SELECT id, name, COALESCE(email, ''), COALESCE(phone, ''), COALESCE(avatar_url, ''), created_at, COALESCE(password_hash, '')
		FROM users
		WHERE email = $1`, email)
}

func (s *PostgresUserStore) FindUserByID(ctx context.Context, id string) (User, error) {
	return s.findUser(ctx, `
		SELECT id, name, COALESCE(email, ''), COALESCE(phone, ''), COALESCE(avatar_url, ''), created_at, COALESCE(password_hash, '')
		FROM users
		WHERE id = $1`, id)
}

func (s *PostgresUserStore) findUser(ctx context.Context, query, value string) (User, error) {
	var user User
	err := s.pool.QueryRow(ctx, query, value).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Phone,
		&user.AvatarURL,
		&user.CreatedAt,
		&user.PasswordHash,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrUserNotFound
	}
	if err != nil {
		return User{}, err
	}

	return user, nil
}
