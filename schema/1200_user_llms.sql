CREATE TABLE IF NOT EXISTS user_llms (
    user_id UUID REFERENCES users(id) PRIMARY KEY,
    llm_provider  smallint NOT NULL, -- 1:Gemini
    key  VARCHAR(100) NOT NULL,
    auto_summary_daily BOOLEAN NOT NULL DEFAULT FALSE, -- 日毎の自動要約生成を行うかどうか
    auto_summary_monthly BOOLEAN NOT NULL DEFAULT FALSE, -- 月毎の自動要約生成を行うかどうか
    auto_latest_trend_enabled BOOLEAN NOT NULL DEFAULT FALSE, -- 直近トレンド分析の自動生成を行うかどうか
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    CONSTRAINT unique_user_llm UNIQUE (user_id, llm_provider) -- ユーザごとにLLM Providerは一意
);
