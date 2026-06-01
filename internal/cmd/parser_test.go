package cmd

import (
	"fmt"
	"testing"
	"time"
)

func TestParseTimeArg(t *testing.T) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	t.Run("正常な時刻フォーマット", func(t *testing.T) {
		result, err := parseTimeArg("14:30")
		if err != nil {
			t.Errorf("parseTimeArg(\"14:30\") returned error: %v", err)
		}
		expected := today.Add(14*time.Hour + 30*time.Minute)
		if !result.Equal(expected) {
			t.Errorf("parseTimeArg(\"14:30\") = %v, want %v", result, expected)
		}
	})

	t.Run("朝の時刻", func(t *testing.T) {
		result, err := parseTimeArg("09:15")
		if err != nil {
			t.Errorf("parseTimeArg(\"09:15\") returned error: %v", err)
		}
		expected := today.Add(9*time.Hour + 15*time.Minute)
		if !result.Equal(expected) {
			t.Errorf("parseTimeArg(\"09:15\") = %v, want %v", result, expected)
		}
	})

	t.Run("深夜0時", func(t *testing.T) {
		result, err := parseTimeArg("00:00")
		if err != nil {
			t.Errorf("parseTimeArg(\"00:00\") returned error: %v", err)
		}
		if !result.Equal(today) {
			t.Errorf("parseTimeArg(\"00:00\") = %v, want %v", result, today)
		}
	})

	t.Run("23時59分", func(t *testing.T) {
		result, err := parseTimeArg("23:59")
		if err != nil {
			t.Errorf("parseTimeArg(\"23:59\") returned error: %v", err)
		}
		expected := today.Add(23*time.Hour + 59*time.Minute)
		if !result.Equal(expected) {
			t.Errorf("parseTimeArg(\"23:59\") = %v, want %v", result, expected)
		}
	})

	t.Run("不正な時刻 - 25時", func(t *testing.T) {
		_, err := parseTimeArg("25:00")
		if err == nil {
			t.Error("parseTimeArg(\"25:00\") should return error, but got nil")
		}
	})

	t.Run("不正な時刻 - 60分", func(t *testing.T) {
		_, err := parseTimeArg("14:60")
		if err == nil {
			t.Error("parseTimeArg(\"14:60\") should return error, but got nil")
		}
	})

	t.Run("HHMM形式 - 1430", func(t *testing.T) {
		result, err := parseTimeArg("1430")
		if err != nil {
			t.Errorf("parseTimeArg(\"1430\") returned error: %v", err)
		}
		expected := today.Add(14*time.Hour + 30*time.Minute)
		if !result.Equal(expected) {
			t.Errorf("parseTimeArg(\"1430\") = %v, want %v", result, expected)
		}
	})

	t.Run("HHMM形式 - 0930", func(t *testing.T) {
		result, err := parseTimeArg("0930")
		if err != nil {
			t.Errorf("parseTimeArg(\"0930\") returned error: %v", err)
		}
		expected := today.Add(9*time.Hour + 30*time.Minute)
		if !result.Equal(expected) {
			t.Errorf("parseTimeArg(\"0930\") = %v, want %v", result, expected)
		}
	})

	t.Run("HHMM形式 - 0000", func(t *testing.T) {
		result, err := parseTimeArg("0000")
		if err != nil {
			t.Errorf("parseTimeArg(\"0000\") returned error: %v", err)
		}
		if !result.Equal(today) {
			t.Errorf("parseTimeArg(\"0000\") = %v, want %v", result, today)
		}
	})

	t.Run("HHMM形式 - 2359", func(t *testing.T) {
		result, err := parseTimeArg("2359")
		if err != nil {
			t.Errorf("parseTimeArg(\"2359\") returned error: %v", err)
		}
		expected := today.Add(23*time.Hour + 59*time.Minute)
		if !result.Equal(expected) {
			t.Errorf("parseTimeArg(\"2359\") = %v, want %v", result, expected)
		}
	})

	t.Run("不正なHHMM形式 - 2500", func(t *testing.T) {
		_, err := parseTimeArg("2500")
		if err == nil {
			t.Error("parseTimeArg(\"2500\") should return error, but got nil")
		}
	})

	t.Run("不正なHHMM形式 - 1460", func(t *testing.T) {
		_, err := parseTimeArg("1460")
		if err == nil {
			t.Error("parseTimeArg(\"1460\") should return error, but got nil")
		}
	})

	t.Run("不正なフォーマット - 3桁", func(t *testing.T) {
		_, err := parseTimeArg("123")
		if err == nil {
			t.Error("parseTimeArg(\"123\") should return error, but got nil")
		}
	})

	t.Run("不正なフォーマット - 5桁", func(t *testing.T) {
		_, err := parseTimeArg("12345")
		if err == nil {
			t.Error("parseTimeArg(\"12345\") should return error, but got nil")
		}
	})

	t.Run("不正なフォーマット - 文字列", func(t *testing.T) {
		_, err := parseTimeArg("abc:def")
		if err == nil {
			t.Error("parseTimeArg(\"abc:def\") should return error, but got nil")
		}
	})

	t.Run("空文字列", func(t *testing.T) {
		_, err := parseTimeArg("")
		if err == nil {
			t.Error("parseTimeArg(\"\") should return error, but got nil")
		}
	})

	t.Run("1桁の時刻", func(t *testing.T) {
		result, err := parseTimeArg("9:5")
		if err != nil {
			t.Errorf("parseTimeArg(\"9:5\") returned error: %v", err)
		}
		expected := today.Add(9*time.Hour + 5*time.Minute)
		if !result.Equal(expected) {
			t.Errorf("parseTimeArg(\"9:5\") = %v, want %v", result, expected)
		}
	})
}

