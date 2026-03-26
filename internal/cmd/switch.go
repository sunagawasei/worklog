package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mattn/go-runewidth"
	"worklog/internal/project"
	"worklog/internal/ui"
)

// handleSwitch はプロジェクトを切り替える
func handleSwitch(manager project.ProjectManager) error {
	var newProject, newTag, newTagName string

	// 引数が十分にある場合は従来の動作
	if len(os.Args) >= 4 {
		newProject = os.Args[2]
		newTag = os.Args[3]
		// タグ名を取得
		if tags, err := manager.GetTags(); err == nil {
			newTagName = resolveTagName(tags, newTag)
		}
	} else {
		// 対話モード：過去2週間のプロジェクトリストから選択
		// 過去2週間のプロジェクト一覧を取得
		summaries, err := manager.ListRecent(14)
		if err != nil {
			return fmt.Errorf("プロジェクト一覧の取得に失敗: %w", err)
		}

		if len(summaries) == 0 {
			fmt.Fprint(os.Stderr, ui.RenderError("過去2週間に作業したプロジェクトがありません"))
			return nil
		}

		// 現在稼働中のプロジェクトを取得
		status, err := manager.Status()
		if err != nil {
			return err
		}

		// UIインスタンスを作成
		promptUI := ui.NewPromptUI()

		// dateGroup は同じ日付ラベルのプロジェクト群
		type dateGroup struct {
			label    string
			projects []ui.ProjectDisplay
		}

		// formatDateLabel の結果ごとにグループ化（順序保持）
		var groups []dateGroup
		groupIndex := map[string]int{}

		for _, summary := range summaries {
			// 現在稼働中のプロジェクトは除外
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

		// 各グループの先頭項目にDateLabelプレフィックスを設定（固定幅: ui.DatePrefixWidth）
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
			fmt.Fprint(os.Stderr, ui.RenderError("切り替え可能なプロジェクトがありません"))
			return nil
		}

		// プロジェクトを選択
		selected, err := promptUI.SelectProjectFromList(selectableProjects)
		if err != nil {
			return fmt.Errorf("プロジェクトの選択に失敗: %w", err)
		}

		newProject = selected.Project
		newTag = selected.Tag
		newTagName = selected.TagName

		// 対話モードで時刻を入力
		timeStr, err := promptUI.InputTime("切替時刻")
		if err != nil {
			return fmt.Errorf("時刻の入力に失敗: %w", err)
		}

		// 切り替え前の状態を保存
		var oldProject string
		var oldStartTime time.Time
		if status != nil {
			oldProject = status.Project
			oldStartTime = status.StartTime
		}

		var switchTime time.Time
		// 時刻が入力された場合
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
			// 空欄の場合は現在時刻を使用
			switchTime = time.Now()
			if err := manager.SwitchAt(newProject, newTag, switchTime); err != nil {
				return err
			}
		}

		// 統一された出力
		tagDisplay := formatTagDisplay(newTag, newTagName)
		output := ui.RenderSwitchMessage(oldProject, oldStartTime, switchTime, newProject, tagDisplay)
		fmt.Print(output)
		return nil
	}

	// コマンドライン引数での実行
	// 現在稼働中のプロジェクトを取得（停止メッセージ表示用）
	status, err := manager.Status()
	if err != nil {
		return err
	}

	// 切り替え前の状態を保存
	var oldProject string
	var oldStartTime time.Time
	if status != nil {
		oldProject = status.Project
		oldStartTime = status.StartTime
	}

	var switchTime time.Time
	// 時刻指定があるかチェック（5番目の引数）
	if len(os.Args) >= 5 && len(os.Args[4]) > 0 {
		timestamp, err := parseTimeArg(os.Args[4])
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

	// 統一された出力
	tagDisplay := formatTagDisplay(newTag, newTagName)
	output := ui.RenderSwitchMessage(oldProject, oldStartTime, switchTime, newProject, tagDisplay)
	fmt.Print(output)
	return nil
}
