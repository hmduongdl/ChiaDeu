// ============================================================================
// GroupCard — Component hiển thị thông tin một nhóm
// ============================================================================
// Hiển thị tên nhóm, mã mời (share_code), số lượng thành viên.
// Có thể click để chọn nhóm (onSelect callback).

import type { Group } from '@/lib/types'

interface GroupCardProps {
  group: Group                              // Dữ liệu nhóm
  onSelect?: (group: Group) => void         // Callback khi click vào card
}

export default function GroupCard({ group, onSelect }: GroupCardProps) {
  return (
    <div
      className="bg-white rounded-xl shadow-sm border p-4 hover:shadow-md transition cursor-pointer"
      onClick={() => onSelect?.(group)}
    >
      {/* Tên nhóm */}
      <h3 className="font-semibold text-lg">{group.name}</h3>

      {/* Mã mời (dạng monospace) */}
      <p className="text-sm text-gray-500 mt-1">
        Mã mời: <span className="font-mono">{group.share_code}</span>
      </p>

      {/* Số lượng thành viên */}
      <p className="text-sm text-gray-500">
        {group.members?.length ?? 0} thành viên
      </p>
    </div>
  )
}