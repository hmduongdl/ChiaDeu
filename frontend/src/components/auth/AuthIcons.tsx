// Tập trung tất cả asset icon dùng trong các màn hình xác thực.
// - authAssets: object chứa URL icon SVG từ Figma cho login, register, forgot-password
// - FigmaIcon: component nhỏ hiển thị icon từ URL (dùng thẻ <img>)
// Tách riêng để dễ bảo trì khi cần đổi icon hoặc chuyển sang sprite sheet.
"use client";

export const authAssets = {
  login: {
    eye: "/icons/eye.svg",
    arrow: "/icons/arrow.svg",
    biometric: "/icons/biometric.svg",
    google: "/icons/google.svg",
    facebook: "/icons/facebook.svg",
    apple: "/icons/apple.svg",
  },
  register: {
    person: "/icons/person.svg",
    email: "/icons/email.svg",
    lock: "/icons/lock.svg",
    confirm: "/icons/confirm.svg",
    google: "/icons/google.svg",
    apple: "/icons/apple.svg",
  },
  forgot: {
    back: "/icons/back.svg",
    arrow: "/icons/arrow.svg",
    key: "/icons/key.svg",
    email: "/icons/email.svg",
  },
} as const;

export function FigmaIcon({ src, alt = "", className = "" }: { src: string; alt?: string; className?: string }) {
  // Đã chuyển sang dùng local assets thay vì MCP URL
  return <img src={src} alt={alt} className={`block object-contain ${className}`} />;
}
