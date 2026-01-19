package ui

import (
	"strings"
	"testing"
	"time"

	"worklog/internal/domain"
)

// TestRenderStatus_Idle はアイドル状態（statusがnil）の表示をテストする
func TestRenderStatus_Idle(t *testing.T) {
	now := time.Date(2025, 9, 30, 9, 29, 0, 0, time.Local)
	todayTotal := 0 * time.Hour // アイドル時は0
	result := RenderStatus(nil, now, todayTotal)

	// 期待する要素が含まれているか確認
	if !strings.Contains(result, "稼働中のプロジェクトはありません") {
		t.Errorf("期待: '稼働中のプロジェクトはありません'を含む, 実際: %q", result)
	}

	// 区切り線が含まれているか
	if !strings.Contains(result, "─") {
		t.Errorf("期待: 区切り線'─'を含む, 実際: %q", result)
	}
}

// TestRenderStatus_Running は稼働中の表示をテストする
func TestRenderStatus_Running(t *testing.T) {
	startTime := time.Date(2025, 9, 30, 10, 0, 0, 0, time.Local)
	now := time.Date(2025, 9, 30, 11, 23, 0, 0, time.Local)

	status := &domain.ProjectStatus{
		Project:   "ProjectA",
		Tag:       "1",
		TagName:   "Development",
		StartTime: startTime,
	}

	// 今日の合計作業時間（3時間と仮定）
	todayTotal := 3 * time.Hour

	result := RenderStatus(status, now, todayTotal)

	// 期待する要素が含まれているか確認
	if !strings.Contains(result, "ProjectA") {
		t.Errorf("期待: 'ProjectA'を含む, 実際: %q", result)
	}

	if !strings.Contains(result, "running") {
		t.Errorf("期待: 'running'を含む, 実際: %q", result)
	}

	if !strings.Contains(result, "10:00") {
		t.Errorf("期待: '10:00'を含む, 実際: %q", result)
	}

	if !strings.Contains(result, "1h 23m") {
		t.Errorf("期待: '1h 23m'を含む, 実際: %q", result)
	}

	if !strings.Contains(result, "Development") {
		t.Errorf("期待: 'Development'を含む, 実際: %q", result)
	}

	// 区切り線が含まれているか
	if !strings.Contains(result, "─") {
		t.Errorf("期待: 区切り線'─'を含む, 実際: %q", result)
	}

	// プログレスバーが含まれているか
	// 1h 23m / 3h 00m = 約46%
	if !strings.Contains(result, BlockFull) {
		t.Errorf("期待: プログレスバー（%s）を含む, 実際: %q", BlockFull, result)
	}

	if !strings.Contains(result, BlockLight) {
		t.Errorf("期待: プログレスバー（%s）を含む, 実際: %q", BlockLight, result)
	}

	// パーセンテージ表示が含まれているか
	if !strings.Contains(result, "%") {
		t.Errorf("期待: パーセンテージ表示を含む, 実際: %q", result)
	}
}

// TestRenderStatus_RunningWithoutTagName はタグ名なしの稼働中表示をテストする
func TestRenderStatus_RunningWithoutTagName(t *testing.T) {
	startTime := time.Date(2025, 9, 30, 10, 0, 0, 0, time.Local)
	now := time.Date(2025, 9, 30, 10, 15, 0, 0, time.Local)

	status := &domain.ProjectStatus{
		Project:   "ProjectB",
		Tag:       "2",
		TagName:   "", // タグ名なし
		StartTime: startTime,
	}

	todayTotal := 1 * time.Hour
	result := RenderStatus(status, now, todayTotal)

	// プロジェクト名は表示される
	if !strings.Contains(result, "ProjectB") {
		t.Errorf("期待: 'ProjectB'を含む, 実際: %q", result)
	}

	// 稼働時間は表示される
	if !strings.Contains(result, "0h 15m") {
		t.Errorf("期待: '0h 15m'を含む, 実際: %q", result)
	}
}

// TestRenderList_Empty はプロジェクトが空の場合の表示をテストする
func TestRenderList_Empty(t *testing.T) {
	summaries := []domain.ProjectSummary{}
	now := time.Date(2025, 9, 30, 15, 0, 0, 0, time.Local)

	result := RenderList(summaries, now)

	// 日付ヘッダーが含まれているか
	if !strings.Contains(result, "Today") {
		t.Errorf("期待: 'Today'を含む, 実際: %q", result)
	}

	if !strings.Contains(result, "2025-09-30") {
		t.Errorf("期待: '2025-09-30'を含む, 実際: %q", result)
	}

	// 空メッセージが含まれているか
	if !strings.Contains(result, "本日の作業履歴はありません") {
		t.Errorf("期待: '本日の作業履歴はありません'を含む, 実際: %q", result)
	}
}

