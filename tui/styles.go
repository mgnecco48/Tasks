package main

import (
	"charm.land/lipgloss/v2"
)

var titleStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.BrightWhite).
	Width(30).
	AlignHorizontal(lipgloss.Center).
	Underline(true)

var iconsStyle = lipgloss.NewStyle().
	Foreground(lipgloss.BrightCyan)

// TODO: Add striketrhough style for the completed tasks

func nicePrint(text string, style lipgloss.Style) string {
	return lipgloss.Sprint(style.Render(text))
}

var completedStyle = lipgloss.NewStyle().
	Strikethrough(true)

var insertModeStyle = lipgloss.NewStyle().
	AlignHorizontal(lipgloss.Center).
	Background(lipgloss.Yellow).
	Foreground(lipgloss.Black).
	PaddingLeft(1).
	PaddingRight(1)

var normalModeStyle = lipgloss.NewStyle().
	AlignHorizontal(lipgloss.Center).
	Background(lipgloss.Magenta).
	Foreground(lipgloss.Black).
	PaddingLeft(1).
	PaddingRight(1)
