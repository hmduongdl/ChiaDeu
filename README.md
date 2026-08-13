# Chia Đều — chia tiền nhóm bạn bè / đi chơi

*Tài liệu định hướng sản phẩm, yêu cầu nghiệp vụ và thiết kế hệ thống dự kiến*

> Trạng thái hiện tại: dự án đã có phần xác thực tài khoản cơ bản. Các API nhóm, khoản chi,
> thanh toán, SePay và PayOS trong tài liệu này là **thiết kế mục tiêu**, chưa phải tính năng
> backend đã hoàn thiện.

---

## 1. Tầm nhìn sản phẩm

Chia Đều số hóa cách một nhóm bạn thường chia tiền sau mỗi lần đi chơi, ăn uống, du lịch:

- App hỗ trợ hai chế độ tính nợ linh hoạt:
  - **Chế độ “chủ xị” (single-creditor):** một người đứng ra ứng tiền và thanh toán các khoản
    chi chung của nhóm. Các thành viên chỉ cần xem phần mình phải chịu và hoàn tiền lại cho
    chủ xị. Chỉ tài khoản nhận tiền của chủ xị cần được cấu hình cho nhóm.
  - **Chế độ “ai nợ ai” (multi-creditor):** nhiều thành viên có thể thay nhau ứng tiền.
    Hệ thống dùng thuật toán min-cash-flow để tối giản số giao dịch thanh toán giữa các
    thành viên, giống cách Splitwise vận hành.
- Cả hai chế độ dùng chung một tầng tính net balance, chỉ khác nhau ở tầng “resolver” khi
  sinh danh sách giao dịch thanh toán.
- SePay là tính năng optional, tích hợp ở bước thanh toán để tạo QR/link và tự động đối
  soát tiền vào. Nhóm vẫn dùng được hoàn toàn ở chế độ thủ công nếu không liên kết SePay.

Mục tiêu của sản phẩm là một ứng dụng chia tiền nhóm bạn bè đơn giản, dễ dùng, không yêu cầu
người dùng phải liên kết tài khoản ngân hàng hay cổng thanh toán để bắt đầu sử dụng.

## 2. Thuật ngữ

| Thuật ngữ | Ý nghĩa |
|---|---|
| Nhóm (`Group`) | Không gian ghi nhận khoản chi và công nợ của một nhóm người |
| Chủ xị (`Leader`) | Trong chế độ "chủ xị": thành viên đang ứng tiền, quản lý nhóm và nhận tiền hoàn lại |
| Thành viên (`Member`) | Người tham gia chia chi phí, có thể là người ứng tiền (chế độ ai nợ ai) hoặc người trả nợ |
| Chế độ nợ (`Settlement mode`) | `SINGLE_CREDITOR` (chủ xị) hoặc `MULTI_CREDITOR` (ai nợ ai), chọn khi tạo nhóm |
| Khoản chi (`Expense`) | Một lần thành viên đã thanh toán cho nhóm |
| Phần chia (`Expense split`) | Số tiền một thành viên phải chịu trong một khoản chi |
| Kỳ chốt (`Settlement batch`) | Ảnh chụp công nợ tại thời điểm yêu cầu thanh toán |
| Net balance | Số dư ròng của mỗi thành viên sau khi tổng hợp mọi khoản chi và phần chia |
| Khoản hoàn (`Settlement`) | Khoản một thành viên phải trả cho người khác trong một kỳ chốt |
| Mã thanh toán | Mã duy nhất, ví dụ `CD4F82A9`, dùng trong nội dung chuyển khoản |
| Đối soát | Ghép giao dịch tiền vào với đúng khoản hoàn và cập nhật trạng thái |

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
5. Thành viên không phải liên kết SePay, không nhập API key và không cấu hình tài khoản
   nhận tiền.
6. Tài khoản nhận tiền và cấu hình SePay thuộc phạm vi nhóm, không phải cấu hình toàn cục
   áp dụng cho mọi nhóm mà người dùng tham gia.
