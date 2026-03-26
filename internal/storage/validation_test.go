package storage

import (
	"strings"
	"testing"
)

func TestValidateProjectName(t *testing.T) {
	t.Run("有効なプロジェクト名", func(t *testing.T) {
		validNames := []string{
			"ProjectA",
			"開発作業",
			"TASK-001",
			"プロジェクト 123",
			"Project:Name", // コロンも許可
		}

		for _, name := range validNames {
			err := ValidateProjectName(name)
			if err != nil {
				t.Errorf("ValidateProjectName(%q) returned error: %v, want nil", name, err)
			}
		}
	})

	t.Run("改行を含むプロジェクト名", func(t *testing.T) {
		invalidNames := []string{
			"Project\nName",
			"Project\rName",
			"Project\r\nName",
		}

		for _, name := range invalidNames {
			err := ValidateProjectName(name)
			if err == nil {
				t.Errorf("ValidateProjectName(%q) returned nil, want error", name)
			}
		}
	})

	t.Run("タブを含むプロジェクト名", func(t *testing.T) {
		name := "Project\tName"
		err := ValidateProjectName(name)
		if err == nil {
			t.Errorf("ValidateProjectName(%q) returned nil, want error", name)
		}
	})

	t.Run("ファイル名不適切文字を含むプロジェクト名", func(t *testing.T) {
		invalidNames := []string{
			"Project/Name",
			"Project\\Name",
			"Project*Name",
			"Project?Name",
			"Project\"Name",
			"Project<Name",
			"Project>Name",
			"Project|Name",
		}

		for _, name := range invalidNames {
			err := ValidateProjectName(name)
			if err == nil {
				t.Errorf("ValidateProjectName(%q) returned nil, want error", name)
			}
		}
	})

	t.Run("空のプロジェクト名", func(t *testing.T) {
		err := ValidateProjectName("")
		if err == nil {
			t.Error("ValidateProjectName(\"\") returned nil, want error")
		}
	})

	t.Run("パストラバーサル(..)を含むプロジェクト名", func(t *testing.T) {
		invalidNames := []string{
			"..",
			"../etc",
			"foo..bar",
		}
		for _, name := range invalidNames {
			err := ValidateProjectName(name)
			if err == nil {
				t.Errorf("ValidateProjectName(%q) returned nil, want error", name)
			}
		}
	})

	t.Run("制御文字を含むプロジェクト名", func(t *testing.T) {
		invalidNames := []string{
			"Project\x00Name",
			"Project\x01Name",
			"Project\x1fName",
			"Project\x7fName", // DEL文字
		}
		for _, name := range invalidNames {
			err := ValidateProjectName(name)
			if err == nil {
				t.Errorf("ValidateProjectName(%q) returned nil, want error", name)
			}
		}
	})

	t.Run("255文字以下は有効（ASCII）", func(t *testing.T) {
		name := strings.Repeat("a", 255)
		err := ValidateProjectName(name)
		if err != nil {
			t.Errorf("ValidateProjectName(255 ASCII) returned error: %v, want nil", err)
		}
	})

	t.Run("256文字以上は無効（ASCII）", func(t *testing.T) {
		name := strings.Repeat("a", 256)
		err := ValidateProjectName(name)
		if err == nil {
			t.Error("ValidateProjectName(256 ASCII) returned nil, want error")
		}
	})

	t.Run("255文字以下は有効（マルチバイト）", func(t *testing.T) {
		name := strings.Repeat("あ", 255)
		err := ValidateProjectName(name)
		if err != nil {
			t.Errorf("ValidateProjectName(255 multibyte) returned error: %v, want nil", err)
		}
	})

	t.Run("256文字以上は無効（マルチバイト）", func(t *testing.T) {
		name := strings.Repeat("あ", 256)
		err := ValidateProjectName(name)
		if err == nil {
			t.Error("ValidateProjectName(256 multibyte) returned nil, want error")
		}
	})
}

func TestValidateTagName(t *testing.T) {
	t.Run("有効なタグ名", func(t *testing.T) {
		validNames := []string{
			"開発",
			"MTG",
			"Review-1",
		}
		for _, name := range validNames {
			err := ValidateTagName(name)
			if err != nil {
				t.Errorf("ValidateTagName(%q) returned error: %v, want nil", name, err)
			}
		}
	})

	t.Run("空のタグ名は無効", func(t *testing.T) {
		err := ValidateTagName("")
		if err == nil {
			t.Error("ValidateTagName(\"\") returned nil, want error")
		}
	})

	t.Run("制御文字を含むタグ名は無効", func(t *testing.T) {
		invalidNames := []string{
			"tag\x00name",
			"tag\x7fname", // DEL文字
		}
		for _, name := range invalidNames {
			err := ValidateTagName(name)
			if err == nil {
				t.Errorf("ValidateTagName(%q) returned nil, want error", name)
			}
		}
	})

	t.Run("255文字以下は有効（マルチバイト）", func(t *testing.T) {
		name := strings.Repeat("あ", 255)
		err := ValidateTagName(name)
		if err != nil {
			t.Errorf("ValidateTagName(255 multibyte) returned error: %v, want nil", err)
		}
	})

	t.Run("256文字以上は無効（マルチバイト）", func(t *testing.T) {
		name := strings.Repeat("あ", 256)
		err := ValidateTagName(name)
		if err == nil {
			t.Error("ValidateTagName(256 multibyte) returned nil, want error")
		}
	})
}
