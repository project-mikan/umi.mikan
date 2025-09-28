# ADR-0007: 依存性注入（DI）コンテナの実装

## 背景

この実装以前、バックエンドでは3つの主要アプリケーション（server、scheduler、subscriber）間で依存関係管理に重大な問題がありました：

### 特定された問題

1. **コードの重複**: 各アプリケーションが同一の依存関係（DB接続、Redisクライアント、設定読み込み）を手動で初期化
2. **密結合**: サービスがハードコーディングされた依存関係で直接インスタンス化
3. **テスタビリティの低さ**: ユニットテストでの依存関係のモック化が困難
4. **リソース管理**: アプリケーション間でのリソースクリーンアップが非一貫
5. **保守負荷**: 新しい依存関係の追加で複数箇所の変更が必要
6. **設定の散在**: 環境変数の読み込みと検証がファイル間に分散

### コードベース分析

- **server/main.go**: 手動依存関係セットアップで123行
- **scheduler/main.go**: 重複した初期化ロジックで526行以上
- **subscriber/main.go**: 手動サービス作成で同様のパターン

3つのアプリケーションすべてが、本質的に同じデータベース接続、Redisセットアップ、サービスインスタンス化コードを繰り返していました。

## 決定

すべてのアプリケーション依存関係を管理するため、**uber-go/dig**を使用した中央集権的な依存性注入（DI）コンテナを実装する。

### 技術選択: uber-go/dig

**digを他の選択肢より選んだ理由:**

1. **成熟かつ実戦経験済み**: 本番環境のGoアプリケーションで広く使用
2. **リフレクションベース**: 手動配線なしでの自動依存関係解決
3. **エラー検出**: 依存関係グラフのコンパイル時・実行時検証
4. **最小限のボイラープレート**: プロバイダー登録と消費がクリーン
5. **Uberの実績**: マイクロサービスアーキテクチャ向けにUberが開発・保守

**検討した代替案:**

- 手動DI（現在のアプローチ） - 保守負荷により却下
- Google Wire - コード生成の複雑さにより却下
- fxフレームワーク - 現在のニーズには過剰として却下

## 実装アーキテクチャ

### 1. コンテナ構造

```go
// digをラップする中央コンテナ
type Container struct {
    container *dig.Container
}

// アプリケーション固有の依存関係グループ
type ServerApp struct {
    DB           database.DB
    Redis        rueidis.Client
    AuthService  *auth.AuthEntry
    DiaryService *diary.DiaryEntry
    UserService  *user.UserEntry
}

type SchedulerApp struct {
    DB              database.DB
    Redis           rueidis.Client
    SchedulerConfig *SchedulerConfig
}

type SubscriberApp struct {
    DB               database.DB
    Redis            rueidis.Client
    LLMFactory       LLMClientFactory
    LockService      LockService
    SubscriberConfig *SubscriberConfig
}
```

### 2. プロバイダーカテゴリ

#### 設定プロバイダー

- 環境変数の読み込みと検証
- 型安全な設定構造体
- 欠落した環境変数に対する中央エラーハンドリング

#### インフラストラクチャプロバイダー

- 適切な設定でのデータベース接続
- 接続検証付きRedisクライアント
- AIサービス抽象化のためのLLMクライアントファクトリー
- 協調のための分散ロックサービス

#### サービスプロバイダー

- 依存関係を注入されたビジネスロジックサービス
- gRPCサービス実装
- 認証・認可サービス

#### アプリケーションプロバイダー

- アプリケーション固有の依存関係バンドル
- リソースクリーンアップ協調
- ライフサイクル管理

### 3. エラーハンドリング思想

```go
// 重要な依存関係に対するfail-fastアプローチ
func mustProvide(container *dig.Container, constructor interface{}) {
    if err := container.Provide(constructor); err != nil {
        panic(fmt.Sprintf("failed to provide dependency: %v", err))
    }
}
```

**思想**: 依存関係が満たされない場合、実行時に失敗するよりもアプリケーション起動時に即座に失敗すべき。

## 実現された利益

### 1. コード削減と明確性

- **server/main.go**: 手動セットアップからクリーンなDI呼び出しに短縮
- **scheduler/main.go**: 526行以上からより良い構造の356行に簡素化
- **subscriber/main.go**: 関心の分離がよりクリーンに

### 2. 保守性の向上

- **単一情報源**: すべての依存関係作成ロジックが中央集権化
- **簡単な拡張**: 新しいサービスの追加にはプロバイダー登録のみが必要
- **一貫したパターン**: すべてのアプリケーションが同じ依存性注入パターンを採用

