// Package config — unit test cho Load().
// File này kiểm thử:
//   - FRONTEND_ORIGIN phải là origin hợp lệ (từ chối "*", ftp://, có path)
//   - CookieSecure mặc định là true nếu không set biến môi trường
package config

import "testing"

func TestLoadRequiresExplicitValidOrigin(t *testing.T) {
	setRequiredEnv(t)

	for _, origin := range []string{"*", "ftp://example.com", "https://example.com/path"} {
		t.Run(origin, func(t *testing.T) {
			t.Setenv("FRONTEND_ORIGIN", origin)
			if _, err := Load(); err == nil {
				t.Fatalf("expected origin %q to fail", origin)
			}
		})
	}
}

func TestLoadDefaultsCookiesToSecure(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("COOKIE_SECURE", "")

	loaded, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !loaded.CookieSecure {
		t.Fatal("cookies must default to secure")
	}
}

func setRequiredEnv(t *testing.T) {
	t.Helper()
	t.Setenv("DATABASE_URL", "postgres://localhost/test")
	t.Setenv("JWT_ACCESS_SECRET", "access-secret")
	t.Setenv("JWT_REFRESH_SECRET", "refresh-secret")
	t.Setenv("FRONTEND_ORIGIN", "https://app.example.com")
}
