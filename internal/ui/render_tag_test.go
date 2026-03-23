package ui

import (
	"strings"
	"testing"

	"worklog/internal/domain"
)

func TestRenderTagList(t *testing.T) {
	t.Run("タグ一覧を表示", func(t *testing.T) {
		tags := []domain.Tag{
			{ID: 1, Name: "開発"},
			{ID: 2, Name: "会議"},
			{ID: 3, Name: "レビュー"},
		}

		result := RenderTagList(tags)

		// タイトルを含むことを確認
		if !strings.Contains(result, "Tags") {
			t.Error("タイトル 'Tags' が含まれていない")
		}

		// 各タグのIDと名前を含むことを確認
		for _, tag := range tags {
			if !strings.Contains(result, tag.Name) {
				t.Errorf("タグ名 %q が含まれていない", tag.Name)
			}
		}

		// 区切り線を含むことを確認
		if !strings.Contains(result, string(LineH)) {
			t.Error("区切り線が含まれていない")
		}
	})

	t.Run("空のタグリストを表示", func(t *testing.T) {
		tags := []domain.Tag{}

		result := RenderTagList(tags)

		// タイトルを含むことを確認
		if !strings.Contains(result, "Tags") {
			t.Error("タイトル 'Tags' が含まれていない")
		}

		// 空メッセージを含むことを確認
		if !strings.Contains(result, "タグがありません") {
			t.Error("空メッセージが含まれていない")
		}
	})
}

func TestRenderTagAdded(t *testing.T) {
	t.Run("タグ追加メッセージを表示", func(t *testing.T) {
		tag := domain.Tag{ID: 4, Name: "テスト"}

		result := RenderTagAdded(tag)

		// タグ名とIDを含むことを確認
		if !strings.Contains(result, tag.Name) {
			t.Errorf("タグ名 %q が含まれていない", tag.Name)
		}

		if !strings.Contains(result, "Tag added") {
			t.Error("'Tag added' メッセージが含まれていない")
		}

		// 丸角ボックスを含むことを確認
		if !strings.Contains(result, RoundTL) {
			t.Errorf("丸角ボックス %q が含まれていない", RoundTL)
		}
	})
}

func TestRenderTagDeleted(t *testing.T) {
	t.Run("タグ削除メッセージを表示", func(t *testing.T) {
		tagID := 2
		tagName := "会議"

		result := RenderTagDeleted(tagID, tagName)

		// タグ名とIDを含むことを確認
		if !strings.Contains(result, tagName) {
			t.Errorf("タグ名 %q が含まれていない", tagName)
		}

		if !strings.Contains(result, "Tag deleted") {
			t.Error("'Tag deleted' メッセージが含まれていない")
		}

		// 丸角ボックスを含むことを確認
		if !strings.Contains(result, RoundTL) {
			t.Errorf("丸角ボックス %q が含まれていない", RoundTL)
		}
	})
}
