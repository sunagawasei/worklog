package ui

import (
	"fmt"
	"strings"
	"time"

	"worklog/internal/domain"
)

// listContentWidth はリスト表示の内容幅（固定）
const listContentWidth = 32

// renderDotLeaderLine はドットリーダーで左右の文字列を結合する
// 例: "ProjectA ··············· 2h 30m"
func renderDotLeaderLine(left, right string, totalWidth int) string {
	leftW := displayWidth(left)
	rightW := displayWidth(right)
	// 左右の間にドットを入れる（最低2個）
	dots := totalWidth - leftW - rightW - 2
	if dots < 2 {
		dots = 2
	}
	return left + " " + strings.Repeat(DotLeader, dots) + " " + right
}

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
		builder.WriteString(renderSeparator(listContentWidth))
		builder.WriteString("\n")
		builder.WriteString(fmt.Sprintf("      %s\n", LineV))
		builder.WriteString(fmt.Sprintf("      %s 本日の作業履歴はありません\n", LineV))
		builder.WriteString(fmt.Sprintf("      %s 'worklog new' で作業を開始できます\n", LineV))
		builder.WriteString(fmt.Sprintf("      %s\n", LineV))
		builder.WriteString(renderSeparator(6))
		builder.WriteString(CrossB)
		builder.WriteString(renderSeparator(listContentWidth))
		builder.WriteString("\n")
		return builder.String()
	}

	// 上部区切り線
	builder.WriteString(renderSeparator(6))
	builder.WriteString(CrossL)
	builder.WriteString(renderSeparator(listContentWidth))
	builder.WriteString("\n")

	// 各プロジェクト（プロジェクト間に空行）
	totalDuration := time.Duration(0)
	for _, summary := range summaries {
		durationStr := FormatDuration(summary.TotalTime)
		totalDuration += summary.TotalTime

		// プロジェクト間の空行
		builder.WriteString(fmt.Sprintf("      %s\n", LineV))

		// プロジェクト名行（ドットリーダーで右揃え時間）
		var leftPart string
		if summary.TagName != "" {
			leftPart = fmt.Sprintf("%s %s %s %s",
				LineV, summary.Project, Bullet, summary.TagName)
		} else {
			leftPart = fmt.Sprintf("%s %s", LineV, summary.Project)
		}
		line := renderDotLeaderLine(leftPart, durationStr, listContentWidth+7)
		builder.WriteString(fmt.Sprintf("      %s\n", line))

		// 時間範囲を表示
		for j, tr := range summary.TimeRanges {
			isLast := j == len(summary.TimeRanges)-1
			connector := "├─"
			if isLast {
				connector = "└─"
			}

			timeRange := fmt.Sprintf("%s-%s",
				tr.Start.Format("15:04"),
				tr.End.Format("15:04"))

			rangeDuration := FormatDurationShort(tr.Duration)
			rangeLeft := fmt.Sprintf("%s %s %s", LineV, connector, timeRange)
			rangeLine := renderDotLeaderLine(rangeLeft, rangeDuration, listContentWidth+7)
			builder.WriteString(fmt.Sprintf("      %s\n", rangeLine))
		}
	}

	// 空行
	builder.WriteString(fmt.Sprintf("      %s\n", LineV))

	// 下部区切り線（二重線でTotal行を強調）
	builder.WriteString(strings.Repeat(LineH2, 6))
	builder.WriteString(CrossB2)
	builder.WriteString(strings.Repeat(LineH2, listContentWidth))
	builder.WriteString("\n")

	// 合計行（ドットリーダーで右揃え）
	totalStr := FormatDuration(totalDuration)
	totalLine := renderDotLeaderLine("Total", totalStr, listContentWidth+7)
	builder.WriteString(fmt.Sprintf("      %s\n", totalLine))

	return builder.String()
}
