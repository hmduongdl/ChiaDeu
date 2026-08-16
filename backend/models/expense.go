package models

import "time"

// Kiểu chia của khoản chi.
const (
	SplitTypeEqual   = "EQUAL"
	SplitTypePercent = "PERCENT"
	SplitTypeWeight  = "WEIGHT"
	SplitTypeCustom  = "CUSTOM"
)

const (
	ExpenseStatusActive = "ACTIVE"
	ExpenseStatusVoided = "VOIDED"
)

// Expense là khoản tiền một thành viên (PaidBy) đã ứng cho một hoặc nhiều
// người trong nhóm.
type Expense struct {
	ID          string     `json:"id"`
	GroupID     string     `json:"groupId"`
	CreatedBy   string     `json:"createdBy"`
	PaidBy      string     `json:"paidBy"`
	Description string     `json:"description"`
	AmountMinor int64      `json:"amountMinor"`
	SplitType   string     `json:"splitType"`
	ExpenseDate time.Time  `json:"expenseDate"`
	BatchID     *string    `json:"batchId,omitempty"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   *time.Time `json:"updatedAt,omitempty"`
}

// ExpenseSplit là phần tiền một thành viên phải chịu trong một khoản chi.
type ExpenseSplit struct {
	ID         string    `json:"id"`
	ExpenseID  string    `json:"expenseId"`
	UserID     string    `json:"userId"`
	ShareMinor int64     `json:"shareMinor"`
	CreatedAt  time.Time `json:"createdAt"`
}

// IsUnsettled cho biết khoản chi còn treo, chưa bị hủy và chưa đưa vào kỳ nào.
func (e Expense) IsUnsettled() bool {
	return e.Status == ExpenseStatusActive && e.BatchID == nil
}
