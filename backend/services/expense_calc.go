// Package services chứa các phép tính thuần túy của nghiệp vụ chia tiền.
// Không phụ thuộc database hay HTTP; mọi comment bằng tiếng Việt.
package services

import (
	"errors"
	"sort"

	"github.com/hmduongdl/ChiaDeu/models"
)

var (
	// ErrSplitsMismatch báo tổng phần chia không khớp số tiền khoản chi.
	ErrSplitsMismatch = errors.New("tổng phần chia không khớp số tiền khoản chi")
	// ErrInvalidAmount báo số tiền phải lớn hơn 0.
	ErrInvalidAmount = errors.New("số tiền phải lớn hơn 0")
	// ErrNoMembers báo danh sách người chia không được để trống.
	ErrNoMembers = errors.New("cần ít nhất một thành viên tham gia chia")
)

// ValidationError mô tả lỗi đầu vào nghiệp vụ; vừa wrap sentinel gốc để errors.Is
// khớp, vừa mang message cho handler trả về cho client.
type ValidationError struct {
	Message string
	Cause   error
}

func (e *ValidationError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.Cause.Error()
}

func (e *ValidationError) Unwrap() error {
	return e.Cause
}

func validationError(message string, cause error) error {
	if cause == nil {
		cause = errors.New(message)
	}
	return &ValidationError{Message: message, Cause: cause}
}

// SumSplits trả về tổng phần chia của một danh sách.
func SumSplits(splits []models.ExpenseSplit) int64 {
	var total int64
	for _, split := range splits {
		total += split.ShareMinor
	}
	return total
}

// SplitEqual chia đều amountMinor cho memberIDs. Để đảm bảo tính xác định
// (deterministic), mảng kết quả sẽ được sắp xếp theo UserID trước khi phân 
// bổ phần dư. Những người có UserID nhỏ nhất sẽ được cộng thêm 1 đơn vị.
// Thuật toán tối ưu zero-allocation: sử dụng chính mảng kết quả để sort in-place.
func SplitEqual(amountMinor int64, memberIDs []string) ([]models.ExpenseSplit, error) {
	if amountMinor <= 0 {
		return nil, ErrInvalidAmount
	}
	if len(memberIDs) == 0 {
		return nil, ErrNoMembers
	}

	count := int64(len(memberIDs))
	base := amountMinor / count
	remainder := int(amountMinor % count)

	splits := make([]models.ExpenseSplit, len(memberIDs))
	for i, userID := range memberIDs {
		splits[i] = models.ExpenseSplit{UserID: userID, ShareMinor: base}
	}

	// Sắp xếp in-place trực tiếp trên mảng kết quả để không tốn thêm bộ nhớ O(N)
	sort.Slice(splits, func(i, j int) bool {
		return splits[i].UserID < splits[j].UserID
	})

	// Phân bổ phần dư cho 'remainder' người đầu tiên
	for i := 0; i < remainder; i++ {
		splits[i].ShareMinor++
	}

	return splits, nil
}

// ValidateExpense kiểm tra các invariant bắt buộc của một khoản chi cùng phần
// chia trước khi ghi vào database. Lỗi trả về là *ValidationError.
func ValidateExpense(expense models.Expense, splits []models.ExpenseSplit) error {
	if expense.AmountMinor <= 0 {
		return validationError("số tiền khoản chi phải lớn hơn 0", ErrInvalidAmount)
	}
	if expense.PaidBy == "" {
		return validationError("paid_by là bắt buộc", nil)
	}
	switch expense.SplitType {
	case models.SplitTypeEqual, models.SplitTypePercent, models.SplitTypeWeight, models.SplitTypeCustom:
	default:
		return validationError("kiểu chia không hợp lệ", nil)
	}
	if len(splits) == 0 {
		return validationError("cần ít nhất một phần chia", nil)
	}
	for _, split := range splits {
		if split.UserID == "" {
			return validationError("phần chia thiếu user_id", nil)
		}
		if split.ShareMinor < 0 {
			return validationError("phần chia không được âm", nil)
		}
	}
	if SumSplits(splits) != expense.AmountMinor {
		return validationError("tổng phần chia không khớp số tiền khoản chi", ErrSplitsMismatch)
	}
	return nil
}
