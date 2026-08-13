# Chia Đều

Chia Đều là ứng dụng ghi nhận chi tiêu chung và tính công nợ cho nhóm bạn bè, chuyến đi
hoặc hộ gia đình. Sản phẩm chỉ sử dụng **Chế độ chia đều linh hoạt**
(`MULTI_CREDITOR`): bất kỳ thành viên nào cũng có thể ứng tiền, và hệ thống tính các giao
dịch hoàn tiền trực tiếp giữa những người đang nợ và những người cần được nhận.

> Trạng thái hiện tại: luồng đăng ký, đăng nhập, làm mới phiên, đăng xuất và bảo vệ route
> đã hoạt động; bốn màn hình chính đã có giao diện dùng dữ liệu trình diễn. API nhóm, khoản
> chi và quyết toán mới là khung endpoint, chưa triển khai nghiệp vụ. Các phần có nhãn
> “mục tiêu” trong tài liệu là hợp đồng cho giai đoạn tiếp theo.

## Nguyên tắc sản phẩm

- Mỗi nhóm dùng duy nhất Chế độ chia đều linh hoạt (`MULTI_CREDITOR`); không có lựa chọn
  chế độ khi tạo nhóm.
- Không có vai trò tài chính cố định. Người tạo hoặc quản trị nhóm không mặc nhiên là
  người ứng tiền hay người nhận tiền.
- Mọi thành viên đang hoạt động đều có thể tạo khoản chi và là `paid_by` của khoản đó.
- Một khoản chi có thể chia đều, theo phần trăm, trọng số hoặc số tiền tùy chỉnh.
- Hệ thống tổng hợp số dư ròng rồi rút gọn danh sách giao dịch cần thanh toán.
- Người trả chỉ được báo “đã chuyển”; chỉ đúng người nhận mới được xác nhận “đã nhận”.
- Tiền được lưu bằng số nguyên theo đơn vị nhỏ nhất của tiền tệ, không dùng `float`.
- Khoản chi và giao dịch đã đưa vào kỳ quyết toán phải được giữ nguyên để bảo toàn lịch sử.

## Thuật ngữ

| Thuật ngữ | Ý nghĩa |
|---|---|
| Nhóm (`Group`) | Không gian chung chứa thành viên, khoản chi và công nợ |
| Quản trị viên (`ADMIN`) | Quản lý tên nhóm, lời mời, thành viên và kỳ quyết toán; không có đặc quyền tài chính |
| Thành viên (`MEMBER`) | Có thể ghi khoản chi, tham gia phần chia và thanh toán công nợ |
| Khoản chi (`Expense`) | Khoản tiền một thành viên đã ứng cho một hoặc nhiều người |
| Phần chia (`ExpenseSplit`) | Số tiền một thành viên phải chịu trong một khoản chi |
| Số dư ròng (`Net balance`) | Tổng tiền đã ứng trừ tổng phần chi phí phải chịu |
| Kỳ quyết toán (`SettlementBatch`) | Ảnh chụp bất biến của các khoản chi được chốt cùng lúc |
| Giao dịch hoàn tiền (`Settlement`) | Chỉ dẫn một thành viên trả một thành viên khác |

`created_by` chỉ lưu người tạo bản ghi. Quyền quản trị nằm ở membership và không được dùng
để quyết định hướng dòng tiền.

## Cách tính công nợ

Với mỗi thành viên:

```text
net_balance = tổng tiền đã ứng - tổng phần chi phí phải chịu
```

- `net_balance > 0`: người đó cần được nhận tiền.
- `net_balance < 0`: người đó cần trả tiền.
- `net_balance = 0`: đã cân bằng, không tạo settlement.

Tổng số dư của một kỳ hợp lệ phải bằng `0`. Backend từ chối chốt kỳ nếu tổng phần chia của
bất kỳ khoản chi nào khác tổng tiền khoản chi hoặc nếu người trả/người tham gia không thuộc
nhóm tại thời điểm ghi nhận.

### Rút gọn giao dịch

Resolver mục tiêu thực hiện theo thứ tự:

1. Tính net balance từ toàn bộ khoản chi chưa chốt.
2. Tách danh sách người cần nhận và người cần trả; bỏ số dư bằng `0`.
3. Sắp xếp theo số dư tuyệt đối giảm dần, dùng `user_id` làm khóa phụ để kết quả ổn định.
4. Ghép người nợ lớn nhất với người được nhận lớn nhất, tạo settlement bằng giá trị nhỏ
   hơn trong hai số dư.
5. Cập nhật số dư và lặp đến khi cả hai phía đều bằng `0`.

