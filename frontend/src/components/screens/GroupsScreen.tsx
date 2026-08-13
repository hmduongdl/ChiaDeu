// Màn hình Nhóm (/groups) — quản lý danh sách nhóm chia tiền.
// Chức năng:
//   - Ô tìm kiếm nhóm theo tên (có hỗ trợ tiếng Việt không dấu)
//   - Bộ lọc: Tất cả / Còn nợ / Được trả
//   - Nút tạo nhóm mới (dấu +)
//   - Danh sách nhóm dạng card: icon, tên nhóm, danh mục, avatar thành viên, số tiền nợ/được trả
//   - Nhóm đã thanh toán xong hiển thị mờ (muted)
//   - Hiển thị "Không tìm thấy nhóm phù hợp" khi không có kết quả
// Dữ liệu hiện tại là mock data.
"use client";

import { useMemo, useState } from "react";
import AvatarStack from "@/components/app/AvatarStack";

const GROUP_ASSET = "/figma/groups";

type GroupFilter = "all" | "debt" | "credit";

type GroupSummary = {
  name: string;
  category: string;
  filter: GroupFilter;
  icon: string;
  avatars: readonly string[];
  overflow?: number;
  label: string;
  amount: string;
  amountClassName: string;
  cornerClassName: string;
  muted?: boolean;
};

const groups: readonly GroupSummary[] = [
  {
    name: "Chuyến đi Đà Lạt",
    category: "Du lịch",
    filter: "debt" as const,
    icon: `${GROUP_ASSET}/imgContainer.svg`,
    avatars: [`${GROUP_ASSET}/imgAvatar.png`, `${GROUP_ASSET}/imgAvatar1.png`],
    overflow: 3,
    label: "Bạn nợ",
    amount: "50.000₫",
    amountClassName: "text-[#dc2626]",
    cornerClassName: "bg-[#fce8e8]",
  },
  {
    name: "Ăn trưa văn phòng",
    category: "Ăn uống",
    filter: "credit" as const,
    icon: `${GROUP_ASSET}/imgContainer1.svg`,
    avatars: [`${GROUP_ASSET}/imgAvatar2.png`, `${GROUP_ASSET}/imgAvatar3.png`],
    label: "Được trả",
    amount: "120.000₫",
    amountClassName: "text-[#065f46]",
    cornerClassName: "bg-[#e7f0ed]",
  },
  {
    name: "Tiền điện tháng 10",
    category: "Sinh hoạt",
    filter: "all" as const,
    icon: `${GROUP_ASSET}/imgContainer2.svg`,
    avatars: [`${GROUP_ASSET}/imgAvatar4.png`, `${GROUP_ASSET}/imgAvatar5.png`],
    label: "Đã thanh toán",
    amount: "",
    amountClassName: "text-[#9ca3af]",
    cornerClassName: "bg-[#edf2f8]",
    muted: true,
  },
  {
    name: "Tiệc sinh nhật",
    category: "Sự kiện",
    filter: "credit" as const,
    icon: `${GROUP_ASSET}/imgContainer3.svg`,
    avatars: [`${GROUP_ASSET}/imgAvatar6.png`],
    overflow: 5,
    label: "Được trả",
    amount: "450.000₫",
    amountClassName: "text-[#065f46]",
    cornerClassName: "bg-[#e7f0ed]",
  },
];

const filters: readonly { key: GroupFilter; label: string }[] = [
  { key: "all", label: "Tất cả" },
  { key: "debt", label: "Còn nợ" },
  { key: "credit", label: "Được trả" },
];

