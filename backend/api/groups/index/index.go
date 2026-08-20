package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/hmduongdl/ChiaDeu/internal/groups"
	"github.com/hmduongdl/ChiaDeu/internal/vercel"
)

type createGroupRequest struct {
	Name     string `json:"name"`
	Currency string `json:"currency"`
}

// Handler xử lý POST /api/groups
func Handler(w http.ResponseWriter, r *http.Request) {
	vercel.WithAuth(handleCreateGroup)(w, r)
}

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
