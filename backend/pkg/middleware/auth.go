// Package middleware chứa các HTTP middleware cho Fiber.
// File này cung cấp:
//   - AuthMiddleware: kiểm tra accessToken cookie, verify JWT, lưu userID vào Locals
//   - UserID: helper lấy userID từ context sau khi middleware đã xác thực
// Nếu token không hợp lệ hoặc thiếu, trả về 401 Unauthorized.
package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/hmduongdl/ChiaDeu/pkg/auth"
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
