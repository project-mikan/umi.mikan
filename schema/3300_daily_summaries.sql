CREATE TABLE daily_summaries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    diary_id UUID NOT NULL REFERENCES diaries(id) ON DELETE CASCADE,
    summary TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Ensure one summary per diary
    UNIQUE(diary_id)
);

-- Index for efficient lookups
CREATE INDEX idx_daily_summaries_diary_id ON daily_summaries(diary_id);
CREATE INDEX idx_daily_summaries_created_at ON daily_summaries(created_at);