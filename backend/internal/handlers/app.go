// Package handlers — HTTP handlers cho nghiệp vụ nhóm, khoản chi và quyết toán.
package handlers

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/hmduongdl/ChiaDeu/internal/expenses"
	"github.com/hmduongdl/ChiaDeu/internal/groups"
	authmiddleware "github.com/hmduongdl/ChiaDeu/internal/middleware"
	"github.com/hmduongdl/ChiaDeu/internal/settlements"
	"github.com/hmduongdl/ChiaDeu/models"
	"github.com/hmduongdl/ChiaDeu/services"
)

// AppHandler tập hợp các handler nghiệp vụ chính.
type AppHandler struct {
	groups      *groups.Service
	expenses    *expenses.Service
	settlements settlements.Store
}

func NewAppHandler(groupsService *groups.Service, expensesService *expenses.Service, settlementsStore settlements.Store) *AppHandler {
	return &AppHandler{groups: groupsService, expenses: expensesService, settlements: settlementsStore}
}

// ----------------------------------------------------------------------------
// Nhóm
// ----------------------------------------------------------------------------

type createGroupRequest struct {
	Name     string `json:"name"`
	Currency string `json:"currency"`
}

func (h *AppHandler) CreateGroup(c *fiber.Ctx) error {
	userID, ok := authmiddleware.UserID(c)
	if !ok {
		return unauthorized(c)
	}
	var request createGroupRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	group, err := h.groups.CreateGroup(c.UserContext(), userID, request.Name, request.Currency)
	if err != nil {
		switch {
		case errors.Is(err, groups.ErrInvalidGroupName), errors.Is(err, groups.ErrInvalidCurrency):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		case errors.Is(err, groups.ErrUserNotFound):
			return unauthorized(c)
		default:
			return internalError(c)
		}
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"group": group})
}

func (h *AppHandler) JoinGroup(c *fiber.Ctx) error {
	userID, ok := authmiddleware.UserID(c)
	if !ok {
		return unauthorized(c)
	}
	shareCode := c.Params("shareCode")

	group, err := h.groups.JoinGroup(c.UserContext(), userID, shareCode)
	if err != nil {
		switch {
		case errors.Is(err, groups.ErrInvalidShareCode):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "mã chia sẻ không hợp lệ"})
		case errors.Is(err, groups.ErrAlreadyMember):
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "bạn đã là thành viên nhóm này"})
		case errors.Is(err, groups.ErrGroupArchived):
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
		case errors.Is(err, groups.ErrUserNotFound):
			return unauthorized(c)
		default:
			return internalError(c)
		}
	}
	return c.JSON(fiber.Map{"group": group})
}

func (h *AppHandler) GetGroup(c *fiber.Ctx) error {
	userID, ok := authmiddleware.UserID(c)
	if !ok {
		return unauthorized(c)
	}
	groupID := c.Params("groupId")

	detail, err := h.groups.GetGroup(c.UserContext(), userID, groupID)
	if err != nil {
		switch {
		case errors.Is(err, groups.ErrGroupNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "nhóm không tồn tại"})
		case errors.Is(err, groups.ErrNotMember):
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "bạn không phải là thành viên nhóm"})
		default:
			return internalError(c)
		}
	}
	return c.JSON(fiber.Map{"group": detail.Group, "members": detail.Members})
}

// ----------------------------------------------------------------------------
// Khoản chi
// ----------------------------------------------------------------------------

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

func (h *AppHandler) CreateExpense(c *fiber.Ctx) error {
	userID, ok := authmiddleware.UserID(c)
	if !ok {
		return unauthorized(c)
	}
	groupID := c.Params("groupId")

	var request expenseRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	input, err := request.toInput()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	expense, splits, err := h.expenses.CreateExpense(c.UserContext(), userID, groupID, input)
	if err != nil {
		return mapExpenseError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"expense": expense, "splits": splits})
}

func (h *AppHandler) UpdateExpense(c *fiber.Ctx) error {
	userID, ok := authmiddleware.UserID(c)
	if !ok {
		return unauthorized(c)
	}
	groupID := c.Params("groupId")
	expenseID := c.Params("expenseId")

	var request expenseRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	input, err := request.toInput()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	expense, splits, err := h.expenses.UpdateExpense(c.UserContext(), userID, groupID, expenseID, input)
	if err != nil {
		return mapExpenseError(c, err)
	}
	return c.JSON(fiber.Map{"expense": expense, "splits": splits})
}