7. Mỗi khoản hoàn có một mã thanh toán duy nhất, không tái sử dụng.
8. Chỉ webhook có mã khớp chính xác mới được tự động đánh dấu `PAID`.
9. Khớp theo số tiền và thời gian chỉ tạo **gợi ý đối soát**, không tự động xác nhận khi
   thiếu mã hoặc còn nhiều ứng viên.
10. Nhóm không liên kết SePay vẫn tạo khoản chi, chốt nợ, sinh thông tin chuyển khoản và
    xác nhận thủ công bình thường.
11. Tiền được lưu bằng đơn vị nhỏ nhất của tiền tệ. Với VND, backend dùng số nguyên đồng;
    không dùng `float` cho tính toán tài chính.
12. Dữ liệu lịch sử phải giữ nguyên người trả, người nhận và cấu hình thanh toán tại thời
    điểm phát sinh, kể cả sau khi đổi chủ xị hoặc đổi chế độ nợ.

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
| Cấu hình tài khoản nhận tiền của nhóm | Có | Không |
| Liên kết/ngắt SePay cho nhóm | Có | Không |
| Xem toàn bộ trạng thái hoàn tiền | Có | Chỉ khoản của mình (*) |
| Yêu cầu xác nhận đã chuyển thủ công | Không | Có |
| Xác nhận/từ chối yêu cầu thủ công | Có | Không |
| Chuyển vai trò chủ xị | Có | Không |
| Rời nhóm | Chỉ sau khi chuyển vai trò | Khi không còn nghĩa vụ đang mở |

(*) Trong chế độ "ai nợ ai", mọi thành viên đều có quyền tạo khoản chi và xem toàn bộ
trạng thái hoàn tiền. |

Mọi quyền đều phải được backend kiểm tra từ membership hiện tại. Frontend ẩn nút không
được xem là biện pháp phân quyền.

## 5. Luồng sử dụng chính

### 5.1 Tạo nhóm và thiết lập

1. Người dùng tạo nhóm, chọn chế độ nợ ("chủ xị" hoặc "ai nợ ai") và tự động trở thành
   thành viên đầu tiên. Nếu chọn chế độ "chủ xị", người tạo đồng thời là chủ xị đầu tiên.
2. Hệ thống sinh mã mời; người khác tham gia bằng mã hoặc liên kết mời.
3. Trong chế độ "chủ xị", chủ xị có thể cấu hình tài khoản nhận tiền gồm tên ngân hàng,
   số tài khoản và tên chủ tài khoản. Trong chế độ "ai nợ ai", mỗi thành viên có thể cấu
   hình tài khoản nhận tiền riêng nếu muốn nhận thanh toán qua app.
4. Không cần liên kết SePay hay bất kỳ cổng thanh toán nào để tạo nhóm và bắt đầu sử dụng.
   SePay là bước tích hợp optional ở giai đoạn sau.
5. Nếu liên kết SePay, backend lưu định danh tài khoản tích hợp và bắt đầu nhận webhook
   tiền vào của tài khoản đó.

Một người có thể là chủ xị ở nhóm A nhưng chỉ là thành viên ở nhóm B. Một tài khoản ngân
hàng có thể được dùng cho nhiều nhóm; mã thanh toán vẫn giúp route giao dịch về đúng
`group_id` và `settlement_id`.

### 5.2 Ghi nhận khoản chi

1. Người tạo khoản chi (chủ xị trong chế độ "chủ xị", hoặc bất kỳ thành viên nào trong
   chế độ "ai nợ ai") nhập mô tả, tổng tiền, ngày chi và người tham gia.
2. Chọn cách chia: đều, phần trăm, theo trọng số hoặc số tiền tùy chỉnh.
3. Backend kiểm tra tổng các phần chia phải bằng tổng khoản chi.
4. Phần của chính người trả (`paid_by`) được ghi nhận là chi phí cá nhân, không tạo khoản
   hoàn cho chính họ.