### 3. テスタビリティの向上

- **インターフェースベース設計**: すべての主要依存関係がインターフェースを実装
- **モックフレンドリー**: テスト実装の代替が簡単
- **分離テスト**: サービスを最小限の依存関係でテスト可能

### 4. より良いリソース管理

- **中央集権化されたクリーンアップ**: `Cleanup`型がすべてのリソース廃棄を管理
- **グレースフルシャットダウン**: アプリケーション終了時の適切なリソースクリーンアップ
- **コネクションプーリング**: データベースとRedis接続の共有

### 5. 設定管理

- **検証**: 必要な環境変数の早期検証
- **型安全性**: 適切な型付けされた設定構造体
- **エラー伝播**: 設定問題の明確なエラーメッセージ

## 実装詳細

### アプリケーション起動パターン

```go
// すべてのアプリケーションで一貫したパターン
func main() {
    diContainer := container.NewContainer()
    if err := diContainer.Invoke(runApplication); err != nil {
        logger.WithError(err).Fatal("Failed to start application")
    }
}

func runApplication(app *container.AppType, cleanup *container.Cleanup) error {
    // 注入された依存関係を使用したアプリケーション固有ロジック
    defer cleanup.Close()
    // ...
}
```

### インターフェース分離

```go
// AIサービスのためのLLM抽象化
type LLMClientFactory interface {
    CreateGeminiClient(ctx context.Context, apiKey string) (*llm.GeminiClient, error)
}

// 分散協調のためのロックサービス抽象化
type LockService interface {
    NewDistributedLock(key string, duration time.Duration) lock.DistributedLockInterface
}
```

## 移行戦略

### フェーズ1: インフラストラクチャ（完了）

- DIコンテナ構造の作成
- コアプロバイダー（DB、Redis、Config）の実装
- サーバーアプリケーションの更新

### フェーズ2: 非同期サービス（完了）

- スケジューラーアプリケーションの移行
- サブスクライバーアプリケーションの移行
- ジョブ抽象化の実装

### フェーズ3: 検証（完了）

- lintエラーとコンパイル問題の修正
- 適切なエラーハンドリングの追加
- テスト互換性の確保

## 結果

### ポジティブ

1. **コード重複の削減**: 約200行以上の繰り返し初期化コードを排除
2. **開発者体験の向上**: 新しいサービスの追加と設定が容易
3. **より良いテスト**: モック注入が直接的で信頼性が高い
4. **一貫したアーキテクチャ**: すべてのアプリケーションが同じパターンを採用
5. **リソース効率**: 共有接続と適切なクリーンアップ
6. **早期エラー検出**: 依存関係の問題が実行時ではなく起動時に表面化

### ネガティブ

1. **学習コスト**: チームメンバーがDIパターンを理解する必要
2. **デバッグの複雑性**: スタックトレースにリフレクションベースの呼び出しが含まれる可能性
3. **実行時依存**: uber-go/digライブラリへの追加依存
4. **起動パフォーマンス**: リフレクションベース注入による最小限のオーバーヘッド

### 中立

1. **コードサイズ**: 重複排除により全体的なコードベースサイズが若干削減
2. **コンパイル時間**: ビルドパフォーマンスへの大きな影響なし

## 将来の検討事項

### 潜在的な拡張

1. **ヘルスチェック**: ヘルスチェックエンドポイントとの統合
2. **メトリクス統合**: 依存関係初期化の自動メトリクス
3. **ホットリロード**: 再起動なしでの設定リロードサポート
4. **サービスディスカバリー**: サービスディスカバリーメカニズムとの統合

### 移行ガイドライン

将来のサービス向け:

1. `New{ServiceName}`命名規則に従ったプロバイダー関数の作成
2. テスタビリティのための明確なインターフェース定義
3. 適切なカテゴリでのプロバイダー登録
4. アプリケーション固有バンドル（`ServerApp`、`SchedulerApp`等）の使用
5. `Cleanup`型での適切なリソースクリーンアップの実装

## 関連ADR

- ADR-0004: Redis Pub/Sub実装（DIで管理される依存関係）
- ADR-0005: スケジューラーシステムアーキテクチャ（DIでリファクタリング）

## 参考文献

- [uber-go/dig documentation](https://pkg.go.dev/go.uber.org/dig)
- [Dependency Injection in Go](https://blog.golang.org/dependency-injection)
- [Clean Architecture patterns](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
