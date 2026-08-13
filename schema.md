# ChiaDeu — Schema & Cấu trúc tổng thể dự án

> Tài liệu này giúp 2 bạn nắm được **bức tranh toàn cảnh** của cả dự án — kể cả phần các bạn không
> trực tiếp code (frontend) — để biết dữ liệu mình viết ra ở backend sẽ được ai dùng, dùng để làm gì.
> Không cần thuộc lòng, chỉ cần đọc hiểu tổng thể 1 lần.

---

## 1. Dự án này làm gì? (tóm tắt 3 câu)

ChiaDeu là app chia tiền nhóm bạn bè sau mỗi lần đi chơi, ăn uống, du lịch. App hỗ trợ 2 chế độ:
(1) **"chủ xị"** — 1 người trong nhóm đứng ra trả tiền trước, những người còn lại trả lại cho người đó;
(2) **"ai nợ ai"** — nhiều người có thể thay nhau ứng tiền, hệ thống tự động tính ai trả ai với số
giao dịch ít nhất (giống Splitwise).
Cả 2 chế độ dùng chung 1 tầng tính net balance, chỉ khác ở cách sinh danh sách thanh toán cuối cùng.
Mọi xác nhận thanh toán đều thực hiện thủ công trong app.
---

## 2. Sơ đồ tổng thể hệ thống

```
┌───────────────────────┐
│   FRONTEND (Next.js)  │   <- Giao diện người dùng bấm vào, xem trên điện thoại/máy tính
│   Người dùng thao tác │
└───────────┬───────────┘
            │ gửi/nhận dữ liệu qua mạng (dạng JSON)
┌───────────▼───────────┐
│   BACKEND (Go)        │   <- Phần 2 bạn phụ trách: xử lý logic, tính toán, lưu trữ
│   Xử lý logic, tính   │
│   toán, kiểm tra quyền│
└───────────┬───────────┘
            │ đọc/ghi dữ liệu
┌───────────▼───────────┐
│  DATABASE (PostgreSQL) │   <- Nơi lưu trữ dữ liệu vĩnh viễn (giống 1 cuốn sổ khổng lồ)
└────────────────────────┘
```

**Ví dụ cụ thể để dễ hình dung:** khi 1 người dùng bấm nút "Tạo khoản chi" trên điện thoại →
Frontend gửi thông tin đó (tên món, số tiền, chia cho ai) sang Backend → Backend kiểm tra hợp lệ,
tính toán từng người nợ bao nhiêu → Backend lưu kết quả vào Database → Backend trả kết quả về →
Frontend hiển thị lên màn hình cho người dùng thấy.

---

## 3. Database Schema — toàn bộ các "cuốn sổ" lưu dữ liệu

Mỗi bảng dưới đây giống như 1 cuốn sổ có nhiều cột, mỗi dòng là 1 bản ghi.

### 3.1 `users` — Sổ người dùng

| Cột | Kiểu | Ý nghĩa |
|---|---|---|
| id | UUID | Mã định danh duy nhất |
| name | text | Tên hiển thị |
| email | text | Email đăng nhập |
| password_hash | text | Mật khẩu đã mã hóa (không lưu mật khẩu gốc) |
| phone | text | Số điện thoại (tùy chọn) |

### 3.2 `groups` — Sổ nhóm

| Cột | Kiểu | Ý nghĩa |
|---|---|---|
| id | UUID | Mã nhóm |
| name | text | Tên nhóm (VD: "Team đi Đà Lạt") |
| created_by | UUID | Người tạo nhóm (ghi nhớ lịch sử, không đổi) |
| leader_id | UUID | Chủ xị **hiện tại** (chỉ dùng trong chế độ "chủ xị", có thể đổi qua thời gian) |
| settlement_mode | text | `SINGLE_CREDITOR` (chủ xị) hoặc `MULTI_CREDITOR` (ai nợ ai), chọn khi tạo nhóm |
| share_code | text | Mã mời để người khác tham gia nhóm |
| currency | text | Đơn vị tiền tệ, mặc định VND |

### 3.3 `group_members` — Sổ thành viên trong từng nhóm

| Cột | Kiểu | Ý nghĩa |
|---|---|---|
| group_id | UUID | Thuộc nhóm nào |
| user_id | UUID | Ai là thành viên |
| joined_at | timestamp | Tham gia lúc nào |

*Ghi chú: 1 user có thể ở nhiều nhóm, mỗi nhóm nhiều user — đây gọi là quan hệ "nhiều-nhiều".*

### 3.4 `expenses` — Sổ khoản chi

| Cột | Kiểu | Ý nghĩa |
|---|---|---|
| id | UUID | Mã khoản chi |
| group_id | UUID | Thuộc nhóm nào |
| paid_by | UUID | Ai đã trả khoản này (chủ xị trong chế độ "chủ xị", bất kỳ thành viên nào trong chế độ "ai nợ ai") |
| description | text | Mô tả (VD: "Ăn tối nhà hàng X") |
| amount_minor | integer | Tổng tiền, tính theo đơn vị nhỏ nhất (VND thì là đồng, KHÔNG dùng số thập phân) |
| split_type | text | Kiểu chia: `EQUAL` (đều) / `PERCENT` (%) / `WEIGHT` (trọng số) / `CUSTOM` (tùy chỉnh) |
| expense_date | date | Ngày phát sinh |

