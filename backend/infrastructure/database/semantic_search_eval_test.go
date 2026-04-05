//go:build integration

package database_test

// セマンティック検索（RAG）の有効性を評価する統合テスト。
// 実行条件:
//   - GEMINI_API_KEY_FOR_TEST 環境変数が設定されていること
//   - テストDBが起動していること（make db-apply-test 済み）
//
// 実行方法:
//   GEMINI_API_KEY_FOR_TEST=xxx make b-test-semantic-eval
//
// このテストは以下を計測する:
//   1. 語彙一致なしでの意味的検索（Recall）— evalPositiveThreshold を使用
//   2. 無関係クエリの誤検出なし（Precision）— evalNegativeThreshold を使用
//   3. マルチチャンク日記での特定チャンクへのアクセス
//
// 閾値の根拠:
//   - 正例の最低スコア ≈ 0.696、雑音の最高スコア ≈ 0.614
//   - evalPositiveThreshold(0.4): 正例を確実に拾う低めの閾値（Recall重視）
//   - evalNegativeThreshold(0.65): 信号と雑音のギャップ中間点（Precision検証用）
//   - 本番コードの閾値(0.3)はさらに低く、Recall最大化を優先した設定

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/project-mikan/umi.mikan/backend/infrastructure/database"
	"github.com/project-mikan/umi.mikan/backend/infrastructure/llm"
	"github.com/project-mikan/umi.mikan/backend/testutil"
)

const (
	// evalPositiveThreshold は正例クエリ評価の類似度閾値（Recall重視）
	// 正例クエリの最低スコア(≈0.696)を十分下回る値を設定
	evalPositiveThreshold = 0.4
	// evalNegativeThreshold はネガティブケース評価の類似度閾値（Precision検証用）
	// Geminiのembeddingモデルは無関係な日本語テキストでも0.55-0.62程度のスコアが出る特性がある。
	// 正例の最低スコア(≈0.696)と雑音の最高スコア(≈0.614)の中間点として0.65を設定する。
	evalNegativeThreshold = 0.65
)

// evalDataset は評価データセットのルート構造
type evalDataset struct {
	Description           string                    `json:"description"`
	DiaryEntries          []evalDiaryEntry          `json:"diary_entries"`
	EvalQueries           []evalQuery               `json:"evaluation_queries"`
	MultiChunkQueries     []multiChunkQuery         `json:"multi_chunk_queries"`
	PredefinedChunkEntries []predefinedChunkEntry   `json:"predefined_chunk_entries"`
	NegativeCases         []evalNegativeCase        `json:"negative_cases"`
}

// predefinedChunkEntry はLLM分割に依存せず事前定義したチャンク構成
// マルチチャンク日記の検索テストで確実に複数チャンクを挿入するために使用する
type predefinedChunkEntry struct {
	DiaryID string              `json:"diary_id"`
	Chunks  []predefinedChunk   `json:"chunks"`
}

type predefinedChunk struct {
	Content string `json:"content"`
	Summary string `json:"summary"`
}

type evalDiaryEntry struct {
	ID      string   `json:"id"`
	Date    string   `json:"date"`
	Content string   `json:"content"`
	Topics  []string `json:"topics"`
}

type evalQuery struct {
	ID               string   `json:"id"`
	Query            string   `json:"query"`
	ExpectedEntryIDs []string `json:"expected_entry_ids"`
	KeywordProbe     string   `json:"keyword_probe"`
	WhyKeywordFails  string   `json:"why_keyword_fails"`
	Category         string   `json:"category"`
}

// multiChunkQuery はマルチチャンク日記の特定チャンクへのアクセスを検証するクエリ
type multiChunkQuery struct {
	ID                  string   `json:"id"`
	Query               string   `json:"query"`
	ExpectedEntryIDs    []string `json:"expected_entry_ids"`
	KeywordProbe        string   `json:"keyword_probe"`
	ExpectedMinChunks   int      `json:"expected_min_chunk_count"`
	WhyKeywordFails     string   `json:"why_keyword_fails"`
	Category            string   `json:"category"`
}

