// Package ui はワークログアプリケーションの表示ロジックを担当する
// ボックス描画文字を使用した構造的な表示を提供
package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/mattn/go-runewidth"
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

	// プログレス用
	BlockFull  = "█"
	BlockDark  = "▓"
	BlockMid   = "▒"
	BlockLight = "░"

	// タイムライン用（Geist Font 1.5.0対応）
	BlockSquare          = "■" // U+25A0 Black Square
	BlockCrosshatch      = "▦" // U+25A6 Square with Orthogonal Crosshatch Fill
	BlockDiagonalPattern = "▨" // U+25A8 Square with Upper Right to Lower Left Fill
	MiddleDot            = "·" // U+00B7 Middle Dot (idle用)

	// インジケーター
	Arrow  = "▸"
	Bullet = "•"
	Square = "■" // BlockSquareと同じ

	// セッション境界マーカー
	SessionStart = "▶" // U+25B6 セッション開始
	SessionEnd   = "◀" // U+25C0 セッション終了
)

// ANSIカラーコード（タイムライン用）
const (
	ColorReset   = "\033[0m"
	ColorRed     = "\033[31m"
	ColorGreen   = "\033[32m"
	ColorYellow  = "\033[33m"
	ColorBlue    = "\033[34m"
	ColorMagenta = "\033[35m"
	ColorCyan    = "\033[36m"
	ColorWhite   = "\033[37m"
	ColorGray    = "\033[90m"
)

// プロジェクト用カラーパレット（高視認性）
var ProjectColors = []string{
	ColorCyan,    // プロジェクト1: シアン
	ColorMagenta, // プロジェクト2: マゼンタ
	ColorYellow,  // プロジェクト3: 黄色
	ColorGreen,   // プロジェクト4: 緑
	ColorBlue,    // プロジェクト5: 青
	ColorRed,     // プロジェクト6: 赤
}

// FormatDuration は時間を "Xh XXm" 形式にフォーマットする
func FormatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	return fmt.Sprintf("%dh %02dm", hours, minutes)
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
