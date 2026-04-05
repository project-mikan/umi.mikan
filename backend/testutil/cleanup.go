package testutil

import (
	"database/sql"
	"fmt"
	"sort"
)

// fkEdge はFKの辺（子テーブルのカラムが親テーブルのカラムを参照）
type fkEdge struct {
	childTable   string
	childColumn  string
	parentTable  string
	parentColumn string
}

// loadFKGraph はDBからpublicスキーマのFK情報を取得する
func loadFKGraph(db *sql.DB) ([]fkEdge, error) {
	rows, err := db.Query(`
		SELECT
			tc.table_name  AS child_table,
			kcu.column_name AS child_column,
			ccu.table_name  AS parent_table,
			ccu.column_name AS parent_column
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu
			ON tc.constraint_name = kcu.constraint_name
			AND tc.table_schema   = kcu.table_schema
		JOIN information_schema.referential_constraints rc
			ON tc.constraint_name  = rc.constraint_name
			AND tc.constraint_schema = rc.constraint_schema
		JOIN information_schema.constraint_column_usage ccu
			ON rc.unique_constraint_name = ccu.constraint_name
		WHERE tc.constraint_type = 'FOREIGN KEY'
		  AND tc.table_schema    = 'public'
	`)
	if err != nil {
		return nil, fmt.Errorf("FK情報の取得失敗: %w", err)
	}
	defer rows.Close()

	var edges []fkEdge
	for rows.Next() {
		var e fkEdge
		if err := rows.Scan(&e.childTable, &e.childColumn, &e.parentTable, &e.parentColumn); err != nil {
			return nil, fmt.Errorf("FK情報のスキャン失敗: %w", err)
		}
		edges = append(edges, e)
	}
	return edges, rows.Err()
}

// buildUserRelatedCleanupOrder はusersテーブルから到達可能なテーブルを
// FK依存順（子→親）で返す（usersテーブル自身を最後に含む）
func buildUserRelatedCleanupOrder(edges []fkEdge) []string {
	// childrenOf[parent] = parentを参照している子テーブル一覧
	childrenOf := map[string][]string{}
	for _, e := range edges {
		childrenOf[e.parentTable] = append(childrenOf[e.parentTable], e.childTable)
	}

	// usersテーブルから到達可能なテーブルをBFSで収集
	reachable := map[string]bool{"users": true}
	queue := []string{"users"}
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		for _, child := range childrenOf[curr] {
			if !reachable[child] {
				reachable[child] = true
				queue = append(queue, child)
			}
		}
	}

	// in-degree[t] = reachableな子テーブルのうちtを参照しているテーブル数
	inDegree := map[string]int{}
	for t := range reachable {
		inDegree[t] = 0
	}
	for _, e := range edges {
		if reachable[e.childTable] && reachable[e.parentTable] {
			inDegree[e.parentTable]++
		}
	}

	// parentsOf[child] = childが参照している親テーブル一覧
	parentsOf := map[string][]string{}
	for _, e := range edges {
		if reachable[e.childTable] && reachable[e.parentTable] {
			parentsOf[e.childTable] = append(parentsOf[e.childTable], e.parentTable)
		}
	}

	// Kahnのアルゴリズム: in-degree=0の葉（誰にも参照されていないテーブル）から処理
	var starts []string
	for t := range reachable {
		if inDegree[t] == 0 {
			starts = append(starts, t)
		}
	}
	sort.Strings(starts)

	var result []string
	processQueue := starts
	for len(processQueue) > 0 {
		curr := processQueue[0]
		processQueue = processQueue[1:]
		result = append(result, curr)

		for _, parent := range parentsOf[curr] {
			inDegree[parent]--
			if inDegree[parent] == 0 {
				processQueue = append(processQueue, parent)
				sort.Strings(processQueue)
			}
		}
	}

	return result
}

// buildDeleteWhereClause はテーブルのFKパスを辿り、usersのIDに基づくWHERE句を生成する
// usersIDCondition: users.idに対する条件式 (例: "= $1" or "IN (SELECT id FROM users WHERE email LIKE $1)")
// 戻り値: WHERE句の文字列 (例: "user_id = $1" or "entity_id IN (SELECT id FROM entities WHERE user_id = $1)")
func buildDeleteWhereClause(tableName string, edgesFrom map[string][]fkEdge, usersIDCondition string, visited map[string]bool) string {
	// usersへの直接参照を優先して探す
	for _, edge := range edgesFrom[tableName] {
		if edge.parentTable == "users" {
			return fmt.Sprintf("%s %s", edge.childColumn, usersIDCondition)
		}
	}

	// 間接パス: 中間テーブルを経由してusersへ辿る
	if visited[tableName] {
		return "" // 循環参照防止
	}
	visited[tableName] = true
	for _, edge := range edgesFrom[tableName] {
		parentWhere := buildDeleteWhereClause(edge.parentTable, edgesFrom, usersIDCondition, visited)
		if parentWhere != "" {
			return fmt.Sprintf("%s IN (SELECT %s FROM %s WHERE %s)",
				edge.childColumn, edge.parentColumn, edge.parentTable, parentWhere)
		}
	}

	return ""
}

// buildDynamicCleanupQueries はDBのFKグラフを解析し、
// userIDまたはemailパターンによるDELETE文一覧を動的に生成する
// スキーマに新しいテーブルが追加されても自動的に対応する
func buildDynamicCleanupQueries(db *sql.DB) (byUserID []string, byEmailPattern []string, err error) {
	edges, err := loadFKGraph(db)
	if err != nil {
		return nil, nil, err
	}

	order := buildUserRelatedCleanupOrder(edges)

	// edgesFrom[child] = そのテーブルが持つFK辺一覧
	edgesFrom := map[string][]fkEdge{}
	for _, e := range edges {
		edgesFrom[e.childTable] = append(edgesFrom[e.childTable], e)
	}

	for _, table := range order {
		if table == "users" {
			// usersテーブル自身は専用の条件で削除
			byUserID = append(byUserID, "DELETE FROM users WHERE id = $1")
			byEmailPattern = append(byEmailPattern, "DELETE FROM users WHERE email LIKE $1")
			continue
		}

		// userIDによる削除クエリ
		whereID := buildDeleteWhereClause(table, edgesFrom, "= $1", map[string]bool{})
		if whereID != "" {
			byUserID = append(byUserID, fmt.Sprintf("DELETE FROM %s WHERE %s", table, whereID))
		}

		// emailパターンによる削除クエリ
		whereEmail := buildDeleteWhereClause(
			table, edgesFrom,
			"IN (SELECT id FROM users WHERE email LIKE $1)",
			map[string]bool{},
		)
		if whereEmail != "" {
			byEmailPattern = append(byEmailPattern, fmt.Sprintf("DELETE FROM %s WHERE %s", table, whereEmail))
		}
	}

	return byUserID, byEmailPattern, nil
}
