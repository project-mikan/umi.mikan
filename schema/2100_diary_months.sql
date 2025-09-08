CREATE TABLE IF NOT EXISTS diary_months (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id) NOT NULL,                           
    content TEXT NOT NULL,
    date DATE NOT NULL,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    CONSTRAINT unique_user_date_months UNIQUE (user_id, date) -- ユーザーごとに日付は一意
);

CREATE INDEX index_diary_months_user_id_and_date ON diary_months (user_id, date);

