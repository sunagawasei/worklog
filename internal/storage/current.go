// Package storage はワークログデータの永続化を担当する
package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const currentFileName = "current"

// CurrentStorage は現在のプロジェクトファイル操作を管理する
type CurrentStorage interface {
	Read() (string, error)
	Write(project, tag string) error
	Clear() error
}

// currentStorage はCurrentStorageの実装
type currentStorage struct {
	baseDir string
}

// NewCurrentStorage は新しいCurrentStorageインスタンスを作成する
func NewCurrentStorage(baseDir string) CurrentStorage {
	return &currentStorage{
		baseDir: baseDir,
	}
}

// Read は今日のcurrentファイルから現在のプロジェクトを読み込む
func (c *currentStorage) Read() (string, error) {
	currentFile := c.getCurrentFilePath()

	content, err := os.ReadFile(currentFile)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("currentファイルが存在しません: %s", currentFile)
		}
		return "", err
	}

	return strings.TrimSpace(string(content)), nil
}

// Write は現在のプロジェクトとタグをcurrentファイルに書き込む
func (c *currentStorage) Write(project, tag string) error {
	// プロジェクト名のバリデーション
	if err := ValidateProjectName(project); err != nil {
		return fmt.Errorf("プロジェクト名が不正です: %w", err)
	}

	dir := c.getCurrentDirPath()
	content := fmt.Sprintf("%s\t%s", project, tag)

	// ディレクトリが存在しない場合は作成
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("ディレクトリ作成に失敗: %s: %w", dir, err)
	}

	// ファイルに書き込み
	currentFile := c.getCurrentFilePath()
	err = os.WriteFile(currentFile, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("currentファイル書き込みに失敗: %s: %w", currentFile, err)
	}

	return nil
}

// getCurrentDirPath は今日の日付ディレクトリパスを生成する
func (c *currentStorage) getCurrentDirPath() string {
	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")
	day := now.Format("02")
	return filepath.Join(c.baseDir, year, month, day)
}

// Clear は現在のプロジェクトファイルを削除する
func (c *currentStorage) Clear() error {
	currentFile := c.getCurrentFilePath()
	err := os.Remove(currentFile)
	if err != nil {
		if os.IsNotExist(err) {
			// ファイルが存在しない場合は成功とみなす
			return nil
		}
		return fmt.Errorf("currentファイル削除に失敗: %s: %w", currentFile, err)
	}
	return nil
}

// getCurrentFilePath は今日のcurrentファイルのパスを生成する
func (c *currentStorage) getCurrentFilePath() string {
	return filepath.Join(c.getCurrentDirPath(), currentFileName)
}
