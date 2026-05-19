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
	Quit     key.Binding
	NextDemo key.Binding
	PrevDemo key.Binding
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
	}
}

func (k keymap) ShortHelp() []key.Binding {
	return []key.Binding{k.Quit, k.NextDemo, k.PrevDemo}
}

func (k keymap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Quit, k.NextDemo, k.PrevDemo}}
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

func newModel() (model, error) {
	h, err := makeHorizontal()
	if err != nil {
		return model{}, err
	}
	v, err := makeVertical()
	if err != nil {
		return model{}, err
	}
	g, err := makeGrid()
	if err != nil {
		return model{}, err
	}

	h.Focus()
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
	}, nil
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
	case 0:
		m.horizontal.Focus()
	case 1:
		m.vertical.Focus()
	case 2:
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

func makeHorizontal() (buttongroup.Model, error) {
	yes := button.New("Yes")
	disabled := button.New("Disabled")
	disabled.SetDisabled(true)
	maybe := button.New("Maybe")
	return buttongroup.HorizontalGroup(yes, disabled, maybe)
}

func makeVertical() (buttongroup.Model, error) {
	one := button.New("One")
	one.SetStyle(button.Parens)
	two := button.New("Two")
	two.SetStyle(button.Parens)
	three := button.New("Three")
	three.SetStyle(button.Parens)
	four := button.New("Four")
	four.SetStyle(button.Parens)
	four.SetDisabled(true)

	return buttongroup.VerticalGroup(one, two, three, four)
}

func makeGrid() (buttongroup.Model, error) {
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
	btnE.SetDisabled(true)
	btnF := button.New("F")
	btnF.SetStyle(button.BoxDrawing)
	btnG := button.New("G")
	btnG.SetStyle(button.BoxDrawing)

	return buttongroup.ButtonGrid(3, 3, btnA, btnB, btnC, btnD, btnE, btnF, btnG)
}

func main() {
	m, err := newModel()
	if err != nil {
		fmt.Printf("Could not create model :(\n%v\n", err)
		os.Exit(1)
	}

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Printf("Could not start program :(\n%v\n", err)
		os.Exit(1)
	}
}
