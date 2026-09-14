package main

import (
	"charm.land/lipgloss/v2"
	"fmt"
)

type key struct {
	key         string
	description string
}

var (
	Up            = key{"k/↑", "move-up"}
	Down          = key{"j/↓", "move-down"}
	Quit          = key{"q", "quit"}
	Help          = key{"?", "toggle help"}
	DeleteTask    = key{"D", "delete focused task"}
	NewParentTask = key{"N", "new main task"}
	NewChildTask  = key{"A", "new child task"}
	ModifyTask    = key{"M", "modify focused task"}
	Esc           = key{"Esc", "cancel"}
	NormalEnter   = key{"Enter", "toggle completion"}
	InsertEnter   = key{"Enter", "save"}
)

func (m model) ShortHelp() []key {
	if m.inserting || m.modifying {
		return []key{Esc, InsertEnter}
	} else {
		return []key{Up, Down, Quit, Help}
	}
}

func (m model) LongHelp() []key {
	return []key{
		Up,
		Down,
		Quit,
		DeleteTask,
		ModifyTask,
		NewParentTask, NewChildTask,
		NormalEnter,
	}
}

func (m model) helpMenu() string {
	var keys []key
	if m.showHelp == false {
		keys = m.ShortHelp()
	} else {
		keys = m.LongHelp()
	}
	view := ""
	keyStyle := helpStyle.AlignHorizontal(lipgloss.Right).Bold(true).Padding(0, 1, 0, 1)
	descStyle := helpStyle.AlignHorizontal(lipgloss.Left)
	for _, key := range keys {
		if m.showHelp == true {
			view += fmt.Sprintln(lipgloss.JoinHorizontal(lipgloss.Bottom, nicePrint(key.key, keyStyle), nicePrint(key.description, descStyle)))
		} else {
			view += lipgloss.JoinHorizontal(lipgloss.Bottom, nicePrint(key.key, keyStyle), nicePrint(key.description, descStyle))
		}
	}
	return view
}
