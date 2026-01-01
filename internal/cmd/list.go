package cmd

import (
	"fmt"
	"os"
	"time"

	"worklog/internal/domain"
	"worklog/internal/project"
	"worklog/internal/ui"
)

// handleList は本日のプロジェクト一覧と稼働時間を表示する
func handleList(manager project.ProjectManager) error {
	var summaries []domain.ProjectSummary
	var displayDate time.Time
	var err error

	// 日付引数があるかチェック（3番目の引数）
	if len(os.Args) >= 3 && len(os.Args[2]) > 0 {
		// 日付が指定された場合
		date, parseErr := parseDateArg(os.Args[2])
		if parseErr != nil {
			return fmt.Errorf("日付の形式が不正です: %w", parseErr)
		}
		displayDate = date
		summaries, err = manager.ListOnDate(date)
	} else {
		// 日付指定なしの場合は本日
		displayDate = time.Now()
		summaries, err = manager.List()
	}

	if err != nil {
		return err
	}

	// 新しいレンダラーを使用して表示
	output := ui.RenderList(summaries, displayDate)
	fmt.Print(output)
	return nil
}
