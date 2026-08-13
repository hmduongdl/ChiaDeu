// Package config nạp và validate cấu hình ứng dụng từ biến môi trường.
// File này định nghĩa struct Config chứa:
//   - Port: cổng server (mặc định 8080)
//   - DatabaseURL: chuỗi kết nối PostgreSQL (bắt buộc)
//   - FrontendOrigin: origin cho CORS (phải là http(s)://host)
//   - JWT secrets: access và refresh (bắt buộc, mỗi cái >= 32 ký tự)
//   - Cookie config: Secure, Domain
// Hỗ trợ hàm envOrDefault và boolEnv để đọc biến môi trường an toàn.
package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port             string
	DatabaseURL      string
	FrontendOrigin   string
	JWTAccessSecret  string
	JWTRefreshSecret string
	CookieSecure     bool
	CookieDomain     string
}

func Load() (Config, error) {
	cookieSecure, err := boolEnv("COOKIE_SECURE", true)
	if err != nil {
		return Config{}, err
	}

	config := Config{
		Port:             envOrDefault("PORT", "8080"),
		DatabaseURL:      strings.TrimSpace(os.Getenv("DATABASE_URL")),
		FrontendOrigin:   strings.TrimRight(envOrDefault("FRONTEND_ORIGIN", "http://localhost:3000"), "/"),
		JWTAccessSecret:  os.Getenv("JWT_ACCESS_SECRET"),
		JWTRefreshSecret: os.Getenv("JWT_REFRESH_SECRET"),
		CookieSecure:     cookieSecure,
		CookieDomain:     strings.TrimSpace(os.Getenv("COOKIE_DOMAIN")),
	}

	if config.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	frontendURL, err := url.ParseRequestURI(config.FrontendOrigin)
	validScheme := err == nil && (frontendURL.Scheme == "http" || frontendURL.Scheme == "https")
	if !validScheme || frontendURL.Host == "" || frontendURL.Path != "" || frontendURL.RawQuery != "" || frontendURL.User != nil {
		return Config{}, errors.New("FRONTEND_ORIGIN must be one explicit http(s) origin")
	}
	if config.JWTAccessSecret == "" || config.JWTRefreshSecret == "" {
		return Config{}, errors.New("JWT_ACCESS_SECRET and JWT_REFRESH_SECRET are required")
	}

	return config, nil
}

func envOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func boolEnv(key string, fallback bool) (bool, error) {
	rawValue := strings.TrimSpace(os.Getenv(key))
	if rawValue == "" {
		return fallback, nil
	}
	value, err := strconv.ParseBool(rawValue)
	if err != nil {
		return false, fmt.Errorf("%s must be true or false: %w", key, err)
	}
	return value, nil
}
