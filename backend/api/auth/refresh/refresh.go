package handler

import (
	"errors"
	"net/http"

	"github.com/hmduongdl/ChiaDeu/pkg/auth"
	"github.com/hmduongdl/ChiaDeu/pkg/config"
	"github.com/hmduongdl/ChiaDeu/pkg/vercel"
)

// Handler xử lý POST /api/auth/refresh
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

	cookie, err := r.Cookie("refreshToken")
	if err != nil {
		vercel.ClearCookies(w, cfg)
		vercel.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	rawRefresh := cookie.Value

	pool, err := vercel.GetDB(ctx)
	if err != nil {
		vercel.SendError(w, http.StatusInternalServerError, "database connection error")
		return
	}

	tokens, err := auth.NewTokenManager(cfg.JWTAccessSecret, cfg.JWTRefreshSecret)
	if err != nil {
		vercel.SendError(w, http.StatusInternalServerError, "token manager error")
		return
	}

	authService := vercel.GetAuthService(ctx, pool, tokens)

	if authService.HasSessions() {
		pair, err := authService.RotateSession(ctx, rawRefresh)
		if err != nil {
			vercel.ClearCookies(w, cfg)
			vercel.SendError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		vercel.SetAccessCookie(w, cfg, pair.AccessToken)
		vercel.SetRefreshCookie(w, cfg, pair.RefreshToken)
		vercel.SendJSON(w, http.StatusOK, map[string]string{"message": "access token refreshed"})
		return
	}

	userID, err := tokens.VerifyRefreshToken(rawRefresh)
	if err != nil {
		vercel.ClearCookies(w, cfg)
		vercel.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if _, err := authService.CurrentUser(ctx, userID); errors.Is(err, auth.ErrUserNotFound) {
		vercel.ClearCookies(w, cfg)
		vercel.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	} else if err != nil {
		vercel.SendError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	accessToken, err := tokens.CreateAccessToken(userID)
	if err != nil {
		vercel.SendError(w, http.StatusInternalServerError, "failed to create access token")
		return
	}
	vercel.SetAccessCookie(w, cfg, accessToken)

	vercel.SendJSON(w, http.StatusOK, map[string]string{"message": "access token refreshed"})
}
