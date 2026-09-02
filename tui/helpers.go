package main

import "fmt"

// Helper function and types to flatten the rows so the cursor and completion work individually.
type taskRow struct {
	task   *Task
	indent int
}

func taskRows(tasks []Task, indent int) []taskRow {
	rows := []taskRow{}

	for i := range tasks {
		rows = append(rows, taskRow{
			task:   &tasks[i],
			indent: indent,
		})

		rows = append(rows, taskRows(tasks[i].Children, indent+1)...)
	}
	return rows
}

// Helper to control row rendering in a single place
func printNiceRow(m model, i int, row taskRow) string {
	s := ""
	cursor := " "
	if m.cursor == i {
		cursor = "\033[91m>\033[0m"
	}
	completed := nicePrint("", iconsStyle)
	if row.task.IsCompleted {
		completed = nicePrint("", iconsStyle)
	}

	if row.indent == 0 {
		if row.task.IsCompleted {
			s += fmt.Sprintf("%s %s %s\n", cursor, completed, nicePrint(row.task.Body, completedStyle))
		} else {
			s += fmt.Sprintf("%s %s %s\n", cursor, completed, row.task.Body)
		}
	} else {
		if row.task.IsCompleted {
			s += fmt.Sprintf("%s    󱞩 %s %s\n", cursor, completed, nicePrint(row.task.Body, completedStyle))
		} else {
			s += fmt.Sprintf("%s    󱞩 %s %s\n", cursor, completed, row.task.Body)
		}
	}
	return s

}
