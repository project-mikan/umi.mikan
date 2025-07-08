CREATE TABLE IF NOT EXISTS entity_aliases (
    id UUID PRIMARY KEY,
    entity_id UUID REFERENCES entities(id) NOT NULL,                           
    name VARCHAR(256) NOT NULL, -- エイリアス
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL
    -- エイリアスではuniqueを考慮しない(例えば同じ名前の人間を別の人物として記録しておきたいこともあるため)
);

