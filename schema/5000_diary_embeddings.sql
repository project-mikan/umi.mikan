-- pgvector拡張の有効化
CREATE EXTENSION IF NOT EXISTS vector;

-- diary_embeddings テーブル
-- 日記を話題ごとに分割したチャンク単位でベクトル埋め込みを格納（意味的検索用）
-- 1日記は複数チャンクに分割され、各チャンクが独立したembeddingを持つ
CREATE TABLE diary_embeddings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    diary_id UUID NOT NULL REFERENCES diaries(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    -- 日記内のチャンクインデックス（0始まり）
    chunk_index INT NOT NULL DEFAULT 0,
    -- チャンクのテキスト内容（検索結果のスニペット表示用）
    chunk_content TEXT NOT NULL DEFAULT '',
    -- ベクトル埋め込み（Gemini gemini-embedding-001 のネイティブ次元数; halfvec はHNSWで4000次元まで対応）
    embedding halfvec(3072) NOT NULL,
    model_version TEXT NOT NULL DEFAULT 'gemini-embedding-001',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(diary_id, chunk_index)
);

-- コサイン類似度でのANN検索インデックス（HNSWはivfflatより行数制限がなく安定）
-- m=16: グラフの接続数（デフォルト16）、ef_construction=64: 構築時の精度（デフォルト64）
CREATE INDEX idx_diary_embeddings_embedding ON diary_embeddings
    USING hnsw (embedding halfvec_cosine_ops)
    WITH (m = 16, ef_construction = 64);

CREATE INDEX idx_diary_embeddings_user_id ON diary_embeddings(user_id);
