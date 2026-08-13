"use client";

import { useRouter } from "next/navigation";
import { useAuthStore } from "@/stores/auth-store";

export default function DashboardPage() {
  const router = useRouter();
  const user = useAuthStore((state) => state.user);
  const logout = useAuthStore((state) => state.logout);

  const handleLogout = async () => {
    await logout();
    router.replace("/login");
    router.refresh();
  };

  return (
    <main className="min-h-screen bg-[#f9f9ff] px-5 py-12 text-[#151c27]">
      <section className="mx-auto max-w-3xl rounded-3xl border border-emerald-950/5 bg-white p-6 shadow-[0_8px_32px_rgba(0,69,50,0.06)]">
        <div className="flex items-center justify-between gap-4">
          <div>
            <p className="text-sm text-[#6f7973]">Xin chào</p>
            <h1 className="text-2xl font-bold text-[#004532]">{user?.name}</h1>
          </div>
          <button
            type="button"
            onClick={() => void handleLogout()}
            className="rounded-xl border border-[#0d7a5f]/30 px-4 py-2 text-sm font-semibold text-[#004532] transition hover:bg-emerald-50"
          >
            Đăng xuất
          </button>
        </div>
        <p className="mt-8 text-sm text-[#3f4944]">Dashboard tài chính của bạn đã được bảo vệ.</p>
      </section>
    </main>
  );
}
