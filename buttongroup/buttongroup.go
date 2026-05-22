// A button group bubble for bubbletea TUI applications.
// Button groups are considered immutable once created: buttons cannot be added/removed
package buttongroup

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/dbarenholz/bubbles/button"
)

//                    .___     .__
//   _____   ____   __| _/____ |  |
//  /     \ /  _ \ / __ |/ __ \|  |
// |  Y Y  (  <_> ) /_/ \  ___/|  |__
// |__|_|  /\____/\____ |\___  >____/
//       \/            \/    \/

// GroupLayout represents the layout of buttons in a group.
type GroupLayout int

const (
	Horizontal GroupLayout = iota
	Vertical
	Grid
)

// KeyMap defines the keybindings for the button group.
type KeyMap struct {
	FocusNext key.Binding
	FocusPrev key.Binding
	Left      key.Binding
	Right     key.Binding
	Up        key.Binding
	Down      key.Binding
}

// DefaultKeyMap returns sensible default key bindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		FocusNext: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next"),
		),
		FocusPrev: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "prev"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "right"),
		),
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
	}
}

// Model represents a group of buttons.
type Model struct {
	buttons     []button.Model
	layout      GroupLayout
	focusIdx    int
	gridRows    int // > 0
	gridColumns int // > 0
	spacing     int
	focused     bool
	KeyMap      KeyMap
}

func (m *Model) buttonIndexByID(id int) (int, bool) {
	for idx := range m.buttons {
		if m.buttons[idx].ID() == id {
			return idx, true
		}
	}
	return -1, false
}

func (m *Model) setFocusedIndex(idx int) {
	if idx < 0 || idx >= len(m.buttons) {
		return
	}
	if m.focusIdx >= 0 && m.focusIdx < len(m.buttons) && m.focusIdx != idx {
		m.buttons[m.focusIdx].Blur()
	}
	m.focusIdx = idx
	if m.focused {
		m.buttons[m.focusIdx].Focus()
	}
}

func (m *Model) blurFocusedButton() {
	if m.focusIdx >= 0 && m.focusIdx < len(m.buttons) {
		m.buttons[m.focusIdx].Blur()
	}
}

// Creates a horizontal button group with the given buttons.
func HorizontalGroup(buttons ...button.Model) Model {
	if len(buttons) == 0 {
		return Model{}
	}
	// copy so callers can't screw up our buttons
	btns := make([]button.Model, len(buttons))
	copy(btns, buttons)
	m := Model{
		buttons:  btns,
		layout:   Horizontal,
		focusIdx: 0,
		spacing:  1,
		focused:  true,
		KeyMap:   DefaultKeyMap(),
	}
	m.buttons[0].Focus()
	return m
}

// Creates a vertical button group with the given buttons.
func VerticalGroup(buttons ...button.Model) Model {
	if len(buttons) == 0 {
		return Model{}
	}
	// copy so callers can't screw up our buttons
	btns := make([]button.Model, len(buttons))
	copy(btns, buttons)
	m := Model{
		buttons:  btns,
		layout:   Vertical,
		focusIdx: 0,
		spacing:  1,
		focused:  true,
		KeyMap:   DefaultKeyMap(),
	}
	m.buttons[0].Focus()
	return m
}

// Creates a grid button group with the given buttons and column count.
// Fills buttons left to right, top to bottom.
// Returns an empty model if the inputs are invalid.
func ButtonGrid(rows, cols int, buttons ...button.Model) Model {
	if len(buttons) == 0 {
		return Model{}
	}
	if rows <= 0 || cols <= 0 {
		return Model{}
	}
	if len(buttons) > rows*cols {
		return Model{}
	}
	// copy so callers can't screw up our buttons
	btns := make([]button.Model, len(buttons))
	copy(btns, buttons)
	m := Model{
		buttons:     btns,
		layout:      Grid,
		focusIdx:    0,
		gridRows:    rows,
		gridColumns: cols,
		spacing:     1,
		focused:     true,
		KeyMap:      DefaultKeyMap(),
	}
	m.buttons[0].Focus()
	return m
}

