// Package services — unit test cho chia tiền.
// File này kiểm thử theo yêu cầu tối thiểu trong phancong.md:
//   - 100 / 3 tạo ba phần có tổng đúng 100
//   - Không có member trả về lỗi
//   - Amount bằng 0 hoặc âm trả về lỗi
//   - Kết quả cùng input luôn giống nhau
package services

import (
	"errors"
	"reflect"
	"testing"

	"github.com/hmduongdl/ChiaDeu/models"
)

func TestSplitEqualDistributesRemainderStably(t *testing.T) {
	memberIDs := []string{"alice", "bob", "carol"}
	first, err := SplitEqual(100, memberIDs)
	if err != nil {
		t.Fatalf("split equal: %v", err)
	}
	second, err := SplitEqual(100, memberIDs)
	if err != nil {
		t.Fatalf("split equal second call: %v", err)
	}

	if !reflect.DeepEqual(first, second) {
		t.Fatalf("kết quả không xác định: %+v vs %+v", first, second)
	}
	if total := SumSplits(first); total != 100 {
		t.Fatalf("tổng phần chia = %d, mong đợi 100", total)
	}
	// Phần dư 1 đơn vị rơi vào thành viên đầu danh sách theo thứ tự ổn định.
	expected := map[string]int64{"alice": 34, "bob": 33, "carol": 33}
	for _, split := range first {
		if split.ShareMinor != expected[split.UserID] {
			t.Fatalf("user %s được chia %d, mong đợi %d", split.UserID, split.ShareMinor, expected[split.UserID])
		}
	}
}

func TestSplitEqualEmptyMembersErrors(t *testing.T) {
	if _, err := SplitEqual(100, nil); !errors.Is(err, ErrNoMembers) {
		t.Fatalf("mong đợi ErrNoMembers, got %v", err)
	}
}

func TestSplitEqualRejectsNonPositiveAmount(t *testing.T) {
	for _, amount := range []int64{0, -100} {
		if _, err := SplitEqual(amount, []string{"alice"}); !errors.Is(err, ErrInvalidAmount) {
			t.Fatalf("amount %d: mong đợi ErrInvalidAmount, got %v", amount, err)
		}
	}
}

func TestSplitEqualSingleMemberTakesEverything(t *testing.T) {
	splits, err := SplitEqual(5000, []string{"alice"})
	if err != nil {
		t.Fatalf("split equal: %v", err)
	}
	if len(splits) != 1 || splits[0].UserID != "alice" || splits[0].ShareMinor != 5000 {
		t.Fatalf("unexpected splits: %+v", splits)
	}
}

func TestValidateExpenseChecksSplitsSum(t *testing.T) {
	expense := models.Expense{
		GroupID:     "group-1",
		PaidBy:      "alice",
		Description: "Tiền ăn",
		AmountMinor: 100000,
		SplitType:   models.SplitTypeEqual,
		Status:      models.ExpenseStatusActive,
	}
	splits := []models.ExpenseSplit{
		{UserID: "alice", ShareMinor: 40000},
		{UserID: "bob", ShareMinor: 40000},
	}
	if err := ValidateExpense(expense, splits); !errors.Is(err, ErrSplitsMismatch) {
		t.Fatalf("mong đợi ErrSplitsMismatch, got %v", err)
	}
}

func TestValidateExpenseRejectsBadSplitTypeAndNegativeShare(t *testing.T) {
	expense := models.Expense{
		PaidBy:      "alice",
		AmountMinor: 100,
		SplitType:   "UNKNOWN",
		Status:      models.ExpenseStatusActive,
	}
	if err := ValidateExpense(expense, []models.ExpenseSplit{{UserID: "a", ShareMinor: 100}}); err == nil {
		t.Fatal("mong đợi lỗi kiểu chia không hợp lệ")
	}

	expense.SplitType = models.SplitTypeCustom
	if err := ValidateExpense(expense, []models.ExpenseSplit{{UserID: "a", ShareMinor: -5}}); err == nil {
		t.Fatal("mong đợi lỗi phần chia âm")
	}
}