Cách làm này có độ phức tạp `O(n log n)`, tạo không quá `d + c - 1` giao dịch với `d` người
nợ và `c` người cần nhận, đồng thời cho kết quả xác định để retry không sinh danh sách khác.
MVP ưu tiên tính đúng, dễ kiểm thử và ít giao dịch; chưa cam kết tìm nghiệm tối ưu toàn cục
cho mọi tổ hợp số dư.

Ví dụ:

| Thành viên | Đã ứng | Phải chịu | Số dư ròng |
|---|---:|---:|---:|
| An | 300.000đ | 100.000đ | +200.000đ |
| Bình | 0đ | 120.000đ | -120.000đ |
| Châu | 60.000đ | 140.000đ | -80.000đ |

Kết quả: Bình trả An `120.000đ`, Châu trả An `80.000đ`.

## Quyền hạn

| Hành động | `ADMIN` | `MEMBER` |
|---|:---:|:---:|
| Xem thành viên, khoản chi và số dư nhóm | Có | Có |
| Tạo khoản chi với chính mình là `paid_by` | Có | Có |
| Sửa/hủy khoản chi chưa chốt do mình tạo | Có | Có |
| Quản lý lời mời và membership | Có | Không |
| Mở hoặc hủy kỳ quyết toán | Có | Không |
| Báo đã chuyển settlement mình phải trả | Có | Có |
| Xác nhận/từ chối settlement mình được nhận | Có | Có |

Một nhóm nên có ít nhất một `ADMIN`, nhưng có thể có nhiều quản trị viên. Backend luôn kiểm
tra membership, quyền quản trị và quan hệ `from_user_id`/`to_user_id`; việc frontend ẩn nút
không được xem là phân quyền.

## Luồng nghiệp vụ mục tiêu

### 1. Tạo nhóm và ghi khoản chi

Người tạo nhóm trở thành membership `ADMIN` đầu tiên. Sau khi các thành viên tham gia bằng
mã mời, bất kỳ ai cũng có thể ghi khoản mình đã trả và chọn những người phải chia.

Backend phải bảo đảm:

- `paid_by` và tất cả người trong phần chia là thành viên đang hoạt động của cùng nhóm.
- `amount_minor > 0`, mọi `share_minor >= 0` và tổng phần chia bằng tổng khoản chi.
- Phần dư khi chia đều được phân bổ theo thứ tự ổn định để không mất đơn vị tiền nhỏ nhất.
- Khoản chi đã chốt không được sửa trực tiếp.

### 2. Chốt kỳ

Mỗi nhóm chỉ có tối đa một kỳ `OPEN`. Việc chốt chạy trong một database transaction:
khóa các khoản chi chưa chốt, tính balance, sinh settlements và gắn các khoản chi vào batch.
Idempotency key giúp retry an toàn khi mạng gián đoạn.

Batch là snapshot. Khoản chi phát sinh sau đó thuộc kỳ tiếp theo. Hủy batch chỉ được phép
khi chưa có settlement `PAID`; thao tác hủy phải trả các khoản chi về trạng thái chưa chốt
và ghi audit log.

### 3. Xác nhận hoàn tiền

```text
PENDING --người trả báo đã chuyển--> AWAITING_CONFIRMATION
AWAITING_CONFIRMATION --người nhận xác nhận--> PAID
AWAITING_CONFIRMATION --người nhận từ chối--> PENDING
PENDING --hủy kỳ--> CANCELLED
```

`PAID` và `CANCELLED` là trạng thái kết thúc. Mọi thay đổi trạng thái cần ghi actor, thời
gian và dữ liệu liên quan vào audit log.

## Kiến trúc

```text
Next.js PWA
    │ REST/JSON qua proxy same-origin /api
    ▼
Go + Fiber API
    │ transaction và truy vấn tham số hóa
    ▼
PostgreSQL
```

| Thành phần | Công nghệ | Trách nhiệm |
|---|---|---|
| Frontend | Next.js 14, React, TypeScript, Tailwind CSS, Zustand | Giao diện, trạng thái phiên, gọi API |
| Backend | Go, Fiber, pgx | Xác thực, phân quyền, nghiệp vụ và resolver |
| Database | PostgreSQL 16 | Dữ liệu, ràng buộc, transaction và audit |
| Auth | bcrypt, HS256 access/refresh JWT trong HttpOnly cookie | Xác thực phiên không dùng Web Storage |

Chi tiết schema mục tiêu nằm trong [`schema.md`](schema.md). Phân công bài tập backend nằm
trong [`phancong.md`](phancong.md).

