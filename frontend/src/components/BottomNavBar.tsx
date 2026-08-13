"use client";

import type { ComponentType, SVGProps } from "react";
import { useState } from "react";
import { usePathname } from "next/navigation";
import { LayoutGroup, motion, MotionConfig } from "framer-motion";

type Tab = "home" | "group" | "activity" | "profile";
type IconProps = SVGProps<SVGSVGElement>;

type TabItem = {
  key: Tab;
  label: string;
  icon: ComponentType<IconProps>;
};

const ICON_PROPS: IconProps = {
  "aria-hidden": true,
  focusable: false,
  fill: "none",
  xmlns: "http://www.w3.org/2000/svg",
};

function HomeIcon(props: IconProps) {
  return (
    <svg {...ICON_PROPS} {...props} width="16" height="18" viewBox="0 0 16 18">
      <path
        d="M0 18V6L8 0L16 6V18H10V11H6V18H0Z"
        fill="currentColor"
      />
    </svg>
  );
}

function GroupIcon(props: IconProps) {
  return (
    <svg {...ICON_PROPS} {...props} width="22" height="20" viewBox="0 0 24 20">
      <circle cx="12" cy="5" r="3" fill="currentColor" />
      <circle cx="4.5" cy="7" r="2.5" fill="currentColor" />
      <circle cx="19.5" cy="7" r="2.5" fill="currentColor" />
      <path
        d="M12 10.5C8.68629 10.5 6 13.1863 6 16.5V19H18V16.5C18 13.1863 15.3137 10.5 12 10.5Z"
        fill="currentColor"
      />
      <path
        d="M0 17C0 13.9624 2.46243 11.5 5.5 11.5C6.22296 11.5 6.91334 11.6395 7.54584 11.8929C5.96675 13.308 5 15.3527 5 17.5V18H0V17Z"
        fill="currentColor"
      />
      <path
        d="M24 17C24 13.9624 21.5376 11.5 18.5 11.5C17.777 11.5 17.0867 11.6395 16.4542 11.8929C18.0332 13.308 19 15.3527 19 17.5V18H24V17Z"
        fill="currentColor"
      />
    </svg>
  );
}

function ActivityIcon(props: IconProps) {
  return (
    <svg {...ICON_PROPS} {...props} width="18" height="20" viewBox="0 0 18 20">
      <path
        d="M9 0C8.17157 0 7.5 0.671573 7.5 1.5V2.08982C4.92327 2.74104 3 5.07342 3 7.85714V11.1716C3 12.1333 2.61795 13.0556 1.93782 13.7357L0.43934 15.2342C-0.19063 15.8642 0.25554 16.9412 1.14645 16.9412H16.8536C17.7445 16.9412 18.1906 15.8642 17.5607 15.2342L16.0622 13.7357C15.3821 13.0556 15 12.1333 15 11.1716V7.85714C15 5.07342 13.0767 2.74104 10.5 2.08982V1.5C10.5 0.671573 9.82843 0 9 0Z"
        fill="currentColor"
      />
      <path
        d="M6.75 18C7.1394 19.1652 8.02913 20 9 20C9.97087 20 10.8606 19.1652 11.25 18H6.75Z"
        fill="currentColor"
      />
    </svg>
  );
}

function ProfileIcon(props: IconProps) {
  return (
    <svg {...ICON_PROPS} {...props} width="16" height="16" viewBox="0 0 16 16">
      <circle cx="8" cy="5" r="3" stroke="currentColor" strokeWidth="1.5" />
      <path
        d="M2 14C2 11.2386 4.23858 9 7 9H9C11.7614 9 14 11.2386 14 14V15C14 15.5523 13.5523 16 13 16H3C2.44772 16 2 15.5523 2 15V14Z"
        fill="currentColor"
      />
    </svg>
  );
}

const TABS: readonly TabItem[] = [
  { key: "home", label: "Trang chủ", icon: HomeIcon },
  { key: "group", label: "Nhóm", icon: GroupIcon },
  { key: "activity", label: "Hoạt động", icon: ActivityIcon },
  { key: "profile", label: "Cá nhân", icon: ProfileIcon },
];

const PILL_TRANSITION = {
  type: "spring",
  stiffness: 500,
  damping: 35,
} as const;

const ICON_TRANSITION = {
  type: "spring",
  stiffness: 400,
  damping: 20,
} as const;

export default function BottomNavBar() {
  const [activeTab, setActiveTab] = useState<Tab>("home");
  const hasNotification = true;
  const pathname = usePathname();

  if (pathname === "/login" || pathname === "/register" || pathname === "/forgot-password") {
    return null;
  }

  return (
    <MotionConfig reducedMotion="user">
      <LayoutGroup id="bottom-navigation">
        <nav
          aria-label="Điều hướng chính"
          className="fixed bottom-4 left-1/2 z-50 isolate grid w-[358px] max-w-[calc(100vw-2rem)] -translate-x-1/2 grid-cols-4 gap-1 rounded-[20px] px-[18.69px] py-[13px] shadow-[0px_-4px_16px_0px_rgba(0,0,0,0.06)]"
        >
          <span
            aria-hidden="true"
            className="pointer-events-none absolute inset-0 -z-10 rounded-[20px] border border-[#eef0f5] bg-white/80 backdrop-blur-[6px]"
          />

          {TABS.map(({ key, label, icon: Icon }) => {
            const isActive = activeTab === key;

            return (
              <button
                key={key}
                type="button"
                aria-pressed={isActive}
                onClick={() => setActiveTab(key)}
                className={`relative flex w-full min-w-0 flex-col items-center justify-center rounded-full px-1 py-2 transition-colors duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[#0d7a5f] focus-visible:ring-offset-2 motion-reduce:transition-none ${
                  isActive ? "text-white" : "text-[#9ca3af]"
                }`}
              >
                {isActive && (
                  <motion.span
                    layoutId="active-pill"
                    aria-hidden="true"
                    className="pointer-events-none absolute inset-0 rounded-full bg-gradient-to-r from-[#065f46] to-[#0d7a5f]"
                    style={{ willChange: "transform" }}
                    transition={PILL_TRANSITION}
                  />
                )}

                <span className="relative z-10 flex h-5 w-[22px] items-center justify-center">
                  <motion.span
                    className="flex h-5 w-[22px] origin-center items-center justify-center [&>svg]:block"
                    animate={{ scale: isActive ? 1.08 : 1 }}
                    transition={ICON_TRANSITION}
                  >
                    <Icon />
                  </motion.span>
                </span>

                <span
                  className={`relative z-10 whitespace-nowrap pt-1 text-[10px] leading-[15px] ${
                    isActive ? "font-bold" : "font-medium"
                  }`}
                >
                  {label}
                </span>

                {key === "activity" && hasNotification && (
                  <>
                    <span className="sr-only">Có hoạt động mới</span>
                    <span
                      aria-hidden="true"
                      className="pointer-events-none absolute right-4 top-[8px] z-20 h-2 w-2 rounded-full border-2 border-white bg-[#ef4444]"
                    />
                  </>
                )}
              </button>
            );
          })}
        </nav>
      </LayoutGroup>
    </MotionConfig>
  );
}
