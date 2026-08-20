package handler

import (
	"errors"
	"net/http"

	"github.com/hmduongdl/ChiaDeu/internal/groups"
	"github.com/hmduongdl/ChiaDeu/internal/vercel"
)

// Handler xử lý GET /api/groups/:groupId
func Handler(w http.ResponseWriter, r *http.Request) {
	vercel.WithAuth(handleGetGroup)(w, r)
}

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
