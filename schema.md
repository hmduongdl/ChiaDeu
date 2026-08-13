# Chia Đều — schema cho Chế độ chia đều linh hoạt

Tài liệu này là nguồn tham chiếu cho nghiệp vụ nhóm, khoản chi và quyết toán. Sản phẩm chỉ
có Chế độ chia đều linh hoạt (`MULTI_CREDITOR`): nhiều thành viên có thể ứng tiền, sau đó
hệ thống sinh các giao dịch hoàn tiền giữa người nợ và người cần được nhận.

> `backend/migrations/001_init.sql` vẫn là migration thử nghiệm cũ. Tên cột, kiểu tiền và
> một số bảng trong migration đó chưa khớp schema mục tiêu dưới đây. Không sửa trực tiếp
> migration đã chạy ở môi trường có dữ liệu; hãy tạo migration kế tiếp để chuyển đổi.

## Quy ước chung

- Khóa chính dùng UUID do PostgreSQL sinh.
- Thời gian dùng `TIMESTAMPTZ` và lưu theo UTC.
- Tiền dùng `BIGINT` theo đơn vị nhỏ nhất, hậu tố `_minor`; VND dùng đơn vị đồng.
- Mọi bảng nghiệp vụ có `created_at`; bảng có thể sửa thêm `updated_at`.
- Không có `leader_id` hoặc `settlement_mode`. Quyền quản trị không quyết định dòng tiền.
- Foreign key tài chính ưu tiên `RESTRICT`; không cascade xóa lịch sử đã chốt.
- Email và mã thanh toán cần unique không phân biệt hoa/thường.

## 1. `users`

| Cột | Kiểu | Ràng buộc / ý nghĩa |
|---|---|---|
| `id` | UUID | Primary key |
| `name` | VARCHAR(100) | Bắt buộc |
| `email` | VARCHAR(254) | Bắt buộc, unique không phân biệt hoa/thường |
| `password_hash` | TEXT | Bcrypt hash; không trả qua API |
| `phone` | VARCHAR(20) | Tùy chọn |
| `avatar_url` | TEXT | Tùy chọn |
| `created_at`, `updated_at` | TIMESTAMPTZ | Thời gian quản lý bản ghi |

## 2. `groups`

| Cột | Kiểu | Ràng buộc / ý nghĩa |
|---|---|---|
| `id` | UUID | Primary key |
| `name` | VARCHAR(100) | Bắt buộc |
| `created_by` | UUID | FK `users.id`; chỉ lưu lịch sử tạo nhóm |
| `share_code` | VARCHAR(16) | Bắt buộc, unique, đủ entropy |
| `currency` | CHAR(3) | Mã ISO 4217; MVP mặc định `VND` |
| `status` | TEXT | `ACTIVE` hoặc `ARCHIVED` |
| `created_at`, `updated_at` | TIMESTAMPTZ | Thời gian quản lý bản ghi |

Tạo group và membership quản trị đầu tiên phải nằm trong cùng transaction.

## 3. `group_members`

| Cột | Kiểu | Ràng buộc / ý nghĩa |
|---|---|---|
| `group_id` | UUID | FK `groups.id` |
| `user_id` | UUID | FK `users.id` |
| `role` | TEXT | `ADMIN` hoặc `MEMBER` |
| `status` | TEXT | `ACTIVE` hoặc `LEFT` |
| `joined_at`, `left_at` | TIMESTAMPTZ | Mốc tham gia/rời nhóm |

Primary key là `(group_id, user_id)`. Một nhóm phải luôn có ít nhất một `ADMIN` đang hoạt
động. `ADMIN` chỉ là quyền vận hành nhóm; cả hai role đều có thể là người ứng, người nợ
hoặc người được nhận.

## 4. `expenses`

| Cột | Kiểu | Ràng buộc / ý nghĩa |
|---|---|---|
| `id` | UUID | Primary key |
| `group_id` | UUID | FK `groups.id` |
| `created_by` | UUID | Người ghi bản ghi; phải là member active |
| `paid_by` | UUID | Người thực sự ứng tiền; phải là member active |
| `description` | VARCHAR(255) | Bắt buộc |
| `amount_minor` | BIGINT | `> 0` |
| `split_type` | TEXT | `EQUAL`, `PERCENT`, `WEIGHT` hoặc `CUSTOM` |
| `expense_date` | DATE | Ngày phát sinh |
| `batch_id` | UUID | Nullable; FK batch sau khi chốt |
| `status` | TEXT | `ACTIVE` hoặc `VOIDED` |
| `created_at`, `updated_at` | TIMESTAMPTZ | Thời gian quản lý bản ghi |

Nên có index `(group_id, batch_id, status)` để lấy nhanh các khoản chưa chốt. Chỉ expense
`ACTIVE` với `batch_id IS NULL` được sửa hoặc đưa vào kỳ mới.

## 5. `expense_splits`

| Cột | Kiểu | Ràng buộc / ý nghĩa |
|---|---|---|
| `id` | UUID | Primary key |
| `expense_id` | UUID | FK `expenses.id` |
| `user_id` | UUID | Người chịu phần chi phí |
| `share_minor` | BIGINT | `>= 0` |
| `created_at` | TIMESTAMPTZ | Thời gian tạo |

Unique `(expense_id, user_id)`. Trong transaction tạo/sửa expense, tổng `share_minor` phải
bằng `expenses.amount_minor`. Dòng có phần chia `0` có thể bỏ để giảm dữ liệu.

## 6. `settlement_batches`

