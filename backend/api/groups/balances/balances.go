package handler

import (
	"errors"
	"net/http"

	"github.com/hmduongdl/ChiaDeu/internal/groups"
	"github.com/hmduongdl/ChiaDeu/internal/vercel"
)

// Handler xử lý GET /api/groups/:groupId/balances
func Handler(w http.ResponseWriter, r *http.Request) {
	vercel.WithAuth(handleBalances)(w, r)
}

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
