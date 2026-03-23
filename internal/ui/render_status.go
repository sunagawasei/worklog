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

	// Status部分の内容を準備（Phase 2 #5: 配置ロジック簡素化）
	var statusLines []string
	if status == nil {
		// アイドル状態（上下に空行を追加して呼吸感を確保）
		statusLines = []string{
			"",
			"稼働中のプロジェクトは",
			"ありません",
			"",
		}
	} else {
		// 稼働中：現在セッション時間 / 累計時間
		currentSessionStr := FormatDurationShort(status.CurrentSessionTime)
		totalTimeStr := FormatDurationShort(status.TotalTime)
		durationStr := fmt.Sprintf("%s / %s", currentSessionStr, totalTimeStr)

		statusLines = []string{
			"",
			fmt.Sprintf("■ %s running", status.Project),
			fmt.Sprintf("%s • %s", status.StartTime.Format("15:04"), durationStr),
		}
		if status.TagName != "" {
			statusLines = append(statusLines, status.TagName)
		}
		// プログレスバーは常に最終行に追加（Phase 2 #6: 実時間/目標時間表示）
		if totalTime > 0 {
			progressBar := RenderTimeProgressBar(totalTime, 8*time.Hour, 16)
			statusLines = append(statusLines, progressBar)
		}
	}

	// Summary部分の内容を準備
	summaryLines := []string{
		"",
		fmt.Sprintf("Today       %s", FormatDuration(totalTime)),
		fmt.Sprintf("Projects    %d", projectCount),
		fmt.Sprintf("Average     %s", FormatDuration(avgTime)),
		"",
	}

	// 上部境界線（カラム幅から動的生成）
	statusLabel := "─ Status "
	summaryLabel := "─ Summary "
	topLine := CornerTL +
		statusLabel + strings.Repeat(LineH, dashStatusCol+2-displayWidth(statusLabel)) +
		CrossT +
		summaryLabel + strings.Repeat(LineH, dashSummaryCol+2-displayWidth(summaryLabel)) +
		CornerTR + "\n"
	builder.WriteString(topLine)

	// 各行を描画（最大行数を決定）
	maxLines := max(len(statusLines), len(summaryLines))

	for i := 0; i < maxLines; i++ {
		// Status部分
		builder.WriteString(LineV + " ")
		if i < len(statusLines) {
			builder.WriteString(padString(statusLines[i], dashStatusCol))
		} else {
			builder.WriteString(strings.Repeat(" ", dashStatusCol))
		}

		// 区切り
		builder.WriteString(" " + LineV + " ")

		// Summary部分
		if i < len(summaryLines) {
			builder.WriteString(padString(summaryLines[i], dashSummaryCol))
		} else {
			builder.WriteString(strings.Repeat(" ", dashSummaryCol))
		}

		builder.WriteString(" " + LineV + "\n")
	}

	// 下部境界線（カラム幅から動的生成）
	bottomLine := CornerBL +
		strings.Repeat(LineH, dashStatusCol+2) +
		CrossB +
		strings.Repeat(LineH, dashSummaryCol+2) +
		CornerBR + "\n"
	builder.WriteString(bottomLine)

	return builder.String()
}
