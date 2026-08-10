import type { Metadata } from 'next'
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
      <body className="min-h-screen bg-gray-50">{children}</body>
    </html>
  )
}