"use client";

import Link from "next/link";
import { FormEvent } from "react";
import { AuthField } from "@/components/auth/AuthField";
import { authAssets, FigmaIcon } from "@/components/auth/AuthIcons";

export default function ForgotPasswordPage() {
  const handleSubmit = (event: FormEvent<HTMLFormElement>) => event.preventDefault();
  return (
    <main className="min-h-screen bg-[#f9f9ff] text-[#151c27]">
      <header className="flex h-16 items-center justify-between border border-[#6f7973]/10 bg-[#f9f9ff] px-5 shadow-sm">
        <Link href="/login" aria-label="Quay lại" className="flex h-10 w-10 items-center justify-center"><FigmaIcon src={authAssets.forgot.back} className="h-4 w-4" /></Link>
        <h1 className="pr-10 text-center text-xl font-semibold leading-7">Quên mật khẩu</h1><span className="w-10" />
      </header>
      <section className="mx-auto max-w-[448px] px-5 pt-8">
        <div className="flex justify-center"><div className="flex h-24 w-24 items-center justify-center rounded-full bg-[#065f46] shadow-[0_4px_16px_rgba(6,95,70,0.15)]"><FigmaIcon src={authAssets.forgot.key} className="h-6 w-11" /></div></div>
        <div className="pt-6 text-center"><h2 className="text-[26px] font-bold leading-8">Khôi phục mật khẩu</h2><p className="mx-auto max-w-[280px] pt-2 text-base leading-6 text-[#3f4944]">Nhập email liên kết với tài khoản của bạn để nhận mã xác thực</p></div>
        <form onSubmit={handleSubmit} className="mt-8 flex flex-col gap-4"><AuthField id="forgot-email" label="Email" placeholder="ví dụ: user@example.com" type="email" icon={authAssets.forgot.email} iconClassName="h-4 w-5" /><button type="submit" className="mt-3 h-12 rounded-xl bg-gradient-to-r from-[#004532] to-[#065f46] text-xl font-semibold leading-7 text-white shadow-[0_4px_6px_rgba(0,69,50,0.2)]">Gửi mã xác nhận</button></form>
        <div className="flex justify-center pt-8"><Link href="/login" className="flex items-center gap-1 text-[12px] font-medium leading-4 text-[#004532]"><FigmaIcon src={authAssets.forgot.arrow} className="h-[13px] w-2" />Quay lại Đăng nhập</Link></div>
      </section>
    </main>
  );
}
