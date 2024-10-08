CREATE TABLE IF NOT EXISTS entities (
    id UUID PRIMARY KEY,
    name VARCHAR(256) NOT NULL,
    user_id UUID REFERENCES users(id) NOT NULL,                           
    category_id UUID REFERENCES entity_categories(id) NOT NULL,                           
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL
);
