package handler

import (
	"errors"
	"net/http"

	"github.com/hmduongdl/ChiaDeu/internal/groups"
	"github.com/hmduongdl/ChiaDeu/internal/vercel"
)

// Handler xử lý POST /api/groups/join/:shareCode
func Handler(w http.ResponseWriter, r *http.Request) {
	vercel.WithAuth(handleJoinGroup)(w, r)
}

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
