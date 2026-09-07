package main

import (
	"charm.land/lipgloss/v2"
)

func nicePrint(text string, style lipgloss.Style) string {
	return lipgloss.Sprint(style.Render(text))
}

func tabBorder(left, middle, right string) lipgloss.Border {
	border := lipgloss.RoundedBorder()
	border.BottomLeft = left
	border.BottomRight = right
	border.Bottom = middle

	return border
}

var titleStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Black).
	Background(lipgloss.Color("#00e5ee")).
	AlignHorizontal(lipgloss.Center).
	Padding(0, 1)

var iconsStyle = lipgloss.NewStyle().
	Foreground(lipgloss.BrightCyan)

var prefixStyle = lipgloss.NewStyle()
var uncompletedStyle = lipgloss.NewStyle()
var completedStyle = lipgloss.NewStyle().Strikethrough(true)

var todoBoxStyle = lipgloss.NewStyle().
	PaddingRight(1).
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("#00e5ee"))

var helpStyle = lipgloss.NewStyle().
	Width(60).
	Foreground(lipgloss.Color("#525252"))
