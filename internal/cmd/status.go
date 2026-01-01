package cmd

import (
	"fmt"
	"time"

	"worklog/internal/project"
	"worklog/internal/ui"
)

// handleStatus は現在の稼働状況を表示する
func handleStatus(manager project.ProjectManager) error {
	status, err := manager.Status()
	if err != nil {
		return err
	}

	// 本日の全プロジェクトの作業時間を取得
	summaries, err := manager.List()
	if err != nil {
		return err
	}

	// ダッシュボードを表示
	output := ui.RenderDashboard(status, summaries, time.Now())
	fmt.Print(output)
	return nil
}
