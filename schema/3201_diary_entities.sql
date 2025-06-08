CREATE TABLE IF NOT EXISTS diary_entities (
    id UUID PRIMARY KEY,
    diary_id UUID REFERENCES diaries(id) NOT NULL,                           
    entity_id UUID REFERENCES entities(id) NOT NULL,                           
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL
);

