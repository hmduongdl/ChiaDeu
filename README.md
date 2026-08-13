# Chia Đều — Cash Flow Minimizer

*Kế hoạch dự án · Schema · Backend Go · Frontend Next.js PWA · Tích hợp SePay + PayOS + MoMo · Deploy Vercel*

---

## 1. Tổng quan dự án

Ứng dụng chia tiền nhóm kiểu Splitwise, với điểm khác biệt: thay vì nhập tay từng khoản chi, người dùng liên kết tài khoản ngân hàng qua SePay để đồng bộ giao dịch tự động, sau đó chọn giao dịch cần chia và hệ thống tự tính toán số tiền tối thiểu cần thanh toán giữa các thành viên.

### Bài toán

Một nhóm bạn đi du lịch/ăn uống chung. Mỗi người trả tiền hộ vài lần cho cả nhóm. Cuối cùng, ai nợ ai bao nhiêu tiền để hoàn lại công bằng?

Vấn đề thực tế: nếu tính thủ công, số giao dịch hoàn tiền sẽ rất nhiều và lộn xộn. Ví dụ:

- A nợ B: 100k
- B nợ C: 150k
- C nợ A: 50k

Nếu chuyển từng khoản riêng lẻ → 3 giao dịch. Nhưng thực ra có thể rút gọn xuống còn 1-2 giao dịch mà vẫn đúng số tiền mỗi người cần nhận/trả. Đây chính là bài toán **Cash Flow Minimization** — tối thiểu hóa số lượng giao dịch thanh toán.

### Luồng sử dụng chính

- Người dùng liên kết tài khoản ngân hàng qua SePay (webhook báo biến động số dư real-time).
- Mỗi giao dịch ngân hàng phát sinh được lưu vào hệ thống qua webhook của SePay.
- Người dùng mở app, chọn một giao dịch ngân hàng muốn chia cho nhóm bạn.
- Chọn nhóm và các thành viên tham gia chia tiền — hệ thống tạo Expense liên kết với giao dịch gốc.
- Thuật toán Minimize Cash Flow tính ra số giao dịch thanh toán tối thiểu cần thực hiện.
- Mỗi giao dịch thanh toán được sinh QR code / payment link (qua PayOS hoặc MoMo).
- Khi người nợ thanh toán, webhook xác nhận tự động cập nhật trạng thái Settlement.

---

## 2. Lựa chọn công nghệ

| Thành phần | Công nghệ | Lý do chọn |
|---|---|---|
| Frontend | Next.js (React + TypeScript) + Tailwind CSS, PWA | Một codebase dùng cho cả web và mobile, cài lên màn hình chính như app thật, deploy Vercel miễn phí |
| Backend API | Go (Fiber) | Hiệu năng cao, single binary deploy, xử lý webhook tốt, toàn bộ hệ thống dùng chung một ngôn ngữ |
| Core Algorithm | Go (`container/heap`) | Thuật toán Minimize Cash Flow cài đặt trực tiếp trong Go, dùng Max-Heap với Greedy, O(N log N), không cần cgo hay external dependency |
| Database | PostgreSQL | Quan hệ dữ liệu rõ ràng giữa Users – Groups – Expenses – Settlements, hỗ trợ transaction an toàn |
| Thanh toán | SePay + PayOS + MoMo (sandbox) | Đa cổng thanh toán — mỗi cổng đảm nhiệm một vai trò riêng |

---

## 3. Kiến trúc hệ thống

```
┌──────────────────────────┐
│  Next.js PWA (Vercel)     │  UI, chọn giao dịch, xem balance, quét QR
└───────────┬────────────────┘
            │ REST / JSON API
┌───────────▼────────────────┐
│      Go Backend (Fiber)     │  REST API, Auth, xử lý webhook
│                             │  Minimize Cash Flow (container/heap)
└──────┬───────────┬──────────┘
       │           │
┌──────▼───┐  ┌────▼─────────────────┐
│ Postgres │  │  SePay / PayOS / MoMo │
│   (DB)   │  │  Webhooks             │
└──────────┘  └──────────────────────┘
```

