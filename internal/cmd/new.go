package cmd

import (
	"fmt"
	"strconv"
	"time"

	"worklog/internal/project"
	"worklog/internal/storage"
	"worklog/internal/ui"
)

// handleNew は新規プロジェクトを開始する
func handleNew(manager project.ProjectManager, opts ExecOptions) error {
	var projectName, tagID, tagName string
	var err error

	// opts.Args: [0]="new", [1]=プロジェクト名, [2]=タグID, [3]=時刻(任意)
	if len(opts.Args) >= 3 {
		projectName = opts.Args[1]
		tagID = opts.Args[2]
		if tags, err := manager.GetTags(); err == nil {
			tagName = resolveTagName(tags, tagID)
		}
	} else if opts.NoInteractive {
		return jsonError(opts, "MISSING_ARGUMENTS", "プロジェクト名とタグIDを指定してください\n使い方: worklog new <プロジェクト名> <タグID> [HH:MM]")
	} else {
		// 対話モード
		promptUI := ui.NewPromptUI()

		projectName, err = promptUI.InputProject()
		if err != nil {
			return fmt.Errorf("プロジェクト名の入力に失敗: %w", err)
		}

		tags, err := manager.GetTags()
		if err != nil {
			return fmt.Errorf("タグの読み込みに失敗: %w", err)
		}

		tagID, err = promptUI.SelectTag(tags)
		if err != nil {
			return fmt.Errorf("タグの選択に失敗: %w", err)
		}
		tagName = resolveTagName(tags, tagID)

		timeStr, err := promptUI.InputTime("開始時刻")
		if err != nil {
			return fmt.Errorf("時刻の入力に失敗: %w", err)
		}

		status, err := manager.Status()
		if err != nil {
			return err
		}

		var startTime time.Time
		if timeStr != "" {
			timestamp, err := parseTimeArg(timeStr)
			if err != nil {
				return fmt.Errorf("時刻の形式が不正です: %w", err)
			}
			startTime = timestamp
			if err := manager.NewAt(projectName, tagID, timestamp); err != nil {
				return err
			}
		} else {
			startTime = time.Now()
			if err := manager.NewAt(projectName, tagID, startTime); err != nil {
				return err
			}
		}

		tagDisplay := formatTagDisplay(tagID, tagName)
		var oldProject string
		var oldStartTime time.Time
		if status != nil {
			oldProject = status.Project
			oldStartTime = status.StartTime
		}
		output := ui.RenderSwitchMessage(oldProject, oldStartTime, startTime, projectName, tagDisplay)
		fmt.Fprint(opts.writer(), output)
		return nil
	}

	// コマンドライン引数での実行
	if err := storage.ValidateProjectName(projectName); err != nil {
		return jsonError(opts, "INVALID_PROJECT_NAME", err.Error())
	}
	if _, err := strconv.Atoi(tagID); err != nil {
		return jsonError(opts, "INVALID_TAG_ID", fmt.Sprintf("タグIDは数値で指定してください: %s", tagID))
	}

	status, err := manager.Status()
	if err != nil {
		return jsonError(opts, "INTERNAL_ERROR", err.Error())
	}

	var startTime time.Time
	if len(opts.Args) >= 4 && len(opts.Args[3]) > 0 {
		timestamp, err := parseTimeArg(opts.Args[3])
		if err != nil {
			return jsonError(opts, "INVALID_TIME_FORMAT", fmt.Sprintf("時刻の形式が不正です: %v", err))
		}
		startTime = timestamp
		if err := manager.NewAt(projectName, tagID, timestamp); err != nil {
			return jsonError(opts, "INTERNAL_ERROR", err.Error())
		}
	} else {
		startTime = time.Now()
		if err := manager.NewAt(projectName, tagID, startTime); err != nil {
			return jsonError(opts, "INTERNAL_ERROR", err.Error())
		}
	}

	if opts.JSONMode {
		out := actionJSON{
			Action:    "new",
			Project:   projectName,
			TagID:     tagID,
			TagName:   tagName,
			StartTime: startTime.Format(time.RFC3339),
		}
		if status != nil {
			out.PrevProject = status.Project
		}
		writeJSON(opts.writer(), out)
		return nil
	}

	tagDisplay := formatTagDisplay(tagID, tagName)
	var oldProject string
	var oldStartTime time.Time
	if status != nil {
		oldProject = status.Project
		oldStartTime = status.StartTime
	}
	output := ui.RenderSwitchMessage(oldProject, oldStartTime, startTime, projectName, tagDisplay)
	fmt.Fprint(opts.writer(), output)
	return nil
}
