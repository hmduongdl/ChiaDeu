package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/hmduongdl/ChiaDeu/internal/auth"
	"github.com/hmduongdl/ChiaDeu/internal/config"
	"github.com/hmduongdl/ChiaDeu/internal/vercel"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Handler xử lý POST /api/auth/login
func Handler(w http.ResponseWriter, r *http.Request) {
	if vercel.HandleCORS(w, r) {
		return
	}

	if r.Method != http.MethodPost {
		vercel.SendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req loginRequest
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
	user, err := authService.Authenticate(ctx, req.Email, req.Password)
	if errors.Is(err, auth.ErrInvalidCredentials) {
		vercel.SendError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}
	if err != nil {
		vercel.SendError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	var pair auth.TokenPair
	if authService.HasSessions() {
		pair, err = authService.StartSession(ctx, user.ID)
	} else {
		pair, err = tokens.CreatePair(user.ID)
	}
	if err != nil {
		vercel.SendError(w, http.StatusInternalServerError, "failed to issue tokens")
		return
	}

	vercel.SetAccessCookie(w, cfg, pair.AccessToken)
	vercel.SetRefreshCookie(w, cfg, pair.RefreshToken)

	vercel.SendJSON(w, http.StatusOK, map[string]interface{}{"user": user})
}
