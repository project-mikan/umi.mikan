CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    name VARCHAR(20) NOT NULL, -- 20文字以内(バイトでなく文字数)
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL
);
