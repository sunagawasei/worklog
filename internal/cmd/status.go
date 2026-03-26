package cmd

import (
	"fmt"
	"time"

	"worklog/internal/project"
	"worklog/internal/ui"
)

// handleStatus は現在の稼働状況を表示する
func handleStatus(manager project.ProjectManager, opts ExecOptions) error {
	status, err := manager.Status()
	if err != nil {
		return jsonError(opts, "INTERNAL_ERROR", err.Error())
	}

	summaries, err := manager.List()
	if err != nil {
		return jsonError(opts, "INTERNAL_ERROR", err.Error())
	}

	if opts.JSONMode {
		out := summaryStatusJSON{
			Summaries: make([]listItemJSON, 0, len(summaries)),
		}
		if status != nil {
			s := toStatusJSON(status)
			out.Status = &s
		}
		for _, s := range summaries {
			out.Summaries = append(out.Summaries, toListItemJSON(s))
		}
		writeJSON(opts.writer(), out)
		return nil
	}

	output := ui.RenderDashboard(status, summaries, time.Now())
	fmt.Fprint(opts.writer(), output)
	return nil
}
