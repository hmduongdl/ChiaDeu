package handler

import (
	"errors"
	"net/http"

	"github.com/hmduongdl/ChiaDeu/pkg/auth"
	"github.com/hmduongdl/ChiaDeu/pkg/config"
	"github.com/hmduongdl/ChiaDeu/pkg/vercel"
)

// Handler xử lý GET /api/auth/me
func Handler(w http.ResponseWriter, r *http.Request) {
	vercel.WithAuth(handleMe)(w, r)
}

func handleMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		vercel.SendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userID, ok := vercel.GetUserID(r)
	if !ok {
		vercel.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	ctx := r.Context()
	pool, err := vercel.GetDB(ctx)
	if err != nil {
		vercel.SendError(w, http.StatusInternalServerError, "database connection error")
		return
	}

	cfg, err := config.Load()
	if err != nil {
		vercel.SendError(w, http.StatusInternalServerError, "configuration error")
		return
	}

	tokens, err := auth.NewTokenManager(cfg.JWTAccessSecret, cfg.JWTRefreshSecret)
	if err != nil {
		vercel.SendError(w, http.StatusInternalServerError, "token manager error")
		return
	}

	authService := vercel.GetAuthService(ctx, pool, tokens)
	user, err := authService.CurrentUser(ctx, userID)
	if errors.Is(err, auth.ErrUserNotFound) {
		vercel.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if err != nil {
		vercel.SendError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	vercel.SendJSON(w, http.StatusOK, map[string]interface{}{"user": user})
}