type evalNegativeCase struct {
	ID               string   `json:"id"`
	Query            string   `json:"query"`
	ExpectedEntryIDs []string `json:"expected_entry_ids"`
	Description      string   `json:"description"`
}

// evalQueryResult はクエリ1件の評価結果
type evalQueryResult struct {
	Query               string
	Category            string
	ExpectedIDs         []string
	SemanticHitIDs      []string
	KeywordHitIDs       []string
	TopSimilarity       float64
	MatchedChunkContent string
	MatchedChunkCount   int
	WhyKeywordFails     string
}

// multiChunkQueryResult はマルチチャンククエリ1件の評価結果
type multiChunkQueryResult struct {
	Query               string
	Category            string
	ExpectedIDs         []string
	ExpectedMinChunks   int
	SemanticHitIDs      []string
	TopSimilarity       float64
	MatchedChunkContent string
	MatchedChunkSummary string
	ActualChunkCount    int
	WhyKeywordFails     string
}

func TestSemanticSearchEvaluation(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY_FOR_TEST")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY_FOR_TEST が設定されていないためスキップ（セマンティック検索の評価には実際のGemini APIが必要）")
	}

	db := testutil.SetupTestDB(t)
	ctx := context.Background()

	geminiClient, err := llm.NewGeminiClient(ctx, apiKey)
	if err != nil {
		t.Fatalf("GeminiClientの初期化に失敗: %v", err)
	}
	defer func() { _ = geminiClient.Close() }()

	dataset := loadEvalDataset(t)
	userID := testutil.CreateTestUser(t, db, "semantic-eval@example.com", "SemanticEvalUser")

	// 評価用日記を挿入してembeddingを生成・保存
	t.Log("=== 日記エントリのembedding生成中（Gemini API使用） ===")
	entryIDMap := make(map[string]uuid.UUID)
	now := time.Now().UnixMilli()

	// 事前定義チャンクのマップを構築（マルチチャンクテスト用）
	predefinedChunkMap := make(map[string]predefinedChunkEntry)
	for _, pce := range dataset.PredefinedChunkEntries {
		predefinedChunkMap[pce.DiaryID] = pce
	}

	for i, entry := range dataset.DiaryEntries {
		diaryID := uuid.New()
		entryIDMap[entry.ID] = diaryID

		_, err := db.ExecContext(ctx,
			`INSERT INTO diaries (id, user_id, content, date, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`,
			diaryID, userID, entry.Content, entry.Date, now+int64(i), now+int64(i),
		)
		if err != nil {
			t.Fatalf("日記[%s]の挿入に失敗: %v", entry.ID, err)
		}

		var chunks []database.DiaryChunk
		// 事前定義チャンクがある場合はLLM分割をバイパスして確実に複数チャンクを挿入する
		if pce, ok := predefinedChunkMap[entry.ID]; ok {
			chunks, err = evalGeneratePredefinedChunks(ctx, geminiClient, pce.Chunks, entry.Date)
			if err != nil {
				t.Fatalf("日記[%s]の事前定義チャンクembedding生成に失敗: %v", entry.ID, err)
			}
		} else {
			chunks, err = evalGenerateChunksWithEmbedding(ctx, geminiClient, entry.Content, entry.Date)
			if err != nil {
				t.Fatalf("日記[%s]のembedding生成に失敗: %v", entry.ID, err)
			}
		}

		if err := database.UpsertDiaryChunkEmbeddings(ctx, db, diaryID, userID, chunks, llm.ModelEmbedding); err != nil {
			t.Fatalf("日記[%s]のchunk upsertに失敗: %v", entry.ID, err)
		}

		t.Logf("  [%d/%d] %s: %d チャンク生成完了", i+1, len(dataset.DiaryEntries), entry.ID, len(chunks))
	}

	// セマンティック vs キーワード 比較評価
	t.Log("")
	t.Log("=== セマンティック検索 vs キーワード検索の比較評価 ===")
	t.Log("")
	results := make([]evalQueryResult, 0, len(dataset.EvalQueries))
	for _, eq := range dataset.EvalQueries {
		results = append(results, runEvalQuery(ctx, t, db, geminiClient, userID, eq, entryIDMap))
	}

	// マルチチャンク評価
	t.Log("=== マルチチャンク：特定チャンクへのアクセス検証 ===")
	t.Log("")
	mcResults := make([]multiChunkQueryResult, 0, len(dataset.MultiChunkQueries))
	for _, mq := range dataset.MultiChunkQueries {
		mcResults = append(mcResults, runMultiChunkQuery(ctx, t, db, geminiClient, userID, mq, entryIDMap))
	}

	// ネガティブケース評価
	t.Log("=== ネガティブケース（誤検出なしを確認） ===")
	for _, nc := range dataset.NegativeCases {
		runNegativeCase(ctx, t, db, geminiClient, userID, nc)
	}

	printEvalReport(t, results)
	printMultiChunkReport(t, mcResults)
}

