CREATE TABLE IF NOT EXISTS diary_months (
    id UUID PRIMARY KEY,
    diary_id UUID REFERENCES users(id) NOT NULL,                           
    content TEXT NOT NULL,
    date DATE NOT NULL,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    CONSTRAINT unique_user_date UNIQUE (user_id, date) -- ユーザーごとに日付は一意
);

CREATE INDEX index_diaries_user_id_and_date ON diaries (user_id, date);

