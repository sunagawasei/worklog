package ui

import (
	"strings"
	"testing"
	"time"

	"worklog/internal/domain"
)


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

		// 停止したプロジェクト名
		if !strings.Contains(result, "ProjectA → stopped") {
			t.Errorf("期待: 'ProjectA → stopped'を含む, 実際: %q", result)
		}

		// 時間範囲が含まれているか（10:00-12:00）
		if !strings.Contains(result, "10:00-12:00") {
			t.Errorf("期待: '10:00-12:00'を含む, 実際: %q", result)
		}

		// 経過時間が含まれているか（2h 00m）
		if !strings.Contains(result, "2h 00m") {
			t.Errorf("期待: '2h 00m'を含む, 実際: %q", result)
		}

		// 開始したプロジェクト
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

		// 開始したプロジェクト
		if !strings.Contains(result, "ProjectA → running") {
			t.Errorf("期待: 'ProjectA → running'を含む, 実際: %q", result)
		}

		// 区切り線が含まれているか
		if !strings.Contains(result, "─") {
			t.Errorf("期待: 区切り線'─'を含む, 実際: %q", result)
		}

		// 旧プロジェクトがない場合は経過時間表示なし
		if strings.Contains(result, "10:00-") {
			t.Errorf("期待: 時間範囲を含まない, 実際: %q", result)
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

	// 丸角ボックス文字が含まれているか
	if !strings.Contains(result, RoundTL) {
		t.Errorf("期待: 丸角ボックス %q を含む, 実際: %q", RoundTL, result)
	}
	if !strings.Contains(result, RoundBR) {
		t.Errorf("期待: 丸角ボックス %q を含む, 実際: %q", RoundBR, result)
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

// === レスポンシブUI テスト ===

func TestContentWidthFor(t *testing.T) {
	t.Run("最小幅保証", func(t *testing.T) {
		if cw := contentWidthFor(20); cw != 44 {
			t.Errorf("期待: 44, 実際: %d", cw)
		}
	})

	t.Run("範囲内はそのまま", func(t *testing.T) {
		if cw := contentWidthFor(60); cw != 60 {
			t.Errorf("期待: 60, 実際: %d", cw)
		}
	})

	t.Run("最大幅制限", func(t *testing.T) {
		if cw := contentWidthFor(120); cw != 80 {
			t.Errorf("期待: 80, 実際: %d", cw)
		}
	})

	t.Run("境界値44", func(t *testing.T) {
		if cw := contentWidthFor(44); cw != 44 {
			t.Errorf("期待: 44, 実際: %d", cw)
		}
	})

	t.Run("境界値80", func(t *testing.T) {
		if cw := contentWidthFor(80); cw != 80 {
			t.Errorf("期待: 80, 実際: %d", cw)
		}
	})
}

func TestDashColumns(t *testing.T) {
	t.Run("最小幅53で現在と同じ", func(t *testing.T) {
		s, sm := dashColumns(53)
		if s != 24 || sm != 22 {
			t.Errorf("期待: status=24, summary=22, 実際: status=%d, summary=%d", s, sm)
		}
	})

	t.Run("幅69でStatus上限", func(t *testing.T) {
		s, sm := dashColumns(69)
		if s != 40 || sm != 22 {
			t.Errorf("期待: status=40, summary=22, 実際: status=%d, summary=%d", s, sm)
		}
	})

	t.Run("幅80で余剰はSummaryへ", func(t *testing.T) {
		s, sm := dashColumns(80)
		if s != 40 || sm != 33 {
			t.Errorf("期待: status=40, summary=33, 実際: status=%d, summary=%d", s, sm)
		}
	})

	t.Run("不変条件: statusCol+summaryCol+7==cw", func(t *testing.T) {
		for cw := 53; cw <= 80; cw++ {
			s, sm := dashColumns(cw)
			if s+sm+dashFrameWidth != cw {
				t.Errorf("cw=%d: status(%d)+summary(%d)+7=%d != %d", cw, s, sm, s+sm+7, cw)
			}
		}
	})
}

func TestRenderDashboard_SingleColumn(t *testing.T) {
	now := time.Date(2025, 9, 30, 15, 0, 0, 0, time.Local)
	status := &domain.ProjectStatus{
		Project:            "TestProject",
		Tag:                "1",
		TagName:            "Development",
		StartTime:          time.Date(2025, 9, 30, 10, 0, 0, 0, time.Local),
		CurrentSessionTime: 1 * time.Hour,
		TotalTime:          3 * time.Hour,
	}
	summaries := []domain.ProjectSummary{
		{Project: "TestProject", TotalTime: 3 * time.Hour},
	}

	result := renderDashboardWithWidth(status, summaries, now, 44)

	// 1カラムなので┬（2カラム区切り）は含まない
	if strings.Contains(result, "┬") {
		t.Error("1カラムモードで┬が含まれている")
	}

	// Status/Summaryラベルが含まれるか
	if !strings.Contains(result, "Status") {
		t.Error("'Status'が含まれていない")
	}
	if !strings.Contains(result, "Summary") {
		t.Error("'Summary'が含まれていない")
	}

	// 中間セパレーター（├...┤）
	if !strings.Contains(result, CrossL) {
		t.Errorf("中間セパレーター'%s'が含まれていない", CrossL)
	}
	if !strings.Contains(result, CrossR) {
		t.Errorf("中間セパレーター'%s'が含まれていない", CrossR)
	}

	// コンテンツ
	if !strings.Contains(result, "TestProject") {
		t.Error("プロジェクト名が含まれていない")
	}
	if !strings.Contains(result, "Today") {
		t.Error("'Today'が含まれていない")
	}
}

func TestRenderDashboard_WideWidth(t *testing.T) {
	now := time.Date(2025, 9, 30, 15, 0, 0, 0, time.Local)
	status := &domain.ProjectStatus{
		Project:            "VeryLongProjectNameThatWouldBeTruncated",
		Tag:                "1",
		TagName:            "Development",
		StartTime:          time.Date(2025, 9, 30, 10, 0, 0, 0, time.Local),
		CurrentSessionTime: 1 * time.Hour,
		TotalTime:          4 * time.Hour,
	}
	summaries := []domain.ProjectSummary{
		{Project: "VeryLongProjectNameThatWouldBeTruncated", TotalTime: 4 * time.Hour},
	}

	result80 := renderDashboardWithWidth(status, summaries, now, 80)
	result53 := renderDashboardWithWidth(status, summaries, now, 53)

	// 幅80ではプロジェクト名がより多く表示される
	if !strings.Contains(result80, "VeryLongProjectName") {
		t.Error("幅80でプロジェクト名が十分に表示されていない")
	}

	// 幅53では切り詰められる
	if strings.Contains(result53, "VeryLongProjectNameThatWouldBeTruncated") {
		t.Error("幅53でプロジェクト名が切り詰められていない")
	}

	// 両方とも2カラム
	if !strings.Contains(result80, "┬") {
		t.Error("幅80で2カラム区切りがない")
	}
	if !strings.Contains(result53, "┬") {
		t.Error("幅53で2カラム区切りがない")
	}
}

func TestRenderListWithWidth(t *testing.T) {
	baseTime := time.Date(2025, 9, 30, 10, 0, 0, 0, time.Local)
	summaries := []domain.ProjectSummary{
		{
			Project:   "ProjectA",
			TagName:   "Dev",
			TotalTime: 2 * time.Hour,
			TimeRanges: []domain.TimeRange{
				{Start: baseTime, End: baseTime.Add(2 * time.Hour), Duration: 2 * time.Hour},
			},
		},
	}
	now := baseTime.Add(8 * time.Hour)

	result44 := renderListWithWidth(summaries, now, 44)
	result80 := renderListWithWidth(summaries, now, 80)

	// 両方にプロジェクト名が含まれる
	if !strings.Contains(result44, "ProjectA") {
		t.Error("幅44でProjectAが含まれていない")
	}
	if !strings.Contains(result80, "ProjectA") {
		t.Error("幅80でProjectAが含まれていない")
	}

	// 幅80のほうがドットリーダーが長い（行が長い）
	lines44 := strings.Split(result44, "\n")
	lines80 := strings.Split(result80, "\n")
	var maxLen44, maxLen80 int
	for _, l := range lines44 {
		if w := displayWidth(l); w > maxLen44 {
			maxLen44 = w
		}
	}
	for _, l := range lines80 {
		if w := displayWidth(l); w > maxLen80 {
			maxLen80 = w
		}
	}
	if maxLen80 <= maxLen44 {
		t.Errorf("幅80の最大行幅(%d)が幅44(%d)以下", maxLen80, maxLen44)
	}
}

func TestProgressBarFit(t *testing.T) {
	t.Run("statusCol=24で収まる", func(t *testing.T) {
		result := renderTimeProgressBarFit(3*time.Hour, 8*time.Hour, 24)
		w := displayWidth(result)
		if w > 24 {
			t.Errorf("表示幅%dが24を超過: %q", w, result)
		}
		if !strings.Contains(result, "3h") {
			t.Errorf("時間情報が含まれていない: %q", result)
		}
	})

	t.Run("statusCol=40で収まる", func(t *testing.T) {
		result := renderTimeProgressBarFit(3*time.Hour, 8*time.Hour, 40)
		w := displayWidth(result)
		if w > 40 {
			t.Errorf("表示幅%dが40を超過: %q", w, result)
		}
		// バー文字が含まれる
		if !strings.Contains(result, BlockFull) && !strings.Contains(result, BlockLight) {
			t.Errorf("バーが含まれていない: %q", result)
		}
	})

	t.Run("statusCol=16でもパニックしない", func(t *testing.T) {
		result := renderTimeProgressBarFit(3*time.Hour, 8*time.Hour, 16)
		if result == "" {
			t.Error("空文字列が返された")
		}
	})
}
