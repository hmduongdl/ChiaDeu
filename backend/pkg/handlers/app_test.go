// Package handlers — test tích hợp nghiệp vụ thông qua HTTP, dùng fake stores
// trong bộ nhớ để chạy toàn bộ luồng: tạo nhóm → tham gia → ghi khoản chi →
// xem số dư → chốt kỳ → báo chuyển → xác nhận đã nhận.
package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/hmduongdl/ChiaDeu/pkg/expenses"
	"github.com/hmduongdl/ChiaDeu/pkg/groups"
	"github.com/hmduongdl/ChiaDeu/pkg/settlements"
	"github.com/hmduongdl/ChiaDeu/models"
)

// ---------------------------------------------------------------------------
// Fake stores
// ---------------------------------------------------------------------------

type fakeGroupStore struct {
	group   models.Group
	users   map[string]bool
	members map[string]models.GroupMember
}

func newFakeGroupStore() *fakeGroupStore {
	store := &fakeGroupStore{
		group: models.Group{
			ID: "group-1", Name: "Nhóm ăn", ShareCode: "ABC12345", Currency: "VND",
			Status: models.GroupStatusActive, CreatedAt: time.Now(),
		},
		// alice/bob tồn tại trên hệ thống; membership được tạo theo luồng nghiệp vụ.
		users:   map[string]bool{"alice": true, "bob": true},
		members: map[string]models.GroupMember{},
	}
	return store
}

func (s *fakeGroupStore) addMember(groupID, userID, role string) {
	s.users[userID] = true
	s.members[groupID+"|"+userID] = models.GroupMember{
		GroupID: groupID, UserID: userID, Role: role,
		Status: models.MemberStatusActive, JoinedAt: time.Now(),
	}
}

func (s *fakeGroupStore) CreateGroupWithAdmin(_ context.Context, group models.Group, admin models.GroupMember) (models.Group, error) {
	if group.ID == "" {
		group.ID = "group-1" // giả lập RETURNING id của database
	}
	s.group = group
	s.members[group.ID+"|"+admin.UserID] = admin
	s.users[admin.UserID] = true
	return group, nil
}

func (s *fakeGroupStore) GetGroup(_ context.Context, groupID string) (models.Group, error) {
	if s.group.ID != groupID {
		return models.Group{}, groups.ErrGroupNotFound
	}
	return s.group, nil
}

func (s *fakeGroupStore) GetGroupByShareCode(_ context.Context, shareCode string) (models.Group, error) {
	if s.group.ShareCode != shareCode {
		return models.Group{}, groups.ErrGroupNotFound
	}
	return s.group, nil
}

func (s *fakeGroupStore) GetMembership(_ context.Context, groupID, userID string) (models.GroupMember, error) {
	member, ok := s.members[groupID+"|"+userID]
	if !ok {
		return models.GroupMember{}, groups.ErrMembershipMissing
	}
	return member, nil
}

func (s *fakeGroupStore) CreateMembership(_ context.Context, member models.GroupMember) error {
	key := member.GroupID + "|" + member.UserID
	if _, ok := s.members[key]; ok {
		return groups.ErrAlreadyMember
	}
	s.members[key] = member
	return nil
}

func (s *fakeGroupStore) ListActiveMembers(_ context.Context, groupID string) ([]models.User, error) {
	var members []models.User
	for _, member := range s.members {
		if member.GroupID == groupID && member.IsActiveMember() {
			members = append(members, models.User{ID: member.UserID, Name: member.UserID})
		}
	}
	return members, nil
}

func (s *fakeGroupStore) ListUserGroups(_ context.Context, userID string) ([]models.Group, error) {
	if _, ok := s.members[s.group.ID+"|"+userID]; ok {
		return []models.Group{s.group}, nil
	}
	return []models.Group{}, nil
}

func (s *fakeGroupStore) UserExists(_ context.Context, userID string) (bool, error) {
	return s.users[userID], nil
}

type fakeExpenseStore struct {
	expenses []models.Expense
	splits   []models.ExpenseSplit
	nextID   int
}

func newFakeExpenseStore() *fakeExpenseStore {
	return &fakeExpenseStore{nextID: 1}
}

