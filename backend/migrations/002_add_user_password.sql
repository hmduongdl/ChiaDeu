-- Migration 002: Thêm cột password_hash vào bảng users.
-- Người dùng đã tồn tại trước migration này sẽ có password_hash = NULL,
-- cần thực hiện quy trình đặt lại mật khẩu trước khi có thể đăng nhập bằng password.
-- Người dùng mới đăng ký sau migration này luôn có password_hash được ghi.
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS password_hash TEXT;

-- Existing rows must complete a password reset before password login can be used.
-- The application always writes this field for newly registered users.
