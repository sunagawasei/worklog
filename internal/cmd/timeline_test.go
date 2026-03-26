package cmd

import (
	"bytes"
	"testing"
	"time"

	"worklog/internal/domain"
)

func TestHandleTimeline_DateArgument(t *testing.T) {
	t.Run("引数なしの場合は本日のデータを取得する", func(t *testing.T) {
		manager := &mockProjectManager{
			summaries: []domain.ProjectSummary{
				{Project: "ProjectA", Tag: "Development", TotalTime: time.Hour},
			},
		}
		opts := parseGlobalFlags([]string{"timeline"})
		opts.Writer = &bytes.Buffer{}

		err := handleTimeline(manager, opts)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(manager.calledMethods) != 1 {
			t.Errorf("expected 1 method call, got %d", len(manager.calledMethods))
		}
		if manager.calledMethods[0] != "List()" {
			t.Errorf("expected List() to be called, got %s", manager.calledMethods[0])
		}
	})

	t.Run("相対日付-1dが指定された場合は昨日のデータを取得する", func(t *testing.T) {
		yesterday := time.Now().AddDate(0, 0, -1)
		manager := &mockProjectManager{
			listOnDateSummaries: []domain.ProjectSummary{
				{Project: "ProjectB", Tag: "MTG", TotalTime: time.Hour * 2},
			},
		}
		opts := parseGlobalFlags([]string{"timeline", "-1d"})
		opts.Writer = &bytes.Buffer{}

		err := handleTimeline(manager, opts)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(manager.calledMethods) != 1 {
			t.Errorf("expected 1 method call, got %d", len(manager.calledMethods))
		}
		expectedPrefix := "ListOnDate("
		if len(manager.calledMethods[0]) < len(expectedPrefix) || manager.calledMethods[0][:len(expectedPrefix)] != expectedPrefix {
			t.Errorf("expected ListOnDate() to be called, got %s", manager.calledMethods[0])
		}
		_ = yesterday
	})

	t.Run("YYYY-MM-DD形式が指定された場合は指定日のデータを取得する", func(t *testing.T) {
		manager := &mockProjectManager{
			listOnDateSummaries: []domain.ProjectSummary{
				{Project: "ProjectC", Tag: "Development", TotalTime: time.Hour * 3},
			},
		}
		opts := parseGlobalFlags([]string{"timeline", "2025-01-15"})
		opts.Writer = &bytes.Buffer{}

		err := handleTimeline(manager, opts)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(manager.calledMethods) != 1 {
			t.Errorf("expected 1 method call, got %d", len(manager.calledMethods))
		}
		expectedPrefix := "ListOnDate("
		if len(manager.calledMethods[0]) < len(expectedPrefix) || manager.calledMethods[0][:len(expectedPrefix)] != expectedPrefix {
			t.Errorf("expected ListOnDate() to be called, got %s", manager.calledMethods[0])
		}
	})

	t.Run("YYYYMMDD形式が指定された場合は指定日のデータを取得する", func(t *testing.T) {
		manager := &mockProjectManager{
			listOnDateSummaries: []domain.ProjectSummary{
				{Project: "ProjectD", Tag: "Review", TotalTime: time.Hour * 4},
			},
		}
		opts := parseGlobalFlags([]string{"timeline", "20250115"})
		opts.Writer = &bytes.Buffer{}

		err := handleTimeline(manager, opts)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(manager.calledMethods) != 1 {
			t.Errorf("expected 1 method call, got %d", len(manager.calledMethods))
		}
		expectedPrefix := "ListOnDate("
		if len(manager.calledMethods[0]) < len(expectedPrefix) || manager.calledMethods[0][:len(expectedPrefix)] != expectedPrefix {
			t.Errorf("expected ListOnDate() to be called, got %s", manager.calledMethods[0])
		}
	})

	t.Run("不正な日付形式の場合はエラーを返す", func(t *testing.T) {
		manager := &mockProjectManager{}
		opts := parseGlobalFlags([]string{"timeline", "invalid-date"})
		opts.Writer = &bytes.Buffer{}

		err := handleTimeline(manager, opts)

		if err == nil {
			t.Error("expected error for invalid date format, got nil")
		}
		if len(manager.calledMethods) != 0 {
			t.Errorf("expected no method calls for invalid date, got %d calls", len(manager.calledMethods))
		}
	})
}
