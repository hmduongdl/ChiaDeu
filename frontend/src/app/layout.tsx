// ============================================================================
// RootLayout — Layout gốc của ứng dụng Next.js
// ============================================================================
// File này định nghĩa cấu trúc HTML chung cho toàn bộ app.
// Metadata dùng cho SEO và PWA (manifest, theme color, viewport).
// Mọi page đều được render bên trong <main> của layout này.

import type { Metadata } from 'next'
import './globals.css'

// Metadata cho SEO và PWA
export const metadata: Metadata = {
  title: 'Cash Flow Minimizer',                             // Tiêu đề trang
  description: 'Ứng dụng chia tiền nhóm thông minh, tự động đồng bộ giao dịch ngân hàng và tối thiểu hoá thanh toán', // Mô tả cho SEO
  manifest: '/manifest.json',                               // PWA manifest
  themeColor: '#3b82f6',                                    // Màu theme cho PWA (xanh primary)
  viewport: 'width=device-width, initial-scale=1',          // Responsive
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="vi">
      <body>
        <main className="min-h-screen">
          {children}
        </main>
      </body>
    </html>
  )
}