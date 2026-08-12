/** @type {import('next').NextConfig} */
const nextConfig = {
  async rewrites() {
    return [
      {
        source: '/api/backend/:path*',
        destination: `${process.env.BACKEND_URL || 'http://localhost:8080'}/api/backend/:path*`,
      },
    ]
  },
}

module.exports = nextConfig
