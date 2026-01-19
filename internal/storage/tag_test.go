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

func TestTagStorage_Save(t *testing.T) {
	t.Run("タグリストを正常に保存", func(t *testing.T) {
		// テンポラリディレクトリの作成
		tmpDir := t.TempDir()
		tagsPath := filepath.Join(tmpDir, "tags.json")

		// TagStorageのインスタンスを作成
		storage := NewTagStorage(tagsPath)

		// 保存するタグリスト
		tags := []Tag{
			{ID: 1, Name: "開発"},
			{ID: 2, Name: "会議"},
			{ID: 3, Name: "レビュー"},
		}

		// Saveメソッドをテスト
		err := storage.Save(tags)
		if err != nil {
			t.Fatalf("Save()でエラーが発生: %v", err)
		}

		// ファイルが作成されたことを確認
		if _, err := os.Stat(tagsPath); os.IsNotExist(err) {
			t.Fatal("ファイルが作成されていない")
		}

		// ファイルから読み込んで確認
		loadedTags, err := storage.Load()
		if err != nil {
			t.Fatalf("Load()でエラーが発生: %v", err)
		}

		// タグ数の確認
		if len(loadedTags) != len(tags) {
			t.Errorf("保存されたタグ数が異なる: got %d, want %d", len(loadedTags), len(tags))
		}

		// 各タグの内容を確認
		for i := range tags {
			if loadedTags[i].ID != tags[i].ID {
				t.Errorf("tags[%d].ID = %d, want %d", i, loadedTags[i].ID, tags[i].ID)
			}
			if loadedTags[i].Name != tags[i].Name {
				t.Errorf("tags[%d].Name = %q, want %q", i, loadedTags[i].Name, tags[i].Name)
			}
		}
	})

	t.Run("空のタグリストを保存", func(t *testing.T) {
		// テンポラリディレクトリの作成
		tmpDir := t.TempDir()
		tagsPath := filepath.Join(tmpDir, "tags.json")

		// TagStorageのインスタンスを作成
		storage := NewTagStorage(tagsPath)

		// 空のタグリスト
		tags := []Tag{}

		// Saveメソッドをテスト
		err := storage.Save(tags)
		if err != nil {
			t.Fatalf("Save()でエラーが発生: %v", err)
		}

		// ファイルから読み込んで確認
		loadedTags, err := storage.Load()
		if err != nil {
			t.Fatalf("Load()でエラーが発生: %v", err)
		}

		// 空のスライスが返ることを確認
		if len(loadedTags) != 0 {
			t.Errorf("空のタグリストを期待したが、%d個のタグが返された", len(loadedTags))
		}
	})
}

