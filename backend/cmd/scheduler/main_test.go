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
