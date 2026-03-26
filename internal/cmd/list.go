package cmd

import (
	"fmt"
	"time"

	"worklog/internal/domain"
	"worklog/internal/project"
	"worklog/internal/ui"
)

// handleList は本日のプロジェクト一覧と稼働時間を表示する
func handleList(manager project.ProjectManager, opts ExecOptions) error {
	var summaries []domain.ProjectSummary
	var displayDate time.Time
	var err error

	// opts.Args[0]="list", opts.Args[1]=日付引数
	if len(opts.Args) >= 2 && len(opts.Args[1]) > 0 {
		date, parseErr := parseDateArg(opts.Args[1])
		if parseErr != nil {
			return jsonError(opts, "INVALID_DATE_FORMAT", fmt.Sprintf("日付の形式が不正です: %v", parseErr))
		}
		displayDate = date
		summaries, err = manager.ListOnDate(date)
	} else {
		displayDate = time.Now()
		summaries, err = manager.List()
	}

	if err != nil {
		return jsonError(opts, "INTERNAL_ERROR", err.Error())
	}

	if opts.JSONMode {
		for _, s := range summaries {
			writeJSON(opts.writer(), toListItemJSON(s))
		}
		return nil
	}

	output := ui.RenderList(summaries, displayDate)
	fmt.Fprint(opts.writer(), output)
	return nil
}
