package cmd

import (
	"testing"
)

func TestParseGlobalFlags(t *testing.T) {
	t.Run("--json フラグを抽出してJSONModeをtrueにする", func(t *testing.T) {
		opts := parseGlobalFlags([]string{"--json", "status"})
		if !opts.JSONMode {
			t.Error("JSONMode should be true")
		}
		if len(opts.Args) != 1 || opts.Args[0] != "status" {
			t.Errorf("Args should be [status], got %v", opts.Args)
		}
	})

	t.Run("--no-interactive フラグを抽出してNoInteractiveをtrueにする", func(t *testing.T) {
		opts := parseGlobalFlags([]string{"status", "--no-interactive"})
		if !opts.NoInteractive {
			t.Error("NoInteractive should be true")
		}
		if len(opts.Args) != 1 || opts.Args[0] != "status" {
			t.Errorf("Args should be [status], got %v", opts.Args)
		}
	})

	t.Run("--json は暗黙的に --no-interactive を有効にする", func(t *testing.T) {
		opts := parseGlobalFlags([]string{"--json", "new", "TASK", "1"})
		if !opts.NoInteractive {
			t.Error("--json should imply --no-interactive")
		}
		if len(opts.Args) != 3 {
			t.Errorf("Args should have 3 elements, got %v", opts.Args)
		}
	})

	t.Run("フラグなしの場合はArgsにそのまま入る", func(t *testing.T) {
		opts := parseGlobalFlags([]string{"new", "PROJECT", "1"})
		if opts.JSONMode || opts.NoInteractive {
			t.Error("no flags should be set")
		}
		if len(opts.Args) != 3 {
			t.Errorf("Args should have 3 elements, got %v", opts.Args)
		}
	})

	t.Run("フラグが位置引数の間にあっても抽出できる", func(t *testing.T) {
		opts := parseGlobalFlags([]string{"new", "--json", "PROJECT", "1"})
		if !opts.JSONMode {
			t.Error("JSONMode should be true")
		}
		if len(opts.Args) != 3 {
			t.Errorf("Args should have 3 elements [new PROJECT 1], got %v", opts.Args)
		}
	})

	t.Run("引数なしの場合はArgs=nilになる", func(t *testing.T) {
		opts := parseGlobalFlags([]string{})
		if opts.JSONMode || opts.NoInteractive {
			t.Error("no flags should be set")
		}
		if len(opts.Args) != 0 {
			t.Errorf("Args should be empty, got %v", opts.Args)
		}
	})
}
