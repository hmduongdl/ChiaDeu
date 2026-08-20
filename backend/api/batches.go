package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/hmduongdl/ChiaDeu/pkg/settlements"
	"github.com/hmduongdl/ChiaDeu/pkg/vercel"
)

// Batches xử lý tất cả /api/groups/:groupId/settlement-batches/* — phân nhánh theo query param "sub"
// sub = create | detail
func Batches(w http.ResponseWriter, r *http.Request) {
	if vercel.HandleCORS(w, r) {
		return
	}

	vercel.WithAuth(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("sub") {
		case "create":
			handleCloseBatch(w, r)
		case "detail":
			handleGetBatch(w, r)
		default:
			vercel.SendError(w, http.StatusNotFound, "route not found")
		}
	})(w, r)
}

// --- close batch ---

type closeBatchRequest struct {
	IdempotencyKey string `json:"idempotencyKey"`
}

// handleCloseBatch xử lý POST /api/groups/:groupId/settlement-batches
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

	// Đọc idempotency key từ header hoặc body
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

// --- get batch ---

// handleGetBatch xử lý GET /api/groups/:groupId/settlement-batches/:batchId
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
