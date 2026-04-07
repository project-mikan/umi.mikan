package database_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/project-mikan/umi.mikan/backend/infrastructure/database"
	"github.com/project-mikan/umi.mikan/backend/testutil"
)

func TestUpsertDiaryChunkEmbeddings(t *testing.T) {
	db := testutil.SetupTestDB(t)
	userID := testutil.CreateTestUser(t, db, "diary-embeddings-test@example.com", "EmbeddingsUser")
	ctx := context.Background()

	// テスト用日記を挿入
	diaryID := uuid.New()
	now := time.Now().UnixMilli()
	_, err := db.ExecContext(ctx,
		`INSERT INTO diaries (id, user_id, content, date, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`,
		diaryID, userID, "今日は朝ジムに行った。夜は友人と食事をした。", "2024-01-15", now, now,
	)
	if err != nil {
		t.Fatalf("日記の挿入に失敗: %v", err)
	}

	// ダミーembedding（3072次元のゼロベクトル）
	dummyEmbedding := make([]float32, 3072)
	dummyEmbedding2 := make([]float32, 3072)
	dummyEmbedding2[0] = 1.0

	t.Run("複数チャンクを正常にUpsertできる", func(t *testing.T) {
		chunks := []database.DiaryChunk{
			{Index: 0, Content: "今日は朝ジムに行った。", Summary: "朝のジムトレーニング", Embedding: dummyEmbedding},
			{Index: 1, Content: "夜は友人と食事をした。", Summary: "友人との夕食", Embedding: dummyEmbedding2},
		}
		err := database.UpsertDiaryChunkEmbeddings(ctx, db, diaryID, userID, chunks, "gemini-embedding-001")
		if err != nil {
			t.Fatalf("UpsertDiaryChunkEmbeddings失敗: %v", err)
		}

		// チャンク数を確認
		var count int
		if err := db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM diary_embeddings WHERE diary_id = $1`, diaryID,
		).Scan(&count); err != nil {
			t.Fatalf("チャンク数の取得に失敗: %v", err)
		}
		if count != 2 {
			t.Errorf("期待チャンク数 2 に対して %d", count)
		}
	})

	t.Run("再Upsert時に既存チャンクが置き換わる", func(t *testing.T) {
		// 1チャンクで上書き
		chunks := []database.DiaryChunk{
			{Index: 0, Content: "更新後のチャンク", Embedding: dummyEmbedding},
		}
		err := database.UpsertDiaryChunkEmbeddings(ctx, db, diaryID, userID, chunks, "gemini-embedding-001")
		if err != nil {
			t.Fatalf("UpsertDiaryChunkEmbeddings失敗: %v", err)
		}

		var count int
		if err := db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM diary_embeddings WHERE diary_id = $1`, diaryID,
		).Scan(&count); err != nil {
			t.Fatalf("チャンク数の取得に失敗: %v", err)
		}
		if count != 1 {
			t.Errorf("再upsert後は1チャンクになるべきところ %d チャンク存在する", count)
		}

		var content string
		if err := db.QueryRowContext(ctx,
			`SELECT chunk_content FROM diary_embeddings WHERE diary_id = $1 AND chunk_index = 0`, diaryID,
		).Scan(&content); err != nil {
			t.Fatalf("chunk_content取得に失敗: %v", err)
		}
		if content != "更新後のチャンク" {
			t.Errorf("chunk_content が期待値と異なる: got %q", content)
		}
	})
}

