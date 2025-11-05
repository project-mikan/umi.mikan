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
	// 新しいロジック: JSTの日付をUTC 00:00:00として表現
	// 昨日（JST 2025/11/3）をUTC 00:00:00として表現
	expectedPeriodEnd := time.Date(2025, 11, 3, 0, 0, 0, 0, time.UTC)
	// 3日前（JST 2025/11/1）をUTC 00:00:00として表現
	expectedPeriodStart := time.Date(2025, 11, 1, 0, 0, 0, 0, time.UTC)

	// periodEndの検証
	if !periodEnd.Equal(expectedPeriodEnd) {
		t.Errorf("periodEnd: expected %v, got %v", expectedPeriodEnd, periodEnd)
	}

	// periodStartの検証
	if !periodStart.Equal(expectedPeriodStart) {
		t.Errorf("periodStart: expected %v, got %v", expectedPeriodStart, periodStart)
	}

	// データベースに格納されている日付（JSTの日付をUTC 00:00:00として表現）と
	// 返却される期間が一致することを確認
	// これにより、date >= periodStart AND date <= periodEnd のクエリで
	// 11/1, 11/2, 11/3 の日記が正しく取得される

	todayJST := time.Date(nowJST.Year(), nowJST.Month(), nowJST.Day(), 0, 0, 0, 0, jst)

	t.Logf("実行日時（JST）: %v", nowJST)
	t.Logf("今日（JST）: %v", todayJST)
	t.Logf("期間開始（UTC 00:00:00として表現されたJST日付）: %v (= JST %s)", periodStart, periodStart.Format("2006/01/02"))
	t.Logf("期間終了（UTC 00:00:00として表現されたJST日付）: %v (= JST %s)", periodEnd, periodEnd.Format("2006/01/02"))
	t.Logf("取得される日記の日付範囲（JST）: %s から %s",
		periodStart.Format("2006/01/02"),
		periodEnd.Format("2006/01/02"))
}
