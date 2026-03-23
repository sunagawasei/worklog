// Package ui はワークログアプリケーションの表示ロジックを担当する
// ボックス描画文字を使用した構造的な表示を提供
package ui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mattn/go-runewidth"
	"golang.org/x/term"
)

// StandardWidth はメッセージ・ヘルプ・タグ・エラー表示の統一幅
const StandardWidth = 44

// ダッシュボードのカラム幅定数
const (
	dashStatusCol  = 24 // Status列の内容幅
	dashSummaryCol = 22 // Summary列の内容幅
)

// ボックス描画文字の定数
const (
	LineH    = "─" // 水平線
	LineV    = "│" // 垂直線
	CornerTL = "┌" // 左上角
	CornerTR = "┐" // 右上角
	CornerBL = "└" // 左下角
	CornerBR = "┘" // 右下角
	CrossT   = "┬" // T字接続（上）
	CrossB   = "┴" // T字接続（下）
	CrossL   = "├" // T字接続（左）
	CrossR   = "┤" // T字接続（右）
	Cross    = "┼" // 十字接続

	// 丸角ボックス（メッセージ系で使用、ダッシュボードの直角と区別）
	RoundTL = "╭" // 丸左上角
	RoundTR = "╮" // 丸右上角
	RoundBL = "╰" // 丸左下角
	RoundBR = "╯" // 丸右下角

	// プログレス用
	BlockFull  = "█"
	BlockLight = "░"

	// タイムライン用（Geist Font 1.5.0対応）
	BlockSquare          = "■" // U+25A0 Black Square
	BlockCrosshatch      = "▦" // U+25A6 Square with Orthogonal Crosshatch Fill
	BlockDiagonalPattern = "▨" // U+25A8 Square with Upper Right to Lower Left Fill
	MiddleDot            = "·" // U+00B7 Middle Dot (idle用)

	// 二重線（重要度の高い区切りに使用）
	LineH2   = "═" // 二重水平線
	CrossB2  = "╧" // 二重水平・単一垂直のT字接続（下）

	// ドットリーダー
	DotLeader = "·" // 列揃え用ドット（MiddleDotと同じ文字）

	// インジケーター
	Arrow  = "▸"
	Bullet = "•"
)

// FormatDuration は時間を "Xh XXm" 形式にフォーマットする
func FormatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	return fmt.Sprintf("%dh %02dm", hours, minutes)
}

// FormatDurationShort は時間を短縮形式でフォーマットする
// 1時間未満の場合は "XXm"、1時間以上の場合は "Xh XXm" 形式
func FormatDurationShort(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60

	if hours == 0 {
		return fmt.Sprintf("%dm", minutes)
	}
	return fmt.Sprintf("%dh %02dm", hours, minutes)
}

// GetTerminalWidth はターミナルの幅を取得する
// 取得できない場合はデフォルト値（80）を返す
func GetTerminalWidth() int {
	fd := int(os.Stdout.Fd())
	width, _, err := term.GetSize(fd)
	if err != nil || width <= 0 {
		return 80 // デフォルト幅
	}
	return width
}

// renderSeparator は指定された幅の区切り線を生成する
func renderSeparator(width int) string {
	return strings.Repeat(LineH, width)
}

// displayWidth は文字列の表示幅を計算する（East Asian Width準拠）
// go-runewidthライブラリを使用して正確な文字幅を取得
func displayWidth(s string) int {
	return runewidth.StringWidth(s)
}

// padString は文字列を指定幅に調整する
// 文字列が長すぎる場合は切り詰め、短い場合はスペースでパディングする
func padString(s string, targetWidth int) string {
	currentWidth := displayWidth(s)

	// 文字列が長すぎる場合は切り詰める
	if currentWidth > targetWidth {
		// 省略記号（…）を含めて切り詰め
		s = runewidth.Truncate(s, targetWidth, "…")
		currentWidth = displayWidth(s)
	}

	// パディングを追加
	padding := targetWidth - currentWidth
	return s + strings.Repeat(" ", padding)
}
