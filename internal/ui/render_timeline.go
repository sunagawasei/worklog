package ui

import (
	"fmt"
	"strings"
	"time"

	"worklog/internal/domain"
)

// selectBlockChar はプロジェクトインデックスに応じたブロック文字を返す
func selectBlockChar(projectIndex int) string {
	chars := []string{BlockSquare, BlockCrosshatch, BlockDiagonalPattern}
	return chars[projectIndex%len(chars)]
}

// RenderTimeline は本日のプロジェクト作業をタイムライン形式で表示する
// summaries: 本日のプロジェクトサマリー一覧
// now: 現在時刻
func RenderTimeline(summaries []domain.ProjectSummary, now time.Time) string {
	var builder strings.Builder

	// ヘッダー
	builder.WriteString("Timeline\n")
	builder.WriteString(strings.Repeat("─", 44) + "\n")

	// プロジェクト名とインデックスのマップを作成
	projectIndexMap := make(map[string]int)
	for i, summary := range summaries {
		projectIndexMap[summary.Project] = i
	}

	// 9:00-20:00の各時間を描画
	for hour := 9; hour <= 20; hour++ {
		// 時刻表示
		builder.WriteString(fmt.Sprintf(" %02d:00 ", hour))

		// 1時間を36文字で表現（1文字≒100秒）
		hourStart := time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, now.Location())

		// この時間帯に作業しているプロジェクトを探す
		var blocks string
		for second := 0; second < 3600; second += 100 {
			// 100秒ごとに判定（36文字で60分を表現するため）
			checkTime := hourStart.Add(time.Duration(second) * time.Second)

			// この時刻にアクティブなプロジェクトを探す
			activeProject := ""
			for _, summary := range summaries {
				for _, tr := range summary.TimeRanges {
					if !checkTime.Before(tr.Start) && checkTime.Before(tr.End) {
						activeProject = summary.Project
						break
					}
				}
				if activeProject != "" {
					break
				}
			}

			// ブロック文字を選択
			if activeProject != "" {
				projectIndex := projectIndexMap[activeProject]
				blocks += selectBlockChar(projectIndex)
			} else {
				blocks += MiddleDot
			}
		}

		builder.WriteString(blocks)
		builder.WriteString("\n")
	}

	// 空行
	builder.WriteString("\n")

	// 凡例
	for _, summary := range summaries {
		builder.WriteString(" ")
		builder.WriteString(selectBlockChar(projectIndexMap[summary.Project]))
		builder.WriteString(" ")
		builder.WriteString(summary.Project)
		builder.WriteString("\n")
	}
	builder.WriteString(" ")
	builder.WriteString(MiddleDot)
	builder.WriteString(" idle")
	builder.WriteString("\n")

	// 下部区切り線
	builder.WriteString(strings.Repeat("─", 44) + "\n")

	return builder.String()
}
