package services

import (
	"container/heap"
	"errors"
	"fmt"

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

// balanceMaxHeap triển khai heap.Interface làm max-heap cho userBalance.
type balanceMaxHeap []userBalance

func (h balanceMaxHeap) Len() int { return len(h) }
func (h balanceMaxHeap) Less(i, j int) bool {
	if h[i].amount != h[j].amount {
		return h[i].amount > h[j].amount // Max-heap: phần tử có số tiền lớn hơn sẽ ưu tiên trước
	}
	return h[i].userID < h[j].userID // Khóa phụ userID để đảm bảo kết quả luôn xác định
}
func (h balanceMaxHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h *balanceMaxHeap) Push(x any)   { *h = append(*h, x.(userBalance)) }
func (h *balanceMaxHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

// SimplifyDebts rút gọn danh sách người nợ và người cần nhận thành các giao dịch
// hoàn tiền trực tiếp bằng 2 max-heap:
//
//  1. Tính số dư (balance) của mỗi người: tổng đã trả - tổng phải trả (đã tính trước khi truyền vào map).
//  2. Đưa tất cả người có balance dương vào max-heap (chủ nợ), người có balance âm vào max-heap khác (con nợ, lấy trị tuyệt đối).
//  3. Lặp lại: lấy phần tử lớn nhất ở mỗi heap (người nợ nhiều nhất, người được nợ nhiều nhất)
//     → tạo giao dịch giữa 2 người này với số tiền = min(hai giá trị)
//     → cập nhật lại balance
//     → đẩy lại vào heap nếu còn dư.
//  4. Dừng khi hết phần tử.
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

	creditors := &balanceMaxHeap{}
	debtors := &balanceMaxHeap{}

	for userID, balance := range balances {
		switch {
		case balance > 0:
			*creditors = append(*creditors, userBalance{userID: userID, amount: balance})
		case balance < 0:
			// Lấy trị tuyệt đối của số dư âm để đưa vào max-heap con nợ.
			*debtors = append(*debtors, userBalance{userID: userID, amount: -balance})
		}
	}

	heap.Init(creditors)
	heap.Init(debtors)

	var settlements []models.Settlement
	for creditors.Len() > 0 && debtors.Len() > 0 {
		// Lấy người được nợ nhiều nhất và người nợ nhiều nhất từ 2 max-heap
		creditor := heap.Pop(creditors).(userBalance)
		debtor := heap.Pop(debtors).(userBalance)

		amount := min(creditor.amount, debtor.amount)
		if amount <= 0 {
			break
		}

		settlements = append(settlements, models.Settlement{
			FromUserID:  debtor.userID,
			ToUserID:    creditor.userID,
			AmountMinor: amount,
			Status:      models.SettlementStatusPending,
		})

		creditor.amount -= amount
		debtor.amount -= amount

		// Cập nhật lại balance và đẩy lại vào heap nếu còn dư
		if creditor.amount > 0 {
			heap.Push(creditors, creditor)
		}
		if debtor.amount > 0 {
			heap.Push(debtors, debtor)
		}
	}

	return settlements, nil
}
