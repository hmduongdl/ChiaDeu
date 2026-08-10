// Package middleware chứa các HTTP middleware dùng chung cho toàn bộ API.
// AuthRequired là middleware xác thực - hiện tại là stub, sẽ implement JWT sau.
package middleware

import "github.com/gofiber/fiber/v2"

// AuthRequired kiểm tra token JWT trong header Authorization.
// Nếu token hợp lệ, trích xuất userID và lưu vào context (c.Locals).
// TODO: Implement JWT validation thực sự khi tích hợp auth.
func AuthRequired(c *fiber.Ctx) error {
	// Lấy token từ header Authorization: Bearer <token>
	// token := c.Get("Authorization")
	// userID, err := validateToken(token)
	// if err != nil {
	//     return c.Status(401).JSON(fiber.Map{"error": "Chưa đăng nhập"})
	// }
	// c.Locals("userID", userID)
	return c.Next()
}