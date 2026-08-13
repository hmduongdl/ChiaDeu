# ChiaDeu — Nhiệm vụ tân thủ: Class, Object & Cấu trúc thư mục (Thành & Phúc Khang)

> Mục tiêu duy nhất của giai đoạn này: biết cách biến schema (bảng dữ liệu) thành **class/struct**
> trong Go, và tự sắp xếp vào **cấu trúc thư mục** hợp lý. CHƯA cần biết API, webhook, JWT là gì —
> những phần đó để sau. Ở đây chỉ code các "khuôn dữ liệu" (giống như định nghĩa 1 cái phiếu thông tin)
> và vài hàm xử lý đơn giản trên khuôn đó.

---

## 0. Khái niệm cần biết trước khi làm (đọc trước, 10 phút)

**Struct trong Go** giống như "class" ở các ngôn ngữ khác — nó là 1 khuôn để mô tả 1 loại dữ liệu
có nhiều thuộc tính. Ví dụ 1 cái bảng `users` trong database sẽ có 1 struct `User` tương ứng:

```go
type User struct {
    ID       string
    Name     string
    Email    string
}
```

**Object** là 1 "bản ghi cụ thể" được tạo ra từ struct đó:

```go
u := User{ID: "1", Name: "Dương", Email: "duong@gmail.com"}
```

**Method** là 1 hàm gắn liền với struct, dùng để xử lý dữ liệu trên chính object đó:

```go
func (u User) GetDisplayName() string {
    return u.Name + " (" + u.Email + ")"
}
```

Bài tập của 2 bạn: nhìn vào schema (bảng dữ liệu) ở mục 1 bên dưới, tự viết struct + vài hàm xử lý
đơn giản cho từng bảng được giao.

---

## 1. Schema cần chuyển thành struct (đã rút gọn, dễ hiểu)

```
users
- id, name, email, password_hash, phone

groups
- id, name, created_by, leader_id, share_code

group_members
- group_id, user_id, joined_at

expenses
- id, group_id, paid_by, description, amount_minor, split_type, expense_date

expense_splits
- id, expense_id, user_id, share_minor

settlement_batches
- id, group_id, leader_id, status

settlements
- id, batch_id, from_user_id, to_user_id, amount_minor, payment_code, status
```

---

## 2. Cấu trúc thư mục (2 bạn tự dựng, mình chỉ gợi ý khung)

```
backend/
├── main.go
├── models/              <-- toàn bộ struct (class) nằm ở đây
│   ├── user.go
│   ├── group.go
│   ├── expense.go
│   └── settlement.go
└── services/             <-- các hàm xử lý logic đơn giản trên struct
    ├── expense_calc.go
    └── settlement_calc.go
```

**Việc đầu tiên của cả 2 bạn (làm chung, 30 phút):** tự tạo cấu trúc thư mục này trong repo
(`backend/models/`, `backend/services/`), tạo sẵn các file rỗng, commit lên nhánh riêng.
Đây là bài tập làm quen với việc tổ chức code, chưa cần viết nội dung vội.

---

## 3. Phân công

### Bạn 1: Nhật Thành — phụ trách: `User`, `Group`, `Expense`

**File cần tạo:** `models/user.go`, `models/group.go`, `models/expense.go`

**Việc cần làm:**

1. Viết struct `User` theo đúng cột trong schema mục 1.
2. Viết struct `Group` và `GroupMember`.
3. Viết struct `Expense` và `ExpenseSplit`.
4. Viết 2–3 method đơn giản (không cần database, chỉ xử lý trên object có sẵn):

```go
// models/user.go
type User struct {
    ID           string
    Name         string
    Email        string
    PasswordHash string
    Phone        string
}

// Method đơn giản: trả về tên hiển thị
func (u User) DisplayName() string {
    return u.Name
}
```

```go
// models/group.go
type Group struct {
    ID        string
    Name      string
    CreatedBy string
    LeaderID  string
    ShareCode string
}

type GroupMember struct {
    GroupID  string
    UserID   string
    JoinedAt string
}

// Method: kiểm tra 1 user có phải leader của group này không
func (g Group) IsLeader(userID string) bool {
    return g.LeaderID == userID
}
```

