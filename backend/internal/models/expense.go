package models

import "time"

// Expense đại diện cho một khoản chi cần chia trong nhóm.
// Có thể liên kết với một giao dịch ngân hàng gốc (source_transaction_id)
// hoặc được nhập tay (source_transaction_id = NULL).
// SplitType: EQUAL (chia đều), PERCENT (chia %), CUSTOM (chia tuỳ chỉnh).
type Expense struct {
	ID                  string    `json:"id"`                             // UUID khoản chi
	GroupID             string    `json:"group_id"`                       // Thuộc nhóm nào
	SourceTransactionID *string   `json:"source_transaction_id,omitempty"` // Giao dịch ngân hàng gốc (NULL nếu nhập tay)
	PaidBy              string    `json:"paid_by"`                        // Người bỏ tiền trả
	Description         string    `json:"description,omitempty"`          // Mô tả khoản chi
	Amount              float64   `json:"amount"`                         // Tổng số tiền
	SplitType           string    `json:"split_type"`                     // Cách chia: EQUAL / PERCENT / CUSTOM
	CreatedAt           time.Time `json:"created_at"`                     // Thời điểm tạo
}

// ExpenseSplit đại diện cho phần tiền của một người trong một khoản chi.
// Tổng share_amount của tất cả splits trong 1 expense = expense.amount.
type ExpenseSplit struct {
	ID          string  `json:"id"`           // UUID dòng split
	ExpenseID   string  `json:"expense_id"`   // Thuộc expense nào
	UserID      string  `json:"user_id"`      // Người phải chịu
	ShareAmount float64 `json:"share_amount"` // Số tiền người này phải chịu
}

// CreateExpenseInput là DTO đầu vào để tạo một khoản chi mới.
// Bao gồm thông tin khoản chi và danh sách splits cho từng thành viên.
type CreateExpenseInput struct {
	GroupID             string              `json:"group_id"`                       // UUID nhóm
	SourceTransactionID *string             `json:"source_transaction_id,omitempty"` // Giao dịch ngân hàng gốc (tuỳ chọn)
	PaidBy              string              `json:"paid_by"`                        // Người bỏ tiền trả
	Description         string              `json:"description,omitempty"`          // Mô tả
	Amount              float64             `json:"amount"`                         // Tổng tiền
	SplitType           string              `json:"split_type"`                     // EQUAL / PERCENT / CUSTOM
	Splits              []ExpenseSplitInput `json:"splits"`                         // Danh sách chia tiền cho từng người
}

// ExpenseSplitInput là DTO đầu vào cho phần chia của một người trong khoản chi.
type ExpenseSplitInput struct {
	UserID      string  `json:"user_id"`      // UUID người phải chịu
	ShareAmount float64 `json:"share_amount"` // Số tiền người này phải chịu
}