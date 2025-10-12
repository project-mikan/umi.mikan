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

### エンティティハイライト表示

- Textareaにおいて、diary_entitiesに登録された位置をハイライトできるようにする
- バックエンドからdiaryとdiary_entitiesが返されるため、フロントエンドで組み立てる
- データを保存するときは再度日記本文とentityを分離し、別でdiary, diary_entitiesとしてバックエンドに送ることで、diaryは純粋な文字列のままとする
- `<a>`タグを用いて該当箇所を囲い、青色のリンクとして表示
  - クラス: `text-blue-600 dark:text-blue-400 hover:underline`
  - 該当箇所をクリックするとentity/{uuid}ページへ遷移
- 文字入力時にentitiesのデータを照合し、候補を出す
- 候補が出ないときは新規登録できる導線を用意する

### エンティティハイライト適用条件

**表示タイミング**:
- 入力中（`isTyping`フラグが立っている間）はハイライトを非表示
- 入力停止後500ms経過してからハイライトを表示
- これにより、入力中のカーソル位置ズレやDOM操作による入力妨害を防止

**適用条件**:
- 現在のテキスト（`value`）が保存されたコンテンツ（`savedContent`）と一致する場合のみハイライトを適用
- テキストが編集されている場合（`value !== savedContent`）はプレーンテキストとして表示
- この仕組みにより、編集中のテキストに古いpositionデータが適用されるのを防ぐ

**データ検証**:
- バックエンドから取得した`diaryEntities`に無効なエンティティが含まれている可能性を考慮
- `validateDiaryEntities`関数で各positionのテキストが実際のエンティティ名/エイリアスと一致するかチェック
- 不一致のpositionは除外してからハイライトを適用

### エンティティ内部編集の検出と紐づけ解除

**課題**: ハイライト済みのエンティティ（`<a>`タグ）内のテキストをユーザーが編集した場合、エンティティの意味が変わるため紐づけを解除する必要がある

**実装**:
1. `input`イベント時にDOM内の全`<a>`タグをチェック
2. 各リンクのテキストが元のエンティティ名/エイリアスと完全一致するか検証
3. 不一致の場合、そのリンクをDOMから削除（テキストノードに置換）
4. 同時に`selectedEntities`からも該当のpositionを削除

**技術詳細**:
- `getTextOffset`関数でリンクのテキスト位置を計算
- `removedPositions`マップで削除されたentityIdとpositionを記録
- `selectedEntities`をフィルタリングして無効なpositionを除外
- Svelteのreactivityのために新しい配列を作成して代入

### 追加ページ

#### URL:/entity/{uuid}

- ユーザが登録している個別のentityの管理ページ
  - entityに紐づくaliasの一覧表示・編集・削除
- entityの削除・aliasの追加・名前・メモの変更が可能
- entityが紐づく日記を表示

#### URL:/entities

- ユーザが登録しているentitiesを一覧で見れる
- 削除・追加も可能

### テキスト入力基盤

**contenteditable div の使用**:
- `<textarea>`の代わりに`contenteditable`属性を持つ`<div>`要素を使用
- これによりエンティティハイライト（`<a>`タグ）とプレーンテキストの混在が可能
- `htmlToPlainText`関数でHTML→プレーンテキスト変換
  - `<br>`タグを改行文字（`\n`）に変換
  - Google Keep等からの貼り付けに対応（`<p>`, `<div>`, `<li>`タグ処理）
  - HTMLエスケープとXSS対策

**改行処理**:
- Enterキー押下時に`<br>`タグを手動挿入
  - デフォルトのcontenteditable動作（`<div>`挿入など）を防止
  - IME入力中（`isComposing`）はEnterキーを無視
- 末尾での改行時は2つの`<br>`を挿入してカーソルが次行に表示されるようにする
- `hasContentAfterNode`関数で末尾判定

**カーソル位置管理**:
- `getTextOffset`: DOM位置→プレーンテキスト位置の変換
  - テキストノードとBR要素を再帰的に走査
  - BRタグは改行1文字としてカウント
- `createRangeAtTextOffset`: プレーンテキスト位置→DOM位置の変換
- `restoreCursorPosition`: テキストオフセットからカーソル位置を復元

