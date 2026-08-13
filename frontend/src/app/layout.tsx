// Root layout của ứng dụng Next.js.
// - Đặt ngôn ngữ html là "vi" (tiếng Việt)
// - Metadata: title "Chia Đều", description, PWA manifest
// - Render children + BottomNavBar (thanh điều hướng dưới cùng)
// - Body có nền xám nhạt (bg-gray-50) và chiều cao tối thiểu toàn màn hình
import type { Metadata } from 'next'
import BottomNavBar from '@/components/BottomNavBar'
import './globals.css'

export const metadata: Metadata = {
  title: 'Chia Đều - Cash Flow Minimizer',
  description: 'Chia tiền nhóm thông minh, tối thiểu giao dịch thanh toán',
  manifest: '/manifest.json',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="vi">
      <body className="min-h-screen bg-gray-50">
        {children}
        <BottomNavBar />
      </body>
    </html>
  )
}
