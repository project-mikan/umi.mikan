CREATE TABLE IF NOT EXISTS entity_aliases (
    id UUID PRIMARY KEY,
    entity_id UUID REFERENCES entities(id) ON DELETE CASCADE NOT NULL,
    alias TEXT NOT NULL,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    -- 同じエンティティに対して同じエイリアスは登録できない
    UNIQUE(entity_id, alias)
);

CREATE INDEX index_entity_aliases_entity_id ON entity_aliases (entity_id);
CREATE INDEX index_entity_aliases_alias ON entity_aliases (alias);

-- エイリアスがユーザー内で一意であることを保証する制約
-- ユーザーは entities テーブル経由で取得される
-- Note: PostgreSQLでは複数テーブルにまたがるユニーク制約は直接作成できないため、
-- アプリケーションロジックとトリガーで対応する
