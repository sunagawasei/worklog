package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// formatTagDisplay はタグの表示形式をフォーマットする
// タグ名がある場合: "タグ名 (ID)"
// タグ名がない場合: "ID"
func formatTagDisplay(tagID, tagName string) string {
	if tagName != "" {
		return fmt.Sprintf("%s (%s)", tagName, tagID)
	}
	return tagID
}

// parseTimeArg は時刻文字列(HH:MM または HHMM 形式)を今日の日付と組み合わせてtime.Timeに変換する
func parseTimeArg(timeStr string) (time.Time, error) {
	if timeStr == "" {
		return time.Time{}, fmt.Errorf("時刻が指定されていません")
	}

	var hour, minute int
	var err error

	// HHMM形式（4桁の数字）かチェック
	if len(timeStr) == 4 {
		// 全て数字かチェック
		if _, err := strconv.Atoi(timeStr); err == nil {
			// 最初の2桁を時間、後の2桁を分として抽出
			hour, _ = strconv.Atoi(timeStr[0:2])
			minute, _ = strconv.Atoi(timeStr[2:4])
		} else {
			return time.Time{}, fmt.Errorf("時刻は HH:MM または HHMM 形式で指定してください")
		}
	} else {
		// HH:MM形式をパース
		parts := strings.Split(timeStr, ":")
		if len(parts) != 2 {
			return time.Time{}, fmt.Errorf("時刻は HH:MM または HHMM 形式で指定してください")
		}

		hour, err = strconv.Atoi(parts[0])
		if err != nil {
			return time.Time{}, fmt.Errorf("時間の形式が不正です: %s", parts[0])
		}

		minute, err = strconv.Atoi(parts[1])
		if err != nil {
			return time.Time{}, fmt.Errorf("分の形式が不正です: %s", parts[1])
		}
	}

	// 時刻の妥当性チェック
	if hour < 0 || hour > 23 {
		return time.Time{}, fmt.Errorf("時間は0-23の範囲で指定してください: %d", hour)
	}
	if minute < 0 || minute > 59 {
		return time.Time{}, fmt.Errorf("分は0-59の範囲で指定してください: %d", minute)
	}

	// 今日の日付と組み合わせる
	now := time.Now()
	result := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())

	return result, nil
}

// parseDateArg は日付文字列(YYYY-MM-DD または YYYYMMDD 形式)をtime.Timeに変換する
// parseRelativeDate は相対日付形式(-1d, -2d等)をパースする
// 対応しない形式の場合は nil, nil を返す（他のパーサーに処理を委譲）
// 形式は正しいが値が不正な場合はエラーを返す
func parseRelativeDate(dateStr string) (*time.Time, error) {
	var num int
	var unit string

	// fmt.Sscanf で "-数値単位" 形式をパース
	n, err := fmt.Sscanf(dateStr, "-%d%s", &num, &unit)
	if n != 2 || err != nil {
		return nil, nil // 非マッチ
	}

	if unit != "d" {
		return nil, nil // 今は "d" のみサポート
	}

	// N日前を計算
	target := time.Now().AddDate(0, 0, -num)
	result := time.Date(target.Year(), target.Month(), target.Day(), 0, 0, 0, 0, time.Local)
	return &result, nil
}

func parseDateArg(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, fmt.Errorf("日付が指定されていません")
	}

	// 相対日付形式（-1d, -2d等）をまずチェック
	if relDate, err := parseRelativeDate(dateStr); err != nil {
		return time.Time{}, err
	} else if relDate != nil {
		return *relDate, nil
	}

	var parsedDate time.Time
	var err error

	// YYYYMMDD形式（8桁の数字）かチェック
	if len(dateStr) == 8 {
		// 全て数字かチェック
		if _, parseErr := strconv.Atoi(dateStr); parseErr == nil {
			// YYYYMMDD形式としてパース
			parsedDate, err = time.ParseInLocation("20060102", dateStr, time.Local)
		} else {
			return time.Time{}, fmt.Errorf("日付は YYYY-MM-DD または YYYYMMDD 形式で指定してください")
		}
	} else {
		// YYYY-MM-DD形式をパース
		parsedDate, err = time.ParseInLocation("2006-01-02", dateStr, time.Local)
	}

	if err != nil {
		return time.Time{}, fmt.Errorf("日付の形式が不正です: %w", err)
	}

	return parsedDate, nil
}

// formatDateLabel は最終作業日から相対的な日付ラベルを生成
func formatDateLabel(lastActivity time.Time) string {
	now := time.Now()

	// 日付部分だけで比較（時刻を無視）
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	activityDate := time.Date(lastActivity.Year(), lastActivity.Month(), lastActivity.Day(), 0, 0, 0, 0, lastActivity.Location())

	daysDiff := int(today.Sub(activityDate).Hours() / 24)

	if daysDiff == 0 {
		return "[Today]"
	} else if daysDiff == 1 {
		return "[Yesterday]"
	} else if daysDiff < 7 {
		return fmt.Sprintf("[%d days ago]", daysDiff)
	} else if daysDiff < 14 {
		weeks := daysDiff / 7
		if weeks == 1 {
			return "[1 week ago]"
		}
		return fmt.Sprintf("[%d weeks ago]", weeks)
	}

	return ""
}