// Init initializes the button group (no-op).
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and returns any commands.
//
// Navigation behaviour:
//   - tab / shift+tab: move linearly through all buttons with wrap-around (all layouts)
//   - Horizontal: Left/Right move without wrapping; Up/Down do nothing
//   - Vertical: Up/Down move without wrapping; Left/Right do nothing
//   - Grid: arrow keys move by row/column without wrapping at edges
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.focused {
		return m, nil
	}
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		switch {
		case key.Matches(keyMsg, m.KeyMap.FocusNext):
			m.focusNext()
			return m, nil
		case key.Matches(keyMsg, m.KeyMap.FocusPrev):
			m.focusPrev()
			return m, nil
		case key.Matches(keyMsg, m.KeyMap.Right):
			m.handleRight()
			return m, nil
		case key.Matches(keyMsg, m.KeyMap.Left):
			m.handleLeft()
			return m, nil
		case key.Matches(keyMsg, m.KeyMap.Down):
			m.handleDown()
			return m, nil
		case key.Matches(keyMsg, m.KeyMap.Up):
			m.handleUp()
			return m, nil
		}
	}
	return m, m.updateFocused(msg)
}

// View renders the button group.
func (m Model) View() string {
	if len(m.buttons) == 0 {
		return ""
	}
	rowSpacing := max(m.spacing-1, 0)

	var rendered string
	switch m.layout {
	case Horizontal:
		rendered = m.renderHorizontal()
	case Vertical:
		rendered = m.renderVertical(rowSpacing)
	case Grid:
		rendered = m.renderGrid(rowSpacing)
	default:
		rendered = m.renderHorizontal()
	}

	return rendered
}

// .__                       .___.__
// |  |__ _____    ____    __| _/|  |   ___________  ______
// |  |  \\__  \  /    \  / __ | |  | _/ __ \_  __ \/  ___/
// |   Y  \/ __ \|   |  \/ /_/ | |  |_\  ___/|  | \/\___ \
// |___|  (____  /___|  /\____ | |____/\___  >__|  /____  >
//      \/     \/     \/      \/           \/           \/

// isDisabled returns true if the button at idx is disabled.
func (m *Model) isDisabled(idx int) bool {
	return m.buttons[idx].State() == button.Disabled
}

// focusNext moves focus to the next non-disabled button, wrapping around.
func (m *Model) focusNext() {
	if len(m.buttons) == 0 {
		return
	}

	start := m.focusIdx
	next := (m.focusIdx + 1) % len(m.buttons)
	for next != start && m.isDisabled(next) {
		next = (next + 1) % len(m.buttons)
	}
	if next == start {
		return // all other buttons are disabled
	}
	m.buttons[m.focusIdx].Blur()
	m.focusIdx = next
	m.buttons[m.focusIdx].Focus()
}

// focusPrev moves focus to the previous non-disabled button, wrapping around.
func (m *Model) focusPrev() {
	if len(m.buttons) == 0 {
		return
	}

	start := m.focusIdx
	prev := m.focusIdx - 1
	if prev < 0 {
		prev = len(m.buttons) - 1
	}
	for prev != start && m.isDisabled(prev) {
		prev--
		if prev < 0 {
			prev = len(m.buttons) - 1
		}
	}
	if prev == start {
		return // all other buttons are disabled
	}
	m.buttons[m.focusIdx].Blur()
	m.focusIdx = prev
	m.buttons[m.focusIdx].Focus()
}

func (m *Model) handleRight() {
	switch m.layout {
	case Horizontal:
		next := m.focusIdx + 1
		for next < len(m.buttons) && m.isDisabled(next) {
			next++
		}
		if next < len(m.buttons) {
			m.buttons[m.focusIdx].Blur()
			m.focusIdx = next
			m.buttons[m.focusIdx].Focus()
		}
	case Grid:
		next := m.focusIdx + 1
		for next < len(m.buttons) && next%m.gridColumns != 0 && m.isDisabled(next) {
			next++
		}
		if next < len(m.buttons) && next%m.gridColumns != 0 && !m.isDisabled(next) {
			m.buttons[m.focusIdx].Blur()
			m.focusIdx = next
			m.buttons[m.focusIdx].Focus()
		}
	}
}

func (m *Model) handleLeft() {
	switch m.layout {
	case Horizontal:
		prev := m.focusIdx - 1
		for prev >= 0 && m.isDisabled(prev) {
			prev--
		}
		if prev >= 0 {
			m.buttons[m.focusIdx].Blur()
			m.focusIdx = prev
			m.buttons[m.focusIdx].Focus()
		}
	case Grid:
		rowStart := (m.focusIdx / m.gridColumns) * m.gridColumns
		prev := m.focusIdx - 1
		for prev >= rowStart && m.isDisabled(prev) {
			prev--
		}
		if prev >= rowStart && !m.isDisabled(prev) {
			m.buttons[m.focusIdx].Blur()
			m.focusIdx = prev
			m.buttons[m.focusIdx].Focus()
		}
	}
}

