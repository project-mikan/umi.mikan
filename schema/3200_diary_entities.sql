CREATE TABLE IF NOT EXISTS diary_entities (
    id UUID PRIMARY KEY,
    diary_id UUID REFERENCES diaries(id) NOT NULL,                           
    entity_alias_id UUID REFERENCES entity_aliases(id) NOT NULL,                           
    start_position INT NOT NULL, -- エンティティの開始位置
    end_position INT NOT NULL, -- エンティティの終了位置
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL
);

