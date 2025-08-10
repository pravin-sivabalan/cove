# 🏔️ Cove

**A focused CLI Pomodoro timer that reads todos from markdown files**

Cove is a terminal-based productivity tool that combines the power of markdown todo lists with time-tracking functionality. Work on your tasks with purpose, track your progress automatically, and maintain focus with a clean, distraction-free interface.

## ✨ Features

### 📋 **Smart Todo Management**
- **Markdown Integration**: Reads standard markdown todo lists (`- [ ] Task`, `- [x] Done`)
- **Intelligent Status Display**: Visual progress indicators (`[ ]`, `[*]`, `[x]`)
- **Smart Sorting**: Active todos at top, completed items at bottom
- **Real-time File Sync**: Automatically detects external file changes
- **Time Tracking**: Automatically records time spent on each task

### ⏱️ **Flexible Timer System** 
- **Custom Durations**: Use star hints (`*` = 5 minutes) for custom timing
- **Precision Timing**: Built on professional Bubbletea timer components  
- **Pause & Resume**: Full timer control with visual feedback
- **Task Switching**: Seamlessly move between tasks while preserving time
- **Auto-persistence**: Time tracking saved to your markdown file

### 🎨 **Professional Interface**
- **Clean Design**: Task-focused UI built with Charm's Bubbles components
- **Keyboard-Driven**: Efficient navigation with standard hotkeys
- **Visual Feedback**: Loading states, status indicators, and progress tracking
- **Responsive Layout**: Adapts to your terminal size

## 🚀 Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/cove.git
cd cove

# Build the application
go build -o cove

# Run with your markdown file
./cove your-todos.md
```

### Usage

**Create a markdown todo file:**
```markdown
# My Tasks

- [ ] Review project proposals
- [ ] Write documentation **
- [ ] Call client about requirements *
- [x] Set up development environment
```

**Run Cove:**
```bash
./cove my-tasks.md
```

## 📖 How It Works

### Todo Selector
Navigate your tasks with a professional list interface:
- **`↑/↓` or `j/k`**: Navigate between tasks
- **`Enter`**: Start working on selected task
- **`q`**: Quit application

```
📝 TODO Selector

[ ] Review project proposals
Not Started

[*] Write documentation  
spent: 15m

[x] Set up development environment
spent: 45m
```

### Cove Timer
Focus on your work with a clean, task-centered timer:
- **`Space`**: Pause/resume timer
- **`Esc`**: Switch to another task (saves time)
- **`d`**: Mark current task as done
- **`q`**: Quit application

```
┌────────────────────────────────────────┐
│        Write documentation             │
└────────────────────────────────────────┘

              Cove Timer

┌─────────────┐
│   15:30     │
└─────────────┘

space: pause/resume • esc: switch task • d: mark done • q: quit
```

## ⚙️ Timer Hints

Control task duration with star notation in your markdown:

| Notation | Duration | Example |
|----------|----------|---------|
| `*` | 5 minutes | `- [ ] Quick call *` |
| `**` | 10 minutes | `- [ ] Code review **` |
| `****` | 20 minutes | `- [ ] Deep work ****` |
| (none) | 20 minutes | `- [ ] Default task` |

## 📁 File Format

Cove reads and writes standard markdown todo lists:

**Input Format:**
```markdown
# My Project

- [ ] Task not started
- [ ] Task with custom time **
- [x] Completed task
```

**After Working (Auto-updated):**
```markdown  
# My Project

- [*] Task in progress (took 15m)
- [ ] Task with custom time **
- [x] Completed task (took 25m)
```

## 🔄 Live File Sync

Cove automatically detects when your markdown file changes externally:
- **Smart Reconciliation**: Matches existing todos even if moved or edited
- **Preserves Time Data**: Time tracking survives file reorganization  
- **Real-time Updates**: UI refreshes automatically when file changes
- **Conflict Resolution**: Intelligent merging of external changes

## 🏗️ Technical Details

**Built With:**
- **Go**: High-performance, cross-platform compatibility
- **Bubbletea**: Professional terminal user interface framework
- **Bubbles Components**: List, Timer, Spinner for robust UI elements
- **Lipgloss**: Beautiful terminal styling and layout

**Architecture:**
- **Single Binary**: No dependencies, easy deployment
- **File-based**: Works with any markdown editor
- **Event-driven**: Responsive, efficient terminal interface
- **Stateless**: Your markdown file is the source of truth

## 🎯 Philosophy

Cove is designed around these core principles:

- **Markdown-Native**: Your todos live in standard markdown files
- **Non-Intrusive**: Works alongside your existing workflow  
- **Focus-First**: Clean interface that emphasizes the current task
- **Time-Aware**: Automatic tracking without manual overhead
- **Terminal-Friendly**: Fast, keyboard-driven, works anywhere

## 🛠️ Development

```bash
# Clone and setup
git clone https://github.com/yourusername/cove.git
cd cove

# Install dependencies  
go mod download

# Run tests
go test ./...

# Build
go build -o cove
```

**Project Structure:**
```
cove/
├── main.go          # Application entry point
├── todo.go          # Todo data structures  
├── file.go          # Markdown reading/writing
├── ui.go            # Bubbletea UI components
├── reconcile.go     # Smart todo reconciliation
└── watcher.go       # File watching functionality
```

## 📝 License

MIT License - see LICENSE file for details.

---

**Focus. Track. Accomplish.** 🏔️

*Built with ❤️ using [Charm](https://charm.sh/)*