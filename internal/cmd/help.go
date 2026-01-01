package cmd

import (
	"fmt"

	"worklog/internal/ui"
)

// showHelp はヘルプメッセージを表示する
func showHelp() {
	output := ui.RenderHelp()
	fmt.Print(output)
}
