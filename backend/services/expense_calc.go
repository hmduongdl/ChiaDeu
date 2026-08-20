// Package services chứa các phép tính thuần túy của nghiệp vụ chia tiền.
// Không phụ thuộc database hay HTTP; mọi comment bằng tiếng Việt.
package services

import (
	"errors"

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

// SplitEqual chia đều amountMinor cho memberIDs theo thứ tự truyền vào. Phần dư
// được phân bổ ổn định: mỗi thành viên đầu danh sách nhận thêm 1 đơn vị cho tới
// khi hết phần dư, nên cùng input luôn cho cùng output và tổng luôn bằng amount.
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
	for index, userID := range memberIDs {
		share := base
		if index < remainder {
			share++
		}
		splits[index] = models.ExpenseSplit{UserID: userID, ShareMinor: share}
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
