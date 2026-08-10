// ============================================================================
// QRCode — Component hiển thị QR code cho thanh toán
// ============================================================================
// Sử dụng thư viện qrcode để render QR code lên canvas.
// Nhận data (payload/link) và size (kích thước pixel).
// QR code dùng để người dùng quét và thanh toán qua PayOS hoặc MoMo.

'use client'

import { useEffect, useRef } from 'react'
import QRCodeLib from 'qrcode'

interface QRCodeProps {
  data: string    // Dữ liệu QR (URL/payload từ PayOS hoặc MoMo)
  size?: number   // Kích thước QR code (pixel), mặc định 200
}

export default function QRCode({ data, size = 200 }: QRCodeProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null) // Tham chiếu đến canvas

  // Vẽ QR code mỗi khi data hoặc size thay đổi
  useEffect(() => {
    if (canvasRef.current && data) {
      QRCodeLib.toCanvas(canvasRef.current, data, { width: size })
    }
  }, [data, size])

  return <canvas ref={canvasRef} />
}