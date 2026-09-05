package main

import (
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
)

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "New Task"
	ti.SetWidth(50)

	vi := viewport.New()
	vi.SetWidth(50)
	vi.SetHeight(120)

	return model{
		tasks:     []Task{},
		textInput: ti,
		viewport:  vi,
		inserting: false,
		modifying: false,
	}
}

func (m model) Init() tea.Cmd {
	return getTasks
}
