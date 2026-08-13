"use client";

import Link from "next/link";
import { FormEvent, useState } from "react";
import { AuthField } from "@/components/auth/AuthField";
import { authAssets, FigmaIcon } from "@/components/auth/AuthIcons";

export default function RegisterPage() {
  const [showPassword, setShowPassword] = useState(false);
  const handleSubmit = (event: FormEvent<HTMLFormElement>) => event.preventDefault();
  const fields = [
    { id: "register-name", label: "Họ và tên", placeholder: "Nhập họ và tên", icon: authAssets.register.person },
    { id: "register-identity", label: "Email hoặc Số điện thoại", placeholder: "Nhập email hoặc SĐT", icon: authAssets.register.email, iconClassName: "h-4 w-5" },
  ];

  return (
    <main className="flex min-h-screen items-center justify-center bg-[#f9f9ff] px-4 py-[69px]">
      <div className="w-full max-w-[448px] rounded-2xl bg-white p-6 shadow-[0_8px_30px_rgba(0,0,0,0.04)]">
        <header className="flex flex-col items-center gap-1">
          <h1 className="text-[26px] font-bold leading-8 tracking-[-0.65px] text-[#004532]">CHIADEU</h1>
          <h2 className="pt-1 text-center text-xl font-semibold leading-7">Tạo tài khoản mới</h2>
          <p className="text-center text-sm leading-5 text-[#3f4944]">Tham gia để cùng chia sẻ chi phí minh bạch</p>
        </header>
        <form onSubmit={handleSubmit} className="mt-8 flex flex-col gap-4">
          {fields.map((field) => <AuthField key={field.id} {...field} />)}
          <div className="relative"><AuthField id="register-password" label="Mật khẩu" placeholder="Tạo mật khẩu" type={showPassword ? "text" : "password"} icon={authAssets.register.lock} autoComplete="new-password" /><button type="button" onClick={() => setShowPassword((value) => !value)} aria-label="Hiện hoặc ẩn mật khẩu" className="absolute right-3 top-7 h-8 w-8" /></div>
          <AuthField id="register-confirm" label="Xác nhận mật khẩu" placeholder="Nhập lại mật khẩu" type="password" icon={authAssets.register.confirm} />
          <button type="submit" className="h-12 rounded-xl bg-gradient-to-br from-[#004532] to-[#065f46] text-[12px] font-semibold leading-4 text-white shadow-sm">Đăng ký</button>
        </form>
        <div className="relative my-8 flex items-center justify-center"><span className="absolute inset-x-0 h-px bg-[#dce2f3]" /><span className="relative bg-white px-2 text-sm leading-5 text-[#6f7973]">Hoặc đăng ký bằng</span></div>
        <div className="flex gap-4"><button type="button" className="flex h-12 flex-1 items-center justify-center gap-2 rounded-xl border border-[#bec9c2] text-[12px] font-medium"><FigmaIcon src={authAssets.register.google} className="h-3 w-5" />Google</button><button type="button" className="flex h-12 flex-1 items-center justify-center gap-2 rounded-xl border border-[#bec9c2] text-[12px] font-medium"><FigmaIcon src={authAssets.register.apple} className="h-[22px] w-[15px]" />Apple</button></div>
        <p className="pt-8 text-center text-sm leading-5 text-[#3f4944]">Đã có tài khoản? <Link href="/login" className="text-[#004532]">Đăng nhập ngay</Link></p>
      </div>
    </main>
  );
}
