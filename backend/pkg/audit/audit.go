// Package audit ghi nhật ký nghiệp vụ append-only vào bảng audit_logs.
// Mọi thay đổi trạng thái quan trọng nên ghi một Entry; không lưu JWT, mật khẩu
// hay bí mật trong Metadata.
package audit

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgconn"
)

// Entry là một dòng nhật ký nghiệp vụ.
type Entry struct {
	GroupID    string
	ActorID    string
	Action     string
	EntityType string
	EntityID   string
	Metadata   map[string]any
}

// Execer cho phép ghi log trong hoặc ngoài transaction; *pgxpool.Pool và pgx.Tx
// đều đáp ứng interface này.
type Execer interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

// Append chèn một dòng audit_logs qua execer (pool hoặc transaction hiện tại).
func Append(ctx context.Context, execer Execer, entry Entry) error {
	var metadataArg any
	if entry.Metadata != nil {
		raw, err := json.Marshal(entry.Metadata)
		if err != nil {
			return err
		}
		metadataArg = raw
	}

	_, err := execer.Exec(ctx, `
		INSERT INTO audit_logs (group_id, actor_id, action, entity_type, entity_id, metadata)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		entry.GroupID, entry.ActorID, entry.Action, entry.EntityType, entry.EntityID, metadataArg)
	return err
}
