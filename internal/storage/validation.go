package storage

import (
	"fmt"
	"strings"
)

// ValidateProjectName はプロジェクト名が有効かどうかを検証する
func ValidateProjectName(name string) error {
	if name == "" {
		return fmt.Errorf("プロジェクト名が空です")
	}

	// 改行文字のチェック
	if strings.ContainsAny(name, "\n\r") {
		return fmt.Errorf("プロジェクト名に改行文字を含めることはできません")
	}

	// タブ文字のチェック
	if strings.Contains(name, "\t") {
		return fmt.Errorf("プロジェクト名にタブ文字を含めることはできません")
	}

	// ファイル名として不適切な文字のチェック
	invalidChars := []string{"/", "\\", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalidChars {
		if strings.Contains(name, char) {
			return fmt.Errorf("プロジェクト名に不適切な文字(%s)を含めることはできません", char)
		}
	}

	return nil
}