func TestTagStorage_Add(t *testing.T) {
	t.Run("新しいタグを追加（ID自動採番）", func(t *testing.T) {
		// テンポラリディレクトリの作成
		tmpDir := t.TempDir()
		tagsPath := filepath.Join(tmpDir, "tags.json")

		// 初期タグリスト
		initialTags := []Tag{
			{ID: 1, Name: "開発"},
			{ID: 2, Name: "会議"},
		}

		// TagStorageのインスタンスを作成
		storage := NewTagStorage(tagsPath)
		err := storage.Save(initialTags)
		if err != nil {
			t.Fatalf("初期データの保存に失敗: %v", err)
		}

		// 新しいタグを追加
		newTag, err := storage.Add("レビュー")
		if err != nil {
			t.Fatalf("Add()でエラーが発生: %v", err)
		}

		// IDが自動採番されているか確認（最大ID+1）
		if newTag.ID != 3 {
			t.Errorf("newTag.ID = %d, want 3", newTag.ID)
		}

		// 名前が正しいか確認
		if newTag.Name != "レビュー" {
			t.Errorf("newTag.Name = %q, want %q", newTag.Name, "レビュー")
		}

		// ファイルから読み込んで確認
		loadedTags, err := storage.Load()
		if err != nil {
			t.Fatalf("Load()でエラーが発生: %v", err)
		}

		// タグ数の確認
		if len(loadedTags) != 3 {
			t.Errorf("タグ数が異なる: got %d, want 3", len(loadedTags))
		}

		// 追加されたタグが含まれているか確認
		found := false
		for _, tag := range loadedTags {
			if tag.ID == 3 && tag.Name == "レビュー" {
				found = true
				break
			}
		}
		if !found {
			t.Error("追加されたタグが見つからない")
		}
	})

	t.Run("空リストへのタグ追加", func(t *testing.T) {
		// テンポラリディレクトリの作成
		tmpDir := t.TempDir()
		tagsPath := filepath.Join(tmpDir, "tags.json")

		// TagStorageのインスタンスを作成
		storage := NewTagStorage(tagsPath)

		// 空リストを保存
		err := storage.Save([]Tag{})
		if err != nil {
			t.Fatalf("初期データの保存に失敗: %v", err)
		}

		// 新しいタグを追加
		newTag, err := storage.Add("開発")
		if err != nil {
			t.Fatalf("Add()でエラーが発生: %v", err)
		}

		// ID=1が割り当てられるか確認
		if newTag.ID != 1 {
			t.Errorf("newTag.ID = %d, want 1", newTag.ID)
		}
	})

	t.Run("同名タグが既に存在する場合はエラー", func(t *testing.T) {
		// テンポラリディレクトリの作成
		tmpDir := t.TempDir()
		tagsPath := filepath.Join(tmpDir, "tags.json")

		// 初期タグリスト
		initialTags := []Tag{
			{ID: 1, Name: "開発"},
			{ID: 2, Name: "会議"},
		}

		// TagStorageのインスタンスを作成
		storage := NewTagStorage(tagsPath)
		err := storage.Save(initialTags)
		if err != nil {
			t.Fatalf("初期データの保存に失敗: %v", err)
		}

		// 既に存在するタグ名で追加を試みる
		_, err = storage.Add("開発")
		if err == nil {
			t.Error("同名タグの追加でエラーが発生しなかった")
		}
	})

	t.Run("削除されたIDは再利用しない（nextID方式）", func(t *testing.T) {
		// テンポラリディレクトリの作成
		tmpDir := t.TempDir()
		tagsPath := filepath.Join(tmpDir, "tags.json")

		// TagStorageのインスタンスを作成
		storage := NewTagStorage(tagsPath)

		// 空リストを初期化
		err := storage.Save([]Tag{})
		if err != nil {
			t.Fatalf("初期データの保存に失敗: %v", err)
		}

		// 1. ID 1, 2, 3 のタグを作成
		tag1, err := storage.Add("開発")
		if err != nil {
			t.Fatalf("タグ1の追加に失敗: %v", err)
		}
		if tag1.ID != 1 {
			t.Errorf("tag1.ID = %d, want 1", tag1.ID)
		}

		tag2, err := storage.Add("会議")
		if err != nil {
			t.Fatalf("タグ2の追加に失敗: %v", err)
		}
		if tag2.ID != 2 {
			t.Errorf("tag2.ID = %d, want 2", tag2.ID)
		}

		tag3, err := storage.Add("レビュー")
		if err != nil {
			t.Fatalf("タグ3の追加に失敗: %v", err)
		}
		if tag3.ID != 3 {
			t.Errorf("tag3.ID = %d, want 3", tag3.ID)
		}

		// 2. ID 3 を削除
		err = storage.Delete(3)
		if err != nil {
			t.Fatalf("タグ3の削除に失敗: %v", err)
		}

		// 3. 新しいタグを追加 → ID 4 であることを確認（ID 3は再利用しない）
		tag4, err := storage.Add("テスト")
		if err != nil {
			t.Fatalf("タグ4の追加に失敗: %v", err)
		}

		if tag4.ID != 4 {
			t.Errorf("tag4.ID = %d, want 4 (削除されたID=3は再利用しない)", tag4.ID)
		}
	})
}