func (s *fakeExpenseStore) CreateExpenseWithSplits(_ context.Context, expense models.Expense, splits []models.ExpenseSplit) (models.Expense, error) {
	expense.ID = "expense-" + string(rune('0'+s.nextID))
	s.nextID++
	expense.CreatedAt = time.Now()
	expense.Status = models.ExpenseStatusActive
	s.expenses = append(s.expenses, expense)
	for _, split := range splits {
		split.ID = "split-" + string(rune('0'+len(s.splits)))
		split.ExpenseID = expense.ID
		split.CreatedAt = time.Now()
		s.splits = append(s.splits, split)
	}
	return expense, nil
}

func (s *fakeExpenseStore) UpdateExpenseWithSplits(_ context.Context, expense models.Expense, splits []models.ExpenseSplit) (models.Expense, error) {
	for index := range s.expenses {
		if s.expenses[index].ID == expense.ID {
			s.expenses[index] = expense
			s.splits = nil
			for _, split := range splits {
				split.ExpenseID = expense.ID
				s.splits = append(s.splits, split)
			}
			return expense, nil
		}
	}
	return models.Expense{}, expenses.ErrExpenseNotFound
}

func (s *fakeExpenseStore) GetExpenseWithSplits(_ context.Context, groupID, expenseID string) (models.Expense, []models.ExpenseSplit, error) {
	for _, expense := range s.expenses {
		if expense.ID == expenseID && expense.GroupID == groupID {
			return expense, s.splitsFor(expenseID), nil
		}
	}
	return models.Expense{}, nil, expenses.ErrExpenseNotFound
}

func (s *fakeExpenseStore) ListUnsettledExpensesWithSplits(_ context.Context, groupID string) ([]models.Expense, []models.ExpenseSplit, error) {
	var result []models.Expense
	var resultSplits []models.ExpenseSplit
	for _, expense := range s.expenses {
		if expense.GroupID == groupID && expense.BatchID == nil {
			result = append(result, expense)
			resultSplits = append(resultSplits, s.splitsFor(expense.ID)...)
		}
	}
	return result, resultSplits, nil
}

func (s *fakeExpenseStore) splitsFor(expenseID string) []models.ExpenseSplit {
	var result []models.ExpenseSplit
	for _, split := range s.splits {
		if split.ExpenseID == expenseID {
			result = append(result, split)
		}
	}
	return result
}

func (s *fakeExpenseStore) ListGroupExpensesWithSplits(_ context.Context, groupID string) ([]models.Expense, []models.ExpenseSplit, error) {
	var result []models.Expense
	var resultSplits []models.ExpenseSplit
	for _, expense := range s.expenses {
		if expense.GroupID == groupID {
			result = append(result, expense)
			resultSplits = append(resultSplits, s.splitsFor(expense.ID)...)
		}
	}
	return result, resultSplits, nil
}

func (s *fakeExpenseStore) VoidExpense(_ context.Context, groupID, expenseID, actorID string) (models.Expense, error) {
	for index := range s.expenses {
		if s.expenses[index].ID == expenseID && s.expenses[index].GroupID == groupID {
			if s.expenses[index].CreatedBy != actorID {
				return models.Expense{}, expenses.ErrNotOwner
			}
			if s.expenses[index].Status != models.ExpenseStatusActive {
				return models.Expense{}, expenses.ErrExpenseVoided
			}
			if s.expenses[index].BatchID != nil {
				return models.Expense{}, expenses.ErrExpenseLocked
			}
			s.expenses[index].Status = models.ExpenseStatusVoided
			return s.expenses[index], nil
		}
	}
	return models.Expense{}, expenses.ErrExpenseNotFound
}

type fakeSettlementsStore struct {
	batch   settlements.BatchSnapshot
	byID    map[string]models.Settlement
	groupID string
}

func newFakeSettlementsStore() *fakeSettlementsStore {
	return &fakeSettlementsStore{
		batch: settlements.BatchSnapshot{
			Batch: models.SettlementBatch{ID: "batch-1", GroupID: "group-1", Status: models.BatchStatusOpen},
		},
		byID:    map[string]models.Settlement{},
		groupID: "group-1",
	}
}

