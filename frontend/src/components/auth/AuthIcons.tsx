// Tập trung tất cả asset icon dùng trong các màn hình xác thực.
// - authAssets: object chứa URL icon SVG từ Figma cho login, register, forgot-password
// - FigmaIcon: component nhỏ hiển thị icon từ URL (dùng thẻ <img>)
// Tách riêng để dễ bảo trì khi cần đổi icon hoặc chuyển sang sprite sheet.
"use client";

export const authAssets = {
  login: {
    eye: "https://www.figma.com/api/mcp/asset/5bbf2052-1cd1-4232-b560-ecce7e8182ed.svg",
    arrow: "https://www.figma.com/api/mcp/asset/ebb1d18b-90e4-4ed7-a68e-d8386e96598d.svg",
    biometric: "https://www.figma.com/api/mcp/asset/f03423b5-61c0-482a-a493-6cf009407de9.svg",
    google: "https://www.figma.com/api/mcp/asset/f5bd3382-df1f-4a6f-b578-15c975ab0dc3.svg",
    facebook: "https://www.figma.com/api/mcp/asset/c2762bf2-ee48-4e68-8663-cabc9209b3ef.svg",
    apple: "https://www.figma.com/api/mcp/asset/c0120fa7-ac9d-4cad-bdf3-48ed9fb7990c.svg",
  },
  register: {
    person: "https://www.figma.com/api/mcp/asset/1c38e7d1-fdf5-43ac-a013-06605fd17337.svg",
    email: "https://www.figma.com/api/mcp/asset/96066051-6ae0-4a92-ad34-aafc4031a8a8.svg",
    lock: "https://www.figma.com/api/mcp/asset/39af368f-5f4f-4dd2-921c-36a39d07d093.svg",
    confirm: "https://www.figma.com/api/mcp/asset/a76a3233-c2e9-407b-bbe4-f6dbb8b3eb95.svg",
    google: "https://www.figma.com/api/mcp/asset/948d239d-d376-4bd3-9f91-c9827a52f5cf.svg",
    apple: "https://www.figma.com/api/mcp/asset/ac1c6602-860d-450e-891c-163b20969122.svg",
  },
  forgot: {
    back: "https://www.figma.com/api/mcp/asset/35f3f1e7-cd99-4067-a9b4-047fe4c30296.svg",
    arrow: "https://www.figma.com/api/mcp/asset/240c0025-754c-49a5-bfe2-6b5623ebee89.svg",
    key: "https://www.figma.com/api/mcp/asset/9d121702-2c5a-49ef-b0b3-9f9b4dbe0f39.svg",
    email: "https://www.figma.com/api/mcp/asset/fc413f21-36eb-4776-a6df-5927d8982c80.svg",
  },
} as const;

export function FigmaIcon({ src, alt = "", className = "" }: { src: string; alt?: string; className?: string }) {
  return <img src={src} alt={alt} className={`block object-contain ${className}`} />;
}
