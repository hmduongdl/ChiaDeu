-- Migration 003: Chuyển schema thử nghiệm (001) sang schema mục tiêu của Chế độ
-- chia đều linh hoạt theo schema.md.
--
-- Các thay đổi chính:
--   - Tiền chuyển từ NUMERIC thành BIGINT integer money (`_minor`).
--   - groups/group_members có status, role, thời gian cập nhật.
--   - expenses có created_by, expense_date, status, batch_id, updated_at.
--   - Thêm settlement_batches, settlements (mới), settlement_events, audit_logs.
--   - Thêm bảng sessions cho refresh token rotation/revocation.
--   - FK tài chính chuyển sang RESTRICT, không cascade lịch sử đã chốt.
--
-- Migration đang chạy trên dữ liệu prototype cũ; mọi chuyển đổi tiền đều round về
-- đơn vị đồng (VND), theo nguyên tắc lưu số nguyên đơn vị nhỏ nhất.

-- ==============================================================================
-- groups
-- ==============================================================================
ALTER TABLE groups ALTER COLUMN share_code TYPE VARCHAR(16);
ALTER TABLE groups ALTER COLUMN currency TYPE CHAR(3) USING (trim(currency)::CHAR(3));
ALTER TABLE groups ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'ACTIVE';
ALTER TABLE groups ADD COLUMN updated_at TIMESTAMPTZ;
ALTER TABLE groups ADD CONSTRAINT groups_status_check CHECK (status IN ('ACTIVE', 'ARCHIVED'));

-- ==============================================================================
-- group_members
-- ==============================================================================
ALTER TABLE group_members ADD COLUMN role TEXT NOT NULL DEFAULT 'MEMBER';
ALTER TABLE group_members ADD COLUMN status TEXT NOT NULL DEFAULT 'ACTIVE';
ALTER TABLE group_members ADD COLUMN left_at TIMESTAMPTZ;
ALTER TABLE group_members ADD CONSTRAINT group_members_role_check CHECK (role IN ('ADMIN', 'MEMBER'));
ALTER TABLE group_members ADD CONSTRAINT group_members_status_check CHECK (status IN ('ACTIVE', 'LEFT'));

-- ==============================================================================
-- expenses
-- ==============================================================================
ALTER TABLE expenses RENAME COLUMN amount TO amount_minor;
ALTER TABLE expenses ALTER COLUMN amount_minor TYPE BIGINT USING round(amount_minor)::BIGINT;
ALTER TABLE expenses ALTER COLUMN description SET DEFAULT '';
UPDATE expenses SET description = '' WHERE description IS NULL;
ALTER TABLE expenses ALTER COLUMN description SET NOT NULL;
ALTER TABLE expenses ADD COLUMN created_by UUID REFERENCES users (id);
ALTER TABLE expenses ADD COLUMN expense_date DATE;
ALTER TABLE expenses ADD COLUMN status TEXT NOT NULL DEFAULT 'ACTIVE';
ALTER TABLE expenses ADD COLUMN updated_at TIMESTAMPTZ;
ALTER TABLE expenses ADD CONSTRAINT expenses_amount_minor_positive CHECK (amount_minor > 0);
ALTER TABLE expenses ADD CONSTRAINT expenses_status_check CHECK (status IN ('ACTIVE', 'VOIDED'));
-- Không cascade xóa lịch sử tài chính.
ALTER TABLE expenses DROP CONSTRAINT IF EXISTS expenses_group_id_fkey;
ALTER TABLE expenses ADD CONSTRAINT fk_expenses_group FOREIGN KEY (group_id) REFERENCES groups (id) ON DELETE RESTRICT;
ALTER TABLE expenses ADD CONSTRAINT fk_expenses_created_by FOREIGN KEY (created_by) REFERENCES users (id) ON DELETE RESTRICT;

