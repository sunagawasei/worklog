package ui

import (
	"fmt"
	"strings"
	"time"
)

// RenderStopMessage は停止完了メッセージを整形して表示する
// project: プロジェクト名
// startTime: 開始時刻
// stopTime: 停止時刻
func RenderStopMessage(project string, startTime, stopTime time.Time) string {
	var builder strings.Builder

	builder.WriteString(renderSeparator(40))
	builder.WriteString("\n")

	// プロジェクト名と状態
	builder.WriteString(fmt.Sprintf("  %s stopped\n", project))

	// 時間範囲と経過時間
	elapsed := stopTime.Sub(startTime)
	builder.WriteString(fmt.Sprintf("  %s-%s %s %s\n",
		startTime.Format("15:04"),
		stopTime.Format("15:04"),
		Bullet,
		FormatDuration(elapsed)))

	builder.WriteString(renderSeparator(40))
	builder.WriteString("\n")

	return builder.String()
}

// RenderSwitchMessage はプロジェクト切り替えの状態遷移を整形して表示する
// oldProject: 停止したプロジェクト名
// oldStartTime: 停止したプロジェクトの開始時刻
// switchTime: 切り替え時刻
// newProject: 開始したプロジェクト名
// newTag: 開始したプロジェクトのタグ表示
func RenderSwitchMessage(oldProject string, oldStartTime, switchTime time.Time, newProject, newTag string) string {
	var builder strings.Builder

	builder.WriteString(renderSeparator(40))
	builder.WriteString("\n")

	// 停止したプロジェクト
	if oldProject != "" {
		builder.WriteString(fmt.Sprintf("  %s → stopped\n", oldProject))

		// 時間範囲と経過時間（RenderStopMessageと同じ形式）
		elapsed := switchTime.Sub(oldStartTime)
		builder.WriteString(fmt.Sprintf("  %s-%s %s %s\n",
			oldStartTime.Format("15:04"),
			switchTime.Format("15:04"),
			Bullet,
			FormatDuration(elapsed)))
	}

	// 開始したプロジェクト
	builder.WriteString(fmt.Sprintf("  %s → running\n", newProject))
	builder.WriteString(fmt.Sprintf("  %s %s %s\n",
		switchTime.Format("15:04"),
		Bullet,
		newTag))

	builder.WriteString(renderSeparator(40))
	builder.WriteString("\n")

	return builder.String()
}

// RenderNewMessage は新規プロジェクト開始メッセージを整形して表示する
// project: プロジェクト名
// startTime: 開始時刻
// tag: タグ表示
func RenderNewMessage(project string, startTime time.Time, tag string) string {
	var builder strings.Builder

	builder.WriteString(renderSeparator(40))
	builder.WriteString("\n")

	// プロジェクト開始
	builder.WriteString(fmt.Sprintf("  %s started\n", project))
	builder.WriteString(fmt.Sprintf("  %s %s %s\n",
		startTime.Format("15:04"),
		Bullet,
		tag))

	builder.WriteString(renderSeparator(40))
	builder.WriteString("\n")

	return builder.String()
}