func (f *fakeSettlementsStore) CloseBatch(_ context.Context, groupID, actorID, idempotencyKey string) (settlements.BatchSnapshot, error) {
	return f.batch, nil
}

func (f *fakeSettlementsStore) GetBatch(_ context.Context, groupID, batchID, actorID string) (settlements.BatchSnapshot, error) {
	return f.batch, nil
}

func (f *fakeSettlementsStore) GetSettlement(_ context.Context, settlementID string) (settlements.SettlementContext, error) {
	settlement, ok := f.byID[settlementID]
	if !ok {
		return settlements.SettlementContext{}, settlements.ErrSettlementNotFound
	}
	return settlements.SettlementContext{Settlement: settlement, GroupID: f.groupID, BatchStatus: models.BatchStatusOpen}, nil
}

func (f *fakeSettlementsStore) MarkSent(_ context.Context, settlementID, actorID string) (models.Settlement, error) {
	return f.transition(settlementID, actorID, func(settlement *models.Settlement) error {
		if settlement.FromUserID != actorID {
			return settlements.ErrNotPayer
		}
		if settlement.Status != models.SettlementStatusPending {
			return settlements.ErrInvalidTransition
		}
		settlement.Status = models.SettlementStatusAwaitingConfirmation
		now := time.Now()
		settlement.MarkedSentAt = &now
		return nil
	})
}

func (f *fakeSettlementsStore) Confirm(_ context.Context, settlementID, actorID string) (models.Settlement, error) {
	return f.transition(settlementID, actorID, func(settlement *models.Settlement) error {
		if settlement.ToUserID != actorID {
			return settlements.ErrNotRecipient
		}
		if settlement.Status != models.SettlementStatusAwaitingConfirmation {
			return settlements.ErrInvalidTransition
		}
		settlement.Status = models.SettlementStatusPaid
		now := time.Now()
		settlement.PaidAt = &now
		return nil
	})
}

func (f *fakeSettlementsStore) Reject(_ context.Context, settlementID, actorID string) (models.Settlement, error) {
	return f.transition(settlementID, actorID, func(settlement *models.Settlement) error {
		if settlement.ToUserID != actorID {
			return settlements.ErrNotRecipient
		}
		if settlement.Status != models.SettlementStatusAwaitingConfirmation {
			return settlements.ErrInvalidTransition
		}
		settlement.Status = models.SettlementStatusPending
		settlement.MarkedSentAt = nil
		return nil
	})
}

func (f *fakeSettlementsStore) CancelBatch(_ context.Context, batchID, actorID string) (settlements.BatchSnapshot, error) {
	return f.batch, nil
}

func (f *fakeSettlementsStore) transition(settlementID, actorID string, apply func(*models.Settlement) error) (models.Settlement, error) {
	settlement, ok := f.byID[settlementID]
	if !ok {
		return models.Settlement{}, settlements.ErrSettlementNotFound
	}
	if err := apply(&settlement); err != nil {
		return models.Settlement{}, err
	}
	f.byID[settlementID] = settlement
	return settlement, nil
}

// ---------------------------------------------------------------------------
// Test app
// ---------------------------------------------------------------------------

