// ============================================================================
// SettlementList — Danh sách giao dịch thanh toán tối ưu
// ============================================================================
// Kết quả sau khi chạy thuật toán Minimize Cash Flow.
// Mỗi dòng: from_user -> to_user, số tiền, phương thức thanh toán, trạng thái.
// Đây là danh sách các khoản cần thanh toán để "chốt sổ" nhóm.

import type { Settlement } from '@/lib/types'

interface SettlementListProps {
  settlements: Settlement[] // Danh sách settlements
}

export default function SettlementList({ settlements }: SettlementListProps) {
  // Trạng thái rỗng: chưa có settlement nào
  if (settlements.length === 0) {
    return (
      <div className="text-center py-8 text-gray-500">
        Chưa có khoản thanh toán nào
      </div>
    )
  }

  return (
    <div className="space-y-2">
      {settlements.map((s) => (
        <div
          key={s.id}
          className="bg-white rounded-lg border p-3 flex justify-between items-center"
        >
          {/* Bên trái: người trả -> người nhận + phương thức + trạng thái */}
          <div>
            <p className="text-sm">
              <span className="font-medium">{s.from_user}</span>
              {' → '}
              <span className="font-medium">{s.to_user}</span>
            </p>
            <p className="text-xs text-gray-500">
              {s.payment_method ?? 'Chưa chọn phương thức'} · {s.status === 'PAID' ? 'Đã trả' : s.status === 'CANCELLED' ? 'Đã huỷ' : 'Chờ thanh toán'}
            </p>
          </div>

          {/* Bên phải: số tiền (định dạng VND) */}
          <p className="font-semibold">
            {s.amount.toLocaleString('vi-VN')}đ
          </p>
        </div>
      ))}
    </div>
  )
}