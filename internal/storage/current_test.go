package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCurrentStorage_Read(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tempDir := t.TempDir()
	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")
	day := now.Format("02")
	testDir := filepath.Join(tempDir, year, month, day)
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("テストディレクトリの作成に失敗: %v", err)
	}

	// テスト用のcurrentファイルを作成
	currentFile := filepath.Join(testDir, "current")
	expectedContent := "TestProject:DEV"
	err = os.WriteFile(currentFile, []byte(expectedContent), 0644)
	if err != nil {
		t.Fatalf("テストファイルの作成に失敗: %v", err)
	}

	// CurrentStorageインスタンスを作成
	storage := NewCurrentStorage(tempDir)

	// Readメソッドをテスト
	content, err := storage.Read()
	if err != nil {
		t.Errorf("Read() returned error: %v", err)
	}

	if content != expectedContent {
		t.Errorf("Read() = %q, want %q", content, expectedContent)
	}
}

func TestCurrentStorage_Write(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tempDir := t.TempDir()

	// CurrentStorageインスタンスを作成
	storage := NewCurrentStorage(tempDir)

	// テストデータ
	project := "TestProject"
	tag := "DEV"

	// Writeメソッドをテスト
	err := storage.Write(project, tag)
	if err != nil {
		t.Errorf("Write() returned error: %v", err)
	}

	// ファイルが作成されたことを確認
	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")
	day := now.Format("02")
	expectedFile := filepath.Join(tempDir, year, month, day, "current")

	content, err := os.ReadFile(expectedFile)
	if err != nil {
		t.Errorf("Failed to read written file: %v", err)
	}

	expectedContent := "TestProject\tDEV"
	if string(content) != expectedContent {
		t.Errorf("Written content = %q, want %q", string(content), expectedContent)
	}
}

func TestCurrentStorage_Read_WithNewline(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tempDir := t.TempDir()
	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")
	day := now.Format("02")
	testDir := filepath.Join(tempDir, year, month, day)
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("テストディレクトリの作成に失敗: %v", err)
	}

	// テスト用のcurrentファイルを作成（末尾に改行を含む）
	currentFile := filepath.Join(testDir, "current")
	contentWithNewline := "TestProject:DEV\n"
	err = os.WriteFile(currentFile, []byte(contentWithNewline), 0644)
	if err != nil {
		t.Fatalf("テストファイルの作成に失敗: %v", err)
	}

	// CurrentStorageインスタンスを作成
	storage := NewCurrentStorage(tempDir)

	// Readメソッドをテスト
	content, err := storage.Read()
	if err != nil {
		t.Errorf("Read() returned error: %v", err)
	}

	// 改行がトリムされていることを期待
	expectedContent := "TestProject:DEV"
	if content != expectedContent {
		t.Errorf("Read() = %q, want %q", content, expectedContent)
	}
}

func TestCurrentStorage_Write_InvalidProjectName(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tempDir := t.TempDir()

	// CurrentStorageインスタンスを作成
	storage := NewCurrentStorage(tempDir)

	// 不正なプロジェクト名のテスト
	invalidNames := []string{
		"Project\nName", // 改行
		"Project/Name",  // スラッシュ
		"Project\tName", // タブ
	}

	for _, name := range invalidNames {
		err := storage.Write(name, "TAG")
		if err == nil {
			t.Errorf("Write(%q, \"TAG\") returned nil, want error", name)
		}
	}
}

func TestCurrentStorage_Clear(t *testing.T) {
	t.Run("ファイルが存在する場合", func(t *testing.T) {
		// テスト用の一時ディレクトリを作成
		tempDir := t.TempDir()
		now := time.Now()
		year := now.Format("2006")
		month := now.Format("01")
		day := now.Format("02")
		testDir := filepath.Join(tempDir, year, month, day)
		err := os.MkdirAll(testDir, 0755)
		if err != nil {
			t.Fatalf("テストディレクトリの作成に失敗: %v", err)
		}

		// テスト用のcurrentファイルを作成
		currentFile := filepath.Join(testDir, "current")
		err = os.WriteFile(currentFile, []byte("TestProject:DEV"), 0644)
		if err != nil {
			t.Fatalf("テストファイルの作成に失敗: %v", err)
		}

		// CurrentStorageインスタンスを作成
		storage := NewCurrentStorage(tempDir)

		// Clearメソッドをテスト
		err = storage.Clear()
		if err != nil {
			t.Errorf("Clear() returned error: %v", err)
		}

		// ファイルが削除されたことを確認
		_, err = os.Stat(currentFile)
		if err == nil {
			t.Error("currentファイルが削除されていません")
		} else if !os.IsNotExist(err) {
			t.Errorf("Unexpected error when checking file: %v", err)
		}
	})

	t.Run("ファイルが存在しない場合", func(t *testing.T) {
		// テスト用の一時ディレクトリを作成（currentファイルは作成しない）
		tempDir := t.TempDir()

		// CurrentStorageインスタンスを作成
		storage := NewCurrentStorage(tempDir)

		// ファイルが存在しない状態でClearメソッドをテスト
		err := storage.Clear()
		if err != nil {
			t.Errorf("Clear() returned error when file doesn't exist: %v", err)
		}
	})
}