// TestRenderList_SingleProject は1つのプロジェクトの場合の表示をテストする
func TestRenderList_SingleProject(t *testing.T) {
	summaries := []domain.ProjectSummary{
		{
			Project:   "ProjectA",
			Tag:       "1",
			TagName:   "Development",
			TotalTime: 2*time.Hour + 30*time.Minute,
		},
	}
	now := time.Date(2025, 9, 30, 15, 0, 0, 0, time.Local)

	result := RenderList(summaries, now)

	// プロジェクト名と稼働時間が含まれているか
	if !strings.Contains(result, "ProjectA") {
		t.Errorf("期待: 'ProjectA'を含む, 実際: %q", result)
	}

	if !strings.Contains(result, "2h 30m") {
		t.Errorf("期待: '2h 30m'を含む, 実際: %q", result)
	}

	// タグ名が含まれているか
	if !strings.Contains(result, "Development") {
		t.Errorf("期待: 'Development'を含む, 実際: %q", result)
	}

	// 合計時間が含まれているか
	if !strings.Contains(result, "Total") {
		t.Errorf("期待: 'Total'を含む, 実際: %q", result)
	}

	// ボックス描画文字が使用されているか
	if !strings.Contains(result, "│") {
		t.Errorf("期待: '│'を含む, 実際: %q", result)
	}
}

// TestRenderList_MultipleProjects は複数のプロジェクトの場合の表示をテストする
func TestRenderList_MultipleProjects(t *testing.T) {
	summaries := []domain.ProjectSummary{
		{
			Project:   "ProjectA",
			Tag:       "1",
			TagName:   "Development",
			TotalTime: 2*time.Hour + 30*time.Minute,
		},
		{
			Project:   "ProjectB",
			Tag:       "2",
			TagName:   "Meeting",
			TotalTime: 1 * time.Hour,
		},
	}
	now := time.Date(2025, 9, 30, 15, 0, 0, 0, time.Local)

	result := RenderList(summaries, now)

	// 両方のプロジェクトが含まれているか
	if !strings.Contains(result, "ProjectA") {
		t.Errorf("期待: 'ProjectA'を含む, 実際: %q", result)
	}

	if !strings.Contains(result, "ProjectB") {
		t.Errorf("期待: 'ProjectB'を含む, 実際: %q", result)
	}

	// タグ名が含まれているか
	if !strings.Contains(result, "Development") {
		t.Errorf("期待: 'Development'を含む, 実際: %q", result)
	}

	if !strings.Contains(result, "Meeting") {
		t.Errorf("期待: 'Meeting'を含む, 実際: %q", result)
	}

	// 両方の稼働時間が含まれているか
	if !strings.Contains(result, "2h 30m") {
		t.Errorf("期待: '2h 30m'を含む, 実際: %q", result)
	}

	if !strings.Contains(result, "1h 00m") {
		t.Errorf("期待: '1h 00m'を含む, 実際: %q", result)
	}

	// 合計時間が含まれているか（3h 30m）
	if !strings.Contains(result, "Total") {
		t.Errorf("期待: 'Total'を含む, 実際: %q", result)
	}

	if !strings.Contains(result, "3h 30m") {
		t.Errorf("期待: '3h 30m'を含む, 実際: %q", result)
	}
}

