-- LLMは全ユーザー共通のため、プロバイダーは持たない
CREATE TABLE IF NOT EXISTS user_llms (
    user_id UUID REFERENCES users(id) PRIMARY KEY,
    auto_summary_monthly BOOLEAN NOT NULL DEFAULT FALSE, -- 月毎の自動要約生成を行うかどうか
    auto_latest_trend_enabled BOOLEAN NOT NULL DEFAULT FALSE, -- 直近トレンド分析の自動生成を行うかどうか
    semantic_search_enabled BOOLEAN NOT NULL DEFAULT FALSE, -- 意味的検索（RAG）機能を有効にするかどうか
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL
);
