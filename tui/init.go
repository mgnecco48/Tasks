package main

import (
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "New Task"
	ti.SetWidth(50)
	return model{
		tasks:     []Task{},
		textInput: ti,
		inserting: false,
		modifying: false,
	}
}

func (m model) Init() tea.Cmd {
	return getTasks
}
