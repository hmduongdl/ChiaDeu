// Zustand store quản lý trạng thái xác thực người dùng.
// Cung cấp:
//   - user: thông tin người dùng hiện tại (null nếu chưa đăng nhập)
//   - isLoading: đang kiểm tra/xử lý xác thực
//   - isAuthenticated: đã xác thực thành công
//   - login(email, password): gọi API đăng nhập, lưu user vào store
//   - register(name, email, password): gọi API đăng ký
//   - logout(): gọi API đăng xuất, reset toàn bộ state
//   - fetchCurrentUser(): gọi GET /auth/me để kiểm tra session
// Khi nhận 401 từ API (token hết hạn), tự động redirect về /login.
"use client";

import { create } from "zustand";
import { apiFetch, setUnauthorizedHandler } from "@/lib/api-client";

export type User = {
  id: string;
  name: string;
  email: string;
  phone?: string;
  avatarUrl?: string;
  createdAt: string;
};

type UserResponse = { user: User };

type AuthState = {
  user: User | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (name: string, email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  fetchCurrentUser: () => Promise<void>;
};

const unauthenticatedState = {
  user: null,
  isLoading: false,
  isAuthenticated: false,
} as const;

export const useAuthStore = create<AuthState>((set) => ({
  ...unauthenticatedState,

  login: async (email, password) => {
    set({ isLoading: true });
    try {
      const { user } = await apiFetch<UserResponse>(
        "/auth/login",
        {
          method: "POST",
          body: JSON.stringify({ email, password }),
        },
        { skipAuthRefresh: true },
      );
      set({ user, isAuthenticated: true, isLoading: false });
    } catch (error) {
      set(unauthenticatedState);
      throw error;
    }
  },

  register: async (name, email, password) => {
    await apiFetch<UserResponse>(
      "/auth/register",
      {
        method: "POST",
        body: JSON.stringify({ name, email, password }),
      },
      { skipAuthRefresh: true },
    );
  },

  logout: async () => {
    try {
      await apiFetch<void>("/auth/logout", { method: "POST" }, { skipAuthRefresh: true });
    } finally {
      set(unauthenticatedState);
    }
  },

  fetchCurrentUser: async () => {
    set({ isLoading: true });
    try {
      const { user } = await apiFetch<UserResponse>("/auth/me");
      set({ user, isAuthenticated: true, isLoading: false });
    } catch {
      set(unauthenticatedState);
    }
  },
}));

if (typeof window !== "undefined") {
  setUnauthorizedHandler(() => {
    useAuthStore.setState(unauthenticatedState);
    const redirect = `${window.location.pathname}${window.location.search}`;
    if (window.location.pathname !== "/login") {
      window.location.assign(`/login?redirect=${encodeURIComponent(redirect)}`);
    }
  });
}