### 3.5 `expense_splits` — Sổ chia phần của từng khoản chi

| Cột | Kiểu | Ý nghĩa |
|---|---|---|
| id | UUID | Mã dòng |
| expense_id | UUID | Thuộc khoản chi nào |
| user_id | UUID | Người phải chịu phần này |
| share_minor | integer | Số tiền người đó phải chịu |
| settlement_batch_id | UUID (có thể trống) | Đã được gộp vào đợt chốt nợ nào chưa |

### 3.6 `settlement_batches` — Sổ đợt chốt nợ

| Cột | Kiểu | Ý nghĩa |
|---|---|---|
| id | UUID | Mã đợt chốt |
| group_id | UUID | Thuộc nhóm nào |
| leader_id | UUID | Chủ xị của đợt này (ghi lại, không đổi dù sau này đổi chủ xị; nullable trong chế độ "ai nợ ai") |
| status | text | `OPEN` (đang mở) / `COMPLETED` / `CANCELLED` |

### 3.7 `settlements` — Sổ khoản cần hoàn tiền

| Cột | Kiểu | Ý nghĩa |
|---|---|---|
| id | UUID | Mã khoản hoàn |
| batch_id | UUID | Thuộc đợt chốt nào |
| from_user_id | UUID | Ai phải trả |
| to_user_id | UUID | Nhận tiền từ ai (chủ xị trong chế độ "chủ xị", thành viên được nhận trong chế độ "ai nợ ai") |
| amount_minor | integer | Số tiền phải trả |
| payment_code | text | Mã riêng để nhận diện khi chuyển khoản, VD `CD4F82A9` |
| status | text | `PENDING` / `AWAITING_CONFIRMATION` / `PAID` / `CANCELLED` |
| paid_at | timestamp | Thời điểm xác nhận đã trả |

> **Ghi chú quan trọng:** `settlements.status` phản ánh vòng đời nghiệp vụ của khoản nợ
> (`PENDING` → `AWAITING_CONFIRMATION` → `PAID`), hoàn toàn độc lập với bất kỳ cổng thanh
> toán hay dịch vụ bên thứ ba nào. Mọi xác nhận đều do người dùng thực hiện thủ công
> trong app: thành viên báo "đã chuyển", người nhận xác nhận "đã nhận".

### 3.8 `audit_logs` — Sổ nhật ký hành động quan trọng

| Cột | Kiểu | Ý nghĩa |
|---|---|---|
| id | UUID | Mã dòng log |
| group_id | UUID | Thuộc nhóm nào |
| actor_id | UUID | Ai thực hiện hành động |
| action | text | Hành động gì (VD: "đổi chủ xị", "xác nhận thanh toán") |
| created_at | timestamp | Lúc nào |

**Sơ đồ quan hệ giữa các bảng (đơn giản hóa):**

```
users ──┬── group_members ──── groups
        │                        │
        │                        ├── expenses ──── expense_splits
        │                        │
        │                        └── settlement_batches ──── settlements
        │
        └── (1 user có thể xuất hiện ở nhiều bảng: là leader, là expense payer, là settlement payer...)
```

---

## 4. Cấu trúc thư mục toàn bộ dự án (bức tranh chung)

```
ChiaDeu/
├── backend/                    <-- 2 bạn phụ trách toàn bộ phần này
│   ├── models/                 <-- struct (class) tương ứng từng bảng ở mục 3
│   ├── services/                <-- các hàm xử lý logic, tính toán
│   ├── repository/               <-- các hàm đọc/ghi dữ liệu vào PostgreSQL
│   └── handler/ (giai đoạn sau)  <-- nơi nhận request từ frontend
│
├── frontend/                   <-- người khác/AI phụ trách, 2 bạn KHÔNG cần code
│   ├── app/                    <-- các trang màn hình (login, danh sách nhóm, chi tiêu...)
│   ├── components/             <-- các mảnh giao diện tái sử dụng (nút, thẻ, form...)
│   └── lib/                    <-- các hàm gọi sang backend để lấy/gửi dữ liệu
│
└── docker-compose.yml          <-- file khởi động database PostgreSQL để test local
```

**Điều quan trọng cần hiểu:** frontend và backend **không đọc trực tiếp code của nhau** — chúng
"nói chuyện" với nhau bằng cách gửi dữ liệu qua mạng theo 1 khuôn mẫu đã thống nhất trước (giống
như 2 người nói chuyện qua điện thoại theo 1 kịch bản có sẵn). Backend chỉ cần đảm bảo trả đúng
dữ liệu theo đúng khuôn, không cần biết frontend hiển thị đẹp xấu ra sao.

