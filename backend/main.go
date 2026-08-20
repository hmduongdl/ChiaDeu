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
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/hmduongdl/ChiaDeu/internal/auth"
	"github.com/hmduongdl/ChiaDeu/internal/config"
	"github.com/hmduongdl/ChiaDeu/internal/expenses"
	"github.com/hmduongdl/ChiaDeu/internal/groups"
	"github.com/hmduongdl/ChiaDeu/internal/handlers"
	authmiddleware "github.com/hmduongdl/ChiaDeu/internal/middleware"
	"github.com/hmduongdl/ChiaDeu/internal/settlements"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

type appDependencies struct {
	frontendOrigin   string
	rateLimitMax     int
	authRateLimitMax int
	authHandler      *handlers.AuthHandler
	authMiddleware   fiber.Handler
	appHandler       *handlers.AppHandler
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
	sessionStore := auth.NewPostgresSessionStore(pool)
	authService := auth.NewServiceWithSessions(auth.NewPostgresUserStore(pool), tokens, sessionStore)
	authHandler := handlers.NewAuthHandler(authService, tokens, handlers.CookieConfig{
		Secure: appConfig.CookieSecure,
		Domain: appConfig.CookieDomain,
	})

	appHandler := handlers.NewAppHandler(
		groups.NewService(groups.NewPostgresStore(pool)),
		expenses.NewService(expenses.NewPostgresStore(pool)),
		settlements.NewPostgresStore(pool),
	)

	app := newApp(appDependencies{
		frontendOrigin:   appConfig.FrontendOrigin,
		rateLimitMax:     appConfig.RateLimitMax,
		authRateLimitMax: appConfig.AuthRateLimitMax,
		authHandler:      authHandler,
		authMiddleware:   authmiddleware.AuthMiddleware(tokens),
		appHandler:       appHandler,
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
	if dependencies.rateLimitMax > 0 {
		api.Use(limiter.New(limiter.Config{Max: dependencies.rateLimitMax, Expiration: time.Minute}))
	}
	registerHealthRoute(api)
	if dependencies.authHandler != nil && dependencies.authMiddleware != nil {
		registerAuthRoutes(api, dependencies.authHandler, dependencies.authMiddleware, dependencies.authRateLimitMax)
	}
	if dependencies.appHandler != nil && dependencies.authMiddleware != nil {
		registerApplicationRoutes(api, dependencies.authMiddleware, dependencies.appHandler)
	}
	registerWebhookRoutes(api)

	// Keep the existing deployment-prefixed health/business URL while auth stays canonical at /api/auth.
	legacyAPI := app.Group("/api/backend")
	registerHealthRoute(legacyAPI)
	if dependencies.appHandler != nil && dependencies.authMiddleware != nil {
		registerApplicationRoutes(legacyAPI, dependencies.authMiddleware, dependencies.appHandler)
	}
	registerWebhookRoutes(legacyAPI)

	return app
}

func registerHealthRoute(api fiber.Router) {
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})
}

func registerAuthRoutes(api fiber.Router, handler *handlers.AuthHandler, requireAuth fiber.Handler, rateLimitMax int) {
	authRoutes := api.Group("/auth")
	if rateLimitMax > 0 {
		authRoutes.Use(limiter.New(limiter.Config{Max: rateLimitMax, Expiration: time.Minute}))
	}
	authRoutes.Post("/register", handler.Register)
	authRoutes.Post("/login", handler.Login)
	authRoutes.Post("/refresh", handler.Refresh)
	authRoutes.Post("/logout", handler.Logout)
	authRoutes.Get("/me", requireAuth, handler.Me)
}

func registerApplicationRoutes(api fiber.Router, requireAuth fiber.Handler, app *handlers.AppHandler) {
	groupsRoutes := api.Group("/groups", requireAuth)
	groupsRoutes.Post("/", app.CreateGroup)
	groupsRoutes.Post("/join/:shareCode", app.JoinGroup)
	groupsRoutes.Get("/:groupId", app.GetGroup)
	groupsRoutes.Post("/:groupId/expenses", app.CreateExpense)
	groupsRoutes.Patch("/:groupId/expenses/:expenseId", app.UpdateExpense)
	groupsRoutes.Get("/:groupId/balances", app.Balances)
	groupsRoutes.Post("/:groupId/settlement-batches", app.CloseBatch)
	groupsRoutes.Get("/:groupId/settlement-batches/:batchId", app.GetBatch)

	settlementRoutes := api.Group("/settlements", requireAuth)
	settlementRoutes.Post("/:settlementId/mark-sent", app.MarkSent)
	settlementRoutes.Post("/:settlementId/confirm", app.Confirm)
	settlementRoutes.Post("/:settlementId/reject", app.Reject)
}

func registerWebhookRoutes(api fiber.Router) {
	api.Post("/webhooks/sepay", notImplemented)
	api.Post("/webhooks/payos", notImplemented)
	api.Post("/webhooks/momo", notImplemented)
}

func notImplemented(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "not implemented"})
}
