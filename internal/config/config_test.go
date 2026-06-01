package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetDataDir(t *testing.T) {
	t.Run("WORKLOG_DATA_DIR設定時はそのまま返す", func(t *testing.T) {
		t.Setenv("WORKLOG_DATA_DIR", "/custom/data")
		t.Setenv("XDG_STATE_HOME", "")
		result := GetDataDir()
		if result != "/custom/data" {
			t.Errorf("GetDataDir() = %q, want %q", result, "/custom/data")
		}
	})

	t.Run("XDG_STATE_HOME設定時はworklogサブディレクトリを返す", func(t *testing.T) {
		t.Setenv("WORKLOG_DATA_DIR", "")
		t.Setenv("XDG_STATE_HOME", "/xdg/state")
		want := "/xdg/state/worklog"
		result := GetDataDir()
		if result != want {
			t.Errorf("GetDataDir() = %q, want %q", result, want)
		}
	})

	t.Run("両方未設定の場合はデフォルトパスを返す", func(t *testing.T) {
		t.Setenv("WORKLOG_DATA_DIR", "")
		t.Setenv("XDG_STATE_HOME", "")
		home, _ := os.UserHomeDir()
		want := filepath.Join(home, ".local", "state", "worklog")
		result := GetDataDir()
		if result != want {
			t.Errorf("GetDataDir() = %q, want %q", result, want)
		}
	})

	t.Run("両方設定時はWORKLOG_DATA_DIRが優先", func(t *testing.T) {
		t.Setenv("WORKLOG_DATA_DIR", "/priority/data")
		t.Setenv("XDG_STATE_HOME", "/xdg/state")
		result := GetDataDir()
		if result != "/priority/data" {
			t.Errorf("GetDataDir() = %q, want %q", result, "/priority/data")
		}
	})
}

func TestGetTagsFile(t *testing.T) {
	t.Run("WORKLOG_TAGS_FILE設定時はそのまま返す", func(t *testing.T) {
		t.Setenv("WORKLOG_TAGS_FILE", "/custom/tags.json")
		t.Setenv("XDG_CONFIG_HOME", "")
		result := GetTagsFile()
		if result != "/custom/tags.json" {
			t.Errorf("GetTagsFile() = %q, want %q", result, "/custom/tags.json")
		}
	})

	t.Run("XDG_CONFIG_HOME設定時はworklog/tags.jsonを返す", func(t *testing.T) {
		t.Setenv("WORKLOG_TAGS_FILE", "")
		t.Setenv("XDG_CONFIG_HOME", "/xdg/config")
		want := "/xdg/config/worklog/tags.json"
		result := GetTagsFile()
		if result != want {
			t.Errorf("GetTagsFile() = %q, want %q", result, want)
		}
	})

	t.Run("両方未設定の場合はデフォルトパスを返す", func(t *testing.T) {
		t.Setenv("WORKLOG_TAGS_FILE", "")
		t.Setenv("XDG_CONFIG_HOME", "")
		home, _ := os.UserHomeDir()
		want := filepath.Join(home, ".config", "worklog", "tags.json")
		result := GetTagsFile()
		if result != want {
			t.Errorf("GetTagsFile() = %q, want %q", result, want)
		}
	})

	t.Run("両方設定時はWORKLOG_TAGS_FILEが優先", func(t *testing.T) {
		t.Setenv("WORKLOG_TAGS_FILE", "/priority/tags.json")
		t.Setenv("XDG_CONFIG_HOME", "/xdg/config")
		result := GetTagsFile()
		if result != "/priority/tags.json" {
			t.Errorf("GetTagsFile() = %q, want %q", result, "/priority/tags.json")
		}
	})
}
