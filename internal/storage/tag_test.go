package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTagStorage_Load(t *testing.T) {
	t.Run("正常なJSONファイルから読み込み", func(t *testing.T) {
		// テスト用の一時ファイルを作成
		tempFile, err := os.CreateTemp("", "tags_*.json")
		if err != nil {
			t.Fatalf("テスト用ファイルの作成に失敗: %v", err)
		}
		defer os.Remove(tempFile.Name())

		// テスト用のJSON（数値IDを使用）
		tagsJSON := `{
			"tags": [
				{"id": 1, "name": "開発"},
				{"id": 2, "name": "会議"},
				{"id": 3, "name": "レビュー"}
			]
		}`

		tagsPath := tempFile.Name()

		// ファイルにJSONを書き込み
		err = os.WriteFile(tagsPath, []byte(tagsJSON), 0644)
		if err != nil {
			t.Fatalf("テスト用ファイルの作成に失敗: %v", err)
		}

		// TagStorageのインスタンスを作成
		storage := NewTagStorage(tagsPath)

		// Loadメソッドをテスト
		tags, err := storage.Load()
		if err != nil {
			t.Errorf("Load()でエラーが発生: %v", err)
		}

		// タグ数の確認
		if len(tags) != 3 {
			t.Errorf("期待したタグ数と異なる: got %d, want 3", len(tags))
		}

		// 各タグの内容を確認
		expectedTags := []struct {
			id   int
			name string
		}{
			{1, "開発"},
			{2, "会議"},
			{3, "レビュー"},
		}

		for i, expected := range expectedTags {
			if tags[i].ID != expected.id {
				t.Errorf("tags[%d].ID = %d, want %d", i, tags[i].ID, expected.id)
			}
			if tags[i].Name != expected.name {
				t.Errorf("tags[%d].Name = %q, want %q", i, tags[i].Name, expected.name)
			}
		}
	})

	t.Run("ファイルが存在しない場合", func(t *testing.T) {
		// テンポラリディレクトリの作成
		tmpDir := t.TempDir()

		// 存在しないファイルのパスを設定
		tagsPath := filepath.Join(tmpDir, "tags.json")

		// TagStorageのインスタンスを作成
		storage := NewTagStorage(tagsPath)

		// Loadメソッドをテスト（エラーが発生することを期待）
		tags, err := storage.Load()
		if err == nil {
			t.Error("ファイルが存在しない場合、エラーが発生するべき")
		}
		if tags != nil {
			t.Error("ファイルが存在しない場合、tagsはnilであるべき")
		}

		// エラーがos.ErrNotExistに関連することを確認
		if !os.IsNotExist(err) {
			t.Errorf("期待したエラーと異なる: got %v", err)
		}
	})

	t.Run("不正なJSONファイルの場合", func(t *testing.T) {
		// テスト用の一時ファイルを作成
		tempFile, err := os.CreateTemp("", "tags_*.json")
		if err != nil {
			t.Fatalf("テスト用ファイルの作成に失敗: %v", err)
		}
		defer os.Remove(tempFile.Name())

		// 不正なJSON（文字列が数値型のフィールドに入っている）
		invalidJSON := `{
			"tags": [
				{"id": "invalid_should_be_int", "name": "開発"}
			]
		}`

		tagsPath := tempFile.Name()

		// ファイルに不正なJSONを書き込み
		err = os.WriteFile(tagsPath, []byte(invalidJSON), 0644)
		if err != nil {
			t.Fatalf("テスト用ファイルの作成に失敗: %v", err)
		}

		// TagStorageのインスタンスを作成
		storage := NewTagStorage(tagsPath)

		// Loadメソッドをテスト（エラーが返ることを期待）
		_, err = storage.Load()
		if err == nil {
			t.Error("Load()が不正なJSONでエラーを返さなかった")
		}
	})

	t.Run("空のタグリストの場合", func(t *testing.T) {
		// テスト用の一時ファイルを作成
		tempFile, err := os.CreateTemp("", "tags_*.json")
		if err != nil {
			t.Fatalf("テスト用ファイルの作成に失敗: %v", err)
		}
		defer os.Remove(tempFile.Name())

		// 空のタグリスト
		emptyJSON := `{
			"tags": []
		}`

		tagsPath := tempFile.Name()

		// ファイルにJSONを書き込み
		err = os.WriteFile(tagsPath, []byte(emptyJSON), 0644)
		if err != nil {
			t.Fatalf("テスト用ファイルの作成に失敗: %v", err)
		}

		// TagStorageのインスタンスを作成
		storage := NewTagStorage(tagsPath)

		// Loadメソッドをテスト
		tags, err := storage.Load()
		if err != nil {
			t.Errorf("Load()でエラーが発生: %v", err)
		}

		// 空のスライスが返ることを確認
		if len(tags) != 0 {
			t.Errorf("空のタグリストを期待したが、%d個のタグが返された", len(tags))
		}
	})

	t.Run("複数のタグタイプがある場合", func(t *testing.T) {
		// テスト用の一時ファイルを作成
		tempFile, err := os.CreateTemp("", "tags_*.json")
		if err != nil {
			t.Fatalf("テスト用ファイルの作成に失敗: %v", err)
		}
		defer os.Remove(tempFile.Name())

		// 様々なIDの値を持つJSON
		mixedJSON := `{
			"tags": [
				{"id": 1, "name": "開発"},
				{"id": 100, "name": "大規模ID"},
				{"id": 0, "name": "ゼロID"},
				{"id": -1, "name": "負のID"}
			]
		}`

		tagsPath := tempFile.Name()

		// ファイルにJSONを書き込み
		err = os.WriteFile(tagsPath, []byte(mixedJSON), 0644)
		if err != nil {
			t.Fatalf("テスト用ファイルの作成に失敗: %v", err)
		}

		// TagStorageのインスタンスを作成
		storage := NewTagStorage(tagsPath)

		// Loadメソッドをテスト
		tags, err := storage.Load()
		if err != nil {
			t.Errorf("Load()でエラーが発生: %v", err)
		}

		// タグ数の確認
		if len(tags) != 4 {
			t.Errorf("期待したタグ数と異なる: got %d, want 4", len(tags))
		}

		// 各タグのIDが正しく読み込まれたか確認
		expectedIDs := []int{1, 100, 0, -1}
		for i, expectedID := range expectedIDs {
			if tags[i].ID != expectedID {
				t.Errorf("tags[%d].ID = %d, want %d", i, tags[i].ID, expectedID)
			}
		}
	})
}