5. Phần của mỗi thành viên khác làm tăng net balance người đó nợ người trả.

Nguồn khoản chi có thể là nhập tay hoặc một giao dịch tiền ra đã đồng bộ. Việc đồng bộ
giao dịch tiền ra là tiện ích, không phải điều kiện để tạo khoản chi.

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
- Tên người nhận và tài khoản nhận tiền (chủ xị trong chế độ “chủ xị”, hoặc thành viên
  được nhận trong chế độ “ai nợ ai”).
- Mã thanh toán bắt buộc.
- QR có sẵn số tài khoản, số tiền và nội dung chuyển khoản nếu cấu hình hỗ trợ.
- Trạng thái cập nhật gần thời gian thực.
- Nút “Tôi đã chuyển khoản” cho chế độ thủ công hoặc khi tự động đối soát chưa nhận ra.

QR chỉ là phương tiện điền thông tin chuyển khoản. Cần xác minh riêng khả năng của từng
nhà cung cấp trước khi khẳng định PayOS có thể chuyển tiền trực tiếp về tài khoản động của
từng chủ xị. MVP có thể dùng VietQR theo tài khoản nhóm và SePay để đối soát tiền vào;
PayOS chỉ bật khi mô hình tài khoản người nhận và hợp đồng tích hợp đã được xác nhận.

### 5.5 Đối soát tự động qua SePay

Khi nhận webhook tiền vào, backend xử lý theo thứ tự:

1. Xác minh webhook theo cơ chế được nhà cung cấp hỗ trợ và từ chối payload không hợp lệ.
2. Dùng mã giao dịch bên cung cấp làm idempotency key để không ghi nhận hai lần.
3. Xác định cấu hình nhận tiền từ tài khoản đã nhận giao dịch (theo nhóm trong chế độ
   "chủ xị", hoặc theo cá nhân trong chế độ "ai nợ ai").
4. Chuẩn hóa nội dung và tìm mã thanh toán theo định dạng cho phép.
5. Yêu cầu mã thuộc một settlement `PENDING`, đúng nhóm và số tiền khớp tuyệt đối.
6. Lưu giao dịch ngân hàng, liên kết settlement rồi chuyển settlement sang `PAID` trong
   cùng một database transaction.
7. Ghi nguồn xác nhận là `SEPAY_AUTO` và thời điểm thanh toán.

Nếu không có mã, sai số tiền, trùng ứng viên hoặc settlement đã đóng, giao dịch đi vào
hàng đợi `UNMATCHED`/`REVIEW_REQUIRED`. Hệ thống có thể xếp hạng gợi ý dựa trên số tiền,
thời gian và tên người gửi, nhưng chỉ chủ xị được quyết định ghép thủ công.

### 5.6 Xác nhận thủ công

1. Thành viên nhấn “Tôi đã chuyển khoản”, có thể đính kèm ghi chú hoặc ảnh biên lai.
2. Settlement chuyển từ `PENDING` sang `AWAITING_CONFIRMATION`; trạng thái này chưa được
   tính là đã thanh toán.
3. Người nhận tiền (chủ xị trong chế độ “chủ xị”, hoặc thành viên được nhận trong chế độ
   “ai nợ ai”) kiểm tra tài khoản và chọn xác nhận hoặc từ chối.
4. Xác nhận chuyển thành `PAID` với nguồn `MANUAL_CONFIRMATION`; từ chối đưa về `PENDING`
   và lưu lý do.

Thành viên không được tự đánh dấu `PAID`, vì đây là trạng thái xác nhận tiền đã đến người
nhận chứ không chỉ là tuyên bố đã gửi.

## 6. Trạng thái nghiệp vụ

### 6.1 Settlement

```text
PENDING ──thành viên báo đã trả──> AWAITING_CONFIRMATION
   │                                      │
   │ webhook khớp chính xác               ├──người nhận từ chối──> PENDING
   │                                      │
   ├──────────────────────────────────────┴──xác nhận──> PAID
   │
   └──hủy kỳ chốt──────────────────────────────────────> CANCELLED
```