// evalGenerateChunksWithEmbedding は日記テキストをチャンク分割しembeddingを生成する
func evalGenerateChunksWithEmbedding(ctx context.Context, client *llm.GeminiClient, content, date string) ([]database.DiaryChunk, error) {
	rawChunks, err := client.SplitDiaryIntoChunks(ctx, content)
	if err != nil || len(rawChunks) == 0 {
		rawChunks = []llm.DiaryChunkData{{Content: content, Summary: ""}}
	}

	datePrefix := evalFormatDatePrefix(date)
	chunks := make([]database.DiaryChunk, 0, len(rawChunks))
	for i, raw := range rawChunks {
		embedding, err := client.GenerateEmbedding(ctx, datePrefix+raw.Content, true)
		if err != nil {
			return nil, fmt.Errorf("チャンク%dのembedding生成に失敗: %w", i, err)
		}
		chunks = append(chunks, database.DiaryChunk{
			Index:     i,
			Content:   raw.Content,
			Summary:   raw.Summary,
			Embedding: embedding,
		})
	}
	return chunks, nil
}

// evalGeneratePredefinedChunks は事前定義チャンクのembeddingを生成する
// LLM分割に頼らず確実に複数チャンクをDBに挿入するためのマルチチャンクテスト用関数
func evalGeneratePredefinedChunks(ctx context.Context, client *llm.GeminiClient, rawChunks []predefinedChunk, date string) ([]database.DiaryChunk, error) {
	datePrefix := evalFormatDatePrefix(date)
	chunks := make([]database.DiaryChunk, 0, len(rawChunks))
	for i, raw := range rawChunks {
		embedding, err := client.GenerateEmbedding(ctx, datePrefix+raw.Content, true)
		if err != nil {
			return nil, fmt.Errorf("事前定義チャンク%dのembedding生成に失敗: %w", i, err)
		}
		chunks = append(chunks, database.DiaryChunk{
			Index:     i,
			Content:   raw.Content,
			Summary:   raw.Summary,
			Embedding: embedding,
		})
	}
	return chunks, nil
}

// evalFormatDatePrefix は "YYYY-MM-DD" を "YYYY年M月D日の日記: " に変換する
func evalFormatDatePrefix(dateStr string) string {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%d年%d月%d日の日記: ", t.Year(), t.Month(), t.Day())
}

