package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const defaultAPIURL = "http://127.0.0.1:8000"

var (
	appStyle = lipgloss.NewStyle().
			Padding(2, 4)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("229")).
			Background(lipgloss.Color("63")).
			Padding(0, 1)

	apiStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244"))

	loadingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	inputBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(0, 1)

	inputLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("111")).
			Bold(true)

	selectedTaskStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("229")).
				Background(lipgloss.Color("238"))

	completedTaskStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("244")).
				Strikethrough(true)

	checkboxStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	emptyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")).
			Italic(true)
)

type task struct {
	ID           int     `json:"id"`
	Body         string  `json:"body"`
	ExtraDetails *string `json:"extra_details"`
	IsCompleted  bool    `json:"is_completed"`
	ParentID     *int    `json:"parent_id"`
}

type visibleTask struct {
	task  task
	depth int
}

type model struct {
	apiURL      string
	tasks       []task
	visible     []visibleTask
	cursor      int
	loading     bool
	error       string
	adding      bool
	input       string
	inputParent *int
}

type tasksLoadedMsg struct {
	tasks []task
	err   error
}

type taskToggledMsg struct {
	err error
}

type taskAddedMsg struct {
	err error
}

func main() {
	apiURL := strings.TrimRight(os.Getenv("TASKS_API_URL"), "/")
	if apiURL == "" {
		apiURL = defaultAPIURL
	}

	m := model{apiURL: apiURL, loading: true}
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func (m model) Init() tea.Cmd {
	return fetchTasks(m.apiURL)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.adding {
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "esc":
				m.adding = false
				m.input = ""
				m.inputParent = nil
				return m, nil
			case "enter":
				body := strings.TrimSpace(m.input)
				if body == "" {
					return m, nil
				}

				parentID := m.inputParent
				m.adding = false
				m.input = ""
				m.inputParent = nil
				m.loading = true
				m.error = ""
				return m, addTask(m.apiURL, body, parentID)
			case "backspace", "ctrl+h":
				runes := []rune(m.input)
				if len(runes) > 0 {
					m.input = string(runes[:len(runes)-1])
				}
				return m, nil
			case " ":
				m.input += " "
				return m, nil
			}

			if msg.Type == tea.KeyRunes {
				m.input += string(msg.Runes)
			}

			return m, nil
		}

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "a":
			m.adding = true
			m.input = ""
			m.inputParent = nil
			m.error = ""
			return m, nil
		case "c":
			if len(m.visible) == 0 {
				return m, nil
			}

			parentID := m.visible[m.cursor].task.ID
			m.adding = true
			m.input = ""
			m.inputParent = &parentID
			m.error = ""
			return m, nil
		case "r":
			m.loading = true
			m.error = ""
			return m, fetchTasks(m.apiURL)
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.visible)-1 {
				m.cursor++
			}
		case " ":
			if len(m.visible) == 0 || m.loading {
				return m, nil
			}

			selected := m.visible[m.cursor].task
			m.loading = true
			m.error = ""
			return m, toggleTask(m.apiURL, selected.ID, !selected.IsCompleted)
		}
	case tasksLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.error = msg.err.Error()
			return m, nil
		}

		m.tasks = msg.tasks
		m.visible = buildVisibleTasks(msg.tasks)
		if m.cursor >= len(m.visible) {
			m.cursor = len(m.visible) - 1
		}
		if m.cursor < 0 {
			m.cursor = 0
		}
	case taskToggledMsg:
		if msg.err != nil {
			m.loading = false
			m.error = msg.err.Error()
			return m, nil
		}

		return m, fetchTasks(m.apiURL)
	case taskAddedMsg:
		if msg.err != nil {
			m.loading = false
			m.error = msg.err.Error()
			return m, nil
		}

		return m, fetchTasks(m.apiURL)
	}

	return m, nil
}

