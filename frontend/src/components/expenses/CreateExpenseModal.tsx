// ============================================================================
// CreateExpenseModal — Modal thêm khoản chi mới
// ============================================================================
// Cho phép người dùng nhập mô tả và số tiền của khoản chi.
// Hiện tại là phiên bản đơn giản (chưa có chọn thành viên và split).
// TODO: Thêm chọn thành viên tham gia chia và loại split (EQUAL/PERCENT/CUSTOM).

'use client'

import { useState } from 'react'
import Button from '@/components/ui/Button'

interface CreateExpenseModalProps {
  isOpen: boolean                               // Trạng thái hiển thị
  onClose: () => void                           // Callback đóng modal
  onCreate: (data: any) => void                 // Callback tạo expense
}

export default function CreateExpenseModal({ isOpen, onClose, onCreate }: CreateExpenseModalProps) {
  const [description, setDescription] = useState('') // Mô tả khoản chi
  const [amount, setAmount] = useState('')           // Số tiền (dạng string để input)

  if (!isOpen) return null

  return (
    // Overlay mờ đen
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-white rounded-xl p-6 w-full max-w-md">
        <h2 className="text-xl font-semibold mb-4">Thêm khoản chi</h2>

        {/* Input mô tả */}
        <input
          type="text"
          placeholder="Mô tả (VD: Ăn tối, Xăng xe...)"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          className="w-full border rounded-lg px-3 py-2 mb-3 focus:outline-none focus:ring-2 focus:ring-primary-500"
        />

        {/* Input số tiền */}
        <input
          type="number"
          placeholder="Số tiền (VNĐ)"
          value={amount}
          onChange={(e) => setAmount(e.target.value)}
          className="w-full border rounded-lg px-3 py-2 mb-4 focus:outline-none focus:ring-2 focus:ring-primary-500"
        />

        {/* Nút hành động */}
        <div className="flex gap-3 justify-end">
          <Button variant="secondary" onClick={onClose}>Huỷ</Button>
          <Button onClick={() => {
            onCreate({ description, amount: parseFloat(amount) })
            setDescription(''); setAmount('') // Reset form
            onClose()
          }}>
            Thêm
          </Button>
        </div>
      </div>
    </div>
  )
}