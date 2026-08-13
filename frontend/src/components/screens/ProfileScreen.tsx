// Màn hình Cá nhân (/profile) — thông tin tài khoản và cài đặt.
// Gồm các phần:
//   - Header: tiêu đề "Trang Cá Nhân" + nút cài đặt
//   - Avatar + tên + email người dùng (có huy hiệu "Đã xác minh")
//   - Card tổng chi tiêu tháng: gradient xanh, hiển thị số tiền + % thay đổi
//   - Menu TÀI KHOẢN: Thông tin cá nhân, Phương thức thanh toán, Bảo mật, Lịch sử giao dịch
//   - Menu ỨNG DỤNG: Thông báo, Ngôn ngữ, Trợ giúp & Hỗ trợ
//   - Nút Đăng xuất: gọi API logout, redirect về /login
// Dữ liệu người dùng lấy từ useAuthStore, các mục menu là tĩnh.
"use client";

import { useRouter } from "next/navigation";
import { useAuthStore } from "@/stores/auth-store";

const PROFILE_ASSET = "/figma/profile";

const accountItems = [
  {
    label: "Thông tin cá nhân",
    icon: `${PROFILE_ASSET}/imgContainer3.svg`,
    iconBackground: "from-[#eff6ff] to-[#dbeafe]",
  },
  {
    label: "Phương thức thanh toán",
    icon: `${PROFILE_ASSET}/imgContainer5.svg`,
    iconBackground: "from-[#ecfdf5] to-[#d1fae5]",
  },
  {
    label: "Bảo mật",
    icon: `${PROFILE_ASSET}/imgContainer6.svg`,
    iconBackground: "from-[#faf5ff] to-[#f3e8ff]",
  },
  {
    label: "Lịch sử giao dịch",
    icon: `${PROFILE_ASSET}/imgContainer7.svg`,
    iconBackground: "from-[#fff1f2] to-[#ffe4e6]",
  },
] as const;

const appItems = [
  {
    label: "Thông báo",
    icon: `${PROFILE_ASSET}/imgContainer8.svg`,
    iconBackground: "from-[#fffbeb] to-[#fef3c7]",
  },
  {
    label: "Ngôn ngữ",
    value: "Tiếng Việt",
    icon: `${PROFILE_ASSET}/imgContainer9.svg`,
    iconBackground: "from-[#eef2ff] to-[#e0e7ff]",
  },
  {
    label: "Trợ giúp & Hỗ trợ",
    icon: `${PROFILE_ASSET}/imgContainer10.svg`,
    iconBackground: "from-[#fff1f2] to-[#ffe4e6]",
  },
] as const;

type MenuItem = (typeof accountItems)[number] | (typeof appItems)[number];

function ProfileMenu({ items }: { items: readonly MenuItem[] }) {
  return (
    <div className="overflow-hidden rounded-[20px] bg-white shadow-[0_4px_14px_rgba(15,23,42,0.06)]">
      {items.map((item, index) => (
        <button
          key={item.label}
          type="button"
          className={`flex h-[61px] w-full items-center gap-4 px-3.5 text-left transition-colors hover:bg-[#f9fafb] ${
            index > 0 ? "border-t border-[#f3f4f6]" : ""
          }`}
        >
          <span className={`flex h-10 w-10 shrink-0 items-center justify-center rounded-2xl bg-gradient-to-br ${item.iconBackground}`}>
            <img src={item.icon} alt="" className="h-[17px] w-[17px]" />
          </span>
          <span className="flex-1 text-sm font-medium text-[#4b5563]">{item.label}</span>
          {"value" in item ? <span className="text-xs text-[#9ca3af]">{item.value}</span> : null}
          <img src={`${PROFILE_ASSET}/imgContainer4.svg`} alt="" className="h-3 w-[7px]" />
        </button>
      ))}
    </div>
  );
}

