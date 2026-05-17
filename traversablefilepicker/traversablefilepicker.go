package traversablefilepicker

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/dustin/go-humanize"
)

var lastID int64

func nextID() int {
	return int(atomic.AddInt64(&lastID, 1))
}

// New returns a new filepicker model with default styling and key bindings.
func New(at string) Model {
	absolutePath, err := filepath.Abs(at)
	if err != nil {
		panic(fmt.Sprintf("could not get absolute path of %s: %v", at, err))
	}
	return Model{
		id:               nextID(),
		currentDirectory: absolutePath,
		Cursor:           ">",
		AllowedTypes:     []string{},
		ShowPermissions:  true,
		ShowSize:         true,
		ShowHidden:       false, // while I would like true, I will keep default charm behaviour for now
		DirAllowed:       false,
		FileAllowed:      true,
		AutoHeight:       true,
		LoopEntries:      false,                                     // while I would like true, I will keep default charm behaviour for now
		entryFilter:      func(de os.DirEntry) bool { return true }, // by default, keep all files
		height:           0,
		history:          newHistory(absolutePath), // uses current directory as parameter
		KeyMap:           DefaultKeyMap(),
		Styles:           DefaultStyles(),
	}
}

type errorMsg struct {
	err error
}

type readDirMsg struct {
	id      int
	entries []os.DirEntry
}

const (
	marginBottom  = 5
	fileSizeWidth = 7
	paddingLeft   = 2
)

// KeyMap defines key bindings for each user action.
type KeyMap struct {
	GoToTop  key.Binding
	GoToLast key.Binding
	Down     key.Binding
	Up       key.Binding
	PageUp   key.Binding
	PageDown key.Binding
	Back     key.Binding
	Open     key.Binding
	Select   key.Binding

	// NOTE: added a convenience keybind to go home
	GoHome key.Binding
}

// DefaultKeyMap defines the default keybindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		GoToTop:  key.NewBinding(key.WithKeys("g", "home"), key.WithHelp("g", "first")),
		GoToLast: key.NewBinding(key.WithKeys("G", "end"), key.WithHelp("G", "last")),
		Down:     key.NewBinding(key.WithKeys("j", "down", "ctrl+n"), key.WithHelp("j", "down")),
		Up:       key.NewBinding(key.WithKeys("k", "up", "ctrl+p"), key.WithHelp("k", "up")),
		PageUp:   key.NewBinding(key.WithKeys("K", "pgup"), key.WithHelp("pgup", "page up")),
		PageDown: key.NewBinding(key.WithKeys("J", "pgdown"), key.WithHelp("pgdown", "page down")),
		Back:     key.NewBinding(key.WithKeys("h", "backspace", "left", "esc"), key.WithHelp("h", "back")),
		Open:     key.NewBinding(key.WithKeys("l", "right", "enter"), key.WithHelp("l", "open")),
		Select:   key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),

		GoHome: key.NewBinding(key.WithKeys("~"), key.WithHelp("~", "go home")),
	}
}

// Styles defines the possible customizations for styles in the file picker.
type Styles struct {
	DisabledCursor   lipgloss.Style
	Cursor           lipgloss.Style
	Symlink          lipgloss.Style
	Directory        lipgloss.Style
	File             lipgloss.Style
	DisabledFile     lipgloss.Style
	Permission       lipgloss.Style
	Selected         lipgloss.Style
	DisabledSelected lipgloss.Style
	FileSize         lipgloss.Style
	EmptyDirectory   lipgloss.Style
}

