````mermaid
  erDiagram

  users {
    uuid id PK
  }

  user_oauthes {
    uuid id PK "OAuth情報"
    string token
  }
  diaries {
    uuid id PK "日記"
    uuid user_id
    text content
  }

users ||--o{ entities : "1人のユーザーは0以上のエンティティを持つ"
users ||--|| user_oauthes : "oauth情報"
users ||--o{ diaries : "1人のユーザーは0以上の日記を持つ"

diaries ||--o{ diary_entities : "1つの日記は0以上の日記登場関連を持つ"
entities ||--o{ diary_entities : "1つのエンティティは0以上の日記登場関連を持つ"


entities ||--o{ entity_aliases : "1人のPersonは0以上の別名を持つ"

entities ||--|| entity_categories : "カテゴリー"

diary_entities{
    uuid id PK "日記登場人物"
    uuid diary_id "出てきた日記"
    uuid person_id "エンティティ"
}

 entities {
    uuid id PK "日記に出てくるエンティティ"
    string name "通常名"
    uuid entitiy_category_id
    uuid user_id
  }
  entity_categories{
    uuid id PK "エンティティのカテゴリ"
    string name "カテゴリー名"

  }

  entity_aliases{
    uuid id PK "出てくる人の別の呼び方"
    string name "呼び方"
    uuid person_id
  }```
````
