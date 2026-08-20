// Package auth xử lý xác thực người dùng.
// File này định nghĩa model User — cấu trúc dữ liệu cốt lõi đại diện
// cho một người dùng trong hệ thống Chia Đều.
package auth

import "time"

type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone,omitempty"`
	AvatarURL string    `json:"avatarUrl,omitempty"`
	CreatedAt time.Time `json:"createdAt"`

	PasswordHash string `json:"-"`
}
