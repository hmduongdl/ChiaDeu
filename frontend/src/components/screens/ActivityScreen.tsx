// Màn hình Hoạt động (/activity) — timeline thông báo và sự kiện.
// Gồm 2 phần:
//   - HÔM NAY: các hoạt động mới nhất (có viền trái xanh, nền gradient nhạt)
//     + Chi phí mới: Minh thêm "Ăn tối" trong nhóm "Chuyến đi Đà Lạt"
//     + Lời mời tham gia nhóm: có nút "Tham gia" / "Từ chối" (tương tác được)
//   - TRƯỚC ĐÓ: các hoạt động cũ hơn (nền trắng, viền xám)
// Có nút "đánh dấu tất cả là đã đọc" và badge số lượng thông báo chưa đọc.
// Dữ liệu hiện tại là mock data.
"use client";

import { useState } from "react";

const ACTIVITY_ASSET = "/figma/activity";

const olderActivities = [
  {
    title: "Yêu cầu tất toán",
    time: "Hôm qua",
    description: "Hoàng đã gửi yêu cầu tất toán nợ cho bạn.",
    icon: `${ACTIVITY_ASSET}/imgContainer2.svg`,
    background: "bg-[#fff1f2]",
  },
  {
    title: "Nhắc nhở",
    time: "2 ngày trước",
    description: "Bạn còn khoản nợ chưa thanh toán trong nhóm 'Đôi lứa'.",
    icon: `${ACTIVITY_ASSET}/imgContainer3.svg`,
    background: "bg-[#fff1f2]",
  },
  {
    title: "Nhận thanh toán",
    time: "Tuần trước",
    description: "An đã thanh toán 150,000đ cho khoản 'Tiền cafe'.",
    icon: `${ACTIVITY_ASSET}/imgContainer4.svg`,
    background: "bg-[#f0fdfa]",
  },
] as const;

function SectionHeading({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex items-center gap-3">
      <h2 className="shrink-0 text-[13px] font-semibold tracking-[0.65px] text-[#6b7280]">{children}</h2>
      <span className="h-px flex-1 bg-[#e5e7eb]" />
    </div>
  );
}

