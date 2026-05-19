// A filepicker bubble that makes it virtually impossible to get stuck in a directory.
// It supports a few extra things over the charm filepicker, too.
// Paths are automatically made absolute, and hence are traversable.
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

//                    .___     .__
//   _____   ____   __| _/____ |  |
//  /     \ /  _ \ / __ |/ __ \|  |
// |  Y Y  (  <_> ) /_/ \  ___/|  |__
// |__|_|  /\____/\____ |\___  >____/
//       \/            \/    \/

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
	GoHome   key.Binding
}

// DefaultKeyMap returns the default key bindings.
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
		GoHome:   key.NewBinding(key.WithKeys("~"), key.WithHelp("~", "go home")),
	}
}

// Styling constants
const (
	// The margin between the bottom of the filepicker and the bottom of the window.
	MARGIN_BOTTOM = 5
	// The width of the file size column. This is necessary to prevent the file size from shifting around as you navigate the file picker.
	FILE_SIZE_WIDTH = 7
	// The padding between the left edge of the filepicker and the text in the empty directory message.
	PADDING_LEFT = 2
)

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
		FileSize:         lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Width(FILE_SIZE_WIDTH).Align(lipgloss.Right),
		EmptyDirectory:   lipgloss.NewStyle().Foreground(lipgloss.Color("240")).PaddingLeft(PADDING_LEFT).SetString("Bummer. No Files Found."),
	}
}

// Model represents a file picker.
type Model struct {
	// private fields
	id               int
	currentDirectory string
	entries          []os.DirEntry
	entryFilter      func(os.DirEntry) bool // true == keep
	height           int
	history          map[string]*historyEntry

	// The path which the user has selected with the file picker.
	// This may be an invalid path according to AllowedTypes.
	Path string

	// AllowedTypes specifies which file types the user may select.
	// If empty the user may select any file.
	AllowedTypes []string

	// The keybindings that are used in the file picker.
	// Defaults are sensible, and have help text for charm's help bubble.
	KeyMap KeyMap

	// Flag to determine whether to show file permissions.
	ShowPermissions bool
	// Flag to determine whether to show file sizes.
	ShowSize bool
	// Flag to determine whether to show hidden files.
	ShowHidden bool
	// Flag to determine whether directories can be selected.
	DirAllowed bool
	// Flag to determine whether files can be selected.
	FileAllowed bool
	// Flag to determine whether selection should loop around when going out of bounds.
	LoopEntries bool

	// Flag to determine whether the filepicker should automatically set its height
	// to the available vertical space in the window.
	// If false, you must set the height manually with SetHeight,
	// and it will not change on window resize.
	AutoHeight bool

	// The string used to indicate the currently highlighted entry.
	Cursor string

	// Styles contains the styling for different parts of the file picker.
	Styles Styles
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

// Init initializes the file picker model.
func (m Model) Init() tea.Cmd {
	return m.readDir(m.currentDirectory, m.ShowHidden)
}

// Update handles user interactions within the file picker model.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case readDirMsg:
		return m.handleReadDirMsg(msg)

	case tea.WindowSizeMsg:
		return m.handleResizeMsg(msg)

	case tea.KeyPressMsg:
		// NOTE: Since Select and Open both use `Enter`, we need to
		// move this check outside of the switch, so the event can bubble through
		if key.Matches(msg, m.KeyMap.Select) {
			m.Path = ""
		}

		switch {
		case key.Matches(msg, m.KeyMap.GoToTop):
			return m.handleGoToTop(msg)
		case key.Matches(msg, m.KeyMap.GoToLast):
			return m.handleGoToLast(msg)
		case key.Matches(msg, m.KeyMap.Down):
			return m.handleDown(msg)
		case key.Matches(msg, m.KeyMap.Up):
			return m.handleUp(msg)
		case key.Matches(msg, m.KeyMap.PageDown):
			return m.handlePageDown(msg)
		case key.Matches(msg, m.KeyMap.PageUp):
			return m.handlePageUp(msg)
		case key.Matches(msg, m.KeyMap.Back):
			return m.handleBack(msg)
		case key.Matches(msg, m.KeyMap.Open):
			return m.handleOpen(msg)
		case key.Matches(msg, m.KeyMap.GoHome):
			return m.handleGoHome(msg)
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
		meta, ok := m.buildEntryMeta(entry)
		if !ok {
			continue
		}
		s.WriteString(m.renderEntry(meta, histEntry.selected == idx))
		s.WriteRune('\n')
	}

	return m.padToHeight(s.String())
}