### エンティティ候補システム仕様

#### 候補表示のトリガー条件

- ユーザが日記本文を入力中に、単語の前方一致でentity候補を表示
  - カーソル位置から後方に向かって、2文字以上の全ての部分文字列をエンティティリストと照合
  - 最長一致を優先して候補を表示
- IME入力中（`compositionupdate`イベント）は候補検索をスキップ
  - IMEの動作を安定させるため、DOM操作を最小限に抑える
  - 確定後（`compositionend`）に候補検索を実行

#### エンティティデータの取得と検索

**全エンティティデータの事前キャッシュ**:
- `loadAllEntities`関数でコンポーネントマウント時に全エンティティを取得
- APIエンドポイント: `/api/entities/search?q=`
- `allEntities`配列に全エンティティを保存
- `allFlatEntities`配列にエンティティ名とエイリアスをフラット化して保存
  - 各項目: `{ entity: Entity, text: string, isAlias: boolean }`

**ブラウザ側での候補フィルタリング**:
- `searchForSuggestions`関数でクライアント側フィルタリング
- 入力クエリとの前方一致でエンティティを特定
- マッチしたエンティティIDの全バリエーション（名前+全エイリアス）を候補に含める
- 完全一致の場合も候補を表示（ユーザーの明示的な選択を待つ）

#### カーソル位置の制御

- エンティティ確定時、カーソルは挿入されたテキストの末尾に配置
- `createRangeAtTextOffset`関数でBR要素を1文字としてカウント
- contenteditable div内のテキストノードとBR要素を再帰的に走査
- `htmlToPlainText`との整合性を保つためBRタグ=改行1文字として扱う
- エンティティ選択後、`generateHTMLFromSelectedEntities`でハイライト付きHTMLを生成してDOMを更新
- `findNodeAtPosition`関数でプレーンテキスト位置からDOM位置を計算してカーソルを配置

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

#### エンティティ選択処理

**選択トリガー**:
- Enterキー押下
- 候補をクリック
- Tabキー/ArrowUp/ArrowDownで候補を選択後、Enterで確定

**`selectSuggestion`関数の処理フロー**:
1. `isSelectingEntity`フラグを立てて、reactive statementからの上書きを防ぐ
2. 現在のcontentElementからvalueを更新（最新のHTML→プレーンテキスト変換）
3. カーソル位置を取得
4. 選択されたテキスト（エンティティ名/エイリアス）でcurrentTriggerPosを再計算
   - カーソル直前から後方に検索し、最長一致を見つける
5. currentTriggerPosからcursorPosまでの文字列を選択されたテキストに置き換え
6. 新しいpositionを計算（`start`, `end`）
7. `selectedEntities`配列に追加（既存entityの場合はpositionsに追加）
8. エンティティ確定直後フラグ（`justSelectedEntity`）を立てる
9. `isTyping`フラグを下ろす
10. `tick()`でreactive statement実行を待つ
11. `isSelectingEntity`フラグを下ろす
12. `generateHTMLFromSelectedEntities`でハイライト付きHTMLを生成
13. カーソルを挿入位置の末尾に配置

**キーボード操作**:
- Tab/Shift+Tab: 候補間を循環
- ArrowDown/ArrowUp: 候補を上下に移動
- Enter: 選択中の候補を確定（選択なしの場合は最長一致を確定）
- Escape: 候補を閉じる

#### 明示的選択のみ登録（2025-10-11追加）

**問題**: 従来は`extractEntitiesFromContent`関数がテキスト内の全ての一致文字列を自動的にエンティティとして登録していた。これにより、ユーザーが候補を選択していない場合でも完全一致する文字列があれば自動登録されてしまう問題があった。

**解決策**: エンティティは明示的に選択した場合のみ登録されるように変更

**実装**:

1. **フロントエンド**:
   - Textareaコンポーネントで`selectedEntities`配列を管理
   - エンティティ候補を明示的に選択（Enter/Click）した時のみ`selectedEntities`に追加
   - エンティティ選択直後を除き、テキスト編集でエンティティ内部が変更された場合は紐づけを解除
   - フォーム送信時に`selectedEntities`をJSON形式でhidden inputとして送信