Điểm khác biệt so với thiết kế ban đầu: **toàn bộ backend là Go thuần**, không cần C++ shared library hay cgo bridge. Thuật toán Minimize Cash Flow được cài đặt trực tiếp trong Go package `algo/`, dùng `container/heap` của standard library. Cách này giúp codebase đồng nhất, build đơn giản (single binary), deploy dễ dàng.

### Luồng authentication

Frontend gửi login tới proxy same-origin `/api/auth/login` với `credentials: include`. Backend kiểm tra bcrypt rồi đặt access JWT 15 phút trong cookie HttpOnly/Secure/SameSite=Lax và refresh JWT 7 ngày trong cookie HttpOnly/Secure/SameSite=Strict chỉ dành cho `/api/auth/refresh`; token không được lưu trong Web Storage hoặc đọc bởi JavaScript. Khi API trả `401`, fetch client chỉ gọi refresh một lần cho các request đồng thời, nhận access cookie mới rồi retry request gốc. Nếu refresh thất bại, Zustand xóa user và chuyển về `/login`. Logout gọi `/api/auth/logout`, backend hết hạn cả hai cookie, sau đó frontend xóa auth state.

Route middleware chỉ kiểm tra sự hiện diện của access cookie để redirect sớm; protected layout luôn gọi `/api/auth/me` để backend xác minh chữ ký, thời hạn và user thật. Cấu hình production cần hai JWT secret độc lập dài ít nhất 32 ký tự, HTTPS (`COOKIE_SECURE=true`), một `FRONTEND_ORIGIN` cụ thể, và nên giữ browser API URL là `/api` để Next middleware và Go backend cùng nhận cookie. Xem `.env.example` cho toàn bộ biến môi trường.

Refresh JWT hiện là stateless và không được rotate/revoke phía server; logout xóa cookie phía browser nhưng không vô hiệu hóa ngay một token đã bị sao chép. Trước khi cần quản lý session theo thiết bị hoặc thu hồi tức thời, bổ sung `jti` cùng session store/revocation list.

---

## 4. Database Schema (PostgreSQL)

### 4.1 `users` — thông tin người dùng

| Cột | Kiểu | Ghi chú |
|---|---|---|
| id | UUID (PK) | `gen_random_uuid()` |
| name | VARCHAR(100) | Tên hiển thị |
| phone | VARCHAR(20) | Unique, dùng để mời/tìm bạn |
| email | VARCHAR(100) | Unique |
| bank_account_no | VARCHAR(50) | Số tài khoản liên kết SePay |
| bank_code | VARCHAR(20) | Mã ngân hàng (VD: MBBank, VCB...) |
| sepay_account_id | VARCHAR(50) | ID tài khoản đã đăng ký trên SePay |
| avatar_url | TEXT | |
| created_at | TIMESTAMPTZ | DEFAULT now() |

### 4.2 `groups` — nhóm chi tiêu

| Cột | Kiểu | Ghi chú |
|---|---|---|
| id | UUID (PK) | |
| name | VARCHAR(100) | VD: "Đà Lạt trip" |
| share_code | VARCHAR(10) | Unique, mã mời ngắn để join nhóm |
| created_by | UUID (FK users) | |
| currency | VARCHAR(10) | Mặc định VND |
| created_at | TIMESTAMPTZ | |

### 4.3 `group_members` — thành viên nhóm

| Cột | Kiểu | Ghi chú |
|---|---|---|
| group_id | UUID (FK groups) | PK kép với user_id |
| user_id | UUID (FK users) | |
| joined_at | TIMESTAMPTZ | |

### 4.4 `bank_transactions` — giao dịch đồng bộ từ SePay

Bảng này lưu toàn bộ giao dịch ngân hàng của người dùng, nhận qua webhook SePay. Đây là nguồn dữ liệu để người dùng chọn khi tạo Expense, thay vì nhập tay.

| Cột | Kiểu | Ghi chú |
|---|---|---|
| id | UUID (PK) | |
| user_id | UUID (FK users) | Chủ tài khoản phát sinh giao dịch |
| sepay_transaction_id | VARCHAR(50) | ID giao dịch từ SePay, unique |
| amount | NUMERIC(14,2) | Số tiền giao dịch |
| transaction_type | VARCHAR(10) | IN hoặc OUT |
| description | TEXT | Nội dung chuyển khoản gốc |
| bank_account_no | VARCHAR(50) | |
| is_used | BOOLEAN | Đã được gán vào 1 expense hay chưa |
| transaction_time | TIMESTAMPTZ | Thời điểm giao dịch thực tế |
| received_at | TIMESTAMPTZ | Thời điểm webhook nhận được |