func TestGetDiaryEmbeddingStatus(t *testing.T) {
	db := testutil.SetupTestDB(t)
	userID := testutil.CreateTestUser(t, db, "embedding-status-test@example.com", "EmbeddingStatusUser")
	ctx := context.Background()

	// テスト用日記を挿入
	diaryID := uuid.New()
	now := time.Now().UnixMilli()
	_, err := db.ExecContext(ctx,
		`INSERT INTO diaries (id, user_id, content, date, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`,
		diaryID, userID, "テスト日記", "2024-01-16", now, now,
	)
	if err != nil {
		t.Fatalf("日記の挿入に失敗: %v", err)
	}

	t.Run("embeddingが存在しない場合はIndexed=falseを返す", func(t *testing.T) {
		status, err := database.GetDiaryEmbeddingStatus(ctx, db, diaryID, userID)
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		if status.Indexed {
			t.Error("embeddingが存在しないのにIndexed=trueになっている")
		}
	})

	t.Run("embeddingが存在する場合はIndexed=trueとmodelVersionを返す", func(t *testing.T) {
		dummyEmbedding := make([]float32, 3072)
		chunks := []database.DiaryChunk{
			{Index: 0, Content: "テスト日記", Embedding: dummyEmbedding},
		}
		if err := database.UpsertDiaryChunkEmbeddings(ctx, db, diaryID, userID, chunks, "gemini-embedding-001"); err != nil {
			t.Fatalf("UpsertDiaryChunkEmbeddings失敗: %v", err)
		}

		status, err := database.GetDiaryEmbeddingStatus(ctx, db, diaryID, userID)
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		if !status.Indexed {
			t.Error("embeddingが存在するのにIndexed=falseになっている")
		}
		if status.ModelVersion != "gemini-embedding-001" {
			t.Errorf("ModelVersion が期待値と異なる: got %q", status.ModelVersion)
		}
	})
}

