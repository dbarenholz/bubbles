// A filefinder is a bubble that allows you to find files.
// Consider it a filepicker, but with a search input that filters the shown entries.
package filefinder

import (
	"os"
	"path/filepath"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"

	tfp "github.com/dbarenholz/bubbles/traversablefilepicker"

	tea "charm.land/bubbletea/v2"
)

type KeyMap struct {
	// FocusSearch is the key binding to focus the search input
	FocusSearch key.Binding
	// FocusFiles is the key binding to focus back on the file list
	FocusFiles key.Binding
	// DiscardSearch is the key binding to discard the search query and go back to the full file list
	DiscardSearch key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		// Open the search by pressing "/"
		FocusSearch:   key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "open search")),
		FocusFiles:    key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "focus files")),
		DiscardSearch: key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "discard search")),
	}
}

// Enum to keep track of which part of the filefinder has focus
type Focus int

const (
	// If we're focused on the files
	FocusOnFiles = iota

	// If we're focused on the search input
	FocusOnSearch
)

type Model struct {
	TextInput  textinput.Model
	FilePicker tfp.Model
	Focus      Focus
	KeyMap     KeyMap
}

func New(at string) Model {
	// grab the default keys
	keys := DefaultKeyMap()

	// create the textinput with the correct keys (charm's textinput default keymap does not have Help on these keys, so we set the ones we want here)
	textInput := textinput.New()
	textInput.Placeholder = "press / to search"
	textInput.SetWidth(len(textInput.Placeholder) + 3)
	textInput.KeyMap = textinput.KeyMap{
		CharacterForward:        key.NewBinding(key.WithKeys("right", "ctrl+f"), key.WithHelp("→", "right")),
		CharacterBackward:       key.NewBinding(key.WithKeys("left", "ctrl+b"), key.WithHelp("←", "left")),
		WordForward:             key.NewBinding(key.WithKeys("alt+right", "ctrl+right", "alt+f"), key.WithHelp("ctrl+→", "word right")),
		WordBackward:            key.NewBinding(key.WithKeys("alt+left", "ctrl+left", "alt+b"), key.WithHelp("ctrl+←", "word left")),
		DeleteWordBackward:      key.NewBinding(key.WithKeys("alt+backspace", "ctrl+w"), key.WithHelp("alt+⌫", "backspace word")),
		DeleteWordForward:       key.NewBinding(key.WithKeys("alt+delete", "alt+d"), key.WithHelp("alt+⌦", "delete word")),
		DeleteBeforeCursor:      key.NewBinding(key.WithKeys("ctrl+u"), key.WithHelp("ctrl+u", "backspace to start")),
		DeleteAfterCursor:       key.NewBinding(key.WithKeys("ctrl+k"), key.WithHelp("ctrl+k", "delete to end")),
		DeleteCharacterBackward: key.NewBinding(key.WithKeys("backspace", "ctrl+h"), key.WithHelp("⌫", "backspace")),
		DeleteCharacterForward:  key.NewBinding(key.WithKeys("delete", "ctrl+d"), key.WithHelp("⌦", "delete")),
		LineStart:               key.NewBinding(key.WithKeys("home", "ctrl+a"), key.WithHelp("home", "to start")),
		LineEnd:                 key.NewBinding(key.WithKeys("end", "ctrl+e"), key.WithHelp("end", "to end")),
		Paste:                   key.NewBinding(key.WithKeys("ctrl+v"), key.WithHelp("ctrl+v", "paste")),
	}
	textInput.Prompt = " "

	// get absolute path to set as current directory of the filepicker
	abs, err := filepath.Abs(at)
	if err != nil {
		abs = "."
	}

	// create the filepicker with the correct keys and current directory
	filePicker := tfp.New(abs)
	filePicker.KeyMap = tfp.DefaultKeyMap()

	// finally, init the model, with initial focus on files
	return Model{
		TextInput:  textInput,
		FilePicker: filePicker,
		Focus:      FocusOnFiles,
		KeyMap:     keys,
	}
}

func (m *Model) Init() tea.Cmd {
	return m.FilePicker.Init()
}

func filterFor(query string) func(entry os.DirEntry) bool {
	return func(entry os.DirEntry) bool {
		// if the query is empty, show all entries
		if query == "" {
			return true
		}
		// otherwise, only show entries that contain the query (case-insensitive)
		return strings.Contains(strings.ToLower(entry.Name()), query)
	}
}