func TestFormatTagDisplay(t *testing.T) {
	t.Run("タグ名が存在する場合", func(t *testing.T) {
		result := formatTagDisplay("5", "Development")
		if result != "Development (5)" {
			t.Errorf("formatTagDisplay(\"5\", \"Development\") = %s, want Development (5)", result)
		}
	})

	t.Run("タグ名が空の場合", func(t *testing.T) {
		result := formatTagDisplay("999", "")
		if result != "999" {
			t.Errorf("formatTagDisplay(\"999\", \"\") = %s, want 999", result)
		}
	})

	t.Run("タグ名に日本語が含まれる場合", func(t *testing.T) {
		result := formatTagDisplay("7", "月次")
		if result != "月次 (7)" {
			t.Errorf("formatTagDisplay(\"7\", \"月次\") = %s, want 月次 (7)", result)
		}
	})
}

func TestParseDateArg(t *testing.T) {
	t.Run("YYYY-MM-DD形式", func(t *testing.T) {
		result, err := parseDateArg("2025-10-07")
		if err != nil {
			t.Errorf("parseDateArg(\"2025-10-07\") returned error: %v", err)
		}
		expected := time.Date(2025, 10, 7, 0, 0, 0, 0, time.Local)
		if !result.Equal(expected) {
			t.Errorf("parseDateArg(\"2025-10-07\") = %v, want %v", result, expected)
		}
	})

	t.Run("YYYYMMDD形式", func(t *testing.T) {
		result, err := parseDateArg("20251007")
		if err != nil {
			t.Errorf("parseDateArg(\"20251007\") returned error: %v", err)
		}
		expected := time.Date(2025, 10, 7, 0, 0, 0, 0, time.Local)
		if !result.Equal(expected) {
			t.Errorf("parseDateArg(\"20251007\") = %v, want %v", result, expected)
		}
	})

	t.Run("相対日付 -1d 形式", func(t *testing.T) {
		result, err := parseDateArg("-1d")
		if err != nil {
			t.Errorf("parseDateArg(\"-1d\") returned error: %v", err)
		}
		yesterday := time.Now().AddDate(0, 0, -1)
		expected := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, time.Local)
		if !result.Equal(expected) {
			t.Errorf("parseDateArg(\"-1d\") = %v, want %v", result, expected)
		}
	})

	t.Run("不正な日付形式", func(t *testing.T) {
		_, err := parseDateArg("2025/10/07")
		if err == nil {
			t.Error("parseDateArg(\"2025/10/07\") should return error, but got nil")
		}
	})

	t.Run("不正な日付 - 13月", func(t *testing.T) {
		_, err := parseDateArg("2025-13-01")
		if err == nil {
			t.Error("parseDateArg(\"2025-13-01\") should return error, but got nil")
		}
	})

	t.Run("不正な日付 - 32日", func(t *testing.T) {
		_, err := parseDateArg("2025-10-32")
		if err == nil {
			t.Error("parseDateArg(\"2025-10-32\") should return error, but got nil")
		}
	})

	t.Run("空文字列", func(t *testing.T) {
		_, err := parseDateArg("")
		if err == nil {
			t.Error("parseDateArg(\"\") should return error, but got nil")
		}
	})
}

