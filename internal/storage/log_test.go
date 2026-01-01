package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLogStorage_Append(t *testing.T) {
	t.Run("新規ログエントリを追加", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage := NewLogStorage(tmpDir)

		now := time.Date(2025, 9, 24, 10, 30, 0, 0, time.Local)
		project := "TestProject"
		action := "start"
		tag := "DEV"

		err := storage.Append(project, action, tag, now)
		if err != nil {
			t.Fatalf("Append failed: %v", err)
		}

		// ログファイルが作成されたことを確認
		expectedPath := filepath.Join(tmpDir, "2025", "09", "24", "TestProject.log")
		content, err := os.ReadFile(expectedPath)
		if err != nil {
			t.Fatalf("Failed to read log file: %v", err)
		}

		expectedContent := "2025/09/24 10:30:00\tstart\tDEV\n"
		if string(content) != expectedContent {
			t.Errorf("Log content mismatch\nExpected: %q\nGot: %q", expectedContent, string(content))
		}
	})

	t.Run("既存ファイルへの追記", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage := NewLogStorage(tmpDir)

		now := time.Date(2025, 9, 24, 10, 30, 0, 0, time.Local)
		project := "TestProject"

		// 1つ目のエントリを追加
		err := storage.Append(project, "start", "DEV", now)
		if err != nil {
			t.Fatalf("First append failed: %v", err)
		}

		// 2つ目のエントリを追加
		stopTime := now.Add(1 * time.Hour)
		err = storage.Append(project, "stop", "DEV", stopTime)
		if err != nil {
			t.Fatalf("Second append failed: %v", err)
		}

		// ファイルの内容を確認
		expectedPath := filepath.Join(tmpDir, "2025", "09", "24", "TestProject.log")
		content, err := os.ReadFile(expectedPath)
		if err != nil {
			t.Fatalf("Failed to read log file: %v", err)
		}

		expectedContent := "2025/09/24 10:30:00\tstart\tDEV\n2025/09/24 11:30:00\tstop\tDEV\n"
		if string(content) != expectedContent {
			t.Errorf("Log content mismatch\nExpected: %q\nGot: %q", expectedContent, string(content))
		}
	})
}

func TestLogStorage_ReadToday(t *testing.T) {
	t.Run("本日のログエントリを読み込み", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage := NewLogStorage(tmpDir)

		// 本日の日付を使用
		now := time.Now()
		project := "TestProject"

		// テスト用のログエントリを追加
		testData := []struct {
			action string
			tag    string
			offset time.Duration
		}{
			{"start", "DEV", 0},
			{"stop", "DEV", 30 * time.Minute},
			{"start", "MTG", 1 * time.Hour},
		}

		// データを追加
		baseTime := time.Date(now.Year(), now.Month(), now.Day(), 10, 0, 0, 0, now.Location())
		for _, data := range testData {
			timestamp := baseTime.Add(data.offset)
			err := storage.Append(project, data.action, data.tag, timestamp)
			if err != nil {
				t.Fatalf("Failed to append test data: %v", err)
			}
		}

		// ReadTodayでデータを読み込み
		entries, err := storage.ReadToday(project)
		if err != nil {
			t.Fatalf("ReadToday failed: %v", err)
		}

		// エントリ数の確認
		if len(entries) != len(testData) {
			t.Fatalf("Expected %d entries, got %d", len(testData), len(entries))
		}

		// 各エントリの内容を検証
		for i, entry := range entries {
			if entry.Action != testData[i].action {
				t.Errorf("Entry %d: action mismatch. Expected %q, got %q",
					i, testData[i].action, entry.Action)
			}
			if entry.Tag != testData[i].tag {
				t.Errorf("Entry %d: tag mismatch. Expected %q, got %q",
					i, testData[i].tag, entry.Tag)
			}
		}
	})

	t.Run("ファイルが存在しない場合", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage := NewLogStorage(tmpDir)

		// 存在しないプロジェクトのログを読み込み
		entries, err := storage.ReadToday("NonExistentProject")
		if err != nil {
			t.Fatalf("ReadToday should not fail for non-existent file: %v", err)
		}

		// 空のスライスが返されることを確認
		if len(entries) != 0 {
			t.Errorf("Expected empty entries for non-existent file, got %d entries", len(entries))
		}
	})
}