// All the things we need to do when files is focused (initial state)
func (m Model) updateForFiles(msg tea.Msg) (Model, tea.Cmd) {
	// if the message is _not_ a key press, delegate it to the filepicker
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		fpModel, cmd := m.FilePicker.Update(msg)
		m.FilePicker = fpModel
		return m, cmd
	}

	// if it is to focus search, then do that!
	if key.Matches(keyMsg, m.KeyMap.FocusSearch) {
		m.Focus = FocusOnSearch
		return m, m.TextInput.Focus()
	}

	// when discarding search, or on path traversals: reset the filter
	if key.Matches(keyMsg, m.FilePicker.KeyMap.Open) || key.Matches(keyMsg, m.FilePicker.KeyMap.Back) {
		m.TextInput.SetValue("")
		m.FilePicker.SetFilterFunc(filterFor(""))
		fpModel, cmd := m.FilePicker.Update(msg)
		m.FilePicker = fpModel
		return m, cmd
	}

	// otherwise, pass on through to filepicker
	fpModel, cmd := m.FilePicker.Update(msg)
	m.FilePicker = fpModel
	return m, cmd

}

// All the things we need to do when search is focused
func (m Model) updateForSearch(msg tea.Msg) (Model, tea.Cmd) {
	// if the message is _not_ a key press, delegate it to the filepicker
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		fpModel, cmd := m.FilePicker.Update(msg)
		m.FilePicker = fpModel
		return m, cmd
	}

	// if it is a keypress, and specifically the one for FocusFiles, then do that!
	if key.Matches(keyMsg, m.KeyMap.FocusFiles) {

		// only one entry, select it
		if m.FilePicker.NumEntries() == 1 {
			fpModel, cmd := m.FilePicker.Update(msg)
			m.FilePicker = fpModel
			return m, cmd
		}

		// multiple entries, so just switch focus back to files
		m.Focus = FocusOnFiles
		m.TextInput.Blur()
		return m, nil
	}

	if key.Matches(keyMsg, m.KeyMap.DiscardSearch) {
		m.Focus = FocusOnFiles
		m.TextInput.Blur()
		m.TextInput.SetValue("")
		m.FilePicker.SetFilterFunc(filterFor(""))
		return m, m.FilePicker.RefreshEntries()
	}

	// in all other cases, send to text input
	tiModel, cmd := m.TextInput.Update(msg)
	m.TextInput = tiModel

	// get search query
	searchQuery := strings.ToLower(m.TextInput.Value())

	// then set filepicker's filter
	m.FilePicker.SetFilterFunc(filterFor(searchQuery))
	fpCmd := m.FilePicker.RefreshEntries()
	return m, tea.Batch(cmd, fpCmd)
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if m.Focus == FocusOnSearch {
		return m.updateForSearch(msg)
	}
	if m.Focus == FocusOnFiles {
		return m.updateForFiles(msg)
	}

	return m, nil
}

// View renders the file picker and search input with styling
func (m *Model) View() string {
	searchBoxStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Margin(0, 1).Width(59)

	// Style for filepicker with left margin to align with search box
	filePickerStyle := lipgloss.NewStyle().MarginLeft(2)

	// Get the number of entries to check for warnings
	numEntries := m.FilePicker.NumEntries()
	focusedColor := m.FilePicker.Styles.Selected.GetForeground()
	dirColor := m.FilePicker.Styles.Directory.GetForeground()
	dimmedColor := m.FilePicker.Styles.DisabledFile.GetForeground()

	// Apply conditional styling to the search input
	searchStyle := searchBoxStyle
	if m.Focus == FocusOnSearch {
		searchStyle = searchStyle.BorderForeground(focusedColor)
	} else {
		searchStyle = searchStyle.BorderForeground(dimmedColor).Foreground(dimmedColor)
	}

	// If no entries, use warning colors
	if numEntries == 0 {
		searchStyle = searchStyle.BorderForeground(lipgloss.Color("196"))
	}

	searchInput := searchStyle.Render(m.TextInput.View())

	pathStyle := lipgloss.NewStyle().MarginLeft(2).Foreground(dirColor)
	currentPath := pathStyle.Render(m.FilePicker.CurrentDirectory())
	filePickerView := filePickerStyle.Render(m.FilePicker.View())

	return searchInput + "\n" + currentPath + "\n" + filePickerView
}
