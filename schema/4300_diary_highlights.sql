-- diary_highlights テーブル
-- 日記エントリのLLM生成ハイライト情報を格納
CREATE TABLE diary_highlights (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    diary_id UUID NOT NULL REFERENCES diaries(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    highlights JSONB NOT NULL, -- ハイライト情報の配列 [{"start": 0, "end": 25, "text": "..."}]
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(diary_id)
);

-- diary_id で検索するためのインデックス
CREATE INDEX idx_diary_highlights_diary_id ON diary_highlights(diary_id);

-- user_id で検索するためのインデックス
CREATE INDEX idx_diary_highlights_user_id ON diary_highlights(user_id);
