package main

import (
	"os"
	"testing"
)

// TestRun tests the run function
func TestRun(t *testing.T) {
	t.Run("正常終了時は終了コード0を返す", func(t *testing.T) {
		// テスト実行時のos.Argsを保存して後で復元
		originalArgs := os.Args
		defer func() { os.Args = originalArgs }()

		// ヘルプを表示する引数を設定（正常終了する）
		os.Args = []string{"worklog", "help"}

		code := run()

		if code != 0 {
			t.Errorf("期待値: 0, 実際: %d", code)
		}
	})

	t.Run("不明なコマンドで終了コード1を返す", func(t *testing.T) {
		// テスト実行時のos.Argsを保存して後で復元
		originalArgs := os.Args
		defer func() { os.Args = originalArgs }()

		// 不明なコマンドを設定
		os.Args = []string{"worklog", "unknown-command"}

		code := run()

		if code != 1 {
			t.Errorf("期待値: 1, 実際: %d", code)
		}
	})
}
