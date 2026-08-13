// Package main khởi tạo và chạy server API của ứng dụng Chia Đều.
// File này chịu trách nhiệm:
//   - Nạp cấu hình từ biến môi trường
//   - Kết nối database PostgreSQL
//   - Khởi tạo JWT token manager và auth service
//   - Đăng ký các route API (health, auth, nghiệp vụ, webhook)
//   - Cấu hình CORS và logging middleware
package main

import (
	"context"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/hmduongdl/ChiaDeu/internal/auth"
	"github.com/hmduongdl/ChiaDeu/internal/config"
	"github.com/hmduongdl/ChiaDeu/internal/handlers"
	authmiddleware "github.com/hmduongdl/ChiaDeu/internal/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

type appDependencies struct {
	frontendOrigin string
	authHandler    *handlers.AuthHandler
	authMiddleware fiber.Handler
}

func main() {
	_ = godotenv.Load()

	appConfig, err := config.Load()
	if err != nil {
		log.Fatalf("load configuration: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, appConfig.DatabaseURL)
	if err != nil {
		log.Fatalf("create database pool: %v", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("connect to database: %v", err)
	}

	tokens, err := auth.NewTokenManager(appConfig.JWTAccessSecret, appConfig.JWTRefreshSecret)
	if err != nil {
		log.Fatalf("configure JWT: %v", err)
	}
	authService := auth.NewService(auth.NewPostgresUserStore(pool))
	authHandler := handlers.NewAuthHandler(authService, tokens, handlers.CookieConfig{
		Secure: appConfig.CookieSecure,
		Domain: appConfig.CookieDomain,
	})

	app := newApp(appDependencies{
		frontendOrigin: appConfig.FrontendOrigin,
		authHandler:    authHandler,
		authMiddleware: authmiddleware.AuthMiddleware(tokens),
	})

	log.Fatal(app.Listen(":" + appConfig.Port))
}

func newApp(dependencies appDependencies) *fiber.App {
	app := fiber.New()

	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     dependencies.frontendOrigin,
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept",
		AllowCredentials: true,
	}))

	api := app.Group("/api")
	registerHealthRoute(api)
	if dependencies.authHandler != nil && dependencies.authMiddleware != nil {
		registerAuthRoutes(api, dependencies.authHandler, dependencies.authMiddleware)
		registerApplicationRoutes(api, dependencies.authMiddleware)
	}
	registerWebhookRoutes(api)

	// Keep the existing deployment-prefixed health URL while auth stays canonical at /api/auth.
	legacyAPI := app.Group("/api/backend")
	registerHealthRoute(legacyAPI)
	if dependencies.authMiddleware != nil {
		registerApplicationRoutes(legacyAPI, dependencies.authMiddleware)
	}
	registerWebhookRoutes(legacyAPI)

	return app
}

func registerHealthRoute(api fiber.Router) {
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})
}

func registerAuthRoutes(api fiber.Router, handler *handlers.AuthHandler, requireAuth fiber.Handler) {
	authRoutes := api.Group("/auth")
	authRoutes.Post("/register", handler.Register)
	authRoutes.Post("/login", handler.Login)
	authRoutes.Post("/refresh", handler.Refresh)
	authRoutes.Post("/logout", handler.Logout)
	authRoutes.Get("/me", requireAuth, handler.Me)
}

func registerApplicationRoutes(api fiber.Router, requireAuth fiber.Handler) {
	protected := api.Group("", requireAuth)

	protected.Post("/auth/link-bank", notImplemented)
	protected.Get("/transactions", notImplemented)
	protected.Post("/groups", notImplemented)
	protected.Post("/groups/join/:shareCode", notImplemented)
	protected.Get("/groups/:id", notImplemented)
	protected.Post("/groups/:id/expenses", notImplemented)
	protected.Get("/groups/:id/balances", notImplemented)
	protected.Post("/groups/:id/settle", notImplemented)
	protected.Post("/settlements/:id/qr", notImplemented)
	protected.Get("/settlements/:id/status", notImplemented)
}

func registerWebhookRoutes(api fiber.Router) {
	api.Post("/webhooks/sepay", notImplemented)
	api.Post("/webhooks/payos", notImplemented)
	api.Post("/webhooks/momo", notImplemented)
}

func notImplemented(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "not implemented"})
}
