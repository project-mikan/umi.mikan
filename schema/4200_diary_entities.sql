CREATE TABLE IF NOT EXISTS diary_entities (
    id UUID PRIMARY KEY,
    diary_id UUID REFERENCES diaries(id) ON DELETE CASCADE NOT NULL,
    entity_id UUID REFERENCES entities(id) ON DELETE CASCADE NOT NULL,
    positions JSONB NOT NULL, -- [{"start": 0, "end": 5}, {"start": 10, "end": 15}]
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    CONSTRAINT unique_diary_entity UNIQUE (diary_id, entity_id)
);

CREATE INDEX index_diary_entities_diary_id ON diary_entities (diary_id);
CREATE INDEX index_diary_entities_entity_id ON diary_entities (entity_id);