func TestSearchDiaryEntriesByEmbedding(t *testing.T) {
	db := testutil.SetupTestDB(t)
	userID := testutil.CreateTestUser(t, db, "embedding-search-test@example.com", "EmbeddingSearchUser")
	ctx := context.Background()
	now := time.Now().UnixMilli()

	// テスト用日記を2件挿入
	diary1ID := uuid.New()
	_, err := db.ExecContext(ctx,
		`INSERT INTO diaries (id, user_id, content, date, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`,
		diary1ID, userID, "今日はランニングをした。", "2024-02-01", now, now,
	)
	if err != nil {
		t.Fatalf("日記1の挿入に失敗: %v", err)
	}

	diary2ID := uuid.New()
	_, err = db.ExecContext(ctx,
		`INSERT INTO diaries (id, user_id, content, date, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`,
		diary2ID, userID, "今日は読書をした。", "2024-02-02", now, now,
	)
	if err != nil {
		t.Fatalf("日記2の挿入に失敗: %v", err)
	}

	// 日記1: 2チャンク（embeddingは互いに直交するベクトル）
	emb1 := make([]float32, 3072)
	emb1[0] = 1.0
	emb2 := make([]float32, 3072)
	emb2[1] = 1.0

	chunks1 := []database.DiaryChunk{
		{Index: 0, Content: "朝ランニング5km走った", Summary: "朝のランニング", Embedding: emb1},
		{Index: 1, Content: "夜はストレッチをした", Summary: "夜のストレッチ", Embedding: emb2},
	}
	if err := database.UpsertDiaryChunkEmbeddings(ctx, db, diary1ID, userID, chunks1, "gemini-embedding-001"); err != nil {
		t.Fatalf("日記1のchunk upsertに失敗: %v", err)
	}

	// 日記2: 1チャンク
	emb3 := make([]float32, 3072)
	emb3[2] = 1.0
	chunks2 := []database.DiaryChunk{
		{Index: 0, Content: "小説を3章まで読んだ", Embedding: emb3},
	}
	if err := database.UpsertDiaryChunkEmbeddings(ctx, db, diary2ID, userID, chunks2, "gemini-embedding-001"); err != nil {
		t.Fatalf("日記2のchunk upsertに失敗: %v", err)
	}

	t.Run("類似度閾値以上の結果のみ返す", func(t *testing.T) {
		// emb1と完全一致するクエリ（類似度1.0）
		queryEmbedding := make([]float32, 3072)
		queryEmbedding[0] = 1.0

		results, err := database.SearchDiaryEntriesByEmbedding(ctx, db, userID, queryEmbedding, 10, 0.9)
		if err != nil {
			t.Fatalf("SearchDiaryEntriesByEmbedding失敗: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("閾値0.9で期待件数 1 に対して %d 件", len(results))
		}
		if len(results) > 0 && results[0].DiaryID != diary1ID {
			t.Errorf("diary1 がヒットするはずが %v がヒットした", results[0].DiaryID)
		}
	})

	t.Run("1日記につき1件のみ返す（最高類似度チャンクを選ぶ）", func(t *testing.T) {
		queryEmbedding := make([]float32, 3072)
		queryEmbedding[0] = 1.0

		results, err := database.SearchDiaryEntriesByEmbedding(ctx, db, userID, queryEmbedding, 10, 0.5)
		if err != nil {
			t.Fatalf("SearchDiaryEntriesByEmbedding失敗: %v", err)
		}

		// 日記ごとに1件であることを確認
		diaryIDsFound := make(map[uuid.UUID]int)
		for _, r := range results {
			diaryIDsFound[r.DiaryID]++
		}
		for id, count := range diaryIDsFound {
			if count > 1 {
				t.Errorf("日記 %v が %d 件重複している（1件のみ期待）", id, count)
			}
		}
	})

	t.Run("ChunkContentが検索結果に含まれる", func(t *testing.T) {
		queryEmbedding := make([]float32, 3072)
		queryEmbedding[0] = 1.0

		results, err := database.SearchDiaryEntriesByEmbedding(ctx, db, userID, queryEmbedding, 10, 0.9)
		if err != nil {
			t.Fatalf("SearchDiaryEntriesByEmbedding失敗: %v", err)
		}
		if len(results) == 0 {
			t.Fatal("結果が0件")
		}
		if results[0].ChunkContent == "" {
			t.Error("ChunkContent が空になっている")
		}
		if results[0].ChunkContent != "朝ランニング5km走った" {
			t.Errorf("ChunkContent が期待値と異なる: got %q", results[0].ChunkContent)
		}
	})

	t.Run("ChunkSummaryとChunkCountが検索結果に含まれる", func(t *testing.T) {
		queryEmbedding := make([]float32, 3072)
		queryEmbedding[0] = 1.0

		results, err := database.SearchDiaryEntriesByEmbedding(ctx, db, userID, queryEmbedding, 10, 0.9)
		if err != nil {
			t.Fatalf("SearchDiaryEntriesByEmbedding失敗: %v", err)
		}
		if len(results) == 0 {
			t.Fatal("結果が0件")
		}
		// 日記1は2チャンク持つ
		if results[0].ChunkCount != 2 {
			t.Errorf("ChunkCount が期待値 2 と異なる: got %d", results[0].ChunkCount)
		}
		// マッチしたのはchunk_index=0のチャンクなのでsummaryは"朝のランニング"
		if results[0].ChunkSummary != "朝のランニング" {
			t.Errorf("ChunkSummary が期待値と異なる: got %q", results[0].ChunkSummary)
		}
	})
}

func TestSearchDiaryEntriesByEmbedding_LimitEdgeCases(t *testing.T) {
	db := testutil.SetupTestDB(t)
	userID := testutil.CreateTestUser(t, db, "embedding-limit-test@example.com", "LimitUser")
	ctx := context.Background()

	// limit=0 と limit>50 のデフォルト制限動作を確認（エラーが出ないことを検証）
	queryEmbedding := make([]float32, 3072)

	t.Run("limit=0の場合デフォルト値で動作する", func(t *testing.T) {
		// エラーなしに動作することを確認（結果は0件でも可）
		results, err := database.SearchDiaryEntriesByEmbedding(ctx, db, userID, queryEmbedding, 0, 0.5)
		if err != nil {
			t.Fatalf("SearchDiaryEntriesByEmbedding(limit=0)失敗: %v", err)
		}
		// 結果は0件でも問題なし（データが存在しないため）
		_ = results
	})

	t.Run("limit>50の場合上限50で動作する", func(t *testing.T) {
		results, err := database.SearchDiaryEntriesByEmbedding(ctx, db, userID, queryEmbedding, 100, 0.5)
		if err != nil {
			t.Fatalf("SearchDiaryEntriesByEmbedding(limit=100)失敗: %v", err)
		}
		_ = results
	})
}

func TestSetHNSWEfSearch(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := context.Background()

	t.Run("トランザクション内でhnsw.ef_searchを設定できる", func(t *testing.T) {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			t.Fatalf("トランザクション開始失敗: %v", err)
		}
		defer func() { _ = tx.Rollback() }()

		if err := database.SetHNSWEfSearch(ctx, tx, 100); err != nil {
			t.Fatalf("SetHNSWEfSearch失敗: %v", err)
		}
	})
}