`PAID` và `CANCELLED` là trạng thái kết thúc. Sửa một settlement đã kết thúc phải thông
qua nghiệp vụ điều chỉnh có audit log, không cập nhật trực tiếp làm mất lịch sử.

### 6.2 Cấu hình thanh toán

Trong chế độ "chủ xị", cấu hình thanh toán thuộc phạm vi nhóm. Trong chế độ "ai nợ ai",
mỗi thành viên có thể cấu hình tài khoản nhận tiền riêng (opt-in cá nhân).

| Trạng thái | Ý nghĩa |
|---|---|
| `MANUAL` | Có thể hiện thông tin chuyển khoản nhưng không tự động đối soát |
| `CONNECTING` | Đang thực hiện luồng liên kết nhà cung cấp |
| `ACTIVE` | Webhook đang hoạt động và có thể tự đối soát |
| `ERROR` | Liên kết lỗi, hết quyền hoặc webhook cần kiểm tra |
| `DISCONNECTED` | Đã ngắt; quay về xác nhận thủ công |

## 7. Quy tắc đổi chủ xị (chỉ áp dụng cho chế độ "chủ xị")

Đổi chủ xị là thay đổi người ứng tiền và người nhận của **các hoạt động tương lai**, không
được làm thay đổi lịch sử.

Để tránh thành viên quét QR cũ nhưng tiền đi vào tài khoản mới, bản MVP áp dụng quy tắc:

1. Người nhận vai trò mới phải đang là thành viên nhóm.
2. Không cho đổi khi còn kỳ chốt có settlement `PENDING` hoặc
   `AWAITING_CONFIRMATION`.
3. Chủ xị cũ phải hoàn tất hoặc hủy kỳ đang mở trước khi chuyển vai trò.
4. Sau khi chuyển, cấu hình SePay cũ bị ngắt khỏi hoạt động mới.
5. Chủ xị mới cấu hình tài khoản nhận tiền hoặc chọn chế độ thủ công.
6. Expense, batch, settlement và bank transaction cũ giữ snapshot người nhận cũ.
7. Mọi lần chuyển vai trò phải ghi audit log gồm người thao tác, người cũ, người mới và
   thời gian.

Giai đoạn sau có thể hỗ trợ nhiều “kỳ chủ xị” đồng thời, nhưng đổi lại UI và webhook
routing phức tạp hơn. MVP ưu tiên quy tắc đóng kỳ trước khi chuyển.

## 8. Kiến trúc dự kiến

```text
┌──────────────────────────────────┐
│ Next.js PWA                      │
│ Nhóm, khoản chi, QR, trạng thái  │
│ hoàn tiền (2 chế độ nợ)          │
└──────────────┬───────────────────┘
               │ REST/JSON + HttpOnly cookie
┌──────────────▼───────────────────┐
│ Go/Fiber API                     │
│ Auth, phân quyền, tính net       │
│ balance, min-cash-flow resolver, │
│ webhook và đối soát              │
└───────┬─────────────┬────────────┘
        │             │
┌───────▼───────┐ ┌───▼─────────────────────┐
│ PostgreSQL    │ │ SePay / QR provider     │
│ Dữ liệu + log │ │ (optional, tích hợp sau)│
└───────────────┘ └─────────────────────────┘
```

### Công nghệ

| Thành phần | Công nghệ dự kiến | Trách nhiệm |
|---|---|---|
| Frontend | Next.js, React, TypeScript, Tailwind CSS | PWA và trải nghiệm theo vai trò/chế độ nợ |
| Backend | Go, Fiber | API, kiểm tra quyền, tính net balance, min-cash-flow và xử lý webhook |
| Database | PostgreSQL | Giao dịch dữ liệu, ràng buộc và audit |
| Auth | JWT ngắn hạn + refresh cookie HttpOnly | Xác thực phiên hiện có |
| Đối soát | SePay (optional) | Nhận biến động tiền vào, tự động đối soát |
| QR/payment link | VietQR hoặc PayOS sau khi xác minh mô hình tích hợp | Điền sẵn thông tin hoàn tiền |

