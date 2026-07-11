package mcpserver

// このファイルはテストビルド専用。
// notifyLastUsedUpdate をテスト用のフック（afterLastUsedUpdate）で上書きすることで
// auth.go の goroutine 完了を待機できるようにする。
// 本番バイナリには含まれないため、本番コードを汚染しない。

// afterLastUsedUpdate はテストケースが注入するコールバック。
// 並列テスト（t.Parallel()）での競合を防ぐため、各テストは書き換え前の値を
// defer で復元すること（callWithTokenAndWaitLastUsedUpdate がこれを担う）。
var afterLastUsedUpdate func(err error)

func init() {
	notifyLastUsedUpdate = func(err error) {
		if afterLastUsedUpdate != nil {
			afterLastUsedUpdate(err)
		}
	}
}
