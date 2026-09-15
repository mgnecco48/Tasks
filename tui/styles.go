package main

import (
	"charm.land/lipgloss/v2"
)

func nicePrint(text string, style lipgloss.Style) string {
	return style.Render(text)
}

var firstTabBorder = lipgloss.Border{
	Top:         "─",
	Bottom:      "",
	Left:        "│",
	Right:       "│",
	TopLeft:     "╭",
	TopRight:    "╮",
	BottomLeft:  "",
	BottomRight: "",
}

var titleStyle = lipgloss.NewStyle().
	Bold(true).
	Border(firstTabBorder).BorderBottom(false).
	BorderForeground(lipgloss.Color("#00e5ee")).
	Foreground(lipgloss.Black).
	Background(lipgloss.Color("#00e5ee")).
	AlignHorizontal(lipgloss.Center).
	Padding(0, 1)

var iconsStyle = lipgloss.NewStyle().
	Foreground(lipgloss.BrightCyan)

var prefixStyle = lipgloss.NewStyle()
var uncompletedStyle = lipgloss.NewStyle().
	Foreground(lipgloss.BrightWhite)
var completedStyle = uncompletedStyle.
	Strikethrough(true).
	Faint(true)

var todoBoxStyle = lipgloss.NewStyle().
	PaddingRight(1).
	Border(lipgloss.RoundedBorder()).BorderTop(false).
	BorderForeground(lipgloss.Color("#00e5ee"))

var helpStyle = lipgloss.NewStyle().
	Foreground(lipgloss.White)

var helpBoxStyle = helpStyle.
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("#00e5ee")).
	Padding(1, 1, 0, 1).
	Align(lipgloss.Left)
