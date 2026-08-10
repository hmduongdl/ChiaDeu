-- ============================================================================
-- Cash Flow Minimizer: Schema khởi tạo cơ sở dữ liệu
-- ============================================================================
-- File này chứa toàn bộ cấu trúc bảng cho hệ thống chia tiền nhóm.
-- Mỗi bảng dùng IF NOT EXISTS để có thể chạy lại an toàn (idempotent).
-- Thứ tự bảng: users -> groups -> group_members -> bank_transactions
--               -> expenses -> expense_splits -> settlements
-- Yêu cầu: PostgreSQL 13+ với extension uuid-ossp.
-- ============================================================================

-- Bật extension sinh UUID tự động
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================================
-- Bảng users: Lưu thông tin người dùng
-- Mỗi user có thể liên kết với một tài khoản ngân hàng qua SePay.
-- Phone và email là unique, dùng để tìm kiếm/mời bạn bè vào nhóm.
-- ============================================================================
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(), -- ID duy nhất, tự sinh UUID v4
    name VARCHAR(100) NOT NULL,                     -- Tên hiển thị
    phone VARCHAR(20) UNIQUE,                       -- Số điện thoại, dùng để mời bạn
    email VARCHAR(100) UNIQUE,                      -- Email đăng nhập
    bank_account_no VARCHAR(50),                    -- Số tài khoản ngân hàng liên kết với SePay
    bank_code VARCHAR(20),                          -- Mã ngân hàng (VD: MBBank, VCB, ACB...)
    sepay_account_id VARCHAR(50),                   -- ID tài khoản đã đăng ký webhook trên SePay
    avatar_url TEXT,                                -- Ảnh đại diện (URL)
    created_at TIMESTAMPTZ DEFAULT now()            -- Thời điểm tạo tài khoản
);

-- ============================================================================
-- Bảng groups: Lưu thông tin nhóm chi tiêu
-- Mỗi nhóm có một share_code ngắn (6-10 ký tự) để mời thành viên.
-- ============================================================================
CREATE TABLE IF NOT EXISTS groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(), -- ID nhóm
    name VARCHAR(100) NOT NULL,                     -- Tên nhóm (VD: "Đà Lạt trip 2024")
    share_code VARCHAR(10) UNIQUE NOT NULL,         -- Mã mời ngắn, unique, dùng để join nhóm
    created_by UUID REFERENCES users(id),           -- Người tạo nhóm
    currency VARCHAR(10) DEFAULT 'VND',             -- Đơn vị tiền tệ, mặc định VND
    created_at TIMESTAMPTZ DEFAULT now()            -- Thời điểm tạo nhóm
);

-- ============================================================================
-- Bảng group_members: Liên kết many-to-many giữa groups và users
-- PK kép (group_id, user_id) đảm bảo mỗi user chỉ join 1 nhóm đúng 1 lần.
-- ON DELETE CASCADE: khi xóa group thì tự động xóa membership.
-- ============================================================================
CREATE TABLE IF NOT EXISTS group_members (
    group_id UUID REFERENCES groups(id) ON DELETE CASCADE, -- Nhóm
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,   -- Thành viên
    joined_at TIMESTAMPTZ DEFAULT now(),                    -- Thời điểm tham gia
    PRIMARY KEY (group_id, user_id)                         -- PK kép
);

-- ============================================================================
-- Bảng bank_transactions: Lưu toàn bộ giao dịch đồng bộ từ SePay
-- Đây là nguồn dữ liệu chính để người dùng chọn khi tạo Expense.
-- Webhook SePay gửi về mỗi khi có biến động số dư tài khoản ngân hàng.
-- Cột is_used = false: giao dịch chưa được gán vào Expense nào.
-- ============================================================================
CREATE TABLE IF NOT EXISTS bank_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),          -- ID nội bộ
    user_id UUID REFERENCES users(id),                      -- Chủ tài khoản phát sinh giao dịch
    sepay_transaction_id VARCHAR(50) UNIQUE,                -- ID giao dịch từ SePay (unique, tránh trùng lặp)
    amount NUMERIC(14,2) NOT NULL,                          -- Số tiền giao dịch (hỗ trợ đến 14 chữ số)
    transaction_type VARCHAR(10) CHECK (transaction_type IN ('IN','OUT')), -- IN = nhận tiền, OUT = chi tiền
    description TEXT,                                       -- Nội dung chuyển khoản gốc từ ngân hàng
    bank_account_no VARCHAR(50),                            -- Số tài khoản phát sinh giao dịch
    is_used BOOLEAN DEFAULT false,                          -- false = chưa gán vào expense nào, true = đã dùng
    transaction_time TIMESTAMPTZ,                           -- Thời điểm giao dịch thực tế (từ SePay)
    received_at TIMESTAMPTZ DEFAULT now()                   -- Thời điểm server nhận webhook
);

