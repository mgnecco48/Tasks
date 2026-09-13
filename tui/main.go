package main

import (
	"fmt"
	"os"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const (
	tree_url   string = "http://127.0.0.1:8000/tasks/tree"
	create_url string = "http://127.0.0.1:8000/tasks/"
	delete_url string = "http://127.0.0.1:8000/tasks/"
	modify_url string = "http://127.0.0.1:8000/tasks/"
)

type model struct {
	tasks           []Task
	cursor          int
	err             error
	textInput       textinput.Model
	inserting       bool
	modifying       bool
	focusedParentId *int
	showError       bool
	width           int
	height          int
}

type Task struct {
	Id          int    `json:"id"`
	Body        string `json:"body"`
	Priority    int    `json:"priority"`
	IsCompleted bool   `json:"is_completed"`
	ParentId    *int   `json:"parent_id"`
	Children    []Task `json:"children"`
}

type TaskCreate struct {
	Body     string `json:"body"`
	ParentId *int   `json:"parent_id"`
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var cmd tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, getTasks

	case taskMsg:
		m.tasks = []Task(msg)

		rows := taskRows(m.tasks, 0)
		if len(rows) == 0 {
			m.cursor = 0
		} else if m.cursor >= len(rows) {
			m.cursor = len(rows) - 1
		}
		return m, nil

	case errMsg:
		m.err = msg.err
		m.showError = true
		return m, nil

	case taskCompletionUpdatedMsg:
		return m, tea.Batch(tea.ClearScreen, getTasks)

	case taskCompletionFailedMsg:
		rows := taskRows(m.tasks, 0)

		for _, row := range rows {
			if row.task.Id == msg.id {
				row.task.IsCompleted = !msg.completed
				break
			}
		}

		m.err = msg.err
		return m, tea.Batch(tea.ClearScreen)

	case createTaskMsg:
		return m, tea.Batch(tea.ClearScreen, getTasks)

	case taskDeletedMsg:
		return m, tea.Batch(tea.ClearScreen, getTasks)

	case taskModifiedMsg:
		return m, tea.Batch(tea.ClearScreen, getTasks)

	case tea.KeyPressMsg:
		rows := taskRows(m.tasks, 0)

		if m.modifying {
			switch msg.String() {
			case "enter":
				id := rows[m.cursor].task.Id
				newBody := m.textInput.Value()
				m.textInput.Reset()
				m.textInput.Blur()
				m.modifying = false
				return m, tea.Batch(tea.ClearScreen, modifyTask(id, newBody))
			case "esc":
				m.textInput.Reset()
				m.textInput.Blur()
				m.modifying = false
				return m, nil

			}
			m.textInput, cmd = m.textInput.Update(msg)

		} else if m.inserting {
			switch msg.String() {
			case "enter":
				taskBody := TaskCreate{
					Body:     m.textInput.Value(),
					ParentId: m.focusedParentId,
				}
				m.textInput.Reset()
				m.textInput.Blur()
				m.inserting = false
				m.focusedParentId = nil
				return m, tea.Batch(tea.ClearScreen, createTask(taskBody))
			case "esc":
				m.textInput.Reset()
				m.textInput.Blur()
				m.inserting = false
				m.focusedParentId = nil
				return m, tea.ClearScreen

			}
			m.textInput, cmd = m.textInput.Update(msg)
		} else if m.showError {
			switch msg.String() {
			case "esc", "enter":
				m.showError = false
				return m, nil
			}

		} else {
			switch msg.String() {
			case "ctrl+c", "q":
				if m.inserting == false {
					return m, tea.Quit
				}
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(rows)-1 {
					m.cursor++
				}
			case "enter":
				if len(rows) > 0 {
					task := rows[m.cursor].task
					task.IsCompleted = !task.IsCompleted
					return m, updateTaskCompletion(task.Id, task.IsCompleted)
				}
			case "A":
				if len(rows) > 0 {
					task := rows[m.cursor].task
					if task.ParentId != nil {
						m.focusedParentId = task.ParentId
					} else {
						m.focusedParentId = &task.Id
					}
					if m.inserting == false {
						m.inserting = true
						m.textInput.Focus()
						return m, nil
					}
				}
			case "N":
				m.inserting = true
				m.textInput.Focus()
				m.focusedParentId = nil
				return m, nil
			case "D":
				if len(rows) > 0 {
					id := rows[m.cursor].task.Id
					return m, deleteTask(id)
				}
			case "C":
				currTaskBody := rows[m.cursor].task.Body
				m.modifying = true
				m.textInput.Focus()
				m.textInput.SetValue(currTaskBody)
				return m, nil

			case "R":
				return m, getTasks

			}
		}
	}
	return m, cmd
}

func (m model) View() tea.View {

	W, H := m.getDimensions()

	if m.err != nil {
		return tea.NewView(fmt.Sprintf("\nThere was an error: %v\n\n", m.err))
	}

	insertingIcon := ""
	if m.inserting || m.modifying {
		insertingIcon = "\033[91m \033[0m"
	}
	tabText := fmt.Sprintf("TASKS %s", insertingIcon)
	titleTab := nicePrint(tabText, titleStyle)

	taskBox := ""
	if m.inserting {
		if m.focusedParentId != nil {
			taskBox += nicePrint(m.childrenInsertView(), todoBoxStyle.Width(W).Height(H-lipgloss.Height(titleTab)))
		} else {
			taskBox += nicePrint(m.parentInsertView(), todoBoxStyle.Width(W).Height(H-lipgloss.Height(titleTab)))
		}
	} else if m.modifying {
		taskBox += nicePrint(m.taskModifyView(), todoBoxStyle.Width(W).Height(H-lipgloss.Height(titleTab)))
	} else {
		taskBox += nicePrint(m.normalView(), todoBoxStyle.Width(W))
	}

	screen := lipgloss.JoinVertical(lipgloss.Left, titleTab, taskBox)
	v := tea.NewView(screen)
	v.AltScreen = true
	return v

}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Print("There was a problem 😭", err)
		os.Exit(1)
	}
}

// TODO: Apply my new line logic. Fix scrolling
// TODO: Add priority funcitionality
// TODO: Add Extra details lookup.
// TODO: Add due dates.
// TODO: Show priority in the thing
// TODO: Add multiple lists, need to fix the backend aswell to do this.
// TODO: Add write error messages to the databse to handle the error gracefully. rightnow  i just return the error but dont rerender the good tasks.
// TODO: Add Undo option for whoopsies
