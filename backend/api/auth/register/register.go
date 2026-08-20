package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/hmduongdl/ChiaDeu/internal/auth"
	"github.com/hmduongdl/ChiaDeu/internal/config"
	"github.com/hmduongdl/ChiaDeu/internal/vercel"
)

type registerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Handler xử lý POST /api/auth/register
func Handler(w http.ResponseWriter, r *http.Request) {
	if vercel.HandleCORS(w, r) {
		return
	}

	if r.Method != http.MethodPost {
		vercel.SendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req registerRequest
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
	user, err := authService.Register(ctx, req.Name, req.Email, req.Password)
	if err != nil {
		var validationError *auth.ValidationError
		switch {
		case errors.As(err, &validationError):
			vercel.SendError(w, http.StatusBadRequest, validationError.Message)
		case errors.Is(err, auth.ErrEmailExists):
			vercel.SendError(w, http.StatusConflict, "email already registered")
		default:
			vercel.SendError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	vercel.SendJSON(w, http.StatusCreated, map[string]interface{}{"user": user})
}