// .__                       .___.__
// |  |__ _____    ____    __| _/|  |   ___________  ______
// |  |  \\__  \  /    \  / __ | |  | _/ __ \_  __ \/  ___/
// |   Y  \/ __ \|   |  \/ /_/ | |  |_\  ___/|  | \/\___ \
// |___|  (____  /___|  /\____ | |____/\___  >__|  /____  >
//      \/     \/     \/      \/           \/           \/

func (m Model) handleReadDirMsg(msg readDirMsg) (Model, tea.Cmd) {
	histEntry := m.getHistEntry()

	if msg.id != m.id {
		return m, nil
	}
	m.entries = msg.entries
	histEntry.maxIdx = max(histEntry.maxIdx, m.Height()-1)

	return m, nil
}

func (m Model) handleResizeMsg(msg tea.WindowSizeMsg) (Model, tea.Cmd) {
	histEntry := m.getHistEntry()

	if m.AutoHeight {
		// set the height on the model
		m.SetHeight(msg.Height - MARGIN_BOTTOM)
		histEntry.maxIdx = m.Height() - 1
		if histEntry.selected > histEntry.maxIdx {
			histEntry.selected = histEntry.maxIdx
		}
		if histEntry.minIdx > histEntry.selected {
			histEntry.minIdx = histEntry.selected
		}
	}

	return m, nil
}

func (m Model) handleGoToTop(_ tea.KeyPressMsg) (Model, tea.Cmd) {
	histEntry := m.getHistEntry()

	histEntry.selected = 0
	histEntry.minIdx = 0
	histEntry.maxIdx = m.Height() - 1
	return m, nil
}

func (m Model) handleGoToLast(_ tea.KeyPressMsg) (Model, tea.Cmd) {
	histEntry := m.getHistEntry()

	histEntry.selected = len(m.entries) - 1
	histEntry.minIdx = len(m.entries) - m.Height()
	histEntry.maxIdx = len(m.entries) - 1
	return m, nil
}

func (m Model) handleDown(_ tea.KeyPressMsg) (Model, tea.Cmd) {
	histEntry := m.getHistEntry()

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
	return m, nil
}

func (m Model) handleUp(_ tea.KeyPressMsg) (Model, tea.Cmd) {
	histEntry := m.getHistEntry()

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
	return m, nil
}

func (m Model) handlePageDown(_ tea.KeyPressMsg) (Model, tea.Cmd) {
	histEntry := m.getHistEntry()

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
	return m, nil
}

func (m Model) handlePageUp(_ tea.KeyPressMsg) (Model, tea.Cmd) {
	histEntry := m.getHistEntry()

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
	return m, nil
}

func (m Model) handleBack(_ tea.KeyPressMsg) (Model, tea.Cmd) {
	histEntry := m.getHistEntry()
	_ = histEntry

	// get parent directory
	parentDir := filepath.Dir(m.currentDirectory)
	// we're at the root
	if parentDir == m.currentDirectory {
		return m, nil
	}
	// we want to go up; current state already save by mutability

	// switch directories
	m.currentDirectory = parentDir

	// we don't need to keep track of history here, this is already done
	// at top of the function by mutability of history map

	// and read it
	return m, m.readDir(m.currentDirectory, m.ShowHidden)
}

