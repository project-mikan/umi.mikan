-- MCPサーバーなど外部クライアント向けのAPIキー
-- キー本体は保存せずSHA-256ハッシュのみを保持する（発行時に一度だけ平文を返す）
CREATE TABLE IF NOT EXISTS user_api_keys (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL, -- キーの用途を示すラベル
    key_hash VARCHAR(64) NOT NULL UNIQUE, -- APIキーのSHA-256ハッシュ（hex）
    key_prefix VARCHAR(16) NOT NULL, -- 一覧表示用のキー先頭部分（例: umi_a1b2c3d4）
    last_used_at BIGINT, -- 最終使用日時（Unix秒、未使用の場合はNULL）
    expires_at BIGINT NOT NULL, -- 有効期限（Unix秒）。長期間有効な認証情報が漏洩した際の被害を限定するため必須とする
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_user_api_keys_user_id ON user_api_keys(user_id);