// DefaultStyles defines the default styling for the file picker.
func DefaultStyles() Styles {
	return Styles{
		DisabledCursor:   lipgloss.NewStyle().Foreground(lipgloss.Color("247")),
		Cursor:           lipgloss.NewStyle().Foreground(lipgloss.Color("212")),
		Symlink:          lipgloss.NewStyle().Foreground(lipgloss.Color("36")),
		Directory:        lipgloss.NewStyle().Foreground(lipgloss.Color("99")),
		File:             lipgloss.NewStyle(),
		DisabledFile:     lipgloss.NewStyle().Foreground(lipgloss.Color("243")),
		DisabledSelected: lipgloss.NewStyle().Foreground(lipgloss.Color("247")),
		Permission:       lipgloss.NewStyle().Foreground(lipgloss.Color("244")),
		Selected:         lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true),
		FileSize:         lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Width(fileSizeWidth).Align(lipgloss.Right),
		EmptyDirectory:   lipgloss.NewStyle().Foreground(lipgloss.Color("240")).PaddingLeft(paddingLeft).SetString("Bummer. No Files Found."),
	}
}

type historyEntry struct {
	selected int
	minIdx   int
	maxIdx   int
}

func newHistory(cwd string) map[string]*historyEntry {
	history := make(map[string]*historyEntry)
	history[cwd] = &historyEntry{selected: 0, minIdx: 0, maxIdx: 0}
	return history
}

func (m *Model) SetFilterFunc(fn func(os.DirEntry) bool) {
	m.entryFilter = fn
}

func (m *Model) RefreshEntries() tea.Cmd {
	return m.readDir(m.currentDirectory, m.ShowHidden)
}

func (m *Model) NumEntries() int {
	return len(m.entries)
}

// CurrentDirectory returns the absolute path of the directory the filepicker is currently showing.
func (m *Model) CurrentDirectory() string {
	return m.currentDirectory
}

// Model represents a file picker.
type Model struct {
	id int

	// Path is the path which the user has selected with the file picker.
	Path string

	// currentDirectory is the directory that the user is currently in.
	currentDirectory string

	// AllowedTypes specifies which file types the user may select.
	// If empty the user may select any file.
	AllowedTypes []string

	KeyMap          KeyMap
	entries         []os.DirEntry
	entryFilter     func(os.DirEntry) bool // a function for filtering; true == keep
	ShowPermissions bool                   // whether to show file permissions; this does not affect whether files can be selected based on type
	ShowSize        bool                   // whether to show file sizes; this does not affect whether files can be selected based on type
	ShowHidden      bool                   // whether hidden files should be shown; this does not affect whether hidden files can be selected
	DirAllowed      bool                   // whether directories can be selected; if false, only files can be selected
	FileAllowed     bool                   // whether files can be selected; if false, only directories can be selected
	LoopEntries     bool                   // whether moving up at the top of the list should loop to the bottom, and vice versa

	FileSelected string

	history map[string]*historyEntry

	height     int
	AutoHeight bool

	Cursor string
	Styles Styles
}

// Gets the history entry for the current directory, or creates one if it doesn't exist.
func (m *Model) histEntry() *historyEntry {
	h, ok := m.history[m.currentDirectory]
	if !ok {
		h = &historyEntry{selected: 0, minIdx: 0, maxIdx: m.Height() - 1}
		m.history[m.currentDirectory] = h
	}
	return h
}

// Gets current valid history entry
func (m *Model) getHistEntry() *historyEntry {
	// init or get

	h := m.histEntry()
	// validate based on current state

	// if there are no entries, then set to default
	if len(m.entries) == 0 {
		h.selected = 0
		h.minIdx = 0
		h.maxIdx = m.Height() - 1
		return h
	}

	// there are entries, so make sure min, max, selected are all valid
	// fixSelected()

	h.selected = max(0, min(h.selected, len(m.entries)-1))
	h.minIdx = max(0, min(h.minIdx, h.selected))
	h.maxIdx = max(h.minIdx, min(h.maxIdx, h.selected+m.Height()-1, len(m.entries)-1))

	// and done!
	return h
}

