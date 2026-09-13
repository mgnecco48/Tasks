package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	tea "charm.land/bubbletea/v2"
)

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

// Changing a task:
func modifyTask(id int, newBody string) tea.Cmd {
	return func() tea.Msg {

		body, err := json.Marshal(struct {
			ID   int    `json:"id"`
			Body string `json:"body"`
		}{
			ID:   id,
			Body: newBody,
		})
		if err != nil {
			return errMsg{err}
		}

		req, err := http.NewRequest(
			http.MethodPatch,
			fmt.Sprintf("http://127.0.0.1:8000/tasks/%d/", id),
			bytes.NewReader(body),
		)
		if err != nil {
			return errMsg{err}
		}

		req.Header.Set("Content-Type", "application/json")

		c := &http.Client{Timeout: 10 * time.Second}
		resp, err := c.Do(req)
		if err != nil {
			return errMsg{err}
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return errMsg{fmt.Errorf("Could not modify the task:\nbody: %s, id: %d", newBody, id)}
		}

		return taskModifiedMsg{id: id, newBody: newBody}
	}
}

type taskModifiedMsg struct {
	id      int
	newBody string
}
