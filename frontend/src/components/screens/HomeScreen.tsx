"use client";

import Link from "next/link";
import AvatarStack from "@/components/app/AvatarStack";
import { useAuthStore } from "@/stores/auth-store";

const HOME_ASSET = "/figma/home";

type HomeGroup = {
  name: string;
  icon: string;
  iconBackground: string;
  avatars: readonly string[];
  overflow?: number;
  status: string;
  statusClassName: string;
};

const groups: readonly HomeGroup[] = [
  {
    name: "Chuyến đi Đà Lạt",
    icon: `${HOME_ASSET}/imgSvg3.svg`,
    iconBackground: "from-[#e0e7ff] to-[#c7d2fe]",
    avatars: [
      `${HOME_ASSET}/imgMember.png`,
      `${HOME_ASSET}/imgMember1.png`,
      `${HOME_ASSET}/imgMember2.png`,
    ],
    overflow: 2,
    status: "Bạn nợ 50k",
    statusClassName: "bg-[#fef2f2] text-[#ef4444]",
  },
  {
    name: "Ăn trưa văn phòng",
    icon: `${HOME_ASSET}/imgSvg4.svg`,
    iconBackground: "from-[#fef3c7] to-[#fde68a]",
    avatars: [`${HOME_ASSET}/imgMember3.png`, `${HOME_ASSET}/imgMember4.png`],
    status: "Được trả 100k",
    statusClassName: "bg-[#ecfdf5] text-[#047857]",
  },
  {
    name: "Tiền điện tháng 10",
    icon: `${HOME_ASSET}/imgSvg5.svg`,
    iconBackground: "from-[#f1f5f9] to-[#e2e8f0]",
    avatars: [
      `${HOME_ASSET}/imgMember5.png`,
      `${HOME_ASSET}/imgMember6.png`,
      `${HOME_ASSET}/imgMember7.png`,
    ],
    status: "Đã thanh toán",
    statusClassName: "bg-[#f3f4f6] text-[#6b7280]",
  },
];

export default function HomeScreen() {
  const user = useAuthStore((state) => state.user);
  const displayName = user?.name?.trim() || "bạn";

  return (
    <main className="mx-auto min-h-dvh w-full max-w-[390px] bg-[#f9fafb] pb-32 text-[#151c27] shadow-[0_0_30px_rgba(15,23,42,0.04)]">
      <header className="flex h-16 items-center justify-between border-b border-[#f3f4f6] bg-white px-5 shadow-[0_1px_2px_rgba(0,0,0,0.04)]">
        <Link href="/dashboard" className="flex items-center gap-2" aria-label="Chia Đều - Trang chủ">
          <span className="flex h-8 w-8 items-center justify-center rounded-lg bg-[#065f46] shadow-sm">
            <img src={`${HOME_ASSET}/imgSvg1.svg`} alt="" className="h-5 w-5" />
          </span>
          <span className="text-xl font-bold tracking-[0.2px] text-[#064e3b]">CHIADEU</span>
        </Link>
        <Link
          href="/profile"
          aria-label="Mở trang cá nhân"
          className="h-10 w-10 rounded-full bg-gradient-to-br from-[#34d399] to-[#065f46] p-0.5 shadow-sm"
        >
          <img
            src={user?.avatarUrl || `${HOME_ASSET}/imgUserAvatar.png`}
            alt="Ảnh đại diện"
            className="h-full w-full rounded-full border-2 border-white object-cover"
          />
        </Link>
      </header>

      <div className="px-5 pt-4">
        <section>
          <h1 className="text-[28px] font-bold leading-[35px] tracking-[-0.5px]">
            Xin chào, <span className="text-[#047857]">{displayName}!</span>
          </h1>
          <p className="mt-0.5 text-sm leading-5 text-[#9ca3af]">
            Chia tiền minh bạch - san sẻ yêu thương!
          </p>
        </section>

        <section className="relative mt-6 overflow-hidden rounded-xl border border-[#e5e7eb] bg-white p-6 shadow-[0_4px_14px_rgba(15,23,42,0.02)]">
          <div aria-hidden="true" className="absolute -right-9 -top-10 h-32 w-32 rounded-full bg-emerald-50/60" />
          <p className="relative text-sm text-[#4b5563]">Tổng số dư</p>
          <p className="relative mt-1 text-[28px] font-bold leading-10 text-[#065f46]">+300.000đ</p>
          <div className="relative mt-2 grid grid-cols-[1fr_1px_1fr] gap-4 border-t border-[#e5e7eb] pt-4">
            <div>
              <div className="flex items-center gap-1 text-sm text-[#4b5563]">
                <img src={`${HOME_ASSET}/imgContainer.svg`} alt="" className="h-[11px] w-[11px]" />
                Bạn nợ
              </div>
              <p className="mt-1 text-xl font-semibold text-[#dc2626]">150.000đ</p>
            </div>
            <span aria-hidden="true" className="h-10 self-center bg-[#e5e7eb]" />
            <div>
              <div className="flex items-center gap-1 text-sm text-[#4b5563]">
                <img src={`${HOME_ASSET}/imgContainer1.svg`} alt="" className="h-[11px] w-[11px]" />
                Bạn được trả
              </div>
              <p className="mt-1 text-xl font-semibold text-[#065f46]">450.000đ</p>
            </div>
          </div>
        </section>

        <section className="mt-10">
          <div className="flex items-center justify-between">
            <h2 className="text-xl font-bold">Nhóm của bạn</h2>
            <Link href="/groups" className="flex items-center gap-1 text-sm font-semibold text-[#047857]">
              Xem tất cả
              <img src={`${HOME_ASSET}/imgSvg2.svg`} alt="" className="h-4 w-4" />
            </Link>
          </div>

          <div className="mt-4 space-y-3">
            {groups.map((group) => (
              <Link
                key={group.name}
                href="/groups"
                className="flex min-h-[84px] items-center gap-4 rounded-2xl bg-white p-4 shadow-[0_3px_12px_rgba(15,23,42,0.06)] transition-transform active:scale-[0.99]"
              >
                <span className={`flex h-12 w-12 shrink-0 items-center justify-center rounded-2xl bg-gradient-to-br ${group.iconBackground}`}>
                  <img src={group.icon} alt="" className="h-6 w-6" />
                </span>
                <span className="min-w-0 flex-1">
                  <span className="block truncate text-base font-semibold">{group.name}</span>
                  <span className="mt-1 block">
                    <AvatarStack avatars={group.avatars} overflow={group.overflow} />
                  </span>
                </span>
                <span className={`shrink-0 rounded-full px-2.5 py-1 text-xs font-medium ${group.statusClassName}`}>
                  {group.status}
                </span>
              </Link>
            ))}
          </div>
        </section>
      </div>

      <Link
        href="/groups"
        aria-label="Tạo nhóm mới"
        className="fixed bottom-[105px] right-[max(16px,calc((100vw-390px)/2+16px))] z-30 flex h-14 w-14 items-center justify-center rounded-full bg-gradient-to-br from-[#10b981] to-[#065f46] shadow-[0_8px_20px_-4px_rgba(6,95,70,0.5)]"
      >
        <img src={`${HOME_ASSET}/imgSvg.svg`} alt="" className="h-6 w-6" />
      </Link>
    </main>
  );
}
