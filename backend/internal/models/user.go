// Package models chứa các struct đại diện cho bảng trong database và các DTO request/response.
// Mỗi file tương ứng với một bảng chính trong schema.
package models

import "time"

// User đại diện cho một người dùng trong hệ thống.
// Mỗi user có thể liên kết tài khoản ngân hàng với SePay để đồng bộ giao dịch tự động.
type User struct {
	ID             string    `json:"id"`               // UUID, tự sinh bởi PostgreSQL
	Name           string    `json:"name"`             // Tên hiển thị
	Phone          string    `json:"phone,omitempty"`  // SĐT, unique, dùng để mời bạn
	Email          string    `json:"email,omitempty"`  // Email, unique, dùng đăng nhập
	BankAccountNo  string    `json:"bank_account_no,omitempty"`  // Số tài khoản ngân hàng
	BankCode       string    `json:"bank_code,omitempty"`        // Mã ngân hàng (VD: MBBank)
	SepayAccountID string    `json:"sepay_account_id,omitempty"` // ID tài khoản trên SePay
	AvatarURL      string    `json:"avatar_url,omitempty"`       // URL ảnh đại diện
	CreatedAt      time.Time `json:"created_at"`                 // Thời điểm tạo tài khoản
}