// TestRenderList_WithTimeRanges は時間範囲を含むリスト表示をテストする
func TestRenderList_WithTimeRanges(t *testing.T) {
	baseTime := time.Date(2025, 9, 30, 10, 0, 0, 0, time.Local)
	summaries := []domain.ProjectSummary{
		{
			Project:   "ProjectA",
			Tag:       "1",
			TagName:   "Development",
			TotalTime: 2*time.Hour + 30*time.Minute,
			TimeRanges: []domain.TimeRange{
				{
					Start:    baseTime,
					End:      baseTime.Add(50 * time.Minute),
					Duration: 50 * time.Minute,
				},
				{
					Start:    baseTime.Add(3 * time.Hour),
					End:      baseTime.Add(4*time.Hour + 40*time.Minute),
					Duration: 1*time.Hour + 40*time.Minute,
				},
			},
		},
		{
			Project:   "ProjectB",
			Tag:       "2",
			TagName:   "Meeting",
			TotalTime: 1 * time.Hour,
			TimeRanges: []domain.TimeRange{
				{
					Start:    baseTime.Add(1 * time.Hour),
					End:      baseTime.Add(2 * time.Hour),
					Duration: 1 * time.Hour,
				},
			},
		},
	}
	now := baseTime.Add(8 * time.Hour)

	result := RenderList(summaries, now)

	// プロジェクト名と合計時間が含まれているか
	if !strings.Contains(result, "ProjectA") {
		t.Errorf("期待: 'ProjectA'を含む, 実際: %q", result)
	}

	if !strings.Contains(result, "2h 30m") {
		t.Errorf("期待: '2h 30m'を含む, 実際: %q", result)
	}

	// タグ名が含まれているか
	if !strings.Contains(result, "Development") {
		t.Errorf("期待: 'Development'を含む, 実際: %q", result)
	}

	if !strings.Contains(result, "Meeting") {
		t.Errorf("期待: 'Meeting'を含む, 実際: %q", result)
	}

	// 時間範囲が含まれているか
	if !strings.Contains(result, "10:00-10:50") {
		t.Errorf("期待: '10:00-10:50'を含む, 実際: %q", result)
	}

	if !strings.Contains(result, "13:00-14:40") {
		t.Errorf("期待: '13:00-14:40'を含む, 実際: %q", result)
	}

	if !strings.Contains(result, "50m") {
		t.Errorf("期待: '50m'（時間範囲の表示）を含む, 実際: %q", result)
	}

	if !strings.Contains(result, "1h 40m") {
		t.Errorf("期待: '1h 40m'（時間範囲の表示）を含む, 実際: %q", result)
	}

	// ツリー構造文字が含まれているか
	if !strings.Contains(result, "├─") {
		t.Errorf("期待: '├─'を含む, 実際: %q", result)
	}

	if !strings.Contains(result, "└─") {
		t.Errorf("期待: '└─'を含む, 実際: %q", result)
	}

	// ProjectBの時間範囲も確認
	if !strings.Contains(result, "11:00-12:00") {
		t.Errorf("期待: '11:00-12:00'を含む, 実際: %q", result)
	}
}

// TestRenderProgressBar_Zero は0%のプログレスバーをテストする
func TestRenderProgressBar_Zero(t *testing.T) {
	result := RenderProgressBar(0.0, 20)

	// BlockLight（░）のみで構成されているか
	expectedBar := strings.Repeat(BlockLight, 20)
	if !strings.Contains(result, expectedBar) {
		t.Errorf("期待: '%s'を含む, 実際: %q", expectedBar, result)
	}

	// 0%表示が含まれているか
	if !strings.Contains(result, "0%") {
		t.Errorf("期待: '0%%'を含む, 実際: %q", result)
	}
}

// TestRenderProgressBar_Half は50%のプログレスバーをテストする
func TestRenderProgressBar_Half(t *testing.T) {
	result := RenderProgressBar(0.5, 20)

	// BlockFull（█）が10文字、BlockLight（░）が10文字
	expectedFull := strings.Repeat(BlockFull, 10)
	expectedLight := strings.Repeat(BlockLight, 10)

	if !strings.Contains(result, expectedFull) {
		t.Errorf("期待: '%s'を含む, 実際: %q", expectedFull, result)
	}

	if !strings.Contains(result, expectedLight) {
		t.Errorf("期待: '%s'を含む, 実際: %q", expectedLight, result)
	}

	// 50%表示が含まれているか
	if !strings.Contains(result, "50%") {
		t.Errorf("期待: '50%%'を含む, 実際: %q", result)
	}
}

// TestRenderProgressBar_Full は100%のプログレスバーをテストする
func TestRenderProgressBar_Full(t *testing.T) {
	result := RenderProgressBar(1.0, 20)

	// BlockFull（█）のみで構成されているか
	expectedBar := strings.Repeat(BlockFull, 20)
	if !strings.Contains(result, expectedBar) {
		t.Errorf("期待: '%s'を含む, 実際: %q", expectedBar, result)
	}

	// 100%表示が含まれているか
	if !strings.Contains(result, "100%") {
		t.Errorf("期待: '100%%'を含む, 実際: %q", result)
	}
}

// TestRenderProgressBar_CustomWidth はカスタム幅のプログレスバーをテストする
func TestRenderProgressBar_CustomWidth(t *testing.T) {
	// 幅10で60%のプログレスバー
	result := RenderProgressBar(0.6, 10)

	// BlockFull（█）が6文字、BlockLight（░）が4文字
	expectedFull := strings.Repeat(BlockFull, 6)
	expectedLight := strings.Repeat(BlockLight, 4)

	if !strings.Contains(result, expectedFull) {
		t.Errorf("期待: '%s'を含む, 実際: %q", expectedFull, result)
	}

	if !strings.Contains(result, expectedLight) {
		t.Errorf("期待: '%s'を含む, 実際: %q", expectedLight, result)
	}

	// 60%表示が含まれているか
	if !strings.Contains(result, "60%") {
		t.Errorf("期待: '60%%'を含む, 実際: %q", result)
	}
}

