// Package main là điểm khởi chạy chính của backend Go.
// Server sử dụng Fiber framework để xử lý REST API, webhook từ SePay/PayOS/MoMo,
// và gọi thuật toán Minimize Cash Flow viết bằng C++ thông qua cgo.
package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/yourusername/cash-flow-minimizer/internal/config"
	"github.com/yourusername/cash-flow-minimizer/internal/database"
	"github.com/yourusername/cash-flow-minimizer/internal/handlers"
	"github.com/yourusername/cash-flow-minimizer/internal/middleware"
)

func main() {
	// Bước 1: Nạp cấu hình từ biến môi trường
	cfg := config.Load()

	// Bước 2: Kết nối PostgreSQL
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Lỗi kết nối database:", err)
	}
	defer db.Close()

	// Bước 3: Chạy migration để tạo/đồng bộ schema
	if err := database.RunMigrations(db); err != nil {
		log.Fatal("Lỗi chạy migration:", err)
	}

	// Bước 4: Khởi tạo Fiber app
	app := fiber.New(fiber.Config{
		AppName: "Cash Flow Minimizer",
	})

	// Middleware toàn cục
	app.Use(logger.New()) // Ghi log mỗi request
	app.Use(cors.New())   // Cho phép frontend Next.js gọi API cross-origin

	// Nhóm các route /api
	api := app.Group("/api")

	// --- Auth: Liên kết tài khoản ngân hàng với SePay ---
	api.Post("/auth/link-bank", handlers.LinkBank(db))

	// --- Webhooks: Nhận callback từ các cổng thanh toán ---
	api.Post("/webhooks/sepay", handlers.SepayWebhook(db)) // Webhook biến động số dư từ SePay
	api.Post("/webhooks/payos", handlers.PayOSWebhook(db)) // Webhook xác nhận thanh toán từ PayOS
	api.Post("/webhooks/momo", handlers.MoMoWebhook(db))   // Webhook xác nhận thanh toán từ MoMo

	// --- Giao dịch ngân hàng: Danh sách giao dịch chưa dùng ---
	api.Get("/transactions", middleware.AuthRequired, handlers.GetTransactions(db))

	// --- Nhóm: CRUD nhóm chi tiêu ---
	api.Post("/groups", middleware.AuthRequired, handlers.CreateGroup(db))                // Tạo nhóm, sinh share_code
	api.Post("/groups/join/:shareCode", middleware.AuthRequired, handlers.JoinGroup(db)) // Tham gia nhóm qua mã mời
	api.Get("/groups/:id", middleware.AuthRequired, handlers.GetGroup(db))                // Xem thông tin nhóm + thành viên

	// --- Khoản chi: Tạo expense và xem balance ---
	api.Post("/groups/:id/expenses", middleware.AuthRequired, handlers.CreateExpense(db)) // Tạo khoản chi, chia tiền
	api.Get("/groups/:id/balances", middleware.AuthRequired, handlers.GetBalances(db))    // Xem net balance từng thành viên

	// --- Thanh toán: Tính toán settlement và sinh QR ---
	api.Post("/groups/:id/settle", middleware.AuthRequired, handlers.SettleGroup(db))           // Gọi thuật toán Minimize Cash Flow
	api.Post("/settlements/:id/qr", middleware.AuthRequired, handlers.GenerateQR(db))           // Sinh QR code / payment link
	api.Get("/settlements/:id/status", middleware.AuthRequired, handlers.GetSettlementStatus(db)) // Kiểm tra trạng thái thanh toán

	// Lấy port từ biến môi trường, mặc định 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server khởi động tại cổng %s", port)
	log.Fatal(app.Listen(":" + port))
}