-- pgvector拡張の有効化
CREATE EXTENSION IF NOT EXISTS vector;

-- diary_embeddings テーブル
-- 日記エントリのベクトル埋め込みを格納（意味的検索用）
CREATE TABLE diary_embeddings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    diary_id UUID NOT NULL REFERENCES diaries(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    embedding vector(1536) NOT NULL, -- Gemini text-embedding-004 の次元数
    model_version TEXT NOT NULL DEFAULT 'text-embedding-004',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(diary_id)
);

-- コサイン類似度でのANN検索インデックス（HNSWはivfflatより行数制限がなく安定）
CREATE INDEX idx_diary_embeddings_embedding ON diary_embeddings
    USING hnsw (embedding vector_cosine_ops);

CREATE INDEX idx_diary_embeddings_user_id ON diary_embeddings(user_id);