// TestRenderProgress_Loading は進行中のプログレス表示をテストする
func TestRenderProgress_Loading(t *testing.T) {
	result := RenderProgress("Loading", 5, 10)

	// ラベルが含まれているか
	if !strings.Contains(result, "Loading") {
		t.Errorf("期待: 'Loading'を含む, 実際: %q", result)
	}

	// プログレスバーが含まれているか（50%）
	if !strings.Contains(result, BlockFull) {
		t.Errorf("期待: プログレスバー（%s）を含む, 実際: %q", BlockFull, result)
	}

	if !strings.Contains(result, BlockLight) {
		t.Errorf("期待: プログレスバー（%s）を含む, 実際: %q", BlockLight, result)
	}

	// 50%表示が含まれているか
	if !strings.Contains(result, "50%") {
		t.Errorf("期待: '50%%'を含む, 実際: %q", result)
	}
}

// TestRenderProgress_Complete は完了状態のプログレス表示をテストする
func TestRenderProgress_Complete(t *testing.T) {
	result := RenderProgress("Complete", 10, 10)

	// ラベルが含まれているか
	if !strings.Contains(result, "Complete") {
		t.Errorf("期待: 'Complete'を含む, 実際: %q", result)
	}

	// 100%プログレスバーが含まれているか
	if !strings.Contains(result, strings.Repeat(BlockFull, 16)) {
		t.Errorf("期待: 100%%プログレスバーを含む, 実際: %q", result)
	}

	// 100%表示が含まれているか
	if !strings.Contains(result, "100%") {
		t.Errorf("期待: '100%%'を含む, 実際: %q", result)
	}
}

// TestRenderProgress_LongLabel は長いラベルのプログレス表示をテストする
func TestRenderProgress_LongLabel(t *testing.T) {
	result := RenderProgress("Exporting Data", 3, 10)

	// ラベルが含まれているか
	if !strings.Contains(result, "Exporting Data") {
		t.Errorf("期待: 'Exporting Data'を含む, 実際: %q", result)
	}

	// プログレスバーが含まれているか（30%）
	if !strings.Contains(result, "30%") {
		t.Errorf("期待: '30%%'を含む, 実際: %q", result)
	}
}

// TestRenderStopMessage は停止メッセージの表示をテストする
func TestRenderStopMessage(t *testing.T) {
	startTime := time.Date(2025, 9, 30, 10, 0, 0, 0, time.Local)
	stopTime := time.Date(2025, 9, 30, 12, 30, 0, 0, time.Local)
	projectName := "ProjectA"

	result := RenderStopMessage(projectName, startTime, stopTime)

	// プロジェクト名が含まれているか
	if !strings.Contains(result, "ProjectA") {
		t.Errorf("期待: 'ProjectA'を含む, 実際: %q", result)
	}

	// "stopped"が含まれているか
	if !strings.Contains(result, "stopped") {
		t.Errorf("期待: 'stopped'を含む, 実際: %q", result)
	}

	// 時間範囲が含まれているか（10:00-12:30）
	if !strings.Contains(result, "10:00-12:30") {
		t.Errorf("期待: '10:00-12:30'を含む, 実際: %q", result)
	}

	// 経過時間が含まれているか（2h 30m）
	if !strings.Contains(result, "2h 30m") {
		t.Errorf("期待: '2h 30m'を含む, 実際: %q", result)
	}

	// 区切り線が含まれているか
	if !strings.Contains(result, "─") {
		t.Errorf("期待: 区切り線'─'を含む, 実際: %q", result)
	}
}

// TestRenderNewMessage は新規プロジェクト開始メッセージの表示をテストする
func TestRenderNewMessage(t *testing.T) {
	startTime := time.Date(2025, 9, 30, 10, 0, 0, 0, time.Local)
	projectName := "Development Work"
	tag := "Development (1)"

	result := RenderNewMessage(projectName, startTime, tag)

	// プロジェクト名が含まれているか
	if !strings.Contains(result, "Development Work") {
		t.Errorf("期待: 'Development Work'を含む, 実際: %q", result)
	}

	// "started"が含まれているか（今回の修正で"running"から変更）
	if !strings.Contains(result, "started") {
		t.Errorf("期待: 'started'を含む, 実際: %q", result)
	}

	// 開始時刻が含まれているか（10:00）
	if !strings.Contains(result, "10:00") {
		t.Errorf("期待: '10:00'を含む, 実際: %q", result)
	}

	// タグが含まれているか
	if !strings.Contains(result, "Development (1)") {
		t.Errorf("期待: 'Development (1)'を含む, 実際: %q", result)
	}

	// 区切り線が含まれているか
	if !strings.Contains(result, "─") {
		t.Errorf("期待: 区切り線'─'を含む, 実際: %q", result)
	}
}

