// Package handlers xử lý HTTP request/response cho các endpoint.
// File này triển khai AuthHandler với các handler:
//   - Register: POST /api/auth/register — tạo tài khoản mới
//   - Login: POST /api/auth/login — đăng nhập, set cookie access + refresh
//   - Refresh: POST /api/auth/refresh — cấp lại access token từ refresh token
//   - Logout: POST /api/auth/logout — xóa cookie
//   - Me: GET /api/auth/me — trả về thông tin user đang đăng nhập
//
// Cookie access token dùng HttpOnly + Lax, refresh token dùng HttpOnly + Strict + path hẹp.
package handlers

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/hmduongdl/ChiaDeu/internal/auth"
	authmiddleware "github.com/hmduongdl/ChiaDeu/internal/middleware"
)

const (
	accessCookieName  = "accessToken"
	refreshCookieName = "refreshToken"
)

type CookieConfig struct {
	Secure bool
	Domain string
}

type AuthHandler struct {
	service *auth.Service
	tokens  *auth.TokenManager
	cookies CookieConfig
}

func NewAuthHandler(service *auth.Service, tokens *auth.TokenManager, cookies CookieConfig) *AuthHandler {
	return &AuthHandler{service: service, tokens: tokens, cookies: cookies}
}

type registerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var request registerRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	// TODO: Add IP/account-based rate limiting before public production launch.
	user, err := h.service.Register(c.UserContext(), request.Name, request.Email, request.Password)
	if err != nil {
		var validationError *auth.ValidationError
		switch {
		case errors.As(err, &validationError):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": validationError.Message})
		case errors.Is(err, auth.ErrEmailExists):
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "email already registered"})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
		}
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"user": user})
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var request loginRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	// TODO: Add IP/account-based rate limiting before public production launch.
	user, err := h.service.Authenticate(c.UserContext(), request.Email, request.Password)
	if errors.Is(err, auth.ErrInvalidCredentials) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid email or password"})
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	pair, err := h.issueTokens(c, user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	h.setAccessCookie(c, pair.AccessToken)
	h.setRefreshCookie(c, pair.RefreshToken)

	return c.JSON(fiber.Map{"user": user})
}

// issueTokens cấp cặp token, kèm lưu phiên khi session store được cấu hình.
func (h *AuthHandler) issueTokens(c *fiber.Ctx, userID string) (auth.TokenPair, error) {
	if h.service.HasSessions() {
		return h.service.StartSession(c.UserContext(), userID)
	}
	return h.tokens.CreatePair(userID)
}

func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	rawRefresh := c.Cookies(refreshCookieName)

	// Chế độ phiên có trạng thái: rotate phiên và thu hồi phiên cũ phía server.
	if h.service.HasSessions() {
		pair, err := h.service.RotateSession(c.UserContext(), rawRefresh)
		if err != nil {
			h.clearCookies(c)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
		}
		h.setAccessCookie(c, pair.AccessToken)
		h.setRefreshCookie(c, pair.RefreshToken)
		return c.JSON(fiber.Map{"message": "access token refreshed"})
	}

	userID, err := h.tokens.VerifyRefreshToken(rawRefresh)
	if err != nil {
		h.clearCookies(c)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	// Ensure the token cannot refresh a session for a user removed from the database.
	if _, err := h.service.CurrentUser(c.UserContext(), userID); errors.Is(err, auth.ErrUserNotFound) {
		h.clearCookies(c)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	accessToken, err := h.tokens.CreateAccessToken(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	h.setAccessCookie(c, accessToken)

	return c.JSON(fiber.Map{"message": "access token refreshed"})
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	_ = h.service.RevokeSession(c.UserContext(), c.Cookies(refreshCookieName))
	h.clearCookies(c)
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *AuthHandler) Me(c *fiber.Ctx) error {
	userID, ok := authmiddleware.UserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	user, err := h.service.CurrentUser(c.UserContext(), userID)
	if errors.Is(err, auth.ErrUserNotFound) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.JSON(fiber.Map{"user": user})
}

func (h *AuthHandler) setAccessCookie(c *fiber.Ctx, value string) {
	// HttpOnly keeps JavaScript from reading the JWT; Lax still supports normal app navigation.
	c.Cookie(&fiber.Cookie{
		Name:     accessCookieName,
		Value:    value,
		Path:     "/",
		Domain:   h.cookies.Domain,
		MaxAge:   int(auth.AccessTokenDuration.Seconds()),
		Expires:  time.Now().Add(auth.AccessTokenDuration),
		Secure:   h.cookies.Secure,
		HTTPOnly: true,
		SameSite: fiber.CookieSameSiteLaxMode,
	})
}

func (h *AuthHandler) setRefreshCookie(c *fiber.Ctx, value string) {
	// Strict and a narrow path reduce where the longer-lived refresh credential is sent.
	c.Cookie(&fiber.Cookie{
		Name:     refreshCookieName,
		Value:    value,
		Path:     "/api/auth/refresh",
		Domain:   h.cookies.Domain,
		MaxAge:   int(auth.RefreshTokenDuration.Seconds()),
		Expires:  time.Now().Add(auth.RefreshTokenDuration),
		Secure:   h.cookies.Secure,
		HTTPOnly: true,
		SameSite: fiber.CookieSameSiteStrictMode,
	})
}

func (h *AuthHandler) clearCookies(c *fiber.Ctx) {
	expires := time.Unix(1, 0)
	for _, cookie := range []fiber.Cookie{
		{Name: accessCookieName, Path: "/", SameSite: fiber.CookieSameSiteLaxMode},
		{Name: refreshCookieName, Path: "/api/auth/refresh", SameSite: fiber.CookieSameSiteStrictMode},
	} {
		cookie.Domain = h.cookies.Domain
		cookie.Value = ""
		cookie.MaxAge = -1
		cookie.Expires = expires
		cookie.Secure = h.cookies.Secure
		cookie.HTTPOnly = true
		c.Cookie(&cookie)
	}
}
