package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"cove/pkg/cove"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <markdown-file>\n", os.Args[0])
		os.Exit(1)
	}

	filename := os.Args[1]
	
	todos, err := cove.ReadTodos(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading todos: %v\n", err)
		os.Exit(1)
	}

	model := cove.NewTodoSelector(todos, filename)
	
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}