// TestRenderSwitchMessage はプロジェクト切り替えメッセージの表示をテストする
func TestRenderSwitchMessage(t *testing.T) {
	t.Run("旧プロジェクトあり", func(t *testing.T) {
		oldStartTime := time.Date(2025, 9, 30, 10, 0, 0, 0, time.Local)
		switchTime := time.Date(2025, 9, 30, 12, 0, 0, 0, time.Local)
		oldProject := "ProjectA"
		newProject := "ProjectB"
		newTag := "Meeting (2)"

		result := RenderSwitchMessage(oldProject, oldStartTime, switchTime, newProject, newTag)

		// issue #5形式: "ProjectA → stopped"
		if !strings.Contains(result, "ProjectA → stopped") {
			t.Errorf("期待: 'ProjectA → stopped'を含む, 実際: %q", result)
		}

		// issue #5形式: "ProjectB → running"
		if !strings.Contains(result, "ProjectB → running") {
			t.Errorf("期待: 'ProjectB → running'を含む, 実際: %q", result)
		}

		// 切り替え時刻が含まれているか（12:00）
		if !strings.Contains(result, "12:00") {
			t.Errorf("期待: '12:00'を含む, 実際: %q", result)
		}

		// タグが含まれているか
		if !strings.Contains(result, "Meeting (2)") {
			t.Errorf("期待: 'Meeting (2)'を含む, 実際: %q", result)
		}

		// 時間範囲・経過時間は表示されない（issue #5コンパクト形式）
		if strings.Contains(result, "10:00-12:00") {
			t.Errorf("期待: 時間範囲'10:00-12:00'を含まない, 実際: %q", result)
		}

		if strings.Contains(result, "2h 00m") {
			t.Errorf("期待: 経過時間'2h 00m'を含まない, 実際: %q", result)
		}

		// 区切り線が含まれているか
		if !strings.Contains(result, "─") {
			t.Errorf("期待: 区切り線'─'を含む, 実際: %q", result)
		}
	})

	t.Run("旧プロジェクトなし（初回開始）", func(t *testing.T) {
		switchTime := time.Date(2025, 9, 30, 10, 0, 0, 0, time.Local)
		oldProject := ""
		oldStartTime := time.Time{} // ゼロ値
		newProject := "ProjectA"
		newTag := "Development (1)"

		result := RenderSwitchMessage(oldProject, oldStartTime, switchTime, newProject, newTag)

		// issue #5形式: "ProjectA → running"
		if !strings.Contains(result, "ProjectA → running") {
			t.Errorf("期待: 'ProjectA → running'を含む, 実際: %q", result)
		}

		// 区切り線が含まれているか
		if !strings.Contains(result, "─") {
			t.Errorf("期待: 区切り線'─'を含む, 実際: %q", result)
		}
	})
}

// TestRenderHelp はヘルプメッセージの表示をテストする
func TestRenderHelp(t *testing.T) {
	result := RenderHelp()

	// タイトルが含まれているか
	if !strings.Contains(result, "worklog") {
		t.Errorf("期待: 'worklog'を含む, 実際: %q", result)
	}

	// 各コマンドが含まれているか
	commands := []string{"new", "switch", "status", "stop", "list", "help"}
	for _, cmd := range commands {
		if !strings.Contains(result, cmd) {
			t.Errorf("期待: '%s'を含む, 実際: %q", cmd, result)
		}
	}

	// Usage が含まれているか
	if !strings.Contains(result, "Usage") {
		t.Errorf("期待: 'Usage'を含む, 実際: %q", result)
	}

	// Commands が含まれているか
	if !strings.Contains(result, "Commands") {
		t.Errorf("期待: 'Commands'を含む, 実際: %q", result)
	}

	// 区切り線が含まれているか
	if !strings.Contains(result, "─") {
		t.Errorf("期待: 区切り線'─'を含む, 実際: %q", result)
	}
}

// TestRenderError はエラーメッセージの表示をテストする
func TestRenderError(t *testing.T) {
	errorMsg := "不明なコマンド: xyz"
	result := RenderError(errorMsg)

	// エラーメッセージが含まれているか
	if !strings.Contains(result, errorMsg) {
		t.Errorf("期待: '%s'を含む, 実際: %q", errorMsg, result)
	}

	// ヘルプへの誘導が含まれているか
	if !strings.Contains(result, "worklog help") {
		t.Errorf("期待: 'worklog help'を含む, 実際: %q", result)
	}

	// 区切り線が含まれているか
	if !strings.Contains(result, "─") {
		t.Errorf("期待: 区切り線'─'を含む, 実際: %q", result)
	}
}

