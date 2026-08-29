package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"
)

const url string = "http://127.0.0.1:8000/tasks/tree"

type model struct {
	tasks    []Task
	cursor   int
	selected map[int]struct{}
	err      error
	// TODO: Add pending map[int]bool if you want to show which tasks are currently syncing.
}

// TODO: get my  taskf from the backend and transform them into GO structs that i can us to render.
type Task struct {
	Id          int    `json:"id"`
	Body        string `json:"body"`
	Priority    int    `json:"priority"`
	IsCompleted bool   `json:"is_completed"`
	ParentId    int    `json:"parent_id"`
	Children    []Task `json:"children"`
}

func getTasks() tea.Msg {

	c := &http.Client{Timeout: 10 * time.Second}
	resp, err := c.Get(url)
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

// TODO: Add updateTaskCompletion(id int, completed bool) tea.Cmd here.

// This are so that bubbletea can actually interpret the "messages"
type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

type taskMsg []Task

// TODO: Add messages for the completion update request.
// Example: taskCompletionUpdatedMsg for success and taskCompletionFailedMsg for failure.

// helper function and types to flatten the rows so the cursor and completion work individually.
type taskRow struct {
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

//

func initialModel() model {
	return model{
		tasks: []Task{},
	}
}

func (m model) Init() tea.Cmd {
	return getTasks
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case taskMsg:
		m.tasks = []Task(msg)
		return m, nil

	case errMsg:
		m.err = msg
		// TODO: Do not quit for a failed completion update later; show a non-fatal error instead.
		return m, tea.Quit

	case tea.KeyPressMsg:
		rows := taskRows(m.tasks, 0)
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(rows)-1 {
				m.cursor++
			}
		case "enter", "space":
			if len(rows) > 0 {
				rows[m.cursor].task.IsCompleted = !rows[m.cursor].task.IsCompleted
				// TODO: After this optimistic local toggle, return updateTaskCompletion(...).
				// Use the task id and the new IsCompleted value, not a generic "toggle" request.
			}
		}
	}
	return m, nil
}

func (m model) View() tea.View {
	if m.err != nil {
		return tea.NewView(fmt.Sprintf("\nThere was an error: %v\n\n", m.err))
	}

	s := "Today's Tasks:\n\n"

	rows := taskRows(m.tasks, 0)
	for i, row := range rows {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		completed := ""
		if row.task.IsCompleted {
			completed = ""
		}

		if row.indent == 0 {
			s += fmt.Sprintf("%s %s %s\n", cursor, completed, row.task.Body)
		} else {
			s += fmt.Sprintf("%s    󱞩 %s %s\n", cursor, completed, row.task.Body)
		}

	}

	s += "\nPress q or ctrl+c to quit.\n"

	return tea.NewView(s)

}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Print("There was a problem 😭")
		os.Exit(1)
	}
}