func (m Model) readDir(path string, showHidden bool) tea.Cmd {
	return func() tea.Msg {
		dirEntries, err := os.ReadDir(path)
		if err != nil {
			return errorMsg{err}
		}

		sort.Slice(dirEntries, func(i, j int) bool {
			if dirEntries[i].IsDir() == dirEntries[j].IsDir() {
				return dirEntries[i].Name() < dirEntries[j].Name()
			}
			return dirEntries[i].IsDir()
		})

		// only keep entries that pass the filter
		filteredEntries := make([]os.DirEntry, 0, len(dirEntries))
		sanitizedAndFilteredEntries := make([]os.DirEntry, 0, len(dirEntries))

		for _, entry := range dirEntries {
			if m.entryFilter(entry) {
				filteredEntries = append(filteredEntries, entry)
				isHidden, _ := IsHidden(entry.Name())
				if !isHidden {
					sanitizedAndFilteredEntries = append(sanitizedAndFilteredEntries, entry)
				}
			}
		}

		if showHidden {
			return readDirMsg{id: m.id, entries: filteredEntries}
		}

		return readDirMsg{id: m.id, entries: sanitizedAndFilteredEntries}
	}
}

// SetHeight sets the height of the file picker.
func (m *Model) SetHeight(h int) {
	histEntry := m.getHistEntry()

	m.height = h
	if histEntry.maxIdx > m.height-1 {
		histEntry.maxIdx = histEntry.minIdx + m.height - 1
	}
}

// Height returns the height of the file picker.
func (m Model) Height() int {
	return m.height
}

// Init initializes the file picker model.
func (m Model) Init() tea.Cmd {
	return m.readDir(m.currentDirectory, m.ShowHidden)
}

