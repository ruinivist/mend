/*
this pakckage basically handles notes which are the individual files for spaced
repetition in mend
- ruinivist, 3Jan26
*/

package main

import (
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
)

type NoteView struct {
	// the actual content
	path    string
	title   string   // the first line is always # title
	content string   // strip the first line from file rest is content
	hints   []string //a hint is anything that matches (hint: <text>) in content
	// display layer
	err        error
	loading    bool
	vp         viewport.Model
	mdRenderer *glamour.TermRenderer
}

// ================== messages ===================
type loadNote struct {
	path string
}

type loadedNote struct {
	title   string
	content string
	hints   []string
	err     error
}

func NewNoteView() *NoteView {
	mdRenderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)
	return &NoteView{
		loading:    true,
		mdRenderer: mdRenderer,
		vp:         viewport.New(0, 0),
	}
}

func (m *NoteView) Init() tea.Cmd {
	return nil
}

func (m *NoteView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.vp.Width = msg.Width
		m.vp.Height = msg.Height
		return m, nil

	case loadNote:
		if m.path == msg.path {
			return m, nil //noop
		}
		m.path = msg.path
		m.loading = true
		return m, fetchContent(msg.path)

	case loadedNote:
		m.loading = false
		m.title = msg.title
		m.content = msg.content
		m.hints = msg.hints
		m.err = msg.err
		m.vp.SetContent(m.renderNote())
		return m, nil
	}
	return m, nil
}

func (m NoteView) View() string {
	if m.loading {
		return "Loading note..."
	}

	if m.err != nil {
		return "Error: " + m.err.Error()
	}

	return m.vp.View()
}

func fetchContent(path string) tea.Cmd {
	return func() tea.Msg {
		data, err := os.ReadFile(path)
		if err != nil {
			return loadedNote{err: err}
		}

		content := string(data)
		lines := strings.Split(content, "\n")

		var title string

		// title
		if len(lines) > 0 && strings.HasPrefix(lines[0], "# ") {
			title = lines[0]
			content = strings.Join(lines[1:], "\n")
		}

		// hints
		hints := extractHints(content)
		content = strings.TrimSpace(content)

		return loadedNote{
			title:   title,
			content: content,
			hints:   hints,
		}
	}
}

func extractHints(content string) []string {
	hints := make([]string, 0)
	return hints // to add later
}

func (m NoteView) renderNote() string {
	title, err1 := m.mdRenderer.Render(m.title)
	content, err2 := m.mdRenderer.Render(m.content)

	if err1 != nil || err2 != nil {
		// some error return raw
		return m.title + "\n\n" + m.content
	}

	return title + content
}
