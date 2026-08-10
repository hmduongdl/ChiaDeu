// ============================================================================
// TransactionList — Danh sách giao dịch ngân hàng đồng bộ từ SePay
// ============================================================================
// Hiển thị các giao dịch chưa dùng (is_used = false).
// Mỗi dòng: mô tả, ngày, số tiền (xanh = IN, đỏ = OUT).
// Có thể click để chọn giao dịch và tạo Expense từ đó.

import type { BankTransaction } from '@/lib/types'

interface TransactionListProps {
  transactions: BankTransaction[]                // Danh sách giao dịch
  onSelect?: (tx: BankTransaction) => void       // Callback khi chọn giao dịch
}

export default function TransactionList({ transactions, onSelect }: TransactionListProps) {
  // Trạng thái rỗng
  if (transactions.length === 0) {
    return (
      <div className="text-center py-8 text-gray-500">
        Không có giao dịch nào
      </div>
    )
  }

  return (
    <div className="space-y-2">
      {transactions.map((tx) => (
        <div
          key={tx.id}
          className="bg-white rounded-lg border p-3 hover:shadow-sm transition cursor-pointer"
          onClick={() => onSelect?.(tx)}
        >
          <div className="flex justify-between items-start">
            {/* Bên trái: mô tả + ngày giao dịch */}
            <div>
              <p className="text-sm text-gray-600">{tx.description}</p>
              <p className="text-xs text-gray-400">
                {new Date(tx.transaction_time).toLocaleDateString('vi-VN')}
              </p>
            </div>

            {/* Bên phải: số tiền (IN = xanh, OUT = đỏ) */}
            <p className={`font-semibold ${tx.transaction_type === 'IN' ? 'text-green-600' : 'text-red-600'}`}>
              {tx.transaction_type === 'IN' ? '+' : '-'}
              {tx.amount.toLocaleString('vi-VN')}đ
            </p>
          </div>
        </div>
      ))}
    </div>
  )
}