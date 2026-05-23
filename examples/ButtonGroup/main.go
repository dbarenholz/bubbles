package main

import (
	"fmt"
	"os"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/dbarenholz/bubbles/button"
	"github.com/dbarenholz/bubbles/buttongroup"
)

type keymap struct {
	Quit key.Binding

	NextDemo key.Binding
	PrevDemo key.Binding

	Tab      key.Binding
	ShiftTab key.Binding
	Left     key.Binding
	Right    key.Binding
	Up       key.Binding
	Down     key.Binding
}

func defaultKeyMap() keymap {
	return keymap{
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c", "q"),
			key.WithHelp("ctrl+c/q", "quit"),
		),
		NextDemo: key.NewBinding(
			key.WithKeys("}", "n"),
			key.WithHelp("n/}", "next demo"),
		),
		PrevDemo: key.NewBinding(
			key.WithKeys("{", "p"),
			key.WithHelp("p/{", "previous demo"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next button"),
		),
		ShiftTab: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "previous button"),
		),
		Left: key.NewBinding(
			key.WithKeys("left"),
			key.WithHelp("left", "left button"),
		),
		Right: key.NewBinding(
			key.WithKeys("right"),
			key.WithHelp("right", "right button"),
		),
		Up: key.NewBinding(
			key.WithKeys("up"),
			key.WithHelp("up", "top button"),
		),
		Down: key.NewBinding(
			key.WithKeys("down"),
			key.WithHelp("down", "down button"),
		),
	}
}

// unused
func (k keymap) ShortHelp() []key.Binding { return []key.Binding{} }
func (k keymap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Quit},
		{k.NextDemo, k.PrevDemo},
		{k.Tab, k.ShiftTab, k.Left, k.Right, k.Up, k.Down},
	}
}

type demo int

const (
	HORIZONTAL_DEMO demo = iota
	VERTICAL_DEMO
	GRID_DEMO
)

type model struct {
	horizontal   buttongroup.Model
	vertical     buttongroup.Model
	grid         buttongroup.Model
	focussedDemo demo
	lastEvent    string
	keys         keymap
	help         help.Model
	quitting     bool
}

func newModel() model {
	h := makeHorizontal()
	v := makeVertical()
	g := makeGrid()

	v.Blur()
	g.Blur()

	hlp := help.New()
	hlp.ShowAll = true

	return model{
		horizontal:   h,
		vertical:     v,
		grid:         g,
		focussedDemo: 0,
		keys:         defaultKeyMap(),
		help:         hlp,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// handle key messages
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		switch {
		case key.Matches(keyMsg, m.keys.Quit):
			m.quitting = true
			return m, tea.Quit
		case key.Matches(keyMsg, m.keys.NextDemo):
			m.focussedDemo = (m.focussedDemo + 1) % 3
			m.applyDemoFocus()
			return m, nil
		case key.Matches(keyMsg, m.keys.PrevDemo):
			m.focussedDemo = (m.focussedDemo - 1) % 3
			if m.focussedDemo < 0 {
				m.focussedDemo += 3
			}
			m.applyDemoFocus()
			return m, nil
		}
	}

	// pass through all other ones to active demo
	return m, m.updateActiveDemo(msg)
}

func (m *model) applyDemoFocus() {
	m.horizontal.Blur()
	m.vertical.Blur()
	m.grid.Blur()

	switch m.focussedDemo {
	case HORIZONTAL_DEMO:
		m.horizontal.Focus()
	case VERTICAL_DEMO:
		m.vertical.Focus()
	case GRID_DEMO:
		m.grid.Focus()
	}
}

func (m *model) updateActiveDemo(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	switch m.focussedDemo {
	case HORIZONTAL_DEMO:
		m.horizontal, cmd = m.horizontal.Update(msg)
		m.recordEvents("Horizontal", m.horizontal, msg)
	case VERTICAL_DEMO:
		m.vertical, cmd = m.vertical.Update(msg)
		m.recordEvents("Vertical", m.vertical, msg)
	case GRID_DEMO:
		m.grid, cmd = m.grid.Update(msg)
		m.recordEvents("Grid", m.grid, msg)
	}

	return cmd
}

func (m *model) recordEvents(name string, group buttongroup.Model, msg tea.Msg) {
	if didPress, id, index := group.DidPress(msg); didPress {
		m.lastEvent = fmt.Sprintf("%s: pressed button #%d at index %d", name, id, index)
	}
	if didPress, id, index := group.DidDisabledPress(msg); didPress {
		m.lastEvent = fmt.Sprintf("%s: disabled press on button #%d at index %d", name, id, index)
	}
}

func (m model) View() tea.View {
	if m.quitting {
		return tea.NewView("")
	}

	active := []string{" ", " ", " "}
	active[m.focussedDemo] = "*"

	var s strings.Builder
	s.WriteString("ButtonGroup demo\n\n")
	fmt.Fprintf(&s, "%s Horizontal\n", active[0])
	s.WriteString(m.horizontal.View())
	s.WriteString("\n\n")
	fmt.Fprintf(&s, "%s Vertical\n", active[1])
	s.WriteString(m.vertical.View())
	s.WriteString("\n\n")
	fmt.Fprintf(&s, "%s Grid\n", active[2])
	s.WriteString(m.grid.View())
	s.WriteString("\n\n")
	if m.lastEvent == "" {
		s.WriteString("Event: none yet\n")
	} else {
		s.WriteString("Event: " + m.lastEvent + "\n")
	}
	s.WriteString("\n")
	s.WriteString(m.help.View(m.keys))

	v := tea.NewView(s.String())
	v.AltScreen = true
	return v
}

func makeHorizontal() buttongroup.Model {
	yes := button.New("Yes")
	disabled := button.New("Disabled")
	maybe := button.New("Maybe")
	group := buttongroup.HorizontalGroup(yes, disabled, maybe)
	group.SetDisabled(disabled.ID(), true)
	return group
}

func makeVertical() buttongroup.Model {
	one := button.New("One")
	one.SetStyle(button.Parens)
	two := button.New("Two")
	two.SetStyle(button.Parens)
	three := button.New("Three")
	three.SetStyle(button.Parens)
	four := button.New("Four")
	four.SetStyle(button.Parens)
	group := buttongroup.VerticalGroup(one, two, three, four)
	group.SetDisabled(four.ID(), true)
	return group
}

func makeGrid() buttongroup.Model {
	btnA := button.New("A")
	btnA.SetStyle(button.BoxDrawing)
	btnB := button.New("B")
	btnB.SetStyle(button.BoxDrawing)
	btnC := button.New("C")
	btnC.SetStyle(button.BoxDrawing)
	btnD := button.New("D")
	btnD.SetStyle(button.BoxDrawing)
	btnE := button.New("E")
	btnE.SetStyle(button.BoxDrawing)
	btnF := button.New("F")
	btnF.SetStyle(button.BoxDrawing)
	btnG := button.New("G")
	btnG.SetStyle(button.BoxDrawing)

	group := buttongroup.ButtonGrid(3, 3, btnA, btnB, btnC, btnD, btnE, btnF, btnG)
	group.SetDisabled(btnE.ID(), true)
	return group
}

func main() {
	if _, err := tea.NewProgram(newModel()).Run(); err != nil {
		fmt.Printf("Could not start program :(\n%v\n", err)
		os.Exit(1)
	}
}