2. **バックエンド**:
   - `extractEntitiesFromContent`関数の使用を廃止
   - フロントエンドから送信された`selectedEntities`を直接使用
   - `DiaryEntityInput`形式に変換して保存

**対象ファイル**:

- `frontend/src/lib/components/atoms/Textarea.svelte`: selectedEntities管理
- `frontend/src/lib/components/molecules/FormField.svelte`: selectedEntities伝播
- `frontend/src/routes/[id]/+page.svelte`: selectedEntitiesバインディング
- `frontend/src/routes/[id]/+page.server.ts`: selectedEntitiesからDiaryEntityInput変換
- `frontend/src/routes/+page.svelte`: トップページの3フォーム対応
- `frontend/src/routes/+page.server.ts`: トップページの3アクション対応

#### diaryEntitiesとselectedEntitiesの同期（2025-10-12追加）

**課題**: バックエンドから返される`diaryEntities`とフロントエンド側の`selectedEntities`の整合性を保つ

**同期ロジック（reactive statement）**:
```typescript
$: {
  if (diaryEntities && diaryEntities.length > 0 && !isTyping && !isSelectingEntity && allEntities && allEntities.length > 0) {
    // diaryEntitiesを検証
    const validatedDiaryEntities = validateDiaryEntities(value, diaryEntities, allEntities);

    // validatedDiaryEntitiesからselectedEntitiesを生成
    const entitiesFromDiary = validatedDiaryEntities.map(...).filter(...);

    // selectedEntitiesが空の場合は無条件で更新
    if (selectedEntities.length === 0) {
      selectedEntities = entitiesFromDiary;
    } else {
      // 既にデータがある場合、positionの数で比較
      // diaryEntitiesの方が多い場合のみ更新（新しいデータ）
      const selectedTotalPositions = selectedEntities.reduce(...);
      const diaryTotalPositions = entitiesFromDiary.reduce(...);

      if (diaryTotalPositions > selectedTotalPositions) {
        selectedEntities = entitiesFromDiary;
      }
    }
  }
}
```

**更新条件**:
- `diaryEntities`が存在し空でない
- 入力中でない（`!isTyping`）
- エンティティ選択中でない（`!isSelectingEntity`）
- 全エンティティデータが読み込まれている（`allEntities.length > 0`）

**更新戦略**:
- `selectedEntities`が空の場合: 無条件で`diaryEntities`から生成
- `selectedEntities`に既にデータがある場合: positionの総数で比較
  - `diaryEntities`の方が多い → バックエンドから新しいデータが返された → 更新
  - `selectedEntities`の方が多い → ユーザーが新しく追加した → 上書きしない

**savedContentの更新**:
- `diaryEntities`が外部から変更されたら、`savedContent`を現在の`value`で更新
- これにより、次回のハイライト適用判定で正しく動作

#### 実装ファイル

- `frontend/src/lib/components/atoms/Textarea.svelte`
  - selectedEntities管理とエンティティ選択時の追加
  - 後方substring一致検索
  - 確定後エンティティ除外ロジック
  - 空入力チェック
  - エンティティ内部編集の検出と紐づけ解除
  - diaryEntitiesとselectedEntitiesの同期（reactive statement）
  - カーソル位置計算関数(`getTextOffset`)
  - カーソル位置設定関数(`createRangeAtTextOffset`)
  - エンティティハイライト適用条件の制御（`isTyping`, `savedContent`）
  - HTML生成関数（`generateHTMLFromSelectedEntities`, `highlightedHTML`）
- `frontend/src/lib/utils/diary-entity-highlighter.ts`
  - `validateDiaryEntities`: diaryEntitiesの検証関数
  - `highlightEntities`: エンティティハイライトHTML生成
  - `escapeHtml`: HTMLエスケープ処理
- `frontend/src/routes/[id]/+page.svelte`
  - `selectedEntities`のバインディング
  - FormFieldへのdiaryEntitiesとselectedEntitiesの受け渡し
  - hidden inputでselectedEntitiesをJSON形式で送信
