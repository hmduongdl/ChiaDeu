package services

import (
	"errors"
	"fmt"
	"sort"

	"github.com/hmduongdl/ChiaDeu/models"
)

// ErrUnbalanced báo tổng số dư khác 0, không thể chốt thành giao dịch.
var ErrUnbalanced = errors.New("tổng số dư phải bằng 0 để tạo settlement")

// CalculateNetBalances tính số dư ròng của từng thành viên trong tập khoản chi:
//
//	net_balance = tổng tiền đã ứng - tổng phần chi phí phải chịu
//
// Khoản chi VOIDED được bỏ qua. Trả về lỗi nếu bất kỳ khoản chi nào có tổng phần
// chia không khớp số tiền, bảo vệ invariant tài chính.
func CalculateNetBalances(expenses []models.Expense, splits []models.ExpenseSplit) (map[string]int64, error) {
	splitsByExpense := make(map[string][]models.ExpenseSplit)
	for _, split := range splits {
		splitsByExpense[split.ExpenseID] = append(splitsByExpense[split.ExpenseID], split)
	}

	balances := make(map[string]int64)
	for _, expense := range expenses {
		if expense.Status == models.ExpenseStatusVoided {
			continue
		}
		expenseSplits := splitsByExpense[expense.ID]
		if SumSplits(expenseSplits) != expense.AmountMinor {
			return nil, fmt.Errorf("khoản chi %s có tổng phần chia không khớp số tiền", expense.ID)
		}
		balances[expense.PaidBy] += expense.AmountMinor
		for _, split := range expenseSplits {
			balances[split.UserID] -= split.ShareMinor
		}
	}
	return balances, nil
}

type userBalance struct {
	userID string
	amount int64
}

// SimplifyDebts rút gọn danh sách người nợ và người cần nhận thành các giao dịch
// hoàn tiền trực tiếp, theo thứ tự quy định trong README:
//
//  1. Số dư dương là cần nhận, số dư âm là cần trả, bỏ số dư bằng 0.
//  2. Sắp xếp cả hai phía theo số dư tuyệt đối giảm dần, dùng userID làm khóa phụ.
//  3. Ghép người nợ lớn nhất với người được nhận lớn nhất, chốt bằng giá trị nhỏ
//     hơn, lặp tới khi cả hai phía cân bằng.
//
// Kết quả xác định với cùng input. Tổng input phải bằng 0.
func SimplifyDebts(balances map[string]int64) ([]models.Settlement, error) {
	var total int64
	for _, balance := range balances {
		total += balance
	}
	if total != 0 {
		return nil, ErrUnbalanced
	}

	var debtors, creditors []userBalance
	for userID, balance := range balances {
		switch {
		case balance > 0:
			creditors = append(creditors, userBalance{userID: userID, amount: balance})
		case balance < 0:
			// Lưu trị tuyệt đối để sắp xếp chung một chiều.
			debtors = append(debtors, userBalance{userID: userID, amount: -balance})
		}
	}

	sortByAmountDesc := func(users []userBalance) {
		sort.Slice(users, func(i, j int) bool {
			if users[i].amount != users[j].amount {
				return users[i].amount > users[j].amount
			}
			return users[i].userID < users[j].userID
		})
	}
	sortByAmountDesc(debtors)
	sortByAmountDesc(creditors)

	var settlements []models.Settlement
	debtorIndex, creditorIndex := 0, 0
	for debtorIndex < len(debtors) && creditorIndex < len(creditors) {
		amount := min(debtors[debtorIndex].amount, creditors[creditorIndex].amount)
		if amount <= 0 {
			break
		}
		settlements = append(settlements, models.Settlement{
			FromUserID:  debtors[debtorIndex].userID,
			ToUserID:    creditors[creditorIndex].userID,
			AmountMinor: amount,
			Status:      models.SettlementStatusPending,
		})
		debtors[debtorIndex].amount -= amount
		creditors[creditorIndex].amount -= amount
		if debtors[debtorIndex].amount == 0 {
			debtorIndex++
		}
		if creditors[creditorIndex].amount == 0 {
			creditorIndex++
		}
	}
	return settlements, nil
}