### 4.5 `expenses` — khoản chi được chia trong nhóm

| Cột | Kiểu | Ghi chú |
|---|---|---|
| id | UUID (PK) | |
| group_id | UUID (FK groups) | |
| source_transaction_id | UUID (FK bank_transactions) | Giao dịch ngân hàng gốc được chọn, có thể NULL nếu nhập tay |
| paid_by | UUID (FK users) | |
| description | VARCHAR(255) | |
| amount | NUMERIC(14,2) | |
| split_type | VARCHAR(20) | EQUAL / PERCENT / CUSTOM |
| created_at | TIMESTAMPTZ | |

### 4.6 `expense_splits` — chi tiết chia tiền từng người

| Cột | Kiểu | Ghi chú |
|---|---|---|
| id | UUID (PK) | |
| expense_id | UUID (FK expenses) | |
| user_id | UUID (FK users) | |
| share_amount | NUMERIC(14,2) | Số tiền người này phải chịu |

### 4.7 `settlements` — giao dịch thanh toán tối ưu

| Cột | Kiểu | Ghi chú |
|---|---|---|
| id | UUID (PK) | |
| group_id | UUID (FK groups) | |
| from_user | UUID (FK users) | Người phải trả |
| to_user | UUID (FK users) | Người được nhận |
| amount | NUMERIC(14,2) | |
| status | VARCHAR(20) | PENDING / PAID / CANCELLED |
| payment_method | VARCHAR(20) | PAYOS_QR / MOMO / SEPAY_TRANSFER / CASH |
| qr_code_data | TEXT | Payload QR hoặc payment link |
| confirmed_transaction_id | UUID (FK bank_transactions) | Giao dịch xác nhận thanh toán, khớp qua webhook |
| paid_at | TIMESTAMPTZ | |
| created_at | TIMESTAMPTZ | |

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    phone VARCHAR(20) UNIQUE,
    email VARCHAR(100) UNIQUE,
    bank_account_no VARCHAR(50),
    bank_code VARCHAR(20),
    sepay_account_id VARCHAR(50),
    avatar_url TEXT,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    share_code VARCHAR(10) UNIQUE NOT NULL,
    created_by UUID REFERENCES users(id),
    currency VARCHAR(10) DEFAULT 'VND',
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE group_members (
    group_id UUID REFERENCES groups(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    joined_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (group_id, user_id)
);

CREATE TABLE bank_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    sepay_transaction_id VARCHAR(50) UNIQUE,
    amount NUMERIC(14,2) NOT NULL,
    transaction_type VARCHAR(10) CHECK (transaction_type IN ('IN','OUT')),
    description TEXT,
    bank_account_no VARCHAR(50),
    is_used BOOLEAN DEFAULT false,
    transaction_time TIMESTAMPTZ,
    received_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE expenses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID REFERENCES groups(id) ON DELETE CASCADE,
    source_transaction_id UUID REFERENCES bank_transactions(id),
    paid_by UUID REFERENCES users(id),
    description VARCHAR(255),
    amount NUMERIC(14,2) NOT NULL,
    split_type VARCHAR(20) DEFAULT 'EQUAL',
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE expense_splits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    expense_id UUID REFERENCES expenses(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id),
    share_amount NUMERIC(14,2) NOT NULL
);

CREATE TABLE settlements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID REFERENCES groups(id),
    from_user UUID REFERENCES users(id),
    to_user UUID REFERENCES users(id),
    amount NUMERIC(14,2) NOT NULL,
    status VARCHAR(20) DEFAULT 'PENDING',
    payment_method VARCHAR(20),
    qr_code_data TEXT,
    confirmed_transaction_id UUID REFERENCES bank_transactions(id),
    paid_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_expenses_group ON expenses(group_id);
