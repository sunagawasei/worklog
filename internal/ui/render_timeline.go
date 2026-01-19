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

// getProjectColor はプロジェクトインデックスに応じた色を返す
func getProjectColor(projectIndex int) string {
	return ProjectColors[projectIndex%len(ProjectColors)]
}

// coloredBlock は色付きブロック文字を返す
func coloredBlock(projectIndex int, char string) string {
	return getProjectColor(projectIndex) + char + ColorReset
}

// sessionPosition はセッション内の位置を表す
type sessionPosition int

const (
	positionNone   sessionPosition = iota // セッション外
	positionStart                         // セッション開始
	positionMiddle                        // セッション中間
	positionEnd                           // セッション終了
	positionSingle                        // 1ブロックのみのセッション
)

// findActiveSession は指定時刻でアクティブなプロジェクトとセッション位置を返す
func findActiveSession(checkTime time.Time, summaries []domain.ProjectSummary, projectIndexMap map[string]int, sampleInterval time.Duration) (string, int, sessionPosition) {
	for _, summary := range summaries {
		for _, tr := range summary.TimeRanges {
			if !checkTime.Before(tr.Start) && checkTime.Before(tr.End) {
				projectIndex := projectIndexMap[summary.Project]
				pos := getSessionPosition(checkTime, tr, sampleInterval)
				return summary.Project, projectIndex, pos
			}
		}
	}
	return "", -1, positionNone
}

// getSessionPosition はセッション内での位置を判定する
func getSessionPosition(checkTime time.Time, tr domain.TimeRange, sampleInterval time.Duration) sessionPosition {
	isStart := checkTime.Sub(tr.Start) < sampleInterval
	isEnd := tr.End.Sub(checkTime) <= sampleInterval

	switch {
	case isStart && isEnd:
		return positionSingle
	case isStart:
		return positionStart
	case isEnd:
		return positionEnd
	default:
		return positionMiddle
	}
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

	const sampleInterval = 100 * time.Second

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

			// この時刻にアクティブなセッションを探す
			_, projectIndex, pos := findActiveSession(checkTime, summaries, projectIndexMap, sampleInterval)

			// ブロック文字を選択（セッション位置に応じた表示）
			switch pos {
			case positionNone:
				blocks += ColorGray + MiddleDot + ColorReset
			case positionStart, positionSingle:
				// セッション開始時は開始マーカー（色付き）
				blocks += coloredBlock(projectIndex, SessionStart)
			case positionMiddle, positionEnd:
				// セッション中間・終了はブロック文字（色付き）
				blocks += coloredBlock(projectIndex, BlockFull)
			}
		}

		builder.WriteString(blocks)
		builder.WriteString("\n")
	}

	// 空行
	builder.WriteString("\n")

	// 凡例（色付き）
	builder.WriteString(" ")
	for i, summary := range summaries {
		if i > 0 {
			builder.WriteString("  ")
		}
		builder.WriteString(coloredBlock(i, SessionStart))
		builder.WriteString(coloredBlock(i, BlockFull))
		builder.WriteString(" ")
		builder.WriteString(summary.Project)
	}
	if len(summaries) > 0 {
		builder.WriteString("  ")
	}
	builder.WriteString(ColorGray + MiddleDot + ColorReset)
	builder.WriteString(" idle")
	builder.WriteString("\n")

	// 下部区切り線
	builder.WriteString(strings.Repeat("─", 44) + "\n")

	return builder.String()
}
