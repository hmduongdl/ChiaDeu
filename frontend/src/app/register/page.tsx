"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { FormEvent, useState } from "react";
import { AuthField } from "@/components/auth/AuthField";
import { authAssets } from "@/components/auth/AuthIcons";
import { ApiError } from "@/lib/api-client";
import { useAuthStore } from "@/stores/auth-store";

export default function RegisterPage() {
  const router = useRouter();
  const register = useAuthStore((state) => state.register);
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState("");

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setError("");
    if (password !== confirmPassword) {
      setError("Mật khẩu xác nhận chưa khớp.");
      return;
    }

    setIsSubmitting(true);
    try {
      await register(name, email, password);
      router.replace("/login?registered=1");
    } catch (requestError) {
      if (requestError instanceof ApiError && requestError.status === 409) {
        setError("Email này đã được đăng ký.");
      } else if (requestError instanceof ApiError && requestError.status === 400) {
        setError("Vui lòng nhập email hợp lệ và mật khẩu từ 8 đến 72 ký tự.");
      } else {
        setError("Không thể tạo tài khoản lúc này. Vui lòng thử lại.");
      }
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <main className="flex min-h-screen items-center justify-center bg-[#f9f9ff] px-4 py-[69px]">
      <div className="w-full max-w-[448px] rounded-2xl bg-white p-6 shadow-[0_8px_30px_rgba(0,0,0,0.04)]">
        <header className="flex flex-col items-center gap-1">
          <h1 className="text-[26px] font-bold leading-8 tracking-[-0.65px] text-[#004532]">CHIADEU</h1>
          <h2 className="pt-1 text-center text-xl font-semibold leading-7">Tạo tài khoản mới</h2>
          <p className="text-center text-sm leading-5 text-[#3f4944]">Tham gia để cùng chia sẻ chi phí minh bạch</p>
        </header>
        <form onSubmit={handleSubmit} className="mt-8 flex flex-col gap-4">
          <AuthField id="register-name" label="Họ và tên" placeholder="Nhập họ và tên" icon={authAssets.register.person} autoComplete="name" value={name} onChange={(event) => setName(event.target.value)} maxLength={100} required />
          <AuthField id="register-email" label="Email" type="email" placeholder="Nhập email" icon={authAssets.register.email} iconClassName="h-4 w-5" autoComplete="email" value={email} onChange={(event) => setEmail(event.target.value)} maxLength={100} required />
          <div className="relative">
            <AuthField id="register-password" label="Mật khẩu" placeholder="Tạo mật khẩu" type={showPassword ? "text" : "password"} icon={authAssets.register.lock} autoComplete="new-password" value={password} onChange={(event) => setPassword(event.target.value)} minLength={8} maxLength={72} required />
            <button type="button" onClick={() => setShowPassword((value) => !value)} aria-label={showPassword ? "Ẩn mật khẩu" : "Hiện mật khẩu"} className="absolute right-3 top-7 h-8 w-8" />
          </div>
          <AuthField id="register-confirm" label="Xác nhận mật khẩu" placeholder="Nhập lại mật khẩu" type="password" icon={authAssets.register.confirm} autoComplete="new-password" value={confirmPassword} onChange={(event) => setConfirmPassword(event.target.value)} minLength={8} maxLength={72} required />
          {error && <p className="rounded-xl bg-red-50 px-4 py-3 text-sm text-red-700" role="alert">{error}</p>}
          <button type="submit" disabled={isSubmitting} className="h-12 rounded-xl bg-gradient-to-br from-[#004532] to-[#065f46] text-[12px] font-semibold leading-4 text-white shadow-sm disabled:cursor-not-allowed disabled:opacity-60">{isSubmitting ? "Đang tạo tài khoản..." : "Đăng ký"}</button>
        </form>
        <p className="pt-8 text-center text-sm leading-5 text-[#3f4944]">Đã có tài khoản? <Link href="/login" className="font-medium text-[#004532]">Đăng nhập ngay</Link></p>
      </div>
    </main>
  );
}
