// Component hiển thị skeleton loading khi lazy-load các trang.
// Dùng animate-pulse của Tailwind để tạo hiệu ứng "đang tải" với:
//   - Một thanh ngang (giả title)
//   - Ba khối chữ nhật (giả nội dung)
// Được dùng làm fallback trong dynamic import của Next.js.
export default function RouteLoading() {
  return (
    <main className="mx-auto min-h-dvh w-full max-w-[390px] bg-[#f9fafb] px-5 pt-20">
      <div className="animate-pulse space-y-5" role="status" aria-label="Đang tải trang">
        <div className="h-8 w-44 rounded-xl bg-emerald-900/10" />
        <div className="h-48 rounded-3xl bg-white shadow-sm" />
        <div className="h-24 rounded-2xl bg-white shadow-sm" />
        <div className="h-24 rounded-2xl bg-white shadow-sm" />
      </div>
    </main>
  );
}
