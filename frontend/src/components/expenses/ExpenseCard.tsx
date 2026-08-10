// ============================================================================
// ExpenseCard — Component hiển thị một khoản chi
// ============================================================================
// Hiển thị mô tả, cách chia (split_type), và số tiền.
// Dùng trong danh sách expenses của một nhóm.

import type { Expense } from '@/lib/types'

interface ExpenseCardProps {
  expense: Expense // Dữ liệu khoản chi
}

export default function ExpenseCard({ expense }: ExpenseCardProps) {
  return (
    <div className="bg-white rounded-lg border p-3">
      <div className="flex justify-between items-start">
        {/* Bên trái: mô tả + cách chia */}
        <div>
          <p className="font-medium">{expense.description}</p>
          <p className="text-sm text-gray-500">
            {expense.split_type === 'EQUAL' ? 'Chia đều' : expense.split_type}
          </p>
        </div>

        {/* Bên phải: số tiền (định dạng VND) */}
        <p className="font-semibold text-primary-600">
          {expense.amount.toLocaleString('vi-VN')}đ
        </p>
      </div>
    </div>
  )
}