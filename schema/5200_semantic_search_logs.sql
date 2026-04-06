-- semantic_search_logs テーブル
-- 意味的検索（RAG）のAIリクエスト履歴を格納（メトリクス集計用）
CREATE TABLE semantic_search_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_semantic_search_logs_user_id ON semantic_search_logs(user_id);
CREATE INDEX idx_semantic_search_logs_created_at ON semantic_search_logs(created_at);
