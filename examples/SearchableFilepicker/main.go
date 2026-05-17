package main

import (
	"errors"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	sfp "github.com/dbarenholz/bubbles/searchablefilepicker"
)

type model struct {
	searchableFilePicker sfp.Model
	selectedFile         string
	quitting             bool
	err                  error
}

type clearErrorMsg struct{}

func clearErrorAfter(t time.Duration) tea.Cmd {
	return tea.Tick(t, func(_ time.Time) tea.Msg {
		return clearErrorMsg{}
	})
}

func (m model) Init() tea.Cmd {
	return m.searchableFilePicker.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		// NOTE: only quit on 'q' if we aren't searching!
		case "ctrl+c", "q":
			if m.searchableFilePicker.Focus == sfp.FocusOnSearch {
				sfpModel, cmd := m.searchableFilePicker.Update(msg)
				m.searchableFilePicker = sfpModel
				return m, cmd
			} else {
				m.quitting = true
				return m, tea.Quit
			}
		}
	case clearErrorMsg:
		m.err = nil
	}

	var cmd tea.Cmd
	m.searchableFilePicker, cmd = m.searchableFilePicker.Update(msg)

	// Did the user select a file?
	if didSelect, path := m.searchableFilePicker.FilePicker.DidSelectFile(msg); didSelect {
		// Get the path of the selected file.
		m.selectedFile = path
	}

	// Did the user select a disabled file?
	// This is only necessary to display an error to the user.
	if didSelect, path := m.searchableFilePicker.FilePicker.DidSelectDisabledFile(msg); didSelect {
		// Let's clear the selectedFile and display an error.
		m.err = errors.New(path + " is not valid.")
		m.selectedFile = ""
		return m, tea.Batch(cmd, clearErrorAfter(2*time.Second))
	}

	return m, cmd
}

func (m model) View() tea.View {
	if m.quitting {
		return tea.NewView("")
	}
	var s strings.Builder
	s.WriteString("\n  ")
	if m.err != nil {
		s.WriteString(m.searchableFilePicker.FilePicker.Styles.DisabledFile.Render(m.err.Error()))
	} else if m.selectedFile == "" {
		s.WriteString("Pick a file:")
	} else {
		s.WriteString("Selected file: " + m.searchableFilePicker.FilePicker.Styles.Selected.Render(m.selectedFile))
	}
	s.WriteString("\n\n" + m.searchableFilePicker.View() + "\n")
	v := tea.NewView(s.String())
	v.AltScreen = true
	return v
}

func main() {
	fp := sfp.New(".") // pass in the directory to start in; don't set fp.CurrentDirectory to "." manually!
	// use any desired options from charm's filepicker, e.g.
	fp.FilePicker.AllowedTypes = []string{".mod", ".sum", ".go", ".txt", ".md"}
	fp.FilePicker.LoopEntries = true // loop selection

	m := model{searchableFilePicker: fp}
	tm, _ := tea.NewProgram(m).Run()
	mm := tm.(model)
	fmt.Println("\n  You selected: " + m.searchableFilePicker.FilePicker.Styles.Selected.Render(mm.selectedFile) + "\n")
}
