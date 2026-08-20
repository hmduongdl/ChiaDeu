package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/hmduongdl/ChiaDeu/pkg/auth"
	"github.com/hmduongdl/ChiaDeu/pkg/config"
	"github.com/hmduongdl/ChiaDeu/pkg/vercel"
)

// Auth xử lý tất cả /api/auth/* — phân nhánh theo query param "sub"
// sub = register | login | refresh | logout | me
func Auth(w http.ResponseWriter, r *http.Request) {
	if vercel.HandleCORS(w, r) {
		return
	}

	switch r.URL.Query().Get("sub") {
	case "register":
		handleRegister(w, r)
	case "login":
		handleLogin(w, r)
	case "refresh":
		handleRefresh(w, r)
	case "logout":
		handleLogout(w, r)
	case "me":
		vercel.WithAuth(handleMe)(w, r)
	default:
		vercel.SendError(w, http.StatusNotFound, "route not found")
	}
}

// --- register ---

type registerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// handleRegister xử lý POST /api/auth/register
func handleRegister(w http.ResponseWriter, r *http.Request) {
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

// --- login ---

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// handleLogin xử lý POST /api/auth/login
func handleLogin(w http.ResponseWriter, r *http.Request) {
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
	cfg, err := config.Load()
	if err != nil {
		vercel.SendError(w, http.StatusInternalServerError, "configuration error")
		return
	}

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

// --- refresh ---

// handleRefresh xử lý POST /api/auth/refresh
func handleRefresh(w http.ResponseWriter, r *http.Request) {
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

// --- logout ---

// handleLogout xử lý POST /api/auth/logout
func handleLogout(w http.ResponseWriter, r *http.Request) {
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

// --- me ---

// handleMe xử lý GET /api/auth/me (đã được bọc WithAuth)
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
