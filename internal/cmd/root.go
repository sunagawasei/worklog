// Package cmd はワークログアプリケーションのCLIコマンド処理を担当する
// コマンドのパース、実行、ユーザーインタラクション機能を提供
package cmd

import (
	"fmt"
	"os"

	"worklog/internal/project"
	"worklog/internal/ui"
)

// Execute はCLIコマンドを実行する（依存性注入版）
func Execute(manager project.ProjectManager) error {
	// 引数なしの場合はヘルプを表示
	if len(os.Args) < 2 {
		showHelp()
		return nil
	}

	// -h または --help フラグのチェック
	if os.Args[1] == "-h" || os.Args[1] == "--help" {
		showHelp()
		return nil
	}

	// コマンドに応じて処理を分岐
	command := os.Args[1]
	switch command {
	case "status":
		return handleStatus(manager)
	case "new":
		return handleNew(manager)
	case "stop":
		return handleStop(manager)
	case "switch":
		return handleSwitch(manager)
	case "list":
		return handleList(manager)
	case "timeline":
		return handleTimeline(manager)
	case "help":
		showHelp()
		return nil
	default:
		errorMsg := fmt.Sprintf("不明なコマンド: %s", command)
		output := ui.RenderError(errorMsg)
		fmt.Fprint(os.Stderr, output)
		return fmt.Errorf("不明なコマンド: %s", command)
	}
}