// newTestApp dựng fiber app với middleware mạo danh user qua header X-Test-User để
// mô phỏng AuthMiddleware mà không cần token thật.
func newTestApp(groupStore groups.Store, expenseStore expenses.Store, settlementStore settlements.Store) *fiber.App {
	app := fiber.New()
	api := app.Group("/api")
	appHandler := NewAppHandler(groups.NewService(groupStore), expenses.NewService(expenseStore), settlementStore)

	requireUser := func(c *fiber.Ctx) error {
		user := c.Get("X-Test-User")
		if user == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
		}
		c.Locals("authenticatedUserID", user)
		return c.Next()
	}

	routes := api.Group("/groups", requireUser)
	routes.Post("/", appHandler.CreateGroup)
	routes.Get("/", appHandler.ListGroups)
	routes.Post("/join/:shareCode", appHandler.JoinGroup)
	routes.Get("/:groupId", appHandler.GetGroup)
	routes.Get("/:groupId/expenses", appHandler.ListExpenses)
	routes.Post("/:groupId/expenses", appHandler.CreateExpense)
	routes.Get("/:groupId/expenses/:expenseId", appHandler.GetExpense)
	routes.Patch("/:groupId/expenses/:expenseId", appHandler.UpdateExpense)
	routes.Post("/:groupId/expenses/:expenseId/void", appHandler.VoidExpense)
	routes.Get("/:groupId/balances", appHandler.Balances)
	routes.Post("/:groupId/settlement-batches", appHandler.CloseBatch)
	routes.Get("/:groupId/settlement-batches/:batchId", appHandler.GetBatch)
	routes.Post("/:groupId/settlement-batches/:batchId/cancel", appHandler.CancelBatch)

	settlementRoutes := api.Group("/settlements", requireUser)
	settlementRoutes.Get("/:settlementId", appHandler.GetSettlement)
	settlementRoutes.Post("/:settlementId/mark-sent", appHandler.MarkSent)
	settlementRoutes.Post("/:settlementId/confirm", appHandler.Confirm)
	settlementRoutes.Post("/:settlementId/reject", appHandler.Reject)

	return app
}

func doRequest(t *testing.T, app *fiber.App, method, path, actor, body string) (*http.Response, map[string]any) {
	t.Helper()
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	if actor != "" {
		request.Header.Set("X-Test-User", actor)
	}
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("request %s %s: %v", method, path, err)
	}
	rawBody, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}
	response.Body.Close()

	var rawPayload map[string]any
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &rawPayload); err != nil {
			t.Fatalf("decode JSON %q: %v", rawBody, err)
		}
	}

	var payload map[string]any
	if rawPayload != nil {
		if successVal, hasSuccess := rawPayload["success"]; hasSuccess {
			if successVal == true {
				if dataVal, hasData := rawPayload["data"]; hasData {
					if dataMap, ok := dataVal.(map[string]any); ok {
						payload = dataMap
					} else {
						payload = rawPayload
					}
				} else {
					payload = rawPayload
				}
			} else {
				payload = rawPayload
			}
		} else {
			payload = rawPayload
		}
	}

	return response, payload
}