-- ============================================================================
-- Bảng expenses: Lưu các khoản chi cần chia trong nhóm
-- Mỗi expense liên kết với một nhóm, một người trả (paid_by),
-- và có thể liên kết với một giao dịch ngân hàng gốc (source_transaction_id).
-- split_type: EQUAL (chia đều), PERCENT (chia theo %), CUSTOM (chia tuỳ chỉnh).
-- ============================================================================
CREATE TABLE IF NOT EXISTS expenses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),                  -- ID khoản chi
    group_id UUID REFERENCES groups(id) ON DELETE CASCADE,          -- Thuộc nhóm nào
    source_transaction_id UUID REFERENCES bank_transactions(id),    -- Giao dịch ngân hàng gốc (NULL nếu nhập tay)
    paid_by UUID REFERENCES users(id),                              -- Người bỏ tiền trả
    description VARCHAR(255),                                       -- Mô tả khoản chi (VD: "Ăn tối", "Xăng xe")
    amount NUMERIC(14,2) NOT NULL,                                  -- Tổng số tiền chi
    split_type VARCHAR(20) DEFAULT 'EQUAL',                         -- Cách chia: EQUAL, PERCENT, CUSTOM
    created_at TIMESTAMPTZ DEFAULT now()                            -- Thời điểm tạo
);

-- ============================================================================
-- Bảng expense_splits: Chi tiết phần tiền từng người phải chịu trong 1 expense
-- Mỗi dòng = một người + số tiền người đó phải trả cho khoản chi này.
-- Tổng share_amount của tất cả splits trong 1 expense = expense.amount.
-- ============================================================================
CREATE TABLE IF NOT EXISTS expense_splits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),              -- ID dòng split
    expense_id UUID REFERENCES expenses(id) ON DELETE CASCADE, -- Thuộc expense nào
    user_id UUID REFERENCES users(id),                          -- Người phải chịu khoản này
    share_amount NUMERIC(14,2) NOT NULL                         -- Số tiền người này phải chịu
);

-- ============================================================================
-- Bảng settlements: Kết quả tính toán tối thiểu hóa giao dịch thanh toán
-- Sau khi chạy thuật toán Minimize Cash Flow, mỗi dòng là một giao dịch
-- thanh toán cần thực hiện giữa 2 người (from_user -> to_user).
-- status: PENDING (chưa trả), PAID (đã trả), CANCELLED (đã hủy).
-- payment_method: PAYOS_QR, MOMO, SEPAY_TRANSFER, CASH.
-- ============================================================================
CREATE TABLE IF NOT EXISTS settlements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),                  -- ID settlement
    group_id UUID REFERENCES groups(id),                            -- Thuộc nhóm nào
    from_user UUID REFERENCES users(id),                            -- Người phải trả tiền (debtor)
    to_user UUID REFERENCES users(id),                              -- Người được nhận tiền (creditor)
    amount NUMERIC(14,2) NOT NULL,                                  -- Số tiền cần thanh toán
    status VARCHAR(20) DEFAULT 'PENDING',                           -- PENDING / PAID / CANCELLED
    payment_method VARCHAR(20),                                     -- PAYOS_QR / MOMO / SEPAY_TRANSFER / CASH
    qr_code_data TEXT,                                              -- Dữ liệu QR code hoặc payment link
    confirmed_transaction_id UUID REFERENCES bank_transactions(id), -- Giao dịch ngân hàng xác nhận đã trả
    paid_at TIMESTAMPTZ,                                            -- Thời điểm thanh toán thành công
    created_at TIMESTAMPTZ DEFAULT now()                            -- Thời điểm tạo settlement
);

-- ============================================================================
-- Indexes: Tối ưu các truy vấn thường xuyên
-- ============================================================================
CREATE INDEX IF NOT EXISTS idx_expenses_group ON expenses(group_id);              -- Lấy danh sách expense theo nhóm
CREATE INDEX IF NOT EXISTS idx_settlements_group ON settlements(group_id);        -- Lấy danh sách settlement theo nhóm
CREATE INDEX IF NOT EXISTS idx_groups_share_code ON groups(share_code);           -- Tìm nhóm bằng mã mời
CREATE INDEX IF NOT EXISTS idx_bank_tx_user_unused ON bank_transactions(user_id, is_used); -- Lấy giao dịch chưa dùng của 1 user