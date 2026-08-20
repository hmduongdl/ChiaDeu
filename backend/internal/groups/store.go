// Package groups chứa nghiệp vụ nhóm: membership, quyền quản trị và mã chia sẻ.
package groups

import (
	"context"
	"errors"

	"github.com/hmduongdl/ChiaDeu/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrGroupNotFound     = errors.New("nhóm không tồn tại")
	ErrShareCodeExists   = errors.New("mã chia sẻ đã tồn tại")
	ErrAlreadyMember     = errors.New("bạn đã là thành viên nhóm này")
	ErrUserNotFound      = errors.New("người dùng không tồn tại")
	ErrNotMember         = errors.New("bạn không phải là thành viên nhóm")
	ErrMembershipMissing = errors.New("không tìm thấy thành viên")
)

// Store giới hạn các thao tác nhóm, giúp handler/service kiểm thử bằng fake.
type Store interface {
	CreateGroupWithAdmin(ctx context.Context, group models.Group, admin models.GroupMember) error
	GetGroup(ctx context.Context, groupID string) (models.Group, error)
	GetGroupByShareCode(ctx context.Context, shareCode string) (models.Group, error)
	GetMembership(ctx context.Context, groupID, userID string) (models.GroupMember, error)
	CreateMembership(ctx context.Context, member models.GroupMember) error
	ListActiveMembers(ctx context.Context, groupID string) ([]models.User, error)
	ListUserGroups(ctx context.Context, userID string) ([]models.Group, error)
	UserExists(ctx context.Context, userID string) (bool, error)
}

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

const groupColumns = `id, name, created_by, share_code, currency, status, created_at, updated_at`

func scanGroup(rows pgx.Row) (models.Group, error) {
	var group models.Group
	var updatedAt pgtype.Timestamptz
	if err := rows.Scan(&group.ID, &group.Name, &group.CreatedBy, &group.ShareCode, &group.Currency, &group.Status, &group.CreatedAt, &updatedAt); err != nil {
		return models.Group{}, err
	}
	if updatedAt.Valid {
		group.UpdatedAt = &updatedAt.Time
	}
	return group, nil
}

func (s *PostgresStore) CreateGroupWithAdmin(ctx context.Context, group models.Group, admin models.GroupMember) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO groups (name, created_by, share_code, currency)
		VALUES ($1, $2, $3, $4)
		RETURNING id`,
		group.Name, group.CreatedBy, group.ShareCode, group.Currency,
	).Scan(&group.ID)
	if err != nil {
		return mapUniqueViolation(err, "groups_share_code_key", ErrShareCodeExists)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO group_members (group_id, user_id, role, status)
		VALUES ($1, $2, $3, $4)`,
		group.ID, admin.UserID, admin.Role, admin.Status,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *PostgresStore) GetGroup(ctx context.Context, groupID string) (models.Group, error) {
	return scanGroup(s.pool.QueryRow(ctx, `SELECT `+groupColumns+` FROM groups WHERE id = $1`, groupID))
}

func (s *PostgresStore) GetGroupByShareCode(ctx context.Context, shareCode string) (models.Group, error) {
	return scanGroup(s.pool.QueryRow(ctx, `SELECT `+groupColumns+` FROM groups WHERE share_code = $1`, shareCode))
}

func (s *PostgresStore) GetMembership(ctx context.Context, groupID, userID string) (models.GroupMember, error) {
	var member models.GroupMember
	var leftAt pgtype.Timestamptz
	err := s.pool.QueryRow(ctx, `
		SELECT group_id, user_id, role, status, joined_at, left_at
		FROM group_members
		WHERE group_id = $1 AND user_id = $2`, groupID, userID,
	).Scan(&member.GroupID, &member.UserID, &member.Role, &member.Status, &member.JoinedAt, &leftAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.GroupMember{}, ErrMembershipMissing
	}
	if err != nil {
		return models.GroupMember{}, err
	}
	if leftAt.Valid {
		member.LeftAt = &leftAt.Time
	}
	return member, nil
}

func (s *PostgresStore) CreateMembership(ctx context.Context, member models.GroupMember) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO group_members (group_id, user_id, role, status)
		VALUES ($1, $2, $3, $4)`,
		member.GroupID, member.UserID, member.Role, member.Status,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrAlreadyMember
		}
		return err
	}
	return nil
}

func (s *PostgresStore) ListActiveMembers(ctx context.Context, groupID string) ([]models.User, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT u.id, u.name, COALESCE(u.email, ''), COALESCE(u.phone, ''), COALESCE(u.avatar_url, ''), u.created_at
		FROM group_members m
		JOIN users u ON u.id = m.user_id
		WHERE m.group_id = $1 AND m.status = $2
		ORDER BY m.joined_at`, groupID, models.MemberStatusActive)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Phone, &user.AvatarURL, &user.CreatedAt); err != nil {
			return nil, err
		}
		members = append(members, user)
	}
	return members, rows.Err()
}

func (s *PostgresStore) ListUserGroups(ctx context.Context, userID string) ([]models.Group, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT g.`+groupColumns+`
		FROM groups g
		JOIN group_members m ON g.id = m.group_id
		WHERE m.user_id = $1 AND m.status = $2
		ORDER BY g.created_at DESC`, userID, models.MemberStatusActive)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []models.Group
	for rows.Next() {
		group, err := scanGroup(rows)
		if err != nil {
			return nil, err
		}
		groups = append(groups, group)
	}
	return groups, rows.Err()
}

func (s *PostgresStore) UserExists(ctx context.Context, userID string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE id = $1)`, userID).Scan(&exists)
	return exists, err
}

// mapUniqueViolation chuyển lỗi unique violation có tên constraint ưng ý thành lỗi
// sentinel, giữ nguyên lỗi khác.
func mapUniqueViolation(err error, constraint string, target error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == constraint {
		return target
	}
	return err
}
