package cmd

import (
	"errors"
	"fmt"
	"os"
	"time"

	"worklog/internal/project"
	"worklog/internal/ui"
)

// handleStop は現在のプロジェクトを停止する
func handleStop(manager project.ProjectManager) error {
	// 先に現在のプロジェクト情報を取得
	status, err := manager.Status()
	if err != nil {
		return err
	}
	if status == nil {
		return errors.New("稼働中のプロジェクトがありません")
	}

	var stopTime time.Time

	// 時刻指定があるかチェック（3番目の引数）
	if len(os.Args) >= 3 && len(os.Args[2]) > 0 {
		// コマンドライン引数で時刻が指定された場合
		timestamp, err := parseTimeArg(os.Args[2])
		if err != nil {
			return fmt.Errorf("時刻の形式が不正です: %w", err)
		}
		stopTime = timestamp
		if err := manager.StopAt(timestamp); err != nil {
			return err
		}
	} else {
		// 引数なしの場合は対話モードで時刻入力
		promptUI := ui.NewPromptUI()
		timeStr, err := promptUI.InputTime("停止時刻")
		if err != nil {
			return fmt.Errorf("時刻の入力に失敗: %w", err)
		}

		if timeStr != "" {
			// 時刻が入力された場合
			timestamp, err := parseTimeArg(timeStr)
			if err != nil {
				return fmt.Errorf("時刻の形式が不正です: %w", err)
			}
			stopTime = timestamp
			if err := manager.StopAt(timestamp); err != nil {
				return err
			}
		} else {
			// 空欄の場合は現在時刻を使用
			stopTime = time.Now()
			if err := manager.StopAt(stopTime); err != nil {
				return err
			}
		}
	}

	// 新しいレンダラーを使用して表示
	output := ui.RenderStopMessage(status.Project, status.StartTime, stopTime)
	fmt.Print(output)
	return nil
}
