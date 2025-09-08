CREATE TABLE IF NOT EXISTS user_llms (
    user_id UUID REFERENCES users(id) PRIMARY KEY,                           
    llm_provider  smallint NOT NULL, -- 0:Gemini
    token  VARCHAR(100) NOT NULL,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    CONSTRAINT unique_user_llm UNIQUE (user_id, llm_provider) -- ユーザごとにLLM Providerは一意
);
