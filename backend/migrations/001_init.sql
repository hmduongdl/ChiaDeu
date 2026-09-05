-- Migration 001: Khởi tạo schema cơ sở dữ liệu Chia Đều.
-- Tạo các bảng chính:
--   users           - Người dùng (id, tên, email, phone, avatar)
--   groups          - Nhóm chia tiền (id, tên, mã share_code, người tạo, tiền tệ)
--   group_members   - Thành viên trong nhóm (quan hệ nhiều-nhiều giữa groups và users)
--   expenses        - Chi phí trong nhóm (người trả, số tiền, kiểu chia)
--   expense_splits  - Phần chia của từng thành viên cho mỗi chi phí
--   settlements     - Thanh toán bù trừ giữa các thành viên (trạng thái, mã thanh toán)
-- Kèm các index cho hiệu năng truy vấn thường dùng.
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    phone VARCHAR(20) UNIQUE,
    email VARCHAR(100) UNIQUE,
    bank_account_no VARCHAR(50),
    bank_code VARCHAR(20),
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

CREATE TABLE expenses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID REFERENCES groups(id) ON DELETE CASCADE,
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
    paid_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_expenses_group ON expenses(group_id);
CREATE INDEX idx_settlements_group ON settlements(group_id);
CREATE INDEX idx_groups_share_code ON groups(share_code);