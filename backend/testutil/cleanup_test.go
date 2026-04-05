package testutil

import (
	"slices"
	"testing"
)

// テスト用のFKグラフ構造（実際のスキーマを模した簡略版）
var testEdges = []fkEdge{
	// user_password_authes → users
	{childTable: "user_password_authes", childColumn: "user_id", parentTable: "users", parentColumn: "id"},
	// user_llms → users
	{childTable: "user_llms", childColumn: "user_id", parentTable: "users", parentColumn: "id"},
	// diaries → users
	{childTable: "diaries", childColumn: "user_id", parentTable: "users", parentColumn: "id"},
	// diary_embeddings → diaries
	{childTable: "diary_embeddings", childColumn: "diary_id", parentTable: "diaries", parentColumn: "id"},
	// diary_embeddings → users
	{childTable: "diary_embeddings", childColumn: "user_id", parentTable: "users", parentColumn: "id"},
	// diary_highlights → diaries
	{childTable: "diary_highlights", childColumn: "diary_id", parentTable: "diaries", parentColumn: "id"},
	// diary_highlights → users
	{childTable: "diary_highlights", childColumn: "user_id", parentTable: "users", parentColumn: "id"},
	// entities → users
	{childTable: "entities", childColumn: "user_id", parentTable: "users", parentColumn: "id"},
	// entity_aliases → entities（usersへの直接参照なし）
	{childTable: "entity_aliases", childColumn: "entity_id", parentTable: "entities", parentColumn: "id"},
	// semantic_search_logs → users
	{childTable: "semantic_search_logs", childColumn: "user_id", parentTable: "users", parentColumn: "id"},
}

func TestBuildUserRelatedCleanupOrder_usersは最後(t *testing.T) {
	order := buildUserRelatedCleanupOrder(testEdges)

	if len(order) == 0 {
		t.Fatal("空の削除順が返された")
	}
	if order[len(order)-1] != "users" {
		t.Errorf("usersは最後に削除される必要があるが、順序: %v", order)
	}
}

func TestBuildUserRelatedCleanupOrder_子テーブルは親より前(t *testing.T) {
	order := buildUserRelatedCleanupOrder(testEdges)

	indexOf := func(table string) int {
		for i, t := range order {
			if t == table {
				return i
			}
		}
		return -1
	}

	// diary_embeddings と diary_highlights は diaries より前
	diariesIdx := indexOf("diaries")
	if diariesIdx == -1 {
		t.Fatal("diariesが削除順に含まれていない")
	}
	for _, child := range []string{"diary_embeddings", "diary_highlights"} {
		if idx := indexOf(child); idx == -1 || idx >= diariesIdx {
			t.Errorf("%s (index=%d) は diaries (index=%d) より前に削除される必要がある", child, idx, diariesIdx)
		}
	}

	// entity_aliases は entities より前
	entitiesIdx := indexOf("entities")
	if entitiesIdx == -1 {
		t.Fatal("entitiesが削除順に含まれていない")
	}
	if idx := indexOf("entity_aliases"); idx == -1 || idx >= entitiesIdx {
		t.Errorf("entity_aliases (index=%d) は entities (index=%d) より前に削除される必要がある", idx, entitiesIdx)
	}

	// diaries と entities は users より前
	usersIdx := indexOf("users")
	for _, tbl := range []string{"diaries", "entities"} {
		if idx := indexOf(tbl); idx == -1 || idx >= usersIdx {
			t.Errorf("%s (index=%d) は users (index=%d) より前に削除される必要がある", tbl, idx, usersIdx)
		}
	}
}

func TestBuildUserRelatedCleanupOrder_全テーブルを含む(t *testing.T) {
	order := buildUserRelatedCleanupOrder(testEdges)

	expected := []string{
		"user_password_authes", "user_llms", "diaries",
		"diary_embeddings", "diary_highlights", "entities",
		"entity_aliases", "semantic_search_logs", "users",
	}
	for _, tbl := range expected {
		if !slices.Contains(order, tbl) {
			t.Errorf("テーブル %s が削除順に含まれていない（順序: %v）", tbl, order)
		}
	}
}

func TestBuildDeleteWhereClause_直接user_id参照(t *testing.T) {
	edgesFrom := map[string][]fkEdge{}
	for _, e := range testEdges {
		edgesFrom[e.childTable] = append(edgesFrom[e.childTable], e)
	}

	// diaries は user_id で users を直接参照
	got := buildDeleteWhereClause("diaries", edgesFrom, "= $1", map[string]bool{})
	want := "user_id = $1"
	if got != want {
		t.Errorf("diaries: got %q, want %q", got, want)
	}
}

func TestBuildDeleteWhereClause_間接参照(t *testing.T) {
	edgesFrom := map[string][]fkEdge{}
	for _, e := range testEdges {
		edgesFrom[e.childTable] = append(edgesFrom[e.childTable], e)
	}

	// entity_aliases は entity_id → entities.id → user_id → users.id という間接パス
	got := buildDeleteWhereClause("entity_aliases", edgesFrom, "= $1", map[string]bool{})
	want := "entity_id IN (SELECT id FROM entities WHERE user_id = $1)"
	if got != want {
		t.Errorf("entity_aliases: got %q, want %q", got, want)
	}
}

func TestBuildDeleteWhereClause_直接参照を優先(t *testing.T) {
	edgesFrom := map[string][]fkEdge{}
	for _, e := range testEdges {
		edgesFrom[e.childTable] = append(edgesFrom[e.childTable], e)
	}

	// diary_embeddings は diary_id と user_id の両方を持つが、user_idへの直接参照を優先すべき
	got := buildDeleteWhereClause("diary_embeddings", edgesFrom, "= $1", map[string]bool{})
	want := "user_id = $1"
	if got != want {
		t.Errorf("diary_embeddings（直接参照優先）: got %q, want %q", got, want)
	}
}

func TestBuildDeleteWhereClause_emailパターン(t *testing.T) {
	edgesFrom := map[string][]fkEdge{}
	for _, e := range testEdges {
		edgesFrom[e.childTable] = append(edgesFrom[e.childTable], e)
	}

	emailCond := "IN (SELECT id FROM users WHERE email LIKE $1)"

	// entity_aliases の email パターン
	got := buildDeleteWhereClause("entity_aliases", edgesFrom, emailCond, map[string]bool{})
	want := "entity_id IN (SELECT id FROM entities WHERE user_id IN (SELECT id FROM users WHERE email LIKE $1))"
	if got != want {
		t.Errorf("entity_aliases（email）: got %q, want %q", got, want)
	}
}
