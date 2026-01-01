package ui

import (
	"fmt"
	"strings"
	"time"

	"worklog/internal/domain"
)

// RenderList は本日のプロジェクト一覧を整形して表示する
func RenderList(summaries []domain.ProjectSummary, now time.Time) string {
	var builder strings.Builder

	// 日付ヘッダー
	dateStr := now.Format("2006-01-02")
	builder.WriteString(fmt.Sprintf("Today %s %s\n", LineV, dateStr))

	// summariesが空の場合
	if len(summaries) == 0 {
		builder.WriteString(renderSeparator(6))
		builder.WriteString(Cross)
		builder.WriteString(renderSeparator(32))
		builder.WriteString("\n")
		builder.WriteString(fmt.Sprintf("      %s\n", LineV))
		builder.WriteString(fmt.Sprintf("      %s 本日の作業履歴はありません\n", LineV))
		builder.WriteString(fmt.Sprintf("      %s\n", LineV))
		builder.WriteString(renderSeparator(6))
		builder.WriteString(CrossB)
		builder.WriteString(renderSeparator(32))
		builder.WriteString("\n")
		return builder.String()
	}

	// 上部区切り線
	builder.WriteString(renderSeparator(6))
	builder.WriteString(CrossL)
	builder.WriteString(renderSeparator(32))
	builder.WriteString("\n")

	// 空行
	builder.WriteString(fmt.Sprintf("      %s\n", LineV))

	// 各プロジェクト
	totalDuration := time.Duration(0)
	for _, summary := range summaries {
		durationStr := FormatDuration(summary.TotalTime)
		totalDuration += summary.TotalTime

		// タグ名がある場合は表示
		if summary.TagName != "" {
			builder.WriteString(fmt.Sprintf("      %s %s %s %s %s\n",
				LineV,
				summary.Project,
				Bullet,
				summary.TagName,
				durationStr))
		} else {
			builder.WriteString(fmt.Sprintf("      %s %s %s\n",
				LineV,
				summary.Project,
				durationStr))
		}

		// 時間範囲を表示
		for i, tr := range summary.TimeRanges {
			isLast := i == len(summary.TimeRanges)-1
			connector := "├─"
			if isLast {
				connector = "└─"
			}

			timeRange := fmt.Sprintf("%s-%s",
				tr.Start.Format("15:04"),
				tr.End.Format("15:04"))

			rangeDuration := FormatDuration(tr.Duration)

			builder.WriteString(fmt.Sprintf("      %s %s %s %6s\n",
				LineV,
				connector,
				timeRange,
				rangeDuration))
		}
	}

	// 空行
	builder.WriteString(fmt.Sprintf("      %s\n", LineV))

	// 下部区切り線
	builder.WriteString(renderSeparator(6))
	builder.WriteString(CrossB)
	builder.WriteString(renderSeparator(32))
	builder.WriteString("\n")

	// 合計行
	totalStr := FormatDuration(totalDuration)
	builder.WriteString(fmt.Sprintf("Total   %s\n", totalStr))

	return builder.String()
}