CREATE INDEX idx_settlements_group ON settlements(group_id);
CREATE INDEX idx_groups_share_code ON groups(share_code);
CREATE INDEX idx_bank_tx_user_unused ON bank_transactions(user_id, is_used);
```

---

## 5. API Endpoints

| Method & Path | Chức năng |
|---|---|
| `POST /api/auth/link-bank` | Liên kết tài khoản ngân hàng với SePay |
| `POST /api/webhooks/sepay` | Nhận webhook giao dịch mới từ SePay |
| `GET /api/transactions?unused=true` | Lấy danh sách giao dịch ngân hàng chưa được gán vào Expense |
| `POST /api/groups` | Tạo nhóm mới, sinh share_code |
| `POST /api/groups/join/:shareCode` | Tham gia nhóm qua mã mời |
| `GET /api/groups/:id` | Thông tin nhóm và thành viên |
| `POST /api/groups/:id/expenses` | Tạo Expense — kèm source_transaction_id nếu chọn từ giao dịch ngân hàng |
| `GET /api/groups/:id/balances` | Tính net balance của từng thành viên |
| `POST /api/groups/:id/settle` | Gọi thuật toán Minimize Cash Flow, trả về danh sách settlements |
| `POST /api/settlements/:id/qr` | Sinh QR / payment link thanh toán (PayOS hoặc MoMo) |
| `POST /api/webhooks/payos` | Webhook xác nhận thanh toán qua PayOS |
| `POST /api/webhooks/momo` | Webhook xác nhận thanh toán qua MoMo (sandbox) |
| `GET /api/settlements/:id/status` | Kiểm tra trạng thái thanh toán |

---

## 6. Thuật toán cốt lõi — Minimize Cash Flow

Dùng Greedy kết hợp Max-Heap để tối thiểu hoá số lượng giao dịch thanh toán. Toàn bộ cài đặt bằng Go thuần (`container/heap`), không phụ thuộc ngoài.

### Các bước xử lý

1. Tính net balance từng người: tổng đã trả hộ trừ tổng phải trả theo phần chia.
2. Đưa các số dư dương (creditor) vào Max-Heap, số dư âm (debtor) vào Max-Heap theo trị tuyệt đối.
3. Lặp lại: lấy creditor và debtor lớn nhất, cho khớp một khoản = `min(hai giá trị)`.
4. Cập nhật lại số dư, loại người về 0 khỏi heap, tiếp tục đến khi hết.
5. Độ phức tạp: **O(N log N)** nhờ dùng Heap thay vì tìm max tuyến tính mỗi vòng lặp.

### Code (Go)

```go
package algo

import "container/heap"

type Balance struct {
    UserID string
    Amount float64 // dương = được nhận, âm = đang nợ
}

type Settlement struct {
    From   string
    To     string
    Amount float64
}

