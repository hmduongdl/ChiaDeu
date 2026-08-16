# Chia Đều — bài tập backend cho Chế độ chia đều linh hoạt

Mục tiêu của bài tập là chuyển schema mục tiêu thành struct Go và luyện cách tính công nợ
theo Chế độ chia đều linh hoạt (`MULTI_CREDITOR`), khi nhiều thành viên có thể thay nhau
ứng tiền. Chưa cần viết HTTP API hoặc kết nối database.

Đọc [`schema.md`](schema.md) trước khi bắt đầu. Trong mô hình này không có trường xác định
một người nhận tiền cố định; hướng thanh toán luôn được suy ra từ số dư ròng.

## 1. Schema rút gọn

```text
users
- id, name, email, password_hash, phone

groups
- id, name, created_by, share_code, currency

group_members
- group_id, user_id, role, status, joined_at

expenses
- id, group_id, created_by, paid_by, description, amount_minor, split_type, expense_date

expense_splits
- id, expense_id, user_id, share_minor

settlement_batches
- id, group_id, created_by, status

settlements
- id, batch_id, from_user_id, to_user_id, amount_minor, payment_code, status
```

## 2. Cấu trúc thư mục bài tập

```text
backend/
├── main.go
├── models/
│   ├── user.go
│   ├── group.go
│   ├── expense.go
│   └── settlement.go
└── services/
    ├── expense_calc.go
    └── settlement_calc.go
```

Đây là khung học tập. Khi tích hợp vào app thật, code mới nên đi vào package theo domain
dưới `backend/internal/`.

## 3. Phân công

### Nhật Thành — user, group và expense

File phụ trách:

- `models/user.go`
- `models/group.go`
- `models/expense.go`
- `services/expense_calc.go`

Yêu cầu:

1. Viết `User`, `Group`, `GroupMember`, `Expense`, `ExpenseSplit` theo schema rút gọn.
2. Viết `DisplayName()` cho user.
3. Viết `IsActiveMember()` cho membership.
4. Viết `SumSplits()` và kiểm tra tổng split bằng amount.
5. Viết `SplitEqual()`; phần dư phải được phân bổ ổn định theo thứ tự `memberIDs`.

Ví dụ khung:

```go
type Group struct {
    ID        string
    Name      string
    CreatedBy string
    ShareCode string
    Currency  string
}

type GroupMember struct {
    GroupID  string
    UserID   string
    Role     string // ADMIN hoặc MEMBER
    Status   string // ACTIVE hoặc LEFT
    JoinedAt string
}

func (m GroupMember) IsActiveMember() bool {
    return m.Status == "ACTIVE"
}
```

```go
type Expense struct {
    ID          string
    GroupID     string
    CreatedBy   string
    PaidBy      string
    Description string
    AmountMinor int64
    SplitType   string // EQUAL, PERCENT, WEIGHT hoặc CUSTOM
    ExpenseDate string
}

type ExpenseSplit struct {
    ID         string
    ExpenseID  string
    UserID     string
    ShareMinor int64
}

func SumSplits(splits []ExpenseSplit) int64 {
    var total int64
    for _, split := range splits {
        total += split.ShareMinor
    }
    return total
}
```

Test tối thiểu cho `SplitEqual()`:

- `100 / 3` tạo ba phần có tổng đúng `100`.
- Không có member trả về lỗi.
- Amount bằng `0` hoặc âm trả về lỗi.
- Kết quả cùng input luôn giống nhau.

### Phúc Khang — balance và settlement resolver

File phụ trách:

- `models/settlement.go`
- `services/settlement_calc.go`

Yêu cầu:

1. Viết `SettlementBatch` và `Settlement`.
2. Viết `MarkAsSent()` và `MarkAsPaid()` có kiểm tra chuyển trạng thái hợp lệ.
3. Viết `CalculateNetBalances(expenses, splits)` theo công thức:

   `số dư = tổng đã ứng - tổng phải chịu`.

4. Viết `SimplifyDebts(balances)` để ghép nhiều người nợ với nhiều người cần được nhận.
5. Kiểm tra invariant trước và sau khi tạo settlement.

Ví dụ khung:

```go
type SettlementBatch struct {
    ID        string
    GroupID   string
    CreatedBy string
    Status    string // OPEN, COMPLETED hoặc CANCELLED
}

type Settlement struct {
    ID          string
    BatchID     string
    FromUserID  string
    ToUserID    string
    AmountMinor int64
    PaymentCode string
    Status      string // PENDING, AWAITING_CONFIRMATION, PAID hoặc CANCELLED
}

func (s Settlement) IsPayer(userID string) bool {
    return s.FromUserID == userID
}

func (s Settlement) IsRecipient(userID string) bool {
    return s.ToUserID == userID
}
```

Quy tắc cho `SimplifyDebts()`:

- Số dư dương là cần nhận, số dư âm là cần trả.
- Tổng input phải bằng `0`; nếu không thì trả lỗi.
- Không tạo giao dịch amount `<= 0` hoặc tự trả cho chính mình.
- Sắp xếp theo số dư tuyệt đối giảm dần và `userID` để kết quả xác định.
- Sau khi áp dụng output, số dư của mọi user phải về `0`.

Test tối thiểu:

| Input balance | Output mong đợi |
|---|---|
| `A:+200, B:-120, C:-80` | `B→A:120`, `C→A:80` |
| `A:+70, B:+30, C:-100` | Hai giao dịch, tổng `100` |
| `A:0, B:0` | Không có settlement |
| `A:+10, B:-9` | Trả lỗi vì tổng khác `0` |

## 4. Làm chung

1. Thống nhất tên field, constant trạng thái và quy ước dấu của balance.
2. Viết table-driven tests cho chia tiền và resolver.
3. Review chéo: Nhật Thành kiểm tra invariant settlement; Phúc Khang kiểm tra phần dư chia đều.
4. Chạy `gofmt` và `go test ./...` trước khi merge.

## 5. Theo dõi tiến độ

| Mã | Công việc | Người phụ trách | Trạng thái | Ghi chú |
|---|---|---|:---:|---|
| 1 | Struct user, group, membership | Nhật Thành | ✅ | `backend/models/` |
| 2 | Struct expense và split | Nhật Thành | ✅ | `backend/models/` |
| 3 | `SumSplits()` và `SplitEqual()` | Nhật Thành | ✅ | `backend/services/expense_calc.go` + test |
| 4 | Struct batch và settlement | Phúc Khang | ✅ | `backend/models/` |
| 5 | `CalculateNetBalances()` | Phúc Khang | ✅ | `backend/services/settlement_calc.go` + test |
| 6 | `SimplifyDebts()` | Phúc Khang | ✅ | `backend/services/settlement_calc.go` + test |
| 7 | Table-driven tests và review chéo | Cả hai | ✅ | `go test ./...` pass toàn bộ |

Trạng thái: ☐ Chưa làm · 🔄 Đang làm · ✅ Xong

## 6. Tiêu chí hoàn thành

- Code compile và `go test ./...` pass.
- Không dùng `float32`/`float64` cho tiền.
- Tổng split luôn bằng amount sau phân bổ phần dư.
- Resolver xử lý được nhiều người ứng và nhiều người nợ.
- Kết quả resolver xác định với cùng input.
- Mọi comment trong code viết bằng tiếng Việt theo quy ước repository.
- Cập nhật `update_log.md` khi code hoặc tài liệu được đưa vào repository.
