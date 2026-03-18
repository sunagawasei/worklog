package ui

import (
	"fmt"
	"strings"
	"time"

	"worklog/internal/domain"
)

// timelineBlocks は1時間あたりのブロック数（100秒/ブロック × 36 ≒ 60分）
const timelineBlocks = 36

// currentTimeMarker は現在時刻を示すマーカー文字
const currentTimeMarker = "▼"

// selectBlockChar はプロジェクトインデックスに応じたブロック文字を返す
func selectBlockChar(projectIndex int) string {
	chars := []string{BlockSquare, BlockCrosshatch, BlockDiagonalPattern}
	return chars[projectIndex%len(chars)]
}

// timelineWidth はタイムラインの区切り線幅を計算する（端末幅に合わせる）
func timelineWidth() int {
	termWidth := GetTerminalWidth()
	// 最低幅: " HH:00 "(7) + ブロック(36) = 43
	if termWidth < 44 {
		return 44
	}
	// 最大幅は80に制限（読みやすさのため）
	if termWidth > 80 {
		return 80
	}
	return termWidth
}

// RenderTimeline は本日のプロジェクト作業をタイムライン形式で表示する
// summaries: プロジェクトサマリー一覧
// displayDate: 表示対象日付（この日のタイムラインを描画）
func RenderTimeline(summaries []domain.ProjectSummary, displayDate time.Time) string {
	var builder strings.Builder

	width := timelineWidth()
	sep := strings.Repeat("─", width)

	// 現在時刻を取得（マーカー表示判定用）
	realNow := time.Now()
	isToday := displayDate.Format("2006-01-02") == realNow.Format("2006-01-02")

	// ヘッダー（9a: 日付表示）
	dateStr := displayDate.Format("2006-01-02")
	builder.WriteString(fmt.Sprintf("Timeline  %s\n", dateStr))
	builder.WriteString(sep + "\n")

	// プロジェクト名とインデックスのマップを作成
	projectIndexMap := make(map[string]int)
	for i, summary := range summaries {
		projectIndexMap[summary.Project] = i
	}

	// 9b: 凡例をヘッダーに移動（タイムライン本体の前に表示）
	var legendParts []string
	for _, summary := range summaries {
		legendParts = append(legendParts,
			fmt.Sprintf(" %s %s", selectBlockChar(projectIndexMap[summary.Project]), summary.Project))
	}
	legendParts = append(legendParts, fmt.Sprintf(" %s idle", MiddleDot))
	builder.WriteString(strings.Join(legendParts, " "))
	builder.WriteString("\n")
	builder.WriteString(sep + "\n")

	// 9f: 時間軸ルーラー（30分刻みの目盛り）
	// " HH:00 " = 7文字のインデント後に :00 と :30 の目盛りを表示
	// :00 は列7（ブロック位置0）、:30 は列25（ブロック位置18）に表示
	ruler := strings.Repeat(" ", 7) + ":00" + strings.Repeat(" ", 15) + ":30"
	builder.WriteString(ruler + "\n")

	// 9:00-20:00の各時間を描画
	for hour := 9; hour <= 20; hour++ {
		// 時刻表示
		builder.WriteString(fmt.Sprintf(" %02d:00 ", hour))

		// 1時間を36文字で表現（100秒/文字）
		hourStart := time.Date(
			displayDate.Year(), displayDate.Month(), displayDate.Day(),
			hour, 0, 0, 0, displayDate.Location(),
		)

		// ブロック文字列を生成
		blocks := make([]string, timelineBlocks)
		for blockIdx := 0; blockIdx < timelineBlocks; blockIdx++ {
			checkTime := hourStart.Add(time.Duration(blockIdx*100) * time.Second)

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

			if activeProject != "" {
				blocks[blockIdx] = selectBlockChar(projectIndexMap[activeProject])
			} else {
				blocks[blockIdx] = MiddleDot
			}
		}

		// 9e: 現在時刻マーカー（今日の表示かつ現在時刻が该当時間内の場合）
		if isToday && realNow.Hour() == hour {
			// 現在分をブロック位置に変換
			totalSeconds := realNow.Minute()*60 + realNow.Second()
			markerPos := totalSeconds / 100
			if markerPos < timelineBlocks {
				blocks[markerPos] = currentTimeMarker
			}
		}

		builder.WriteString(strings.Join(blocks, ""))
		builder.WriteString("\n")
	}

	// 下部区切り線
	builder.WriteString(sep + "\n")

	return builder.String()
}
