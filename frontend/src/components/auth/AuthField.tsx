// Component input dùng chung cho các form xác thực (đăng nhập, đăng ký, quên mật khẩu).
// Props:
//   - label: nhãn hiển thị phía trên input
//   - icon (tùy chọn): icon SVG hiển thị bên trái input (dùng FigmaIcon)
//   - Các props còn lại được truyền thẳng xuống thẻ <input> (type, placeholder, value, onChange, ...)
// Style: bo góc 12px, viền xám xanh, nền trắng ngà.
"use client";

import type { InputHTMLAttributes } from "react";
import { FigmaIcon } from "./AuthIcons";

type AuthFieldProps = InputHTMLAttributes<HTMLInputElement> & {
  label: string;
  icon?: string;
  iconClassName?: string;
};

export function AuthField({ label, icon, iconClassName = "h-5 w-5", id, ...props }: AuthFieldProps) {
  return (
    <label htmlFor={id} className="flex w-full flex-col gap-1 text-[12px] font-medium leading-4 text-[#3f4944]">
      <span>{label}</span>
      <span className="relative flex h-12 w-full items-center rounded-[12px] border border-[#bec9c2] bg-[#f9f9ff] shadow-[0px_1px_2px_rgba(0,0,0,0.05)]">
        {icon && <FigmaIcon src={icon} className={`absolute left-3 ${iconClassName}`} />}
        <input
          id={id}
          className={`h-full w-full rounded-[12px] bg-transparent px-[17px] text-[16px] text-[#3f4944] outline-none placeholder:text-[#bec9c2] ${icon ? "pl-[41px]" : ""}`}
          {...props}
        />
      </span>
    </label>
  );
}