func TestGroupExpenseCloseAndPaymentFlow(t *testing.T) {
	settlementStore := newFakeSettlementsStore()
	settlementStore.byID["settlement-1"] = models.Settlement{
		ID: "settlement-1", BatchID: "batch-1", FromUserID: "bob", ToUserID: "alice",
		AmountMinor: 5000, PaymentCode: "ABC123", Status: models.SettlementStatusPending, CreatedAt: time.Now(),
	}
	app := newTestApp(newFakeGroupStore(), newFakeExpenseStore(), settlementStore)

	// Tạo nhóm
	response, payload := doRequest(t, app, http.MethodPost, "/api/groups", "alice", `{"name":"Nhóm ăn"}`)
	if response.StatusCode != http.StatusCreated {
		t.Fatalf("create group: status=%d body=%v", response.StatusCode, payload)
	}
	group := payload["group"].(map[string]any)
	groupID := group["id"].(string)
	shareCode := group["shareCode"].(string)
	if groupID != "group-1" || shareCode == "" {
		t.Fatalf("mong đợi group-1 và share code hợp lệ, got %v", group)
	}

	// Tham gia bằng mã chia sẻ
	response, _ = doRequest(t, app, http.MethodPost, "/api/groups/join/"+shareCode, "bob", "")
	if response.StatusCode != http.StatusOK {
		t.Fatalf("join group: status=%d", response.StatusCode)
	}

	// Ghi khoản chi chia đều
	expenseBody := `{"paidBy":"alice","description":"Tiền ăn","amountMinor":10000,"splitType":"EQUAL","splits":[{"userId":"alice","shareMinor":5000},{"userId":"bob","shareMinor":5000}]}`
	response, payload = doRequest(t, app, http.MethodPost, "/api/groups/"+groupID+"/expenses", "alice", expenseBody)
	if response.StatusCode != http.StatusCreated {
		t.Fatalf("create expense: status=%d body=%v", response.StatusCode, payload)
	}

	// Xem số dư — bob đang nợ 5000
	response, payload = doRequest(t, app, http.MethodGet, "/api/groups/"+groupID+"/balances", "bob", "")
	if response.StatusCode != http.StatusOK {
		t.Fatalf("balances: status=%d body=%v", response.StatusCode, payload)
	}
	balances := payload["balances"].([]any)
	found := false
	for _, entry := range balances {
		item := entry.(map[string]any)
		if item["userId"] == "bob" {
			found = true
			if item["balanceMinor"].(float64) != -5000 {
				t.Fatalf("mong đợi bob nợ 5000, got %v", item)
			}
		}
	}
	if !found {
		t.Fatalf("không tìm thấy số dư của bob: %v", balances)
	}

	// Chốt kỳ với idempotency key
	response, payload = doRequest(t, app, http.MethodPost, "/api/groups/"+groupID+"/settlement-batches", "alice", `{"idempotencyKey":"close-001"}`)
	if response.StatusCode != http.StatusCreated {
		t.Fatalf("close batch: status=%d body=%v", response.StatusCode, payload)
	}

	// Bob báo đã chuyển
	response, payload = doRequest(t, app, http.MethodPost, "/api/settlements/settlement-1/mark-sent", "bob", "")
	if response.StatusCode != http.StatusOK {
		t.Fatalf("mark-sent: status=%d body=%v", response.StatusCode, payload)
	}
	settlement := payload["settlement"].(map[string]any)
	if settlement["status"] != "AWAITING_CONFIRMATION" {
		t.Fatalf("status sau mark-sent phải là AWAITING_CONFIRMATION, got %v", settlement["status"])
	}

	// Alice xác nhận đã nhận
	response, payload = doRequest(t, app, http.MethodPost, "/api/settlements/settlement-1/confirm", "alice", "")
	if response.StatusCode != http.StatusOK {
		t.Fatalf("confirm: status=%d body=%v", response.StatusCode, payload)
	}
	settlement = payload["settlement"].(map[string]any)
	if settlement["status"] != "PAID" {
		t.Fatalf("status sau confirm phải là PAID, got %v", settlement["status"])
	}

	// Xác nhận lần nữa phải thất bại (đã kết thúc)
	response, _ = doRequest(t, app, http.MethodPost, "/api/settlements/settlement-1/confirm", "alice", "")
	if response.StatusCode != http.StatusConflict {
		t.Fatalf("mong đợi 409 khi confirm trạng thái kết thúc, got %d", response.StatusCode)
	}
}

func TestSettlementActionsEnforcePayerAndRecipient(t *testing.T) {
	settlementStore := newFakeSettlementsStore()
	settlementStore.byID["settlement-1"] = models.Settlement{
		ID: "settlement-1", BatchID: "batch-1", FromUserID: "bob", ToUserID: "alice",
		AmountMinor: 5000, PaymentCode: "ABC123", Status: models.SettlementStatusPending, CreatedAt: time.Now(),
	}
	app := newTestApp(newFakeGroupStore(), newFakeExpenseStore(), settlementStore)

	// alice không phải người trả → không được phép báo chuyển
	response, _ := doRequest(t, app, http.MethodPost, "/api/settlements/settlement-1/mark-sent", "alice", "")
	if response.StatusCode != http.StatusForbidden {
		t.Fatalf("mark-sent bởi người nhận: mong đợi 403, got %d", response.StatusCode)
	}

	// chưa báo chuyển mà xác nhận → không được phép
	response, _ = doRequest(t, app, http.MethodPost, "/api/settlements/settlement-1/confirm", "alice", "")
	if response.StatusCode != http.StatusConflict {
		t.Fatalf("confirm khi chưa báo chuyển: mong đợi 409, got %d", response.StatusCode)
	}
}

func TestProtectedRoutesRejectAnonymous(t *testing.T) {
	app := newTestApp(newFakeGroupStore(), newFakeExpenseStore(), newFakeSettlementsStore())

	response, _ := doRequest(t, app, http.MethodGet, "/api/groups/group-1", "", "")
	if response.StatusCode != http.StatusUnauthorized {
		t.Fatalf("mong đợi 401 khi chưa xác thực, got %d", response.StatusCode)
	}
}

