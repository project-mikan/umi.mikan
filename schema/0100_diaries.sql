CREATE TABLE IF NOT EXISTS diaries (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id) NOT NULL,                           
    content TEXT NOT NULL,
    date DATE NOT NULL,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    UNIQUE (user_id, date) -- ユーザごとに日付をユニークにする
);