// TestRenderDashboard_Idle はアイドル状態のダッシュボード表示をテストする
func TestRenderDashboard_Idle(t *testing.T) {
	now := time.Date(2025, 9, 30, 15, 0, 0, 0, time.Local)
	summaries := []domain.ProjectSummary{
		{
			Project:   "ProjectA",
			TotalTime: 2 * time.Hour,
		},
		{
			Project:   "ProjectB",
			TotalTime: 1 * time.Hour,
		},
	}

	result := RenderDashboard(nil, summaries, now)

	// 2カラムレイアウトの区切り文字が含まれているか
	if !strings.Contains(result, "┬") {
		t.Errorf("期待: 2カラム区切り'┬'を含む, 実際: %q", result)
	}

	if !strings.Contains(result, "┴") {
		t.Errorf("期待: 2カラム区切り'┴'を含む, 実際: %q", result)
	}

	// Status部分：アイドル状態メッセージ
	if !strings.Contains(result, "稼働中のプロジェクトは") {
		t.Errorf("期待: 'アイドル状態メッセージ（前半）'を含む, 実際: %q", result)
	}
	if !strings.Contains(result, "ありません") {
		t.Errorf("期待: 'アイドル状態メッセージ（後半）'を含む, 実際: %q", result)
	}

	// Summary部分：Today（合計時間）
	if !strings.Contains(result, "Today") {
		t.Errorf("期待: 'Today'を含む, 実際: %q", result)
	}

	if !strings.Contains(result, "3h 00m") {
		t.Errorf("期待: '3h 00m'（合計時間）を含む, 実際: %q", result)
	}

	// Summary部分：Projects（プロジェクト数）
	if !strings.Contains(result, "Projects") {
		t.Errorf("期待: 'Projects'を含む, 実際: %q", result)
	}

	if !strings.Contains(result, "2") {
		t.Errorf("期待: '2'（プロジェクト数）を含む, 実際: %q", result)
	}

	// Summary部分：Average（平均作業時間）
	if !strings.Contains(result, "Average") {
		t.Errorf("期待: 'Average'を含む, 実際: %q", result)
	}

	if !strings.Contains(result, "1h 30m") {
		t.Errorf("期待: '1h 30m'（平均時間）を含む, 実際: %q", result)
	}
}

// TestRenderDashboard_Running は稼働中のダッシュボード表示をテストする
func TestRenderDashboard_Running(t *testing.T) {
	startTime := time.Date(2025, 9, 30, 10, 0, 0, 0, time.Local)
	now := time.Date(2025, 9, 30, 11, 30, 0, 0, time.Local)

	status := &domain.ProjectStatus{
		Project:            "ProjectA",
		Tag:                "1",
		TagName:            "Development",
		StartTime:          startTime,
		CurrentSessionTime: 1*time.Hour + 30*time.Minute, // 現在セッションの経過時間（11:30 - 10:00）
		TotalTime:          4*time.Hour + 15*time.Minute, // 本日の累計稼働時間
	}

	summaries := []domain.ProjectSummary{
		{
			Project:   "ProjectA",
			TotalTime: 4*time.Hour + 15*time.Minute,
		},
		{
			Project:   "ProjectB",
			TotalTime: 1*time.Hour + 30*time.Minute,
		},
		{
			Project:   "ProjectC",
			TotalTime: 2 * time.Hour,
		},
	}

	result := RenderDashboard(status, summaries, now)

	// 2カラムレイアウトの区切り文字が含まれているか
	if !strings.Contains(result, "┬") {
		t.Errorf("期待: 2カラム区切り'┬'を含む, 実際: %q", result)
	}

	// Status部分：プロジェクト名と稼働状態
	if !strings.Contains(result, "ProjectA") {
		t.Errorf("期待: 'ProjectA'を含む, 実際: %q", result)
	}

	if !strings.Contains(result, "running") {
		t.Errorf("期待: 'running'を含む, 実際: %q", result)
	}

	// Status部分：開始時刻と時間表示（新形式：現在セッション時間 / 累計時間）
	if !strings.Contains(result, "10:00") {
		t.Errorf("期待: '10:00'を含む, 実際: %q", result)
	}

	// 新しい表示形式：「1h 30m / 4h 15m」（現在セッション時間 / 累計時間）
	if !strings.Contains(result, "1h 30m / 4h 15m") {
		t.Errorf("期待: '1h 30m / 4h 15m'（現在セッション時間 / 累計時間）を含む, 実際: %q", result)
	}

	// Status部分：タグ名
	if !strings.Contains(result, "Development") {
		t.Errorf("期待: 'Development'を含む, 実際: %q", result)
	}

	// Summary部分：Today（合計時間）
	if !strings.Contains(result, "7h 45m") {
		t.Errorf("期待: '7h 45m'（合計時間）を含む, 実際: %q", result)
	}

	// Summary部分：Projects（プロジェクト数）
	if !strings.Contains(result, "3") {
		t.Errorf("期待: '3'（プロジェクト数）を含む, 実際: %q", result)
	}

	// Summary部分：Average（平均作業時間）
	// 7h 45m / 3 = 2h 35m
	if !strings.Contains(result, "2h 35m") {
		t.Errorf("期待: '2h 35m'（平均時間）を含む, 実際: %q", result)
	}

	// Status部分：プログレスバー
	if !strings.Contains(result, BlockFull) {
		t.Errorf("期待: プログレスバー（%s）を含む, 実際: %q", BlockFull, result)
	}
}

