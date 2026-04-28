package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
	// "regexp"
	"strings"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/glamour/v2"
	"charm.land/lipgloss/v2"
)

type Note struct {
	Title string
	Path  string
}

type item struct {
	title string
	path  string
}

type notesLoadedMsg []Note

func (i item) Title() string       { return i.title }
func (i item) Description() string { return "" }
func (i item) FilterValue() string { return i.title }

func (m *model) selectedNote() (item, bool) {
	selected := m.list.SelectedItem()
	if selected == nil {
		return item{}, false
	}

	it, ok := selected.(item)
	return it, ok
}

type model struct {
	list        list.Model
	content     string
	viewing     bool
	editing     bool
	settingPath bool
	notesPath   string
	loading     bool
	loadingMsg  string

	width  int
	height int

	spinner   spinner.Model
	viewport  viewport.Model
	renderer  *glamour.TermRenderer
	textinput textinput.Model
	textarea  textarea.Model
}

// HELPER FUNCTIONS
func loadNotes(path string) []Note {
	files, err := os.ReadDir(path)
	if err != nil {
		return nil
	}
	var notes []Note

	for _, file := range files {
		name := file.Name()

		if strings.HasPrefix(name, ".") {
			continue
		}

		fullPath := filepath.Join(path, name)

		note := Note{
			Title: name,
			Path:  fullPath,
		}
		notes = append(notes, note)
	}

	return notes
}

func (m *model) layoutSizes() {
	titleStyle := lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center)

	contentStyle := lipgloss.NewStyle().
		Width(m.width).
		BorderStyle(lipgloss.RoundedBorder())

	helpStyle := lipgloss.NewStyle().
		Width(m.width)

	title := titleStyle.Render("X")
	help := helpStyle.Render("")
	usedHeight := lipgloss.Height(title) + lipgloss.Height(help)
	contentHeight := m.height - usedHeight
	innerHeight := contentHeight - contentStyle.GetVerticalFrameSize()

	if innerHeight < 1 {
		innerHeight = 1
	}

	if m.viewing || m.editing {
		viewportWidth := m.width - contentStyle.GetHorizontalFrameSize()
		if viewportWidth < 1 {
			viewportWidth = 1
		}

		if m.editing {
			m.textarea.SetWidth(viewportWidth)
			m.textarea.SetHeight(innerHeight)
		} else {
			m.viewport.SetWidth(viewportWidth)
			m.viewport.SetHeight(innerHeight)
		}
	} else {
		leftWidth := m.width / 4
		rightWidth := m.width - leftWidth

		listWidth := leftWidth - contentStyle.GetHorizontalFrameSize()
		viewportWidth := rightWidth - contentStyle.GetHorizontalFrameSize()

		if listWidth < 1 {
			listWidth = 1
		}
		if viewportWidth < 1 {
			viewportWidth = 1
		}

		m.list.SetWidth(listWidth)
		m.list.SetHeight(innerHeight)

		m.viewport.SetWidth(viewportWidth)
		m.viewport.SetHeight(innerHeight)
	}
}

func (m *model) renderSelectedNote() {
	it, ok := m.selectedNote()
	if !ok {
		return
	}

	content, err := os.ReadFile(it.path)
	if err != nil {
		return
	}

	m.content = string(content)
	r, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithPreservedNewLines(),
		glamour.WithWordWrap(0),
	)
	if err != nil {
		return
	}

	text, err := r.Render(m.content)
	if err != nil {
		return
	}

	m.viewport.SetContent(text)
}

func (m *model) enterEditMode() tea.Cmd {
	it, ok := m.selectedNote()
	if !ok {
		return nil
	}

	content, err := os.ReadFile(it.path)
	if err != nil {
		return nil
	}

	m.textarea.SetValue(string(content))
	m.editing = true
	m.layoutSizes()
	return m.textarea.Focus()
}

func (m *model) saveEditedNote() {
	it, ok := m.selectedNote()
	if !ok {
		return
	}

	if err := os.WriteFile(it.path, []byte(m.textarea.Value()), 0644); err != nil {
		return
	}

	m.editing = false
	m.textarea.Blur()
	m.renderSelectedNote()
	m.layoutSizes()
}

func (m *model) cancelEditMode() {
	m.editing = false
	m.textarea.Blur()
	m.renderSelectedNote()
	m.layoutSizes()
}

func fancyHeader(title string, width int) string {
	if width == 0 {
		return title
	}

	titleLen := len([]rune(title))
	padding := width - titleLen
	lpad := (padding / 2) - 1
	rpad := (padding - lpad) - 1

	line := strings.Repeat("/", lpad) + " " + title + " " + strings.Repeat("/", rpad)

	runes := []rune(line)

	colors := lipgloss.Blend1D(
		len(runes),
		lipgloss.Color("#7D56F4"),
		lipgloss.Color("#FF5F87"),
	)

	out := ""
	for i, r := range runes {
		style := lipgloss.NewStyle().Foreground(colors[i])
		out += style.Render(string(r))
	}

	return out
}