// runEvalQuery はクエリ1件のセマンティック検索とキーワード検索を実行して比較する
func runEvalQuery(
	ctx context.Context,
	t *testing.T,
	db *sql.DB,
	client *llm.GeminiClient,
	userID uuid.UUID,
	eq evalQuery,
	entryIDMap map[string]uuid.UUID,
) evalQueryResult {
	t.Helper()

	queryEmbedding, err := client.GenerateEmbedding(ctx, eq.Query, false)
	if err != nil {
		t.Fatalf("クエリ[%s]のembedding生成に失敗: %v", eq.ID, err)
	}

	semanticResults, err := database.SearchDiaryEntriesByEmbedding(ctx, db, userID, queryEmbedding, 10, evalPositiveThreshold)
	if err != nil {
		t.Fatalf("クエリ[%s]のセマンティック検索に失敗: %v", eq.ID, err)
	}

	semanticHitIDs := make([]string, 0)
	topSimilarity := 0.0
	var matchedChunkContent, matchedChunkSummary string
	matchedChunkCount := 0
	for _, sr := range semanticResults {
		for evalID, dbID := range entryIDMap {
			if sr.DiaryID == dbID {
				semanticHitIDs = append(semanticHitIDs, evalID)
				if sr.Similarity > topSimilarity {
					topSimilarity = sr.Similarity
					matchedChunkContent = sr.ChunkContent
					matchedChunkSummary = sr.ChunkSummary
					matchedChunkCount = sr.ChunkCount
				}
			}
		}
	}
	_ = matchedChunkSummary

	keywordHitIDs := evalKeywordSearch(ctx, t, db, userID, eq.KeywordProbe, entryIDMap)

	result := evalQueryResult{
		Query:               eq.Query,
		Category:            eq.Category,
		ExpectedIDs:         eq.ExpectedEntryIDs,
		SemanticHitIDs:      semanticHitIDs,
		KeywordHitIDs:       keywordHitIDs,
		TopSimilarity:       topSimilarity,
		MatchedChunkContent: matchedChunkContent,
		MatchedChunkCount:   matchedChunkCount,
		WhyKeywordFails:     eq.WhyKeywordFails,
	}

	t.Logf("クエリ: 「%s」", eq.Query)
	t.Logf("  カテゴリ: %s", eq.Category)
	t.Logf("  期待: %v", eq.ExpectedEntryIDs)
	t.Logf("  セマンティック結果: %v（最高類似度: %.3f）", semanticHitIDs, topSimilarity)
	if matchedChunkContent != "" {
		t.Logf("  マッチチャンク(チャンク数:%d): %s", matchedChunkCount, evalTruncate(matchedChunkContent, 50))
	}
	t.Logf("  キーワード(%q)結果: %v", eq.KeywordProbe, keywordHitIDs)
	t.Logf("  なぜキーワード検索が失敗するか: %s", eq.WhyKeywordFails)

	if evalContainsAny(semanticHitIDs, eq.ExpectedEntryIDs) {
		t.Logf("  ✓ PASS: セマンティック検索が期待エントリを正しく発見")
	} else {
		t.Errorf("  ✗ FAIL: セマンティック検索が期待エントリ%vを見つけられなかった（閾値%.2f）", eq.ExpectedEntryIDs, evalPositiveThreshold)
	}

	if evalContainsAny(keywordHitIDs, eq.ExpectedEntryIDs) {
		t.Logf("  ℹ 注意: このクエリはキーワード検索でも見つかる（語彙が重複している）")
	} else {
		t.Logf("  ✓ 確認: キーワード検索では見つからず → セマンティック検索の優位性を示す")
	}

	t.Log("")
	return result
}