## 9. Mô hình dữ liệu đề xuất

Đây là schema mục tiêu cho migration tương lai. Migration hiện tại vẫn phản ánh mô hình cũ
và **chưa được xem là đã đáp ứng tài liệu này**.

### 9.1 `users`

Chỉ lưu hồ sơ và thông tin xác thực người dùng. Không đặt `sepay_account_id` hoặc tài khoản
nhận tiền mặc định ở đây làm nguồn sự thật cho settlement.

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

Ràng buộc “leader phải là member” nên được bảo vệ bằng transaction/service rule; nếu dùng
ràng buộc database thì phải thiết kế khóa phù hợp để tránh trạng thái trung gian khi tạo nhóm.

### 9.3 `group_payment_profiles`

Một cấu hình nhận tiền cho mỗi nhóm (chế độ "chủ xị") hoặc mỗi thành viên (chế độ
"ai nợ ai", opt-in cá nhân). Dữ liệu bí mật của nhà cung cấp phải mã hóa ở tầng ứng
dụng hoặc lưu trong secret manager, không trả về frontend.

| Cột chính | Ý nghĩa |
|---|---|
| `group_id` | Phạm vi cấu hình (nullable nếu là cấu hình cá nhân) |
| `user_id` | Chủ sở hữu cấu hình (luôn có) |
| `receiver_user_id` | Người nhận tiền tại thời điểm kích hoạt |
| `bank_code`, `bank_account_no`, `account_name` | Thông tin sinh QR/hiển thị |
| `provider`, `provider_account_id` | Nhà cung cấp và ID tích hợp, nullable ở manual mode |
| `status`, `connected_at`, `disconnected_at` | Vòng đời liên kết |

### 9.4 `expenses` và `expense_splits`

| Bảng/cột chính | Ý nghĩa |
|---|---|
| `expenses.group_id` | Nhóm sở hữu khoản chi |
| `expenses.paid_by` | Người đã trả khoản này (chủ xị trong chế độ "chủ xị", bất kỳ thành viên nào trong chế độ "ai nợ ai") |
| `expenses.amount_minor` | Tổng tiền bằng số nguyên đơn vị nhỏ nhất |
| `expenses.split_type`, `description`, `expense_date` | Quy tắc chia và nội dung |
| `expense_splits.user_id`, `share_minor` | Phần mỗi người phải chịu |
| `expense_splits.settlement_batch_id` | Nullable cho đến khi được chốt |

### 9.5 `settlement_batches` và `settlements`

| Bảng/cột chính | Ý nghĩa |
|---|---|
| `settlement_batches.leader_id` | Snapshot chủ xị của kỳ chốt |
| `settlement_batches.status` | `OPEN`, `COMPLETED`, `CANCELLED` |
| `settlements.from_user_id` | Thành viên trả tiền |
| `settlements.to_user_id` | Người nhận tiền (chủ xị trong chế độ "chủ xị", thành viên được nhận trong chế độ "ai nợ ai"), snapshot bất biến |
| `settlements.amount_minor` | Số tiền phải trả |
| `settlements.payment_code` | Mã duy nhất, không đoán được và dễ nhập |
| `settlements.status` | Trạng thái theo mục 6 |
| `settlements.confirmation_source` | `SEPAY_AUTO`, `MANUAL_CONFIRMATION` hoặc null |
| `settlements.paid_at` | Thời điểm xác nhận tiền đến |

Nên có unique key `(batch_id, from_user_id)` để một thành viên chỉ có một settlement trong
mỗi kỳ và unique index không phân biệt hoa/thường cho `payment_code`.

### 9.6 `bank_transactions`, `reconciliation_candidates` và `audit_logs`

