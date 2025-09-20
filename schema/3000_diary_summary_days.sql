CREATE TABLE IF NOT EXISTS diary_summary_days (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id) NOT NULL,
    date DATE NOT NULL,
    summary TEXT NOT NULL,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    CONSTRAINT unique_user_date_summary UNIQUE (user_id, date)
);

CREATE INDEX index_diary_summary_days_user_id_date ON diary_summary_days (user_id, date);
