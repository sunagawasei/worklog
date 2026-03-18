package ui

import (
	"fmt"
	"strings"
	"time"
)

// RenderProgressBar はパーセンテージからプログレスバーを生成する
// percent: 0.0-1.0の範囲
// width: バーの幅（文字数）
func RenderProgressBar(percent float64, width int) string {
	// パーセンテージを0.0-1.0の範囲に制限
	if percent < 0.0 {
		percent = 0.0
	}
	if percent > 1.0 {
		percent = 1.0
	}

	// 埋める文字数を計算
	filled := int(percent * float64(width))
	empty := width - filled

	// プログレスバーを構築
	bar := strings.Repeat(BlockFull, filled) + strings.Repeat(BlockLight, empty)

	// パーセンテージ表示（整数）
	percentInt := int(percent * 100)

	return fmt.Sprintf("%s  %d%%", bar, percentInt)
}

// RenderTimeProgressBar は作業時間をプログレスバーとして表示する
// 形式: ████░░░░  2h/8h (25%)
// current: 現在の作業時間、totalGoal: 目標時間（通常8h）、width: バーの幅
func RenderTimeProgressBar(current time.Duration, totalGoal time.Duration, width int) string {
	if totalGoal <= 0 {
		return RenderProgressBar(0.0, width)
	}

	percent := float64(current) / float64(totalGoal)
	bar := renderBarOnly(percent, width)

	// 時間表示（1時間未満は分単位）
	currentH := int(current.Hours())
	currentM := int(current.Minutes()) % 60
	goalH := int(totalGoal.Hours())

	var currentStr string
	if currentH == 0 {
		currentStr = fmt.Sprintf("%dm", currentM)
	} else {
		currentStr = fmt.Sprintf("%dh %02dm", currentH, currentM)
	}

	percentInt := int(percent * 100)
	if percentInt > 100 {
		percentInt = 100
	}

	return fmt.Sprintf("%s  %s/%dh (%d%%)", bar, currentStr, goalH, percentInt)
}

// renderBarOnly はパーセンテージからバー文字列のみを生成する（数値なし）
func renderBarOnly(percent float64, width int) string {
	if percent < 0.0 {
		percent = 0.0
	}
	if percent > 1.0 {
		percent = 1.0
	}
	filled := int(percent * float64(width))
	empty := width - filled
	return strings.Repeat(BlockFull, filled) + strings.Repeat(BlockLight, empty)
}

// RenderProgress はラベルとプログレスバーを組み合わせて表示する
// label: 処理名
// current: 現在の進捗数
// total: 全体数
func RenderProgress(label string, current, total int) string {
	// パーセンテージを計算
	var percent float64
	if total > 0 {
		percent = float64(current) / float64(total)
	} else {
		percent = 0.0
	}

	// プログレスバーを生成（幅16文字）
	bar := RenderProgressBar(percent, 16)

	return fmt.Sprintf("%s %s", label, bar)
}
