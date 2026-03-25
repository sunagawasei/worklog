package ui

import (
	"fmt"
	"strings"
	"time"

	"worklog/internal/domain"
)

// RenderDashboard は現在の稼働状況とサマリー統計を表示する
func RenderDashboard(status *domain.ProjectStatus, summaries []domain.ProjectSummary, now time.Time) string {
	return renderDashboardWithWidth(status, summaries, now, contentWidth())
}

func renderDashboardWithWidth(status *domain.ProjectStatus, summaries []domain.ProjectSummary, now time.Time, width int) string {
	if width < 53 {
		return renderDashboardSingleColumn(status, summaries, now, width)
	}
	return renderDashboardTwoColumn(status, summaries, now, width)
}

// dashboardData はダッシュボードのStatus/Summary行を準備する
func dashboardData(status *domain.ProjectStatus, summaries []domain.ProjectSummary, statusCol int) (statusLines, summaryLines []string, totalTime time.Duration) {
	projectCount := len(summaries)
	for _, summary := range summaries {
		totalTime += summary.TotalTime
	}

	var avgTime time.Duration
	if projectCount > 0 {
		avgTime = totalTime / time.Duration(projectCount)
	}

	// Status部分
	if status == nil {
		statusLines = []string{
			"",
			"稼働中のプロジェクトは",
			"ありません",
			"",
		}
	} else {
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
		if totalTime > 0 {
			progressBar := renderTimeProgressBarFit(totalTime, 8*time.Hour, statusCol)
			statusLines = append(statusLines, progressBar)
		}
	}

	// Summary部分
	summaryLines = []string{
		"",
		fmt.Sprintf("Today       %s", FormatDuration(totalTime)),
		fmt.Sprintf("Projects    %d", projectCount),
		fmt.Sprintf("Average     %s", FormatDuration(avgTime)),
		"",
	}

	return
}

// renderTimeProgressBarFit はmaxWidth内に収まるプログレスバーを生成する
func renderTimeProgressBarFit(current time.Duration, totalGoal time.Duration, maxWidth int) string {
	if totalGoal <= 0 {
		return ""
	}

	percent := float64(current) / float64(totalGoal)

	// 時間テキストを先に生成
	currentH := int(current.Hours())
	currentM := int(current.Minutes()) % 60
	goalH := int(totalGoal.Hours())

	var currentStr string
	if currentH == 0 {
		currentStr = fmt.Sprintf("%dm", currentM)
	} else {
		currentStr = fmt.Sprintf("%dh %02dm", currentH, currentM)
	}

	percentInt := int(percent * 100)
	if percentInt > 100 {
		percentInt = 100
	}

	timeText := fmt.Sprintf("  %s/%dh (%d%%)", currentStr, goalH, percentInt)
	textWidth := displayWidth(timeText)

	barWidth := maxWidth - textWidth
	if barWidth > 20 {
		barWidth = 20
	}
	if barWidth < 4 {
		// バーを表示する余裕がない場合はテキストのみ
		return strings.TrimLeft(timeText, " ")
	}

	bar := renderBarOnly(percent, barWidth)
	return bar + timeText
}

func renderDashboardTwoColumn(status *domain.ProjectStatus, summaries []domain.ProjectSummary, now time.Time, width int) string {
	var builder strings.Builder
	statusCol, summaryCol := dashColumns(width)

	statusLines, summaryLines, _ := dashboardData(status, summaries, statusCol)

	// 上部境界線
	statusLabel := "─ Status "
	summaryLabel := "─ Summary "
	topLine := CornerTL +
		statusLabel + strings.Repeat(LineH, statusCol+2-displayWidth(statusLabel)) +
		CrossT +
		summaryLabel + strings.Repeat(LineH, summaryCol+2-displayWidth(summaryLabel)) +
		CornerTR + "\n"
	builder.WriteString(topLine)

	// 各行を描画
	maxLines := max(len(statusLines), len(summaryLines))

	for i := 0; i < maxLines; i++ {
		builder.WriteString(LineV + " ")
		if i < len(statusLines) {
			builder.WriteString(padString(statusLines[i], statusCol))
		} else {
			builder.WriteString(strings.Repeat(" ", statusCol))
		}

		builder.WriteString(" " + LineV + " ")

		if i < len(summaryLines) {
			builder.WriteString(padString(summaryLines[i], summaryCol))
		} else {
			builder.WriteString(strings.Repeat(" ", summaryCol))
		}

		builder.WriteString(" " + LineV + "\n")
	}

	// 下部境界線
	bottomLine := CornerBL +
		strings.Repeat(LineH, statusCol+2) +
		CrossB +
		strings.Repeat(LineH, summaryCol+2) +
		CornerBR + "\n"
	builder.WriteString(bottomLine)

	return builder.String()
}

func renderDashboardSingleColumn(status *domain.ProjectStatus, summaries []domain.ProjectSummary, now time.Time, width int) string {
	var builder strings.Builder
	innerW := width - dashSingleFrame

	statusLines, summaryLines, _ := dashboardData(status, summaries, innerW)

	// 上部: ┌─ Status ──...──┐
	statusLabel := "─ Status "
	topLine := CornerTL +
		statusLabel + strings.Repeat(LineH, width-2-displayWidth(statusLabel)) +
		CornerTR + "\n"
	builder.WriteString(topLine)

	// Status行
	for _, line := range statusLines {
		builder.WriteString(LineV + " " + padString(line, innerW) + " " + LineV + "\n")
	}

	// 中間: ├─ Summary ──...──┤
	summaryLabel := "─ Summary "
	midLine := CrossL +
		summaryLabel + strings.Repeat(LineH, width-2-displayWidth(summaryLabel)) +
		CrossR + "\n"
	builder.WriteString(midLine)

	// Summary行
	for _, line := range summaryLines {
		builder.WriteString(LineV + " " + padString(line, innerW) + " " + LineV + "\n")
	}

	// 下部: └──...──┘
	builder.WriteString(CornerBL + strings.Repeat(LineH, width-2) + CornerBR + "\n")

	return builder.String()
}