func (h *AppHandler) Balances(c *fiber.Ctx) error {
	userID, ok := authmiddleware.UserID(c)
	if !ok {
		return unauthorized(c)
	}
	groupID := c.Params("groupId")

	balances, err := h.expenses.UnsettledBalances(c.UserContext(), groupID)
	if err != nil {
		return internalError(c)
	}
	detail, err := h.groups.GetGroup(c.UserContext(), userID, groupID)
	if err != nil {
		if errors.Is(err, groups.ErrNotMember) || errors.Is(err, groups.ErrGroupNotFound) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "không thể xem số dư nhóm này"})
		}
		return internalError(c)
	}

	names := make(map[string]string, len(detail.Members))
	for _, member := range detail.Members {
		names[member.ID] = member.DisplayName()
	}

	result := make([]fiber.Map, 0, len(balances))
	for memberID, balanceMinor := range balances {
		result = append(result, fiber.Map{
			"userId":       memberID,
			"name":         names[memberID],
			"balanceMinor": balanceMinor,
		})
	}
	return c.JSON(fiber.Map{"balances": result})
}

// ----------------------------------------------------------------------------
// Quyết toán
// ----------------------------------------------------------------------------

type closeBatchRequest struct {
	IdempotencyKey string `json:"idempotencyKey"`
}

func (h *AppHandler) CloseBatch(c *fiber.Ctx) error {
	userID, ok := authmiddleware.UserID(c)
	if !ok {
		return unauthorized(c)
	}
	groupID := c.Params("groupId")

	key := strings.TrimSpace(c.Get("Idempotency-Key"))
	if key == "" {
		var request closeBatchRequest
		if err := c.BodyParser(&request); err == nil {
			key = strings.TrimSpace(request.IdempotencyKey)
		}
	}
	if key == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "thiếu idempotency key"})
	}

	snapshot, err := h.settlements.CloseBatch(c.UserContext(), groupID, userID, key)
	if err != nil {
		switch {
		case errors.Is(err, settlements.ErrNotAdmin):
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "chỉ quản trị viên mới được chốt kỳ"})
		case errors.Is(err, settlements.ErrBatchAlreadyOpen):
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "nhóm đã có một kỳ đang mở"})
		case errors.Is(err, settlements.ErrNoOpenExpenses):
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": "không có khoản chi nào để chốt"})
		case errors.Is(err, settlements.ErrInvalidExpenseData):
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": "dữ liệu khoản chi không hợp lệ"})
		default:
			return internalError(c)
		}
	}
	return c.Status(fiber.StatusCreated).JSON(snapshot)
}

func (h *AppHandler) GetBatch(c *fiber.Ctx) error {
	userID, ok := authmiddleware.UserID(c)
	if !ok {
		return unauthorized(c)
	}
	groupID := c.Params("groupId")
	batchID := c.Params("batchId")

	snapshot, err := h.settlements.GetBatch(c.UserContext(), groupID, batchID, userID)
	if err != nil {
		switch {
		case errors.Is(err, settlements.ErrBatchNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "kỳ quyết toán không tồn tại"})
		case errors.Is(err, settlements.ErrNotGroupMember):
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "bạn không phải là thành viên nhóm"})
		default:
			return internalError(c)
		}
	}
	return c.JSON(snapshot)
}

func (h *AppHandler) MarkSent(c *fiber.Ctx) error {
	return h.settleTransition(c, h.settlements.MarkSent)
}

func (h *AppHandler) Confirm(c *fiber.Ctx) error {
	return h.settleTransition(c, h.settlements.Confirm)
}

func (h *AppHandler) Reject(c *fiber.Ctx) error {
	return h.settleTransition(c, h.settlements.Reject)
}

type settleFunc func(ctx context.Context, settlementID, actorID string) (models.Settlement, error)

func (h *AppHandler) settleTransition(c *fiber.Ctx, action settleFunc) error {
	userID, ok := authmiddleware.UserID(c)
	if !ok {
		return unauthorized(c)
	}
	settlementID := c.Params("settlementId")

	settlement, err := action(c.UserContext(), settlementID, userID)
	if err != nil {
		switch {
		case errors.Is(err, settlements.ErrSettlementNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "giao dịch hoàn tiền không tồn tại"})
		case errors.Is(err, settlements.ErrNotGroupMember),
			errors.Is(err, settlements.ErrNotPayer),
			errors.Is(err, settlements.ErrNotRecipient):
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		case errors.Is(err, settlements.ErrBatchClosed), errors.Is(err, settlements.ErrInvalidTransition):
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
		default:
			return internalError(c)
		}
	}
	return c.JSON(fiber.Map{"settlement": settlement})
}

// ----------------------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------------------

func unauthorized(c *fiber.Ctx) error {
	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
}

func internalError(c *fiber.Ctx) error {
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
}

func mapExpenseError(c *fiber.Ctx, err error) error {
	var validationError *services.ValidationError
	switch {
	case errors.As(err, &validationError):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": validationError.Message})
	case errors.Is(err, expenses.ErrNotActiveMember):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "người ứng hoặc người chia không phải thành viên hoạt động"})
	case errors.Is(err, expenses.ErrNotOwner):
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "chỉ người tạo khoản chi mới được sửa"})
	case errors.Is(err, expenses.ErrExpenseLocked), errors.Is(err, expenses.ErrExpenseVoided):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "khoản chi không thể sửa"})
	case errors.Is(err, expenses.ErrExpenseNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "khoản chi không tồn tại"})
	default:
		return internalError(c)
	}
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