func (m Model) handleOpen(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	histEntry := m.getHistEntry()

	if len(m.entries) == 0 {
		return m, nil
	}

	entry := m.entries[histEntry.selected]
	info, err := entry.Info()
	if err != nil {
		return m, nil
	}
	isSymlink := info.Mode()&os.ModeSymlink != 0
	isDir := entry.IsDir()

	if isSymlink {
		symlinkPath, _ := filepath.EvalSymlinks(filepath.Join(m.currentDirectory, entry.Name()))
		info, err := os.Stat(symlinkPath)
		if err != nil {
			return m, nil
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
		return m, nil
	}

	// save folder name as current directory
	m.currentDirectory = filepath.Join(m.currentDirectory, entry.Name())
	// and read it -- no need for history management here
	// this is done by mutability of history map at top of the function
	return m, m.readDir(m.currentDirectory, m.ShowHidden)
}

func (m Model) handleGoHome(_ tea.KeyPressMsg) (Model, tea.Cmd) {
	histEntry := m.getHistEntry()
	_ = histEntry

	// get the homedir
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return m, nil
	}
	// set current directory to homedir
	m.currentDirectory = homeDir
	// and then read it -- no need for history management here
	return m, m.readDir(m.currentDirectory, m.ShowHidden)
}

//                            .___           .__
// _______   ____   ____    __| _/___________|__| ____    ____
// \_  __ \_/ __ \ /    \  / __ |/ __ \_  __ \  |/    \  / ___\
//  |  | \/\  ___/|   |  \/ /_/ \  ___/|  | \/  |   |  \/ /_/  >
//  |__|    \___  >___|  /\____ |\___  >__|  |__|___|  /\___  /
//              \/     \/      \/    \/              \//_____/

type entryMeta struct {
	name        string
	size        string
	mode        os.FileMode
	isDir       bool
	isSymlink   bool
	symlinkPath string
	disabled    bool
}

func (m Model) buildEntryMeta(entry os.DirEntry) (entryMeta, bool) {
	info, err := entry.Info()
	if err != nil {
		return entryMeta{}, false
	}

	name := entry.Name()
	size := strings.Replace(humanize.Bytes(uint64(info.Size())), " ", "", 1)
	mode := info.Mode()
	isDir := entry.IsDir()
	isSymlink := info.Mode()&os.ModeSymlink != 0
	var symlinkPath string
	if isSymlink {
		symlinkPath, _ = filepath.EvalSymlinks(filepath.Join(m.currentDirectory, name))
	}
	disabled := !m.canSelect(name) && !isDir

	return entryMeta{
		name:        name,
		size:        size,
		mode:        mode,
		isDir:       isDir,
		isSymlink:   isSymlink,
		symlinkPath: symlinkPath,
		disabled:    disabled,
	}, true
}

func (m Model) renderEntry(meta entryMeta, selected bool) string {
	if selected {
		selectedText := ""
		if m.ShowPermissions {
			selectedText += " " + meta.mode.String()
		}
		if m.ShowSize {
			selectedText += fmt.Sprintf("%"+strconv.Itoa(m.Styles.FileSize.GetWidth())+"s", meta.size)
		}
		selectedText += " " + meta.name
		if meta.isSymlink {
			selectedText += " → " + meta.symlinkPath
		}

		if meta.disabled {
			return m.Styles.DisabledCursor.Render(m.Cursor) + m.Styles.DisabledSelected.Render(selectedText)
		}
		return m.Styles.Cursor.Render(m.Cursor) + m.Styles.Selected.Render(selectedText)
	}

	fileName := m.styleForEntry(meta).Render(meta.name)
	if meta.isSymlink {
		fileName += " → " + meta.symlinkPath
	}

	var line strings.Builder
	line.WriteString(m.Styles.Cursor.Render(" "))
	if m.ShowPermissions {
		line.WriteString(" " + m.Styles.Permission.Render(meta.mode.String()))
	}
	if m.ShowSize {
		line.WriteString(m.Styles.FileSize.Render(meta.size))
	}
	line.WriteString(" " + fileName)

	return line.String()
}

