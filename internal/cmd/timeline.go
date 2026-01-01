package cmd

import (
	"fmt"
	"time"

	"worklog/internal/project"
	"worklog/internal/ui"
)

// handleTimeline は本日のタイムライン表示を処理する
func handleTimeline(manager project.ProjectManager) error {
	summaries, err := manager.List()
	if err != nil {
		return err
	}

	// タイムラインを表示
	output := ui.RenderTimeline(summaries, time.Now())
	fmt.Print(output)
	return nil
}
