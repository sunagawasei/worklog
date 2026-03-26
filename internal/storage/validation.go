package storage

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

const maxNameLen = 255

// ValidateProjectName はプロジェクト名が有効かどうかを検証する
func ValidateProjectName(name string) error {
	return validateName(name, "プロジェクト名", true)
}

// ValidateTagName はタグ名が有効かどうかを検証する
func ValidateTagName(name string) error {
	return validateName(name, "タグ名", false)
}

func validateName(name, label string, checkFileChars bool) error {
	if name == "" {
		return fmt.Errorf("%sが空です", label)
	}

	if utf8.RuneCountInString(name) > maxNameLen {
		return fmt.Errorf("%sは%d文字以内にしてください", label, maxNameLen)
	}

	// パストラバーサル対策
	if strings.Contains(name, "..") {
		return fmt.Errorf("%sに\"..\"を含めることはできません", label)
	}

	// 制御文字（U+0000-U+001F, U+007F）チェック
	for _, r := range name {
		if r < 0x20 || r == 0x7f {
			return fmt.Errorf("%sに制御文字を含めることはできません", label)
		}
	}

	// ファイル名として不適切な文字のチェック（プロジェクト名のみ）
	if checkFileChars {
		invalidChars := []string{"/", "\\", "*", "?", "\"", "<", ">", "|"}
		for _, char := range invalidChars {
			if strings.Contains(name, char) {
				return fmt.Errorf("%sに不適切な文字(%s)を含めることはできません", label, char)
			}
		}
	}

	return nil
}