func (m *Model) handleDown() {
	switch m.layout {
	case Vertical:
		next := m.focusIdx + 1
		for next < len(m.buttons) && m.isDisabled(next) {
			next++
		}
		if next < len(m.buttons) {
			m.buttons[m.focusIdx].Blur()
			m.focusIdx = next
			m.buttons[m.focusIdx].Focus()
		}
	case Grid:
		next := m.focusIdx + m.gridColumns
		for next < len(m.buttons) && m.isDisabled(next) {
			next += m.gridColumns
		}
		if next < len(m.buttons) {
			m.buttons[m.focusIdx].Blur()
			m.focusIdx = next
			m.buttons[m.focusIdx].Focus()
		}
	}
}

func (m *Model) handleUp() {
	switch m.layout {
	case Vertical:
		prev := m.focusIdx - 1
		for prev >= 0 && m.isDisabled(prev) {
			prev--
		}
		if prev >= 0 {
			m.buttons[m.focusIdx].Blur()
			m.focusIdx = prev
			m.buttons[m.focusIdx].Focus()
		}
	case Grid:
		prev := m.focusIdx - m.gridColumns
		for prev >= 0 && m.isDisabled(prev) {
			prev -= m.gridColumns
		}
		if prev >= 0 {
			m.buttons[m.focusIdx].Blur()
			m.focusIdx = prev
			m.buttons[m.focusIdx].Focus()
		}
	}
}

//                            .___           .__
// _______   ____   ____    __| _/___________|__| ____    ____
// \_  __ \_/ __ \ /    \  / __ |/ __ \_  __ \  |/    \  / ___\
//  |  | \/\  ___/|   |  \/ /_/ \  ___/|  | \/  |   |  \/ /_/  >
//  |__|    \___  >___|  /\____ |\___  >__|  |__|___|  /\___  /
//              \/     \/      \/    \/              \//_____/

// renderRow renders a slice of buttons side by side, correctly handling
// multi-line button views (e.g. BoxDrawing style) by aligning each line.
func renderRow(buttons []button.Model, spacing int) string {
	if len(buttons) == 0 {
		return ""
	}
	sep := strings.Repeat(" ", spacing)
	buttonLines := make([][]string, len(buttons))
	maxLines := 0
	for i, btn := range buttons {
		lines := strings.Split(btn.View(), "\n")
		buttonLines[i] = lines
		if len(lines) > maxLines {
			maxLines = len(lines)
		}
	}
	rows := make([]string, maxLines)
	for lineIdx := range maxLines {
		parts := make([]string, len(buttons))
		for btnIdx, lines := range buttonLines {
			if lineIdx < len(lines) {
				parts[btnIdx] = lines[lineIdx]
			} else {
				// pad with spaces matching the width of the first line
				if len(lines) > 0 {
					parts[btnIdx] = strings.Repeat(" ", len([]rune(lines[0])))
				}
			}
		}
		rows[lineIdx] = strings.Join(parts, sep)
	}
	return strings.Join(rows, "\n")
}

func (m Model) renderHorizontal() string {
	return renderRow(m.buttons, m.spacing)
}

func (m Model) renderVertical(rowSpacing int) string {
	var rendered strings.Builder
	for i, item := range m.buttons {
		rendered.WriteString(item.View())
		if i < len(m.buttons)-1 {
			rendered.WriteString("\n")
			rendered.WriteString(strings.Repeat("\n", rowSpacing))
		}
	}
	return rendered.String()
}

func (m Model) renderGridItems(count int, rowSpacing int) string {
	rowSep := "\n" + strings.Repeat("\n", rowSpacing)
	var rows []string
	for rowStart := 0; rowStart < count; rowStart += m.gridColumns {
		rowEnd := min(rowStart+m.gridColumns, count)
		rows = append(rows, renderRow(m.buttons[rowStart:rowEnd], m.spacing))
	}
	return strings.Join(rows, rowSep)
}

func (m Model) renderGrid(rowSpacing int) string {
	count := len(m.buttons)
	if m.gridRows > 0 {
		max := m.gridRows * m.gridColumns
		if max < count {
			count = max
		}
	}
	return m.renderGridItems(count, rowSpacing)
}

