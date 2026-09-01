package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

const (
	tree_url   string = "http://127.0.0.1:8000/tasks/tree"
	create_url string = "http://127.0.0.1:8000/tasks/"
	delete_url string = "http://127.0.0.1:8000/tasks/"
)

type model struct {
	tasks           []Task
	cursor          int
	err             error
	textInput       textinput.Model
	inserting       bool
	focusedParentId *int
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

// MY OWN Commands/Messages:
// Messages are what bubbletea reveices in the Update function to trigger a UI "refresh"
// Messages:
type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

type taskMsg []Task

// Command
func getTasks() tea.Msg {

	c := &http.Client{Timeout: 10 * time.Second}
	resp, err := c.Get(tree_url)
	if err != nil {
		return errMsg{err}
	}
	defer resp.Body.Close()

	var tasks []Task
	if err := json.NewDecoder(resp.Body).Decode(&tasks); err != nil {
		return errMsg{err}

	}

	return taskMsg(tasks)
}

// Messages for completeing a task and updating it in the database
type taskCompletionFailedMsg struct {
	id        int
	completed bool
	err       error
}

type taskCompletionUpdatedMsg struct {
	id        int
	completed bool
}

// Updating completion command
func updateTaskCompletion(id int, completed bool) tea.Cmd {
	return func() tea.Msg {
		body, err := json.Marshal(struct {
			IsCompleted bool `json:"is_completed"`
		}{
			IsCompleted: completed,
		})
		if err != nil {
			return taskCompletionFailedMsg{id: id, completed: completed, err: err}
		}

		req, err := http.NewRequest(
			http.MethodPatch,
			fmt.Sprintf("http://127.0.0.1:8000/tasks/%d/completion", id),
			bytes.NewReader(body),
		)
		if err != nil {
			return taskCompletionFailedMsg{id: id, completed: completed, err: err}
		}

		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return taskCompletionFailedMsg{id: id, completed: completed, err: err}
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return taskCompletionFailedMsg{
				id:        id,
				completed: completed,
				err:       fmt.Errorf("completion update failed: %s", resp.Status),
			}
		}

		return taskCompletionUpdatedMsg{id: id, completed: completed}
	}
}

// Create Task:
type createTaskMsg struct {
	created bool
}

func createTask(task TaskCreate) tea.Cmd {
	return func() tea.Msg {

		body, err := json.Marshal(task)
		if err != nil {
			return errMsg{err}
		}

		c := &http.Client{Timeout: 10 * time.Second}
		resp, err := c.Post(
			create_url,
			"application/json",
			bytes.NewReader(body),
		)
		if err != nil {
			return errMsg{err}
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return errMsg{fmt.Errorf("task creation failed: %s", resp.Status)}
		}

		return createTaskMsg{true}

	}
}

// Delete task:
type taskDeletedMsg struct {
	id int
}

func deleteTask(taskId int) tea.Cmd {
	return func() tea.Msg {

		req, err := http.NewRequest(
			http.MethodDelete,
			fmt.Sprintf("%s%d/", delete_url, taskId),
			nil,
		)
		if err != nil {
			return errMsg{err}
		}

		c := &http.Client{Timeout: 10 * time.Second}
		resp, err := c.Do(req)
		if err != nil {
			return errMsg{err}
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return errMsg{fmt.Errorf("task deletion failed: %s", resp.Status)}
		}

		return taskDeletedMsg{taskId}

	}
}

// ======================================

type taskRow struct { // helper function and types to flatten the rows so the cursor and completion work individually.
	task   *Task
	indent int
}

func taskRows(tasks []Task, indent int) []taskRow {
	rows := []taskRow{}

	for i := range tasks {
		rows = append(rows, taskRow{
			task:   &tasks[i],
			indent: indent,
		})

		rows = append(rows, taskRows(tasks[i].Children, indent+1)...)
	}
	return rows
}

// ======================================================

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "New Task"
	ti.SetWidth(50)
	return model{
		tasks:     []Task{},
		textInput: ti,
		inserting: false,
	}
}

func (m model) Init() tea.Cmd {
	return getTasks
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var cmd tea.Cmd

	switch msg := msg.(type) {

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
		m.err = msg
		return m, nil

	case taskCompletionUpdatedMsg:
		return m, getTasks

	case taskCompletionFailedMsg:
		rows := taskRows(m.tasks, 0)

		for _, row := range rows {
			if row.task.Id == msg.id {
				row.task.IsCompleted = !msg.completed
				break
			}
		}

		m.err = msg.err
		return m, nil

	case createTaskMsg:
		return m, tea.Batch(tea.ClearScreen, getTasks)

	case taskDeletedMsg:
		return m, tea.Batch(tea.ClearScreen, getTasks)

	case tea.KeyPressMsg:
		rows := taskRows(m.tasks, 0)

		if m.inserting {
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
			case "a":
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

			case "n":
				m.inserting = true
				m.textInput.Focus()
				m.focusedParentId = nil
				return m, nil

			case "D":
				if len(rows) > 0 {
					id := rows[m.cursor].task.Id
					tea.ClearScreen()
					return m, deleteTask(id)
				}
			}
		}
	}
	return m, cmd
}

func (m model) View() tea.View {
	if m.err != nil {
		return tea.NewView(fmt.Sprintf("\nThere was an error: %v\n\n", m.err))
	}

	insertingIcon := ""
	if m.inserting {
		insertingIcon = "\033[91m*\033[0m"
	}

	s := fmt.Sprintf(nicePrint("Today's Tasks:", titleStyle)+"%s\n", insertingIcon)

	if m.inserting == true && m.focusedParentId != nil {
		s += m.childrenInsertView()
		return tea.NewView(s)
	} else {
		s += m.normalView()
		v := tea.NewView(s)
		v.AltScreen = true
		return v
	}

}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Print("There was a problem 😭", err)
		os.Exit(1)
	}
}

// TODO: make it look nice with lipgloss, check other methods i could add (modify task?)
// TODO: add edit functionality
// TODO: Add priority funcitionality
// TODO: Add Extra details lookup.
// TODO: Add due dates.
// TODO: Add multiple lists, need to fix the backend aswell to do this.
