package main

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/hmduongdl/ChiaDeu/internal/auth"
	"github.com/hmduongdl/ChiaDeu/internal/handlers"
	authmiddleware "github.com/hmduongdl/ChiaDeu/internal/middleware"
	"golang.org/x/crypto/bcrypt"
)

const (
	testAccessSecret  = "main-test-access-secret-longer-than-thirty-two-characters"
	testRefreshSecret = "main-test-refresh-secret-longer-than-thirty-two-characters"
)

type testUserStore struct {
	user auth.User
}

func (s *testUserStore) CreateUser(_ context.Context, name, email, passwordHash string) (auth.User, error) {
	s.user = auth.User{ID: "user-123", Name: name, Email: email, PasswordHash: passwordHash, CreatedAt: time.Now()}
	return s.user, nil
}

func (s *testUserStore) FindUserByEmail(_ context.Context, email string) (auth.User, error) {
	if s.user.Email != email {
		return auth.User{}, auth.ErrUserNotFound
	}
	return s.user, nil
}

func (s *testUserStore) FindUserByID(_ context.Context, id string) (auth.User, error) {
	if s.user.ID != id {
		return auth.User{}, auth.ErrUserNotFound
	}
	return s.user, nil
}

func TestHealthRoutes(t *testing.T) {
	app := newApp(appDependencies{frontendOrigin: "http://localhost:3000"})

	for _, path := range []string{"/api/health", "/api/backend/health"} {
		t.Run(path, func(t *testing.T) {
			response, err := app.Test(httptest.NewRequest(http.MethodGet, path, nil))
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			defer response.Body.Close()

			if response.StatusCode != http.StatusOK {
				t.Fatalf("expected status 200, got %d", response.StatusCode)
			}
			body, err := io.ReadAll(response.Body)
			if err != nil {
				t.Fatalf("read response: %v", err)
			}
			if string(body) != `{"status":"ok"}` {
				t.Fatalf("unexpected response body: %s", body)
			}
		})
	}
}

func TestLoginMeRefreshAndLogout(t *testing.T) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash test password: %v", err)
	}
	store := &testUserStore{user: auth.User{
		ID: "user-123", Name: "Test User", Email: "test@example.com", PasswordHash: string(passwordHash), CreatedAt: time.Now(),
	}}
	app, tokens := newAuthenticatedTestApp(t, store)

	loginRequest := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{"email":"test@example.com","password":"password123"}`))
	loginRequest.Header.Set("Content-Type", "application/json")
	loginResponse, err := app.Test(loginRequest)
	if err != nil {
		t.Fatalf("login request: %v", err)
	}
	defer loginResponse.Body.Close()
	if loginResponse.StatusCode != http.StatusOK {
		t.Fatalf("expected login status 200, got %d", loginResponse.StatusCode)
	}

	accessCookie := findCookie(t, loginResponse.Cookies(), "accessToken")
	refreshCookie := findCookie(t, loginResponse.Cookies(), "refreshToken")
	if !accessCookie.HttpOnly || !accessCookie.Secure || accessCookie.Path != "/" || accessCookie.SameSite != http.SameSiteLaxMode || accessCookie.MaxAge != 15*60 {
		t.Fatalf("unexpected access cookie: %+v", accessCookie)
	}
	if !refreshCookie.HttpOnly || !refreshCookie.Secure || refreshCookie.Path != "/api/auth/refresh" || refreshCookie.SameSite != http.SameSiteStrictMode || refreshCookie.MaxAge != 7*24*60*60 {
		t.Fatalf("unexpected refresh cookie: %+v", refreshCookie)
	}

	meRequest := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	meRequest.AddCookie(accessCookie)
	meResponse, err := app.Test(meRequest)
	if err != nil {
		t.Fatalf("me request: %v", err)
	}
	defer meResponse.Body.Close()
	if meResponse.StatusCode != http.StatusOK {
		t.Fatalf("expected me status 200, got %d", meResponse.StatusCode)
	}

	refreshRequest := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	refreshRequest.AddCookie(refreshCookie)
	refreshResponse, err := app.Test(refreshRequest)
	if err != nil {
		t.Fatalf("refresh request: %v", err)
	}
	defer refreshResponse.Body.Close()
	if refreshResponse.StatusCode != http.StatusOK {
		t.Fatalf("expected refresh status 200, got %d", refreshResponse.StatusCode)
	}
	refreshedAccess := findCookie(t, refreshResponse.Cookies(), "accessToken")
	if _, err := tokens.VerifyAccessToken(refreshedAccess.Value); err != nil {
		t.Fatalf("refresh did not issue a valid access token: %v", err)
	}

	logoutResponse, err := app.Test(httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil))
	if err != nil {
		t.Fatalf("logout request: %v", err)
	}
	defer logoutResponse.Body.Close()
	if logoutResponse.StatusCode != http.StatusNoContent {
		t.Fatalf("expected logout status 204, got %d", logoutResponse.StatusCode)
	}
	for _, name := range []string{"accessToken", "refreshToken"} {
		if cookie := findCookie(t, logoutResponse.Cookies(), name); !cookie.Expires.Before(time.Now()) {
			t.Fatalf("expected %s deletion cookie, got %+v", name, cookie)
		}
	}
}

func TestAuthMiddlewareAndCredentialedCORS(t *testing.T) {
	app, _ := newAuthenticatedTestApp(t, &testUserStore{})

	unauthorizedResponse, err := app.Test(httptest.NewRequest(http.MethodGet, "/api/auth/me", nil))
	if err != nil {
		t.Fatalf("unauthorized request: %v", err)
	}
	defer unauthorizedResponse.Body.Close()
	body, _ := io.ReadAll(unauthorizedResponse.Body)
	if unauthorizedResponse.StatusCode != http.StatusUnauthorized || string(body) != `{"error":"unauthorized"}` {
		t.Fatalf("unexpected unauthorized response: status=%d body=%s", unauthorizedResponse.StatusCode, body)
	}

	preflight := httptest.NewRequest(http.MethodOptions, "/api/auth/login", nil)
	preflight.Header.Set("Origin", "http://localhost:3000")
	preflight.Header.Set("Access-Control-Request-Method", http.MethodPost)
	preflightResponse, err := app.Test(preflight)
	if err != nil {
		t.Fatalf("preflight request: %v", err)
	}
	defer preflightResponse.Body.Close()
	if preflightResponse.Header.Get("Access-Control-Allow-Origin") != "http://localhost:3000" || preflightResponse.Header.Get("Access-Control-Allow-Credentials") != "true" {
		t.Fatalf("unexpected CORS headers: %+v", preflightResponse.Header)
	}
}

func newAuthenticatedTestApp(t *testing.T, store auth.UserStore) (*fiber.App, *auth.TokenManager) {
	t.Helper()
	tokens, err := auth.NewTokenManager(testAccessSecret, testRefreshSecret)
	if err != nil {
		t.Fatalf("create token manager: %v", err)
	}
	service := auth.NewService(store)
	handler := handlers.NewAuthHandler(service, tokens, handlers.CookieConfig{Secure: true})
	return newApp(appDependencies{
		frontendOrigin: "http://localhost:3000",
		authHandler:    handler,
		authMiddleware: authmiddleware.AuthMiddleware(tokens),
	}), tokens
}

func findCookie(t *testing.T, cookies []*http.Cookie, name string) *http.Cookie {
	t.Helper()
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie
		}
	}
	t.Fatalf("cookie %q not found", name)
	return nil
}