export default function ActivityScreen() {
  const [isRead, setIsRead] = useState(false);
  const [invitationResponse, setInvitationResponse] = useState<"accepted" | "declined" | null>(null);

  return (
    <main className="mx-auto min-h-dvh w-full max-w-[390px] bg-[#f9fafb] pb-32 text-[#151c27] shadow-[0_20px_25px_-5px_rgba(0,0,0,0.1)]">
      <header className="sticky top-0 z-20 flex h-16 items-center justify-center border-b border-[#f3f4f6] bg-white/90 px-5 shadow-[0_2px_8px_rgba(0,0,0,0.02)] backdrop-blur-md">
        <div className="flex items-center gap-2">
          <h1 className="text-xl font-bold text-[#047857]">Hoạt động</h1>
          {!isRead ? (
            <span className="flex h-4 min-w-4 items-center justify-center rounded-full bg-[#065f46] px-1 text-[9px] font-bold text-white">2</span>
          ) : null}
        </div>
        <button
          type="button"
          onClick={() => setIsRead(true)}
          aria-label="Đánh dấu tất cả là đã đọc"
          className="absolute right-3 flex h-10 w-10 items-center justify-center rounded-full bg-white shadow-[1.4px_3px_3px_rgba(0,0,0,0.2)]"
        >
          <img src={`${ACTIVITY_ASSET}/imgContainer5.svg`} alt="" className="h-[11px] w-[19px]" />
        </button>
      </header>

      <div className="space-y-7 px-5 pb-8 pt-5">
        <section>
          <SectionHeading>HÔM NAY</SectionHeading>
          <div className="mt-3 space-y-4">
            <article className={`relative flex gap-3 overflow-hidden rounded-2xl border-l-4 border-[#065f46] px-4 py-4 shadow-[0_2px_8px_rgba(0,0,0,0.05)] ${isRead ? "bg-white" : "bg-gradient-to-r from-[#eef2ff] to-[#f5f7ff]"}`}>
              <span className="flex h-12 w-12 shrink-0 items-center justify-center rounded-full bg-gradient-to-br from-[#d1fae5] to-[#a7f3d0]">
                <img src={`${ACTIVITY_ASSET}/imgContainer.svg`} alt="" className="h-5 w-[18px]" />
              </span>
              <div className="min-w-0 flex-1">
                <div className="flex items-start justify-between gap-2">
                  <h2 className="text-[21px] font-semibold leading-6">Chi phí mới</h2>
                  <span className="mt-1 shrink-0 text-xs text-[#6b7280]">2 phút trước</span>
                </div>
                <p className="mt-1 text-[13px] font-medium leading-[18px] text-[#4b5563]">
                  <strong>Minh</strong> đã thêm một chi phí mới &apos;Ăn tối&apos; trong nhóm &apos;Chuyến đi Đà Lạt&apos;.
                </p>
              </div>
              {!isRead ? <span className="absolute right-3 top-3 h-2.5 w-2.5 rounded-full bg-[#047857] shadow-[0_0_8px_rgba(4,120,87,0.65)]" /> : null}
            </article>

            <article className={`relative flex gap-3 overflow-hidden rounded-2xl border-l-4 border-[#065f46] px-4 py-4 shadow-[0_2px_8px_rgba(0,0,0,0.05)] ${isRead ? "bg-white" : "bg-gradient-to-r from-[#eef2ff] to-[#f5f7ff]"}`}>
              <span className="flex h-12 w-12 shrink-0 items-center justify-center rounded-full bg-gradient-to-br from-[#fef3c7] to-[#fde68a]">
                <img src={`${ACTIVITY_ASSET}/imgContainer1.svg`} alt="" className="h-4 w-6" />
              </span>
              <div className="min-w-0 flex-1">
                <div className="flex items-start justify-between gap-2">
                  <h2 className="text-[21px] font-semibold leading-6">Lời mời tham gia nhóm</h2>
                  <span className="mt-1 shrink-0 text-xs text-[#6b7280]">1 giờ trước</span>
                </div>
                <p className="mt-1 text-[13px] font-medium leading-[18px] text-[#4b5563]">
                  <strong>Linh</strong> đã mời bạn vào nhóm &apos;Nhà trọ 123&apos;.
                </p>
                {invitationResponse ? (
                  <p className="mt-3 text-sm font-semibold text-[#047857]">
                    {invitationResponse === "accepted" ? "Đã tham gia nhóm" : "Đã từ chối lời mời"}
                  </p>
                ) : (
                  <div className="mt-3 grid grid-cols-2 gap-2">
                    <button type="button" onClick={() => setInvitationResponse("accepted")} className="h-10 rounded-full bg-[#047857] text-sm font-semibold text-white">
                      Tham gia
                    </button>
                    <button type="button" onClick={() => setInvitationResponse("declined")} className="h-10 rounded-full border border-[#cbd5e1] text-sm font-semibold text-[#4b5563]">
                      Từ chối
                    </button>
                  </div>
                )}
              </div>
              {!isRead ? <span className="absolute right-3 top-3 h-2.5 w-2.5 rounded-full bg-[#047857] shadow-[0_0_8px_rgba(4,120,87,0.65)]" /> : null}
            </article>
          </div>
        </section>

        <section>
          <SectionHeading>TRƯỚC ĐÓ</SectionHeading>
          <div className="mt-3 space-y-4">
            {olderActivities.map((activity) => (
              <article key={activity.title} className="flex min-h-[102px] gap-3 rounded-2xl border border-[#e5e7eb] bg-white px-4 py-4">
                <span className={`flex h-12 w-12 shrink-0 items-center justify-center rounded-full ${activity.background}`}>
                  <img src={activity.icon} alt="" className="h-5 w-[22px]" />
                </span>
                <div className="min-w-0 flex-1">
                  <div className="flex items-start justify-between gap-2">
                    <h2 className="text-xl font-semibold leading-6 text-[#374151]">{activity.title}</h2>
                    <span className="mt-1 shrink-0 text-xs text-[#9ca3af]">{activity.time}</span>
                  </div>
                  <p className="mt-1 text-[13px] leading-[18px] text-[#6b7280]">{activity.description}</p>
                </div>
              </article>
            ))}
          </div>
        </section>
      </div>
    </main>
  );
}
