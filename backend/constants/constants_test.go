package constants

import "testing"

func TestMinDiaryEntriesForTrend(t *testing.T) {
	// トレンド分析に必要な最小日記エントリ数は3日分
	expectedValue := 3
	if MinDiaryEntriesForTrend != expectedValue {
		t.Errorf("expected MinDiaryEntriesForTrend to be %d, got %d", expectedValue, MinDiaryEntriesForTrend)
	}
}
