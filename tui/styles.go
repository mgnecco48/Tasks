package main

import (
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/exp/charmtone"
)

func nicePrint(text string, style lipgloss.Style) string {
	return lipgloss.Sprint(style.Render(text))
}

var titleStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.BrightWhite).
	Width(50).
	AlignHorizontal(lipgloss.Center).
	Underline(true)

var iconsStyle = lipgloss.NewStyle().
	Foreground(lipgloss.BrightCyan)

var completedStyle = lipgloss.NewStyle().
	Strikethrough(true)

/* Status Bar Stuff*/
var insertModeStyle = lipgloss.NewStyle().
	AlignHorizontal(lipgloss.Center).
	Background(lipgloss.Yellow).
	Foreground(lipgloss.Black).
	PaddingLeft(1).
	PaddingRight(1).
	Width(10)

var normalModeStyle = lipgloss.NewStyle().
	AlignHorizontal(lipgloss.Center).
	Background(lipgloss.Magenta).
	Foreground(lipgloss.Black).
	PaddingLeft(1).
	PaddingRight(1).
	Width(10)

var errorStatusStyle = lipgloss.NewStyle().
	AlignHorizontal(lipgloss.Center).
	Background(lipgloss.Red).
	Foreground(lipgloss.Black).
	PaddingLeft(1).
	PaddingRight(1).
	Width(10)

var barSpaceStyle = lipgloss.NewStyle().
	Width(30).
	Background(lipgloss.Black)

var todoBoxStyle = lipgloss.NewStyle().
	PaddingLeft(1).
	PaddingRight(1).
	Border(lipgloss.RoundedBorder()).
	BorderForegroundBlend(
		charmtone.Cherry,
		charmtone.Charple,
		charmtone.Guac,
		charmtone.Charple,
		charmtone.Sriracha)