| Cột | Kiểu | Ràng buộc / ý nghĩa |
|---|---|---|
| `id` | UUID | Primary key |
| `group_id` | UUID | FK `groups.id` |
| `created_by` | UUID | Admin mở kỳ; không ảnh hưởng người nhận tiền |
| `idempotency_key` | TEXT | Unique trong phạm vi group |
| `status` | TEXT | `OPEN`, `COMPLETED` hoặc `CANCELLED` |
| `created_at`, `completed_at`, `cancelled_at` | TIMESTAMPTZ | Mốc vòng đời |

Dùng partial unique index theo `group_id WHERE status = 'OPEN'` để mỗi nhóm chỉ có một kỳ
đang mở. Batch và danh sách expense/settlement của batch là snapshot bất biến.

## 7. `settlements`

| Cột | Kiểu | Ràng buộc / ý nghĩa |
|---|---|---|
| `id` | UUID | Primary key |
| `batch_id` | UUID | FK `settlement_batches.id` |
| `from_user_id` | UUID | Người phải trả |
| `to_user_id` | UUID | Người cần được nhận |
| `amount_minor` | BIGINT | `> 0` |
| `payment_code` | VARCHAR(20) | Unique không phân biệt hoa/thường |
| `status` | TEXT | `PENDING`, `AWAITING_CONFIRMATION`, `PAID`, `CANCELLED` |
| `marked_sent_at`, `paid_at` | TIMESTAMPTZ | Nullable theo trạng thái |
| `created_at`, `updated_at` | TIMESTAMPTZ | Thời gian quản lý bản ghi |

Các check bắt buộc:

- `from_user_id <> to_user_id`.
- Cả hai user thuộc cùng group của batch tại thời điểm batch được tạo.
- Unique `(batch_id, from_user_id, to_user_id)`.
- Chỉ `from_user_id` được chuyển sang `AWAITING_CONFIRMATION`.
- Chỉ `to_user_id` được xác nhận `PAID` hoặc từ chối về `PENDING`.

## 8. `settlement_events`

| Cột | Kiểu | Ràng buộc / ý nghĩa |
|---|---|---|
| `id` | UUID | Primary key |
| `settlement_id` | UUID | FK `settlements.id` |
| `actor_id` | UUID | User thực hiện |
| `event_type` | TEXT | `MARKED_SENT`, `CONFIRMED`, `REJECTED`, `CANCELLED` |
| `note`, `receipt_url` | TEXT | Tùy chọn |
| `created_at` | TIMESTAMPTZ | Thời điểm sự kiện |

Bảng sự kiện là lịch sử append-only, không thay thế trạng thái hiện tại ở `settlements`.

## 9. `audit_logs`

| Cột | Kiểu | Ràng buộc / ý nghĩa |
|---|---|---|
| `id` | UUID | Primary key |
| `group_id` | UUID | FK `groups.id` |
| `actor_id` | UUID | User thực hiện |
| `action` | TEXT | Tên hành động nghiệp vụ |
| `entity_type`, `entity_id` | TEXT, UUID | Bản ghi bị tác động |
| `metadata` | JSONB | Snapshot tối thiểu phục vụ truy vết |
| `created_at` | TIMESTAMPTZ | Thời điểm hành động |

Không lưu JWT, mật khẩu, secret hoặc dữ liệu biên lai nhạy cảm trong `metadata`.

## Quan hệ chính

```text
users ──< group_members >── groups
  │                            │
  │                            ├──< expenses ──< expense_splits
  │                            │        │
  │                            │        └── settlement_batches
  │                            │
  └── from/to ── settlements >─┴── settlement_batches
                         │
                         └──< settlement_events

groups ──< audit_logs
```

## Invariant tài chính

1. Với mỗi expense: `SUM(expense_splits.share_minor) = expenses.amount_minor`.
2. Với mỗi batch: `SUM(net_balance) = 0`.
3. Tổng tiền từ các settlement bằng tổng số dư dương và trị tuyệt đối tổng số dư âm.
4. Không tạo settlement số tiền `0`, âm hoặc từ một user tới chính họ.
5. Một expense chỉ thuộc tối đa một batch.
6. Không sửa expense, split hoặc settlement đã chốt; điều chỉnh bằng bản ghi mới và audit.

## Transaction chốt kỳ

Trong một transaction:

1. Kiểm tra quyền `ADMIN`, idempotency key và không có batch `OPEN` khác.
2. Khóa danh sách expense `ACTIVE` chưa có `batch_id` của nhóm.
3. Kiểm tra membership và invariant tổng phần chia.
4. Tạo batch, tính net balance, chạy resolver của Chế độ chia đều linh hoạt.
5. Tạo settlements và gắn toàn bộ expense đã chọn vào batch.
6. Kiểm tra lại tổng settlement rồi commit.

Nếu bất kỳ bước nào lỗi, rollback toàn bộ. Resolver không đọc dữ liệu ngoài tập expense đã
khóa, nhờ đó request đồng thời không chốt trùng.

## Cấu trúc code mục tiêu

```text
backend/
├── internal/
│   ├── auth/                  # Xác thực đã triển khai
│   ├── groups/                # Membership và quyền quản trị
│   ├── expenses/              # Expense, split và validation
│   ├── settlements/           # Balance, resolver và vòng đời thanh toán
│   ├── audit/                 # Nhật ký append-only
│   ├── handlers/              # HTTP handlers
│   └── middleware/            # Xác thực và request context
└── migrations/

frontend/src/
├── app/                       # Route công khai và được bảo vệ
├── components/screens/        # Màn hình cấp trang
├── components/group/          # Dashboard nhóm thống nhất
├── lib/                       # API client
└── stores/                    # Client state
```

Chỉ cần một dashboard nhóm. Giao diện thay đổi theo quan hệ thực tế của user với từng
settlement: người trả thấy hành động báo chuyển, người nhận thấy hành động xác nhận.