// runMultiChunkQuery はマルチチャンク日記に対して特定チャンクでヒットすることを検証する
func runMultiChunkQuery(
	ctx context.Context,
	t *testing.T,
	db *sql.DB,
	client *llm.GeminiClient,
	userID uuid.UUID,
	mq multiChunkQuery,
	entryIDMap map[string]uuid.UUID,
) multiChunkQueryResult {
	t.Helper()

	queryEmbedding, err := client.GenerateEmbedding(ctx, mq.Query, false)
	if err != nil {
		t.Fatalf("マルチチャンククエリ[%s]のembedding生成に失敗: %v", mq.ID, err)
	}

	semanticResults, err := database.SearchDiaryEntriesByEmbedding(ctx, db, userID, queryEmbedding, 10, evalPositiveThreshold)
	if err != nil {
		t.Fatalf("マルチチャンククエリ[%s]のセマンティック検索に失敗: %v", mq.ID, err)
	}

	semanticHitIDs := make([]string, 0)
	topSimilarity := 0.0
	var matchedChunkContent, matchedChunkSummary string
	actualChunkCount := 0

	for _, sr := range semanticResults {
		for evalID, dbID := range entryIDMap {
			if sr.DiaryID == dbID {
				semanticHitIDs = append(semanticHitIDs, evalID)
				if sr.Similarity > topSimilarity {
					topSimilarity = sr.Similarity
					matchedChunkContent = sr.ChunkContent
					matchedChunkSummary = sr.ChunkSummary
					actualChunkCount = sr.ChunkCount
				}
			}
		}
	}

	result := multiChunkQueryResult{
		Query:               mq.Query,
		Category:            mq.Category,
		ExpectedIDs:         mq.ExpectedEntryIDs,
		ExpectedMinChunks:   mq.ExpectedMinChunks,
		SemanticHitIDs:      semanticHitIDs,
		TopSimilarity:       topSimilarity,
		MatchedChunkContent: matchedChunkContent,
		MatchedChunkSummary: matchedChunkSummary,
		ActualChunkCount:    actualChunkCount,
		WhyKeywordFails:     mq.WhyKeywordFails,
	}

	t.Logf("マルチチャンククエリ: 「%s」", mq.Query)
	t.Logf("  カテゴリ: %s", mq.Category)
	t.Logf("  期待エントリ: %v（最小チャンク数: %d以上）", mq.ExpectedEntryIDs, mq.ExpectedMinChunks)
	t.Logf("  セマンティック結果: %v（最高類似度: %.3f）", semanticHitIDs, topSimilarity)
	if matchedChunkContent != "" {
		t.Logf("  マッチチャンク内容: %s", evalTruncate(matchedChunkContent, 60))
		t.Logf("  マッチチャンク概要: %s", matchedChunkSummary)
		t.Logf("  チャンク総数: %d（期待最小値: %d）", actualChunkCount, mq.ExpectedMinChunks)
	}
	t.Logf("  なぜキーワード検索が失敗するか: %s", mq.WhyKeywordFails)

	// アサーション1: 期待エントリが見つかること
	if evalContainsAny(semanticHitIDs, mq.ExpectedEntryIDs) {
		t.Logf("  ✓ PASS: セマンティック検索が期待エントリを発見")
	} else {
		t.Errorf("  ✗ FAIL: 期待エントリ%vが見つからなかった", mq.ExpectedEntryIDs)
	}

	// アサーション2: マッチした日記が複数チャンクを持つこと
	if actualChunkCount >= mq.ExpectedMinChunks {
		t.Logf("  ✓ PASS: マルチチャンク確認（%dチャンク ≥ 最小値%d）", actualChunkCount, mq.ExpectedMinChunks)
	} else if actualChunkCount > 0 {
		t.Errorf("  ✗ FAIL: チャンク数が期待を下回る（%dチャンク < 最小値%d）— 日記が十分に分割されていない可能性", actualChunkCount, mq.ExpectedMinChunks)
	}

	t.Log("")
	return result
}

// runNegativeCase は無関係なクエリで誤検出がないことを確認する
func runNegativeCase(
	ctx context.Context,
	t *testing.T,
	db *sql.DB,
	client *llm.GeminiClient,
	userID uuid.UUID,
	nc evalNegativeCase,
) {
	t.Helper()

	queryEmbedding, err := client.GenerateEmbedding(ctx, nc.Query, false)
	if err != nil {
		t.Fatalf("ネガティブクエリ[%s]のembedding生成に失敗: %v", nc.ID, err)
	}

	semanticResults, err := database.SearchDiaryEntriesByEmbedding(ctx, db, userID, queryEmbedding, 10, evalNegativeThreshold)
	if err != nil {
		t.Fatalf("ネガティブクエリ[%s]の検索に失敗: %v", nc.ID, err)
	}

	t.Logf("ネガティブクエリ: 「%s」", nc.Query)
	t.Logf("  セマンティック結果件数: %d件（閾値%.2f）", len(semanticResults), evalNegativeThreshold)

	if len(semanticResults) == 0 {
		t.Logf("  ✓ PASS: 誤検出なし（全エントリが閾値%.2f未満）", evalNegativeThreshold)
	} else {
		maxSim := 0.0
		for _, r := range semanticResults {
			if r.Similarity > maxSim {
				maxSim = r.Similarity
			}
		}
		t.Errorf("  ✗ FAIL: %d件ヒット（最高類似度: %.3f）— 無関係クエリが閾値%.2fを超えている", len(semanticResults), maxSim, evalNegativeThreshold)
	}
	t.Log("")
}