func TestFormatDateLabel(t *testing.T) {
	// 相対日数をもとに lastActivity を作成するヘルパー
	daysAgo := func(n int) time.Time {
		d := time.Now().AddDate(0, 0, -n)
		// 時刻は当日正午にして日付計算が安定するようにする
		return time.Date(d.Year(), d.Month(), d.Day(), 12, 0, 0, 0, d.Location())
	}

	tests := []struct {
		name  string
		input time.Time
		want  string
		note  string
	}{
		{
			name:  "0日前（今日） → [Today]",
			input: daysAgo(0),
			want:  "[Today]",
		},
		{
			name:  "1日前（昨日） → [Yesterday]",
			input: daysAgo(1),
			want:  "[Yesterday]",
		},
		{
			name:  "3日前 → [3 days ago]",
			input: daysAgo(3),
			want:  "[3 days ago]",
		},
		{
			name:  "6日前（< 7） → [6 days ago]",
			input: daysAgo(6),
			want:  "[6 days ago]",
		},
		{
			name:  "7日前（境界: ちょうど1週間） → [1 week ago]",
			input: daysAgo(7),
			want:  "[1 week ago]",
		},
		{
			name:  "8日前（7日超・14日以内） → [1 week ago]",
			input: daysAgo(8),
			want:  "[1 week ago]",
		},
		{
			name:  "13日前（7 / 7 = 1） → [1 week ago]",
			input: daysAgo(13),
			want:  "[1 week ago]",
		},
		{
			name:  "14日前（14 / 7 = 2） → [2 weeks ago]",
			input: daysAgo(14),
			want:  "[2 weeks ago]",
		},
		{
			name:  "15日前（>= 15） → 空文字列（ラベルなし）",
			input: daysAgo(15),
			want:  "",
		},
		{
			name:  "30日前 → 空文字列（ラベルなし）",
			input: daysAgo(30),
			want:  "",
		},
		{
			// NOTE: 未来日付（daysDiff < 0）に対する処理は未定義。
			// daysDiff < 0 の場合、daysDiff < 7 が true となるため
			// "[-N days ago]" のような文字列が返る（バグ疑い）。
			// 修正候補: daysDiff < 0 の場合 "" を返すべきか要検討。
			name:  "未来日付（1日後） → \"[-1 days ago]\"（現状挙動・バグ疑い）",
			input: time.Now().AddDate(0, 0, 1),
			want:  fmt.Sprintf("[%d days ago]", -1),
			note:  "未来日付の扱いは未定義。daysDiff=-1 が daysDiff<7 の分岐に入るため、意図しない文字列が返る",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := formatDateLabel(tc.input)
			if got != tc.want {
				extra := ""
				if tc.note != "" {
					extra = " [NOTE: " + tc.note + "]"
				}
				t.Errorf("formatDateLabel() = %q, want %q%s", got, tc.want, extra)
			}
		})
	}
}

