package ui

import (
	"fmt"
	"strings"

	"worklog/internal/domain"
)

// RenderTagList はタグ一覧を整形して表示する
func RenderTagList(tags []domain.Tag) string {
	var builder strings.Builder

	builder.WriteString("Tags\n")
	builder.WriteString(renderSeparator(40))
	builder.WriteString("\n")

	if len(tags) == 0 {
		builder.WriteString("  (タグがありません)\n")
	} else {
		for _, tag := range tags {
			builder.WriteString(fmt.Sprintf("  %2d - %s\n", tag.ID, tag.Name))
		}
	}

	builder.WriteString(renderSeparator(40))
	builder.WriteString("\n")

	return builder.String()
}

// RenderTagAdded はタグ追加完了メッセージを整形して表示する
func RenderTagAdded(tag domain.Tag) string {
	var builder strings.Builder

	builder.WriteString(renderSeparator(40))
	builder.WriteString("\n")

	builder.WriteString(fmt.Sprintf("  Tag added: %s (ID: %d)\n", tag.Name, tag.ID))

	builder.WriteString(renderSeparator(40))
	builder.WriteString("\n")

	return builder.String()
}

// RenderTagDeleted はタグ削除完了メッセージを整形して表示する
func RenderTagDeleted(tagID int, tagName string) string {
	var builder strings.Builder

	builder.WriteString(renderSeparator(40))
	builder.WriteString("\n")

	builder.WriteString(fmt.Sprintf("  Tag deleted: %s (ID: %d)\n", tagName, tagID))

	builder.WriteString(renderSeparator(40))
	builder.WriteString("\n")

	return builder.String()
}
