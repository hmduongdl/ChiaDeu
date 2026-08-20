package handler

import (
	"errors"
	"net/http"

	"github.com/hmduongdl/ChiaDeu/pkg/settlements"
	"github.com/hmduongdl/ChiaDeu/pkg/vercel"
)

// Handler xử lý GET /api/groups/:groupId/settlement-batches/:batchId
func Handler(w http.ResponseWriter, r *http.Request) {
	vercel.WithAuth(handleGetBatch)(w, r)
}

func handleGetBatch(w http.ResponseWriter, r *http.Request) {
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
	batchID := r.URL.Query().Get("batchId")
	if groupID == "" || batchID == "" {
		vercel.SendError(w, http.StatusBadRequest, "mã nhóm và mã kỳ là bắt buộc")
		return
	}

	ctx := r.Context()
	pool, err := vercel.GetDB(ctx)
	if err != nil {
		vercel.SendError(w, http.StatusInternalServerError, "database connection error")
		return
	}

	settlementsStore := vercel.GetSettlementsStore(pool)
	snapshot, err := settlementsStore.GetBatch(ctx, groupID, batchID, userID)
	if err != nil {
		switch {
		case errors.Is(err, settlements.ErrBatchNotFound):
			vercel.SendError(w, http.StatusNotFound, "kỳ quyết toán không tồn tại")
		case errors.Is(err, settlements.ErrNotGroupMember):
			vercel.SendError(w, http.StatusForbidden, "bạn không phải là thành viên nhóm")
		default:
			vercel.SendError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	vercel.SendJSON(w, http.StatusOK, snapshot)
}
