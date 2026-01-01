package storage

import (
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
}