func TestParseRelativeDate(t *testing.T) {
	t.Run("相対日付 -1d は昨日を返す", func(t *testing.T) {
		result, err := parseRelativeDate("-1d")
		if err != nil {
			t.Errorf("parseRelativeDate(\"-1d\") returned error: %v", err)
		}
		if result == nil {
			t.Fatal("parseRelativeDate(\"-1d\") returned nil, want date")
		}
		yesterday := time.Now().AddDate(0, 0, -1)
		expected := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, time.Local)
		if !result.Equal(expected) {
			t.Errorf("parseRelativeDate(\"-1d\") = %v, want %v", result, expected)
		}
	})

	t.Run("相対日付 -2d は一昨日を返す", func(t *testing.T) {
		result, err := parseRelativeDate("-2d")
		if err != nil {
			t.Errorf("parseRelativeDate(\"-2d\") returned error: %v", err)
		}
		if result == nil {
			t.Fatal("parseRelativeDate(\"-2d\") returned nil, want date")
		}
		twoDaysAgo := time.Now().AddDate(0, 0, -2)
		expected := time.Date(twoDaysAgo.Year(), twoDaysAgo.Month(), twoDaysAgo.Day(), 0, 0, 0, 0, time.Local)
		if !result.Equal(expected) {
			t.Errorf("parseRelativeDate(\"-2d\") = %v, want %v", result, expected)
		}
	})

	t.Run("相対日付 -0d は今日を返す", func(t *testing.T) {
		result, err := parseRelativeDate("-0d")
		if err != nil {
			t.Errorf("parseRelativeDate(\"-0d\") returned error: %v", err)
		}
		if result == nil {
			t.Fatal("parseRelativeDate(\"-0d\") returned nil, want date")
		}
		today := time.Now()
		expected := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.Local)
		if !result.Equal(expected) {
			t.Errorf("parseRelativeDate(\"-0d\") = %v, want %v", result, expected)
		}
	})

	t.Run("相対日付 -30d は30日前を返す", func(t *testing.T) {
		result, err := parseRelativeDate("-30d")
		if err != nil {
			t.Errorf("parseRelativeDate(\"-30d\") returned error: %v", err)
		}
		if result == nil {
			t.Fatal("parseRelativeDate(\"-30d\") returned nil, want date")
		}
		thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
		expected := time.Date(thirtyDaysAgo.Year(), thirtyDaysAgo.Month(), thirtyDaysAgo.Day(), 0, 0, 0, 0, time.Local)
		if !result.Equal(expected) {
			t.Errorf("parseRelativeDate(\"-30d\") = %v, want %v", result, expected)
		}
	})

	t.Run("非マッチ形式 YYYY-MM-DD は nil, nil を返す", func(t *testing.T) {
		result, err := parseRelativeDate("2025-01-01")
		if err != nil {
			t.Errorf("parseRelativeDate(\"2025-01-01\") returned error: %v", err)
		}
		if result != nil {
			t.Errorf("parseRelativeDate(\"2025-01-01\") = %v, want nil", result)
		}
	})

	t.Run("非マッチ形式 YYYYMMDD は nil, nil を返す", func(t *testing.T) {
		result, err := parseRelativeDate("20250101")
		if err != nil {
			t.Errorf("parseRelativeDate(\"20250101\") returned error: %v", err)
		}
		if result != nil {
			t.Errorf("parseRelativeDate(\"20250101\") = %v, want nil", result)
		}
	})

	t.Run("非マッチ形式 空文字列 は nil, nil を返す", func(t *testing.T) {
		result, err := parseRelativeDate("")
		if err != nil {
			t.Errorf("parseRelativeDate(\"\") returned error: %v", err)
		}
		if result != nil {
			t.Errorf("parseRelativeDate(\"\") = %v, want nil", result)
		}
	})
}
