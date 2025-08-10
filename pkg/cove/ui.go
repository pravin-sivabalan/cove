package cove

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/timer"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Message types
type fileChangedMsg struct{}
type checkFileMsg struct{}


// Helper function to sort todos (completed items last)
func sortTodos(todos []Todo) []Todo {
	sorted := make([]Todo, len(todos))
	copy(sorted, todos)
	
	// Sort: not completed first, then completed
	// Within each group, maintain original order
	var notCompleted, completed []Todo
	
	for _, todo := range sorted {
		if todo.State == Done {
			completed = append(completed, todo)
		} else {
			notCompleted = append(notCompleted, todo)
		}
	}
	
	// Combine: not completed first, then completed
	result := make([]Todo, 0, len(sorted))
	result = append(result, notCompleted...)
	result = append(result, completed...)
	
	return result
}

// ===== TODO SELECTOR =====

type TodoSelectorModel struct {
	todos        []Todo
	filename     string
	lastModified time.Time
	spinner      spinner.Model
	loading      bool
	cursor       int
}

func NewTodoSelector(todos []Todo, filename string) TodoSelectorModel {
	// Sort todos (completed items last)
	sortedTodos := sortTodos(todos)
	
	// Create spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))
	
	// Get initial file modification time
	modTime := time.Time{}
	if stat, err := os.Stat(filename); err == nil {
		modTime = stat.ModTime()
	}
	
	return TodoSelectorModel{
		todos:        sortedTodos,
		filename:     filename,
		lastModified: modTime,
		spinner:      s,
		loading:      false,
		cursor:       0,
	}
}

func (m TodoSelectorModel) Init() tea.Cmd {
	return tea.Batch(
		m.checkFile(),
		m.spinner.Tick,
	)
}

func (m TodoSelectorModel) checkFile() tea.Cmd {
	return tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
		return checkFileMsg{}
	})
}

func (m TodoSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.todos)-1 {
				m.cursor++
			}
		case "enter":
			if m.cursor < len(m.todos) {
				timerModel := NewBubblesTimer(&m.todos[m.cursor], m, m.cursor)
				return timerModel, timerModel.Init()
			}
		}
		
	case checkFileMsg:
		// Check if file has been modified
		if stat, err := os.Stat(m.filename); err == nil {
			if stat.ModTime().After(m.lastModified) {
				m.lastModified = stat.ModTime()
				m.loading = true
				return m, func() tea.Msg { return fileChangedMsg{} }
			}
		}
		// Continue checking
		return m, m.checkFile()
		
	case fileChangedMsg:
		m.loading = false
		// Reload todos from file
		if newTodos, err := ReadTodos(m.filename); err == nil {
			// Reconcile old todos with new ones
			reconciledTodos := ReconcileTodos(m.todos, newTodos)
			// Sort todos (completed items last)
			sortedTodos := sortTodos(reconciledTodos)
			m.todos = sortedTodos
			
			// Keep cursor within bounds
			if m.cursor >= len(m.todos) {
				m.cursor = len(m.todos) - 1
			}
			if m.cursor < 0 {
				m.cursor = 0
			}
		}
		return m, m.checkFile()
		
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}
	
	return m, tea.Batch(cmds...)
}

func (m TodoSelectorModel) View() string {
	if m.loading {
		return fmt.Sprintf("\n   %s Loading todos...\n\n", m.spinner.View())
	}
	
	var s strings.Builder
	
	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1)
	
	s.WriteString(titleStyle.Render("📝 TODO Selector"))
	s.WriteString("\n\n")
	
	// Show todos in markdown style
	for i, todo := range m.todos {
		var checkbox string
		switch todo.State {
		case Done:
			checkbox = "[x]"
		default:
			if todo.TimeSpent > 0 {
				checkbox = "[*]"
			} else {
				checkbox = "[ ]"
			}
		}
		
		// Create the todo line
		todoLine := fmt.Sprintf("- %s %s", checkbox, todo.Description)
		
		// Style for cursor selection
		if i == m.cursor {
			selectedStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#01BE85")).
				Bold(true)
			s.WriteString("> " + selectedStyle.Render(todoLine))
		} else {
			s.WriteString("  " + todoLine)
		}
		
		s.WriteString("\n")
		
		// Add time info on next line if available
		if todo.TimeSpent > 0 {
			timeInfo := fmt.Sprintf("    spent: %v", todo.TimeSpent.Round(time.Minute))
			if i == m.cursor {
				selectedStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("#01BE85"))
				s.WriteString("  " + selectedStyle.Render(timeInfo))
			} else {
				timeStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("#888888"))
				s.WriteString("  " + timeStyle.Render(timeInfo))
			}
			s.WriteString("\n")
		}
	}
	
	// Help text
	s.WriteString("\n")
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#555555"))
	s.WriteString(helpStyle.Render("↑/↓ or j/k: navigate • enter: start timer • q: quit"))
	
	return s.String()
}

// ===== TIMER WITH BUBBLES TIMER =====

type TimerModel struct {
	todo        *Todo
	parentModel TodoSelectorModel
	timer       timer.Model
	startTime   time.Time
	todoIndex   int
}

