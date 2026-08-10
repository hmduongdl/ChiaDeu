package services

import (
	"database/sql"

	"github.com/yourusername/cash-flow-minimizer/internal/models"
)

// SettlementService xử lý tất cả nghiệp vụ liên quan đến thanh toán:
// - Gọi thuật toán Minimize Cash Flow để tính settlements tối ưu
// - Sinh QR code / payment link qua PayOS hoặc MoMo
// - Xác nhận thanh toán khi nhận webhook từ cổng thanh toán
type SettlementService struct {
	db *sql.DB
}

// NewSettlementService khởi tạo service với database connection.
func NewSettlementService(db *sql.DB) *SettlementService {
	return &SettlementService{db: db}
}

// SettleGroup tính toán danh sách giao dịch thanh toán tối thiểu cho một nhóm.
// Luồng xử lý:
//  1. Tính net balance từng thành viên (gọi ExpenseService.GetGroupBalances)
//  2. Gọi C++ algorithm.MinimizeCashFlow(balances) qua cgo bridge
//  3. Lưu từng settlement vào bảng settlements
//  4. Trả về danh sách settlements
// Độ phức tạp thuật toán: O(N log N) với N là số thành viên trong nhóm.
func (s *SettlementService) SettleGroup(groupID string) ([]models.Settlement, error) {
	// TODO: Lấy balances từ ExpenseService
	// TODO: Gọi algorithm.MinimizeCashFlow(balances)
	// TODO: INSERT từng settlement vào DB
	return nil, nil
}

// GeneratePaymentQR sinh QR code hoặc payment link cho một settlement.
// method: "PAYOS_QR" hoặc "MOMO"
// Luồng xử lý:
//  1. Lấy thông tin settlement (from_user, to_user, amount)
//  2. Gọi PayOS API hoặc MoMo API để tạo payment link
//  3. Lưu qr_code_data vào settlement trong DB
//  4. Trả về dữ liệu QR (URL hoặc payload)
func (s *SettlementService) GeneratePaymentQR(settlementID string, method string) (string, error) {
	// TODO: Lấy settlement từ DB
	// TODO: if method == "PAYOS_QR": gọi PayOS API
	// TODO: if method == "MOMO": gọi MoMo sandbox API
	// TODO: UPDATE settlements SET qr_code_data = ?, payment_method = ?
	return "", nil
}

// ConfirmPayment xác nhận một settlement đã được thanh toán.
// Được gọi khi webhook PayOS/MoMo báo thanh toán thành công,
// hoặc khi SePay phát hiện giao dịch chuyển khoản khớp với settlement.
// Luồng xử lý:
//  1. Cập nhật settlement.status = PAID
//  2. Gán confirmed_transaction_id (nếu có)
//  3. Lưu paid_at = now()
func (s *SettlementService) ConfirmPayment(settlementID string, transactionID string) error {
	// TODO: UPDATE settlements SET status = 'PAID', confirmed_transaction_id = ?, paid_at = now()
	return nil
}