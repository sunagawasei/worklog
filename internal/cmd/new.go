package cmd

import (
	"fmt"
	"os"
	"time"

	"worklog/internal/project"
	"worklog/internal/ui"
)

// handleNew は新規プロジェクトを開始する
func handleNew(manager project.ProjectManager) error {
	var projectName, tagID, tagName string
	var err error

	// 引数が十分にある場合は従来の動作
	if len(os.Args) >= 4 {
		projectName = os.Args[2]
		tagID = os.Args[3]
		// タグ名を取得
		if tags, err := manager.GetTags(); err == nil {
			tagName = resolveTagName(tags, tagID)
		}
	} else {
		// 対話モード
		// UIインスタンスを作成
		promptUI := ui.NewPromptUI()

		// プロジェクト名を入力
		projectName, err = promptUI.InputProject()
		if err != nil {
			return fmt.Errorf("プロジェクト名の入力に失敗: %w", err)
		}

		// タグを選択
		tags, err := manager.GetTags()
		if err != nil {
			return fmt.Errorf("タグの読み込みに失敗: %w", err)
		}

		tagID, err = promptUI.SelectTag(tags)
		if err != nil {
			return fmt.Errorf("タグの選択に失敗: %w", err)
		}

		// 選択されたタグ名を取得
		tagName = resolveTagName(tags, tagID)

		// 対話モードで時刻を入力
		timeStr, err := promptUI.InputTime("開始時刻")
		if err != nil {
			return fmt.Errorf("時刻の入力に失敗: %w", err)
		}

		// 現在の稼働状況を取得（停止メッセージ表示用）
		status, err := manager.Status()
		if err != nil {
			return err
		}

		var startTime time.Time
		// 時刻が入力された場合
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
			// 空欄の場合は現在時刻を使用
			startTime = time.Now()
			if err := manager.NewAt(projectName, tagID, startTime); err != nil {
				return err
			}
		}

		// 停止したプロジェクトがあれば表示
		if status != nil {
			stopTime := startTime
			stopOutput := ui.RenderStopMessage(status.Project, status.StartTime, stopTime)
			fmt.Print(stopOutput)
		}

		// 統一された出力
		tagDisplay := formatTagDisplay(tagID, tagName)
		output := ui.RenderNewMessage(projectName, startTime, tagDisplay)
		fmt.Print(output)
		return nil
	}

	// コマンドライン引数での実行
	// 現在稼働中のプロジェクトを取得（停止メッセージ表示用）
	status, err := manager.Status()
	if err != nil {
		return err
	}

	var startTime time.Time
	// 時刻指定があるかチェック（5番目の引数）
	if len(os.Args) >= 5 && len(os.Args[4]) > 0 {
		timestamp, err := parseTimeArg(os.Args[4])
		if err != nil {
			return fmt.Errorf("時刻の形式が不正です: %w", err)
		}
		startTime = timestamp
		if err := manager.NewAt(projectName, tagID, timestamp); err != nil {
			return err
		}
	} else {
		startTime = time.Now()
		if err := manager.NewAt(projectName, tagID, time.Now()); err != nil {
			return err
		}
	}

	// 停止したプロジェクトがあれば表示
	if status != nil {
		stopTime := startTime
		stopOutput := ui.RenderStopMessage(status.Project, status.StartTime, stopTime)
		fmt.Print(stopOutput)
	}

	// 統一された出力
	tagDisplay := formatTagDisplay(tagID, tagName)
	output := ui.RenderNewMessage(projectName, startTime, tagDisplay)
	fmt.Print(output)
	return nil
}