func NewBubblesTimer(todo *Todo, parent TodoSelectorModel, todoIndex int) TimerModel {
	t := timer.NewWithInterval(todo.EstimatedTime, time.Second)
	
	return TimerModel{
		todo:        todo,
		parentModel: parent,
		timer:       t,
		startTime:   time.Now(),
		todoIndex:   todoIndex,
	}
}

func (m TimerModel) Init() tea.Cmd {
	return m.timer.Init()
}

func (m TimerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case " ":
			if m.timer.Running() {
				return m, m.timer.Stop()
			} else {
				return m, m.timer.Start()
			}
		case "h":
			// Calculate elapsed time since timer started
			elapsed := time.Since(m.startTime)
			if elapsed > 0 && m.todoIndex < len(m.parentModel.todos) {
				m.parentModel.todos[m.todoIndex].AddTime(elapsed)
			}
			
			// Write updated todos back to file
			if err := WriteTodos(m.parentModel.filename, m.parentModel.todos); err != nil {
				// Handle error silently for now
			}
			
			return m.parentModel, m.parentModel.checkFile()
		case "d":
			// Calculate elapsed time since timer started
			elapsed := time.Since(m.startTime)
			if elapsed > 0 && m.todoIndex < len(m.parentModel.todos) {
				m.parentModel.todos[m.todoIndex].AddTime(elapsed)
				m.parentModel.todos[m.todoIndex].MarkDone()
			}
			
			// Write updated todos back to file
			if err := WriteTodos(m.parentModel.filename, m.parentModel.todos); err != nil {
				// Handle error silently for now
			}
			
			return m.parentModel, m.parentModel.checkFile()
		case "y":
			if m.timer.Timedout() {
				newTimer := timer.NewWithInterval(m.todo.EstimatedTime, time.Second)
				m.timer = newTimer
				m.startTime = time.Now()
				return m, m.timer.Init()
			}
		case "n":
			if m.timer.Timedout() {
				return m.parentModel, m.parentModel.checkFile()
			}
		}
		
	case timer.TickMsg:
		var cmd tea.Cmd
		m.timer, cmd = m.timer.Update(msg)
		return m, cmd
		
	case timer.StartStopMsg:
		var cmd tea.Cmd
		m.timer, cmd = m.timer.Update(msg)
		return m, cmd
		
	case timer.TimeoutMsg:
		return m, nil
	}
	
	var cmd tea.Cmd
	m.timer, cmd = m.timer.Update(msg)
	return m, cmd
}

func (m TimerModel) View() string {
	var s strings.Builder
	
	// Task name with highlighting (like before)
	taskStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(1, 2)
	
	s.WriteString(taskStyle.Render(m.todo.Description))
	s.WriteString("\n\n\n")
	
	if m.timer.Timedout() {
		// Completion state
		completeStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#01BE85"))
		
		s.WriteString(completeStyle.Render("✅ COMPLETE"))
		s.WriteString("\n\n")
		
		durationMinutes := int(m.todo.EstimatedTime.Minutes())
		
		statusStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))
		
		s.WriteString(statusStyle.Render(fmt.Sprintf("Add another %d minutes? (y/n)", durationMinutes)))
		s.WriteString("\n\n")
		
		controlsStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#555555"))
		
		s.WriteString(controlsStyle.Render("d: mark done  h: switch task  q: quit"))
	} else {
		// Parse the timer to get minutes and seconds
		timerText := m.timer.View()
		
		// Define styles for enhanced timer display
		minutesStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#F25D94")).
			Padding(0, 1)
		
		secondsStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA"))
		
		colonStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#888888"))
		
		// Try to parse different timer formats
		var minutes, seconds string
		parsed := false
		
		// Format 1: "9:30" or "09:30"
		if strings.Contains(timerText, ":") {
			parts := strings.Split(timerText, ":")
			if len(parts) == 2 {
				minutes = strings.TrimSpace(parts[0])
				seconds = strings.TrimSpace(parts[1])
				parsed = true
			}
		}
		
		// Format 2: "9m30s" or similar
		if !parsed && strings.Contains(timerText, "m") && strings.Contains(timerText, "s") {
			// Extract minutes and seconds from format like "9m30s"
			text := timerText
			if mIndex := strings.Index(text, "m"); mIndex > 0 {
				minutes = text[:mIndex]
				remaining := text[mIndex+1:]
				if sIndex := strings.Index(remaining, "s"); sIndex > 0 {
					seconds = remaining[:sIndex]
					parsed = true
				}
			}
		}
		
		if parsed {
			// Create enhanced timer display with highlighted minutes
			timerDisplay := minutesStyle.Render(minutes) + "  " + colonStyle.Render(":") + "  " + secondsStyle.Render(seconds)
			s.WriteString(timerDisplay)
		} else {
			// Fallback to regular timer display if parsing fails
			timerStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FAFAFA"))
			s.WriteString(timerStyle.Render(timerText))
		}
		
		s.WriteString("\n\n")
		
		// Status line
		statusStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))
		
		if m.timer.Running() {
			s.WriteString(statusStyle.Render("remaining"))
		} else {
			s.WriteString(statusStyle.Render("⏸️ PAUSED"))
		}
		
		s.WriteString("\n\n")
		
		controlsStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#555555"))
		
		s.WriteString(controlsStyle.Render("space: resume  h: switch  d: done  q: quit"))
	}
	
	return s.String()
}