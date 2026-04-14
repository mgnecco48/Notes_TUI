package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/glamour/v2"
	"charm.land/lipgloss/v2"
)

// GETTING THE PATH FOR THE NOTES DIRECTORY
const configFile = ".notes_app"

func loadPathFile() (string, error) {
	file, err := os.ReadFile(configFile)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(file)), nil
}

func savePathToFile(path string) error {
	return os.WriteFile(configFile, []byte(path), 0644)
}

type Note struct {
	Title string
	Path  string
}

type item struct {
	title string
	path  string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return "" }
func (i item) FilterValue() string { return i.title }

type model struct {
	list        list.Model
	cursor      int
	content     string
	viewing     bool
	settingPath bool

	width  int
	height int

	viewport  viewport.Model
	renderer  *glamour.TermRenderer
	textInput textinput.Model
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

	title := titleStyle.Render("X")
	usedHeight := lipgloss.Height(title)
	contentHeight := m.height - usedHeight
	innerHeight := contentHeight - contentStyle.GetVerticalFrameSize()

	if innerHeight < 1 {
		innerHeight = 1
	}

	if m.viewing {
		viewportWidth := m.width - contentStyle.GetHorizontalFrameSize()
		if viewportWidth < 1 {
			viewportWidth = 1
		}

		m.viewport.SetWidth(viewportWidth)
		m.viewport.SetHeight(innerHeight)
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
	selected := m.list.SelectedItem()
	if selected == nil {
		return
	}

	it := selected.(item)

	content, err := os.ReadFile(it.path)
	if err != nil {
		return
	}

	m.content = string(content)
	r, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
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

// INPUT COMMAND FUNCTION

// ACTUAL TUI MODEL

func initialModel() model {

	items := []list.Item{}
	for _, n := range loadNotes("/Users/martin/Documents/notes/") {
		items = append(items, item{title: n.Title, path: n.Path})

	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "NOTES"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)

	pathPrompt := textinput.New()
	pathPrompt.Placeholder = "~/path/to/notes"
	pathPrompt.Focus()
	pathPrompt.SetWidth(40)

	return model{

		list:        l,
		viewport:    viewport.Model{},
		settingPath: true,
		textInput:   pathPrompt,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var cmd tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:

		m.width = msg.Width
		m.height = msg.Height
		m.layoutSizes()
		m.renderSelectedNote()
		m.viewport.HighlightStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("238")).Background(lipgloss.Color("34"))
		m.viewport.SelectedHighlightStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("238")).Background(lipgloss.Color("47"))
		m.viewport.SetHighlights(regexp.MustCompile("is").FindAllStringIndex(m.content, -1))
		m.viewport.HighlightNext()

	case tea.KeyPressMsg:

		switch msg.String() {

		case "ctrl+c", "q":
			if m.viewing {
				m.viewing = false
				m.layoutSizes()
				m.renderSelectedNote()
				return m, nil
			}
			return m, tea.Quit

		case "enter":
			m.viewing = !m.viewing
			m.layoutSizes()
			m.renderSelectedNote()
			return m, nil
		}

	}

	if m.viewing {
		m.viewport, cmd = m.viewport.Update(msg)

	}
	if !m.viewing {
		m.list, cmd = m.list.Update(msg)
	}
	m.renderSelectedNote()

	return m, cmd
}

func (m model) View() tea.View {
	if m.viewing {
		selected := m.list.SelectedItem().(item)

		contentStyle := lipgloss.NewStyle().
			Width(m.width).
			Align(lipgloss.Left).
			Foreground(lipgloss.BrightWhite).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240"))

		hintStyle := lipgloss.NewStyle().
			Width(m.width).
			Align(lipgloss.Center).
			Foreground(lipgloss.White).Faint(true)

		title := fancyHeader(selected.title, m.width)
		content := contentStyle.Render(m.viewport.View())
		hint := hintStyle.Render("Press q to go back")

		v := tea.NewView(title + "\n" + content + "\n" + hint)
		v.AltScreen = true

		return v
	}

	leftWidth := m.width / 4
	rightWidth := m.width - leftWidth

	leftStyle := lipgloss.NewStyle().
		Width(leftWidth).
		Height(m.viewport.Height()).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240"))

	rightStyle := lipgloss.NewStyle().
		Width(rightWidth).
		Height(m.viewport.Height()).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240"))

	m.list.SetHeight(m.viewport.Height())

	title := fancyHeader("MY NOTES APP", m.width)
	left := leftStyle.Render(m.list.View())
	right := rightStyle.Render(m.viewport.View())

	content := lipgloss.JoinHorizontal(
		lipgloss.Top,
		left,
		right,
	)

	v := tea.NewView(title + "\n" + content)
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