// TestRenderDashboard_Running_ShortFormat は現在セッション時間が1時間未満の場合の短縮形式をテストする
func TestRenderDashboard_Running_ShortFormat(t *testing.T) {
	startTime := time.Date(2025, 9, 30, 10, 0, 0, 0, time.Local)
	now := time.Date(2025, 9, 30, 10, 38, 0, 0, time.Local)

	status := &domain.ProjectStatus{
		Project:            "ProjectA",
		Tag:                "1",
		TagName:            "Development",
		StartTime:          startTime,
		CurrentSessionTime: 38 * time.Minute,             // 現在セッション時間（38分）
		TotalTime:          1*time.Hour + 20*time.Minute, // 本日の累計稼働時間
	}

	summaries := []domain.ProjectSummary{
		{
			Project:   "ProjectA",
			TotalTime: 1*time.Hour + 20*time.Minute,
		},
	}

	result := RenderDashboard(status, summaries, now)

	// 新しい短縮形式：「38m / 1h 20m」（現在セッション時間が1時間未満）
	if !strings.Contains(result, "38m / 1h 20m") {
		t.Errorf("期待: '38m / 1h 20m'（短縮形式）を含む, 実際: %q", result)
	}
}

// TestRenderDashboard_NoProjects はプロジェクトが存在しない場合のダッシュボード表示をテストする
func TestRenderDashboard_NoProjects(t *testing.T) {
	now := time.Date(2025, 9, 30, 15, 0, 0, 0, time.Local)
	summaries := []domain.ProjectSummary{}

	result := RenderDashboard(nil, summaries, now)

	// 2カラムレイアウトの区切り文字が含まれているか
	if !strings.Contains(result, "┬") {
		t.Errorf("期待: 2カラム区切り'┬'を含む, 実際: %q", result)
	}

	// Status部分：アイドル状態メッセージ
	if !strings.Contains(result, "稼働中のプロジェクトは") {
		t.Errorf("期待: 'アイドル状態メッセージ（前半）'を含む, 実際: %q", result)
	}
	if !strings.Contains(result, "ありません") {
		t.Errorf("期待: 'アイドル状態メッセージ（後半）'を含む, 実際: %q", result)
	}

	// Summary部分：Today（0時間）
	if !strings.Contains(result, "0h 00m") {
		t.Errorf("期待: '0h 00m'（合計時間）を含む, 実際: %q", result)
	}

	// Summary部分：Projects（0プロジェクト）
	if !strings.Contains(result, "0") {
		t.Errorf("期待: '0'（プロジェクト数）を含む, 実際: %q", result)
	}

	// Summary部分：Average（0時間）
	if !strings.Contains(result, "0h 00m") {
		t.Errorf("期待: '0h 00m'（平均時間）を含む, 実際: %q", result)
	}
}

// TestRenderTimeline_Empty はプロジェクトが空の場合のタイムライン表示をテストする
func TestRenderTimeline_Empty(t *testing.T) {
	summaries := []domain.ProjectSummary{}
	now := time.Date(2025, 9, 30, 15, 0, 0, 0, time.Local)

	result := RenderTimeline(summaries, now)

	// タイトルが含まれているか
	if !strings.Contains(result, "Timeline") {
		t.Errorf("期待: 'Timeline'を含む, 実際: %q", result)
	}

	// 区切り線が含まれているか
	if !strings.Contains(result, "─") {
		t.Errorf("期待: 区切り線'─'を含む, 実際: %q", result)
	}

	// アイドル状態のブロック文字が含まれているか
	if !strings.Contains(result, MiddleDot) {
		t.Errorf("期待: アイドル状態ブロック（%s）を含む, 実際: %q", MiddleDot, result)
	}

	// 時間軸（9:00-20:00）が含まれているか
	if !strings.Contains(result, "09:00") {
		t.Errorf("期待: '09:00'を含む, 実際: %q", result)
	}

	if !strings.Contains(result, "20:00") {
		t.Errorf("期待: '20:00'を含む, 実際: %q", result)
	}
}

