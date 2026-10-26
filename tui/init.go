package main

import (
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "New Task"

	return model{
		tasks:     []Task{},
		lists:     []List{},
		textInput: ti,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(getTasks, getLists)
}
