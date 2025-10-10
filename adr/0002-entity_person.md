# entityにおける人物機能

## 概要

- 日記において固有名詞の存在は極めて重要
- 例えば人物を明示的に指定することで、人物のハイライトや検索を容易とする
- 固有名詞をentityとして管理し、まずその1つとして「人物」を登録できるようにする

## バックエンド

### DB

- entitiesテーブルを用いて固有名詞を管理
  - entityが抽象で人物はその一部
- entity_aliasesテーブルではentityの別名を登録
  - usuyuki, うすゆきなど表記ゆれに対応
- diary_entitiesでdiariesの日記のどこに登場するかを紐づける
  - diariesの更新時にdiary_entitiesも同時に更新する

#### diaryとentityの紐づけ検討

登場位置ごとにレコードを持つ

##### 本文中での位置をレコードごとに分けて持つ

- entityページで該当diaryの表示が容易
- レコード数が極めて多くなる
- 1日記の更新で複数レコードの更新が必要

##### 1つのdiaryに対して1つレコードを作る

jsonカラムを作り位置とentityの対応付けを記録

- diary_entitiesを作る
- diary,entityの複合PK
  - entityとdiaryを1対1にすることでentity→diaryの探索コストを下げる
- positionsにJSONで該当のentityのstart,endを持つ
  - このときaliasを使用していてもaliasのuuidなどは保持しない
  - 理由：名前が変わったとき、diaryテーブルの名前は変えられないため、あくまでentityの位置を保持することで、名前が変わっても位置を保持し続けられる
  - aliasはユーザの参照用でDB側では保存塩内
- 1つのdiaryに複数のentityがあるときは、entityごとにdiary_entitiesのレコードが作られる

#### category_id

- 0:未分類(no_category)
- 1:人物(people)

## フロントエンド

- Textareaにおいて、diary_entitiesに登録された位置をハイライトできるようにする
- バックエンドからdiaryとdiary_entitiesが返されるため、フロントエンドで組み立てる
- データを保存するときは再度日記本文とentityを分離し、別でdiary, diary_entitiesとしてバックエンドに送ることで、diaryは純粋な文字列のままとする
- <span>を用いて該当箇所を囲い、色を付ける
  - 該当箇所をクリックするとentity/{uuiod}ページへ遷移
- 文字入力時にentitiesのデータを照合し、候補を出す
- 候補が出ないときは新規登録できる導線を用意する

### 追加ページ

#### URL:/entity/{uuid}

- ユーザが登録している個別のentityの管理ページ
  - entityに紐づくaliasの一覧表示・編集・削除
- entityの削除・aliasの追加・名前・メモの変更が可能
- entityが紐づく日記を表示

#### URL:/entities

- ユーザが登録しているentitiesを一覧で見れる
- 削除・追加も可能