// TestRenderTimeline_SingleProject は1つのプロジェクトのタイムライン表示をテストする
func TestRenderTimeline_SingleProject(t *testing.T) {
	baseTime := time.Date(2025, 9, 30, 10, 0, 0, 0, time.Local)
	summaries := []domain.ProjectSummary{
		{
			Project:   "ProjectA",
			TotalTime: 2 * time.Hour,
			TimeRanges: []domain.TimeRange{
				{
					Start:    baseTime,
					End:      baseTime.Add(2 * time.Hour),
					Duration: 2 * time.Hour,
				},
			},
		},
	}
	now := baseTime.Add(8 * time.Hour)

	result := RenderTimeline(summaries, now)

	// タイトルが含まれているか
	if !strings.Contains(result, "Timeline") {
		t.Errorf("期待: 'Timeline'を含む, 実際: %q", result)
	}

	// 作業ブロック文字が含まれているか
	if !strings.Contains(result, BlockSquare) {
		t.Errorf("期待: 作業ブロック（%s）を含む, 実際: %q", BlockSquare, result)
	}

	// プロジェクト名が凡例に含まれているか
	if !strings.Contains(result, "ProjectA") {
		t.Errorf("期待: 'ProjectA'を含む, 実際: %q", result)
	}

	// 時間軸が含まれているか
	if !strings.Contains(result, "10:00") {
		t.Errorf("期待: '10:00'を含む, 実際: %q", result)
	}
}

// TestRenderTimeline_MultipleProjects は複数プロジェクトのタイムライン表示をテストする
func TestRenderTimeline_MultipleProjects(t *testing.T) {
	baseTime := time.Date(2025, 9, 30, 10, 0, 0, 0, time.Local)
	summaries := []domain.ProjectSummary{
		{
			Project:   "ProjectA",
			TotalTime: 2 * time.Hour,
			TimeRanges: []domain.TimeRange{
				{
					Start:    baseTime,
					End:      baseTime.Add(2 * time.Hour),
					Duration: 2 * time.Hour,
				},
			},
		},
		{
			Project:   "ProjectB",
			TotalTime: 1 * time.Hour,
			TimeRanges: []domain.TimeRange{
				{
					Start:    baseTime.Add(2 * time.Hour),
					End:      baseTime.Add(3 * time.Hour),
					Duration: 1 * time.Hour,
				},
			},
		},
	}
	now := baseTime.Add(8 * time.Hour)

	result := RenderTimeline(summaries, now)

	// 複数の異なるブロック文字が使用されているか
	if !strings.Contains(result, BlockSquare) {
		t.Errorf("期待: 第1プロジェクトブロック（%s）を含む, 実際: %q", BlockSquare, result)
	}

	if !strings.Contains(result, BlockCrosshatch) {
		t.Errorf("期待: 第2プロジェクトブロック（%s）を含む, 実際: %q", BlockCrosshatch, result)
	}

	// 両プロジェクトが凡例に含まれているか
	if !strings.Contains(result, "ProjectA") {
		t.Errorf("期待: 'ProjectA'を含む, 実際: %q", result)
	}

	if !strings.Contains(result, "ProjectB") {
		t.Errorf("期待: 'ProjectB'を含む, 実際: %q", result)
	}
}

// TestFormatDurationShort は短縮形式の時間フォーマットをテストする
func TestFormatDurationShort(t *testing.T) {
	t.Run("0分", func(t *testing.T) {
		result := FormatDurationShort(0)
		if result != "0m" {
			t.Errorf("期待: '0m', 実際: %q", result)
		}
	})

	t.Run("1時間未満（分のみ）", func(t *testing.T) {
		result := FormatDurationShort(38 * time.Minute)
		if result != "38m" {
			t.Errorf("期待: '38m', 実際: %q", result)
		}
	})

	t.Run("ちょうど1時間", func(t *testing.T) {
		result := FormatDurationShort(1 * time.Hour)
		if result != "1h 00m" {
			t.Errorf("期待: '1h 00m', 実際: %q", result)
		}
	})

	t.Run("1時間以上", func(t *testing.T) {
		result := FormatDurationShort(1*time.Hour + 20*time.Minute)
		if result != "1h 20m" {
			t.Errorf("期待: '1h 20m', 実際: %q", result)
		}
	})

	t.Run("複数時間", func(t *testing.T) {
		result := FormatDurationShort(3*time.Hour + 45*time.Minute)
		if result != "3h 45m" {
			t.Errorf("期待: '3h 45m', 実際: %q", result)
		}
	})
}
