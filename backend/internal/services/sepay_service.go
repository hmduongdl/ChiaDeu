package services

// SepayService xử lý các nghiệp vụ liên quan đến SePay:
// - Đăng ký webhook theo dõi biến động số dư tài khoản ngân hàng
// - Xác thực chữ ký webhook từ SePay
// SePay là cổng đồng bộ giao dịch ngân hàng chính của hệ thống.
// Đăng ký cá nhân dễ dàng, webhook realtime, có gói miễn phí.
type SepayService struct {
	apiKey string // API key của tài khoản SePay
}

// NewSepayService khởi tạo service với API key SePay.
func NewSepayService(apiKey string) *SepayService {
	return &SepayService{apiKey: apiKey}
}

// RegisterWebhook đăng ký theo dõi biến động số dư cho một tài khoản ngân hàng.
// Gọi SePay API để thiết lập webhook, mỗi khi tài khoản có giao dịch mới
// SePay sẽ gửi POST request đến endpoint /api/webhooks/sepay của hệ thống.
// Tham số:
//   - bankAccountNo: số tài khoản ngân hàng cần theo dõi
//   - bankCode: mã ngân hàng (VD: MBBank, VCB, ACB...)
func (s *SepayService) RegisterWebhook(bankAccountNo, bankCode string) error {
	// TODO: Gọi SePay API POST /api/v1/webhooks/register
	// TODO: Truyền bank_account_no, bank_code, webhook_url (đến endpoint của mình)
	return nil
}

// VerifyWebhookSignature xác thực chữ ký của webhook gửi từ SePay.
// SePay ký mỗi webhook bằng HMAC-SHA256 với API key.
// Việc xác thực này đảm bảo webhook đến từ SePay thật, không phải giả mạo.
// Tham số:
//   - signature: chữ ký từ header X-Sepay-Signature
//   - payload: body của webhook (raw bytes)
// Trả về true nếu chữ ký hợp lệ.
func (s *SepayService) VerifyWebhookSignature(signature string, payload []byte) bool {
	// TODO: Tính HMAC-SHA256(payload, apiKey)
	// TODO: So sánh với signature từ header
	return false
}