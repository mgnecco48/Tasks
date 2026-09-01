package main

import (
	"charm.land/lipgloss/v2"
)

var titleStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Yellow).
	Width(50).
	AlignHorizontal(lipgloss.Center).
	Underline(true)

var iconsStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Yellow)

// TODO: Add striketrhough style for the completed tasks

func nicePrint(text string, style lipgloss.Style) string {
	return lipgloss.Sprint(style.Render(text))
}
