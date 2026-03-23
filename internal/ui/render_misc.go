package ui

import (
	"fmt"
	"strings"

	"worklog/internal/domain"
)

// RenderHelp はヘルプメッセージを整形して表示する
func RenderHelp() string {
	var builder strings.Builder

	builder.WriteString("worklog ")
	builder.WriteString(Bullet)
	builder.WriteString(" プロジェクト稼働時間管理\n")
	builder.WriteString(renderSeparator(StandardWidth))
	builder.WriteString("\n")

	builder.WriteString("Usage\n")
	builder.WriteString("  worklog <command> [options]\n\n")

	builder.WriteString("Commands\n")
	builder.WriteString("  new       新規プロジェクトを開始\n")
	builder.WriteString("    worklog new [プロジェクト名] [タグID] [HH:MM]\n")
	builder.WriteString("  switch    プロジェクトを切り替え\n")
	builder.WriteString("    worklog switch [プロジェクト名] [タグID] [HH:MM]\n")
	builder.WriteString("  status    現在の稼働状況を表示\n")
	builder.WriteString("  stop      プロジェクトを停止\n")
	builder.WriteString("    worklog stop [HH:MM]\n")
	builder.WriteString("  list      本日の作業履歴を表示\n")
	builder.WriteString("  timeline  本日のタイムラインを表示\n")
	builder.WriteString("    worklog timeline [-1d|-2d|YYYY-MM-DD]\n")
	builder.WriteString("  tag       タグを管理 (list/add/delete)\n")
	builder.WriteString("    worklog tag add <名前>\n")
	builder.WriteString("    worklog tag delete <ID>\n")
	builder.WriteString("  help      ヘルプを表示\n")

	builder.WriteString("\n")
	builder.WriteString("Examples\n")
	builder.WriteString("  worklog new TASK-001 1         # 対話的にタグを選択\n")
	builder.WriteString("  worklog new TASK-001 1 09:30   # 時刻を指定して開始\n")
	builder.WriteString("  worklog switch                 # 対話的に切り替え\n")
	builder.WriteString("  worklog stop 18:00             # 18:00に停止\n")
	builder.WriteString("  worklog timeline -1d           # 昨日のタイムライン\n")

	builder.WriteString(renderSeparator(StandardWidth))
	builder.WriteString("\n")

	return builder.String()
}

// RenderError はエラーメッセージを丸角ボックスで整形して表示する
func RenderError(message string) string {
	lines := []string{
		fmt.Sprintf("[!] %s", message),
		"",
		"'worklog help' で利用可能なコマンドを確認できます",
	}
	return renderRoundBox(lines, StandardWidth)
}

// RenderTagList はタグ一覧を整形して表示する
func RenderTagList(tags []domain.Tag) string {
	var builder strings.Builder

	builder.WriteString("Tags\n")
	builder.WriteString(renderSeparator(StandardWidth))
	builder.WriteString("\n")

	if len(tags) == 0 {
		builder.WriteString("  (タグがありません)\n")
	} else {
		for _, tag := range tags {
			builder.WriteString(fmt.Sprintf("  %2d - %s\n", tag.ID, tag.Name))
		}
	}

	builder.WriteString(renderSeparator(StandardWidth))
	builder.WriteString("\n")

	return builder.String()
}

// RenderTagAdded はタグ追加完了メッセージを丸角ボックスで整形して表示する
func RenderTagAdded(tag domain.Tag) string {
	lines := []string{fmt.Sprintf("Tag added: %s (ID: %d)", tag.Name, tag.ID)}
	return renderRoundBox(lines, StandardWidth)
}

// RenderTagDeleted はタグ削除完了メッセージを丸角ボックスで整形して表示する
func RenderTagDeleted(tagID int, tagName string) string {
	lines := []string{fmt.Sprintf("Tag deleted: %s (ID: %d)", tagName, tagID)}
	return renderRoundBox(lines, StandardWidth)
}
