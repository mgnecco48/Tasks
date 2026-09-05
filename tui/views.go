package main

import (
	"fmt"

	"charm.land/lipgloss/v2"
)

func (m model) normalView() string {
	s := ""
	rows := taskRows(m.tasks, 0)
	for i, row := range rows {
		s += printNiceRow(m, i, row)
	}
	return s
}

func (m model) parentInsertView() string {
	s := ""
	rows := taskRows(m.tasks, 0)
	for i, row := range rows {
		s += printNiceRow(m, i, row)
	}

	s += fmt.Sprintf("  %s\n", m.textInput.View())

	return s
}

func (m model) childrenInsertView() string {
	rows := taskRows(m.tasks, 0)
	parentPos := m.cursor
	s := ""

	for i, row := range rows[:parentPos+1] {
		s += printNiceRow(m, i, row)
	}

	s += fmt.Sprintf("     %s\n", m.textInput.View())

	for i, row := range rows[parentPos+1:] {
		s += printNiceRow(m, i, row)
	}
	return s
}

func (m model) taskModifyView() string {
	rows := taskRows(m.tasks, 0)
	currentPos := m.cursor
	s := ""

	for i, row := range rows[:currentPos] {
		s += printNiceRow(m, i, row)
	}

	if rows[m.cursor].task.ParentId != nil {
		s += fmt.Sprintf("     %s\n", m.textInput.View())
	} else {
		s += fmt.Sprintf("  %s\n", m.textInput.View())

	}
	for i, row := range rows[currentPos+1:] {
		s += printNiceRow(m, i, row)
	}
	return s
}

type statusBar struct {
	mode  string
	space string
	extra string
}

func (m model) statusBar() string {
	var bar statusBar

	if m.inserting || m.modifying {
		bar.mode = nicePrint("INSERT", insertModeStyle)
		bar.space = nicePrint("                              ", barSpaceStyle)
		bar.extra = nicePrint("😁 : 👍🏻", insertModeStyle)
	} else if m.err != nil {
		bar.mode = nicePrint("ERROR", errorStatusStyle)
		bar.space = m.err.Error()
		bar.extra = nicePrint("😁 : 👍🏻", errorStatusStyle)
	} else {
		bar.mode = nicePrint("NORMAL", normalModeStyle)
		bar.space = nicePrint("                              ", barSpaceStyle)
		bar.extra = nicePrint("😁 : 👍🏻", normalModeStyle)
	}
	return lipgloss.JoinHorizontal(lipgloss.Bottom, bar.mode, bar.space, bar.extra)
}
