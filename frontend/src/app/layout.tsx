// Root layout của ứng dụng Next.js.
// - Đặt ngôn ngữ html là "vi" (tiếng Việt)
// - Metadata: title "Chia Đều", description, PWA manifest
// - Render children + BottomNavBar (thanh điều hướng dưới cùng)
// - Body có nền xám nhạt (bg-gray-50) và chiều cao tối thiểu toàn màn hình
import type { Metadata } from 'next'
import { Inter } from 'next/font/google'
import BottomNavBar from '@/components/BottomNavBar'
import './globals.css'

const inter = Inter({ subsets: ['vietnamese'] })

export const metadata: Metadata = {
  title: 'Chia Đều - Chia tiền minh bạch, San sẻ yêu thương',
  description: 'Chia tiền nhóm thông minh, tối thiểu giao dịch thanh toán',
  manifest: '/manifest.json',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="vi" className="light">
      <head>
        <link href="https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:wght,FILL@100..700,0..1&display=swap" rel="stylesheet" />
      </head>
      <body className={`${inter.className} bg-background text-on-background min-h-[max(884px,100dvh)] flex flex-col antialiased selection:bg-primary-container selection:text-on-primary-container`}>
        {children}
        <BottomNavBar />
      </body>
    </html>
  )
}
