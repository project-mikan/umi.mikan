CREATE TABLE IF NOT EXISTS diary_trends (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id) NOT NULL,
    trend_type VARCHAR(20) NOT NULL, -- 'weekly', 'monthly', 'yearly'
    reference_date DATE NOT NULL,
    analysis TEXT NOT NULL,
    deprecated BOOLEAN DEFAULT FALSE,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    CONSTRAINT unique_user_trend UNIQUE (user_id, trend_type, reference_date),
    CONSTRAINT check_trend_type CHECK (trend_type IN ('weekly', 'monthly', 'yearly'))
);

CREATE INDEX index_diary_trends_user_id_type_date ON diary_trends (user_id, trend_type, reference_date);
CREATE INDEX index_diary_trends_deprecated ON diary_trends (deprecated);
