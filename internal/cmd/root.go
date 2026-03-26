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
func Execute(manager project.ProjectManager, opts ExecOptions) error {
	if opts.Writer == nil {
		opts.Writer = os.Stdout
	}

	// 引数なしの場合はヘルプを表示
	if len(opts.Args) == 0 {
		if opts.JSONMode {
			return jsonError(opts, "MISSING_ARGUMENTS", "コマンドを指定してください。使い方: worklog <command>")
		}
		showHelp(opts)
		return nil
	}

	// -h または --help フラグのチェック
	if opts.Args[0] == "-h" || opts.Args[0] == "--help" {
		if opts.JSONMode {
			return jsonError(opts, "MISSING_ARGUMENTS", "コマンドを指定してください。使い方: worklog <command>")
		}
		showHelp(opts)
		return nil
	}

	// コマンドに応じて処理を分岐
	command := opts.Args[0]
	switch command {
	case "status":
		return handleStatus(manager, opts)
	case "new":
		return handleNew(manager, opts)
	case "stop":
		return handleStop(manager, opts)
	case "switch":
		return handleSwitch(manager, opts)
	case "list":
		return handleList(manager, opts)
	case "timeline":
		return handleTimeline(manager, opts)
	case "tag":
		return handleTag(manager, opts)
	case "help":
		if opts.JSONMode {
			return jsonError(opts, "MISSING_ARGUMENTS", "コマンドを指定してください。使い方: worklog <command>")
		}
		showHelp(opts)
		return nil
	default:
		if opts.JSONMode {
			writeJSONError(opts.writer(), "UNKNOWN_COMMAND", fmt.Sprintf("不明なコマンド: %s", command))
			return HandledError{}
		}
		errorMsg := fmt.Sprintf("不明なコマンド: %s", command)
		output := ui.RenderError(errorMsg)
		fmt.Fprint(os.Stderr, output)
		return fmt.Errorf("不明なコマンド: %s", command)
	}
}
