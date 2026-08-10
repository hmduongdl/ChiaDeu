// ============================================================================
// Tailwind CSS Configuration
// ============================================================================
// Cấu hình content paths để Tailwind quét class và tree-shake CSS không dùng.
// Màu primary được định nghĩa thêm với các shade: 50, 500, 600, 700.

import type { Config } from 'tailwindcss'

const config: Config = {
  content: [
    './src/pages/**/*.{js,ts,jsx,tsx,mdx}',
    './src/components/**/*.{js,ts,jsx,tsx,mdx}',
    './src/app/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  theme: {
    extend: {
      colors: {
        // Màu chủ đạo xanh dương (primary)
        primary: {
          50: '#eff6ff',  // Xanh nhạt (nền hover outline)
          500: '#3b82f6', // Xanh trung bình (ring focus)
          600: '#2563eb', // Xanh chính (nút, text)
          700: '#1d4ed8', // Xanh đậm (hover nút)
        },
      },
    },
  },
  plugins: [],
}
export default config