-- ==============================================================================
-- expense_splits
-- ==============================================================================
ALTER TABLE expense_splits RENAME COLUMN share_amount TO share_minor;
ALTER TABLE expense_splits ALTER COLUMN share_minor TYPE BIGINT USING round(share_minor)::BIGINT;
ALTER TABLE expense_splits ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT now();
ALTER TABLE expense_splits ADD CONSTRAINT expense_splits_share_minor_nonneg CHECK (share_minor >= 0);
ALTER TABLE expense_splits DROP CONSTRAINT IF EXISTS expense_splits_expense_id_fkey;
ALTER TABLE expense_splits ADD CONSTRAINT fk_splits_expense FOREIGN KEY (expense_id) REFERENCES expenses (id) ON DELETE RESTRICT;
ALTER TABLE expense_splits ADD CONSTRAINT fk_splits_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE RESTRICT;
CREATE UNIQUE INDEX uq_expense_splits_expense_user ON expense_splits (expense_id, user_id);

-- ==============================================================================
-- users — email unique không phân biệt hoa/thường (bảng đã tồn tại từ 001)
-- ==============================================================================
CREATE UNIQUE INDEX uq_users_email_lower ON users (lower(email)) WHERE email IS NOT NULL;

-- ==============================================================================
-- settlement_batches (mới)
-- ==============================================================================
CREATE TABLE settlement_batches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID NOT NULL REFERENCES groups (id) ON DELETE RESTRICT,
    created_by UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    idempotency_key VARCHAR(64) NOT NULL,
    status TEXT NOT NULL DEFAULT 'OPEN' CHECK (status IN ('OPEN', 'COMPLETED', 'CANCELLED')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    CONSTRAINT uq_batches_idempotency_per_group UNIQUE (group_id, idempotency_key)
);
-- Mỗi nhóm chỉ được có tối đa một kỳ đang mở.
CREATE UNIQUE INDEX idx_batches_one_open_per_group ON settlement_batches (group_id) WHERE status = 'OPEN';

-- Thêm liên kết batch vào expenses (cho cả bảng còn rỗng lẫn có dữ liệu).
ALTER TABLE expenses ADD COLUMN batch_id UUID REFERENCES settlement_batches (id) ON DELETE RESTRICT;
CREATE INDEX idx_expenses_group_batch_status ON expenses (group_id, batch_id, status);

-- ==============================================================================
-- settlements — thay thế bảng prototype cũ (từ_user/to_user, QR) bằng schema mục tiêu
-- ==============================================================================
DROP TABLE IF EXISTS settlements;

CREATE TABLE settlements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    batch_id UUID NOT NULL REFERENCES settlement_batches (id) ON DELETE RESTRICT,
    from_user_id UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    to_user_id UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    amount_minor BIGINT NOT NULL CHECK (amount_minor > 0),
    payment_code VARCHAR(20) NOT NULL,
    status TEXT NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'AWAITING_CONFIRMATION', 'PAID', 'CANCELLED')),
    marked_sent_at TIMESTAMPTZ,
    paid_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ,
    CONSTRAINT settlements_no_self_payment CHECK (from_user_id <> to_user_id)
);

-- payment_code unique không phân biệt hoa/thường.
CREATE UNIQUE INDEX uq_settlements_payment_code_lower ON settlements (lower(payment_code));
-- Không lặp cặp người trả/người nhận trong cùng một kỳ.
CREATE UNIQUE INDEX uq_settlements_batch_pair ON settlements (batch_id, from_user_id, to_user_id);
CREATE INDEX idx_settlements_batch ON settlements (batch_id);

-- ==============================================================================
-- settlement_events — lịch sử append-only của vòng đời settlement
-- ==============================================================================
CREATE TABLE settlement_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    settlement_id UUID NOT NULL REFERENCES settlements (id) ON DELETE RESTRICT,
    actor_id UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    event_type TEXT NOT NULL CHECK (event_type IN ('MARKED_SENT', 'CONFIRMED', 'REJECTED', 'CANCELLED')),
    note TEXT,
    receipt_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_settlement_events_settlement ON settlement_events (settlement_id);

-- ==============================================================================
-- audit_logs — nhật ký nghiệp vụ append-only
-- ==============================================================================
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID NOT NULL REFERENCES groups (id) ON DELETE RESTRICT,
    actor_id UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    action TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id TEXT,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_audit_logs_group ON audit_logs (group_id, created_at);

-- ==============================================================================
-- sessions — lưu refresh token để rotate và thu hồi phía server
-- ==============================================================================
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    refresh_token_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expired_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ
);
CREATE INDEX idx_sessions_user ON sessions (user_id);
CREATE INDEX idx_sessions_expired_at ON sessions (expired_at);