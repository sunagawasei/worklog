package ui

import (
	"fmt"
	"strings"
	"time"
)

// renderRoundBox は丸角ボックス内に行を表示する
// 形式:
//
//	╭────────────────────────────────────────╮
//	│  content                               │
//	╰────────────────────────────────────────╯
func renderRoundBox(lines []string, width int) string {
	var builder strings.Builder
	innerWidth := width - 6 // "│  " と "  │" で6文字使用

	builder.WriteString(RoundTL + strings.Repeat(LineH, width-2) + RoundTR + "\n")
	for _, line := range lines {
		content := padString(line, innerWidth)
		builder.WriteString(LineV + "  " + content + "  " + LineV + "\n")
	}
	builder.WriteString(RoundBL + strings.Repeat(LineH, width-2) + RoundBR + "\n")

	return builder.String()
}

// RenderStopMessage は停止完了メッセージを整形して表示する
func RenderStopMessage(project string, startTime, stopTime time.Time) string {
	return renderStopMessageWithWidth(project, startTime, stopTime, contentWidth())
}

func renderStopMessageWithWidth(project string, startTime, stopTime time.Time, width int) string {
	lines := []string{
		fmt.Sprintf("%s stopped", project),
	}
	if startTime.IsZero() {
		lines = append(lines, stopTime.Format("15:04"))
	} else {
		elapsed := stopTime.Sub(startTime)
		lines = append(lines, fmt.Sprintf("%s-%s %s %s",
			startTime.Format("15:04"),
			stopTime.Format("15:04"),
			Bullet,
			FormatDuration(elapsed)))
	}
	return renderRoundBox(lines, width)
}

// RenderSwitchMessage はプロジェクト切り替えの状態遷移を整形して表示する
func RenderSwitchMessage(oldProject string, oldStartTime, switchTime time.Time, newProject, newTag string) string {
	return renderSwitchMessageWithWidth(oldProject, oldStartTime, switchTime, newProject, newTag, contentWidth())
}

func renderSwitchMessageWithWidth(oldProject string, oldStartTime, switchTime time.Time, newProject, newTag string, width int) string {
	var lines []string

	// 停止したプロジェクト
	if oldProject != "" {
		lines = append(lines, fmt.Sprintf("%s → stopped", oldProject))
		if oldStartTime.IsZero() {
			lines = append(lines, switchTime.Format("15:04"))
		} else {
			elapsed := switchTime.Sub(oldStartTime)
			lines = append(lines, fmt.Sprintf("%s-%s %s %s",
				oldStartTime.Format("15:04"),
				switchTime.Format("15:04"),
				Bullet,
				FormatDuration(elapsed)))
		}
	}

	// 開始したプロジェクト
	lines = append(lines, fmt.Sprintf("%s → running", newProject))
	lines = append(lines, fmt.Sprintf("%s %s %s",
		switchTime.Format("15:04"),
		Bullet,
		newTag))

	return renderRoundBox(lines, width)
}

// RenderNewMessage は新規プロジェクト開始メッセージを整形して表示する
func RenderNewMessage(project string, startTime time.Time, tag string) string {
	return renderNewMessageWithWidth(project, startTime, tag, contentWidth())
}

func renderNewMessageWithWidth(project string, startTime time.Time, tag string, width int) string {
	lines := []string{
		fmt.Sprintf("%s started", project),
		fmt.Sprintf("%s %s %s", startTime.Format("15:04"), Bullet, tag),
	}
	return renderRoundBox(lines, width)
}
