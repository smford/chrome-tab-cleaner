package main

import (
	"fmt"
	"os"

	"github.com/smford/chrome-tab-cleaner/internal/chrome"
	"github.com/smford/chrome-tab-cleaner/internal/tui"
)

func main() {
	client := chrome.NewClient()
	app := tui.NewApp(client)

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running application: %v\n", err)
		os.Exit(1)
	}
}
