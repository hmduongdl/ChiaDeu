package models

import (
	"strings"
	"time"
)

// Vai trò của thành viên trong nhóm. ADMIN chỉ quản lý vận hành nhóm, không
// quyết định hướng dòng tiền.
const (
	RoleAdmin  = "ADMIN"
	RoleMember = "MEMBER"
)

// Trạng thái membership.
const (
	MemberStatusActive = "ACTIVE"
	MemberStatusLeft   = "LEFT"
)

const (
	GroupStatusActive   = "ACTIVE"
	GroupStatusArchived = "ARCHIVED"
)

// Group là không gian chung chứa thành viên, khoản chi và công nợ.
type Group struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	CreatedBy string     `json:"createdBy"`
	ShareCode string     `json:"shareCode"`
	Currency  string     `json:"currency"`
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`
}

// GroupMember là quan hệ giữa một người dùng và một nhóm.
type GroupMember struct {
	GroupID  string     `json:"groupId"`
	UserID   string     `json:"userId"`
	Role     string     `json:"role"`
	Status   string     `json:"status"`
	JoinedAt time.Time  `json:"joinedAt"`
	LeftAt   *time.Time `json:"leftAt,omitempty"`
}

// IsActiveMember cho biết thành viên hiện đang hoạt động trong nhóm.
func (m GroupMember) IsActiveMember() bool {
	return m.Status == MemberStatusActive
}

// IsAdmin cho biết thành viên có quyền quản trị nhóm.
func (m GroupMember) IsAdmin() bool {
	return m.IsActiveMember() && m.Role == RoleAdmin
}

// NormalizeShareCode chuẩn hóa mã chia sẻ người dùng nhập: bỏ khoảng trắng và
// chuyển về chữ in hoa để so khớp không phân biệt hoa/thường.
func NormalizeShareCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}
