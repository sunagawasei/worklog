package storage

import (
	"encoding/json"
	"os"

	"worklog/internal/domain"
)

// Tag は domain.Tag のエイリアス（後方互換性のため）
type Tag = domain.Tag

// TagStorage はタグ情報の読み込みを行うインターフェース
type TagStorage interface {
	Load() ([]Tag, error)
	Save(tags []Tag) error
	Add(name string) (Tag, error)
	Delete(id int) error
}

// tagStorage はTagStorageインターフェースの実装
type tagStorage struct {
	filePath string
}

// NewTagStorage は新しいTagStorageインスタンスを作成する
func NewTagStorage(filePath string) TagStorage {
	return &tagStorage{
		filePath: filePath,
	}
}

// tagsData はJSONファイルの構造を表す
type tagsData struct {
	Tags   []Tag `json:"tags"`
	NextID int   `json:"nextID,omitempty"`
}

// Load はタグ情報をJSONファイルから読み込む
func (s *tagStorage) Load() ([]Tag, error) {
	// ファイルを読み込む
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return nil, err
	}

	// JSONをパース
	var td tagsData
	err = json.Unmarshal(data, &td)
	if err != nil {
		return nil, err
	}

	return td.Tags, nil
}

// saveData は内部用の保存メソッド（NextIDを保持）
func (s *tagStorage) saveData(td tagsData) error {
	// JSONにエンコード（インデント付き）
	data, err := json.MarshalIndent(td, "", "  ")
	if err != nil {
		return err
	}
	// ファイルに書き込み
	return os.WriteFile(s.filePath, data, 0644)
}

// Save はタグ情報をJSONファイルに保存する（外部用）
func (s *tagStorage) Save(tags []Tag) error {
	// 既存のNextIDを読み込む
	var td tagsData
	data, err := os.ReadFile(s.filePath)
	if err == nil {
		json.Unmarshal(data, &td)
	}

	td.Tags = tags
	// NextIDがない場合は最大ID+1で設定
	if td.NextID == 0 {
		maxID := 0
		for _, tag := range tags {
			if tag.ID > maxID {
				maxID = tag.ID
			}
		}
		td.NextID = maxID + 1
	}

	return s.saveData(td)
}

// Add は新しいタグを追加する（ID自動採番）
func (s *tagStorage) Add(name string) (Tag, error) {
	// 現在のデータを読み込む（tagsDataをそのまま使用）
	data, err := os.ReadFile(s.filePath)
	var td tagsData
	if err == nil {
		json.Unmarshal(data, &td)
	}

	// 同名タグのチェック
	for _, tag := range td.Tags {
		if tag.Name == name {
			return Tag{}, os.ErrExist
		}
	}

	// NextIDがない場合は最大ID+1で初期化（既存データのマイグレーション）
	if td.NextID == 0 {
		maxID := 0
		for _, tag := range td.Tags {
			if tag.ID > maxID {
				maxID = tag.ID
			}
		}
		td.NextID = maxID + 1
	}

	// 新しいタグを作成
	newTag := Tag{
		ID:   td.NextID,
		Name: name,
	}

	// NextIDをインクリメント
	td.Tags = append(td.Tags, newTag)
	td.NextID++

	// 保存
	err = s.saveData(td)
	if err != nil {
		return Tag{}, err
	}

	return newTag, nil
}

// Delete は指定したIDのタグを削除する
func (s *tagStorage) Delete(id int) error {
	// 現在のタグリストを読み込む
	tags, err := s.Load()
	if err != nil {
		return err
	}

	// 指定したIDのタグを探して削除
	found := false
	newTags := make([]Tag, 0, len(tags))
	for _, tag := range tags {
		if tag.ID == id {
			found = true
			// このタグは追加しない（削除）
		} else {
			newTags = append(newTags, tag)
		}
	}

	// タグが見つからなかった場合はエラー
	if !found {
		return os.ErrNotExist
	}

	// 保存
	return s.Save(newTags)
}
