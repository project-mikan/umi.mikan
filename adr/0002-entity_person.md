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

### エンティティ候補システム仕様

#### 候補表示のトリガー条件

- ユーザが日記本文を入力中に、単語の前方一致でentity候補を表示
  - カーソル位置から後方に向かって、2文字以上の全ての部分文字列をエンティティリストと照合
  - 最長一致を優先して候補を表示

#### カーソル位置の制御

- エンティティ確定時、カーソルは挿入されたテキストの末尾に配置
- `createRangeAtTextOffset`関数でBR要素を1文字としてカウント
- contenteditable div内のテキストノードとBR要素を再帰的に走査
- `htmlToPlainText`との整合性を保つためBRタグ=改行1文字として扱う

#### 確定後の再候補表示

**要件**: エンティティ確定後も新しい入力で候補を表示し続ける

**実装方法**:

- `justSelectedEntity`フラグと`lastSelectedEntityText`変数で状態管理
- エンティティ確定時:
  1. `justSelectedEntity = true`
  2. `lastSelectedEntityText`に確定したエンティティ名を保存
- 次の入力イベント時:
  1. 確定したエンティティ名を現在の単語から除外
  2. 残った部分で候補検索を実行
  3. フラグとテキストをリセット

**例**:

1. "na"と入力 → "名取さな"の候補表示
2. "natori"を確定 → テキストは"natori"
3. "na"と続けて入力 → テキストは"natorina"
4. システムが"natori"部分を除外 → "na"で検索
5. 再度"名取さな"の候補表示

**制約**: スペースは追加しない（ユーザ要件）

#### 空入力時の候補非表示

**要件**: 改行直後や何も入力していない状態では候補を表示しない

**実装**:

- `textAfterLastNewline`（最後の改行以降のテキスト）を`searchText`として使用
- `searchText.length === 0`の場合、候補を非表示にする
- これにより、改行直後（カーソルの前がすべて改行）や入力開始前に候補が表示されなくなる

#### 実装ファイル

- `frontend/src/lib/components/atoms/Textarea.svelte`
  - 後方substring一致検索: lines 318-362
  - 確定後エンティティ除外ロジック: lines 301-316
  - 空入力チェック: lines 318-323
  - エンティティ確定時フラグ設定: lines 364-366
  - カーソル位置計算関数(`getTextOffset`): lines 378-444