func getConfigFile() (string, error) {
	directory, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	appDirectory := filepath.Join(directory, "notes_tui")
	err = os.MkdirAll(appDirectory, 0755)
	if err != nil {
		return "", err
	}

	return filepath.Join(appDirectory, "path.txt"), nil
}

func savePath(path string) error {
	file, err := getConfigFile()
	if err != nil {
		return err
	}

	return os.WriteFile(file, []byte(path), 0644)
}

func loadPathFile() (string, error) {
	pathFile, err := getConfigFile()
	if err != nil {
		return "", err
	}

	path, err := os.ReadFile(pathFile)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(path)), nil
}

// Bubbletea Commands
func loadNotesCmd(path string) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(1000 * time.Millisecond)

		notes := loadNotes(path)
		return notesLoadedMsg(notes)
	}
}

// ACTUAL TUI MODEL

func initialModel() model {
	notesPath, err := loadPathFile()
	if err != nil {

	}

	items := []list.Item{}

	if notesPath != "" {

		for _, n := range loadNotes(notesPath) {
			items = append(items, item{title: n.Title, path: n.Path})
		}
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "NOTES"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.SetShowFilter(true)
	l.SetShowHelp(false)

	textinput := textinput.New()
	textinput.Styles().Blurred.Text.AlignHorizontal(lipgloss.Center)
	textinput.Placeholder = "full/path/to/notes"
	textinput.Focus()

	ta := textarea.New()
	ta.SetStyles(textarea.DefaultDarkStyles())
	ta.Prompt = ""
	ta.Placeholder = "Write your note..."
	ta.Blur()

	s := spinner.New()
	s.Spinner = spinner.Meter

	havePath := true
	if notesPath == "" {
		havePath = false
	}

	return model{

		list:        l,
		viewport:    viewport.Model{},
		spinner:     s,
		settingPath: !havePath,
		textinput:   textinput,
		textarea:    ta,
		notesPath:   notesPath,
	}
}

func (m model) Init() tea.Cmd {

	return m.spinner.Tick

}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var cmd tea.Cmd
	var cmds []tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	cmds = append(cmds, cmd)

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:

		m.width = msg.Width
		m.height = msg.Height
		m.layoutSizes()
		if !m.settingPath && !m.editing {
			m.renderSelectedNote()
		}

	case notesLoadedMsg:
		m.loading = false

		items := []list.Item{}
		for _, n := range msg {
			items = append(items, item{title: n.Title, path: n.Path})
		}

		m.list.SetItems(items)
		m.settingPath = false
		m.layoutSizes()

		return m, nil

	case tea.KeyPressMsg:
		if m.list.SettingFilter() {
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}

		switch msg.String() {
		case "e":
			if !m.editing {
				return m, m.enterEditMode()
			}

		case "ctrl+s":
			if m.editing {
				m.saveEditedNote()
				return m, nil
			}

		case "esc":
			if m.editing {
				m.cancelEditMode()
				return m, nil
			}

		case "tab", "ctrl+i":
			if m.editing {
				m.textarea.InsertString("    ")
				return m, nil
			}

		case "ctrl+c", "q":
			if m.viewing {
				m.viewing = false
				m.layoutSizes()
				m.renderSelectedNote()
				return m, nil
			}
			if !m.editing {
				return m, tea.Quit
			}
		case "enter":
			if m.editing {
				m.textarea.InsertString("\n")
				return m, nil
			}

			if m.viewing == false && m.settingPath == false {
				m.viewing = true
			}
			if m.settingPath == true {
				path := strings.TrimSpace(m.textinput.Value())
				err := savePath(path)
				if err != nil {
					fmt.Println("File not found!, please write the whole path, '~' expansion is not supported yet.")
				}

				if path != "" {
					m.loading = true
					m.loadingMsg = "Loading notes..."
					return m, tea.Batch(
						m.spinner.Tick,
						loadNotesCmd(path),
					)
				}

			}
			m.layoutSizes()
			m.renderSelectedNote()
			return m, nil
		}

	}

	if m.settingPath {
		m.textinput, cmd = m.textinput.Update(msg)
		cmds = append(cmds, cmd)
	}
	if m.editing {
		m.textarea, cmd = m.textarea.Update(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)
	}
	if m.viewing {
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)

	}
	if !m.viewing && !m.settingPath {
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
	}
	if !m.settingPath && !m.editing {
		m.renderSelectedNote()
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() tea.View {
	if m.settingPath {

		promptWidth := m.width / 3
		promptHeight := 5

		contentStyle := lipgloss.NewStyle().
			Width(m.width).
			Height(m.height - 2).
			Foreground(lipgloss.Color("240")).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("243"))

		promptStyle := lipgloss.NewStyle().
			Width(promptWidth).
			Align(lipgloss.Center).
			Foreground(lipgloss.BrightWhite).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForegroundBlend(lipgloss.Color("#FF5F87"), lipgloss.Color("#7D56F4")).
			Padding(1)

		helpStyle := lipgloss.NewStyle().
			Width(m.width).
			Align(lipgloss.Center).
			Foreground(lipgloss.Color("243"))

		backgroundFiller := strings.Builder{}

		rows := m.height - 4
		cols := m.width - 2
		for range rows {
			for range cols {
				backgroundFiller.WriteString("#")
			}
		}

		x := (m.width - promptWidth) / 2
		y := (m.height - promptHeight) / 2

		m.textinput.SetWidth(promptWidth)

		title := fancyHeader("TERMINAL NOTES", m.width)
		content := contentStyle.Render(backgroundFiller.String())
		help := helpStyle.Render("q quit")

		var promptContent string

		if m.loading {
			promptContent = lipgloss.JoinVertical(
				lipgloss.Center,
				"Loading notes...",
				"",
				m.spinner.View(),
			)
		} else {
			promptContent = lipgloss.JoinVertical(
				lipgloss.Top,
				"Enter your notes's file path:",
				"",
				lipgloss.NewStyle().Foreground(lipgloss.BrightWhite).Render(m.textinput.View()),
				"",
				"",
				lipgloss.NewStyle().
					Foreground(lipgloss.Color("243")).
					Align(lipgloss.Center).
					Render("<enter> to load the notes."),
			)
		}

		prompt := promptStyle.Render(promptContent)

		background := lipgloss.JoinVertical(lipgloss.Top, title, content, help)

		layers := []*lipgloss.Layer{
			lipgloss.NewLayer(background),
			lipgloss.NewLayer(prompt).X(x).Y(y).Z(1),
		}

		screen := lipgloss.NewCompositor(layers...)
		v := tea.NewView(screen.Render())
		v.AltScreen = true

		return v
	}

	if m.editing {
		selected, ok := m.selectedNote()
		if !ok {
			return tea.NewView("")
		}

		contentStyle := lipgloss.NewStyle().
			Width(m.width).
			Align(lipgloss.Left).
			Foreground(lipgloss.BrightWhite).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("243"))

		helpStyle := lipgloss.NewStyle().
			Width(m.width).
			Align(lipgloss.Center).
			Foreground(lipgloss.Color("243"))

		title := fancyHeader("EDITING: "+selected.title, m.width)
		content := contentStyle.Render(m.textarea.View())
		help := helpStyle.Render("Ctrl+S save • Esc cancel")

		v := tea.NewView(title + "\n" + content + "\n" + help)
		v.AltScreen = true

		return v
	}

	if m.viewing {
		selected, ok := m.selectedNote()
		if !ok {
			return tea.NewView("")
		}

		contentStyle := lipgloss.NewStyle().
			Width(m.width).
			Align(lipgloss.Left).
			Foreground(lipgloss.BrightWhite).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("243"))

		helpStyle := lipgloss.NewStyle().
			Width(m.width).
			Align(lipgloss.Center).
			Foreground(lipgloss.Color("243"))

		title := fancyHeader(selected.title, m.width)
		content := contentStyle.Render(m.viewport.View())
		help := helpStyle.Render("↑/k scroll up • ↓/j scroll down • ←/h left • →/l right • e edit • q back")

		v := tea.NewView(title + "\n" + content + "\n" + help)
		v.AltScreen = true

		return v
	}

	leftWidth := m.width / 4
	rightWidth := m.width - leftWidth

	leftStyle := lipgloss.NewStyle().
		Width(leftWidth).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("243"))

	rightStyle := lipgloss.NewStyle().
		Width(rightWidth).
		Height(m.viewport.Height()).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("243"))

	helpStyle := lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center).
		Foreground(lipgloss.Color("243"))

	title := fancyHeader("TEMINAL NOTES", m.width)
	left := leftStyle.Render(m.list.View())
	right := rightStyle.Render(m.viewport.View())

	content := lipgloss.JoinHorizontal(
		lipgloss.Top,
		left,
		right,
	)

	help := helpStyle.Render("↑/k up • ↓/j down • / filter • enter full view • q quit")

	v := tea.NewView(title + "\n" + content + "\n" + help)
	v.AltScreen = true

	return v
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
