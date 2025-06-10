CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    email VARCHAR(320) NOT NULL UNIQUE, -- ログイン用のやつ
    name VARCHAR(20) NOT NULL, -- 20文字以内(バイトでなく文字数)
    auth_type smallint NOT NULL, -- 0:password
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL
);
CREATE INDEX idx_users_email ON users (email);
