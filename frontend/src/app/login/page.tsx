"use client";

import Link from "next/link";
import { FormEvent, useState } from "react";
import { AuthField } from "@/components/auth/AuthField";
import { authAssets, FigmaIcon } from "@/components/auth/AuthIcons";

export default function LoginPage() {
  const [showPassword, setShowPassword] = useState(false);
  const handleSubmit = (event: FormEvent<HTMLFormElement>) => event.preventDefault();

  return (
    <main className="min-h-screen bg-[#f9f9ff] px-5 py-[79px] text-[#151c27]">
      <div className="mx-auto flex w-full max-w-[448px] flex-col items-center">
        <header className="flex w-full flex-col items-center gap-2">
          <h1 className="text-center text-[26px] font-bold leading-8 tracking-[-0.65px] text-[#004532]">CHIADEU</h1>
          <p className="text-center text-base leading-6 text-[#3f4944]">Chào mừng bạn quay lại!</p>
        </header>

        <section className="mt-8 w-full rounded-[24px] border border-white/50 bg-white/70 p-[25px] shadow-[0_8px_32px_rgba(0,69,50,0.04)] backdrop-blur-[6px]">
          <form onSubmit={handleSubmit} className="flex flex-col gap-4">
            <AuthField id="login-identity" label="Email hoặc Số điện thoại" placeholder="Nhập email hoặc SĐT" autoComplete="username" />
            <div className="relative">
              <AuthField id="login-password" label="Mật khẩu" type={showPassword ? "text" : "password"} placeholder="••••••••" autoComplete="current-password" />
              <button type="button" aria-label={showPassword ? "Ẩn mật khẩu" : "Hiện mật khẩu"} onClick={() => setShowPassword((value) => !value)} className="absolute right-3 top-7 flex h-8 w-8 items-center justify-center">
                <FigmaIcon src={authAssets.login.eye} className="h-[19.8px] w-[22px]" />
              </button>
            </div>
            <div className="flex justify-end pb-3 pt-1"><Link href="/forgot-password" className="text-[12px] font-medium leading-4 text-[#004532]">Quên mật khẩu?</Link></div>
            <button type="submit" className="flex h-12 items-center justify-center gap-2 rounded-2xl bg-gradient-to-r from-[#004532] to-[#065f46] text-[12px] font-semibold leading-4 text-white shadow-[0_4px_6px_rgba(0,69,50,0.2)]">Đăng nhập <FigmaIcon src={authAssets.login.arrow} className="h-3 w-3" /></button>
          </form>
          <div className="flex items-center py-2 text-[12px] leading-4 text-[#3f4944]"><span className="h-px flex-1 bg-[#bec9c2]/50" /><span className="px-4">Hoặc đăng nhập bằng</span><span className="h-px flex-1 bg-[#bec9c2]/50" /></div>
          <div className="flex justify-center gap-4 pt-2">
            <button type="button" aria-label="Đăng nhập bằng Google" className="flex h-12 w-12 items-center justify-center rounded-2xl border border-[#bec9c2]/50 bg-[#f9f9ff] shadow-sm"><FigmaIcon src={authAssets.login.google} className="h-4 w-5" /></button>
            <button type="button" aria-label="Đăng nhập bằng Facebook" className="flex h-12 w-12 items-center justify-center rounded-2xl bg-[#1877f2] shadow-sm"><FigmaIcon src={authAssets.login.facebook} className="h-[18px] w-[18px]" /></button>
            <button type="button" aria-label="Đăng nhập bằng Apple" className="flex h-12 w-12 items-center justify-center rounded-2xl bg-black shadow-sm"><FigmaIcon src={authAssets.login.apple} className="h-4 w-4" /></button>
          </div>
        </section>
        <p className="pt-12 text-center text-sm leading-5 text-[#3f4944]">Chưa có tài khoản? <Link href="/register" className="text-[12px] font-semibold text-[#004532]">Đăng ký ngay</Link></p>
      </div>
    </main>
  );
}
