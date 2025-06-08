CREATE TABLE IF NOT EXISTS user_password_authes (
    user_id UUID REFERENCES users(id) PRIMARY KEY,                           
    password_hashed VARCHAR(20) NOT NULL, -- ハッシュ化されたパスワード
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL
);
