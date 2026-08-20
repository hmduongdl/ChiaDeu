package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/hmduongdl/ChiaDeu/models"
	"github.com/hmduongdl/ChiaDeu/pkg/settlements"
	"github.com/hmduongdl/ChiaDeu/pkg/vercel"
)

// Settlements xử lý tất cả /api/settlements/:settlementId/:action
// action (từ query param "action") = mark-sent | confirm | reject
func Settlements(w http.ResponseWriter, r *http.Request) {
	if vercel.HandleCORS(w, r) {
		return
	}

	vercel.WithAuth(handleSettlementAction)(w, r)
}

type settleFunc func(ctx context.Context, settlementID, actorID string) (models.Settlement, error)

// handleSettlementAction phân nhánh theo query param "action"
func handleSettlementAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		vercel.SendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userID, ok := vercel.GetUserID(r)
	if !ok {
		vercel.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	settlementID := r.URL.Query().Get("settlementId")
	actionStr := r.URL.Query().Get("action")
	if settlementID == "" || actionStr == "" {
		vercel.SendError(w, http.StatusBadRequest, "mã giao dịch và hành động là bắt buộc")
		return
	}

	ctx := r.Context()
	pool, err := vercel.GetDB(ctx)
	if err != nil {
		vercel.SendError(w, http.StatusInternalServerError, "database connection error")
		return
	}

	store := vercel.GetSettlementsStore(pool)

	var fn settleFunc
	switch actionStr {
	case "mark-sent":
		fn = store.MarkSent
	case "confirm":
		fn = store.Confirm
	case "reject":
		fn = store.Reject
	default:
		vercel.SendError(w, http.StatusBadRequest, "hành động không hợp lệ, dùng: mark-sent | confirm | reject")
		return
	}

	settlement, err := fn(ctx, settlementID, userID)
	if err != nil {
		switch {
		case errors.Is(err, settlements.ErrSettlementNotFound):
			vercel.SendError(w, http.StatusNotFound, "giao dịch hoàn tiền không tồn tại")
		case errors.Is(err, settlements.ErrNotGroupMember),
			errors.Is(err, settlements.ErrNotPayer),
			errors.Is(err, settlements.ErrNotRecipient):
			vercel.SendError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, settlements.ErrBatchClosed),
			errors.Is(err, settlements.ErrInvalidTransition):
			vercel.SendError(w, http.StatusConflict, err.Error())
		default:
			vercel.SendError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	vercel.SendJSON(w, http.StatusOK, map[string]interface{}{"settlement": settlement})
}
