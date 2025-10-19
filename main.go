package main

import (
	"fmt"
	"os"

	"scottgcooper-cbz-webp-converter/tui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-isatty"
)

func main() {
	// Check if we should run in CLI mode (for backwards compatibility)
	if len(os.Args) > 1 && os.Args[1] == "--cli" {
		runCLIMode()
		return
	}

	// Check if we're in an interactive terminal
	if !isatty.IsTerminal(os.Stdin.Fd()) || !isatty.IsTerminal(os.Stdout.Fd()) {
		fmt.Println("This program requires an interactive terminal.")
		fmt.Println("Please run it in a terminal or use --cli mode for non-interactive use.")
		os.Exit(1)
	}

	// Run TUI mode
	p := tea.NewProgram(tui.InitialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running TUI: %v\n", err)
		os.Exit(1)
	}
}

func runCLIMode() {
	// This is the original CLI functionality for backwards compatibility
	// You can implement this if needed
	fmt.Println("CLI mode not implemented. Use the TUI interface instead.")
}
