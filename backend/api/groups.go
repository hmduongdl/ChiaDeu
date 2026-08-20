package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/hmduongdl/ChiaDeu/pkg/groups"
	"github.com/hmduongdl/ChiaDeu/pkg/vercel"
)

// Groups xử lý tất cả /api/groups/* — phân nhánh theo query param "sub"
// sub = create | join | detail | balances
func Groups(w http.ResponseWriter, r *http.Request) {
	if vercel.HandleCORS(w, r) {
		return
	}

	vercel.WithAuth(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("sub") {
		case "create":
			handleCreateGroup(w, r)
		case "join":
			handleJoinGroup(w, r)
		case "detail":
			handleGetGroup(w, r)
		case "balances":
			handleBalances(w, r)
		default:
			vercel.SendError(w, http.StatusNotFound, "route not found")
		}
	})(w, r)
}

// --- create group ---

type createGroupRequest struct {
	Name     string `json:"name"`
	Currency string `json:"currency"`
}

// handleCreateGroup xử lý POST /api/groups
func handleCreateGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		vercel.SendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userID, ok := vercel.GetUserID(r)
	if !ok {
		vercel.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req createGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		vercel.SendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ctx := r.Context()
	pool, err := vercel.GetDB(ctx)
	if err != nil {
		vercel.SendError(w, http.StatusInternalServerError, "database connection error")
		return
	}

	groupsService := vercel.GetGroupsService(pool)
	group, err := groupsService.CreateGroup(ctx, userID, req.Name, req.Currency)
	if err != nil {
		switch {
		case errors.Is(err, groups.ErrInvalidGroupName), errors.Is(err, groups.ErrInvalidCurrency):
			vercel.SendError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, groups.ErrUserNotFound):
			vercel.SendError(w, http.StatusUnauthorized, "unauthorized")
		default:
			vercel.SendError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	vercel.SendJSON(w, http.StatusCreated, map[string]interface{}{"group": group})
}

// --- join group ---

// handleJoinGroup xử lý POST /api/groups/join/:shareCode
func handleJoinGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		vercel.SendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userID, ok := vercel.GetUserID(r)
	if !ok {
		vercel.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	shareCode := r.URL.Query().Get("shareCode")
	if shareCode == "" {
		vercel.SendError(w, http.StatusBadRequest, "mã chia sẻ là bắt buộc")
		return
	}

	ctx := r.Context()
	pool, err := vercel.GetDB(ctx)
	if err != nil {
		vercel.SendError(w, http.StatusInternalServerError, "database connection error")
		return
	}

	groupsService := vercel.GetGroupsService(pool)
	group, err := groupsService.JoinGroup(ctx, userID, shareCode)
	if err != nil {
		switch {
		case errors.Is(err, groups.ErrInvalidShareCode):
			vercel.SendError(w, http.StatusNotFound, "mã chia sẻ không hợp lệ")
		case errors.Is(err, groups.ErrAlreadyMember):
			vercel.SendError(w, http.StatusConflict, "bạn đã là thành viên nhóm này")
		case errors.Is(err, groups.ErrGroupArchived):
			vercel.SendError(w, http.StatusConflict, err.Error())
		case errors.Is(err, groups.ErrUserNotFound):
			vercel.SendError(w, http.StatusUnauthorized, "unauthorized")
		default:
			vercel.SendError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	vercel.SendJSON(w, http.StatusOK, map[string]interface{}{"group": group})
}

// --- group detail ---

// handleGetGroup xử lý GET /api/groups/:groupId
func handleGetGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
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

	ctx := r.Context()
	pool, err := vercel.GetDB(ctx)
	if err != nil {
		vercel.SendError(w, http.StatusInternalServerError, "database connection error")
		return
	}

	groupsService := vercel.GetGroupsService(pool)
	detail, err := groupsService.GetGroup(ctx, userID, groupID)
	if err != nil {
		switch {
		case errors.Is(err, groups.ErrGroupNotFound):
			vercel.SendError(w, http.StatusNotFound, "nhóm không tồn tại")
		case errors.Is(err, groups.ErrNotMember):
			vercel.SendError(w, http.StatusForbidden, "bạn không phải là thành viên nhóm")
		default:
			vercel.SendError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	vercel.SendJSON(w, http.StatusOK, map[string]interface{}{
		"group":   detail.Group,
		"members": detail.Members,
	})
}

// --- balances ---

// handleBalances xử lý GET /api/groups/:groupId/balances
func handleBalances(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
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

	ctx := r.Context()
	pool, err := vercel.GetDB(ctx)
	if err != nil {
		vercel.SendError(w, http.StatusInternalServerError, "database connection error")
		return
	}

	expensesService := vercel.GetExpensesService(pool)
	groupsService := vercel.GetGroupsService(pool)

	balances, err := expensesService.UnsettledBalances(ctx, groupID)
	if err != nil {
		vercel.SendError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	detail, err := groupsService.GetGroup(ctx, userID, groupID)
	if err != nil {
		if errors.Is(err, groups.ErrNotMember) || errors.Is(err, groups.ErrGroupNotFound) {
			vercel.SendError(w, http.StatusForbidden, "không thể xem số dư nhóm này")
			return
		}
		vercel.SendError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	names := make(map[string]string, len(detail.Members))
	for _, member := range detail.Members {
		names[member.ID] = member.DisplayName()
	}

	type balanceEntry struct {
		UserID       string `json:"userId"`
		Name         string `json:"name"`
		BalanceMinor int64  `json:"balanceMinor"`
	}

	result := make([]balanceEntry, 0, len(balances))
	for memberID, balanceMinor := range balances {
		result = append(result, balanceEntry{
			UserID:       memberID,
			Name:         names[memberID],
			BalanceMinor: balanceMinor,
		})
	}

	vercel.SendJSON(w, http.StatusOK, map[string]interface{}{"balances": result})
}
