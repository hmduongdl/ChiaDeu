package models

import "time"

// Settlement đại diện cho một giao dịch thanh toán tối ưu giữa 2 người trong nhóm.
// Được sinh ra sau khi chạy thuật toán Minimize Cash Flow.
// Status: PENDING (chưa thanh toán), PAID (đã thanh toán), CANCELLED (đã huỷ).
// PaymentMethod: PAYOS_QR, MOMO, SEPAY_TRANSFER, CASH.
type Settlement struct {
	ID                     string     `json:"id"`                               // UUID settlement
	GroupID                string     `json:"group_id"`                         // Thuộc nhóm nào
	FromUser               string     `json:"from_user"`                        // Người phải trả (debtor)
	ToUser                 string     `json:"to_user"`                          // Người nhận (creditor)
	Amount                 float64    `json:"amount"`                           // Số tiền cần thanh toán
	Status                 string     `json:"status"`                           // PENDING / PAID / CANCELLED
	PaymentMethod          string     `json:"payment_method,omitempty"`         // PAYOS_QR / MOMO / SEPAY_TRANSFER / CASH
	QRCodeData             string     `json:"qr_code_data,omitempty"`           // Payload QR hoặc payment link
	ConfirmedTransactionID *string    `json:"confirmed_transaction_id,omitempty"` // Giao dịch xác nhận thanh toán
	PaidAt                 *time.Time `json:"paid_at,omitempty"`                // Thời điểm thanh toán thành công
	CreatedAt              time.Time  `json:"created_at"`                       // Thời điểm tạo settlement
}