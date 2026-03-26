package cmd

import (
	"fmt"

	"worklog/internal/ui"
)

// showHelp はヘルプメッセージを表示する
func showHelp(opts ExecOptions) {
	output := ui.RenderHelp()
	fmt.Fprint(opts.writer(), output)
}
