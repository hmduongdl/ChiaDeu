package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/hmduongdl/ChiaDeu/internal/settlements"
	"github.com/hmduongdl/ChiaDeu/internal/vercel"
)

type closeBatchRequest struct {
	IdempotencyKey string `json:"idempotencyKey"`
}

// Handler xử lý POST /api/groups/:groupId/settlement-batches
func Handler(w http.ResponseWriter, r *http.Request) {
	vercel.WithAuth(handleCloseBatch)(w, r)
}

func handleCloseBatch(w http.ResponseWriter, r *http.Request) {
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

	key := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if key == "" {
		var req closeBatchRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
			key = strings.TrimSpace(req.IdempotencyKey)
		}
	}

	if key == "" {
		vercel.SendError(w, http.StatusBadRequest, "thiếu idempotency key")
		return
	}

	ctx := r.Context()
	pool, err := vercel.GetDB(ctx)
	if err != nil {
		vercel.SendError(w, http.StatusInternalServerError, "database connection error")
		return
	}

	settlementsStore := vercel.GetSettlementsStore(pool)
	snapshot, err := settlementsStore.CloseBatch(ctx, groupID, userID, key)
	if err != nil {
		switch {
		case errors.Is(err, settlements.ErrNotAdmin):
			vercel.SendError(w, http.StatusForbidden, "chỉ quản trị viên mới được chốt kỳ")
		case errors.Is(err, settlements.ErrBatchAlreadyOpen):
			vercel.SendError(w, http.StatusConflict, "nhóm đã có một kỳ đang mở")
		case errors.Is(err, settlements.ErrNoOpenExpenses):
			vercel.SendError(w, http.StatusUnprocessableEntity, "không có khoản chi nào để chốt")
		case errors.Is(err, settlements.ErrInvalidExpenseData):
			vercel.SendError(w, http.StatusUnprocessableEntity, "dữ liệu khoản chi không hợp lệ")
		default:
			vercel.SendError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	vercel.SendJSON(w, http.StatusCreated, snapshot)
}
