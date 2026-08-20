# Update log

## 2026-08-16 — Backend: models, schema, nghiệp vụ nhóm/khoản chi/quyết toán, API

Triển khai toàn bộ roadmap backend theo Chế độ chia đều linh hoạt, tách theo feature
branch:

- **Domain models + calc** (`feature/backend-domain-models`): `backend/models/{user,group,expense,settlement}.go`,
  `backend/services/{expense_calc,settlement_calc}.go` — struct theo schema mục tiêu,
  `SumSplits`/`SplitEqual` (phần dư chia ổn định), `ValidateExpense`, `CalculateNetBalances`,
  `SimplifyDebts` (resolver xác định O(n log n)), table-driven tests theo phancong.md.
- **Migration schema** (`feature/backend-schema-migration`): `backend/migrations/003_target_schema.sql`
  chuyển schema prototype → integer money `_minor`, role/status membership,
  `settlement_batches`/`settlements`/`settlement_events`/`audit_logs`/`sessions`,
  partial unique index một kỳ `OPEN`/nhóm, FK tài chính `RESTRICT`.
- **Groups/Expenses** (`feature/backend-groups-expenses`): `internal/groups` (tạo nhóm +
  membership ADMIN trong transaction, join bằng share code, danh sách thành viên),
  `internal/expenses` (tạo/sửa khoản chi kèm phần chia, kiểm tra thành viên hoạt động,
  invariant tổng split), `internal/audit` (log append-only), `services.ValidationError`.
- **Settlements** (`feature/backend-settlements`): `internal/settlements` — chốt kỳ trong
  transaction (khóa khoản chi chưa chốt, tính balance, chạy resolver, sinh settlements,
  idempotency key), vòng đời `PENDING → AWAITING_CONFIRMATION → PAID`
  (`mark-sent`/`confirm`/`reject`, hoàn tất kỳ khi mọi settlement PAID), hủy kỳ khi chưa
  có giao dịch thanh toán, ghi `settlement_events` + audit.
- **Refresh token rotation** (`feature/backend-refresh-sessions`): refresh token mang claim
  `jti`, phiên lưu hash trong `sessions`, mỗi refresh thu hồi phiên cũ và cấp phiên mới,
  logout thu hồi phía server; vẫn giữ luồng stateless khi chưa cấu hình session store.
- **API + hardening** (`feature/backend-api`): `internal/handlers/app.go` triển khai contract
  API nhóm/khoản chi/balance/chốt kỳ/settlement, wire vào `main.go`, rate limit
  (fiber limiter, `RATE_LIMIT_MAX`, `AUTH_RATE_LIMIT_MAX`), test tích hợp handlers với
  fake stores cho toàn bộ luồng nghiệp vụ.

- Tệp/dir thay đổi chính: `backend/main.go`, `backend/main_test.go`,
  `backend/internal/{auth,config,groups,expenses,settlements,audit,handlers,middleware}`,
  `backend/models/`, `backend/services/`, `backend/migrations/003_target_schema.sql`,
  `.env.example`, `update_log.md`, `phancong.md`.
- Kiểm tra: `go build ./...`, `go vet ./...`, `go test ./...` pass toàn bộ; `git diff --check`
  không lỗi. Chưa chạy kiểm thử chạy thật với database PostgreSQL (không có Docker/DB
  trong môi trường) — SQL trong migration 003 và các store cần được xác minh bằng
  integration test thật trước production.
- Giới hạn/theo dõi: chưa có endpoint lấy danh sách nhóm của user qua API (`GET /groups`);
  chưa triển khai `VOID` (hủy) khoản chi riêng lẻ; webhook ngân hàng vẫn là stub;
  vòng đời settlement chưa có màn hình dashboard nhóm phía frontend; `cancel` kỳ chưa xoá
  `batch_id` trên settlements đã đánh dấu sent (chỉ CANCELLED).

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

## 2026-08-13 — Multi-creditor-only product documentation

- Rewrote the product, schema, API target, business rules, permissions, settlement flow, roadmap, and beginner backend assignment around a single `MULTI_CREDITOR` model in which every active member can pay expenses.
- Removed the legacy fixed-recipient mode, mode selection, leadership transfer flow, leader-specific data fields, resolver, dashboard, and exercises from the target documentation.
- Defined a deterministic multi-creditor balance resolver, financial invariants, transaction boundaries, idempotency expectations, unified group permissions, and clearer separation between implemented authentication and planned business APIs.
- Main changed files: `README.md`, `schema.md`, `phancong.md`, and `update_log.md`.
- Validation: Markdown/reference consistency, stale terminology search, and `git diff --check` performed; application tests/build were not run because this change only updates documentation.
- Known limitations/follow-up: the existing `001_init.sql` migration and placeholder business routes still reflect an earlier prototype and must be replaced through a safe follow-up migration and implementation. The documented resolver prioritizes deterministic, bounded transaction reduction and does not claim a globally minimal solution for every balance combination.

## 2026-08-13 — Flexible splitting mode terminology

