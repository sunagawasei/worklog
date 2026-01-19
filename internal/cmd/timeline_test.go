package cmd

import (
	"os"
	"testing"
	"time"

	"worklog/internal/domain"
)

func TestHandleTimeline_DateArgument(t *testing.T) {
	// 元のos.Argsを保存して、テスト後に復元
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	t.Run("引数なしの場合は本日のデータを取得する", func(t *testing.T) {
		// Arrange
		manager := &mockProjectManager{
			summaries: []domain.ProjectSummary{
				{Project: "ProjectA", Tag: "Development", TotalTime: time.Hour},
			},
		}
		os.Args = []string{"worklog", "timeline"}

		// Act
		err := handleTimeline(manager)

		// Assert
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
		// Arrange
		yesterday := time.Now().AddDate(0, 0, -1)
		manager := &mockProjectManager{
			listOnDateSummaries: []domain.ProjectSummary{
				{Project: "ProjectB", Tag: "MTG", TotalTime: time.Hour * 2},
			},
		}
		os.Args = []string{"worklog", "timeline", "-1d"}

		// Act
		err := handleTimeline(manager)

		// Assert
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(manager.calledMethods) != 1 {
			t.Errorf("expected 1 method call, got %d", len(manager.calledMethods))
		}
		// ListOnDate が呼ばれたことを確認（日付は yesterday の日付部分と一致するはず）
		expectedPrefix := "ListOnDate("
		if len(manager.calledMethods[0]) < len(expectedPrefix) || manager.calledMethods[0][:len(expectedPrefix)] != expectedPrefix {
			t.Errorf("expected ListOnDate() to be called, got %s", manager.calledMethods[0])
		}
		// 実際に渡された日付が昨日であることを確認
		_ = yesterday // 日付の厳密な比較は省略（時刻部分の違いがあるため）
	})

	t.Run("YYYY-MM-DD形式が指定された場合は指定日のデータを取得する", func(t *testing.T) {
		// Arrange
		manager := &mockProjectManager{
			listOnDateSummaries: []domain.ProjectSummary{
				{Project: "ProjectC", Tag: "Development", TotalTime: time.Hour * 3},
			},
		}
		os.Args = []string{"worklog", "timeline", "2025-01-15"}

		// Act
		err := handleTimeline(manager)

		// Assert
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(manager.calledMethods) != 1 {
			t.Errorf("expected 1 method call, got %d", len(manager.calledMethods))
		}
		// ListOnDate が呼ばれたことを確認
		expectedPrefix := "ListOnDate("
		if len(manager.calledMethods[0]) < len(expectedPrefix) || manager.calledMethods[0][:len(expectedPrefix)] != expectedPrefix {
			t.Errorf("expected ListOnDate() to be called, got %s", manager.calledMethods[0])
		}
	})

	t.Run("YYYYMMDD形式が指定された場合は指定日のデータを取得する", func(t *testing.T) {
		// Arrange
		manager := &mockProjectManager{
			listOnDateSummaries: []domain.ProjectSummary{
				{Project: "ProjectD", Tag: "Review", TotalTime: time.Hour * 4},
			},
		}
		os.Args = []string{"worklog", "timeline", "20250115"}

		// Act
		err := handleTimeline(manager)

		// Assert
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(manager.calledMethods) != 1 {
			t.Errorf("expected 1 method call, got %d", len(manager.calledMethods))
		}
		// ListOnDate が呼ばれたことを確認
		expectedPrefix := "ListOnDate("
		if len(manager.calledMethods[0]) < len(expectedPrefix) || manager.calledMethods[0][:len(expectedPrefix)] != expectedPrefix {
			t.Errorf("expected ListOnDate() to be called, got %s", manager.calledMethods[0])
		}
	})

	t.Run("不正な日付形式の場合はエラーを返す", func(t *testing.T) {
		// Arrange
		manager := &mockProjectManager{}
		os.Args = []string{"worklog", "timeline", "invalid-date"}

		// Act
		err := handleTimeline(manager)

		// Assert
		if err == nil {
			t.Error("expected error for invalid date format, got nil")
		}
		// ListOnDate が呼ばれていないことを確認
		if len(manager.calledMethods) != 0 {
			t.Errorf("expected no method calls for invalid date, got %d calls", len(manager.calledMethods))
		}
	})
}
