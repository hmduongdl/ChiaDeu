# Chia Đều — chia tiền nhóm bạn bè / đi chơi

*Tài liệu định hướng sản phẩm, yêu cầu nghiệp vụ và thiết kế hệ thống dự kiến*

> **Trạng thái hiện tại:** dự án đã có phần xác thực tài khoản cơ bản. Các API nhóm, khoản
> chi, thanh toán mô tả trong tài liệu này là **thiết kế mục tiêu**, chưa phải tính năng
> backend đã hoàn thiện. Xem mục [16](#16-xác-thực-và-chạy-dự-án-hiện-tại) để biết phần đã
> chạy được thật sự.

---

## Mục lục

1. [Tầm nhìn sản phẩm](#1-tầm-nhìn-sản-phẩm)
2. [Thuật ngữ](#2-thuật-ngữ)
3. [Nguyên tắc nghiệp vụ bắt buộc](#3-nguyên-tắc-nghiệp-vụ-bắt-buộc)
4. [Vai trò và quyền hạn](#4-vai-trò-và-quyền-hạn)
5. [Luồng sử dụng chính](#5-luồng-sử-dụng-chính)
6. [Trạng thái nghiệp vụ](#6-trạng-thái-nghiệp-vụ)
7. [Quy tắc đổi chủ xị](#7-quy-tắc-đổi-chủ-xị-chỉ-áp-dụng-cho-chế-độ-chủ-xị)
8. [Kiến trúc dự kiến](#8-kiến-trúc-dự-kiến)
9. [Mô hình dữ liệu đề xuất](#9-mô-hình-dữ-liệu-đề-xuất)
10. [API mục tiêu](#10-api-mục-tiêu)
11. [Bảo mật và độ tin cậy](#11-yêu-cầu-bảo-mật-và-độ-tin-cậy)
12. [Yêu cầu UX](#12-yêu-cầu-ux)
13. [Các trường hợp biên](#13-các-trường-hợp-biên-cần-xử-lý)
14. [Tiêu chí nghiệm thu MVP](#14-tiêu-chí-nghiệm-thu-mvp)
15. [Lộ trình đề xuất](#15-lộ-trình-đề-xuất)
16. [Xác thực và chạy dự án hiện tại](#16-xác-thực-và-chạy-dự-án-hiện-tại)

---

## 1. Tầm nhìn sản phẩm

Chia Đều là ứng dụng chia tiền nhóm bạn bè sau mỗi lần đi chơi, ăn uống, du lịch. Không
cần liên kết ngân hàng, không cần cổng thanh toán — mở app lên, nhập khoản chi, hệ thống
tự tính ai nợ ai bao nhiêu.

App hỗ trợ **hai chế độ tính nợ** linh hoạt, chọn khi tạo nhóm:

- **Chế độ "chủ xị" (single-creditor)** — một người đứng ra ứng tiền và thanh toán các
  khoản chi chung của nhóm. Các thành viên chỉ cần xem phần mình phải chịu và hoàn tiền
  lại cho chủ xị.
- **Chế độ "ai nợ ai" (multi-creditor)** — nhiều thành viên có thể thay nhau ứng tiền.
  Hệ thống dùng thuật toán min-cash-flow để tối giản số giao dịch thanh toán giữa các
  thành viên, giống cách Splitwise vận hành.

Cả hai chế độ dùng chung một tầng tính net balance, chỉ khác nhau ở tầng "resolver" khi
sinh danh sách giao dịch thanh toán. Mọi xác nhận thanh toán đều thực hiện thủ công trong
app: thành viên báo "đã trả", người nhận xác nhận "đã nhận".

## 2. Thuật ngữ

| Thuật ngữ | Ý nghĩa |
|---|---|
| Nhóm (`Group`) | Không gian ghi nhận khoản chi và công nợ của một nhóm người |
| Chủ xị (`Leader`) | Trong chế độ "chủ xị": thành viên đang ứng tiền, quản lý nhóm và nhận tiền hoàn lại |
| Thành viên (`Member`) | Người tham gia chia chi phí, có thể là người ứng tiền hoặc người trả nợ |
| Chế độ nợ (`Settlement mode`) | `SINGLE_CREDITOR` (chủ xị) hoặc `MULTI_CREDITOR` (ai nợ ai), chọn khi tạo nhóm |
| Khoản chi (`Expense`) | Một lần thành viên đã thanh toán cho nhóm |
| Phần chia (`Expense split`) | Số tiền một thành viên phải chịu trong một khoản chi |
| Kỳ chốt (`Settlement batch`) | Ảnh chụp công nợ tại thời điểm yêu cầu thanh toán |
| Net balance | Số dư ròng của mỗi thành viên sau khi tổng hợp mọi khoản chi và phần chia |
| Khoản hoàn (`Settlement`) | Khoản một thành viên phải trả cho người khác trong một kỳ chốt |
| Mã thanh toán | Mã duy nhất, ví dụ `CD4F82A9`, dùng để nhận diện khoản thanh toán |

`created_by` và `leader_id` là hai khái niệm khác nhau. Người tạo nhóm được lưu để phục
vụ lịch sử; chủ xị là vai trò hiện tại và có thể được chuyển cho thành viên khác (chỉ áp
dụng trong chế độ "chủ xị"). Trong chế độ "ai nợ ai", `leader_id` có thể để trống hoặc
không sử dụng.

## 3. Nguyên tắc nghiệp vụ bắt buộc

1. Nhóm có thể hoạt động ở chế độ "chủ xị" (có đúng một leader) hoặc "ai nợ ai" (không
   có leader cố định, mọi thành viên đều có thể ứng tiền).
2. Trong chế độ "chủ xị", chủ xị phải là một thành viên của chính nhóm đó.
3. Trong chế độ "chủ xị", mọi khoản chi mới được xem là do chủ xị hiện tại thanh toán.
   Trong chế độ "ai nợ ai", bất kỳ thành viên nào cũng có thể là người trả (`paid_by`).
4. Trong chế độ "chủ xị", mọi khoản hoàn được tạo trong một kỳ chốt phải có người nhận
   là chủ xị của kỳ đó. Trong chế độ "ai nợ ai", hệ thống dùng min-cash-flow để xác định
   ai trả ai với số giao dịch tối thiểu.
5. Mỗi khoản hoàn có một mã thanh toán duy nhất, không tái sử dụng.
6. Tiền được lưu bằng đơn vị nhỏ nhất của tiền tệ. Với VND, backend dùng số nguyên đồng;
   không dùng `float` cho tính toán tài chính.
7. Dữ liệu lịch sử phải giữ nguyên người trả, người nhận tại thời điểm phát sinh, kể cả
   sau khi đổi chủ xị hoặc đổi chế độ nợ.
8. Thành viên không được tự đánh dấu `PAID` — chỉ người nhận tiền mới có quyền xác nhận
   đã nhận.

## 4. Vai trò và quyền hạn

Trong chế độ "chủ xị", bảng quyền dưới đây phân biệt rõ chủ xị và thành viên. Trong chế
độ "ai nợ ai", mọi thành viên đều có quyền tạo khoản chi (với `paid_by` là chính mình) và
xem toàn bộ trạng thái hoàn tiền của nhóm.

| Hành động | Chủ xị | Thành viên |
|---|:---:|:---:|
| Xem nhóm, thành viên, khoản chi và công nợ của mình | Có | Có |
| Tạo/sửa/hủy khoản chi | Có | Không (*) |
| Chọn cách chia khoản chi | Có | Không (*) |
| Mở hoặc hủy kỳ chốt | Có | Không |
| Xem toàn bộ trạng thái hoàn tiền | Có | Chỉ khoản của mình (*) |
| Yêu cầu xác nhận đã chuyển thủ công | Không | Có |
| Xác nhận/từ chối yêu cầu thủ công | Có | Không |
| Chuyển vai trò chủ xị | Có | Không |
| Rời nhóm | Chỉ sau khi chuyển vai trò | Khi không còn nghĩa vụ đang mở |

(*) Trong chế độ "ai nợ ai", mọi thành viên đều có quyền tạo khoản chi và xem toàn bộ
trạng thái hoàn tiền.

Mọi quyền đều phải được backend kiểm tra từ membership hiện tại. Frontend ẩn nút không
được xem là biện pháp phân quyền.

## 5. Luồng sử dụng chính

### 5.1 Tạo nhóm và thiết lập

1. Người dùng tạo nhóm, chọn chế độ nợ ("chủ xị" hoặc "ai nợ ai") và tự động trở thành
   thành viên đầu tiên. Nếu chọn chế độ "chủ xị", người tạo đồng thời là chủ xị đầu tiên.
2. Hệ thống sinh mã mời; người khác tham gia bằng mã hoặc liên kết mời.
3. Không cần bất kỳ liên kết ngân hàng hay cổng thanh toán nào để tạo nhóm và bắt đầu
   sử dụng.

Một người có thể là chủ xị ở nhóm A nhưng chỉ là thành viên ở nhóm B.

### 5.2 Ghi nhận khoản chi

1. Người tạo khoản chi (chủ xị trong chế độ "chủ xị", hoặc bất kỳ thành viên nào trong
   chế độ "ai nợ ai") nhập mô tả, tổng tiền, ngày chi và người tham gia.
2. Chọn cách chia: đều, phần trăm, theo trọng số hoặc số tiền tùy chỉnh.
3. Backend kiểm tra tổng các phần chia phải bằng tổng khoản chi.
4. Phần của chính người trả (`paid_by`) được ghi nhận là chi phí cá nhân, không tạo khoản
   hoàn cho chính họ.
5. Phần của mỗi thành viên khác làm tăng net balance người đó nợ người trả.

### 5.3 Chốt công nợ

Hệ thống mở một kỳ chốt, lấy các phần chia chưa thuộc kỳ chốt trước đó và tính net
balance cho từng thành viên:

```text
NetBalance(thành viên) = Tổng phần chia phải chịu - Tổng tiền đã ứng
```

Từ net balance, hệ thống sinh danh sách settlement theo chế độ nợ của nhóm:

**Chế độ "chủ xị" (single-creditor):** mọi người có net balance âm (đang nợ) sẽ có một
settlement trả về chủ xị. Không cần thuật toán phức tạp vì tất cả dòng tiền đều hướng về
một người nhận duy nhất.

Ví dụ, An nợ 120.000đ và Bình nợ 80.000đ thì kỳ chốt tạo:

- An → Chủ xị: 120.000đ, mã `CD4F82A9`.
- Bình → Chủ xị: 80.000đ, mã `CD91B7C2`.

**Chế độ "ai nợ ai" (multi-creditor):** dùng thuật toán min-cash-flow để tối giản số giao
dịch thanh toán. Thuật toán hoạt động trên danh sách net balance: người có balance dương
(được nhận) sẽ nhận tiền từ người có balance âm (phải trả), ưu tiên khớp các cặp có số dư
đối ứng để giảm số lượng giao dịch.

Ví dụ, An nợ ròng 100.000đ, Bình nợ ròng 50.000đ, Châu được nhận 150.000đ:

- An → Châu: 100.000đ
- Bình → Châu: 50.000đ

Kỳ chốt là ảnh chụp bất biến. Khoản chi tạo sau thời điểm chốt sẽ thuộc kỳ kế tiếp; không
âm thầm thay đổi số tiền mà thành viên đang chuẩn bị thanh toán.

### 5.4 Thành viên hoàn tiền

Tại màn hình khoản cần trả, thành viên thấy:

- Số tiền chính xác.
- Tên người nhận.
- Mã thanh toán để ghi vào nội dung chuyển khoản.
- Trạng thái cập nhật gần thời gian thực.
- Nút "Tôi đã chuyển khoản" để báo đã thanh toán.

### 5.5 Xác nhận thanh toán

1. Thành viên nhấn "Tôi đã chuyển khoản", có thể đính kèm ghi chú hoặc ảnh biên lai.
2. Settlement chuyển từ `PENDING` sang `AWAITING_CONFIRMATION`; trạng thái này chưa được
   tính là đã thanh toán.
3. Người nhận tiền (chủ xị trong chế độ "chủ xị", hoặc thành viên được nhận trong chế độ
   "ai nợ ai") kiểm tra tài khoản và chọn xác nhận hoặc từ chối.
4. Xác nhận chuyển thành `PAID`; từ chối đưa về `PENDING` và lưu lý do.

Thành viên không được tự đánh dấu `PAID`, vì đây là trạng thái xác nhận tiền đã đến người
nhận chứ không chỉ là tuyên bố đã gửi.

## 6. Trạng thái nghiệp vụ

### 6.1 Settlement

```text
PENDING ──thành viên báo đã trả──> AWAITING_CONFIRMATION
   │                                      │
   │                                      ├──người nhận từ chối──> PENDING
   │                                      │
   └──────────────────────────────────────┴──xác nhận──> PAID
   │
   └──hủy kỳ chốt──────────────────────────────────────> CANCELLED
```

`PAID` và `CANCELLED` là trạng thái kết thúc. Sửa một settlement đã kết thúc phải thông
qua nghiệp vụ điều chỉnh có audit log, không cập nhật trực tiếp làm mất lịch sử.

## 7. Quy tắc đổi chủ xị (chỉ áp dụng cho chế độ "chủ xị")

Đổi chủ xị là thay đổi người ứng tiền và người nhận của **các hoạt động tương lai**, không
được làm thay đổi lịch sử.

1. Người nhận vai trò mới phải đang là thành viên nhóm.
2. Không cho đổi khi còn kỳ chốt có settlement `PENDING` hoặc `AWAITING_CONFIRMATION`.
3. Chủ xị cũ phải hoàn tất hoặc hủy kỳ đang mở trước khi chuyển vai trò.
4. Expense, batch, settlement cũ giữ snapshot người nhận cũ.
5. Mọi lần chuyển vai trò phải ghi audit log gồm người thao tác, người cũ, người mới và
   thời gian.

## 8. Kiến trúc dự kiến

```text
┌──────────────────────────────────┐
│ Next.js PWA                      │
│ Nhóm, khoản chi, trạng thái      │
│ hoàn tiền (2 chế độ nợ)          │
└──────────────┬───────────────────┘
               │ REST/JSON + HttpOnly cookie
┌──────────────▼───────────────────┐
│ Go/Fiber API                     │
│ Auth, phân quyền, tính net       │
│ balance, min-cash-flow resolver  │
└──────────────┬───────────────────┘
               │
┌──────────────▼───────────────────┐
│ PostgreSQL                       │
│ Dữ liệu + audit log              │
└──────────────────────────────────┘
```

### Công nghệ

| Thành phần | Công nghệ dự kiến | Trách nhiệm |
|---|---|---|
| Frontend | Next.js, React, TypeScript, Tailwind CSS | PWA và trải nghiệm theo vai trò/chế độ nợ |
| Backend | Go, Fiber | API, kiểm tra quyền, tính net balance, min-cash-flow |
| Database | PostgreSQL | Giao dịch dữ liệu, ràng buộc và audit |
| Auth | JWT ngắn hạn + refresh cookie HttpOnly | Xác thực phiên |

## 9. Mô hình dữ liệu đề xuất

Đây là schema mục tiêu cho migration tương lai. Migration hiện tại vẫn phản ánh mô hình cũ
và **chưa được xem là đã đáp ứng tài liệu này**.

### 9.1 `users`

| Cột chính | Ý nghĩa |
|---|---|
| `id`, `name`, `email`, `password_hash` | Danh tính và đăng nhập |
| `phone`, `avatar_url` | Hồ sơ tùy chọn |
| `created_at`, `updated_at` | Thời gian quản lý bản ghi |

### 9.2 `groups` và `group_members`

| Bảng/cột chính | Ý nghĩa |
|---|---|
| `groups.created_by` | Người tạo, không thay đổi theo vai trò |
| `groups.leader_id` | Chủ xị hiện tại (chỉ dùng trong chế độ "chủ xị", nullable) |
| `groups.settlement_mode` | `SINGLE_CREDITOR` hoặc `MULTI_CREDITOR`, chọn khi tạo nhóm |
| `groups.share_code`, `currency` | Mời thành viên và đơn vị tiền |
| `group_members(group_id, user_id)` | Thành viên và thời điểm tham gia |

Ràng buộc "leader phải là member" nên được bảo vệ bằng transaction/service rule; nếu dùng
ràng buộc database thì phải thiết kế khóa phù hợp để tránh trạng thái trung gian khi tạo
nhóm.

### 9.3 `expenses` và `expense_splits`

| Bảng/cột chính | Ý nghĩa |
|---|---|
| `expenses.group_id` | Nhóm sở hữu khoản chi |
| `expenses.paid_by` | Người đã trả khoản này (chủ xị trong chế độ "chủ xị", bất kỳ thành viên nào trong chế độ "ai nợ ai") |
| `expenses.amount_minor` | Tổng tiền bằng số nguyên đơn vị nhỏ nhất |
| `expenses.split_type`, `description`, `expense_date` | Quy tắc chia và nội dung |
| `expense_splits.user_id`, `share_minor` | Phần mỗi người phải chịu |
| `expense_splits.settlement_batch_id` | Nullable cho đến khi được chốt |

### 9.4 `settlement_batches` và `settlements`

| Bảng/cột chính | Ý nghĩa |
|---|---|
| `settlement_batches.leader_id` | Snapshot chủ xị của kỳ chốt (nullable trong multi-creditor) |
| `settlement_batches.status` | `OPEN`, `COMPLETED`, `CANCELLED` |
| `settlements.from_user_id` | Thành viên trả tiền |
| `settlements.to_user_id` | Người nhận tiền (chủ xị trong chế độ "chủ xị", thành viên được nhận trong chế độ "ai nợ ai"), snapshot bất biến |
| `settlements.amount_minor` | Số tiền phải trả |
| `settlements.payment_code` | Mã duy nhất, không đoán được và dễ nhập |
| `settlements.status` | Trạng thái theo mục 6 |
| `settlements.paid_at` | Thời điểm xác nhận tiền đến |

Nên có unique key `(batch_id, from_user_id)` để một thành viên chỉ có một settlement trong
mỗi kỳ và unique index không phân biệt hoa/thường cho `payment_code`.

### 9.5 `audit_logs`

| Bảng | Mục đích |
|---|---|
| `audit_logs` | Lịch sử đổi vai trò, sửa chi phí, xác nhận/từ chối thanh toán |

## 10. API mục tiêu

Các endpoint dưới đây là hợp đồng dự kiến, chưa khẳng định đã tồn tại trong backend.

### Nhóm và vai trò

| Method & path | Chức năng |
|---|---|
| `POST /api/groups` | Tạo nhóm; body gồm `settlement_mode` (`SINGLE_CREDITOR` hoặc `MULTI_CREDITOR`). Người tạo là thành viên đầu tiên (và là chủ xị nếu chọn single-creditor) |
| `POST /api/groups/join/:shareCode` | Tham gia nhóm |
| `GET /api/groups/:id` | Lấy nhóm, vai trò, thành viên |
| `POST /api/groups/:id/transfer-leadership` | Chuyển chủ xị khi không còn settlement mở (chỉ áp dụng single-creditor) |

### Khoản chi và kỳ chốt

| Method & path | Chức năng |
|---|---|
| `POST /api/groups/:id/expenses` | Tạo khoản chi và phần chia (chủ xị trong single-creditor, bất kỳ thành viên nào trong multi-creditor) |
| `PATCH /api/groups/:id/expenses/:expenseId` | Sửa khoản chưa chốt |
| `GET /api/groups/:id/balances` | Xem nợ chưa chốt theo thành viên |
| `POST /api/groups/:id/settlement-batches` | Chốt nợ và tạo settlement |
| `GET /api/groups/:id/settlement-batches/:batchId` | Xem tiến độ thu tiền của kỳ |

### Thanh toán

| Method & path | Chức năng |
|---|---|
| `POST /api/settlements/:id/mark-sent` | Thành viên báo đã chuyển khoản |
| `POST /api/settlements/:id/confirm` | Người nhận xác nhận đã nhận tiền |
| `POST /api/settlements/:id/reject` | Người nhận từ chối yêu cầu |

Các lệnh tạo/chốt/xác nhận nên nhận idempotency key từ client để tránh nhân đôi khi mạng
chập chờn hoặc người dùng bấm lại.

## 11. Yêu cầu bảo mật và độ tin cậy

- API trả dữ liệu theo nguyên tắc tối thiểu: thành viên không xem dữ liệu không liên quan
  đến mình (trừ chế độ "ai nợ ai" nơi mọi người đều xem được toàn bộ).
- Mã thanh toán cần đủ entropy, không chứa thông tin cá nhân.
- Mọi hành động tài chính quan trọng phải có audit log.
- Tiền dùng integer/decimal chính xác; validate currency và không cho số âm hoặc bằng 0.
- Backend kiểm tra quyền từ membership hiện tại cho mọi request.

## 12. Yêu cầu UX

### Dashboard chủ xị (chế độ "chủ xị")

- Tổng đã ứng, tổng chưa thu, tổng đã thu.
- Danh sách ai chưa trả, ai báo đã chuyển và ai đã được xác nhận.
- Cảnh báo và hướng dẫn đóng kỳ trước khi đổi chủ xị.

### Dashboard nhóm (chế độ "ai nợ ai")

- Tổng chi tiêu của nhóm, net balance của từng thành viên.
- Danh sách giao dịch thanh toán tối giản (ai trả ai, bao nhiêu).
- Mọi thành viên đều xem được toàn bộ trạng thái hoàn tiền của nhóm.
- Nút "Tôi đã chuyển khoản" cho từng settlement của mình.

### Dashboard thành viên

- Ưu tiên con số "Bạn cần trả" và "Bạn sẽ nhận" thay vì bảng balance phức tạp.
- Hiển thị rõ người nhận/người trả, số tiền, mã thanh toán.
- Phân biệt "Đã báo chuyển" với "Đã xác nhận"; không dùng cùng màu/trạng thái.

### Trạng thái rỗng và lỗi

- Chưa có khoản chi: hướng dẫn thêm khoản đầu tiên.
- Sai nội dung chuyển khoản: hướng dẫn liên hệ người nhận để xác nhận thủ công.

## 13. Các trường hợp biên cần xử lý

- Thành viên rời nhóm khi còn nợ hoặc còn settlement mở.
- Người trả tự đưa mình ra khỏi phần chia của một expense.
- Tổng CUSTOM/PERCENT không bằng tổng expense do làm tròn.
- Thành viên trả thiếu, trả thừa hoặc chia thành nhiều lần chuyển.
- Khoản chi bị sửa sau khi đã chốt.
- Kỳ chốt bị hủy khi một số settlement đã `PAID`.

Với MVP, không hỗ trợ trả một settlement bằng nhiều giao dịch hoặc gộp nhiều settlement
vào một giao dịch. Các trường hợp trả thiếu/thừa do người nhận xử lý thủ công; không tự
động suy diễn.

## 14. Tiêu chí nghiệm thu MVP

- Tạo nhóm tự gán người tạo làm thành viên đầu tiên (và chủ xị nếu single-creditor) trong
  một transaction.
- Thành viên không thể gọi API tạo expense hoặc xác nhận nhận tiền (trừ chế độ ai nợ ai).
- Tạo expense với các kiểu chia và backend từ chối tổng phần chia sai.
- Chốt nợ tạo đúng settlement theo chế độ: hình sao về chủ xị (single-creditor) hoặc
  min-cash-flow (multi-creditor).
- Mỗi settlement có mã duy nhất.
- Thành viên báo đã trả → `AWAITING_CONFIRMATION`; người nhận xác nhận → `PAID`.
- Không đổi được chủ xị khi còn settlement mở; lịch sử không đổi sau khi chuyển vai trò.
- Tất cả hành động tài chính và đổi vai trò có audit log.

## 15. Lộ trình đề xuất

### Giai đoạn 1: Core — trải nghiệm chia tiền cơ bản

| Hạng mục | Nội dung |
|---|---|
| 1.1 Nhóm tự do | Tạo nhóm không cần liên kết gì, chọn chế độ nợ (`SINGLE_CREDITOR` hoặc `MULTI_CREDITOR`), membership, mã mời |
| 1.2 Chi tiêu | Expense/split với các kiểu chia (đều, %, trọng số, tùy chỉnh), kiểm tra tổng tiền |
| 1.3 Tính nợ | Net balance engine chung cho cả 2 chế độ |
| 1.4 Resolver | Chốt nợ: single-creditor (hình sao về chủ xị) + multi-creditor (min-cash-flow) |
| 1.5 Thanh toán thủ công | Payment code, đánh dấu đã trả, xác nhận đã nhận, audit log |
| 1.6 Chuyển chủ xị | Quy tắc đóng kỳ, lịch sử bất biến (chỉ cho single-creditor) |

### Giai đoạn 2: Trải nghiệm nâng cao

| Hạng mục | Nội dung |
|---|---|
| 2.1 Thông báo | Push notification khi có khoản chi mới, khi được nhắc trả tiền, khi được xác nhận |
| 2.2 Ảnh biên lai | Upload ảnh khi tạo khoản chi hoặc khi báo đã chuyển khoản |
| 2.3 Lịch sử & báo cáo | Xem lại lịch sử chi tiêu theo nhóm, export CSV |
| 2.4 Nhắc nợ | Gửi nhắc nhở tự động tới thành viên chưa thanh toán |

### Giai đoạn 3: Mở rộng & hoàn thiện

| Hạng mục | Nội dung |
|---|---|
| 3.1 PWA polish | Offline mode, cài đặt lên màn hình chính, animation |
| 3.2 Đa tiền tệ | Hỗ trợ nhóm dùng USD, EUR... với tỉ giá |
| 3.3 Hoàn thiện | Rate limit, quan sát hệ thống, test tích hợp |

## 16. Xác thực và chạy dự án hiện tại

Frontend gửi login tới proxy same-origin `/api/auth/login` với `credentials: include`.
Backend dùng access/refresh JWT trong cookie HttpOnly; token không được lưu trong Web
Storage. Protected layout gọi `/api/auth/me` để xác minh user thực tế.

Cấu hình production cần hai JWT secret độc lập dài ít nhất 32 ký tự, HTTPS
(`COOKIE_SECURE=true`), một `FRONTEND_ORIGIN` cụ thể và browser API URL `/api`. Xem
`.env.example` cho các biến môi trường hiện có.

### Local development

```bash
# Khởi động PostgreSQL
docker-compose up -d postgres

# Chạy backend
cd backend && go run .

# Chạy frontend
cd frontend && npm install && npm run dev
```

### Deploy hiện tại

`vercel.json` ở thư mục gốc khai báo Next.js frontend và Go backend. API backend triển
khai dưới prefix `/api/backend`; health check là `GET /api/backend/health`. Các API nghiệp
vụ trong mục 10 chỉ là hợp đồng mục tiêu cho các giai đoạn tiếp theo.
