package main

import (
	"fmt"

	"charm.land/lipgloss/v2"
)

func (m model) normalView() string {
	s := ""
	rows := taskRows(m.tasks, 0)
	for i, row := range rows {
		if i == len(rows)-1 {
			s += printNiceRow(m, i, row)
		} else {
			s += printNiceRow(m, i, row) + "\n"
		}
	}
	return s
}

func (m model) listView() string {
	lists := m.lists
	s := nicePrint("Your Lists", titleStyle.Align(lipgloss.Center)) + "\n\n"
	for i, list := range lists {
		s += nicePrint(fmt.Sprintf("%d-"+list.Name, i+1), listsStyle) + "\n"
		s += nicePrint("-"+list.Desc, listsStyle.Faint(true).Italic(true)) + "\n"
	}
	return s
}

func (m model) parentInsertView() string {
	s := ""
	rows := taskRows(m.tasks, 0)
	for i, row := range rows {
		s += printNiceRow(m, i, row) + "\n"
	}
	s += fmt.Sprintf("  %s", m.textInput.View())
	return s
}

func (m model) childrenInsertView() string {
	rows := taskRows(m.tasks, 0)
	parentPos := m.cursor
	s := ""

	for i, row := range rows[:parentPos+1] {
		s += printNiceRow(m, i, row) + "\n"
	}

	s += fmt.Sprintf("     %s\n", m.textInput.View())

	for i, row := range rows[parentPos+1:] {
		if i == len(rows)-1 {
			s += printNiceRow(m, i, row)
		} else {
			s += printNiceRow(m, i, row) + "\n"
		}
	}
	return s
}

func (m model) taskModifyView() string {
	rows := taskRows(m.tasks, 0)
	currentPos := m.cursor
	s := ""

	for i, row := range rows[:currentPos] {
		s += printNiceRow(m, i, row) + "\n"
	}

	if rows[m.cursor].task.ParentId != nil {
		s += fmt.Sprintf("     %s\n", m.textInput.View())
	} else {
		s += fmt.Sprintf("  %s\n", m.textInput.View())

	}
	for i, row := range rows[currentPos+1:] {
		if i == len(rows)-1 {
			s += printNiceRow(m, i, row)
		} else {
			s += printNiceRow(m, i, row) + "\n"
		}
	}
	return s
}
