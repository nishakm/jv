package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {

	p := tea.NewProgram(NewModel(),
		tea.WithAltScreen(),       // opens up a new terminal screen
		tea.WithMouseCellMotion()) // takes mouse input

	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