func (m model) View() string {
	var b strings.Builder

	b.WriteString(headerStyle.Render("Tasks TUI"))
	b.WriteString("\n")
	b.WriteString(apiStyle.Render(fmt.Sprintf("API: %s", m.apiURL)))
	b.WriteString("\n\n")

	if m.loading {
		b.WriteString(loadingStyle.Render("Loading..."))
		b.WriteString("\n")
	}

	if m.error != "" {
		b.WriteString(errorStyle.Render(fmt.Sprintf("Error: %s", m.error)))
		b.WriteString("\n\n")
	}

	if len(m.visible) == 0 && !m.loading {
		b.WriteString(emptyStyle.Render("No tasks found."))
		b.WriteString("\n")
	}

	if m.adding {
		label := "New task"
		if m.inputParent != nil {
			label = fmt.Sprintf("New child for task #%d", *m.inputParent)
		}

		input := fmt.Sprintf("%s\n%s", inputLabelStyle.Render(label), m.input)
		b.WriteString(inputBoxStyle.Render(input))
		b.WriteString("\n\n")
	}

	for i, row := range m.visible {
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}

		checkbox := "[ ]"
		if row.task.IsCompleted {
			checkbox = "[x]"
		}

		indent := strings.Repeat("  ", row.depth)
		body := row.task.Body
		if row.task.IsCompleted {
			body = completedTaskStyle.Render(body)
		}

		line := fmt.Sprintf("%s %s%s %s", cursor, indent, checkboxStyle.Render(checkbox), body)
		if i == m.cursor {
			line = selectedTaskStyle.Render(line)
		}

		b.WriteString(line)
		b.WriteString("\n")
	}

	b.WriteString("\n")
	if m.adding {
		b.WriteString(helpStyle.Render("enter: save | esc: cancel | ctrl+c: quit"))
		b.WriteString("\n")
	} else {
		b.WriteString(helpStyle.Render("j/down: move down | k/up: move up | space: toggle | a: add root | c: add child | r: refresh | q: quit"))
		b.WriteString("\n")
	}

	return appStyle.Render(b.String())
}

func fetchTasks(apiURL string) tea.Cmd {
	return func() tea.Msg {
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get(apiURL + "/tasks/")
		if err != nil {
			return tasksLoadedMsg{err: err}
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return tasksLoadedMsg{err: err}
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return tasksLoadedMsg{err: fmt.Errorf("GET /tasks/ failed: %s: %s", resp.Status, strings.TrimSpace(string(body)))}
		}

		var tasks []task
		if err := json.Unmarshal(body, &tasks); err != nil {
			return tasksLoadedMsg{err: err}
		}

		return tasksLoadedMsg{tasks: tasks}
	}
}

func toggleTask(apiURL string, id int, completed bool) tea.Cmd {
	return func() tea.Msg {
		payload, err := json.Marshal(map[string]bool{"is_completed": completed})
		if err != nil {
			return taskToggledMsg{err: err}
		}

		url := fmt.Sprintf("%s/tasks/%d/completion", apiURL, id)
		req, err := http.NewRequest(http.MethodPatch, url, bytes.NewReader(payload))
		if err != nil {
			return taskToggledMsg{err: err}
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return taskToggledMsg{err: err}
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return taskToggledMsg{err: err}
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return taskToggledMsg{err: fmt.Errorf("PATCH completion failed: %s: %s", resp.Status, strings.TrimSpace(string(body)))}
		}

		return taskToggledMsg{}
	}
}

func addTask(apiURL string, body string, parentID *int) tea.Cmd {
	return func() tea.Msg {
		payload, err := json.Marshal(struct {
			Body         string  `json:"body"`
			ExtraDetails *string `json:"extra_details"`
			ParentID     *int    `json:"parent_id"`
		}{
			Body:     body,
			ParentID: parentID,
		})
		if err != nil {
			return taskAddedMsg{err: err}
		}

		req, err := http.NewRequest(http.MethodPost, apiURL+"/tasks/", bytes.NewReader(payload))
		if err != nil {
			return taskAddedMsg{err: err}
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return taskAddedMsg{err: err}
		}
		defer resp.Body.Close()

		responseBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return taskAddedMsg{err: err}
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return taskAddedMsg{err: fmt.Errorf("POST /tasks/ failed: %s: %s", resp.Status, strings.TrimSpace(string(responseBody)))}
		}

		return taskAddedMsg{}
	}
}

func buildVisibleTasks(tasks []task) []visibleTask {
	childrenByParent := map[int][]task{}
	roots := []task{}
	tasksByID := map[int]task{}

	for _, t := range tasks {
		tasksByID[t.ID] = t
	}

	for _, t := range tasks {
		if t.ParentID == nil {
			roots = append(roots, t)
			continue
		}

		if _, ok := tasksByID[*t.ParentID]; !ok {
			roots = append(roots, t)
			continue
		}

		childrenByParent[*t.ParentID] = append(childrenByParent[*t.ParentID], t)
	}

	sortTasks(roots)
	for parentID := range childrenByParent {
		sortTasks(childrenByParent[parentID])
	}

	visible := []visibleTask{}
	var addTask func(task, int)
	addTask = func(t task, depth int) {
		visible = append(visible, visibleTask{task: t, depth: depth})
		for _, child := range childrenByParent[t.ID] {
			addTask(child, depth+1)
		}
	}

	for _, root := range roots {
		addTask(root, 0)
	}

	return visible
}

func sortTasks(tasks []task) {
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].ID < tasks[j].ID
	})
}