func (m Model) styleForEntry(meta entryMeta) lipgloss.Style {
	if meta.isDir {
		return m.Styles.Directory
	}
	if meta.isSymlink {
		return m.Styles.Symlink
	}
	if meta.disabled {
		return m.Styles.DisabledFile
	}
	return m.Styles.File
}

func (m Model) padToHeight(s string) string {
	height := lipgloss.Height(s)
	if height >= m.Height() {
		return s
	}

	var out strings.Builder
	out.WriteString(s)
	for i := height; i < m.Height(); i++ {
		out.WriteRune('\n')
	}

	return out.String()
}

//             ___.   .__  .__                       .__
// ______  __ _\_ |__ |  | |__| ____   _____  ______ |__|
// \____ \|  |  \ __ \|  | |  |/ ___\  \__  \ \____ \|  |
// |  |_> >  |  / \_\ \  |_|  \  \___   / __ \|  |_> >  |
// |   __/|____/|___  /____/__|\___  > (____  /   __/|__|
// |__|             \/             \/       \/|__|

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

// Set the filter function for the filepicker.
// Entries that evaluate to true will be shown.
func (m *Model) SetFilterFunc(fn func(os.DirEntry) bool) {
	m.entryFilter = fn
}

// Ask the filepicker to refresh its entries by re-reading the current directory.
// Necessary to call after SetFilterFunc for the new filter function to take effect.
func (m *Model) RefreshEntries() tea.Cmd {
	return m.readDir(m.currentDirectory, m.ShowHidden)
}

// The number of entries in the current directory. Can be used to check for empty directories.
func (m *Model) NumEntries() int {
	return len(m.entries)
}

// CurrentDirectory returns the absolute path of the directory the filepicker is currently showing.
func (m *Model) CurrentDirectory() string {
	return m.currentDirectory
}

// HighlightedPath returns the path of the currently highlighted file or directory.
func (m Model) HighlightedPath() string {
	histEntry := m.getHistEntry()

	if len(m.entries) == 0 || histEntry.selected < 0 || histEntry.selected >= len(m.entries) {
		return ""
	}
	return filepath.Join(m.currentDirectory, m.entries[histEntry.selected].Name())
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

type errorMsg struct {
	err error
}

type readDirMsg struct {
	id      int
	entries []os.DirEntry
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

func (m *Model) histEntry() *historyEntry {
	// get hist entry
	h, ok := m.history[m.currentDirectory]
	if !ok {
		// create if doesn't exist
		h = &historyEntry{selected: 0, minIdx: 0, maxIdx: m.Height() - 1}
		m.history[m.currentDirectory] = h
	}
	// return it
	return h
}

func (m *Model) getHistEntry() *historyEntry {
	// init or get
	h := m.histEntry()

	// if there are no entries, then set to default
	if len(m.entries) == 0 {
		h.selected = 0
		h.minIdx = 0
		h.maxIdx = m.Height() - 1
		return h
	}

	// there are entries, so make sure min, max, selected are all valid
	h.selected = max(0, min(h.selected, len(m.entries)-1))
	h.minIdx = max(0, min(h.minIdx, h.selected))
	h.maxIdx = max(h.minIdx, min(h.maxIdx, h.selected+m.Height()-1, len(m.entries)-1))

	// and done!
	return h
}

func (m Model) sortedEntriesFor(path string) ([]os.DirEntry, error) {
	dirEntries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	sort.Slice(dirEntries, func(i, j int) bool {
		if dirEntries[i].IsDir() == dirEntries[j].IsDir() {
			return dirEntries[i].Name() < dirEntries[j].Name()
		}
		return dirEntries[i].IsDir()
	})

	return dirEntries, nil
}

func (m Model) readDir(path string, showHidden bool) tea.Cmd {
	return func() tea.Msg {
		// get entries for this path in known sorted order
		dirEntries, err := m.sortedEntriesFor(path)
		if err != nil {
			return errorMsg{err}
		}

		// only keep entries that pass the filter; at the same time, remove hidden files if necessary
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
