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

func renderPriority(priority int) string {
	symbol := "󱥸"
	style := lipgloss.NewStyle().Width(5)

	switch priority {
	case 1:
		style = style.Foreground(lipgloss.Red)
	case 2:
		style = style.Foreground(lipgloss.Yellow)
	case 3:
		style = style.Foreground(lipgloss.Green)
	}

	return nicePrint(symbol, style)
}

// Helper to control row rendering in a single place
func printNiceRow(m model, i int, row taskRow) string {

	W, _ := m.getDimensions()

	style := lipgloss.NewStyle()
	s := ""
	cursor := " "
	if m.cursor == i {
		cursor = "\033[91m>\033[0m"
		style = style.Background(lipgloss.Color("#3c3836"))
	}
	completed := nicePrint("", iconsStyle)
	if row.task.IsCompleted {
		completed = nicePrint("", iconsStyle)
	}

	parentPrefix := nicePrint(fmt.Sprintf("%s %s ", cursor, completed), prefixStyle)
	childPrefix := nicePrint(fmt.Sprintf("%s    󱞩 %s ", cursor, completed), prefixStyle)

	pPrefixW := lipgloss.Width(parentPrefix)
	cPrefixW := lipgloss.Width(childPrefix)

	prioritySymbol := renderPriority(row.task.Priority)
	pSymbolW := lipgloss.Width(prioritySymbol)

	if row.indent == 0 {
		if row.task.IsCompleted {
			s += lipgloss.JoinHorizontal(lipgloss.Top, parentPrefix, nicePrint(row.task.Body, completedStyle.Width(W-pPrefixW-pSymbolW)), prioritySymbol)
		} else {
			s += lipgloss.JoinHorizontal(lipgloss.Top, parentPrefix, nicePrint(row.task.Body, uncompletedStyle.Width(W-pPrefixW-pSymbolW)), prioritySymbol)
		}
	} else {
		if row.task.IsCompleted {
			s += lipgloss.JoinHorizontal(lipgloss.Top, childPrefix, nicePrint(row.task.Body, completedStyle.Width(W-cPrefixW-pSymbolW)), prioritySymbol)
		} else {
			s += lipgloss.JoinHorizontal(lipgloss.Top, childPrefix, nicePrint(row.task.Body, uncompletedStyle.Width(W-cPrefixW-pSymbolW)), prioritySymbol)
		}
	}
	return s

}
