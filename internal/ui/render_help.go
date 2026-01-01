package ui

import (
	"fmt"
	"strings"
)

// RenderHelp はヘルプメッセージを整形して表示する
func RenderHelp() string {
	var builder strings.Builder

	builder.WriteString("worklog ")
	builder.WriteString(Bullet)
	builder.WriteString(" プロジェクト稼働時間管理\n")
	builder.WriteString(renderSeparator(40))
	builder.WriteString("\n")

	builder.WriteString("Usage\n")
	builder.WriteString("  worklog <command> [options]\n\n")

	builder.WriteString("Commands\n")
	builder.WriteString("  new       新規プロジェクトを開始\n")
	builder.WriteString("  switch    プロジェクトを切り替え\n")
	builder.WriteString("  status    現在の稼働状況を表示\n")
	builder.WriteString("  stop      プロジェクトを停止\n")
	builder.WriteString("  list      本日の作業履歴を表示\n")
	builder.WriteString("  timeline  本日のタイムラインを表示\n")
	builder.WriteString("  help      ヘルプを表示\n")

	builder.WriteString(renderSeparator(40))
	builder.WriteString("\n")

	return builder.String()
}

// RenderError はエラーメッセージを整形して表示する
func RenderError(message string) string {
	var builder strings.Builder

	builder.WriteString(renderSeparator(40))
	builder.WriteString("\n")
	builder.WriteString(fmt.Sprintf("  \uf00d %s\n", message))
	builder.WriteString("  \n")
	builder.WriteString("  Try 'worklog help' for available commands\n")
	builder.WriteString(renderSeparator(40))
	builder.WriteString("\n")

	return builder.String()
}
