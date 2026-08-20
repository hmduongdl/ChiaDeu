// Package expenses — nghiệp vụ tạo/sửa/đọc khoản chi cho nhóm.
package expenses

import (
	"context"
	"time"

	"github.com/hmduongdl/ChiaDeu/models"
	"github.com/hmduongdl/ChiaDeu/services"
)

// Input là dữ liệu người dùng gửi lên để tạo hoặc sửa một khoản chi.
type Input struct {
	PaidBy      string
	Description string
	AmountMinor int64
	SplitType   string
	ExpenseDate time.Time
	Splits      []models.ExpenseSplit
}

func (i Input) toExpense(groupID, createdBy string) models.Expense {
	return models.Expense{
		GroupID:     groupID,
		CreatedBy:   createdBy,
		PaidBy:      i.PaidBy,
		Description: i.Description,
		AmountMinor: i.AmountMinor,
		SplitType:   i.SplitType,
		ExpenseDate: i.ExpenseDate,
		Status:      models.ExpenseStatusActive,
	}
}

// Service triển khai nghiệp vụ khoản chi trên nền Store.
type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

func (s *Service) CreateExpense(ctx context.Context, actorID, groupID string, input Input) (models.Expense, []models.ExpenseSplit, error) {
	expense := input.toExpense(groupID, actorID)
	if err := services.ValidateExpense(expense, input.Splits); err != nil {
		return models.Expense{}, nil, err
	}
	created, err := s.store.CreateExpenseWithSplits(ctx, expense, input.Splits)
	if err != nil {
		return models.Expense{}, nil, err
	}
	return created, input.Splits, nil
}

func (s *Service) UpdateExpense(ctx context.Context, actorID, groupID, expenseID string, input Input) (models.Expense, []models.ExpenseSplit, error) {
	expense := input.toExpense(groupID, actorID)
	expense.ID = expenseID
	if err := services.ValidateExpense(expense, input.Splits); err != nil {
		return models.Expense{}, nil, err
	}
	updated, err := s.store.UpdateExpenseWithSplits(ctx, expense, input.Splits)
	if err != nil {
		return models.Expense{}, nil, err
	}
	return updated, input.Splits, nil
}

// UnsettledBalances tính số dư ròng từ các khoản chi chưa chốt của nhóm.
func (s *Service) UnsettledBalances(ctx context.Context, groupID string) (map[string]int64, error) {
	expenses, splits, err := s.store.ListUnsettledExpensesWithSplits(ctx, groupID)
	if err != nil {
		return nil, err
	}
	return services.CalculateNetBalances(expenses, splits)
}
