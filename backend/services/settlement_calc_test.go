// Package services — unit test cho resolver công nợ.
// File này kiểm thử theo bảng minh họa trong phancong.md:
//
//	| Input balance        | Output mong đợi               |
//	| A:+200, B:-120, C:-80| B→A:120, C→A:80               |
//	| A:+70, B:+30, C:-100 | Hai giao dịch, tổng 100        |
//	| A:0, B:0             | Không có settlement            |
//	| A:+10, B:-9          | Trả lỗi vì tổng khác 0         |
package services

import (
	"errors"
	"reflect"
	"testing"

	"github.com/hmduongdl/ChiaDeu/models"
)

func TestCalculateNetBalances(t *testing.T) {
	expenses := []models.Expense{
		{ID: "e1", PaidBy: "an", AmountMinor: 300000, Status: models.ExpenseStatusActive},
		{ID: "e2", PaidBy: "chau", AmountMinor: 60000, Status: models.ExpenseStatusActive},
		{ID: "e3", PaidBy: "void", AmountMinor: 999999, Status: models.ExpenseStatusVoided},
	}
	splits := []models.ExpenseSplit{
		{ExpenseID: "e1", UserID: "an", ShareMinor: 100000},
		{ExpenseID: "e1", UserID: "binh", ShareMinor: 120000},
		{ExpenseID: "e1", UserID: "chau", ShareMinor: 80000},
		{ExpenseID: "e2", UserID: "chau", ShareMinor: 60000},
	}

	balances, err := CalculateNetBalances(expenses, splits)
	if err != nil {
		t.Fatalf("calculate net balances: %v", err)
	}
	expected := map[string]int64{"an": 200000, "binh": -120000, "chau": -80000}
	if !reflect.DeepEqual(balances, expected) {
		t.Fatalf("balances = %+v, mong đợi %+v", balances, expected)
	}
}

func TestCalculateNetBalancesRejectsBrokenExpense(t *testing.T) {
	expenses := []models.Expense{{ID: "e1", PaidBy: "an", AmountMinor: 100, Status: models.ExpenseStatusActive}}
	splits := []models.ExpenseSplit{{ExpenseID: "e1", UserID: "an", ShareMinor: 50}}
	if _, err := CalculateNetBalances(expenses, splits); err == nil {
		t.Fatal("mong đợi lỗi tổng phần chia không khớp")
	}
}

func TestSimplifyDebtsExampleBalances(t *testing.T) {
	settlements, err := SimplifyDebts(map[string]int64{"an": 200000, "binh": -120000, "chau": -80000})
	if err != nil {
		t.Fatalf("simplify debts: %v", err)
	}
	assertSettlements(t, settlements, []models.Settlement{
		{FromUserID: "binh", ToUserID: "an", AmountMinor: 120000, Status: models.SettlementStatusPending},
		{FromUserID: "chau", ToUserID: "an", AmountMinor: 80000, Status: models.SettlementStatusPending},
	})
}

func TestSimplifyDebtsTwoDebtorsToOneCreditor(t *testing.T) {
	settlements, err := SimplifyDebts(map[string]int64{"a": 70, "b": 30, "c": -100})
	if err != nil {
		t.Fatalf("simplify debts: %v", err)
	}
	var total int64
	recipients := make(map[string]struct{})
	for _, s := range settlements {
		total += s.AmountMinor
		if s.FromUserID != "c" {
			t.Fatalf("người nợ phải là c, got %+v", s)
		}
		recipients[s.ToUserID] = struct{}{}
	}
	if len(settlements) != 2 || total != 100 {
		t.Fatalf("mong đợi hai giao dịch tổng 100, got %+v", settlements)
	}
	if _, okA := recipients["a"]; !okA {
		t.Fatalf("mong đợi có người nhận a, got %+v", settlements)
	}
	if _, okB := recipients["b"]; !okB {
		t.Fatalf("mong đợi có người nhận b, got %+v", settlements)
	}
}

func TestSimplifyDebtsAllZeroProducesNothing(t *testing.T) {
	settlements, err := SimplifyDebts(map[string]int64{"a": 0, "b": 0})
	if err != nil {
		t.Fatalf("simplify debts: %v", err)
	}
	if len(settlements) != 0 {
		t.Fatalf("mong đợi không có settlement, got %+v", settlements)
	}
}

func TestSimplifyDebtsRejectsUnbalancedInput(t *testing.T) {
	if _, err := SimplifyDebts(map[string]int64{"a": 10, "b": -9}); !errors.Is(err, ErrUnbalanced) {
		t.Fatalf("mong đợi ErrUnbalanced, got %v", err)
	}
}

func assertSettlements(t *testing.T, got, want []models.Settlement) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("settlements = %+v, mong đợi %+v", got, want)
	}
}
