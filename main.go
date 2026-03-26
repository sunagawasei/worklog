package main

import (
	"fmt"
	"os"
	"worklog/internal/cmd"
	"worklog/internal/config"
	"worklog/internal/project"
	"worklog/internal/storage"
)

func main() {
	os.Exit(run())
}

func run() int {
	// 依存性の組み立て（Dependency Injection Container的な役割）
	dataDir := config.GetDataDir()
	tagsFile := config.GetTagsFile()

	currentStorage := storage.NewCurrentStorage(dataDir)
	logStorage := storage.NewLogStorage(dataDir)
	tagStorage := storage.NewTagStorage(tagsFile)

	// ProjectManagerを作成
	manager := project.NewProjectManager(currentStorage, logStorage, tagStorage)

	// グローバルフラグを抽出してExecOptionsを構築
	opts := cmd.ParseGlobalFlags(os.Args[1:])
	opts.Writer = os.Stdout

	// コマンドを実行（依存性を注入）
	if err := cmd.Execute(manager, opts); err != nil {
		if _, ok := err.(cmd.HandledError); !ok {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		return 1
	}
	return 0
}
