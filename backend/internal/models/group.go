package models

import "time"

// Group đại diện cho một nhóm chi tiêu (VD: nhóm đi du lịch, nhóm ăn uống...).
// Mỗi nhóm có một share_code ngắn dùng để mời thành viên tham gia.
type Group struct {
	ID        string    `json:"id"`         // UUID nhóm
	Name      string    `json:"name"`       // Tên nhóm (VD: "Đà Lạt trip 2024")
	ShareCode string    `json:"share_code"` // Mã mời ngắn (6-10 ký tự), unique
	CreatedBy string    `json:"created_by"` // UUID của người tạo nhóm
	Currency  string    `json:"currency"`   // Đơn vị tiền tệ, mặc định VND
	CreatedAt time.Time `json:"created_at"` // Thời điểm tạo nhóm
}

// GroupMember đại diện cho một thành viên trong nhóm.
// Bảng group_members có PK kép (group_id, user_id).
type GroupMember struct {
	GroupID  string    `json:"group_id"`  // UUID nhóm
	UserID   string    `json:"user_id"`   // UUID thành viên
	JoinedAt time.Time `json:"joined_at"` // Thời điểm tham gia
}