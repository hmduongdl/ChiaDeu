// ============================================================================
// Button — Component nút bấm dùng chung toàn app
// ============================================================================
// Hỗ trợ 3 biến thể (variant): primary, secondary, outline
// Hỗ trợ 3 kích thước (size): sm, md, lg
// Kế thừa tất cả props của <button> HTML native.

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'outline' // Kiểu nút
  size?: 'sm' | 'md' | 'lg'                      // Kích thước
}

export default function Button({
  variant = 'primary',
  size = 'md',
  className = '',
  ...props
}: ButtonProps) {
  // Base styles: bo góc, font, transition, focus ring
  const base = 'rounded-lg font-medium transition-colors focus:outline-none focus:ring-2'

  // Các biến thể màu sắc
  const variants = {
    primary: 'bg-primary-600 text-white hover:bg-primary-700 focus:ring-primary-500',    // Nút chính, xanh đậm
    secondary: 'bg-gray-200 text-gray-800 hover:bg-gray-300 focus:ring-gray-400',        // Nút phụ, xám
    outline: 'border border-primary-600 text-primary-600 hover:bg-primary-50 focus:ring-primary-500', // Nút viền
  }

  // Các kích thước
  const sizes = {
    sm: 'px-3 py-1.5 text-sm',
    md: 'px-4 py-2 text-sm',
    lg: 'px-6 py-3 text-base',
  }

  return (
    <button
      className={`${base} ${variants[variant]} ${sizes[size]} ${className}`}
      {...props}
    />
  )
}