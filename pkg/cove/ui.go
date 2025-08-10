package cove

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/timer"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Message types
type fileChangedMsg struct{}
type checkFileMsg struct{}

// ===== TODO LIST ITEM =====

type todoItem struct {
	todo *Todo
}

func (i todoItem) FilterValue() string { return i.todo.Description }

func (i todoItem) Title() string { 
	var checkbox string
	switch i.todo.State {
	case Done:
		checkbox = "[x]"
	default:
		if i.todo.TimeSpent > 0 {
			checkbox = "[*]"  // In progress
		} else {
			checkbox = "[ ]"  // Not started
		}
	}
	
	return fmt.Sprintf("%s %s", checkbox, i.todo.Description)
}

func (i todoItem) Description() string {
	if i.todo.TimeSpent > 0 {
		return fmt.Sprintf("spent: %v", i.todo.TimeSpent.Round(time.Minute))
	}
	return ""
}

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

// ===== TODO SELECTOR WITH BUBBLES LIST =====

type TodoSelectorModel struct {
	list         list.Model
	todos        []Todo
	filename     string
	lastModified time.Time
	spinner      spinner.Model
	loading      bool
}

func NewTodoSelector(todos []Todo, filename string) TodoSelectorModel {
	// Sort todos (completed items last)
	sortedTodos := sortTodos(todos)
	
	// Convert todos to list items
	items := make([]list.Item, len(sortedTodos))
	for i := range sortedTodos {
		items[i] = todoItem{todo: &sortedTodos[i]}
	}
	
	// Create list with default delegate
	delegate := list.NewDefaultDelegate()
	
	// Customize the delegate styles
	purple := lipgloss.Color("#7D56F4")
	green := lipgloss.Color("#01BE85")
	
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(green).
		BorderLeftForeground(green)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(green)
		
	l := list.New(items, delegate, 80, 20)
	l.Title = "📝 TODO Selector"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false) // Keep it simple for now
	l.SetShowHelp(true)
	
	// Style the title
	l.Styles.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(purple).
		Padding(0, 1)
	
	// Create spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(purple)
	
	// Get initial file modification time
	modTime := time.Time{}
	if stat, err := os.Stat(filename); err == nil {
		modTime = stat.ModTime()
	}
	
	return TodoSelectorModel{
		list:         l,
		todos:        sortedTodos,
		filename:     filename,
		lastModified: modTime,
		spinner:      s,
		loading:      false,
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
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil
		
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			if selectedItem, ok := m.list.SelectedItem().(todoItem); ok {
				timerModel := NewBubblesTimer(selectedItem.todo, m)
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
			
			// Update list items
			items := make([]list.Item, len(sortedTodos))
			for i := range sortedTodos {
				items[i] = todoItem{todo: &sortedTodos[i]}
			}
			m.list.SetItems(items)
		}
		return m, m.checkFile()
		
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}
	
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)
	
	return m, tea.Batch(cmds...)
}

func (m TodoSelectorModel) View() string {
	if m.loading {
		return fmt.Sprintf("\n   %s Loading todos...\n\n", m.spinner.View())
	}
	
	return "\n" + m.list.View()
}

// ===== TIMER WITH BUBBLES TIMER =====

type TimerModel struct {
	todo        *Todo
	parentModel TodoSelectorModel
	timer       timer.Model
	startTime   time.Time
}

func NewBubblesTimer(todo *Todo, parent TodoSelectorModel) TimerModel {
	t := timer.NewWithInterval(todo.EstimatedTime, time.Second)
	
	return TimerModel{
		todo:        todo,
		parentModel: parent,
		timer:       t,
		startTime:   time.Now(),
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
			// Calculate elapsed time using the timer's remaining time
			elapsed := m.todo.EstimatedTime - m.timer.Timeout
			if elapsed > 0 {
				m.todo.AddTime(elapsed)
			}
			
			// Write updated todos back to file
			if err := WriteTodos(m.parentModel.filename, m.parentModel.todos); err != nil {
				// Handle error silently for now
			}
			
			return m.parentModel, m.parentModel.checkFile()
		case "d":
			// Calculate elapsed time using the timer's remaining time
			elapsed := m.todo.EstimatedTime - m.timer.Timeout
			if elapsed > 0 {
				m.todo.AddTime(elapsed)
			}
			m.todo.MarkDone()
			
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
		
		// Add highlighting to minutes if we can parse the timer format
		if strings.Contains(timerText, ":") {
			parts := strings.Split(timerText, ":")
			if len(parts) == 2 {
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
				
				timerDisplay := minutesStyle.Render(parts[0]) + "  " + colonStyle.Render(":") + "  " + secondsStyle.Render(parts[1])
				s.WriteString(timerDisplay)
			} else {
				// Fallback to regular timer display
				timerStyle := lipgloss.NewStyle().
					Bold(true).
					Foreground(lipgloss.Color("#FAFAFA"))
				s.WriteString(timerStyle.Render(timerText))
			}
		} else {
			// Fallback to regular timer display
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