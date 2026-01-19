package cmd

import (
	"fmt"
	"os"
	"strconv"

	"worklog/internal/domain"
	"worklog/internal/project"
	"worklog/internal/ui"
)

// handleTag はタグ管理コマンドのメインハンドラー
func handleTag(manager project.ProjectManager) error {
	// サブコマンドなし、または "list" の場合は一覧表示
	if len(os.Args) < 3 || os.Args[2] == "list" {
		return handleTagList(manager)
	}

	// サブコマンドで分岐
	switch os.Args[2] {
	case "add":
		return handleTagAdd(manager)
	case "delete":
		return handleTagDelete(manager)
	default:
		return fmt.Errorf("不明なサブコマンド: %s\n使い方: worklog tag [list|add|delete]", os.Args[2])
	}
}

// handleTagList はタグ一覧を表示する
func handleTagList(manager project.ProjectManager) error {
	tags, err := manager.GetTags()
	if err != nil {
		return fmt.Errorf("タグの読み込みに失敗しました: %w", err)
	}

	output := ui.RenderTagList(tags)
	fmt.Print(output)
	return nil
}

// handleTagAdd は新しいタグを追加する
func handleTagAdd(manager project.ProjectManager) error {
	// タグ名が指定されているかチェック
	if len(os.Args) < 4 {
		return fmt.Errorf("タグ名を指定してください\n使い方: worklog tag add <name>")
	}

	tagName := os.Args[3]

	// 空の名前チェック
	if tagName == "" {
		return fmt.Errorf("タグ名を指定してください")
	}

	// タグを追加
	tag, err := manager.AddTag(tagName)
	if err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("タグ \"%s\" は既に存在します", tagName)
		}
		return fmt.Errorf("タグの追加に失敗しました: %w", err)
	}

	// 成功メッセージを表示
	output := ui.RenderTagAdded(tag)
	fmt.Print(output)
	return nil
}

// handleTagDelete はタグを削除する
func handleTagDelete(manager project.ProjectManager) error {
	// タグIDが指定されているかチェック
	if len(os.Args) < 4 {
		return fmt.Errorf("タグIDを指定してください\n使い方: worklog tag delete <id>")
	}

	// タグIDをパース
	tagID, err := strconv.Atoi(os.Args[3])
	if err != nil {
		return fmt.Errorf("タグIDは数値で指定してください: %s", os.Args[3])
	}

	// 削除前にタグ一覧を取得して、削除対象のタグを確認
	tags, err := manager.GetTags()
	if err != nil {
		return fmt.Errorf("タグの読み込みに失敗しました: %w", err)
	}

	// 削除対象のタグを検索
	var targetTag *domain.Tag
	for _, tag := range tags {
		if tag.ID == tagID {
			t := tag
			targetTag = &t
			break
		}
	}

	if targetTag == nil {
		return fmt.Errorf("ID %d のタグは存在しません", tagID)
	}

	// 削除確認プロンプト
	confirmed, err := ui.ConfirmAction(fmt.Sprintf("タグ \"%s\" を削除しますか？", targetTag.Name))
	if err != nil {
		return fmt.Errorf("確認プロンプトでエラーが発生しました: %w", err)
	}

	if !confirmed {
		fmt.Println("削除をキャンセルしました")
		return nil
	}

	// タグを削除
	err = manager.DeleteTag(tagID)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("ID %d のタグは存在しません", tagID)
		}
		return fmt.Errorf("タグの削除に失敗しました: %w", err)
	}

	// 成功メッセージを表示
	output := ui.RenderTagDeleted(targetTag.ID, targetTag.Name)
	fmt.Print(output)
	return nil
}
