// ============================================================================
// CreateGroupModal — Modal tạo nhóm mới
// ============================================================================
// Overlay modal với input nhập tên nhóm và 2 nút Huỷ / Tạo nhóm.
// Hiển thị khi isOpen = true, ẩn khi isOpen = false.
// Khi tạo xong: gọi onCreate(name) và tự động đóng modal.

'use client'

import { useState } from 'react'
import Button from '@/components/ui/Button'

interface CreateGroupModalProps {
  isOpen: boolean                        // Trạng thái hiển thị modal
  onClose: () => void                    // Callback khi đóng modal
  onCreate: (name: string) => void       // Callback khi nhấn "Tạo nhóm"
}

export default function CreateGroupModal({ isOpen, onClose, onCreate }: CreateGroupModalProps) {
  const [name, setName] = useState('')   // Tên nhóm người dùng nhập

  // Không render gì nếu modal đang đóng
  if (!isOpen) return null

  return (
    // Overlay mờ đen phủ toàn màn hình
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      {/* Nội dung modal */}
      <div className="bg-white rounded-xl p-6 w-full max-w-md">
        <h2 className="text-xl font-semibold mb-4">Tạo nhóm mới</h2>

        {/* Input tên nhóm */}
        <input
          type="text"
          placeholder="Tên nhóm (VD: Đà Lạt trip)"
          value={name}
          onChange={(e) => setName(e.target.value)}
          className="w-full border rounded-lg px-3 py-2 mb-4 focus:outline-none focus:ring-2 focus:ring-primary-500"
        />

        {/* Nút hành động */}
        <div className="flex gap-3 justify-end">
          <Button variant="secondary" onClick={onClose}>Huỷ</Button>
          <Button onClick={() => { onCreate(name); setName(''); onClose() }}>
            Tạo nhóm
          </Button>
        </div>
      </div>
    </div>
  )
}