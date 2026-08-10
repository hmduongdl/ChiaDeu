// Package services chứa logic nghiệp vụ (business logic) của ứng dụng.
// Các service được tách riêng khỏi handler để dễ kiểm thử và tái sử dụng.
// Mỗi service nhận *sql.DB qua constructor và hoạt động độc lập.
package services

import (
	"database/sql"

	"github.com/yourusername/cash-flow-minimizer/internal/models"
)

// ExpenseService xử lý tất cả nghiệp vụ liên quan đến khoản chi:
// - Tạo expense mới kèm splits
// - Tính toán net balance cho từng thành viên trong nhóm
type ExpenseService struct {
	db *sql.DB
}

// NewExpenseService khởi tạo service với database connection.
func NewExpenseService(db *sql.DB) *ExpenseService {
	return &ExpenseService{db: db}
}

// CreateExpense tạo một khoản chi mới trong nhóm, bao gồm cả việc chia tiền.
// Sử dụng database transaction để đảm bảo tính nhất quán:
// - Nếu bất kỳ bước nào thất bại, toàn bộ thay đổi được rollback.
// Luồng xử lý:
//  1. INSERT vào bảng expenses
//  2. INSERT từng dòng vào bảng expense_splits
//  3. Nếu có source_transaction_id: UPDATE bank_transactions SET is_used = true
func (s *ExpenseService) CreateExpense(input models.CreateExpenseInput) (*models.Expense, error) {
	// TODO: Bắt đầu transaction: tx, err := s.db.Begin()
	// TODO: INSERT INTO expenses (...) VALUES (...) RETURNING id
	// TODO: for each split: INSERT INTO expense_splits (...)
	// TODO: if input.SourceTransactionID != nil: UPDATE bank_transactions SET is_used = true
	// TODO: Commit transaction
	return nil, nil
}

// GetGroupBalances tính toán net balance của từng thành viên trong nhóm.
// Công thức:
//
//	net = SUM(paid_by) - SUM(share_amount trong splits)
//
// net > 0: người này được nhận tiền (creditor)
// net < 0: người này phải trả tiền (debtor)
// net = 0: hoà vốn
// Kết quả được dùng làm đầu vào cho thuật toán Minimize Cash Flow.
func (s *ExpenseService) GetGroupBalances(groupID string) ([]models.Balance, error) {
	// TODO: SELECT user_id, SUM(paid_amount) - SUM(share_amount) as net
	//       FROM (tổng hợp từ expenses và expense_splits)
	//       WHERE group_id = ?
	//       GROUP BY user_id
	return nil, nil
}