// evalKeywordSearch はdiariesテーブルにLIKE検索を行い、ヒットした日記のevalIDを返す
func evalKeywordSearch(
	ctx context.Context,
	t *testing.T,
	db *sql.DB,
	userID uuid.UUID,
	keyword string,
	entryIDMap map[string]uuid.UUID,
) []string {
	t.Helper()

	rows, err := db.QueryContext(ctx,
		`SELECT id FROM diaries WHERE user_id = $1 AND content LIKE $2`,
		userID, "%"+keyword+"%",
	)
	if err != nil {
		t.Fatalf("キーワード検索に失敗: %v", err)
	}
	defer func() { _ = rows.Close() }()

	hits := make([]string, 0)
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			t.Fatalf("キーワード検索結果のスキャンに失敗: %v", err)
		}
		for evalID, dbID := range entryIDMap {
			if dbID == id {
				hits = append(hits, evalID)
			}
		}
	}
	return hits
}

// printEvalReport は通常クエリの評価レポートを表示する
func printEvalReport(t *testing.T, results []evalQueryResult) {
	t.Helper()

	t.Log("============================================================")
	t.Log("  セマンティック検索 評価レポート（語彙非一致クエリ）")
	t.Log("============================================================")

	totalQueries := len(results)
	semanticRecall := 0
	keywordRecall := 0
	semanticWins := 0

	for _, r := range results {
		semanticFound := evalContainsAny(r.SemanticHitIDs, r.ExpectedIDs)
		keywordFound := evalContainsAny(r.KeywordHitIDs, r.ExpectedIDs)
		if semanticFound {
			semanticRecall++
		}
		if keywordFound {
			keywordRecall++
		}
		if semanticFound && !keywordFound {
			semanticWins++
		}
	}

	semanticRecallRate := float64(semanticRecall) / float64(totalQueries) * 100
	keywordRecallRate := float64(keywordRecall) / float64(totalQueries) * 100

	t.Log("")
	t.Logf("  総クエリ数:                   %d", totalQueries)
	t.Logf("  セマンティック検索 Recall:    %.0f%% (%d/%d)", semanticRecallRate, semanticRecall, totalQueries)
	t.Logf("  キーワード検索 Recall:        %.0f%% (%d/%d)", keywordRecallRate, keywordRecall, totalQueries)
	t.Logf("  セマンティックのみ成功件数:   %d件（キーワードでは見つからない日記を発見）", semanticWins)
	t.Log("")

	type catStats struct{ semantic, keyword int }
	categories := make(map[string]*catStats)
	for _, r := range results {
		if _, ok := categories[r.Category]; !ok {
			categories[r.Category] = &catStats{}
		}
		if evalContainsAny(r.SemanticHitIDs, r.ExpectedIDs) {
			categories[r.Category].semantic++
		}
		if evalContainsAny(r.KeywordHitIDs, r.ExpectedIDs) {
			categories[r.Category].keyword++
		}
	}

	t.Log("  カテゴリ別成功率:")
	for cat, stats := range categories {
		t.Logf("    %-25s  セマンティック: %d件  キーワード: %d件", cat, stats.semantic, stats.keyword)
	}

	t.Log("")
	t.Log("  結論:")
	if semanticRecall > keywordRecall {
		t.Logf("  ✓ セマンティック検索はキーワード検索より %.0f ポイント高い Recall を達成",
			semanticRecallRate-keywordRecallRate)
		t.Logf("  ✓ %d件のクエリで、語彙が一致しなくても意味的に正しい日記を発見できた", semanticWins)
	} else {
		t.Log("  ℹ セマンティック検索の優位性を示す十分な差が得られなかった")
	}
	t.Log("============================================================")

	if semanticRecall < keywordRecall {
		t.Errorf("セマンティック検索のRecall(%d)がキーワード検索(%d)を下回った", semanticRecall, keywordRecall)
	}
	if semanticRecallRate < 70.0 {
		t.Errorf("セマンティック検索のRecall(%.0f%%)が目標値(70%%)を下回った", semanticRecallRate)
	}
}

