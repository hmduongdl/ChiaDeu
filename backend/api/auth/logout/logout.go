package handler

import (
	"net/http"

	"github.com/hmduongdl/ChiaDeu/pkg/auth"
	"github.com/hmduongdl/ChiaDeu/pkg/config"
	"github.com/hmduongdl/ChiaDeu/pkg/vercel"
)

// Handler xử lý POST /api/auth/logout
func Handler(w http.ResponseWriter, r *http.Request) {
	if vercel.HandleCORS(w, r) {
		return
	}

	if r.Method != http.MethodPost {
		vercel.SendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	ctx := r.Context()
	cfg, err := config.Load()
	if err != nil {
		vercel.SendError(w, http.StatusInternalServerError, "configuration error")
		return
	}

	pool, err := vercel.GetDB(ctx)
	if err == nil {
		tokens, _ := auth.NewTokenManager(cfg.JWTAccessSecret, cfg.JWTRefreshSecret)
		authService := vercel.GetAuthService(ctx, pool, tokens)
		if authService.HasSessions() {
			var refreshVal string
			if cookie, err := r.Cookie("refreshToken"); err == nil {
				refreshVal = cookie.Value
			}
			_ = authService.RevokeSession(ctx, refreshVal)
		}
	}

	vercel.ClearCookies(w, cfg)
	w.WriteHeader(http.StatusNoContent)
}
