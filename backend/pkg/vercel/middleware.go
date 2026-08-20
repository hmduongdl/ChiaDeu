package vercel

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/hmduongdl/ChiaDeu/pkg/auth"
	"github.com/hmduongdl/ChiaDeu/pkg/config"
	"github.com/hmduongdl/ChiaDeu/pkg/expenses"
	"github.com/hmduongdl/ChiaDeu/pkg/groups"
	"github.com/hmduongdl/ChiaDeu/pkg/settlements"
	"github.com/jackc/pgx/v5/pgxpool"
)

type contextKey string

const UserIDKey contextKey = "userID"

var (
	dbPool *pgxpool.Pool
	poolMu sync.Mutex
)

// ApiResponse đại diện cho cấu trúc phản hồi API chuẩn hoá.
type ApiResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// GetDB trả về pool kết nối database dùng chung, tối ưu hóa cho môi trường serverless.
func GetDB(ctx context.Context) (*pgxpool.Pool, error) {
	poolMu.Lock()
	defer poolMu.Unlock()
	if dbPool != nil {
		return dbPool, nil
	}

	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	poolCfg, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	// Tối ưu cho Serverless: Giới hạn pool tối đa 2 connection để tránh cạn kiệt kết nối DB
	poolCfg.MaxConns = 2
	poolCfg.MinConns = 0
	poolCfg.MaxConnIdleTime = 15 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, err
	}

	dbPool = pool
	return dbPool, nil
}

// WithAuth là middleware xác thực JWT từ cookie accessToken.
func WithAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if HandleCORS(w, r) {
			return
		}

		cookie, err := r.Cookie("accessToken")
		if err != nil {
			SendError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		cfg, err := config.Load()
		if err != nil {
			SendError(w, http.StatusInternalServerError, "configuration error")
			return
		}

		tokens, err := auth.NewTokenManager(cfg.JWTAccessSecret, cfg.JWTRefreshSecret)
		if err != nil {
			SendError(w, http.StatusInternalServerError, "token manager error")
			return
		}

		userID, err := tokens.VerifyAccessToken(cookie.Value)
		if err != nil {
			SendError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next(w, r.WithContext(ctx))
	}
}

// GetUserID lấy userID đã được xác thực từ context request.
func GetUserID(r *http.Request) (string, bool) {
	userID, ok := r.Context().Value(UserIDKey).(string)
	return userID, ok && userID != ""
}

// HandleCORS cấu hình CORS thủ công cho từng handler.
func HandleCORS(w http.ResponseWriter, r *http.Request) bool {
	cfg, err := config.Load()
	origin := "http://localhost:3000"
	if err == nil && cfg.FrontendOrigin != "" {
		origin = cfg.FrontendOrigin
	}

	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Idempotency-Key")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return true
	}
	return false
}

// SendJSON trả về JSON phản hồi thành công.
func SendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ApiResponse{
		Success: true,
		Data:    data,
	})
}

// SendError trả về JSON phản hồi lỗi.
func SendError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ApiResponse{
		Success: false,
		Error:   message,
	})
}

// Cookie helpers
func SetAccessCookie(w http.ResponseWriter, cfg config.Config, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "accessToken",
		Value:    value,
		Path:     "/",
		Domain:   cfg.CookieDomain,
		MaxAge:   int(auth.AccessTokenDuration.Seconds()),
		Expires:  time.Now().Add(auth.AccessTokenDuration),
		Secure:   cfg.CookieSecure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func SetRefreshCookie(w http.ResponseWriter, cfg config.Config, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refreshToken",
		Value:    value,
		Path:     "/api/auth/refresh",
		Domain:   cfg.CookieDomain,
		MaxAge:   int(auth.RefreshTokenDuration.Seconds()),
		Expires:  time.Now().Add(auth.RefreshTokenDuration),
		Secure:   cfg.CookieSecure,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}

func ClearCookies(w http.ResponseWriter, cfg config.Config) {
	expires := time.Unix(1, 0)
	http.SetCookie(w, &http.Cookie{
		Name:     "accessToken",
		Value:    "",
		Path:     "/",
		Domain:   cfg.CookieDomain,
		MaxAge:   -1,
		Expires:  expires,
		Secure:   cfg.CookieSecure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refreshToken",
		Value:    "",
		Path:     "/api/auth/refresh",
		Domain:   cfg.CookieDomain,
		MaxAge:   -1,
		Expires:  expires,
		Secure:   cfg.CookieSecure,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}

// Services initialization helpers
func GetAuthService(ctx context.Context, pool *pgxpool.Pool, tokens *auth.TokenManager) *auth.Service {
	sessionStore := auth.NewPostgresSessionStore(pool)
	return auth.NewServiceWithSessions(auth.NewPostgresUserStore(pool), tokens, sessionStore)
}

func GetGroupsService(pool *pgxpool.Pool) *groups.Service {
	return groups.NewService(groups.NewPostgresStore(pool))
}

func GetExpensesService(pool *pgxpool.Pool) *expenses.Service {
	return expenses.NewService(expenses.NewPostgresStore(pool))
}

func GetSettlementsStore(pool *pgxpool.Pool) settlements.Store {
	return settlements.NewPostgresStore(pool)
}