```go
// models/expense.go
type Expense struct {
    ID          string
    GroupID     string
    PaidBy      string
    Description string
    AmountMinor int64  // số tiền tính theo đồng, ví dụ 120000 = 120.000đ
    SplitType   string // "EQUAL", "PERCENT", "WEIGHT", "CUSTOM"
    ExpenseDate string
}

type ExpenseSplit struct {
    ID          string
    ExpenseID   string
    UserID      string
    ShareMinor  int64
}

// Method: cộng tổng các phần chia, dùng để kiểm tra sau này
func SumSplits(splits []ExpenseSplit) int64 {
    var total int64
    for _, s := range splits {
        total += s.ShareMinor
    }
    return total
}
```

**Bài tập thêm (khi xong phần trên):** viết 1 hàm `SplitEqual(amount int64, memberIDs []string) []ExpenseSplit`
tự chia đều `amount` cho từng người trong `memberIDs`, trả về danh sách `ExpenseSplit`.
Đây là bài tập rèn tư duy hàm + slice/mảng trong Go, chưa liên quan gì tới web hay database.

---

### Bạn 2: Phúc Khang — phụ trách: `SettlementBatch`, `Settlement`

**File cần tạo:** `models/settlement.go`

**Việc cần làm:**

1. Viết struct `SettlementBatch`.
2. Viết struct `Settlement`.
3. Viết 2–3 method đơn giản:

```go
// models/settlement.go
type SettlementBatch struct {
    ID       string
    GroupID  string
    LeaderID string
    Status   string // "OPEN", "COMPLETED", "CANCELLED"
}

type Settlement struct {
    ID          string
    BatchID     string
    FromUserID  string
    ToUserID    string
    AmountMinor int64
    PaymentCode string
    Status      string // "PENDING", "AWAITING_CONFIRMATION", "PAID", "CANCELLED"
}

// Method: kiểm tra settlement này có phải của user X không (X là người phải trả)
func (s Settlement) IsOwnedBy(userID string) bool {
    return s.FromUserID == userID
}

// Method: đánh dấu đã báo chuyển khoản (chỉ đổi giá trị trong object, chưa lưu DB)
func (s *Settlement) MarkAsSent() {
    s.Status = "AWAITING_CONFIRMATION"
}

// Method: đánh dấu đã xác nhận nhận tiền
func (s *Settlement) MarkAsPaid() {
    s.Status = "PAID"
}
```

**Bài tập thêm (khi xong phần trên):** viết hàm `CalculateSettlements(balances map[string]int64, leaderID string) []Settlement`
— nhận vào 1 danh sách "ai nợ bao nhiêu" (`map[string]int64`, key là `userID`, value là số tiền nợ),
tự tạo ra danh sách `Settlement` tương ứng (mỗi người nợ > 0 thì tạo 1 settlement gửi về `leaderID`).

---

## 4. Form theo dõi tiến độ

| Mã | Công việc | Người phụ trách | Trạng thái | Ngày xong | Ghi chú |
|---|---|---|:---:|---|---|
| 1 | Tạo cấu trúc thư mục `models/`, `services/` | Cả 2 | ☐ | | |
| 2 | Struct `User` + method `DisplayName()` | Bạn 1 | ☐ | | |
| 3 | Struct `Group`, `GroupMember` + method `IsLeader()` | Bạn 1 | ☐ | | |
| 4 | Struct `Expense`, `ExpenseSplit` + hàm `SumSplits()` | Bạn 1 | ☐ | | |
| 5 | Bài tập thêm: hàm `SplitEqual()` | Bạn 1 | ☐ | | |
| 6 | Struct `SettlementBatch`, `Settlement` | Bạn 2 | ☐ | | |
| 7 | Method `IsOwnedBy()`, `MarkAsSent()`, `MarkAsPaid()` | Bạn 2 | ☐ | | |
| 8 | Bài tập thêm: hàm `CalculateSettlements()` | Bạn 2 | ☐ | | |

Trạng thái: ☐ Chưa làm · 🔄 Đang làm · ✅ Xong

---

