package cmd

import (
	"fmt"
	"time"

	"worklog/internal/project"
	"worklog/internal/ui"
)

// handleStop は現在のプロジェクトを停止する
func handleStop(manager project.ProjectManager, opts ExecOptions) error {
	status, err := manager.Status()
	if err != nil {
		return jsonError(opts, "INTERNAL_ERROR", err.Error())
	}
	if status == nil {
		return jsonError(opts, "NO_ACTIVE_PROJECT", "稼働中のプロジェクトがありません")
	}

	var stopTime time.Time

	// opts.Args: [0]="stop", [1]=時刻(任意)
	if len(opts.Args) >= 2 && len(opts.Args[1]) > 0 {
		timestamp, err := parseTimeArg(opts.Args[1])
		if err != nil {
			return jsonError(opts, "INVALID_TIME_FORMAT", fmt.Sprintf("時刻の形式が不正です: %v", err))
		}
		stopTime = timestamp
		if err := manager.StopAt(timestamp); err != nil {
			return jsonError(opts, "INTERNAL_ERROR", err.Error())
		}
	} else if opts.NoInteractive {
		stopTime = time.Now()
		if err := manager.StopAt(stopTime); err != nil {
			return jsonError(opts, "INTERNAL_ERROR", err.Error())
		}
	} else {
		// 対話モード
		promptUI := ui.NewPromptUI()
		timeStr, err := promptUI.InputTime("停止時刻")
		if err != nil {
			return fmt.Errorf("時刻の入力に失敗: %w", err)
		}

		if timeStr != "" {
			timestamp, err := parseTimeArg(timeStr)
			if err != nil {
				return fmt.Errorf("時刻の形式が不正です: %w", err)
			}
			stopTime = timestamp
			if err := manager.StopAt(timestamp); err != nil {
				return err
			}
		} else {
			stopTime = time.Now()
			if err := manager.StopAt(stopTime); err != nil {
				return err
			}
		}
	}

	if opts.JSONMode {
		out := actionJSON{
			Action:   "stop",
			Project:  status.Project,
			TagID:    status.Tag,
			TagName:  status.TagName,
			StopTime: stopTime.Format(time.RFC3339),
		}
		writeJSON(opts.writer(), out)
		return nil
	}

	output := ui.RenderStopMessage(status.Project, status.StartTime, stopTime)
	fmt.Fprint(opts.writer(), output)
	return nil
}
