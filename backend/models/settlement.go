package models

import "time"

// Trạng thái của kỳ quyết toán.
const (
	BatchStatusOpen      = "OPEN"
	BatchStatusCompleted = "COMPLETED"
	BatchStatusCancelled = "CANCELLED"
)

// Trạng thái của giao dịch hoàn tiền.
const (
	SettlementStatusPending              = "PENDING"
	SettlementStatusAwaitingConfirmation = "AWAITING_CONFIRMATION"
	SettlementStatusPaid                 = "PAID"
	SettlementStatusCancelled            = "CANCELLED"
)

// SettlementBatch là ảnh chụp bất biến của các khoản chi được chốt cùng lúc.
type SettlementBatch struct {
	ID             string     `json:"id"`
	GroupID        string     `json:"groupId"`
	CreatedBy      string     `json:"createdBy"`
	IdempotencyKey string     `json:"idempotencyKey"`
	Status         string     `json:"status"`
	CreatedAt      time.Time  `json:"createdAt"`
	CompletedAt    *time.Time `json:"completedAt,omitempty"`
	CancelledAt    *time.Time `json:"cancelledAt,omitempty"`
}

// Settlement là chỉ dẫn một thành viên (FromUserID) phải trả cho một thành
// viên khác (ToUserID).
type Settlement struct {
	ID           string     `json:"id"`
	BatchID      string     `json:"batchId"`
	FromUserID   string     `json:"fromUserId"`
	ToUserID     string     `json:"toUserId"`
	AmountMinor  int64      `json:"amountMinor"`
	PaymentCode  string     `json:"paymentCode"`
	Status       string     `json:"status"`
	MarkedSentAt *time.Time `json:"markedSentAt,omitempty"`
	PaidAt       *time.Time `json:"paidAt,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    *time.Time `json:"updatedAt,omitempty"`
}

// IsPayer cho biết người dùng có phải là người phải trả khoản này hay không.
func (s Settlement) IsPayer(userID string) bool {
	return s.FromUserID == userID
}

// IsRecipient cho biết người dùng có phải là người cần được nhận hay không.
func (s Settlement) IsRecipient(userID string) bool {
	return s.ToUserID == userID
}

// InTerminalState cho biết settlement đã đi tới trạng thái kết thúc.
func (s Settlement) InTerminalState() bool {
	return s.Status == SettlementStatusPaid || s.Status == SettlementStatusCancelled
}
