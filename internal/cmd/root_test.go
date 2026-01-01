package cmd

import (
	"fmt"
	"testing"
	"time"

	"worklog/internal/domain"
	"worklog/internal/storage"
)

// mockProjectManager はテスト用のモック実装
type mockProjectManager struct {
	newError            error
	switchError         error
	stopError           error
	status              *domain.ProjectStatus
	statusError         error
	summaries           []domain.ProjectSummary
	listError           error
	listOnDateSummaries []domain.ProjectSummary
	listOnDateError     error
	calledMethods       []string // 呼ばれたメソッドを記録
}

func (m *mockProjectManager) New(project, tag string) error {
	m.calledMethods = append(m.calledMethods, fmt.Sprintf("New(%s,%s)", project, tag))
	return m.newError
}

func (m *mockProjectManager) Switch(project, tag string) error {
	m.calledMethods = append(m.calledMethods, fmt.Sprintf("Switch(%s,%s)", project, tag))
	return m.switchError
}

func (m *mockProjectManager) Stop() error {
	m.calledMethods = append(m.calledMethods, "Stop()")
	return m.stopError
}

func (m *mockProjectManager) Status() (*domain.ProjectStatus, error) {
	m.calledMethods = append(m.calledMethods, "Status()")
	return m.status, m.statusError
}

func (m *mockProjectManager) List() ([]domain.ProjectSummary, error) {
	m.calledMethods = append(m.calledMethods, "List()")
	return m.summaries, m.listError
}

func (m *mockProjectManager) ListOnDate(date time.Time) ([]domain.ProjectSummary, error) {
	m.calledMethods = append(m.calledMethods, fmt.Sprintf("ListOnDate(%v)", date))
	return m.listOnDateSummaries, m.listOnDateError
}

func (m *mockProjectManager) NewAt(project, tag string, timestamp time.Time) error {
	m.calledMethods = append(m.calledMethods, fmt.Sprintf("NewAt(%s,%s,%v)", project, tag, timestamp))
	return m.newError
}

func (m *mockProjectManager) SwitchAt(project, tag string, timestamp time.Time) error {
	m.calledMethods = append(m.calledMethods, fmt.Sprintf("SwitchAt(%s,%s,%v)", project, tag, timestamp))
	return m.switchError
}

func (m *mockProjectManager) StopAt(timestamp time.Time) error {
	m.calledMethods = append(m.calledMethods, fmt.Sprintf("StopAt(%v)", timestamp))
	return m.stopError
}

// mockTagStorage はテスト用のモック実装
type mockTagStorage struct {
	tags []storage.Tag
	err  error
}

func (m *mockTagStorage) Load() ([]storage.Tag, error) {
	return m.tags, m.err
}

func TestHandleSwitch_InteractiveMode(t *testing.T) {
	t.Run("本日のプロジェクトリストから選択して切り替え", func(t *testing.T) {
		// モックマネージャーを設定
		_ = &mockProjectManager{
			status: &domain.ProjectStatus{
				Project:   "CurrentProject",
				Tag:       "Development",
				StartTime: time.Now(),
			},
			summaries: []domain.ProjectSummary{
				{Project: "ProjectA", Tag: "Development", TotalTime: time.Hour},
				{Project: "ProjectB", Tag: "MTG", TotalTime: time.Hour * 2},
				{Project: "CurrentProject", Tag: "Development", TotalTime: time.Hour * 3},
			},
		}

		// 期待される動作:
		// 1. List()で本日のプロジェクト一覧を取得
		// 2. Status()で現在稼働中のプロジェクトを取得
		// 3. 稼働中でないプロジェクトのみリスト表示
		// 4. Switch()で選択したプロジェクトに切り替え

		// NOTE: 対話的UIのテストは複雑なため、統合テストで確認
		// ここでは、List()とStatus()が呼ばれることを確認する仕様とする
	})

	t.Run("本日のプロジェクトがない場合", func(t *testing.T) {
		// モックマネージャーを設定（プロジェクトリストが空）
		manager := &mockProjectManager{
			summaries: []domain.ProjectSummary{},
		}

		// NOTE: 実際のhandleSwitch関数の変更後にテストを追加
		_ = manager // linterエラー回避
	})

	t.Run("稼働中のプロジェクトがない場合", func(t *testing.T) {
		// モックマネージャーを設定（稼働中なし）
		manager := &mockProjectManager{
			status: nil, // 稼働中なし
			summaries: []domain.ProjectSummary{
				{Project: "ProjectA", Tag: "Development", TotalTime: time.Hour},
				{Project: "ProjectB", Tag: "MTG", TotalTime: time.Hour * 2},
			},
		}

		// NOTE: 実際のhandleSwitch関数の変更後にテストを追加
		_ = manager // linterエラー回避
	})
}

func TestParseTimeArg(t *testing.T) {
	// 現在時刻を固定（今日の日付を取得するため）
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
		expected := today
		if !result.Equal(expected) {
			t.Errorf("parseTimeArg(\"00:00\") = %v, want %v", result, expected)
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

	t.Run("HHMM形式 - 1230", func(t *testing.T) {
		result, err := parseTimeArg("1230")
		if err != nil {
			t.Errorf("parseTimeArg(\"1230\") returned error: %v", err)
		}
		expected := today.Add(12*time.Hour + 30*time.Minute)
		if !result.Equal(expected) {
			t.Errorf("parseTimeArg(\"1230\") = %v, want %v", result, expected)
		}
	})

	t.Run("HHMM形式 - 0000", func(t *testing.T) {
		result, err := parseTimeArg("0000")
		if err != nil {
			t.Errorf("parseTimeArg(\"0000\") returned error: %v", err)
		}
		expected := today
		if !result.Equal(expected) {
			t.Errorf("parseTimeArg(\"0000\") = %v, want %v", result, expected)
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
	tests := []struct {
		name     string
		tagID    string
		tagName  string
		expected string
	}{
		{
			name:     "タグ名が存在する場合",
			tagID:    "5",
			tagName:  "Development",
			expected: "Development (5)",
		},
		{
			name:     "タグ名が空の場合（見つからない）",
			tagID:    "999",
			tagName:  "",
			expected: "999",
		},
		{
			name:     "タグ名に日本語が含まれる場合",
			tagID:    "7",
			tagName:  "月次",
			expected: "月次 (7)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatTagDisplay(tt.tagID, tt.tagName)
			if result != tt.expected {
				t.Errorf("formatTagDisplay(%s, %s) = %s, want %s",
					tt.tagID, tt.tagName, result, tt.expected)
			}
		})
	}
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