// printMultiChunkReport はマルチチャンク評価のレポートを表示する
func printMultiChunkReport(t *testing.T, results []multiChunkQueryResult) {
	t.Helper()

	t.Log("")
	t.Log("============================================================")
	t.Log("  マルチチャンク評価レポート")
	t.Log("============================================================")

	totalQueries := len(results)
	entryFound := 0
	chunkCountPassed := 0

	for _, r := range results {
		if evalContainsAny(r.SemanticHitIDs, r.ExpectedIDs) {
			entryFound++
		}
		if r.ActualChunkCount >= r.ExpectedMinChunks {
			chunkCountPassed++
		}
	}

	t.Log("")
	t.Logf("  マルチチャンククエリ数:         %d", totalQueries)
	t.Logf("  期待エントリ発見率:             %d/%d", entryFound, totalQueries)
	t.Logf("  チャンク数条件クリア:           %d/%d（ChunkCount ≥ 期待最小値）", chunkCountPassed, totalQueries)
	t.Log("")
	t.Log("  クエリ別マッチチャンク:")
	for _, r := range results {
		t.Logf("    「%s」", evalTruncate(r.Query, 25))
		t.Logf("      → チャンク数:%d  類似度:%.3f  マッチ内容:%s",
			r.ActualChunkCount, r.TopSimilarity, evalTruncate(r.MatchedChunkContent, 40))
		t.Logf("      → 概要: %s", r.MatchedChunkSummary)
	}
	t.Log("============================================================")
}

// evalContainsAny はsliceの中にtargetsの要素が1つでも含まれるかを確認する
func evalContainsAny(slice, targets []string) bool {
	for _, target := range targets {
		for _, s := range slice {
			if s == target {
				return true
			}
		}
	}
	return false
}

// loadEvalDataset は評価データセットJSONファイルを読み込む
func loadEvalDataset(t *testing.T) *evalDataset {
	t.Helper()

	data, err := os.ReadFile("../../testdata/semantic_eval/dataset.json")
	if err != nil {
		t.Fatalf("評価データセットの読み込みに失敗: %v", err)
	}

	var dataset evalDataset
	if err := json.Unmarshal(data, &dataset); err != nil {
		t.Fatalf("評価データセットのパースに失敗: %v", err)
	}

	if len(dataset.DiaryEntries) == 0 {
		t.Fatal("評価データセットに日記エントリが存在しない")
	}
	if len(dataset.EvalQueries) == 0 {
		t.Fatal("評価データセットに評価クエリが存在しない")
	}

	return &dataset
}

// TestSemanticSearchEvaluation_PrintDataset はデータセットの内容をダンプするヘルパーテスト
// GEMINI_API_KEY不要で実行可能（データセットの確認用）
func TestSemanticSearchEvaluation_PrintDataset(t *testing.T) {
	dataset := loadEvalDataset(t)

	t.Logf("評価データセット: %s", dataset.Description)
	t.Logf("日記エントリ数: %d", len(dataset.DiaryEntries))
	t.Logf("評価クエリ数: %d", len(dataset.EvalQueries))
	t.Logf("マルチチャンククエリ数: %d", len(dataset.MultiChunkQueries))
	t.Logf("ネガティブケース数: %d", len(dataset.NegativeCases))
	t.Log("")

	t.Log("== 日記エントリ一覧 ==")
	for _, e := range dataset.DiaryEntries {
		t.Logf("  [%s] %s: %s...", e.ID, e.Date, evalTruncate(e.Content, 40))
	}
	t.Log("")

	t.Log("== 評価クエリ一覧 ==")
	for _, q := range dataset.EvalQueries {
		t.Logf("  [%s] %s", q.ID, q.Query)
		t.Logf("    期待: %v / キーワードプローブ: %q", q.ExpectedEntryIDs, q.KeywordProbe)
	}
	t.Log("")

	t.Log("== マルチチャンククエリ一覧 ==")
	for _, q := range dataset.MultiChunkQueries {
		t.Logf("  [%s] %s", q.ID, q.Query)
		t.Logf("    期待: %v / 最小チャンク数: %d / キーワードプローブ: %q", q.ExpectedEntryIDs, q.ExpectedMinChunks, q.KeywordProbe)
	}
}

func evalTruncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "..."
}