func MinimizeCashFlow(balances []Balance) []Settlement {
    // 1. Tách thành 2 Max-Heap: creditors (dương) và debtors (âm)
    // 2. Lặp: lấy max creditor + max debtor
    //    settleAmount = min(creditor, |debtor|)
    //    tạo Settlement{debtor, creditor, settleAmount}
    //    cập nhật lại heap, loại người có balance = 0
    // 3. Trả về danh sách settlements tối thiểu
}
```

Vị trí trong codebase: `backend/algo/minimize_cash_flow.go`. Package `algo` được import trực tiếp từ các handler trong `internal/handlers/`, không cần cgo bridge.

---

## 7. Tích hợp thanh toán đa cổng (SePay + PayOS + MoMo)

MoMo Business API (M4B) yêu cầu tài khoản doanh nghiệp với mã số thuế, không phù hợp để đăng ký trực tiếp cho dự án cá nhân/sinh viên. Vì vậy hệ thống tách vai trò rõ ràng giữa 3 cổng.

### 7.1 Vai trò từng cổng

| Cổng | Vai trò trong hệ thống | Lý do |
|---|---|---|
| **SePay** | Đồng bộ lịch sử giao dịch ngân hàng (luồng chọn giao dịch để tạo Expense) | Đăng ký cá nhân dễ dàng, webhook realtime, đúng thế mạnh theo dõi biến động số dư |
| **PayOS** | Sinh QR / payment link cho từng Settlement khi cần thanh toán chéo giữa các thành viên | Không yêu cầu người nhận có tài khoản SePay riêng, chỉ cần số tài khoản ngân hàng để tạo QR, có sandbox test miễn phí |
| **MoMo (Sandbox)** | Lựa chọn thanh toán phụ, hiển thị trong UI như một phương thức thay thế | Tích hợp được ở môi trường Test mà không cần tài khoản Business, đủ để demo đa dạng phương thức thanh toán |

### 7.2 Luồng đồng bộ giao dịch (SePay)

- Người dùng liên kết tài khoản ngân hàng ngay trong bước onboarding — nhập số tài khoản và ngân hàng, hệ thống gọi API SePay để đăng ký theo dõi.
- SePay gửi webhook mỗi khi có biến động số dư — hệ thống lưu vào bảng `bank_transactions` với `is_used = false`.
- Trong giao diện tạo Expense, người dùng thấy danh sách giao dịch `is_used = false`, chọn một giao dịch để gán làm khoản chi cần chia.

### 7.3 Luồng thanh toán Settlement (PayOS chính, MoMo phụ)

- Khi Settlement được tạo, hệ thống gọi PayOS để sinh QR code/payment link tương ứng với số tiền cần trả.
- Người dùng có thể chọn thanh toán qua MoMo sandbox nếu muốn demo phương thức thay thế — sinh QR/deeplink MoMo tương ứng.
- Webhook PayOS (và MoMo ở môi trường test) báo về khi thanh toán thành công — hệ thống cập nhật Settlement sang trạng thái PAID.
- Song song đó, nếu giao dịch chuyển khoản thật đi qua ngân hàng đã liên kết SePay, webhook SePay cũng có thể đối chiếu số tiền + nội dung để xác nhận chéo, tăng độ tin cậy.

### 7.4 Ghi chú triển khai thực tế

- Giai đoạn demo/đồ án: dùng PayOS thật (có sandbox miễn phí) + MoMo sandbox — không cần giấy phép kinh doanh.
- Nếu sau này muốn vận hành thật với MoMo, cần đăng ký tài khoản Business (M4B) qua doanh nghiệp.
- SePay có gói miễn phí giới hạn số lượng giao dịch/tháng — đủ dùng cho quy mô demo nhóm nhỏ.

---

## 8. Deploy

### Vercel Services — Frontend + Backend

File `vercel.json` ở thư mục gốc khai báo hai service và dùng chung một domain:

- Next.js frontend phục vụ tất cả các URL thông thường.
- Go backend phục vụ dưới prefix `/api/backend`.
- Ví dụ health check production: `GET /api/backend/health`.

Các bước deploy:

1. Push toàn bộ repository lên GitHub và import repository vào Vercel.
2. Trong **Build and Deployment Settings**, chọn Framework Preset là **Services** và giữ Root Directory ở thư mục gốc repository.
3. Khai báo các biến môi trường backend cần dùng, ví dụ `DATABASE_URL`, trong Vercel Dashboard.
4. Deploy. Vercel sẽ build `frontend` và `backend` thành hai service trong cùng deployment.

Có thể chạy cấu hình Vercel Services trên máy bằng `vercel dev`. Khi chỉ chạy Next.js bằng `npm run dev`, rewrite trong `frontend/next.config.js` sẽ chuyển `/api/backend/*` đến backend ở `http://localhost:8080`.

### Local Development

```bash
# Khởi động PostgreSQL
docker-compose up -d postgres

# Backend
cd backend && go run .

# Frontend
cd frontend && npm install && npm run dev
```

---

## 9. Roadmap thực hiện

| Giai đoạn | Nội dung |
|---|---|
| Tuần 1 | Setup database schema, khởi tạo Go backend (CRUD groups, users, expenses cơ bản) |
| Tuần 2 | Viết thuật toán Minimize Cash Flow bằng Go (`container/heap`), tích hợp vào handler |
| Tuần 3 | Tích hợp webhook SePay, xử lý và lưu bank_transactions |
| Tuần 4 | Frontend Next.js: tạo nhóm, chọn giao dịch ngân hàng để tạo expense, xem balance |
| Tuần 5 | Sinh QR thanh toán cho settlement (PayOS + MoMo), xác nhận tự động qua webhook |
| Tuần 6 | PWA polish, testing, viết báo cáo DSA (chứng minh độ phức tạp thuật toán) |

---