| Bảng | Mục đích |
|---|---|
| `bank_transactions` | Lưu webhook đã chuẩn hóa, raw payload được bảo vệ, provider transaction ID unique |
| `reconciliation_candidates` | Gợi ý ghép không chắc chắn, điểm tin cậy và quyết định của chủ xị |
| `audit_logs` | Lịch sử đổi vai trò, sửa chi phí, xác nhận/từ chối và ghép giao dịch |

`bank_transactions` cần giữ `group_id`, payment profile và receiver snapshot để dữ liệu cũ
không bị route lại khi nhóm đổi chủ xị hoặc thành viên thay đổi cấu hình nhận tiền.

## 10. API mục tiêu

Các endpoint dưới đây là hợp đồng dự kiến, chưa khẳng định đã tồn tại trong backend.

### Nhóm và vai trò

| Method & path | Chức năng |
|---|---|
| `POST /api/groups` | Tạo nhóm; body gồm `settlement_mode` (`SINGLE_CREDITOR` hoặc `MULTI_CREDITOR`). Người tạo là thành viên đầu tiên (và là chủ xị nếu chọn single-creditor) |
| `POST /api/groups/join/:shareCode` | Tham gia nhóm |
| `GET /api/groups/:id` | Lấy nhóm, vai trò, thành viên và trạng thái payment profile |
| `POST /api/groups/:id/transfer-leadership` | Chuyển chủ xị khi không còn settlement mở (chỉ áp dụng single-creditor) |

### Cấu hình nhận tiền

| Method & path | Chức năng |
|---|---|
| `PUT /api/groups/:id/payment-profile` | Cập nhật thông tin nhận tiền cho nhóm (single-creditor) |
| `PUT /api/users/me/payment-profile` | Cập nhật thông tin nhận tiền cá nhân (multi-creditor, opt-in) |
| `POST /api/groups/:id/sepay/connect` | Bắt đầu liên kết SePay (optional) |
| `DELETE /api/groups/:id/sepay/connect` | Ngắt liên kết và chuyển về manual mode |

### Khoản chi và kỳ chốt

| Method & path | Chức năng |
|---|---|
| `POST /api/groups/:id/expenses` | Tạo khoản chi và phần chia (chủ xị trong single-creditor, bất kỳ thành viên nào trong multi-creditor) |
| `PATCH /api/groups/:id/expenses/:expenseId` | Sửa khoản chưa chốt |
| `GET /api/groups/:id/balances` | Xem nợ chưa chốt theo thành viên |
| `POST /api/groups/:id/settlement-batches` | Chốt nợ và tạo settlement hình sao |
| `GET /api/groups/:id/settlement-batches/:batchId` | Xem tiến độ thu tiền của kỳ |

### Thanh toán và đối soát

| Method & path | Chức năng |
|---|---|
| `GET /api/settlements/:id/payment-instructions` | Trả QR/payload và thông tin người nhận |
| `POST /api/settlements/:id/mark-sent` | Thành viên yêu cầu xác nhận thủ công |
| `POST /api/settlements/:id/confirm` | Chủ xị xác nhận đã nhận tiền |
| `POST /api/settlements/:id/reject` | Chủ xị từ chối yêu cầu thủ công |
| `POST /api/webhooks/sepay` | Nhận webhook công khai, có xác minh và idempotency |
| `GET /api/groups/:id/reconciliation` | Chủ xị xem giao dịch chưa khớp và gợi ý |
| `POST /api/groups/:id/reconciliation/:transactionId/match` | Chủ xị ghép thủ công với settlement |

Các lệnh tạo/chốt/xác nhận nên nhận idempotency key từ client để tránh nhân đôi khi mạng
chập chờn hoặc người dùng bấm lại.

## 11. Yêu cầu bảo mật và độ tin cậy

- Không lưu API key, access token nhà cung cấp hoặc raw webhook nhạy cảm dưới dạng plain
  text trong log.
