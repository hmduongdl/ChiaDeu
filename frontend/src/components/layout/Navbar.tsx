// ============================================================================
// Navbar — Thanh điều hướng chính của ứng dụng
// ============================================================================
// Hiển thị logo "CashFlow" bên trái và các link điều hướng bên phải.
// Fixed ở đỉnh trang với shadow nhẹ và border bottom.

export default function Navbar() {
  return (
    <nav className="bg-white shadow-sm border-b">
      <div className="max-w-4xl mx-auto px-4 h-14 flex items-center justify-between">
        {/* Logo / Tên app */}
        <a href="/" className="font-bold text-primary-600">
          CashFlow
        </a>

        {/* Menu điều hướng */}
        <div className="flex gap-4">
          <a href="/groups" className="text-sm text-gray-600 hover:text-primary-600">
            Nhóm
          </a>
          <a href="/transactions" className="text-sm text-gray-600 hover:text-primary-600">
            Giao dịch
          </a>
        </div>
      </div>
    </nav>
  )
}