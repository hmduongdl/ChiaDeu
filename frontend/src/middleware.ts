import { NextRequest, NextResponse } from "next/server";

const PROTECTED_ROUTES = ["/dashboard", "/wallet", "/transactions", "/settings"];

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
    "/wallet/:path*",
    "/transactions/:path*",
    "/settings/:path*",
    "/login",
  ],
};
