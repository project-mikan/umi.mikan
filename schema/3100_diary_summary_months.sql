CREATE TABLE IF NOT EXISTS diary_summary_months (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id) NOT NULL,
    year INTEGER NOT NULL,
    month INTEGER NOT NULL,
    summary TEXT NOT NULL,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    CONSTRAINT unique_user_month UNIQUE (user_id, year, month),
    CONSTRAINT check_month CHECK (month >= 1 AND month <= 12)
);

CREATE INDEX index_diary_summary_months_user_id_year_month ON diary_summary_months (user_id, year, month);