func TestTagStorage_Delete(t *testing.T) {
	t.Run("指定したIDのタグを削除", func(t *testing.T) {
		// テンポラリディレクトリの作成
		tmpDir := t.TempDir()
		tagsPath := filepath.Join(tmpDir, "tags.json")

		// 初期タグリスト
		initialTags := []Tag{
			{ID: 1, Name: "開発"},
			{ID: 2, Name: "会議"},
			{ID: 3, Name: "レビュー"},
		}

		// TagStorageのインスタンスを作成
		storage := NewTagStorage(tagsPath)
		err := storage.Save(initialTags)
		if err != nil {
			t.Fatalf("初期データの保存に失敗: %v", err)
		}

		// ID=2のタグを削除
		err = storage.Delete(2)
		if err != nil {
			t.Fatalf("Delete()でエラーが発生: %v", err)
		}

		// ファイルから読み込んで確認
		loadedTags, err := storage.Load()
		if err != nil {
			t.Fatalf("Load()でエラーが発生: %v", err)
		}

		// タグ数の確認
		if len(loadedTags) != 2 {
			t.Errorf("タグ数が異なる: got %d, want 2", len(loadedTags))
		}

		// 削除されたタグが含まれていないことを確認
		for _, tag := range loadedTags {
			if tag.ID == 2 {
				t.Error("削除されたはずのタグが含まれている")
			}
		}

		// 残りのタグを確認
		expectedTags := []Tag{
			{ID: 1, Name: "開発"},
			{ID: 3, Name: "レビュー"},
		}

		for i, expected := range expectedTags {
			if loadedTags[i].ID != expected.ID {
				t.Errorf("tags[%d].ID = %d, want %d", i, loadedTags[i].ID, expected.ID)
			}
			if loadedTags[i].Name != expected.Name {
				t.Errorf("tags[%d].Name = %q, want %q", i, loadedTags[i].Name, expected.Name)
			}
		}
	})

	t.Run("存在しないIDを指定した場合はエラー", func(t *testing.T) {
		// テンポラリディレクトリの作成
		tmpDir := t.TempDir()
		tagsPath := filepath.Join(tmpDir, "tags.json")

		// 初期タグリスト
		initialTags := []Tag{
			{ID: 1, Name: "開発"},
			{ID: 2, Name: "会議"},
		}

		// TagStorageのインスタンスを作成
		storage := NewTagStorage(tagsPath)
		err := storage.Save(initialTags)
		if err != nil {
			t.Fatalf("初期データの保存に失敗: %v", err)
		}

		// 存在しないID=99のタグを削除しようとする
		err = storage.Delete(99)
		if err == nil {
			t.Error("存在しないIDの削除でエラーが発生しなかった")
		}
	})

	t.Run("唯一のタグを削除して空リストになる", func(t *testing.T) {
		// テンポラリディレクトリの作成
		tmpDir := t.TempDir()
		tagsPath := filepath.Join(tmpDir, "tags.json")

		// 初期タグリスト（1つだけ）
		initialTags := []Tag{
			{ID: 1, Name: "開発"},
		}

		// TagStorageのインスタンスを作成
		storage := NewTagStorage(tagsPath)
		err := storage.Save(initialTags)
		if err != nil {
			t.Fatalf("初期データの保存に失敗: %v", err)
		}

		// タグを削除
		err = storage.Delete(1)
		if err != nil {
			t.Fatalf("Delete()でエラーが発生: %v", err)
		}

		// ファイルから読み込んで確認
		loadedTags, err := storage.Load()
		if err != nil {
			t.Fatalf("Load()でエラーが発生: %v", err)
		}

		// 空のスライスが返ることを確認
		if len(loadedTags) != 0 {
			t.Errorf("空のタグリストを期待したが、%d個のタグが返された", len(loadedTags))
		}
	})
}
