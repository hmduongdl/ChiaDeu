-- Migration 001: Khởi tạo schema cơ sở dữ liệu Chia Đều.
-- Tạo các bảng chính:
--   users           - Người dùng (id, tên, email, phone, tài khoản ngân hàng, avatar)
--   groups          - Nhóm chia tiền (id, tên, mã share_code, người tạo, tiền tệ)
--   group_members   - Thành viên trong nhóm (quan hệ nhiều-nhiều giữa groups và users)
--   bank_transactions - Giao dịch ngân hàng từ webhook (Sepay/PayOS/Momo)
--   expenses        - Chi phí trong nhóm (người trả, số tiền, kiểu chia)
--   expense_splits  - Phần chia của từng thành viên cho mỗi chi phí
--   settlements     - Thanh toán bù trừ giữa các thành viên (trạng thái, QR code)
-- Kèm các index cho hiệu năng truy vấn thường dùng.
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

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