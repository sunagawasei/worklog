package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mattn/go-runewidth"
	"worklog/internal/project"
	"worklog/internal/storage"
	"worklog/internal/ui"
)

// handleSwitch はプロジェクトを切り替える
func handleSwitch(manager project.ProjectManager, opts ExecOptions) error {
	var newProject, newTag, newTagName string

	// opts.Args: [0]="switch", [1]=プロジェクト名, [2]=タグID, [3]=時刻(任意)
	if len(opts.Args) >= 3 {
		newProject = opts.Args[1]
		newTag = opts.Args[2]
		if tags, err := manager.GetTags(); err == nil {
			newTagName = resolveTagName(tags, newTag)
		}
	} else if opts.NoInteractive {
		return jsonError(opts, "MISSING_ARGUMENTS", "プロジェクト名とタグIDを指定してください\n使い方: worklog switch <プロジェクト名> <タグID> [HH:MM]")
	} else {
		// 対話モード：過去2週間のプロジェクトリストから選択
		summaries, err := manager.ListRecent(14)
		if err != nil {
			return fmt.Errorf("プロジェクト一覧の取得に失敗: %w", err)
		}

		if len(summaries) == 0 {
			fmt.Fprint(opts.writer(), ui.RenderError("過去2週間に作業したプロジェクトがありません"))
			return nil
		}

		status, err := manager.Status()
		if err != nil {
			return err
		}

		promptUI := ui.NewPromptUI()

		type dateGroup struct {
			label    string
			projects []ui.ProjectDisplay
		}
		var groups []dateGroup
		groupIndex := map[string]int{}

		for _, summary := range summaries {
			if status != nil && summary.Project == status.Project {
				continue
			}

			timeStr := ui.FormatDuration(summary.TotalTime)
			dateLabel := formatDateLabel(summary.LastActivity)

			pd := ui.ProjectDisplay{
				Project:   summary.Project,
				Tag:       summary.Tag,
				TagName:   summary.TagName,
				Time:      timeStr,
				Status:    "▫ paused",
				IsRunning: false,
			}

			if idx, ok := groupIndex[dateLabel]; ok {
				groups[idx].projects = append(groups[idx].projects, pd)
			} else {
				groupIndex[dateLabel] = len(groups)
				groups = append(groups, dateGroup{label: dateLabel, projects: []ui.ProjectDisplay{pd}})
			}
		}

		var selectableProjects []ui.ProjectDisplay
		for _, g := range groups {
			for i, pd := range g.projects {
				if i == 0 {
					padding := ui.DatePrefixWidth - runewidth.StringWidth(g.label)
					pd.DateLabel = g.label + strings.Repeat(" ", padding)
				} else {
					pd.DateLabel = strings.Repeat(" ", ui.DatePrefixWidth)
				}
				selectableProjects = append(selectableProjects, pd)
			}
		}

		if len(selectableProjects) == 0 {
			fmt.Fprint(opts.writer(), ui.RenderError("切り替え可能なプロジェクトがありません"))
			return nil
		}

		selected, err := promptUI.SelectProjectFromList(selectableProjects)
		if err != nil {
			return fmt.Errorf("プロジェクトの選択に失敗: %w", err)
		}

		newProject = selected.Project
		newTag = selected.Tag
		newTagName = selected.TagName

		timeStr, err := promptUI.InputTime("切替時刻")
		if err != nil {
			return fmt.Errorf("時刻の入力に失敗: %w", err)
		}

		var oldProject string
		var oldStartTime time.Time
		if status != nil {
			oldProject = status.Project
			oldStartTime = status.StartTime
		}

		var switchTime time.Time
		if timeStr != "" {
			timestamp, err := parseTimeArg(timeStr)
			if err != nil {
				return fmt.Errorf("時刻の形式が不正です: %w", err)
			}
			switchTime = timestamp
			if err := manager.SwitchAt(newProject, newTag, timestamp); err != nil {
				return err
			}
		} else {
			switchTime = time.Now()
			if err := manager.SwitchAt(newProject, newTag, switchTime); err != nil {
				return err
			}
		}

		tagDisplay := formatTagDisplay(newTag, newTagName)
		output := ui.RenderSwitchMessage(oldProject, oldStartTime, switchTime, newProject, tagDisplay)
		fmt.Fprint(opts.writer(), output)
		return nil
	}

	// コマンドライン引数での実行
	if err := storage.ValidateProjectName(newProject); err != nil {
		return jsonError(opts, "INVALID_PROJECT_NAME", err.Error())
	}
	if _, err := strconv.Atoi(newTag); err != nil {
		return jsonError(opts, "INVALID_TAG_ID", fmt.Sprintf("タグIDは数値で指定してください: %s", newTag))
	}

	status, err := manager.Status()
	if err != nil {
		return jsonError(opts, "INTERNAL_ERROR", err.Error())
	}

	var oldProject string
	var oldStartTime time.Time
	if status != nil {
		oldProject = status.Project
		oldStartTime = status.StartTime
	}

	var switchTime time.Time
	if len(opts.Args) >= 4 && len(opts.Args[3]) > 0 {
		timestamp, err := parseTimeArg(opts.Args[3])
		if err != nil {
			return jsonError(opts, "INVALID_TIME_FORMAT", fmt.Sprintf("時刻の形式が不正です: %v", err))
		}
		switchTime = timestamp
		if err := manager.SwitchAt(newProject, newTag, timestamp); err != nil {
			return jsonError(opts, "INTERNAL_ERROR", err.Error())
		}
	} else {
		switchTime = time.Now()
		if err := manager.SwitchAt(newProject, newTag, switchTime); err != nil {
			return jsonError(opts, "INTERNAL_ERROR", err.Error())
		}
	}

	if opts.JSONMode {
		out := actionJSON{
			Action:    "switch",
			Project:   newProject,
			TagID:     newTag,
			TagName:   newTagName,
			StartTime: switchTime.Format(time.RFC3339),
		}
		if oldProject != "" {
			out.PrevProject = oldProject
		}
		writeJSON(opts.writer(), out)
		return nil
	}

	tagDisplay := formatTagDisplay(newTag, newTagName)
	output := ui.RenderSwitchMessage(oldProject, oldStartTime, switchTime, newProject, tagDisplay)
	fmt.Fprint(opts.writer(), output)
	return nil
}
