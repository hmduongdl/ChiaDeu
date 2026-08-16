// Package models định nghĩa các cấu trúc dữ liệu domain của ứng dụng Chia Đều
// theo schema mục tiêu trong schema.md. Đây là tầng thuần túy, không phụ thuộc
// database hay HTTP.
package models

import "time"

// User là thông tin cá nhân của một người dùng.
type User struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Email     string     `json:"email"`
	Phone     string     `json:"phone,omitempty"`
	AvatarURL string     `json:"avatarUrl,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`
}

// DisplayName trả về tên hiển thị, không bao giờ rỗng ngay cả khi người dùng
// thiếu thông tin tên.
func (u User) DisplayName() string {
	if u.Name != "" {
		return u.Name
	}
	return u.Email
}
