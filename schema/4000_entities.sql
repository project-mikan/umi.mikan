CREATE TABLE IF NOT EXISTS entities (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id) NOT NULL,
    name TEXT NOT NULL,
    category_id INTEGER NOT NULL DEFAULT 0, -- 0:未分類(no_category), 1:人物(people)
    memo TEXT,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    UNIQUE(user_id, name) -- 同一ユーザー内でエンティティ名の重複を禁止
);

CREATE INDEX index_entities_user_id ON entities (user_id);
CREATE INDEX index_entities_category_id ON entities (category_id);
