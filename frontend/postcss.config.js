// ============================================================================
// PostCSS Configuration
// ============================================================================
// Đăng ký 2 plugin: tailwindcss (xử lý utility class) và autoprefixer (thêm vendor prefix).

module.exports = {
  plugins: {
    tailwindcss: {},  // Biên dịch Tailwind directives → CSS
    autoprefixer: {}, // Tự động thêm -webkit-, -moz- prefix cho cross-browser
  },
}