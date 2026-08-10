// ============================================================================
// Next.js Configuration
// ============================================================================
// - reactStrictMode: bật Strict Mode để phát hiện lỗi sớm
// - rewrites: proxy /api/* đến Go backend (localhost:8080) để tránh CORS
//   Khi frontend gọi /api/groups, Next.js chuyển tiếp đến http://localhost:8080/api/groups

/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,

  // Proxy API: chuyển tiếp request từ frontend đến Go backend
  async rewrites() {
    return [
      {
        source: '/api/:path*',                          // Tất cả request bắt đầu bằng /api
        destination: 'http://localhost:8080/api/:path*', // Chuyển đến Go backend
      },
    ]
  },
}

module.exports = nextConfig