package ui

import (
	"fmt"
	"strings"
	"time"

	"worklog/internal/domain"
)

// RenderDashboard は現在の稼働状況とサマリー統計を2カラムで表示する
// status: 現在の稼働状況（nilの場合はアイドル）
// summaries: 本日のプロジェクトサマリー一覧
// now: 現在時刻
func RenderDashboard(status *domain.ProjectStatus, summaries []domain.ProjectSummary, now time.Time) string {
	var builder strings.Builder

	// 合計時間とプロジェクト数を計算
	totalTime := time.Duration(0)
	projectCount := len(summaries)
	for _, summary := range summaries {
		totalTime += summary.TotalTime
	}

	// 平均作業時間を計算
	var avgTime time.Duration
	if projectCount > 0 {
		avgTime = totalTime / time.Duration(projectCount)
	}

	// Status部分の内容を準備
	var statusLines []string
	if status == nil {
		// アイドル状態
		statusLines = append(statusLines, "稼働中のプロジェクトは")
		statusLines = append(statusLines, "ありません")
		statusLines = append(statusLines, "")
	} else {
		// 稼働中
		// 新しい形式：現在セッション時間 / 累計時間
		currentSessionStr := FormatDurationShort(status.CurrentSessionTime)
		totalTimeStr := FormatDurationShort(status.TotalTime)
		durationStr := fmt.Sprintf("%s / %s", currentSessionStr, totalTimeStr)

		statusLines = append(statusLines, fmt.Sprintf("■ %s running", status.Project))
		if status.TagName != "" {
			statusLines = append(statusLines, fmt.Sprintf("%s • %s", status.StartTime.Format("15:04"), durationStr))
			statusLines = append(statusLines, status.TagName)
		} else {
			statusLines = append(statusLines, fmt.Sprintf("%s • %s", status.StartTime.Format("15:04"), durationStr))
			statusLines = append(statusLines, "")
		}

		// プログレスバーを追加（totalTimeが0より大きい場合）
		if totalTime > 0 {
			percent := float64(totalTime) / float64(8*time.Hour) // 8時間を100%と想定
			progressBar := RenderProgressBar(percent, 16)
			// プログレスバーを追加する代わりに、statusLinesを調整
			if len(statusLines) > 2 && statusLines[2] != "" {
				// タグ名がある場合は、プログレスバーを追加
				statusLines = append(statusLines, progressBar)
			} else {
				// タグ名がない場合は、空行をプログレスバーに置き換え
				statusLines[2] = progressBar
			}
		}
	}

	// Summary部分の内容を準備
	summaryLines := []string{
		fmt.Sprintf("Today       %s", FormatDuration(totalTime)),
		fmt.Sprintf("Projects    %d", projectCount),
		fmt.Sprintf("Average     %s", FormatDuration(avgTime)),
	}

	// 上部境界線
	builder.WriteString("┌─ Status ─────────────────┬─ Summary ──────────────┐\n")

	// 各行を描画（最大行数を決定）
	maxLines := len(statusLines)
	if len(summaryLines) > maxLines {
		maxLines = len(summaryLines)
	}

	for i := 0; i < maxLines; i++ {
		// Status部分
		builder.WriteString("│ ")
		if i < len(statusLines) {
			builder.WriteString(padString(statusLines[i], 24))
		} else {
			builder.WriteString(strings.Repeat(" ", 24))
		}

		// 区切り
		builder.WriteString(" │ ")

		// Summary部分
		if i < len(summaryLines) {
			builder.WriteString(padString(summaryLines[i], 22))
		} else {
			builder.WriteString(strings.Repeat(" ", 22))
		}

		builder.WriteString(" │\n")
	}

	// 下部境界線
	builder.WriteString("└──────────────────────────┴────────────────────────┘\n")

	return builder.String()
}

// RenderStatus は現在の稼働状況を整形して表示する
// statusがnilの場合はアイドル状態として表示
// todayTotal: 今日の合計作業時間（プログレスバー計算用）
func RenderStatus(status *domain.ProjectStatus, now time.Time, todayTotal time.Duration) string {
	var builder strings.Builder

	// 区切り線
	builder.WriteString(renderSeparator(38))
	builder.WriteString("\n")

	if status == nil {
		// アイドル状態
		builder.WriteString("  稼働中のプロジェクトはありません\n")
	} else {
		// 稼働中
		builder.WriteString(fmt.Sprintf("  %s %s running\n",
			status.Project,
			Arrow))

		// 経過時間を計算
		elapsed := now.Sub(status.StartTime)
		durationStr := FormatDuration(elapsed)

		// タグ名がある場合は表示
		if status.TagName != "" {
			builder.WriteString(fmt.Sprintf("  %s %s %s %s %s\n",
				status.StartTime.Format("15:04"),
				Bullet,
				durationStr,
				Bullet,
				status.TagName))
		} else {
			builder.WriteString(fmt.Sprintf("  %s %s %s\n",
				status.StartTime.Format("15:04"),
				Bullet,
				durationStr))
		}

		// プログレスバーを表示（todayTotalが0より大きい場合）
		if todayTotal > 0 {
			percent := float64(todayTotal) / float64(8*time.Hour) // 8時間を100%と想定
			progressBar := RenderProgressBar(percent, 20)
			builder.WriteString(fmt.Sprintf("  %s\n", progressBar))
		}
	}

	return builder.String()
}
