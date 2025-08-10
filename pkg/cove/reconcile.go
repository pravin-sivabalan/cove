package cove

import (
	"strings"
)

// ReconcileTodos matches old todos with new todos from the file
// It handles:
// - Todos moving to different lines
// - Todo descriptions changing
// - New todos being added
// - Todos being deleted
func ReconcileTodos(oldTodos, newTodos []Todo) []Todo {
	if len(oldTodos) == 0 {
		return newTodos
	}

	// Create a copy of new todos to modify
	reconciledTodos := make([]Todo, len(newTodos))
	copy(reconciledTodos, newTodos)

	// Track which old todos have been matched
	matched := make([]bool, len(oldTodos))

	// First pass: exact line number matches (todos that haven't moved)
	for i, newTodo := range reconciledTodos {
		for j, oldTodo := range oldTodos {
			if matched[j] {
				continue
			}
			if oldTodo.LineNumber == newTodo.LineNumber && 
			   similarDescriptions(oldTodo.Description, newTodo.Description) {
				// Keep time spent and state from old todo if new todo isn't done
				if newTodo.State == Open && oldTodo.TimeSpent > 0 {
					reconciledTodos[i].TimeSpent = oldTodo.TimeSpent
				}
				if newTodo.State == Open && oldTodo.State == Done {
					// If old todo was done but new todo isn't, keep it as done
					reconciledTodos[i].State = oldTodo.State
				}
				matched[j] = true
				break
			}
		}
	}

	// Second pass: match by description similarity (todos that moved)
	for i, newTodo := range reconciledTodos {
		if newTodo.TimeSpent > 0 {
			continue // Already matched
		}
		
		for j, oldTodo := range oldTodos {
			if matched[j] {
				continue
			}
			
			// Match by description similarity
			if similarDescriptions(oldTodo.Description, newTodo.Description) {
				// Transfer time spent and state
				if newTodo.State == Open && oldTodo.TimeSpent > 0 {
					reconciledTodos[i].TimeSpent = oldTodo.TimeSpent
				}
				if newTodo.State == Open && oldTodo.State == Done {
					reconciledTodos[i].State = oldTodo.State
				}
				matched[j] = true
				break
			}
		}
	}

	// Third pass: match by position for unmatched todos (best effort)
	unmatched := []int{}
	for j := range oldTodos {
		if !matched[j] {
			unmatched = append(unmatched, j)
		}
	}

	newTodoIndex := 0
	for _, oldIndex := range unmatched {
		// Find the next unmatched new todo
		for newTodoIndex < len(reconciledTodos) {
			if reconciledTodos[newTodoIndex].TimeSpent == 0 {
				if reconciledTodos[newTodoIndex].State == Open && oldTodos[oldIndex].TimeSpent > 0 {
					reconciledTodos[newTodoIndex].TimeSpent = oldTodos[oldIndex].TimeSpent
				}
				if reconciledTodos[newTodoIndex].State == Open && oldTodos[oldIndex].State == Done {
					reconciledTodos[newTodoIndex].State = oldTodos[oldIndex].State
				}
				newTodoIndex++
				break
			}
			newTodoIndex++
		}
	}

	return reconciledTodos
}

// similarDescriptions checks if two descriptions are similar enough to be considered the same todo
func similarDescriptions(desc1, desc2 string) bool {
	// Exact match
	if desc1 == desc2 {
		return true
	}
	
	// Case-insensitive match
	if strings.EqualFold(desc1, desc2) {
		return true
	}
	
	// Check if one is a substring of the other (for minor edits)
	desc1Lower := strings.ToLower(strings.TrimSpace(desc1))
	desc2Lower := strings.ToLower(strings.TrimSpace(desc2))
	
	if len(desc1Lower) > 0 && len(desc2Lower) > 0 {
		// If one description contains the other and they're at least 70% similar
		if strings.Contains(desc1Lower, desc2Lower) || strings.Contains(desc2Lower, desc1Lower) {
			shorter := len(desc1Lower)
			if len(desc2Lower) < shorter {
				shorter = len(desc2Lower)
			}
			longer := len(desc1Lower)
			if len(desc2Lower) > longer {
				longer = len(desc2Lower)
			}
			
			// If the shorter description is at least 70% of the longer one
			if float64(shorter)/float64(longer) >= 0.7 {
				return true
			}
		}
	}
	
	return false
}