// Update handles user interactions within the file picker model.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	// NOTE: This will initialize according to current working directory
	// path traversal need not do any initialization
	histEntry := m.getHistEntry()

	switch msg := msg.(type) {
	case readDirMsg:
		if msg.id != m.id {
			break
		}
		m.entries = msg.entries
		histEntry.maxIdx = max(histEntry.maxIdx, m.Height()-1)

	case tea.WindowSizeMsg:
		if m.AutoHeight {
			// set the height on the model
			m.SetHeight(msg.Height - marginBottom)
			histEntry.maxIdx = m.Height() - 1
			if histEntry.selected > histEntry.maxIdx {
				histEntry.selected = histEntry.maxIdx
			}
			if histEntry.minIdx > histEntry.selected {
				histEntry.minIdx = histEntry.selected
			}
		}

	case tea.KeyPressMsg:
		if key.Matches(msg, m.KeyMap.Select) {
			// Clear stale selection; this keypress must set m.Path in this frame to count as a selection.
			m.Path = ""
		}
		switch {
		case key.Matches(msg, m.KeyMap.GoToTop):
			histEntry.selected = 0
			histEntry.minIdx = 0
			histEntry.maxIdx = m.Height() - 1
		case key.Matches(msg, m.KeyMap.GoToLast):
			histEntry.selected = len(m.entries) - 1
			histEntry.minIdx = len(m.entries) - m.Height()
			histEntry.maxIdx = len(m.entries) - 1
		case key.Matches(msg, m.KeyMap.Down):
			histEntry.selected++
			// went out of bounds, but want to loop
			if histEntry.selected >= len(m.entries) && m.LoopEntries {
				histEntry.selected = 0
				histEntry.minIdx = 0
				histEntry.maxIdx = m.Height() - 1
			}
			// went out of bounds, but don't want to loop
			if histEntry.selected >= len(m.entries) && !m.LoopEntries {
				histEntry.selected = len(m.entries) - 1
			}
			// scroll window down if necessary
			if histEntry.selected > histEntry.maxIdx {
				histEntry.minIdx++
				histEntry.maxIdx++
			}
		case key.Matches(msg, m.KeyMap.Up):
			histEntry.selected--
			// went out of bounds, but want to loop
			if histEntry.selected < 0 && m.LoopEntries {
				histEntry.selected = len(m.entries) - 1
				histEntry.minIdx = len(m.entries) - m.Height()
				histEntry.maxIdx = len(m.entries) - 1
			}
			// went out of bounds, don't want to loop
			if histEntry.selected < 0 && !m.LoopEntries {
				histEntry.selected = 0
			}
			// scroll window up if necessary
			if histEntry.selected < histEntry.minIdx {
				histEntry.minIdx--
				histEntry.maxIdx--
			}
		case key.Matches(msg, m.KeyMap.PageDown):
			histEntry.selected += m.Height()
			if histEntry.selected >= len(m.entries) {
				histEntry.selected = len(m.entries) - 1
			}
			histEntry.minIdx += m.Height()
			histEntry.maxIdx += m.Height()

			if histEntry.maxIdx >= len(m.entries) {
				histEntry.maxIdx = len(m.entries) - 1
				histEntry.minIdx = histEntry.maxIdx - m.Height() + 1 // fix visual discrepancy between GoToLast and PageDowns
			}
		case key.Matches(msg, m.KeyMap.PageUp):
			histEntry.selected -= m.Height()
			if histEntry.selected < 0 {
				histEntry.selected = 0
			}
			histEntry.minIdx -= m.Height()
			histEntry.maxIdx -= m.Height()

			if histEntry.minIdx < 0 {
				histEntry.minIdx = 0
				histEntry.maxIdx = histEntry.minIdx + m.Height()
			}
		// NOTE: modified back behaviour; back always goes up one level
		case key.Matches(msg, m.KeyMap.Back):
			// get parent directory
			parentDir := filepath.Dir(m.currentDirectory)
			// we're at the root
			if parentDir == m.currentDirectory {
				break
			}
			// we want to go up; current state already save by mutability

			// switch directories
			m.currentDirectory = parentDir

			// we don't need to keep track of history here, this is already done
			// at top of the function by mutability of history map

			// and read it
			return m, m.readDir(m.currentDirectory, m.ShowHidden)

		// NOTE: modified open implementation; behaviour is identical
		case key.Matches(msg, m.KeyMap.Open):
			if len(m.entries) == 0 {
				break
			}

			entry := m.entries[histEntry.selected]
			info, err := entry.Info()
			if err != nil {
				break
			}
			isSymlink := info.Mode()&os.ModeSymlink != 0
			isDir := entry.IsDir()

			if isSymlink {
				symlinkPath, _ := filepath.EvalSymlinks(filepath.Join(m.currentDirectory, entry.Name()))
				info, err := os.Stat(symlinkPath)
				if err != nil {
					break
				}
				if info.IsDir() {
					isDir = true
				}
			}

			if (!isDir && m.FileAllowed) || (isDir && m.DirAllowed) {
				if key.Matches(msg, m.KeyMap.Select) {
					// Select the current path as the selection
					m.Path = filepath.Join(m.currentDirectory, entry.Name())
				}
			}

			if !isDir {
				break
			}

			// save folder name as current directory
			m.currentDirectory = filepath.Join(m.currentDirectory, entry.Name())
			// and read it -- no need for history management here
			// this is done by mutability of history map at top of the function
			return m, m.readDir(m.currentDirectory, m.ShowHidden)
		case key.Matches(msg, m.KeyMap.GoHome):
			// get the homedir
			homeDir, err := os.UserHomeDir()
			if err != nil {
				break
			}
			// set current directory to homedir
			m.currentDirectory = homeDir
			// and then read it -- no need for history management here
			return m, m.readDir(m.currentDirectory, m.ShowHidden)
		}
	}

	return m, nil
}

