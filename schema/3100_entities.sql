CREATE TABLE IF NOT EXISTS entities (
    id UUID PRIMARY KEY,
    name VARCHAR(256) NOT NULL,
    memo TEXT,
    user_id UUID REFERENCES users(id) NOT NULL,                           
    category_type int NOT NULL,                           
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    CONSTRAINT unique_entity_name UNIQUE (user_id, name) -- ユーザーごとにnameはunique
);
