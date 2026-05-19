package button

import (
	"fmt"
	"strings"
	"sync/atomic"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

//                    .___     .__
//   _____   ____   __| _/____ |  |
//  /     \ /  _ \ / __ |/ __ \|  |
// |  Y Y  (  <_> ) /_/ \  ___/|  |__
// |__|_|  /\____/\____ |\___  >____/
//       \/            \/    \/

// ButtonType represents the visual style of a button.
type ButtonType int

const (
	// Brackets style: [ text ]
	Brackets ButtonType = iota
	// Parens style: ( text )
	Parens
	// BoxDrawing style:
	// ┌──────┐
	// │ text │
	// └──────┘
	BoxDrawing
)

// ButtonState represents the button's current state.
type ButtonState int

const (
	Enabled ButtonState = iota
	Disabled
	Focused
)

// KeyMap defines the keybindings for the button.
type KeyMap struct {
	ButtonPress key.Binding
}

// DefaultKeyMap returns a sensible default key map.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		ButtonPress: key.NewBinding(
			key.WithKeys("enter", "space"),
			key.WithHelp("space/enter", "press button"),
		),
	}
}

// Styles contains the styling for different button states.
type Styles struct {
	Enabled  lipgloss.Style
	Disabled lipgloss.Style
	Focused  lipgloss.Style
}

// DefaultStyles returns sensible default styles.
func DefaultStyles() Styles {
	return Styles{
		Enabled:  lipgloss.NewStyle().Foreground(lipgloss.Color("255")),
		Disabled: lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		Focused:  lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true),
	}
}

// Model represents a button.
type Model struct {
	// private fields
	id    int
	label string
	style ButtonType
	state ButtonState

	// The keybindings used by the button.
	KeyMap KeyMap
	// Styles contains the styling for different button states.
	Styles Styles
}

// New creates a new button with the given label.
func New(label string) Model {
	keys := DefaultKeyMap()
	styles := DefaultStyles()

	return Model{
		id:     nextID(),
		label:  label,
		style:  Brackets,
		state:  Enabled,
		KeyMap: keys,
		Styles: styles,
	}
}

// Init initializes the button (no-op).
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and returns any commands.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}

	return m.handlePress(keyMsg)
}

// View renders the button.
func (m Model) View() string {
	return m.styleForState().Render(m.renderedLabel())
}

// .__                       .___.__
// |  |__ _____    ____    __| _/|  |   ___________  ______
// |  |  \\__  \  /    \  / __ | |  | _/ __ \_  __ \/  ___/
// |   Y  \/ __ \|   |  \/ /_/ | |  |_\  ___/|  | \/\___ \
// |___|  (____  /___|  /\____ | |____/\___  >__|  /____  >
//      \/     \/     \/      \/           \/           \/

func (m Model) handlePress(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	if key.Matches(msg, m.KeyMap.ButtonPress) {
		if m.state == Disabled {
			return m, func() tea.Msg { return disabledPressMsg{ID: m.id} }
		}
		if m.state == Focused {
			return m, func() tea.Msg { return pressMsg{ID: m.id} }
		}
	}

	return m, nil
}

//                            .___           .__
// _______   ____   ____    __| _/___________|__| ____    ____
// \_  __ \_/ __ \ /    \  / __ |/ __ \_  __ \  |/    \  / ___\
//  |  | \/\  ___/|   |  \/ /_/ \  ___/|  | \/  |   |  \/ /_/  >
//  |__|    \___  >___|  /\____ |\___  >__|  |__|___|  /\___  /
//              \/     \/      \/    \/              \//_____/

func (m Model) styleForState() lipgloss.Style {
	switch m.state {
	case Disabled:
		return m.Styles.Disabled
	case Focused:
		return m.Styles.Focused
	default:
		return m.Styles.Enabled
	}
}

func (m Model) renderedLabel() string {
	switch m.style {
	case Brackets:
		return fmt.Sprintf("[ %s ]", m.label)
	case Parens:
		return fmt.Sprintf("( %s )", m.label)
	case BoxDrawing:
		return boxDrawing(m.label)
	default:
		return fmt.Sprintf("[ %s ]", m.label)
	}
}

func boxDrawing(label string) string {
	inner := " " + label + " "
	border := strings.Repeat("─", len(inner))
	return "┌" + border + "┐\n│" + inner + "│\n└" + border + "┘"
}

//             ___.   .__  .__                       .__
// ______  __ _\_ |__ |  | |__| ____   _____  ______ |__|
// \____ \|  |  \ __ \|  | |  |/ ___\  \__  \ \____ \|  |
// |  |_> >  |  / \_\ \  |_|  \  \___   / __ \|  |_> >  |
// |   __/|____/|___  /____/__|\___  > (____  /   __/|__|
// |__|             \/             \/       \/|__|

// ID returns the button's unique identifier.
func (m Model) ID() int {
	return m.id
}

// Label returns the button's label text.
func (m Model) Label() string {
	return m.label
}

// SetLabel sets the button's label text.
func (m *Model) SetLabel(label string) {
	m.label = label
}

// SetStyle sets the button's visual style.
func (m *Model) SetStyle(style ButtonType) {
	m.style = style
}

// Style returns the button's current visual style.
func (m Model) Style() ButtonType {
	return m.style
}

// Focus marks the button as focused.
func (m *Model) Focus() {
	if m.state != Disabled {
		m.state = Focused
	}
}

// Blur marks the button as unfocused.
func (m *Model) Blur() {
	if m.state != Disabled {
		m.state = Enabled
	}
}

// State returns the button's state.
func (m Model) State() ButtonState {
	return m.state
}

// SetDisabled enables or disables the button.
func (m *Model) SetDisabled(disable bool) {
	if disable {
		m.state = Disabled
	} else {
		m.state = Enabled
		m.Focus()
	}
}

// DidPress returns wheter a user has pressed the button (on this msg).
func (m Model) DidPress(msg tea.Msg) bool {
	press, ok := msg.(pressMsg)
	return ok && press.ID == m.id
}

// DidDisabledPress returns whether a user attempted to press a disabled button (on this msg).
func (m Model) DidDisabledPress(msg tea.Msg) bool {
	press, ok := msg.(disabledPressMsg)
	return ok && press.ID == m.id
}

//              .__               __
// _____________|__|__  _______ _/  |_  ____
// \____ \_  __ \  \  \/ /\__  \\   __\/ __ \
// |  |_> >  | \/  |\   /  / __ \|  | \  ___/
// |   __/|__|  |__| \_/  (____  /__|  \___  >
// |__|                        \/          \/

var lastID int64

func nextID() int {
	return int(atomic.AddInt64(&lastID, 1))
}

type pressMsg struct {
	ID int
}

type disabledPressMsg struct {
	ID int
}
