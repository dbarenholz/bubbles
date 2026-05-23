package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/dbarenholz/bubbles/button"
)

type keymap struct {
	Quit        key.Binding
	ToBracket   key.Binding
	ToParen     key.Binding
	ToBox       key.Binding
	ToggleState key.Binding
	Press       key.Binding
}

// unused
func (k keymap) ShortHelp() []key.Binding { return []key.Binding{} }
func (k keymap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Quit, k.ToggleState, k.Press},
		{k.ToBracket, k.ToParen, k.ToBox},
	}
}

func defaultKeyMap() keymap {
	return keymap{
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c", "q"),
			key.WithHelp("ctrl+c/q", "quit"),
		),
		ToBracket: key.NewBinding(
			key.WithKeys("[", "]"),
			key.WithHelp("[/]", "bracket style"),
		),
		ToParen: key.NewBinding(
			key.WithKeys("(", ")"),
			key.WithHelp("(/)", "paren style"),
		),
		ToBox: key.NewBinding(
			key.WithKeys("#"),
			key.WithHelp("#", "box drawing style"),
		),
		ToggleState: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "toggle disabled"),
		),
		Press: key.NewBinding(
			key.WithKeys("space", "enter"),
			key.WithHelp("enter/space", "press"),
		),
	}
}

type model struct {
	btn      button.Model
	keys     keymap
	help     help.Model
	count    int
	quitting bool
	msg      string
}

type clearMsg struct{}

func clearErrorAfter(t time.Duration) tea.Cmd {
	return tea.Tick(t, func(_ time.Time) tea.Msg {
		return clearMsg{}
	})
}

func newModel() model {
	btn := button.New("Press me!")
	btn.Focus()
	help := help.New()
	help.ShowAll = true
	return model{
		btn:  btn,
		keys: defaultKeyMap(),
		help: help,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			m.quitting = true
			return m, tea.Quit
		case key.Matches(msg, m.keys.ToBracket):
			m.btn.SetStyle(button.Brackets)
		case key.Matches(msg, m.keys.ToParen):
			m.btn.SetStyle(button.Parens)
		case key.Matches(msg, m.keys.ToBox):
			m.btn.SetStyle(button.BoxDrawing)
		case key.Matches(msg, m.keys.ToggleState):
			enabled, _ := m.btn.State()
			m.btn.SetDisabled(enabled)
		}

	case clearMsg:
		m.msg = ""
	}

	var cmd tea.Cmd
	m.btn, cmd = m.btn.Update(msg)

	// did we press the button?
	if m.btn.DidPress(msg) {
		m.count++
	}

	// did we try to press a disabled button?
	if m.btn.DidDisabledPress(msg) {
		m.msg = "Button is disabled!"
		return m, tea.Batch(cmd, clearErrorAfter(2*time.Second))
	}

	return m, cmd
}

func (m model) View() tea.View {
	if m.quitting {
		return tea.NewView("")
	}

	var b strings.Builder
	b.WriteString("A cute button stands before you.\n")
	fmt.Fprintf(&b, "\n%s\n\n", m.btn.View())
	fmt.Fprintf(&b, "Pressed count: %d\n", m.count)

	if m.msg != "" {
		fmt.Fprintf(&b, "Message: %s\n", m.msg)
	}
	help := m.help.View(m.keys)

	v := tea.NewView(b.String() + help)
	v.AltScreen = true
	return v
}

func main() {
	if _, err := tea.NewProgram(newModel()).Run(); err != nil {
		fmt.Printf("Could not start program :(\n%v\n", err)
		os.Exit(1)
	}
}
