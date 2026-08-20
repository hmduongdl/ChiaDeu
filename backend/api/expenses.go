package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/hmduongdl/ChiaDeu/pkg/expenses"
	"github.com/hmduongdl/ChiaDeu/pkg/vercel"
	"github.com/hmduongdl/ChiaDeu/models"
	"github.com/hmduongdl/ChiaDeu/services"
)

// Expenses xử lý tất cả /api/groups/:groupId/expenses/* — phân nhánh theo query param "sub"
// sub = create | update
func Expenses(w http.ResponseWriter, r *http.Request) {
	if vercel.HandleCORS(w, r) {
		return
	}

	vercel.WithAuth(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("sub") {
		case "create":
			handleCreateExpense(w, r)
		case "update":
			handleUpdateExpense(w, r)
		default:
			vercel.SendError(w, http.StatusNotFound, "route not found")
		}
	})(w, r)
}

// --- shared types ---

type expenseSplitRequest struct {
	UserID     string `json:"userId"`
	ShareMinor int64  `json:"shareMinor"`
}

type expenseBodyRequest struct {
	PaidBy      string                `json:"paidBy"`
	Description string                `json:"description"`
	AmountMinor int64                 `json:"amountMinor"`
	SplitType   string                `json:"splitType"`
	ExpenseDate string                `json:"expenseDate"`
	Splits      []expenseSplitRequest `json:"splits"`
}

func (req expenseBodyRequest) toInput() (expenses.Input, error) {
	input := expenses.Input{
		PaidBy:      req.PaidBy,
		Description: req.Description,
		AmountMinor: req.AmountMinor,
		SplitType:   req.SplitType,
	}
	if input.SplitType == "" {
		input.SplitType = models.SplitTypeEqual
	}
	val := strings.TrimSpace(req.ExpenseDate)
	if val == "" {
		input.ExpenseDate = time.Now()
	} else {
		parsed, err := time.Parse("2006-01-02", val)
		if err != nil {
			return expenses.Input{}, errors.New("expenseDate phải theo định dạng YYYY-MM-DD")
		}
		input.ExpenseDate = parsed
	}
	for _, s := range req.Splits {
		input.Splits = append(input.Splits, models.ExpenseSplit{UserID: s.UserID, ShareMinor: s.ShareMinor})
	}
	return input, nil
}

func mapExpenseErr(w http.ResponseWriter, err error) {
	var ve *services.ValidationError
	switch {
	case errors.As(err, &ve):
		vercel.SendError(w, http.StatusBadRequest, ve.Message)
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

// --- create expense ---

// handleCreateExpense xử lý POST /api/groups/:groupId/expenses
func handleCreateExpense(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		vercel.SendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userID, ok := vercel.GetUserID(r)
	if !ok {
		vercel.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	groupID := r.URL.Query().Get("groupId")
	if groupID == "" {
		vercel.SendError(w, http.StatusBadRequest, "mã nhóm là bắt buộc")
		return
	}

	var req expenseBodyRequest
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
	expense, splits, err := expensesService.CreateExpense(ctx, userID, groupID, input)
	if err != nil {
		mapExpenseErr(w, err)
		return
	}

	vercel.SendJSON(w, http.StatusCreated, map[string]interface{}{
		"expense": expense,
		"splits":  splits,
	})
}

// --- update expense ---

// handleUpdateExpense xử lý PATCH /api/groups/:groupId/expenses/:expenseId
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

	var req expenseBodyRequest
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
		mapExpenseErr(w, err)
		return
	}

	vercel.SendJSON(w, http.StatusOK, map[string]interface{}{
		"expense": expense,
		"splits":  splits,
	})
}
