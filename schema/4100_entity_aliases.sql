CREATE TABLE IF NOT EXISTS entity_aliases (
    id UUID PRIMARY KEY,
    entity_id UUID REFERENCES entities(id) ON DELETE CASCADE NOT NULL,
    alias TEXT NOT NULL,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL
);

CREATE INDEX index_entity_aliases_entity_id ON entity_aliases (entity_id);
CREATE INDEX index_entity_aliases_alias ON entity_aliases (alias);
