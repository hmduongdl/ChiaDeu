// Package config quản lý tất cả cấu hình của ứng dụng thông qua biến môi trường.
// Các cấu hình bao gồm: kết nối database, API key cho SePay/PayOS/MoMo, và cổng server.
package config

import "os"

// Config chứa toàn bộ cấu hình cần thiết để chạy backend.
type Config struct {
	DatabaseURL string // Chuỗi kết nối PostgreSQL (VD: postgres://user:pass@host:5432/db?sslmode=disable)
	Port        string // Cổng HTTP server lắng nghe
	SepayAPIKey string // API key dùng để gọi SePay API (đăng ký webhook, xác thực)
	PayOSAPIKey string // API key dùng để gọi PayOS API (sinh QR, tạo payment link)
	MoMoAPIKey  string // API key dùng để gọi MoMo API (sandbox, sinh QR)
}

// Load đọc tất cả biến môi trường và trả về struct Config với giá trị mặc định hợp lý.
func Load() *Config {
	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://localhost:5432/cashflow?sslmode=disable"),
		Port:        getEnv("PORT", "8080"),
		SepayAPIKey: getEnv("SEPAY_API_KEY", ""),
		PayOSAPIKey: getEnv("PAYOS_API_KEY", ""),
		MoMoAPIKey:  getEnv("MOMO_API_KEY", ""),
	}
}

// getEnv trả về giá trị biến môi trường nếu có, ngược lại trả về fallback.
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}