export default function GroupsScreen() {
  const [activeFilter, setActiveFilter] = useState<GroupFilter>("all");
  const [searchTerm, setSearchTerm] = useState("");

  const visibleGroups = useMemo(() => {
    const normalizedSearch = searchTerm.trim().toLocaleLowerCase("vi");

    return groups.filter((group) => {
      const matchesFilter = activeFilter === "all" || group.filter === activeFilter;
      const matchesSearch = group.name.toLocaleLowerCase("vi").includes(normalizedSearch);
      return matchesFilter && matchesSearch;
    });
  }, [activeFilter, searchTerm]);

  return (
    <main className="mx-auto min-h-dvh w-full max-w-[390px] bg-[#f9f9ff] px-5 pb-32 pt-6 text-[#151c27] shadow-[0_0_30px_rgba(15,23,42,0.04)]">
      <h1 className="text-[26px] font-bold leading-8 text-[#004532]">Nhóm của bạn</h1>

      <label className="relative mt-4 block">
        <span className="sr-only">Tìm kiếm nhóm</span>
        <img
          src={`${GROUP_ASSET}/imgIcon.svg`}
          alt=""
          className="pointer-events-none absolute left-3 top-1/2 h-[18px] w-[18px] -translate-y-1/2"
        />
        <input
          type="search"
          value={searchTerm}
          onChange={(event) => setSearchTerm(event.target.value)}
          placeholder="Tìm kiếm nhóm..."
          className="h-12 w-full rounded-full border border-[#bec9c2] bg-white py-3 pl-10 pr-4 text-base outline-none shadow-[0_2px_8px_rgba(0,0,0,0.04)] placeholder:text-[#9ca3af] focus:border-[#0d7a5f] focus:ring-2 focus:ring-[#0d7a5f]/15"
        />
      </label>

      <div className="mt-6 flex items-center gap-3">
        {filters.map((filter) => (
          <button
            key={filter.key}
            type="button"
            onClick={() => setActiveFilter(filter.key)}
            aria-pressed={activeFilter === filter.key}
            className={`h-[34px] rounded-full border px-4 text-sm font-medium transition-colors ${
              activeFilter === filter.key
                ? "border-[#047857] bg-[#047857] text-white"
                : "border-[#d1d5db] bg-white text-[#4b5563]"
            }`}
          >
            {filter.label}
          </button>
        ))}
        <button
          type="button"
          aria-label="Tạo nhóm mới"
          className="ml-auto flex h-10 w-10 items-center justify-center rounded-2xl bg-[#004532] shadow-[0_4px_6px_rgba(0,0,0,0.12)]"
        >
          <img src={`${GROUP_ASSET}/imgContainer4.svg`} alt="" className="h-[15px] w-[15px]" />
        </button>
      </div>

      <section className="mt-7 space-y-4" aria-live="polite">
        {visibleGroups.map((group) => (
          <article
            key={group.name}
            className={`relative min-h-[144px] overflow-hidden rounded-[14px] bg-white p-4 shadow-[0_4px_14px_rgba(15,23,42,0.04)] ${
              group.muted ? "opacity-50 grayscale" : ""
            }`}
          >
            <span aria-hidden="true" className={`absolute -right-7 -top-7 h-[74px] w-[74px] rounded-full ${group.cornerClassName}`} />
            <div className="relative flex items-center gap-3">
              <span className="flex h-12 w-12 shrink-0 items-center justify-center rounded-full bg-[#e2e8f8]">
                <img src={group.icon} alt="" className="h-5 w-5" />
              </span>
              <div>
                <h2 className="text-[21px] font-semibold leading-7">{group.name}</h2>
                <p className="text-sm text-[#6b7280]">{group.category}</p>
              </div>
            </div>
            <div className="relative mt-7 flex items-end justify-between">
              <AvatarStack avatars={group.avatars} overflow={group.overflow} size="medium" muted={group.muted} />
              <div className="text-right">
                <p className={`text-xs ${group.amountClassName}`}>{group.label}</p>
                {group.amount ? <p className={`text-[24px] font-semibold leading-8 ${group.amountClassName}`}>{group.amount}</p> : null}
              </div>
            </div>
          </article>
        ))}

        {visibleGroups.length === 0 ? (
          <div className="rounded-2xl border border-dashed border-[#bec9c2] bg-white px-5 py-10 text-center text-sm text-[#6b7280]">
            Không tìm thấy nhóm phù hợp.
          </div>
        ) : null}
      </section>
    </main>
  );
}
