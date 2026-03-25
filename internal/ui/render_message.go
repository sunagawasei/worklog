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
// project: プロジェクト名
// startTime: 開始時刻
// stopTime: 停止時刻
func RenderStopMessage(project string, startTime, stopTime time.Time) string {
	elapsed := stopTime.Sub(startTime)
	lines := []string{
		fmt.Sprintf("%s stopped", project),
		fmt.Sprintf("%s-%s %s %s",
			startTime.Format("15:04"),
			stopTime.Format("15:04"),
			Bullet,
			FormatDuration(elapsed)),
	}
	return renderRoundBox(lines, StandardWidth)
}

// RenderSwitchMessage はプロジェクト切り替えの状態遷移を整形して表示する
// oldProject: 停止したプロジェクト名
// oldStartTime: 停止したプロジェクトの開始時刻
// switchTime: 切り替え時刻
// newProject: 開始したプロジェクト名
// newTag: 開始したプロジェクトのタグ表示
func RenderSwitchMessage(oldProject string, oldStartTime, switchTime time.Time, newProject, newTag string) string {
	var lines []string

	// 停止したプロジェクト
	if oldProject != "" {
		elapsed := switchTime.Sub(oldStartTime)
		lines = append(lines, fmt.Sprintf("%s → stopped", oldProject))
		lines = append(lines, fmt.Sprintf("%s-%s %s %s",
			oldStartTime.Format("15:04"),
			switchTime.Format("15:04"),
			Bullet,
			FormatDuration(elapsed)))
	}

	// 開始したプロジェクト
	lines = append(lines, fmt.Sprintf("%s → running", newProject))
	lines = append(lines, fmt.Sprintf("%s %s %s",
		switchTime.Format("15:04"),
		Bullet,
		newTag))

	return renderRoundBox(lines, StandardWidth)
}

// RenderNewMessage は新規プロジェクト開始メッセージを整形して表示する
// project: プロジェクト名
// startTime: 開始時刻
// tag: タグ表示
func RenderNewMessage(project string, startTime time.Time, tag string) string {
	lines := []string{
		fmt.Sprintf("%s started", project),
		fmt.Sprintf("%s %s %s", startTime.Format("15:04"), Bullet, tag),
	}
	return renderRoundBox(lines, StandardWidth)
}
