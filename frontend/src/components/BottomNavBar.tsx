"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { LayoutGroup, motion, MotionConfig } from "framer-motion";

type TabItem = {
  href: string;
  label: string;
  icon: string;
  iconClassName: string;
  hasNotification?: boolean;
};

const TABS: readonly TabItem[] = [
  {
    href: "/dashboard",
    label: "Trang chủ",
    icon: "/figma/navigation/imgContainer.svg",
    iconClassName: "h-[18px] w-4",
  },
  {
    href: "/groups",
    label: "Nhóm",
    icon: "/figma/navigation/imgContainer1.svg",
    iconClassName: "h-4 w-[22px]",
  },
  {
    href: "/activity",
    label: "Hoạt động",
    icon: "/figma/navigation/imgContainer2.svg",
    iconClassName: "h-5 w-4",
    hasNotification: true,
  },
  {
    href: "/profile",
    label: "Cá nhân",
    icon: "/figma/navigation/imgContainer3.svg",
    iconClassName: "h-4 w-4",
  },
];

const DOCK_ROUTES = new Set(TABS.map((tab) => tab.href));

export default function BottomNavBar() {
  const pathname = usePathname();

  if (!DOCK_ROUTES.has(pathname)) {
    return null;
  }

  return (
    <MotionConfig reducedMotion="user">
      <LayoutGroup id="bottom-navigation">
        <nav
          aria-label="Điều hướng chính"
          className="fixed bottom-4 left-1/2 z-50 flex h-[81px] w-[358px] max-w-[calc(100vw-2rem)] -translate-x-1/2 items-center justify-between rounded-[20px] border border-[#eef0f5] bg-white/80 px-[18.69px] py-[13px] shadow-[0_-4px_16px_rgba(0,0,0,0.06)] backdrop-blur-md"
        >
          {TABS.map((tab) => {
            const isActive = pathname === tab.href;

            return (
              <Link
                key={tab.href}
                href={tab.href}
                aria-current={isActive ? "page" : undefined}
                className={`relative flex h-[55px] flex-col items-center justify-center rounded-full py-2 text-[10px] transition-[padding,color] duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[#0d7a5f] focus-visible:ring-offset-2 ${
                  isActive
                    ? "px-6 font-bold text-white"
                    : "px-2 font-medium text-[#9ca3af]"
                }`}
              >
                {isActive ? (
                  <motion.span
                    layoutId="active-dock-pill"
                    aria-hidden="true"
                    className="absolute inset-0 rounded-full bg-gradient-to-r from-[#065f46] to-[#0d7a5f] shadow-[0_4px_6px_rgba(6,95,70,0.25)]"
                    transition={{ type: "spring", stiffness: 500, damping: 38 }}
                  />
                ) : null}

                <span
                  aria-hidden="true"
                  className={`relative z-10 shrink-0 bg-current ${tab.iconClassName}`}
                  style={{
                    WebkitMask: `url(${tab.icon}) center / contain no-repeat`,
                    mask: `url(${tab.icon}) center / contain no-repeat`,
                  }}
                />
                <span className="relative z-10 mt-1 whitespace-nowrap leading-[15px]">
                  {tab.label}
                </span>

                {tab.hasNotification ? (
                  <span
                    aria-label="Có hoạt động mới"
                    className={`absolute z-20 h-2 w-2 rounded-full border-2 ${
                      isActive
                        ? "right-5 top-1.5 border-[#065f46] bg-[#ef4444]"
                        : "right-1 top-1.5 border-white bg-[#ef4444]"
                    }`}
                  />
                ) : null}
              </Link>
            );
          })}
        </nav>
      </LayoutGroup>
    </MotionConfig>
  );
}
