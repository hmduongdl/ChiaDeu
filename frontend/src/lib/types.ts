// ============================================================================
// Cash Flow Minimizer — Các kiểu dữ liệu TypeScript dùng chung toàn app
// ============================================================================
// File này định nghĩa interfaces tương ứng với models bên Go backend.
// Mọi component và API call đều dùng chung các types từ đây.

// User: thông tin người dùng
export interface User {
  id: string // UUID
  name: string // Tên hiển thị
  phone?: string // SĐT (unique)
  email?: string // Email (unique)
  bank_account_no?: string // Số tài khoản liên kết SePay
  avatar_url?: string // URL ảnh đại diện
}

// Group: nhóm chi tiêu
export interface Group {
  id: string // UUID nhóm
  name: string // Tên nhóm (VD: "Đà Lạt trip")
  share_code: string // Mã mời ngắn
  created_by: string // UUID người tạo
  currency: string // Đơn vị tiền tệ, mặc định VND
  members: User[] // Danh sách thành viên
}

// BankTransaction: giao dịch ngân hàng đồng bộ từ SePay
export interface BankTransaction {
  id: string // UUID
  user_id: string // Chủ tài khoản
  amount: number // Số tiền
  transaction_type: 'IN' | 'OUT' // IN = nhận, OUT = chi
  description: string // Nội dung chuyển khoản
  is_used: boolean // Đã gán vào expense nào chưa
  transaction_time: string // Thời điểm giao dịch thực tế
}

// Expense: khoản chi cần chia trong nhóm
export interface Expense {
  id: string // UUID
  group_id: string // Thuộc nhóm nào
  source_transaction_id?: string // Giao dịch ngân hàng gốc (nếu có)
  paid_by: string // Người bỏ tiền trả
  description: string // Mô tả
  amount: number // Tổng tiền
  split_type: 'EQUAL' | 'PERCENT' | 'CUSTOM' // Cách chia
  splits: ExpenseSplit[] // Chi tiết chia cho từng người
}

// ExpenseSplit: phần tiền của một người trong khoản chi
export interface ExpenseSplit {
  user_id: string // UUID người phải chịu
  share_amount: number // Số tiền phải chịu
}

// Settlement: giao dịch thanh toán tối ưu (kết quả thuật toán)
export interface Settlement {
  id: string // UUID
  group_id: string // Thuộc nhóm nào
  from_user: string // Người phải trả (debtor)
  to_user: string // Người nhận (creditor)
  amount: number // Số tiền cần thanh toán
  status: 'PENDING' | 'PAID' | 'CANCELLED' // Trạng thái
  payment_method?: 'PAYOS_QR' | 'MOMO' | 'CASH' // Phương thức thanh toán
  qr_code_data?: string // Dữ liệu QR code / payment link
}

// Balance: số dư ròng của một thành viên trong nhóm
// amount > 0: được nhận tiền (creditor)
// amount < 0: phải trả tiền (debtor)
export interface Balance {
  user_id: string // UUID
  amount: number // Số dư ròng
}