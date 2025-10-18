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
	interval := 24 * time.Hour
	job := NewLatestTrendJob(interval)

	if job.Name() != "LatestTrendGeneration" {
		t.Errorf("expected job name 'LatestTrendGeneration', got '%s'", job.Name())
	}

	if job.Interval() != interval {
		t.Errorf("expected interval %v, got %v", interval, job.Interval())
	}
}
