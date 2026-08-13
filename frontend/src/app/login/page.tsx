"use client";

import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { FormEvent, Suspense, useState } from "react";
import { AuthField } from "@/components/auth/AuthField";
import { authAssets, FigmaIcon } from "@/components/auth/AuthIcons";
import { ApiError } from "@/lib/api-client";
import { useAuthStore } from "@/stores/auth-store";

function LoginForm() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const login = useAuthStore((state) => state.login);
  const isLoading = useAuthStore((state) => state.isLoading);
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState("");

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setError("");

    try {
      await login(email, password);
      const requestedRedirect = searchParams.get("redirect") || "/dashboard";
      const redirect = requestedRedirect.startsWith("/") && !requestedRedirect.startsWith("//")
        ? requestedRedirect
        : "/dashboard";
      router.replace(redirect);
      router.refresh();
    } catch (requestError) {
      setError(
        requestError instanceof ApiError && requestError.status === 401
          ? "Email hoặc mật khẩu không đúng."
          : "Không thể đăng nhập lúc này. Vui lòng thử lại.",
      );
    }
  };

  return (
    <main className="min-h-screen bg-[#f9f9ff] px-5 py-[79px] text-[#151c27]">
      <div className="mx-auto flex w-full max-w-[448px] flex-col items-center">
        <header className="flex w-full flex-col items-center gap-2">
          <h1 className="text-center text-[26px] font-bold leading-8 tracking-[-0.65px] text-[#004532]">CHIADEU</h1>
          <p className="text-center text-base leading-6 text-[#3f4944]">Chào mừng bạn quay lại!</p>
        </header>

        <section className="mt-8 w-full rounded-[24px] border border-white/50 bg-white/70 p-[25px] shadow-[0_8px_32px_rgba(0,69,50,0.04)] backdrop-blur-[6px]">
          {searchParams.get("registered") === "1" && (
            <p className="mb-4 rounded-xl bg-emerald-50 px-4 py-3 text-sm text-emerald-800" role="status">
              Tạo tài khoản thành công. Bạn có thể đăng nhập ngay.
            </p>
          )}
          <form onSubmit={handleSubmit} className="flex flex-col gap-4">
            <AuthField
              id="login-email"
              label="Email"
              type="email"
              placeholder="Nhập email"
              autoComplete="email"
              value={email}
              onChange={(event) => setEmail(event.target.value)}
              required
            />
            <div className="relative">
              <AuthField
                id="login-password"
                label="Mật khẩu"
                type={showPassword ? "text" : "password"}
                placeholder="••••••••"
                autoComplete="current-password"
                value={password}
                onChange={(event) => setPassword(event.target.value)}
                required
              />
              <button type="button" aria-label={showPassword ? "Ẩn mật khẩu" : "Hiện mật khẩu"} onClick={() => setShowPassword((value) => !value)} className="absolute right-3 top-7 flex h-8 w-8 items-center justify-center">
                <FigmaIcon src={authAssets.login.eye} className="h-[19.8px] w-[22px]" />
              </button>
            </div>
            <div className="flex justify-end pb-3 pt-1"><Link href="/forgot-password" className="text-[12px] font-medium leading-4 text-[#004532]">Quên mật khẩu?</Link></div>
            {error && <p className="rounded-xl bg-red-50 px-4 py-3 text-sm text-red-700" role="alert">{error}</p>}
            <button type="submit" disabled={isLoading} className="flex h-12 items-center justify-center gap-2 rounded-2xl bg-gradient-to-r from-[#004532] to-[#065f46] text-[12px] font-semibold leading-4 text-white shadow-[0_4px_6px_rgba(0,69,50,0.2)] disabled:cursor-not-allowed disabled:opacity-60">
              {isLoading ? "Đang đăng nhập..." : "Đăng nhập"}
              {!isLoading && <FigmaIcon src={authAssets.login.arrow} className="h-3 w-3" />}
            </button>
          </form>
        </section>
        <p className="pt-12 text-center text-sm leading-5 text-[#3f4944]">Chưa có tài khoản? <Link href="/register" className="text-[12px] font-semibold text-[#004532]">Đăng ký ngay</Link></p>
      </div>
    </main>
  );
}

export default function LoginPage() {
  return (
    <Suspense fallback={<main className="min-h-screen bg-[#f9f9ff]" />}>
      <LoginForm />
    </Suspense>
  );
}