---

## 5. Cấu trúc thư mục Backend chi tiết (phần 2 bạn trực tiếp làm)

```
backend/
├── main.go                     <-- điểm bắt đầu chạy chương trình
├── models/
│   ├── user.go                 <-- struct User
│   ├── group.go                <-- struct Group, GroupMember
│   ├── expense.go               <-- struct Expense, ExpenseSplit
│   ├── settlement.go            <-- struct SettlementBatch, Settlement
│   └── audit.go                 <-- struct AuditLog
├── services/
│   ├── expense_calc.go          <-- hàm chia tiền (EQUAL/PERCENT/WEIGHT/CUSTOM)
│   └── settlement_calc.go       <-- hàm tính net balance, single-creditor resolver, min-cash-flow resolver
└── repository/ (giai đoạn sau, khi học kết nối database)
    ├── user_repo.go
    ├── group_repo.go
    └── ...
```

**Ánh xạ: mỗi bảng schema (mục 3) → 1 struct trong `models/`:**

| Bảng trong Database | Struct trong `models/` | File |
|---|---|---|
| `users` | `User` | `models/user.go` |
| `groups` | `Group` | `models/group.go` |
| `group_members` | `GroupMember` | `models/group.go` |
| `expenses` | `Expense` | `models/expense.go` |
| `expense_splits` | `ExpenseSplit` | `models/expense.go` |
| `settlement_batches` | `SettlementBatch` | `models/settlement.go` |
| `settlements` | `Settlement` | `models/settlement.go` |
| `audit_logs` | `AuditLog` | `models/audit.go` |

---

## 6. Cấu trúc thư mục Frontend (chỉ để tham khảo, KHÔNG cần code)

```
frontend/
├── app/
│   ├── (auth)/login, register         <-- màn hình đăng nhập/đăng ký
│   └── (protected)/
│       └── groups/[groupId]/
│           ├── page.tsx                <-- màn hình chính của 1 nhóm
│           ├── expenses/                <-- màn hình khoản chi
│           └── settlements/             <-- màn hình chốt nợ, thanh toán
├── components/
│   ├── group/LeaderDashboard.tsx        <-- giao diện dành cho chủ xị (chế độ "chủ xị")
│   ├── group/MultiCreditorDashboard.tsx  <-- giao diện dành cho chế độ "ai nợ ai"
│   └── group/MemberDashboard.tsx         <-- giao diện dành cho thành viên
└── lib/api/                             <-- nơi gọi sang backend lấy dữ liệu
```

**2 bạn chỉ cần nhớ 1 điều:** khi frontend gọi sang backend, dữ liệu trả về phải đúng khớp với
struct mà 2 bạn đã định nghĩa trong `models/` — vì vậy đặt tên field trong struct cần rõ ràng,
đúng ý nghĩa, để sau này việc nối 2 bên không bị lệch.

---

## 7. Vai trò người dùng chính trong app (để hiểu dữ liệu dùng vào việc gì)

| Vai trò | Họ thấy gì trên app | Dữ liệu backend cần cung cấp |
|---|---|---|
| **Chủ xị (Leader)** — chế độ "chủ xị" | Tạo khoản chi, chốt nợ, xác nhận ai đã trả | `Expense`, `SettlementBatch`, danh sách `Settlement` của cả nhóm |
| **Thành viên (Member)** — chế độ "chủ xị" | Chỉ xem mình nợ bao nhiêu, báo đã chuyển khoản | Chỉ `Settlement` liên quan tới chính họ (`from_user_id` = họ) |
| **Thành viên** — chế độ "ai nợ ai" | Tạo khoản chi (mình là `paid_by`), xem toàn bộ nợ nhóm, báo đã chuyển | `Expense`, toàn bộ `Settlement` của nhóm, net balance từng người |

Trong cả 2 chế độ, backend tính net balance giống nhau: `NetBalance = Tổng phần chia phải chịu - Tổng tiền đã ứng`. Sự khác biệt chỉ nằm ở bước sinh danh sách settlement: chế độ "chủ xị" tạo settlement hình sao về 1 người, chế độ "ai nợ ai" dùng min-cash-flow để tối giản giao dịch.

---

## 8. Lộ trình học tiếp theo (để 2 bạn hình dung đường đi)

```
Bước 1 (hiện tại): Viết struct (class) + method đơn giản              <-- đang làm
Bước 2: Kết nối struct với database PostgreSQL thật (repository)
Bước 3: Viết hàm xử lý logic phức tạp hơn (services):
        - Net balance engine (chung cho cả 2 chế độ)
        - Single-creditor resolver (hình sao về chủ xị)
        - Multi-creditor resolver (min-cash-flow, ai nợ ai)
Bước 4: Mở "cổng" để frontend gọi vào backend (API/handler)
```

Mỗi bước sẽ có tài liệu riêng, đơn giản hóa dần — không cần lo về Bước 4 lúc này.