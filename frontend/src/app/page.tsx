// ============================================================================
// HomePage — Trang chủ của ứng dụng
// ============================================================================
// Hiển thị tiêu đề, mô tả ngắn và 2 nút điều hướng chính:
// - "Nhóm của tôi": đến danh sách nhóm
// - "Giao dịch ngân hàng": đến danh sách giao dịch đồng bộ từ SePay

export default function Home() {
  return (
    <div className="flex flex-col items-center justify-center min-h-screen p-8">
      {/* Tiêu đề chính */}
      <h1 className="text-4xl font-bold text-primary-600 mb-4">
        Cash Flow Minimizer
      </h1>

      {/* Mô tả ngắn */}
      <p className="text-lg text-gray-600 mb-8">
        Chia tiền nhóm thông minh — Tối thiểu hoá giao dịch thanh toán
      </p>

      {/* Nút điều hướng chính */}
      <div className="flex gap-4">
        <a
          href="/groups"
          className="px-6 py-3 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition"
        >
          Nhóm của tôi
        </a>
        <a
          href="/transactions"
          className="px-6 py-3 bg-white text-primary-600 border border-primary-600 rounded-lg hover:bg-primary-50 transition"
        >
          Giao dịch ngân hàng
        </a>
      </div>
    </div>
  )
}