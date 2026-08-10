// Package handlers chứa tất cả HTTP handler cho các endpoint REST API.
// Mỗi handler nhận *sql.DB qua closure pattern để inject database dependency.
// Các handler hiện tại là stub với TODO - sẽ implement logic thật ở bước tiếp theo.
package handlers

import (
	"database/sql"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/cash-flow-minimizer/internal/models"
)

// ============================================================================
// Auth Handlers
// ============================================================================

// LinkBank xử lý yêu cầu liên kết tài khoản ngân hàng với SePay.
// Endpoint: POST /api/auth/link-bank
// Input: user_id, bank_account_no, bank_code
// Luồng xử lý:
//  1. Nhận thông tin tài khoản ngân hàng từ client
//  2. Gọi SePay API để đăng ký webhook theo dõi biến động số dư
//  3. Cập nhật bank_account_no, bank_code, sepay_account_id vào bảng users
func LinkBank(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input struct {
			UserID        string `json:"user_id"`
			BankAccountNo string `json:"bank_account_no"`
			BankCode      string `json:"bank_code"`
		}
		if err := c.BodyParser(&input); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Dữ liệu không hợp lệ"})
		}

		// TODO: Gọi SePay API để đăng ký webhook cho tài khoản này
		// TODO: Cập nhật thông tin ngân hàng vào bảng users

		return c.Status(200).JSON(fiber.Map{"message": "Liên kết ngân hàng thành công"})
	}
}

// ============================================================================
// Webhook Handlers - Nhận callback từ các cổng thanh toán
// ============================================================================

// SepayWebhook nhận webhook báo biến động số dư từ SePay.
// Endpoint: POST /api/webhooks/sepay
// Mỗi khi tài khoản ngân hàng có giao dịch (nhận/chuyển tiền),
// SePay sẽ gửi POST request đến endpoint này với thông tin giao dịch.
// Luồng xử lý:
//  1. Xác thực chữ ký webhook từ SePay để đảm bảo request hợp lệ
//  2. Parse dữ liệu giao dịch (số tiền, nội dung, loại giao dịch...)
//  3. Lưu vào bảng bank_transactions với is_used = false
func SepayWebhook(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var tx models.BankTransaction
		if err := c.BodyParser(&tx); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Payload webhook không hợp lệ"})
		}

		// TODO: Xác thực chữ ký webhook SePay
		// TODO: Lưu giao dịch vào bảng bank_transactions với is_used = false
		// TODO: Tránh trùng lặp bằng cách kiểm tra sepay_transaction_id unique

		return c.Status(200).JSON(fiber.Map{"message": "Đã nhận giao dịch"})
	}
}

// PayOSWebhook nhận webhook xác nhận thanh toán từ PayOS.
// Endpoint: POST /api/webhooks/payos
// Khi người dùng quét QR PayOS và thanh toán thành công, PayOS gửi callback về đây.
// Luồng xử lý:
//  1. Xác thực chữ ký webhook PayOS
//  2. Tìm settlement tương ứng với payment link
//  3. Cập nhật settlement.status = PAID, paid_at = now()
func PayOSWebhook(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// TODO: Xác thực chữ ký PayOS
		// TODO: Tìm settlement theo payment link ID
		// TODO: Cập nhật trạng thái settlement sang PAID
		return c.Status(200).JSON(fiber.Map{"message": "Đã nhận webhook PayOS"})
	}
}

// MoMoWebhook nhận webhook xác nhận thanh toán từ MoMo (sandbox).
// Endpoint: POST /api/webhooks/momo
// Tương tự PayOS, nhưng dành cho phương thức thanh toán MoMo.
// Luồng xử lý:
//  1. Xác thực chữ ký MoMo
//  2. Tìm settlement tương ứng
//  3. Cập nhật settlement.status = PAID
func MoMoWebhook(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// TODO: Xác thực chữ ký MoMo
		// TODO: Tìm settlement và cập nhật trạng thái PAID
		return c.Status(200).JSON(fiber.Map{"message": "Đã nhận webhook MoMo"})
	}
}

// ============================================================================
// Transaction Handlers - Quản lý giao dịch ngân hàng
// ============================================================================

// GetTransactions trả về danh sách giao dịch ngân hàng chưa được gán vào Expense.
// Endpoint: GET /api/transactions?unused=true
// Mặc định chỉ trả về giao dịch có is_used = false.
// Người dùng sẽ chọn từ danh sách này để tạo Expense.
func GetTransactions(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// TODO: Query bank_transactions WHERE user_id = ? AND is_used = false
		// TODO: Sắp xếp theo transaction_time giảm dần
		return c.Status(200).JSON(fiber.Map{"transactions": []models.BankTransaction{}})
	}
}

// ============================================================================
// Group Handlers - Quản lý nhóm chi tiêu
// ============================================================================

