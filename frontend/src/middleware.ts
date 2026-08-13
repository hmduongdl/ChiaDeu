// Next.js middleware — chạy ở edge runtime trước mỗi request.
// Chức năng:
//   - Bảo vệ các route cần đăng nhập (/dashboard, /groups, /activity, /profile, ...)
//   - Nếu chưa có accessToken cookie → redirect về /login kèm tham số redirect
//   - Nếu đã đăng nhập mà vào /login → redirect về /dashboard
// Sử dụng matcher config để giới hạn route nào middleware được kích hoạt.
import { NextRequest, NextResponse } from "next/server";

const PROTECTED_ROUTES = [
  "/dashboard",
  "/groups",
  "/activity",
  "/profile",
  "/wallet",
  "/transactions",
  "/settings",
];

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;
  const hasAccessToken = Boolean(request.cookies.get("accessToken")?.value);
  const isProtectedRoute = PROTECTED_ROUTES.some(
    (route) => pathname === route || pathname.startsWith(`${route}/`),
  );

  if (isProtectedRoute && !hasAccessToken) {
    const loginURL = new URL("/login", request.url);
    loginURL.searchParams.set("redirect", pathname);
    return NextResponse.redirect(loginURL);
  }

  if (pathname === "/login" && hasAccessToken) {
    return NextResponse.redirect(new URL("/dashboard", request.url));
  }

  return NextResponse.next();
}

export const config = {
  matcher: [
    "/dashboard/:path*",
    "/groups/:path*",
    "/activity/:path*",
    "/profile/:path*",
    "/wallet/:path*",
    "/transactions/:path*",
    "/settings/:path*",
    "/login",
  ],
};