// View returns the view of the file picker.
func (m Model) View() string {
	histEntry := m.getHistEntry()

	if len(m.entries) == 0 {
		return m.Styles.EmptyDirectory.Height(m.Height()).MaxHeight(m.Height()).String()
	}
	var s strings.Builder

	for idx, entry := range m.entries {
		if idx < histEntry.minIdx || idx > histEntry.maxIdx {
			continue
		}

		var symlinkPath string
		info, err := entry.Info()
		if err != nil {
			continue
		}
		isSymlink := info.Mode()&os.ModeSymlink != 0
		size := strings.Replace(humanize.Bytes(uint64(info.Size())), " ", "", 1)
		name := entry.Name()

		if isSymlink {
			symlinkPath, _ = filepath.EvalSymlinks(filepath.Join(m.currentDirectory, name))
		}

		disabled := !m.canSelect(name) && !entry.IsDir()

		if histEntry.selected == idx {
			selected := ""
			if m.ShowPermissions {
				selected += " " + info.Mode().String()
			}
			if m.ShowSize {
				selected += fmt.Sprintf("%"+strconv.Itoa(m.Styles.FileSize.GetWidth())+"s", size)
			}
			selected += " " + name
			if isSymlink {
				selected += " → " + symlinkPath
			}
			if disabled {
				s.WriteString(m.Styles.DisabledCursor.Render(m.Cursor) + m.Styles.DisabledSelected.Render(selected))
			} else {
				s.WriteString(m.Styles.Cursor.Render(m.Cursor) + m.Styles.Selected.Render(selected))
			}
			s.WriteRune('\n')
			continue
		}

		style := m.Styles.File
		if entry.IsDir() {
			style = m.Styles.Directory
		} else if isSymlink {
			style = m.Styles.Symlink
		} else if disabled {
			style = m.Styles.DisabledFile
		}

		fileName := style.Render(name)
		s.WriteString(m.Styles.Cursor.Render(" "))
		if isSymlink {
			fileName += " → " + symlinkPath
		}
		if m.ShowPermissions {
			s.WriteString(" " + m.Styles.Permission.Render(info.Mode().String()))
		}
		if m.ShowSize {
			s.WriteString(m.Styles.FileSize.Render(size))
		}
		s.WriteString(" " + fileName)
		s.WriteRune('\n')
	}

	for i := lipgloss.Height(s.String()); i < m.Height(); i++ {
		s.WriteRune('\n')
	}

	return s.String()
}

// DidSelectFile returns whether a user has selected a file (on this msg).
func (m Model) DidSelectFile(msg tea.Msg) (bool, string) {
	didSelect, path := m.didSelectFile(msg)
	if didSelect && m.canSelect(path) {
		return true, path
	}
	return false, ""
}

// DidSelectDisabledFile returns whether a user tried to select a disabled file
// (on this msg). This is necessary only if you would like to warn the user that
// they tried to select a disabled file.
func (m Model) DidSelectDisabledFile(msg tea.Msg) (bool, string) {
	didSelect, path := m.didSelectFile(msg)
	if didSelect && !m.canSelect(path) {
		return true, path
	}
	return false, ""
}

func (m Model) didSelectFile(msg tea.Msg) (bool, string) {
	histEntry := m.getHistEntry()

	// shortcut on no entries
	if len(m.entries) == 0 {
		return false, ""
	}

	// selection can only happen on keypressmessages, when key matches Select
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok || !key.Matches(keyMsg, m.KeyMap.Select) {
		return false, ""
	}

	entry := m.entries[histEntry.selected]
	info, err := entry.Info()
	if err != nil {
		return false, ""
	}
	isSymlink := info.Mode()&os.ModeSymlink != 0
	isDir := entry.IsDir()

	if isSymlink {
		symlinkPath, _ := filepath.EvalSymlinks(filepath.Join(m.currentDirectory, entry.Name()))
		info, err := os.Stat(symlinkPath)
		if err != nil {
			return false, ""
		}
		isDir = isDir || info.IsDir()
	}

	if ((!isDir && m.FileAllowed) || (isDir && m.DirAllowed)) && m.Path != "" {
		return true, m.Path
	}

	return false, ""
}

func (m Model) canSelect(file string) bool {
	if len(m.AllowedTypes) <= 0 {
		return true
	}

	for _, ext := range m.AllowedTypes {
		if strings.HasSuffix(file, ext) {
			return true
		}
	}
	return false
}

// HighlightedPath returns the path of the currently highlighted file or directory.
func (m Model) HighlightedPath() string {
	histEntry := m.getHistEntry()

	if len(m.entries) == 0 || histEntry.selected < 0 || histEntry.selected >= len(m.entries) {
		return ""
	}
	return filepath.Join(m.currentDirectory, m.entries[histEntry.selected].Name())
}
