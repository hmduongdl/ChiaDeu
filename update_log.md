# Update log

## 2026-08-12

- Added Figma-based authentication routes:
  - `/login` — Emerald Minimalist login screen (node `25:795`).
  - `/register` — account creation screen (node `25:848`).
  - `/forgot-password` — password recovery screen (node `25:917`).
- Added shared auth field and Figma asset helpers under `frontend/src/components/auth/`.
- Added password visibility toggle on login and registration forms.
- Added navigation links between login, registration, and password recovery screens.
- Bottom Nav is hidden on all three auth routes via `usePathname`, so it does not overlap the Figma auth layouts.
- Forms currently prevent submission while backend authentication endpoints are not implemented.
- Figma SVG assets are referenced through the connector's temporary URLs for this implementation preview; they should be downloaded into `frontend/public` before a production release.

## 2026-08-12 — AI workflow rule

- Added root `AGENTS.md` with a mandatory rule requiring every AI-assisted code, configuration, dependency, documentation, or tooling change to append an entry to `update_log.md`.
- The rule requires the date, summary, changed paths, validation, and known limitations/follow-up notes.
- Validation: verified the rule file and this log entry are both present.

## 2026-08-13 — JWT cookie authentication

- Implemented PostgreSQL-backed registration and login with bcrypt password hashing, typed HS256 access/refresh JWTs, strict environment validation, secure HttpOnly cookie issuance/deletion, access-token middleware, `/api/auth/register`, `/login`, `/refresh`, `/logout`, and `/me`, credentialed single-origin CORS, and auth protection for the existing non-webhook API routes.
- Added migration `backend/migrations/002_add_user_password.sql`, PostgreSQL/crypto/JWT dependencies, auth/config/handler/middleware packages under `backend/internal/`, and backend tests covering password hashing/validation, token type and expiry enforcement, cookie security attributes, refresh/logout, unauthorized JSON, and CORS.
- Added the Zustand auth store, shared fetch client with single-flight refresh and original-request retry, Next route middleware, protected layout/loading/redirect behavior, a protected dashboard/logout screen, and working login/registration forms. Browser API traffic now uses the same-origin `/api` proxy so both Next middleware and Go receive the cookies; no token is stored in Web Storage or logged.
- Added root `.env.example`, Docker Compose JWT/origin/cookie/proxy configuration, standalone Next Docker output, and an authentication flow/maintenance note in `README.md`.
- Main changed paths: `.env.example`, `README.md`, `backend/main.go`, `backend/internal/`, `backend/migrations/`, `backend/go.mod`, `backend/go.sum`, `frontend/src/lib/`, `frontend/src/stores/`, `frontend/src/middleware.ts`, `frontend/src/app/(protected)/`, auth pages, frontend package manifests/config/Dockerfile, and `docker-compose.yml`.
- Validation: `go test ./...` passed; `npm run build` passed with compilation, static generation, and TypeScript checks; `git diff --check` passed. `npm run lint` could not run non-interactively because this repository has no ESLint configuration. Docker Compose/runtime database integration was not executed because Docker is unavailable in the environment.
- Known limitations/follow-up: refresh tokens are stateless and are not rotated or server-revoked, so add a `jti` session store/revocation list when device sessions or immediate revocation are required. Existing users with no `password_hash` need a password reset path before password login. Detailed rate limiting remains a TODO. `npm audit --omit=dev` reports two high-severity advisories in the existing Next.js 14/PostCSS dependency chain; the offered automatic fix upgrades to breaking Next.js 16 and should be handled as a dedicated migration before production deployment. Production must use independent random JWT secrets, HTTPS with `COOKIE_SECURE=true`, and an explicit frontend origin.

## 2026-08-13 — Các màn hình chính và dock điều hướng từ Figma

- Triển khai bốn màn hình mobile theo các frame Figma `Trang chủ` (`7:2`), `Nhóm (list)` (`25:678`), `Hoạt động` (`16:214`) và `Profile` (`16:2`) tại các route được bảo vệ `/dashboard`, `/groups`, `/activity` và `/profile`.
- Kết nối dock với route thật bằng Next `Link`, đồng bộ trạng thái active từ pathname, giữ animation pill và chỉ render dock tại đúng bốn route trên. Bổ sung tìm kiếm/lọc nhóm, đánh dấu hoạt động đã đọc, phản hồi lời mời, dữ liệu người dùng từ auth store và đăng xuất thật ở trang cá nhân.
- Mỗi route dùng dynamic import và loading skeleton riêng; kết quả build xác nhận từng route có bundle độc lập. Toàn bộ 55 avatar/icon Figma cần dùng đã được tải về `frontend/public/figma/`, không còn phụ thuộc URL asset tạm thời.
- Tệp/thư mục thay đổi chính: `frontend/src/app/(protected)/{dashboard,groups,activity,profile}/`, `frontend/src/components/screens/`, `frontend/src/components/app/`, `frontend/src/components/BottomNavBar.tsx`, `frontend/src/middleware.ts`, `frontend/src/app/globals.css`, `frontend/public/figma/`, `update_log.md`.
- Kiểm tra: `npx tsc --noEmit --pretty false` đạt; `npm run build` đạt, tạo đủ route và bundle riêng; smoke test dev server xác nhận cả bốn route chưa đăng nhập trả `307` về login đúng redirect và asset Figma trả `200`; `git diff --check` đạt; xác nhận 55 asset Figma có dữ liệu trên đĩa.
- Giới hạn/theo dõi: số dư, danh sách nhóm và hoạt động hiện là dữ liệu trình diễn theo Figma; tìm kiếm/lọc và thao tác thông báo chạy ở client, chưa có API lưu bền. Các nút tạo nhóm, cài đặt và mục chi tiết profile mới dừng ở trạng thái giao diện cho tới khi có route/API tương ứng.
