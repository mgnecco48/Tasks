package main

import (
	"fmt"

	"charm.land/lipgloss/v2"
)

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

func (m model) getDimensions() (int, int) {
	return m.width, m.height
}

// Helper to control row rendering in a single place
func printNiceRow(m model, i int, row taskRow) string {

	W, _ := m.getDimensions()

	s := ""
	cursor := " "
	if m.cursor == i {
		cursor = "\033[91m>\033[0m"
	}
	completed := nicePrint("", iconsStyle)
	if row.task.IsCompleted {
		completed = nicePrint("", iconsStyle)
	}
	parentPrefix := nicePrint(fmt.Sprintf("%s %s ", cursor, completed), prefixStyle)
	childPrefix := nicePrint(fmt.Sprintf("%s    󱞩 %s ", cursor, completed), prefixStyle)

	pPrefixW := lipgloss.Width(parentPrefix)
	cPrefixW := lipgloss.Width(childPrefix)

	if row.indent == 0 {
		if row.task.IsCompleted {
			s += lipgloss.JoinHorizontal(lipgloss.Top, parentPrefix, nicePrint(row.task.Body, completedStyle.Width(W-pPrefixW)))
		} else {
			s += lipgloss.JoinHorizontal(lipgloss.Top, parentPrefix, nicePrint(row.task.Body, uncompletedStyle.Width(W-pPrefixW)))
		}
	} else {
		if row.task.IsCompleted {
			s += lipgloss.JoinHorizontal(lipgloss.Top, childPrefix, nicePrint(row.task.Body, completedStyle.Width(W-cPrefixW)))
		} else {
			s += lipgloss.JoinHorizontal(lipgloss.Top, childPrefix, nicePrint(row.task.Body, uncompletedStyle.Width(W-cPrefixW)))
		}
	}
	return s

}
