package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	app := fiber.New()

	app.Use(logger.New())
	app.Use(cors.New())

	app.Get("/api/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	api := app.Group("/api")

	// Auth
	api.Post("/auth/link-bank", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "not implemented"})
	})

	// Webhooks
	api.Post("/webhooks/sepay", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "not implemented"})
	})
	api.Post("/webhooks/payos", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "not implemented"})
	})
	api.Post("/webhooks/momo", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "not implemented"})
	})

	// Transactions
	api.Get("/transactions", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "not implemented"})
	})

	// Groups
	api.Post("/groups", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "not implemented"})
	})
	api.Post("/groups/join/:shareCode", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "not implemented"})
	})
	api.Get("/groups/:id", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "not implemented"})
	})
	api.Post("/groups/:id/expenses", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "not implemented"})
	})
	api.Get("/groups/:id/balances", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "not implemented"})
	})
	api.Post("/groups/:id/settle", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "not implemented"})
	})

	// Settlements
	api.Post("/settlements/:id/qr", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "not implemented"})
	})
	api.Get("/settlements/:id/status", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "not implemented"})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Fatal(app.Listen(":" + port))
}