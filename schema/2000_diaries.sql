CREATE TABLE IF NOT EXISTS diaries (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,                           
    content TEXT NOT NULL,
    date DATE NOT NULL,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL
);

