ALTER TABLE users
    ADD COLUMN IF NOT EXISTS password_hash TEXT;

-- Existing rows must complete a password reset before password login can be used.
-- The application always writes this field for newly registered users.
