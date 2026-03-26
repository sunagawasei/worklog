package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// ExecOptions はCLI実行時のグローバルオプションを保持する
type ExecOptions struct {
	JSONMode      bool
	NoInteractive bool
	Args          []string  // グローバルフラグを除いた引数 [0]=サブコマンド, [1+]=位置引数
	Writer        io.Writer // 出力先（nilの場合はos.Stdout）
}

// writer は出力先を返す（nil の場合は os.Stdout をフォールバック）
func (o ExecOptions) writer() io.Writer {
	if o.Writer == nil {
		return os.Stdout
	}
	return o.Writer
}

// ParseGlobalFlags は args からグローバルフラグを抽出し ExecOptions を返す
// --json と --no-interactive を認識し、残りを Args に格納する
// --json は暗黙的に --no-interactive を有効にする
func ParseGlobalFlags(args []string) ExecOptions { return parseGlobalFlags(args) }

func parseGlobalFlags(args []string) ExecOptions {
	var opts ExecOptions
	remaining := make([]string, 0, len(args))

	for _, arg := range args {
		switch arg {
		case "--json":
			opts.JSONMode = true
		case "--no-interactive":
			opts.NoInteractive = true
		default:
			remaining = append(remaining, arg)
		}
	}

	if opts.JSONMode {
		opts.NoInteractive = true
	}

	opts.Args = remaining
	return opts
}

// HandledError は既に JSON 出力済みのエラーを示す型
// main.go はこの型のエラーを stderr に出力しない
type HandledError struct{}

func (e HandledError) Error() string { return "" }

// jsonError は JSON モード時に JSON エラーを Writer に書き込み HandledError を返す
// 非 JSON モード時は通常の error を返す
func jsonError(opts ExecOptions, code, message string) error {
	if opts.JSONMode {
		writeJSONError(opts.writer(), code, message)
		return HandledError{}
	}
	return fmt.Errorf("%s", message)
}

// writeJSON は v を JSON にエンコードして w に書き込む
func writeJSON(w io.Writer, v any) {
	enc := json.NewEncoder(w)
	enc.Encode(v) //nolint:errcheck
}

// writeJSONError は JSON エラーを w に書き込む
func writeJSONError(w io.Writer, code, message string) {
	writeJSON(w, errorJSON{Error: code, Message: message})
}
