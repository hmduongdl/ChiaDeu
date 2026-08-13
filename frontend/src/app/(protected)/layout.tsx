"use client";

import { useEffect, useState } from "react";
import { usePathname, useRouter } from "next/navigation";
import { useAuthStore } from "@/stores/auth-store";

export default function ProtectedLayout({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const pathname = usePathname();
  const user = useAuthStore((state) => state.user);
  const isLoading = useAuthStore((state) => state.isLoading);
  const fetchCurrentUser = useAuthStore((state) => state.fetchCurrentUser);
  const [hasCheckedSession, setHasCheckedSession] = useState(false);

  useEffect(() => {
    void fetchCurrentUser().finally(() => setHasCheckedSession(true));
  }, [fetchCurrentUser]);

  useEffect(() => {
    if (hasCheckedSession && !isLoading && !user) {
      router.replace(`/login?redirect=${encodeURIComponent(pathname)}`);
    }
  }, [hasCheckedSession, isLoading, pathname, router, user]);

  if (!hasCheckedSession || isLoading || !user) {
    return (
      <main className="flex min-h-screen items-center justify-center bg-[#f9f9ff] text-[#004532]">
        <div className="flex items-center gap-3" role="status" aria-live="polite">
          <span className="h-5 w-5 animate-spin rounded-full border-2 border-[#0d7a5f]/30 border-t-[#0d7a5f]" />
          <span className="text-sm font-medium">Đang xác thực...</span>
        </div>
      </main>
    );
  }

  return children;
}
