package cmd

import (
	"fmt"
	"os"
	"time"

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
		tags, err := manager.GetTags()
		if err == nil {
			id := 0
			fmt.Sscanf(newTag, "%d", &id)
			for _, t := range tags {
				if t.ID == id {
					newTagName = t.Name
					break
				}
			}
		}
	} else {
		// 対話モード：過去2週間のプロジェクトリストから選択
		// 過去2週間のプロジェクト一覧を取得
		summaries, err := manager.ListRecent(14)
		if err != nil {
			return fmt.Errorf("プロジェクト一覧の取得に失敗: %w", err)
		}

		if len(summaries) == 0 {
			fmt.Println("過去2週間に作業したプロジェクトがありません")
			fmt.Println("新規プロジェクトを開始してください")
			return nil
		}

		// 現在稼働中のプロジェクトを取得
		status, err := manager.Status()
		if err != nil {
			return err
		}

		// UIインスタンスを作成
		promptUI := ui.NewPromptUI()

		// 稼働中でないプロジェクトを今日と過去に分けてリストアップ
		var todayProjects []ui.ProjectDisplay
		var pastProjects []ui.ProjectDisplay

		for _, summary := range summaries {
			// 現在稼働中のプロジェクトは除外
			if status != nil && summary.Project == status.Project {
				continue
			}

			// 稼働時間をフォーマット
			timeStr := ui.FormatDuration(summary.TotalTime)

			// 状態アイコン（一時停止中のプロジェクト）
			statusIcon := "▫ paused"

			// 日付ラベルを生成
			dateLabel := formatDateLabel(summary.LastActivity)

			pd := ui.ProjectDisplay{
				Project:   summary.Project,
				Tag:       summary.Tag,
				TagName:   summary.TagName,
				Time:      timeStr,
				Status:    statusIcon,
				IsRunning: false,
				DateLabel: dateLabel,
			}

			// 今日と過去でグループ分け
			if dateLabel == "[Today]" {
				todayProjects = append(todayProjects, pd)
			} else {
				pastProjects = append(pastProjects, pd)
			}
		}

		// 最終的なリストを作成（今日 → セパレーター → 過去）
		var selectableProjects []ui.ProjectDisplay
		selectableProjects = append(selectableProjects, todayProjects...)

		// 両方のグループが存在する場合のみセパレーターを挿入
		if len(todayProjects) > 0 && len(pastProjects) > 0 {
			selectableProjects = append(selectableProjects, ui.ProjectDisplay{
				IsSeparator:   true,
				SeparatorText: "────────────────────────────────",
			})
		}

		selectableProjects = append(selectableProjects, pastProjects...)

		if len(selectableProjects) == 0 {
			fmt.Println("切り替え可能なプロジェクトがありません")
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
		timeStr, err := promptUI.InputTime("Switch time")
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
			if err := manager.Switch(newProject, newTag); err != nil {
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
		if err := manager.Switch(newProject, newTag); err != nil {
			return err
		}
	}

	// 統一された出力
	tagDisplay := formatTagDisplay(newTag, newTagName)
	output := ui.RenderSwitchMessage(oldProject, oldStartTime, switchTime, newProject, tagDisplay)
	fmt.Print(output)
	return nil
}