func TestListUserGroups(t *testing.T) {
	groupStore := newFakeGroupStore()
	app := newTestApp(groupStore, newFakeExpenseStore(), newFakeSettlementsStore())

	// Alice tạo nhóm
	response, payload := doRequest(t, app, http.MethodPost, "/api/groups", "alice", `{"name":"Nhóm du lịch"}`)
	if response.StatusCode != http.StatusCreated {
		t.Fatalf("create group: status=%d body=%v", response.StatusCode, payload)
	}

	// Alice lấy danh sách nhóm của mình
	response, payload = doRequest(t, app, http.MethodGet, "/api/groups", "alice", "")
	if response.StatusCode != http.StatusOK {
		t.Fatalf("list groups alice: status=%d body=%v", response.StatusCode, payload)
	}
	groupsList, ok := payload["groups"].([]any)
	if !ok || len(groupsList) != 1 {
		t.Fatalf("mong đợi 1 nhóm cho alice, got %v", payload)
	}

	// Bob chưa tham gia nhóm nào → trả mảng rỗng
	response, payload = doRequest(t, app, http.MethodGet, "/api/groups", "bob", "")
	if response.StatusCode != http.StatusOK {
		t.Fatalf("list groups bob: status=%d body=%v", response.StatusCode, payload)
	}
	groupsList, ok = payload["groups"].([]any)
	if !ok || len(groupsList) != 0 {
		t.Fatalf("mong đợi 0 nhóm cho bob, got %v", payload)
	}
}

func TestListAndGetExpensesAndVoid(t *testing.T) {
	groupStore := newFakeGroupStore()
	expenseStore := newFakeExpenseStore()
	app := newTestApp(groupStore, expenseStore, newFakeSettlementsStore())

	// Tạo nhóm và bob tham gia
	response, payload := doRequest(t, app, http.MethodPost, "/api/groups", "alice", `{"name":"Nhóm xem phim"}`)
	if response.StatusCode != http.StatusCreated {
		t.Fatalf("create group: status=%d", response.StatusCode)
	}
	group := payload["group"].(map[string]any)
	groupID := group["id"].(string)
	shareCode := group["shareCode"].(string)

	doRequest(t, app, http.MethodPost, "/api/groups/join/"+shareCode, "bob", "")

	// Tạo khoản chi
	expenseBody := `{"paidBy":"alice","description":"Vé xem phim","amountMinor":80000,"splitType":"EQUAL","splits":[{"userId":"alice","shareMinor":40000},{"userId":"bob","shareMinor":40000}]}`
	response, payload = doRequest(t, app, http.MethodPost, "/api/groups/"+groupID+"/expenses", "alice", expenseBody)
	if response.StatusCode != http.StatusCreated {
		t.Fatalf("create expense: status=%d body=%v", response.StatusCode, payload)
	}
	expenseID := payload["expense"].(map[string]any)["id"].(string)

	// Lấy danh sách khoản chi của nhóm
	response, payload = doRequest(t, app, http.MethodGet, "/api/groups/"+groupID+"/expenses", "bob", "")
	if response.StatusCode != http.StatusOK {
		t.Fatalf("list expenses: status=%d body=%v", response.StatusCode, payload)
	}
	expensesList := payload["expenses"].([]any)
	if len(expensesList) != 1 {
		t.Fatalf("mong đợi 1 khoản chi, got %d", len(expensesList))
	}
	item := expensesList[0].(map[string]any)
	splits := item["splits"].([]any)
	if len(splits) != 2 {
		t.Fatalf("mong đợi 2 phần chia, got %d", len(splits))
	}

	// Lấy chi tiết một khoản chi
	response, payload = doRequest(t, app, http.MethodGet, "/api/groups/"+groupID+"/expenses/"+expenseID, "bob", "")
	if response.StatusCode != http.StatusOK {
		t.Fatalf("get expense: status=%d body=%v", response.StatusCode, payload)
	}
	if payload["expense"].(map[string]any)["id"] != expenseID {
		t.Fatalf("mong đợi id khoản chi %s, got %v", expenseID, payload)
	}

	// Bob không phải người tạo → không được phép hủy khoản chi
	response, _ = doRequest(t, app, http.MethodPost, "/api/groups/"+groupID+"/expenses/"+expenseID+"/void", "bob", "")
	if response.StatusCode != http.StatusForbidden {
		t.Fatalf("void by non-owner: mong đợi 403, got %d", response.StatusCode)
	}

	// Alice là người tạo → hủy thành công
	response, payload = doRequest(t, app, http.MethodPost, "/api/groups/"+groupID+"/expenses/"+expenseID+"/void", "alice", "")
	if response.StatusCode != http.StatusOK {
		t.Fatalf("void by owner: status=%d body=%v", response.StatusCode, payload)
	}
	if payload["expense"].(map[string]any)["status"] != "VOIDED" {
		t.Fatalf("trạng thái sau void phải là VOIDED, got %v", payload["expense"])
	}

	// Hủy lần nữa phải trả 409 Conflict
	response, _ = doRequest(t, app, http.MethodPost, "/api/groups/"+groupID+"/expenses/"+expenseID+"/void", "alice", "")
	if response.StatusCode != http.StatusConflict {
		t.Fatalf("void already voided: mong đợi 409, got %d", response.StatusCode)
	}
}

