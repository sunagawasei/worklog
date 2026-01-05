package config

import (
	"os"
	"path/filepath"
)

// GetDataDir returns the data directory path.
// Priority: WORKLOG_DATA_DIR > XDG_STATE_HOME/worklog > ~/.local/state/worklog
func GetDataDir() string {
	if dir := os.Getenv("WORKLOG_DATA_DIR"); dir != "" {
		return dir
	}
	if xdgState := os.Getenv("XDG_STATE_HOME"); xdgState != "" {
		return filepath.Join(xdgState, "worklog")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "state", "worklog")
}

// GetTagsFile returns the tags file path.
// Priority: WORKLOG_TAGS_FILE > XDG_CONFIG_HOME/worklog/tags.json > ~/.config/worklog/tags.json
func GetTagsFile() string {
	if file := os.Getenv("WORKLOG_TAGS_FILE"); file != "" {
		return file
	}
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		return filepath.Join(xdgConfig, "worklog", "tags.json")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "worklog", "tags.json")
}
