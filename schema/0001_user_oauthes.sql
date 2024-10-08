CREATE TABLE IF NOT EXISTS user_oauthes (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id) NOT NULL,
    provider_name VARCHAR(20) NOT NULL, -- provider名
    provider_user_id VARCHAR(255) NOT NULL,
    access_token TEXT NOT NULL,
    refresh_token TEXT NOT NULL,
    token_expires_at BIGINT NOT NULL,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    UNIQUE (provider_name, provider_user_id)    -- プロバイダーごとに一意のユーザー
);


