package cmd

import (
	"fmt"
	"os"
	"strconv"

	"worklog/internal/domain"
	"worklog/internal/project"
	"worklog/internal/storage"
	"worklog/internal/ui"
)

// handleTag はタグ管理コマンドのメインハンドラー
func handleTag(manager project.ProjectManager, opts ExecOptions) error {
	// opts.Args: [0]="tag", [1]=サブコマンド, [2]=引数
	// サブコマンドなし、または "list" の場合は一覧表示
	if len(opts.Args) < 2 || opts.Args[1] == "list" {
		return handleTagList(manager, opts)
	}

	switch opts.Args[1] {
	case "add":
		return handleTagAdd(manager, opts)
	case "delete":
		return handleTagDelete(manager, opts)
	default:
		return jsonError(opts, "UNKNOWN_SUBCOMMAND", fmt.Sprintf("不明なサブコマンド: %s\n使い方: worklog tag [list|add|delete]", opts.Args[1]))
	}
}

// handleTagList はタグ一覧を表示する
func handleTagList(manager project.ProjectManager, opts ExecOptions) error {
	tags, err := manager.GetTags()
	if err != nil {
		return jsonError(opts, "INTERNAL_ERROR", fmt.Sprintf("タグの読み込みに失敗しました: %v", err))
	}

	if opts.JSONMode {
		if tags == nil {
			tags = []domain.Tag{}
		}
		writeJSON(opts.writer(), tags)
		return nil
	}

	output := ui.RenderTagList(tags)
	fmt.Fprint(opts.writer(), output)
	return nil
}

// handleTagAdd は新しいタグを追加する
func handleTagAdd(manager project.ProjectManager, opts ExecOptions) error {
	// opts.Args: [0]="tag", [1]="add", [2]=タグ名
	if len(opts.Args) < 3 {
		return jsonError(opts, "MISSING_ARGUMENTS", "タグ名を指定してください\n使い方: worklog tag add <name>")
	}

	tagName := opts.Args[2]
	if tagName == "" {
		return jsonError(opts, "MISSING_ARGUMENTS", "タグ名を指定してください")
	}

	if err := storage.ValidateTagName(tagName); err != nil {
		return jsonError(opts, "INVALID_TAG_NAME", err.Error())
	}

	tag, err := manager.AddTag(tagName)
	if err != nil {
		if os.IsExist(err) {
			return jsonError(opts, "TAG_ALREADY_EXISTS", fmt.Sprintf("タグ \"%s\" は既に存在します", tagName))
		}
		return jsonError(opts, "INTERNAL_ERROR", fmt.Sprintf("タグの追加に失敗しました: %v", err))
	}

	if opts.JSONMode {
		writeJSON(opts.writer(), tagResultJSON{Action: "added", ID: tag.ID, Name: tag.Name})
		return nil
	}

	output := ui.RenderTagAdded(tag)
	fmt.Fprint(opts.writer(), output)
	return nil
}

// handleTagDelete はタグを削除する
func handleTagDelete(manager project.ProjectManager, opts ExecOptions) error {
	// opts.Args: [0]="tag", [1]="delete", [2]=タグID
	if len(opts.Args) < 3 {
		return jsonError(opts, "MISSING_ARGUMENTS", "タグIDを指定してください\n使い方: worklog tag delete <id>")
	}

	tagID, err := strconv.Atoi(opts.Args[2])
	if err != nil {
		return jsonError(opts, "INVALID_TAG_ID", fmt.Sprintf("タグIDは数値で指定してください: %s", opts.Args[2]))
	}

	tags, err := manager.GetTags()
	if err != nil {
		return jsonError(opts, "INTERNAL_ERROR", fmt.Sprintf("タグの読み込みに失敗しました: %v", err))
	}

	var targetTag *domain.Tag
	for _, tag := range tags {
		if tag.ID == tagID {
			t := tag
			targetTag = &t
			break
		}
	}

	if targetTag == nil {
		return jsonError(opts, "TAG_NOT_FOUND", fmt.Sprintf("ID %d のタグは存在しません", tagID))
	}

	// --no-interactive モードでは確認プロンプトをスキップ（安全のためエラー）
	if opts.NoInteractive {
		return jsonError(opts, "INTERACTIVE_REQUIRED", "tag delete は確認のため --no-interactive モードでは実行できません。--force オプションは未実装です")
	}

	confirmed, err := ui.NewPromptUI().ConfirmAction(fmt.Sprintf("タグ \"%s\" を削除", targetTag.Name))
	if err != nil {
		return fmt.Errorf("確認プロンプトでエラーが発生しました: %w", err)
	}

	if !confirmed {
		fmt.Fprintln(opts.writer(), "削除をキャンセルしました")
		return nil
	}

	err = manager.DeleteTag(tagID)
	if err != nil {
		if os.IsNotExist(err) {
			return jsonError(opts, "TAG_NOT_FOUND", fmt.Sprintf("ID %d のタグは存在しません", tagID))
		}
		return jsonError(opts, "INTERNAL_ERROR", fmt.Sprintf("タグの削除に失敗しました: %v", err))
	}

	output := ui.RenderTagDeleted(targetTag.ID, targetTag.Name)
	fmt.Fprint(opts.writer(), output)
	return nil
}
