package main

import (
	"testing"
	"time"
)

func TestDailySummaryJob(t *testing.T) {
	interval := 10 * time.Minute
	job := NewDailySummaryJob(interval)

	if job.Name() != "DailySummaryGeneration" {
		t.Errorf("expected job name 'DailySummaryGeneration', got '%s'", job.Name())
	}

	if job.Interval() != interval {
		t.Errorf("expected interval %v, got %v", interval, job.Interval())
	}
}

func TestMonthlySummaryJob(t *testing.T) {
	interval := 30 * time.Minute
	job := NewMonthlySummaryJob(interval)

	if job.Name() != "MonthlySummaryGeneration" {
		t.Errorf("expected job name 'MonthlySummaryGeneration', got '%s'", job.Name())
	}

	if job.Interval() != interval {
		t.Errorf("expected interval %v, got %v", interval, job.Interval())
	}
}

func TestLatestTrendJob(t *testing.T) {
	targetHour := 4
	targetMinute := 30
	job := NewLatestTrendJob(targetHour, targetMinute)

	if job.Name() != "LatestTrendGeneration" {
		t.Errorf("expected job name 'LatestTrendGeneration', got '%s'", job.Name())
	}

	// TargetHourが正しく設定されているか確認
	if job.TargetHour() != targetHour {
		t.Errorf("expected targetHour %d, got %d", targetHour, job.TargetHour())
	}

	// TargetMinuteが正しく設定されているか確認
	if job.TargetMinute() != targetMinute {
		t.Errorf("expected targetMinute %d, got %d", targetMinute, job.TargetMinute())
	}

	// DailyScheduledJobインターフェースを実装しているか確認
	var _ DailyScheduledJob = job
}

// TestCalculateTrendPeriod は、2025/11/4 4:00 JST の実行で
// 11/1, 11/2, 11/3 の日記が取得されることを確認するテスト
func TestCalculateTrendPeriod(t *testing.T) {
	// 2025/11/4 4:00 JST
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		jst = time.FixedZone("Asia/Tokyo", 9*60*60)
	}
	nowJST := time.Date(2025, 11, 4, 4, 0, 0, 0, jst)

	// 実際の関数を呼び出す
	periodStart, periodEnd := calculateTrendPeriod(nowJST)

	// 期待値の計算
	// 2025/11/3 00:00:00 JST = 2025/11/2 15:00:00 UTC (昨日)
	expectedPeriodEnd := time.Date(2025, 11, 2, 15, 0, 0, 0, time.UTC)
	// 2025/11/1 00:00:00 JST = 2025/10/31 15:00:00 UTC (3日前)
	expectedPeriodStart := time.Date(2025, 10, 31, 15, 0, 0, 0, time.UTC)

	// periodEndの検証
	if !periodEnd.Equal(expectedPeriodEnd) {
		t.Errorf("periodEnd: expected %v, got %v", expectedPeriodEnd, periodEnd)
	}

	// periodStartの検証
	if !periodStart.Equal(expectedPeriodStart) {
		t.Errorf("periodStart: expected %v, got %v", expectedPeriodStart, periodStart)
	}

	// 日本時間での日付を確認
	periodEndJST := periodEnd.In(jst)
	periodStartJST := periodStart.In(jst)

	// periodEndは日本時間で11/3の00:00:00であるべき
	expectedPeriodEndJST := time.Date(2025, 11, 3, 0, 0, 0, 0, jst)
	if !periodEndJST.Equal(expectedPeriodEndJST) {
		t.Errorf("periodEndJST: expected %v, got %v", expectedPeriodEndJST, periodEndJST)
	}

	// periodStartは日本時間で11/1の00:00:00であるべき
	expectedPeriodStartJST := time.Date(2025, 11, 1, 0, 0, 0, 0, jst)
	if !periodStartJST.Equal(expectedPeriodStartJST) {
		t.Errorf("periodStartJST: expected %v, got %v", expectedPeriodStartJST, periodStartJST)
	}

	// 取得される日記の日付範囲を確認
	// データベースのクエリは date >= periodStart AND date <= periodEnd
	// これにより、11/1, 11/2, 11/3 の日記が取得される

	todayJST := time.Date(nowJST.Year(), nowJST.Month(), nowJST.Day(), 0, 0, 0, 0, jst)
	todayUTC := todayJST.UTC()

	t.Logf("実行日時（JST）: %v", nowJST)
	t.Logf("今日（JST）: %v", todayJST)
	t.Logf("今日（UTC）: %v", todayUTC)
	t.Logf("期間開始（UTC）: %v (JST: %v)", periodStart, periodStartJST)
	t.Logf("期間終了（UTC）: %v (JST: %v)", periodEnd, periodEndJST)
	t.Logf("取得される日記の日付範囲（JST）: %s から %s",
		periodStartJST.Format("2006/01/02"),
		periodEndJST.Format("2006/01/02"))
}
