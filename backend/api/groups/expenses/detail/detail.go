package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/hmduongdl/ChiaDeu/internal/expenses"
	"github.com/hmduongdl/ChiaDeu/internal/vercel"
	"github.com/hmduongdl/ChiaDeu/models"
	"github.com/hmduongdl/ChiaDeu/services"
)

type splitRequest struct {
	UserID     string `json:"userId"`
	ShareMinor int64  `json:"shareMinor"`
}

type expenseRequest struct {
	PaidBy      string         `json:"paidBy"`
	Description string         `json:"description"`
	AmountMinor int64          `json:"amountMinor"`
	SplitType   string         `json:"splitType"`
	ExpenseDate string         `json:"expenseDate"`
	Splits      []splitRequest `json:"splits"`
}

func (r expenseRequest) toInput() (expenses.Input, error) {
	input := expenses.Input{
		PaidBy:      r.PaidBy,
		Description: r.Description,
		AmountMinor: r.AmountMinor,
		SplitType:   r.SplitType,
	}
	if input.SplitType == "" {
		input.SplitType = models.SplitTypeEqual
	}
	expenseDate, err := parseExpenseDate(r.ExpenseDate)
	if err != nil {
		return expenses.Input{}, err
	}
	input.ExpenseDate = expenseDate
	for _, split := range r.Splits {
		input.Splits = append(input.Splits, models.ExpenseSplit{UserID: split.UserID, ShareMinor: split.ShareMinor})
	}
	return input, nil
}

func parseExpenseDate(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Now(), nil
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}, errors.New("expenseDate phải theo định dạng YYYY-MM-DD")
	}
	return parsed, nil
}

func mapExpenseError(w http.ResponseWriter, err error) {
	var validationError *services.ValidationError
	switch {
	case errors.As(err, &validationError):
		vercel.SendError(w, http.StatusBadRequest, validationError.Message)
	case errors.Is(err, expenses.ErrNotActiveMember):
		vercel.SendError(w, http.StatusBadRequest, "người ứng hoặc người chia không phải thành viên hoạt động")
	case errors.Is(err, expenses.ErrNotOwner):
		vercel.SendError(w, http.StatusForbidden, "chỉ người tạo khoản chi mới được sửa")
	case errors.Is(err, expenses.ErrExpenseLocked), errors.Is(err, expenses.ErrExpenseVoided):
		vercel.SendError(w, http.StatusConflict, "khoản chi không thể sửa")
	case errors.Is(err, expenses.ErrExpenseNotFound):
		vercel.SendError(w, http.StatusNotFound, "khoản chi không tồn tại")
	default:
		vercel.SendError(w, http.StatusInternalServerError, "internal server error")
	}
}

// Handler xử lý PATCH /api/groups/:groupId/expenses/:expenseId
func Handler(w http.ResponseWriter, r *http.Request) {
	vercel.WithAuth(handleUpdateExpense)(w, r)
}

func handleUpdateExpense(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch && r.Method != http.MethodPut {
		vercel.SendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userID, ok := vercel.GetUserID(r)
	if !ok {
		vercel.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	groupID := r.URL.Query().Get("groupId")
	expenseID := r.URL.Query().Get("expenseId")
	if groupID == "" || expenseID == "" {
		vercel.SendError(w, http.StatusBadRequest, "mã nhóm và mã khoản chi là bắt buộc")
		return
	}

	var req expenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		vercel.SendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	input, err := req.toInput()
	if err != nil {
		vercel.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	ctx := r.Context()
	pool, err := vercel.GetDB(ctx)
	if err != nil {
		vercel.SendError(w, http.StatusInternalServerError, "database connection error")
		return
	}

	expensesService := vercel.GetExpensesService(pool)
	expense, splits, err := expensesService.UpdateExpense(ctx, userID, groupID, expenseID, input)
	if err != nil {
		mapExpenseError(w, err)
		return
	}

	vercel.SendJSON(w, http.StatusOK, map[string]interface{}{
		"expense": expense,
		"splits":  splits,
	})
}
