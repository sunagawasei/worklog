package ui

import (
	"fmt"
	"strings"
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
