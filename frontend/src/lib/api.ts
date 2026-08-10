// ============================================================================
// API Client — Giao tiếp với Go backend qua REST API
// ============================================================================
// Tất cả request được proxy qua Next.js (xem next.config.js) để tránh CORS.
// Mỗi hàm trả về Promise<T> với T là kiểu dữ liệu tương ứng.
// Khi có lỗi HTTP, ném Error với status code.

const API_BASE = '/api' // Proxy đến Go backend (localhost:8080)

// Hàm helper: gửi request và parse JSON response
async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
    ...options,
  })
  if (!res.ok) {
    throw new Error(`Lỗi API: ${res.status}`)
  }
  return res.json()
}

export const api = {
  // ==========================================================================
  // Groups — Quản lý nhóm chi tiêu
  // ==========================================================================

  // Tạo nhóm mới, trả về id và share_code
  createGroup: (name: string) =>
    request<{ id: string; share_code: string }>('/groups', {
      method: 'POST',
      body: JSON.stringify({ name }),
    }),

  // Tham gia nhóm bằng mã mời
  joinGroup: (shareCode: string) =>
    request<{ id: string }>(`/groups/join/${shareCode}`, { method: 'POST' }),

  // Lấy thông tin nhóm + danh sách thành viên
  getGroup: (id: string) =>
    request<{ group: any }>(`/groups/${id}`),

  // ==========================================================================
  // Expenses — Quản lý khoản chi
  // ==========================================================================

  // Tạo khoản chi mới trong nhóm
  createExpense: (groupId: string, data: any) =>
    request<{ id: string }>(`/groups/${groupId}/expenses`, {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  // Lấy net balance của từng thành viên trong nhóm
  getBalances: (groupId: string) =>
    request<{ balances: any[] }>(`/groups/${groupId}/balances`),

  // ==========================================================================
  // Settlements — Thanh toán tối ưu
  // ==========================================================================

  // Gọi thuật toán Minimize Cash Flow, trả về danh sách settlements
  settleGroup: (groupId: string) =>
    request<{ settlements: any[] }>(`/groups/${groupId}/settle`, {
      method: 'POST',
    }),

  // Sinh QR code / payment link cho một settlement
  generateQR: (settlementId: string) =>
    request<{ qr_code: string }>(`/settlements/${settlementId}/qr`, {
      method: 'POST',
    }),

  // ==========================================================================
  // Transactions — Giao dịch ngân hàng
  // ==========================================================================

  // Lấy danh sách giao dịch ngân hàng (mặc định chỉ lấy chưa dùng)
  getTransactions: (unusedOnly = true) =>
    request<{ transactions: any[] }>(`/transactions?unused=${unusedOnly}`),

  // ==========================================================================
  // Auth — Xác thực & liên kết ngân hàng
  // ==========================================================================

  // Liên kết tài khoản ngân hàng với SePay
  linkBank: (data: any) =>
    request<{ message: string }>('/auth/link-bank', {
      method: 'POST',
      body: JSON.stringify(data),
    }),
}