package main

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	args := os.Args[1:]
	clog := initClog(ctx, args)

	app := tea.NewProgram(
		clog,
		tea.WithAltScreen(),
	)

	if err := app.Start(); err != nil {
		cancel()
		fmt.Printf("error: %v", err)
		os.Exit(1)
	}
}