- Webhook không dùng phiên đăng nhập người dùng nhưng phải có xác minh nhà cung cấp,
  chống replay nếu giao thức hỗ trợ và giới hạn kích thước payload.
- Mọi webhook phải idempotent theo provider transaction ID.
- Cập nhật bank transaction và settlement phải nằm trong cùng database transaction.
- API trả dữ liệu theo nguyên tắc tối thiểu: thành viên không xem cấu hình bí mật hoặc giao
  dịch ngân hàng không liên quan.
- QR và payment instruction phải được tạo từ snapshot của settlement, không đọc mù quáng
  cấu hình chủ xị mới nhất.
- Mã thanh toán cần đủ entropy, không chứa thông tin cá nhân và được parse không phân biệt
  hoa/thường nhưng không match theo chuỗi con quá rộng.
- Mọi hành động tài chính quan trọng phải có audit log.
- Tiền dùng integer/decimal chính xác; validate currency và không cho số âm hoặc bằng 0.

## 12. Yêu cầu UX

### Dashboard chủ xị (chế độ "chủ xị")

- Tổng đã ứng, tổng chưa thu, tổng đã thu và giao dịch cần xem xét.
- Trạng thái liên kết ngân hàng rõ ràng; lỗi SePay không chặn thao tác thủ công.
- Danh sách ai chưa trả, ai báo đã chuyển và ai đã được xác nhận.
- Cảnh báo và hướng dẫn đóng kỳ trước khi đổi chủ xị.

### Dashboard nhóm (chế độ "ai nợ ai")

- Tổng chi tiêu của nhóm, net balance của từng thành viên.
- Danh sách giao dịch thanh toán tối giản (ai trả ai, bao nhiêu).
- Mọi thành viên đều xem được toàn bộ trạng thái hoàn tiền của nhóm.
- Nút "Tôi đã chuyển khoản" cho từng settlement của mình.

### Dashboard thành viên

- Ưu tiên con số “Bạn cần trả” và “Bạn sẽ nhận” thay vì bảng balance phức tạp.
- Hiển thị rõ người nhận/người trả, số tiền, mã thanh toán và một nút quét/sao chép.
- Phân biệt “Đã báo chuyển” với “Đã xác nhận”; không dùng cùng màu/trạng thái.
- Không xuất hiện màn hình liên kết SePay ép buộc; SePay là opt-in cá nhân.

### Trạng thái rỗng và lỗi

- Chưa có khoản chi: hướng dẫn chủ xị thêm khoản đầu tiên.
- Chưa liên kết SePay: giải thích nhóm vẫn hoạt động ở chế độ thủ công.
- Webhook chậm: cho phép báo đã chuyển và hiển thị “đang chờ đối soát”.
- Sai nội dung chuyển khoản: hướng dẫn liên hệ chủ xị để ghép thủ công, không yêu cầu trả
  lần hai ngay lập tức.

## 13. Các trường hợp biên cần xử lý

- Thành viên rời nhóm khi còn nợ hoặc còn settlement mở.
- Chủ xị tự đưa mình ra khỏi phần chia của một expense.
- Tổng CUSTOM/PERCENT không bằng tổng expense do làm tròn.
- Thành viên trả thiếu, trả thừa hoặc chia thành nhiều lần chuyển.
- Hai settlement cùng số tiền được chuyển gần nhau nhưng không có mã.
- Một giao dịch chứa nhiều chuỗi giống mã thanh toán.
- Webhook bị gửi lặp, đến sai thứ tự hoặc đến sau khi đã xác nhận thủ công.
- Chủ xị nhận tiền thủ công rồi sau đó webhook của cùng giao dịch xuất hiện.
- Khoản chi bị sửa sau khi đã chốt.
- Kỳ chốt bị hủy khi một số settlement đã `PAID`.
- Tài khoản ngân hàng của chủ xị thay đổi trong khi còn QR chưa thanh toán.
- Một payment profile được dùng bởi nhiều nhóm.