## API

### Đã triển khai

| Method & path | Chức năng |
|---|---|
| `GET /api/health` | Kiểm tra API |
| `POST /api/auth/register` | Tạo tài khoản |
| `POST /api/auth/login` | Đăng nhập và cấp cookie |
| `POST /api/auth/refresh` | Cấp lại access token từ refresh cookie |
| `POST /api/auth/logout` | Xóa cookie phiên |
| `GET /api/auth/me` | Lấy người dùng hiện tại; yêu cầu access token |

### Hợp đồng mục tiêu cho nghiệp vụ nhóm

Các route dưới đây chưa được triển khai đầy đủ. Request tạo nhóm không nhận
`settlement_mode`, vì toàn hệ thống dùng Chế độ chia đều linh hoạt (`MULTI_CREDITOR`).

| Method & path | Chức năng |
|---|---|
| `POST /api/groups` | Tạo nhóm và membership `ADMIN` đầu tiên |
| `POST /api/groups/join/:shareCode` | Tham gia nhóm |
| `GET /api/groups/:id` | Lấy thông tin và thành viên nhóm |
| `POST /api/groups/:id/expenses` | Tạo khoản chi cùng các phần chia |
| `PATCH /api/groups/:id/expenses/:expenseId` | Sửa khoản chi chưa chốt do mình tạo |
| `GET /api/groups/:id/balances` | Lấy số dư ròng chưa chốt |
| `POST /api/groups/:id/settlement-batches` | Chốt công nợ bằng resolver của Chế độ chia đều linh hoạt |
| `GET /api/groups/:id/settlement-batches/:batchId` | Xem snapshot và tiến độ thanh toán |
| `POST /api/settlements/:id/mark-sent` | Người trả báo đã chuyển |
| `POST /api/settlements/:id/confirm` | Người nhận xác nhận đã nhận |
| `POST /api/settlements/:id/reject` | Người nhận từ chối xác nhận |

Các lệnh tạo/chốt/xác nhận nên hỗ trợ idempotency key. API nghiệp vụ cũ đang có trong
`backend/main.go` hiện chỉ trả thông báo `not implemented`.

## Chạy local

### Yêu cầu

- Go theo phiên bản trong `backend/go.mod`.
- Node.js tương thích Next.js 14 và npm.
- PostgreSQL 16, hoặc Docker Compose.

### Cấu hình

```bash
cp .env.example .env
```

Thay hai JWT secret bằng hai giá trị ngẫu nhiên, độc lập, dài ít nhất 32 ký tự. Với local
HTTP đặt `COOKIE_SECURE=false`; production phải dùng HTTPS và `COOKIE_SECURE=true`.

### Docker Compose

```bash
docker-compose up --build
```

- Frontend: `http://localhost:3000`
- Backend: `http://localhost:8080`
- Health check: `http://localhost:8080/api/health`

### Chạy từng phần

```bash
# Terminal 1
cd backend
go run .

# Terminal 2
cd frontend
npm install
npm run dev
```

Browser gọi `/api/*` cùng origin với frontend; Next.js chuyển tiếp sang `BACKEND_URL`.
Access/refresh token chỉ nằm trong HttpOnly cookie và không được ghi log hay lưu vào
`localStorage`/`sessionStorage`.

## Kiểm thử

```bash
cd backend && go test ./...
cd frontend && npm run build
git diff --check
```

`npm run lint` hiện yêu cầu bổ sung cấu hình ESLint. Docker/runtime integration cần được
chạy ở môi trường có Docker.

## Lộ trình

1. Thay schema cũ bằng migration mục tiêu cho Chế độ chia đều linh hoạt, gồm integer money,
   batch và audit.
2. Triển khai membership, expense/split validation và truy vấn balance.
3. Triển khai resolver xác định, test invariant và chốt kỳ bằng transaction.
4. Hoàn thiện vòng đời settlement, idempotency và phân quyền người trả/người nhận.
5. Xây dashboard nhóm thống nhất: “Bạn cần trả”, “Bạn sẽ nhận”, balance và lịch sử.
6. Bổ sung rate limit, refresh-token rotation/revocation và kiểm thử tích hợp trước production.

## Giới hạn hiện tại

- Migration `001_init.sql` là schema thử nghiệm cũ; chưa đáp ứng schema mục tiêu trong
  `schema.md` và cần migration thay thế an toàn.
- Refresh token chưa rotate hoặc revoke phía server.
- Chưa hỗ trợ thanh toán một phần, gộp nhiều settlement vào một lần chuyển hoặc đa tiền tệ
  trong cùng nhóm.