func TestLogStorage_ReadAllOnDate(t *testing.T) {
	t.Run("指定日のログエントリを読み込み", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage := NewLogStorage(tmpDir)

		// 特定の日付を指定
		targetDate := time.Date(2025, 10, 5, 0, 0, 0, 0, time.Local)

		// 複数プロジェクトのテストデータを準備
		projects := []struct {
			name   string
			action string
			tag    string
		}{
			{"ProjectA", "start", "Development"},
			{"ProjectA", "stop", "Development"},
			{"ProjectB", "start", "MTG"},
		}

		// データを追加
		baseTime := time.Date(2025, 10, 5, 10, 0, 0, 0, time.Local)
		for i, p := range projects {
			timestamp := baseTime.Add(time.Duration(i) * 30 * time.Minute)
			err := storage.Append(p.name, p.action, p.tag, timestamp)
			if err != nil {
				t.Fatalf("Failed to append test data: %v", err)
			}
		}

		// ReadAllOnDateでデータを読み込み
		allEntries, err := storage.ReadAllOnDate(targetDate)
		if err != nil {
			t.Fatalf("ReadAllOnDate failed: %v", err)
		}

		// プロジェクト数の確認
		if len(allEntries) != 2 {
			t.Fatalf("Expected 2 projects, got %d", len(allEntries))
		}

		// ProjectAのエントリ数を確認
		if len(allEntries["ProjectA"]) != 2 {
			t.Errorf("ProjectA: expected 2 entries, got %d", len(allEntries["ProjectA"]))
		}

		// ProjectBのエントリ数を確認
		if len(allEntries["ProjectB"]) != 1 {
			t.Errorf("ProjectB: expected 1 entry, got %d", len(allEntries["ProjectB"]))
		}
	})
}

func TestLogStorage_ReadRange(t *testing.T) {
	t.Run("複数日のログエントリを読み込み", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage := NewLogStorage(tmpDir)

		// 3日間のテストデータを準備
		day1 := time.Date(2025, 10, 1, 10, 0, 0, 0, time.Local)
		day2 := time.Date(2025, 10, 2, 10, 0, 0, 0, time.Local)
		day3 := time.Date(2025, 10, 3, 10, 0, 0, 0, time.Local)

		// Day1: ProjectAとProjectB
		_ = storage.Append("ProjectA", "start", "Development", day1)
		_ = storage.Append("ProjectA", "stop", "Development", day1.Add(1*time.Hour))
		_ = storage.Append("ProjectB", "start", "MTG", day1.Add(2*time.Hour))

		// Day2: ProjectAのみ
		_ = storage.Append("ProjectA", "start", "Development", day2)
		_ = storage.Append("ProjectA", "stop", "Development", day2.Add(1*time.Hour))

		// Day3: ProjectCのみ
		_ = storage.Append("ProjectC", "start", "REV", day3)

		// Day1からDay3までを読み込み
		allEntries, err := storage.ReadRange(day1, day3)
		if err != nil {
			t.Fatalf("ReadRange failed: %v", err)
		}

		// プロジェクト数の確認
		if len(allEntries) != 3 {
			t.Fatalf("Expected 3 projects, got %d", len(allEntries))
		}

		// ProjectAのエントリ数を確認（day1の2エントリ + day2の2エントリ = 4エントリ）
		if len(allEntries["ProjectA"]) != 4 {
			t.Errorf("ProjectA: expected 4 entries, got %d", len(allEntries["ProjectA"]))
		}

		// ProjectBのエントリ数を確認（day1の1エントリ）
		if len(allEntries["ProjectB"]) != 1 {
			t.Errorf("ProjectB: expected 1 entry, got %d", len(allEntries["ProjectB"]))
		}

		// ProjectCのエントリ数を確認（day3の1エントリ）
		if len(allEntries["ProjectC"]) != 1 {
			t.Errorf("ProjectC: expected 1 entry, got %d", len(allEntries["ProjectC"]))
		}
	})

	t.Run("範囲外のデータは含まれない", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage := NewLogStorage(tmpDir)

		day1 := time.Date(2025, 10, 1, 10, 0, 0, 0, time.Local)
		day2 := time.Date(2025, 10, 2, 10, 0, 0, 0, time.Local)
		day5 := time.Date(2025, 10, 5, 10, 0, 0, 0, time.Local)

		// Day1, Day2, Day5にデータを追加
		_ = storage.Append("ProjectA", "start", "Development", day1)
		_ = storage.Append("ProjectA", "start", "Development", day2)
		_ = storage.Append("ProjectA", "start", "Development", day5)

		// Day1からDay2までを読み込み（Day5は含まれない）
		allEntries, err := storage.ReadRange(day1, day2)
		if err != nil {
			t.Fatalf("ReadRange failed: %v", err)
		}

		// ProjectAのエントリ数を確認（day1とday2のみ = 2エントリ）
		if len(allEntries["ProjectA"]) != 2 {
			t.Errorf("ProjectA: expected 2 entries (day1, day2), got %d", len(allEntries["ProjectA"]))
		}
	})

	t.Run("データが存在しない期間", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage := NewLogStorage(tmpDir)

		startDate := time.Date(2025, 10, 1, 0, 0, 0, 0, time.Local)
		endDate := time.Date(2025, 10, 3, 0, 0, 0, 0, time.Local)

		// データを追加せずに読み込み
		allEntries, err := storage.ReadRange(startDate, endDate)
		if err != nil {
			t.Fatalf("ReadRange should not fail for empty range: %v", err)
		}

		// 空のマップが返されることを確認
		if len(allEntries) != 0 {
			t.Errorf("Expected empty map for empty range, got %d projects", len(allEntries))
		}
	})
}
