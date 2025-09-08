CREATE TABLE IF NOT EXISTS diary_summary_days (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id) NOT NULL,
    diary_id UUID REFERENCES diaries(id) NOT NULL,
    summary TEXT NOT NULL,
    date DATE NOT NULL,
    deprecated BOOLEAN DEFAULT FALSE,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    CONSTRAINT unique_user_date_summary UNIQUE (user_id, date)
);

CREATE INDEX index_diary_summary_days_user_id_date ON diary_summary_days (user_id, date);
CREATE INDEX index_diary_summary_days_deprecated ON diary_summary_days (deprecated);