export default function ProfileScreen() {
  const router = useRouter();
  const user = useAuthStore((state) => state.user);
  const logout = useAuthStore((state) => state.logout);

  const handleLogout = async () => {
    await logout();
    router.replace("/login");
    router.refresh();
  };

  return (
    <main className="mx-auto min-h-dvh w-full max-w-[390px] bg-[#f9fafb] pb-32 text-[#151c27] shadow-[0_0_30px_rgba(15,23,42,0.04)]">
      <header className="sticky top-0 z-20 flex h-16 items-center justify-center border-b border-[#f3f4f6] bg-white/90 px-5 shadow-[0_2px_8px_rgba(0,0,0,0.02)] backdrop-blur-md">
        <h1 className="text-lg font-semibold text-[#065f46]">Trang Cá Nhân</h1>
        <button
          type="button"
          aria-label="Mở cài đặt"
          className="absolute right-5 flex h-10 w-10 items-center justify-center rounded-full bg-white shadow-[0_2px_8px_rgba(0,0,0,0.06)]"
        >
          <img src={`${PROFILE_ASSET}/imgContainer12.svg`} alt="" className="h-[17px] w-[17px]" />
        </button>
      </header>

      <div className="px-4 pt-7">
        <section className="flex flex-col items-center">
          <div className="relative h-[88px] w-[88px] rounded-full bg-gradient-to-br from-[#10b981] to-[#2dd4bf] p-[3px] shadow-[0_4px_10px_rgba(16,185,129,0.25)]">
            <img
              src={user?.avatarUrl || `${PROFILE_ASSET}/imgUserAvatar.png`}
              alt="Ảnh đại diện"
              className="h-full w-full rounded-full border-4 border-white object-cover"
            />
            <button
              type="button"
              aria-label="Đổi ảnh đại diện"
              className="absolute bottom-0 right-0 flex h-7 w-7 items-center justify-center rounded-full border border-[#f3f4f6] bg-white shadow-md"
            >
              <img src={`${PROFILE_ASSET}/imgContainer.svg`} alt="" className="h-3 w-3" />
            </button>
          </div>
          <div className="mt-4 flex items-center gap-1.5">
            <h2 className="text-xl font-bold">{user?.name || "Minh Nguyễn"}</h2>
            <img src={`${PROFILE_ASSET}/imgContainer1.svg`} alt="Đã xác minh" className="h-4 w-[17px]" />
          </div>
          <p className="mt-1 text-sm text-[#9ca3af]">{user?.email || "minh.nguyen@example.com"}</p>
        </section>

        <section className="relative mt-5 overflow-hidden rounded-[18px] bg-gradient-to-r from-[#047857] to-[#065f46] p-3.5 text-white shadow-[0_5px_12px_rgba(6,95,70,0.18)]">
          <div className="flex items-center gap-2 text-[11px] font-medium">
            <img src={`${PROFILE_ASSET}/imgContainer2.svg`} alt="" className="h-3.5 w-3.5" />
            Tổng chi tiêu tháng
          </div>
          <div className="mt-1 flex items-end gap-1.5">
            <strong className="text-[22px] leading-7">12.680.000</strong>
            <span className="pb-0.5 text-xs">VND</span>
            <span className="pb-0.5 text-xs font-semibold">/ 4 nhóm tham gia</span>
          </div>
          <span className="absolute right-3 top-3 rounded-full bg-white px-2 py-1 text-[10px] font-medium text-[#d97706]">
            +12% so với tháng trước
          </span>
        </section>

        <section className="mt-4">
          <h2 className="mb-2 px-0.5 text-[11px] font-semibold tracking-[0.5px] text-[#6b7280]">TÀI KHOẢN</h2>
          <ProfileMenu items={accountItems} />
        </section>

        <section className="mt-5">
          <h2 className="mb-2 px-0.5 text-[11px] font-semibold tracking-[0.5px] text-[#6b7280]">ỨNG DỤNG</h2>
          <ProfileMenu items={appItems} />
        </section>

        <button
          type="button"
          onClick={() => void handleLogout()}
          className="mt-9 flex h-12 w-full items-center justify-center gap-2 rounded-[20px] bg-[#fef2f2] text-sm font-semibold text-[#ef4444]"
        >
          <img src={`${PROFILE_ASSET}/imgContainer11.svg`} alt="" className="h-[15px] w-[15px]" />
          Đăng xuất
        </button>
        <p className="mt-7 text-center text-xs text-[#9ca3af]">Phiên bản 2.1.0</p>
      </div>
    </main>
  );
}