Với MVP, không hỗ trợ trả một settlement bằng nhiều giao dịch hoặc gộp nhiều settlement
vào một giao dịch. Các trường hợp trả thiếu/thừa đi vào hàng đợi để chủ xị xử lý thủ công;
không tự động suy diễn.

## 14. Tiêu chí nghiệm thu MVP

- Tạo nhóm tự gán người tạo làm chủ xị và thành viên đầu tiên trong một transaction.
- Thành viên không thể gọi API tạo expense, cấu hình ngân hàng hoặc xác nhận nhận tiền.
- Chủ xị tạo expense với các kiểu chia và backend từ chối tổng phần chia sai.
- Chốt nợ tạo đúng một settlement cho mỗi thành viên còn nợ, tất cả cùng trả về chủ xị.
- Mỗi settlement có mã duy nhất và payment instruction đúng snapshot người nhận.
- Webhook lặp không tạo bank transaction hoặc ghi nhận thanh toán lần hai.
- Webhook mã đúng + tiền đúng tự động chuyển `PENDING` thành `PAID`.
- Webhook thiếu mã/sai tiền không tự động `PAID` và xuất hiện trong hàng đợi xem xét.
- Manual mode hoạt động khi chưa có SePay.
- Không đổi được chủ xị khi còn settlement mở; lịch sử không đổi sau khi chuyển vai trò.
- Tất cả hành động tài chính và đổi vai trò có audit log.

## 15. Lộ trình đề xuất

### Giai đoạn 1: Core — trải nghiệm chia tiền cơ bản

| Hạng mục | Nội dung |
|---|---|
| 1.1 Nhóm tự do | Tạo nhóm không cần liên kết ngân hàng, chọn chế độ nợ (`SINGLE_CREDITOR` hoặc `MULTI_CREDITOR`), membership, mã mời |
| 1.2 Chi tiêu | Expense/split với các kiểu chia (đều, %, trọng số, tùy chỉnh), kiểm tra tổng tiền |
| 1.3 Tính nợ | Net balance engine chung cho cả 2 chế độ |
| 1.4 Resolver | Chốt nợ: single-creditor (hình sao về chủ xị) + multi-creditor (min-cash-flow) |
| 1.5 Thanh toán thủ công | Payment code, QR tĩnh, đánh dấu đã trả/xác nhận thủ công, audit log |
| 1.6 Chuyển chủ xị | Quy tắc đóng kỳ, lịch sử bất biến (chỉ cho single-creditor) |

### Giai đoạn 2: Tích hợp SePay optional

| Hạng mục | Nội dung |
|---|---|
| 2.1 Liên kết cá nhân | Opt-in SePay cho từng thành viên (không ép buộc), quản lý API key an toàn |
| 2.2 QR/link thanh toán | QR động và payment link tích hợp thông tin chuyển khoản |
| 2.3 Đối soát tự động | Webhook, exact match theo mã + số tiền, idempotency, review queue |
| 2.4 Cấu hình nhận tiền | Payment profile theo nhóm (single-creditor) hoặc cá nhân (multi-creditor) |

### Giai đoạn 3: Monetize qua SePay

| Hạng mục | Nội dung |
|---|---|
| 3.1 Premium | Tính năng nâng cao cho người dùng trả phí (báo cáo, xuất dữ liệu, ưu tiên hỗ trợ) |
| 3.2 Phí giao dịch | Phí nhỏ trên giao dịch qua SePay, miễn phí cho chế độ thủ công |
| 3.3 Hoàn thiện | Thông báo, rate limit, quan sát hệ thống, test tích hợp và PWA polish |

Nên hoàn thành luồng thủ công (giai đoạn 1) trước khi tích hợp SePay. Khi đó sản phẩm đã
sử dụng được và webhook chỉ là lớp tự động hóa, không trở thành điểm lỗi duy nhất.

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
vụ chủ xị trong mục 10 chỉ là hợp đồng mục tiêu cho các giai đoạn tiếp theo.
