package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/hmduongdl/ChiaDeu/internal/auth"
)

const userIDLocal = "authenticatedUserID"

func AuthMiddleware(tokens *auth.TokenManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		rawToken := c.Cookies("accessToken")
		userID, err := tokens.VerifyAccessToken(rawToken)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
		}

		c.Locals(userIDLocal, userID)
		return c.Next()
	}
}

func UserID(c *fiber.Ctx) (string, bool) {
	userID, ok := c.Locals(userIDLocal).(string)
	return userID, ok && userID != ""
}