func TestCancelBatch(t *testing.T) {
	groupStore := newFakeGroupStore()
	settlementStore := newFakeSettlementsStore()
	app := newTestApp(groupStore, newFakeExpenseStore(), settlementStore)

	// Alice tạo nhóm
	doRequest(t, app, http.MethodPost, "/api/groups", "alice", `{"name":"Nhóm tất toán"}`)

	// Hủy kỳ quyết toán
	response, payload := doRequest(t, app, http.MethodPost, "/api/groups/group-1/settlement-batches/batch-1/cancel", "alice", "")
	if response.StatusCode != http.StatusOK {
		t.Fatalf("cancel batch: status=%d body=%v", response.StatusCode, payload)
	}
	if payload["batch"] == nil {
		t.Fatalf("mong đợi batch snapshot trong kết quả, got %v", payload)
	}
}

func TestGetSettlement(t *testing.T) {
	groupStore := newFakeGroupStore()
	settlementStore := newFakeSettlementsStore()
	settlementStore.byID["settlement-1"] = models.Settlement{
		ID: "settlement-1", BatchID: "batch-1", FromUserID: "bob", ToUserID: "alice",
		AmountMinor: 5000, PaymentCode: "ABC123", Status: models.SettlementStatusPending, CreatedAt: time.Now(),
	}
	app := newTestApp(groupStore, newFakeExpenseStore(), settlementStore)

	// Tạo nhóm cho alice
	doRequest(t, app, http.MethodPost, "/api/groups", "alice", `{"name":"Nhóm hoàn tiền"}`)

	// Alice (thành viên nhóm) lấy chi tiết giao dịch
	response, payload := doRequest(t, app, http.MethodGet, "/api/settlements/settlement-1", "alice", "")
	if response.StatusCode != http.StatusOK {
		t.Fatalf("get settlement: status=%d body=%v", response.StatusCode, payload)
	}
	settleData := payload["settlement"].(map[string]any)
	if settleData["id"] != "settlement-1" {
		t.Fatalf("mong đợi settlement-1, got %v", settleData)
	}

	// Charlie (không thuộc nhóm) lấy chi tiết giao dịch → 403 Forbidden
	response, _ = doRequest(t, app, http.MethodGet, "/api/settlements/settlement-1", "charlie", "")
	if response.StatusCode != http.StatusForbidden {
		t.Fatalf("get settlement non-member: mong đợi 403, got %d", response.StatusCode)
	}

	// Settlement không tồn tại → 404 Not Found
	response, _ = doRequest(t, app, http.MethodGet, "/api/settlements/nonexistent", "alice", "")
	if response.StatusCode != http.StatusNotFound {
		t.Fatalf("get settlement nonexistent: mong đợi 404, got %d", response.StatusCode)
	}
}
