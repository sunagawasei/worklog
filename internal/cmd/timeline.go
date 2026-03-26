package cmd

import (
	"fmt"
	"time"

	"worklog/internal/domain"
	"worklog/internal/project"
	"worklog/internal/ui"
)

// handleTimeline は本日のタイムライン表示を処理する
func handleTimeline(manager project.ProjectManager, opts ExecOptions) error {
	var summaries []domain.ProjectSummary
	var displayDate time.Time
	var err error

	// opts.Args[0]="timeline", opts.Args[1]=日付引数
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
			writeJSON(opts.writer(), toTimelineItemJSON(s))
		}
		return nil
	}

	output := ui.RenderTimeline(summaries, displayDate)
	fmt.Fprint(opts.writer(), output)
	return nil
}
