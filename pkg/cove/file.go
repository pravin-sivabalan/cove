package cove

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
)

var todoRegex = regexp.MustCompile(`^[\s]*-\s+\[([x\s])\]\s+(.+)$`)
var starRegex = regexp.MustCompile(`\*+`)

func ReadTodos(filename string) ([]Todo, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var todos []Todo
	scanner := bufio.NewScanner(file)
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		if match := todoRegex.FindStringSubmatch(line); match != nil {
			checkbox := strings.TrimSpace(match[1])
			description := strings.TrimSpace(match[2])
			
			// Check for timer hints (stars)
			var todo Todo
			if starMatch := starRegex.FindString(description); starMatch != "" {
				starCount := len(starMatch)
				estimatedMinutes := starCount * 5 // each star = 5 minutes
				// Remove stars from description
				cleanDescription := strings.TrimSpace(starRegex.ReplaceAllString(description, ""))
				todo = NewTodoWithEstimate(cleanDescription, estimatedMinutes)
			} else {
				todo = NewTodo(description)
			}
			
			if checkbox == "x" {
				todo.State = Done
			}
			
			todo.OriginalLine = line
			todo.LineNumber = lineNumber
			
			todos = append(todos, todo)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return todos, nil
}

func WriteTodos(filename string, todos []Todo) error {
	// Read all lines from the original file
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file for reading: %w", err)
	}
	defer file.Close()

	var allLines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		allLines = append(allLines, scanner.Text())
	}
	
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	// Update lines with completed todos
	for _, todo := range todos {
		if todo.LineNumber > 0 && todo.LineNumber <= len(allLines) {
			lineIndex := todo.LineNumber - 1
			
			// Generate the updated line
			checkbox := " "
			if todo.State == Done {
				checkbox = "x"
			}
			
			// Format time spent
			timeSpentStr := ""
			if todo.TimeSpent > 0 {
				minutes := int(todo.TimeSpent.Minutes())
				if minutes > 0 {
					timeSpentStr = fmt.Sprintf(" (took %dm)", minutes)
				}
			}
			
			// Reconstruct the line with stars if they were originally present
			stars := ""
			if todo.EstimatedTime != 20*time.Minute {
				starCount := int(todo.EstimatedTime.Minutes()) / 5
				stars = strings.Repeat("*", starCount)
				if stars != "" {
					stars = " " + stars
				}
			}
			
			// Extract indentation from original line
			indent := ""
			if match := regexp.MustCompile(`^(\s*)`).FindStringSubmatch(todo.OriginalLine); match != nil {
				indent = match[1]
			}
			
			newLine := fmt.Sprintf("%s- [%s] %s%s%s", indent, checkbox, todo.Description, stars, timeSpentStr)
			allLines[lineIndex] = newLine
		}
	}

	// Write back to file
	outFile, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	for _, line := range allLines {
		if _, err := fmt.Fprintln(outFile, line); err != nil {
			return fmt.Errorf("error writing line: %w", err)
		}
	}

	return nil
}