- Standardized the user-facing name of the `MULTI_CREDITOR` model as “Chế độ chia đều linh hoạt” across the product overview, target schema, API description, roadmap, and backend assignment.
- Main changed files: `README.md`, `schema.md`, `phancong.md`, and `update_log.md`.
- Validation: searched the repository documentation to confirm the superseded Vietnamese label is absent; `git diff --check` passed. Application tests/build were not run because this is a terminology-only documentation change.
- Known limitations/follow-up: `MULTI_CREDITOR` remains the internal technical identifier; UI copy should use “Chế độ chia đều linh hoạt” when the business screens are implemented.

## 2026-08-20 — Phân tích kiến trúc Backend

- Tạo file tài liệu `backend_analysis.md` phân tích toàn bộ cấu trúc thư mục, kiến trúc và công nghệ sử dụng của phần Backend (Golang, Fiber, PostgreSQL). Tài liệu cung cấp cái nhìn tổng quan về chức năng của từng thư mục (`internal`, `models`, `services`, `migrations`) và luồng hoạt động chính để hỗ trợ việc bảo trì và phát triển tiếp theo.
- Tệp/dir thay đổi chính: `backend_analysis.md` (tạo mới trong thư mục artifacts), `update_log.md`.
- Kiểm tra: Không yêu cầu build code do chỉ tạo tài liệu phân tích. Đã đọc và tổng hợp thông tin từ `go.mod`, `main.go` và cấu trúc thư mục thực tế của `backend`.
- Giới hạn/theo dõi: Phân tích dựa trên hiện trạng code, tài liệu này có thể cần cập nhật nếu kiến trúc backend có thay đổi lớn trong tương lai.

## 2026-08-20 — Phân tích thuật toán chia tiền và chốt công nợ

- Giải thích chi tiết hai thuật toán cốt lõi trong `services/`: thuật toán chia đều khoản chi (`SplitEqual` trong `expense_calc.go`) và thuật toán tối giản hóa công nợ (`SimplifyDebts` trong `settlement_calc.go`). Thuật toán xử lý tốt các trường hợp số lẻ (phân bổ phần dư) và sử dụng phương pháp tham lam (greedy) có sắp xếp để giảm thiểu số lượng giao dịch thanh toán chéo.
- Tệp/dir thay đổi chính: Không sửa code, chỉ tạo log và giải thích trên chat.
- Kiểm tra: Không yêu cầu build code.
- Giới hạn/theo dõi: Thuật toán SimplifyDebts hiện tại giải quyết bài toán theo hướng xác định (deterministic) O(N log N) để tối thiểu hóa số giao dịch cục bộ, không phải thuật toán tìm số giao dịch tuyệt đối nhỏ nhất (NP-Hard).

## 2026-08-20 — Cập nhật SplitEqual đảm bảo tính xác định

- Cập nhật thuật toán `SplitEqual` trong `services/expense_calc.go` để luôn sort `memberIDs` (theo chuẩn từ điển) trước khi phân bổ phần dư (số lẻ). Điều này giải quyết rủi ro kết quả chia bị thay đổi nếu thứ tự query mảng user từ DB trả về khác nhau giữa các lần chạy, đảm bảo hoàn toàn tính xác định (deterministic) cho thuật toán chia số lẻ. 
- Tệp/dir thay đổi chính: `backend/services/expense_calc.go`
- Kiểm tra: `go test ./...` chạy thành công toàn bộ bài test ở backend.
- Giới hạn/theo dõi: Không có.

## 2026-08-20 — Tối ưu zero-allocation cho SplitEqual

- Áp dụng thuật toán sắp xếp in-place trực tiếp trên mảng kết quả `splits` thay vì tạo mảng copy `memberIDs` mới để sort. Việc này giúp khử hoàn toàn chi phí khởi tạo memory (Zero Extra Allocation - O(1) bộ nhớ phụ) trong khi vẫn giữ độ phức tạp thời gian O(N log N) cho việc chia phần dư của hàm `SplitEqual`.
- Tệp/dir thay đổi chính: `backend/services/expense_calc.go`
- Kiểm tra: `go test ./...` đạt 100% (pass xanh).
- Giới hạn/theo dõi: Không có.

## 2026-08-20 — Refactor Backend sang Vercel Serverless Functions

- Tái cấu trúc backend từ Fiber server truyền thống sang Vercel Serverless Go Functions.
- Tạo gói helper dùng chung `backend/internal/vercel` chứa middleware CORS, Auth (JWT cookie), cấu hình database pool tối ưu (MaxConns = 2) cho môi trường serverless, và các helper response định dạng JSON chuẩn.
- Triển khai 16 file serverless handlers tương ứng các API dưới thư mục `backend/api/` (chia theo từng folder con độc lập để tránh xung đột package name khi compile cục bộ).
- Cấu hình lại `vercel.json` ở root để map các route cụ thể (specific) và động (dynamic) theo đúng thứ tự ưu tiên, kết nối cả frontend Next.js và các Serverless Functions của backend.
- Tệp/dir thay đổi chính: `vercel.json`, `backend/api/`, `backend/internal/vercel/`, `update_log.md`, `AGENTS.md`.
- Kiểm tra: `go build ./...` và `go test ./...` chạy thành công toàn bộ.
- Giới hạn/theo dõi: Cần thiết lập biến môi trường trên Vercel Dashboard trước khi deploy. Database PostgreSQL nên bật connection pooler (như Neon/Supabase Pooler) vì môi trường serverless sẽ sinh ra nhiều connection ngắn hạn.
