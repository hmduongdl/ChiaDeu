package models

import "time"

// BankTransaction đại diện cho một giao dịch ngân hàng được đồng bộ từ SePay.
// Đây là nguồn dữ liệu chính để người dùng chọn khi tạo Expense.
// TransactionType: IN (nhận tiền), OUT (chi tiền).
// IsUsed: false = chưa gán vào expense nào, true = đã được dùng.
type BankTransaction struct {
	ID                 string    `json:"id"`                   // UUID nội bộ
	UserID             string    `json:"user_id"`              // Chủ tài khoản
	SepayTransactionID string    `json:"sepay_transaction_id"` // ID giao dịch từ SePay (unique)
	Amount             float64   `json:"amount"`               // Số tiền giao dịch
	TransactionType    string    `json:"transaction_type"`     // IN hoặc OUT
	Description        string    `json:"description,omitempty"` // Nội dung chuyển khoản gốc
	BankAccountNo      string    `json:"bank_account_no,omitempty"` // Số tài khoản phát sinh
	IsUsed             bool      `json:"is_used"`              // Đã gán vào expense chưa
	TransactionTime    time.Time `json:"transaction_time"`     // Thời điểm giao dịch thực tế (từ SePay)
	ReceivedAt         time.Time `json:"received_at"`          // Thời điểm server nhận webhook
}