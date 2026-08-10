// Package database quản lý kết nối PostgreSQL và chạy migration tự động khi server khởi động.
// Migration được đọc từ thư mục migrations/ và chạy theo thứ tự alphabet của tên file.
package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	_ "github.com/lib/pq" // PostgreSQL driver cho database/sql
)

// Connect mở kết nối đến PostgreSQL và kiểm tra kết nối bằng ping.
// databaseURL có dạng: postgres://user:password@host:port/dbname?sslmode=disable
func Connect(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("không thể mở kết nối database: %w", err)
	}

	// Kiểm tra kết nối thực sự hoạt động
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("không thể ping database: %w", err)
	}

	// Giới hạn số kết nối đồng thời để tránh quá tải PostgreSQL
	db.SetMaxOpenConns(25) // Tối đa 25 kết nối mở
	db.SetMaxIdleConns(5)  // Giữ sẵn 5 kết nối idle trong pool

	return db, nil
}

// RunMigrations đọc tất cả file .sql trong thư mục migrations/ và thực thi theo thứ tự.
// Mỗi file migration dùng cú pháp IF NOT EXISTS nên có thể chạy lại an toàn (idempotent).
func RunMigrations(db *sql.DB) error {
	// Lấy đường dẫn tuyệt đối đến thư mục migrations/ dựa vào vị trí file hiện tại
	_, filename, _, _ := runtime.Caller(0)
	migrationsDir := filepath.Join(filepath.Dir(filename), "../../migrations")

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("không thể đọc thư mục migrations: %w", err)
	}

	for _, entry := range entries {
		// Bỏ qua thư mục và file không phải .sql
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".sql" {
			continue
		}

		content, err := os.ReadFile(filepath.Join(migrationsDir, entry.Name()))
		if err != nil {
			return fmt.Errorf("không thể đọc file migration %s: %w", entry.Name(), err)
		}

		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("không thể thực thi migration %s: %w", entry.Name(), err)
		}

		fmt.Printf("Đã chạy migration: %s\n", entry.Name())
	}

	return nil
}