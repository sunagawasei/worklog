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
	Tags []Tag `json:"tags"`
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
