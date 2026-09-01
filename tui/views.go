package main

import "fmt"

func (m model) childrenInsertView() string {
	rows := taskRows(m.tasks, 0)
	parentPos := m.cursor
	s := ""

	for i, row := range rows[:parentPos+1] {
		cursor := " "
		if m.cursor == i {
			cursor = "\033[91m>\033[0m"
		}
		completed := nicePrint("", iconsStyle)
		if row.task.IsCompleted {
			completed = nicePrint("", iconsStyle)
		}

		if row.indent == 0 {
			s += fmt.Sprintf("%s %s %s\n", cursor, completed, row.task.Body)
		} else {
			s += fmt.Sprintf("%s    󱞩 %s %s\n", cursor, completed, row.task.Body)
		}

	}

	s += fmt.Sprintf("     %s\n", m.textInput.View())

	for _, row := range rows[parentPos+1:] {
		cursor := " "
		completed := nicePrint("", iconsStyle)
		if row.task.IsCompleted {
			completed = nicePrint("", iconsStyle)
		}

		if row.indent == 0 {
			s += fmt.Sprintf("%s %s %s\n", cursor, completed, row.task.Body)
		} else {
			s += fmt.Sprintf("%s    󱞩 %s %s\n", cursor, completed, row.task.Body)
		}

	}
	s += "\nPress 'enter' to insert or 'esc' to cancel.\n"

	return s
}

func (m model) normalView() string {
	s := ""
	rows := taskRows(m.tasks, 0)
	for i, row := range rows {
		cursor := " "
		if m.cursor == i {
			cursor = "\033[91m>\033[0m"
		}

		completed := nicePrint("", iconsStyle)
		if row.task.IsCompleted {
			completed = nicePrint("", iconsStyle)
		}

		if row.indent == 0 {
			s += fmt.Sprintf("%s %s %s\n", cursor, completed, row.task.Body)
		} else {
			s += fmt.Sprintf("%s    󱞩 %s %s\n", cursor, completed, row.task.Body)
		}

	}
	if m.inserting {
		s += fmt.Sprintf("  %s\n", m.textInput.View())
		s += "\nPress 'enter' to insert or 'esc' to cancel.\n"
		return s

	}
	s += "\nPress q or ctrl+c to quit.\n"

	return s
}