func TestInsertSemanticSearchLog(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := context.Background()
	userID := testutil.CreateTestUser(t, db, "semantic-search-log@example.com", "User")

	t.Run("意味的検索ログを挿入できる", func(t *testing.T) {
		if err := database.InsertSemanticSearchLog(ctx, db, userID); err != nil {
			t.Fatalf("InsertSemanticSearchLog失敗: %v", err)
		}

		var count int
		if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM semantic_search_logs WHERE user_id = $1`, userID).Scan(&count); err != nil {
			t.Fatalf("カウントクエリ失敗: %v", err)
		}
		if count != 1 {
			t.Errorf("期待 1, 実際 %d", count)
		}
	})
}

func TestDiaryIDsWithoutEmbeddings(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := context.Background()
	userID := testutil.CreateTestUser(t, db, "diary-ids-without-embeddings@example.com", "User")

	diaryID := uuid.New()
	now := time.Now().UnixMilli()
	if _, err := db.ExecContext(ctx,
		`INSERT INTO diaries (id, user_id, content, date, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`,
		diaryID, userID, "embedding未生成の日記", "2024-01-01", now, now,
	); err != nil {
		t.Fatalf("日記の挿入に失敗: %v", err)
	}

	t.Run("embedding未生成の日記IDを返す", func(t *testing.T) {
		ids, err := database.DiaryIDsWithoutEmbeddings(ctx, db, userID)
		if err != nil {
			t.Fatalf("DiaryIDsWithoutEmbeddings失敗: %v", err)
		}
		if len(ids) != 1 {
			t.Errorf("期待 1件, 実際 %d件", len(ids))
		}
		if len(ids) > 0 && ids[0] != diaryID.String() {
			t.Errorf("期待ID %s, 実際 %s", diaryID, ids[0])
		}
	})

	t.Run("embedding生成済みの日記は含まない", func(t *testing.T) {
		dummyEmbedding := make([]float32, 3072)
		chunks := []database.DiaryChunk{
			{Index: 0, Content: "テスト", Summary: "テスト概要", Embedding: dummyEmbedding},
		}
		if err := database.UpsertDiaryChunkEmbeddings(ctx, db, diaryID, userID, chunks, "gemini-embedding-001"); err != nil {
			t.Fatalf("UpsertDiaryChunkEmbeddings失敗: %v", err)
		}

		ids, err := database.DiaryIDsWithoutEmbeddings(ctx, db, userID)
		if err != nil {
			t.Fatalf("DiaryIDsWithoutEmbeddings失敗: %v", err)
		}
		if len(ids) != 0 {
			t.Errorf("embedding生成済みで0件を期待したが %d件", len(ids))
		}
	})
}

func TestUpsertDiaryChunkEmbeddings_NonSQLDB(t *testing.T) {
	// db が *sql.DB でない場合にエラーが返ることを確認
	// NOTE: 現在の実装では *sql.DB 以外は受け付けないが、
	// DBインターフェースの別実装はないためこのケースはスキップ
	t.Skip("*sql.DB以外のDBインターフェース実装がないためスキップ")
}