// CreateGroup tạo nhóm mới và sinh mã mời share_code ngẫu nhiên.
// Endpoint: POST /api/groups
// Input: { "name": "Đà Lạt trip" }
// Luồng xử lý:
//  1. Sinh share_code ngẫu nhiên (6-10 ký tự, unique)
//  2. INSERT vào bảng groups
//  3. Tự động thêm người tạo vào group_members
//  4. Trả về group_id và share_code
func CreateGroup(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// TODO: Sinh share_code ngẫu nhiên, đảm bảo unique
		// TODO: INSERT groups + INSERT group_members (người tạo là member đầu tiên)
		return c.Status(201).JSON(fiber.Map{"message": "Đã tạo nhóm"})
	}
}

// JoinGroup cho phép người dùng tham gia nhóm bằng mã mời share_code.
// Endpoint: POST /api/groups/join/:shareCode
// Luồng xử lý:
//  1. Tìm group theo share_code
//  2. Kiểm tra user đã là thành viên chưa
//  3. Thêm user vào group_members
func JoinGroup(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		shareCode := c.Params("shareCode")
		// TODO: Tìm group theo share_code
		// TODO: Kiểm tra user chưa là thành viên
		// TODO: INSERT vào group_members
		return c.Status(200).JSON(fiber.Map{"message": "Đã tham gia nhóm", "share_code": shareCode})
	}
}

// GetGroup trả về thông tin chi tiết của nhóm và danh sách thành viên.
// Endpoint: GET /api/groups/:id
// Trả về: group info + members[] + expenses[] + balances
func GetGroup(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// TODO: Query group + JOIN group_members + users để lấy danh sách thành viên
		return c.Status(200).JSON(fiber.Map{"group": nil})
	}
}

// ============================================================================
// Expense Handlers - Quản lý khoản chi
// ============================================================================

// CreateExpense tạo khoản chi mới trong nhóm và chia tiền cho các thành viên.
// Endpoint: POST /api/groups/:id/expenses
// Input: { paid_by, description, amount, split_type, splits[], source_transaction_id? }
// Luồng xử lý (dùng database transaction):
//  1. INSERT vào expenses
//  2. INSERT từng dòng vào expense_splits
//  3. Nếu có source_transaction_id, UPDATE bank_transactions SET is_used = true
//  4. Commit transaction
func CreateExpense(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// TODO: Bắt đầu DB transaction
		// TODO: INSERT expense
		// TODO: INSERT expense_splits (từng người một)
		// TODO: Nếu có source_transaction_id -> UPDATE bank_transactions.is_used = true
		// TODO: Commit transaction
		return c.Status(201).JSON(fiber.Map{"message": "Đã tạo khoản chi"})
	}
}

// GetBalances tính toán và trả về net balance của từng thành viên trong nhóm.
// Endpoint: GET /api/groups/:id/balances
// Công thức: net = tổng tiền đã trả hộ (paid_by) - tổng tiền phải chịu (share_amount)
// Kết quả dương (+) = người khác nợ mình (creditor)
// Kết quả âm (-) = mình nợ người khác (debtor)
func GetBalances(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// TODO: Tính net balance = SUM(paid_by) - SUM(share_amount) cho từng user trong nhóm
		return c.Status(200).JSON(fiber.Map{"balances": []interface{}{}})
	}
}

// ============================================================================
// Settlement Handlers - Thanh toán tối ưu
// ============================================================================

// SettleGroup gọi thuật toán Minimize Cash Flow để tính toán giao dịch thanh toán tối thiểu.
// Endpoint: POST /api/groups/:id/settle
// Luồng xử lý:
//  1. Tính net balance từng thành viên
//  2. Gọi C++ algorithm qua cgo bridge để tính settlements tối ưu
//  3. Lưu danh sách settlements vào database
//  4. Trả về danh sách các khoản cần thanh toán
func SettleGroup(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// TODO: Lấy balances từ DB
		// TODO: Gọi algorithm.MinimizeCashFlow(balances) để tính settlements
		// TODO: INSERT từng settlement vào bảng settlements
		return c.Status(200).JSON(fiber.Map{"settlements": []models.Settlement{}})
	}
}

// GenerateQR sinh QR code hoặc payment link cho một settlement.
// Endpoint: POST /api/settlements/:id/qr
// Input: { "method": "PAYOS_QR" | "MOMO" }
// Luồng xử lý:
//  1. Lấy thông tin settlement từ DB
//  2. Gọi PayOS API hoặc MoMo API để sinh QR/payment link
//  3. Lưu qr_code_data vào settlement
//  4. Trả về QR code hoặc payment link cho client
func GenerateQR(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// TODO: Lấy settlement từ DB
		// TODO: Gọi PayOS hoặc MoMo API để sinh QR
		// TODO: Lưu qr_code_data vào DB
		return c.Status(200).JSON(fiber.Map{"qr_code": ""})
	}
}

// GetSettlementStatus kiểm tra trạng thái thanh toán của một settlement.
// Endpoint: GET /api/settlements/:id/status
// Trả về trạng thái hiện tại: PENDING / PAID / CANCELLED
func GetSettlementStatus(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// TODO: Query settlement status từ DB
		return c.Status(200).JSON(fiber.Map{"status": "PENDING"})
	}
}