//             ___.   .__  .__                       .__
// ______  __ _\_ |__ |  | |__| ____   _____  ______ |__|
// \____ \|  |  \ __ \|  | |  |/ ___\  \__  \ \____ \|  |
// |  |_> >  |  / \_\ \  |_|  \  \___   / __ \|  |_> >  |
// |   __/|____/|___  /____/__|\___  > (____  /   __/|__|
// |__|             \/             \/       \/|__|

// Focus restores the group focus, or focuses a specific button when an id is provided.
func (m *Model) Focus(ids ...int) {
	if len(m.buttons) == 0 {
		m.focused = true
		return
	}
	if len(ids) == 0 {
		m.focused = true
		m.buttons[m.focusIdx].Focus()
		return
	}
	if idx, ok := m.buttonIndexByID(ids[0]); ok {
		m.focused = true
		m.setFocusedIndex(idx)
	}
}

// Blur removes group focus, or blurs a specific button when an id is provided.
func (m *Model) Blur(ids ...int) {
	if len(m.buttons) == 0 {
		m.focused = false
		return
	}
	if len(ids) == 0 {
		m.blurFocusedButton()
		m.focused = false
		return
	}
	if idx, ok := m.buttonIndexByID(ids[0]); ok {
		m.buttons[idx].Blur()
		if idx == m.focusIdx && m.focused {
			m.focusIdx = idx
		}
	}
}

// SetDisabled disables or enables a button in the group by id.
func (m *Model) SetDisabled(id int, disable bool) {
	if idx, ok := m.buttonIndexByID(id); ok {
		m.buttons[idx].SetDisabled(disable)
	}
}

// FocusedButton returns the currently focused button.
func (m Model) FocusedButton() *button.Model {
	if m.focusIdx < 0 || m.focusIdx >= len(m.buttons) {
		return nil
	}
	return &m.buttons[m.focusIdx]
}

// NumButtons returns the number of buttons in the group.
func (m Model) NumButtons() int {
	return len(m.buttons)
}

// SetSpacing sets the spacing between items (in characters).
func (m *Model) SetSpacing(spacing int) {
	if spacing < 0 {
		spacing = 0
	}
	m.spacing = spacing
}

// DidPress returns whether a user has pressed a button in the group (on this msg).
// It also returns the pressed button ID and index.
func (m Model) DidPress(msg tea.Msg) (bool, int, int) {
	press, ok := msg.(pressMsg)
	if !ok {
		return false, 0, -1
	}

	if press.GroupIndex < 0 || press.GroupIndex >= len(m.buttons) {
		return false, 0, -1
	}

	if m.buttons[press.GroupIndex].ID() != press.ID {
		return false, 0, -1
	}

	return true, press.ID, press.GroupIndex
}

// DidDisabledPress returns whether a user attempted to press a disabled button
// in the group (on this msg). It also returns the button ID and index.
func (m Model) DidDisabledPress(msg tea.Msg) (bool, int, int) {
	press, ok := msg.(disabledPressMsg)
	if !ok {
		return false, 0, -1
	}

	if press.GroupIndex < 0 || press.GroupIndex >= len(m.buttons) {
		return false, 0, -1
	}

	if m.buttons[press.GroupIndex].ID() != press.ID {
		return false, 0, -1
	}

	return true, press.ID, press.GroupIndex
}

//              .__               __
// _____________|__|__  _______ _/  |_  ____
// \____ \_  __ \  \  \/ /\__  \\   __\/ __ \
// |  |_> >  | \/  |\   /  / __ \|  | \  ___/
// |   __/|__|  |__| \_/  (____  /__|  \___  >
// |__|                        \/          \/

type pressMsg struct {
	ID         int
	GroupIndex int
}

type disabledPressMsg struct {
	ID         int
	GroupIndex int
}

func (m *Model) updateFocused(msg tea.Msg) tea.Cmd {
	if m.focusIdx >= 0 && m.focusIdx < len(m.buttons) {
		focusedIndex := m.focusIdx
		b, cmd := m.buttons[focusedIndex].Update(msg)
		m.buttons[focusedIndex] = b
		if !m.focused {
			return nil
		}
		focusedButton := m.buttons[focusedIndex]
		focusedButtonID := focusedButton.ID()
		if cmd == nil {
			return nil
		}
		return func() tea.Msg {
			result := cmd()
			if focusedButton.DidPress(result) {
				return pressMsg{ID: focusedButtonID, GroupIndex: focusedIndex}
			}
			if focusedButton.DidDisabledPress(result) {
				return disabledPressMsg{ID: focusedButtonID, GroupIndex: focusedIndex}
			}
			return result
		}
	}
	return nil
}
