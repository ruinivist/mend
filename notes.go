/*
this pakckage basically handles notes which are the individual files for spaced
repetition in mend
- ruinivist, 3Jan26
*/

package main

import (
	"os"
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
)

type ViewState int

const (
	StateTitleOnly ViewState = iota
	StateContent
	StateHints
)

type NoteView struct {
	// the actual content
	path       string
	title      string   // the first line is always # title
	content    string   // strip the first line from file rest is content
	rawContent string   // full content for editing
	hints      []string //a hint is anything that matches **text** or __text__ in content
	// display layer
	err        error
	loading    bool
	vp         viewport.Model
	mdRenderer *glamour.TermRenderer
	viewState  ViewState
	// editing
	textarea  textarea.Model
	isEditing bool
}

// ================== messages ===================
type loadNote struct {
	path string
}

type loadedNote struct {
	title      string
	content    string
	rawContent string
	hints      []string
	err        error
}

func NewNoteView() *NoteView {
	mdRenderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)
	ta := textarea.New()
	ta.Focus()
	return &NoteView{
		loading:    true,
		mdRenderer: mdRenderer,
		vp:         viewport.New(0, 0),
		viewState:  StateTitleOnly,
		textarea:   ta,
	}
}

func (m *NoteView) Init() tea.Cmd {
	return textarea.Blink
}

func (m *NoteView) IsEditing() bool {
	return m.isEditing
}

func (m *NoteView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.vp.Width = msg.Width
		m.vp.Height = msg.Height
		m.textarea.SetWidth(msg.Width)
		m.textarea.SetHeight(msg.Height)
		return m, nil

	case tea.KeyMsg:
		if m.isEditing {
			switch msg.String() {
			case "ctrl+w":
				m.isEditing = false
				return m, nil
			case "ctrl+s":
				return m, saveContent(m.path, m.textarea.Value())
			}
			var cmd tea.Cmd
			m.textarea, cmd = m.textarea.Update(msg)
			return m, cmd
		}

		switch msg.String() {
		case "enter":
			if m.path != "" && !m.loading {
				m.isEditing = true
				m.textarea.SetValue(m.rawContent)
				m.textarea.Focus()
				return m, textarea.Blink
			}
		case "d", "right":
			m.viewState = StateContent
			m.vp.SetContent(m.renderNote())
		case "a", "left":
			m.viewState = StateHints
			m.vp.SetContent(m.renderNote())
		case "pgup":
			m.vp.PageUp()
			return m, nil
		case "pgdown":
			m.vp.PageDown()
			return m, nil
		}
		var cmd tea.Cmd
		m.vp, cmd = m.vp.Update(msg)
		return m, cmd

	case loadNote:
		if m.path == msg.path {
			return m, nil //noop
		}
		m.path = msg.path
		m.isEditing = false
		m.loading = true
		return m, fetchContent(msg.path)

	case loadedNote:
		m.loading = false
		m.title = msg.title
		m.content = msg.content
		m.rawContent = msg.rawContent
		m.hints = msg.hints
		m.err = msg.err
		m.viewState = StateTitleOnly
		m.vp.SetContent(m.renderNote())

		return m, nil

	case tea.MouseMsg:
		if m.isEditing {
			return m, nil // or handle mouse in textarea if supported
		}
		var cmd tea.Cmd
		m.vp, cmd = m.vp.Update(msg)
		return m, cmd
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

	if m.isEditing {
		return m.textarea.View()
	}

	return m.vp.View()
}

func fetchContent(path string) tea.Cmd {
	return func() tea.Msg {
		data, err := os.ReadFile(path)
		if err != nil {
			return loadedNote{err: err}
		}

		rawContent := string(data)
		content := rawContent
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
			title:      title,
			content:    content,
			rawContent: rawContent,
			hints:      hints,
		}
	}
}

func saveContent(path, content string) tea.Cmd {
	return func() tea.Msg {
		err := os.WriteFile(path, []byte(content), 0644)
		if err != nil {
			return loadedNote{err: err}
		}
		return fetchContent(path)()
	}
}

func extractHints(content string) []string {
	re := regexp.MustCompile(`(?s)\*\*(.*?)\*\*|__(.*?)__`)
	matches := re.FindAllStringSubmatch(content, -1)
	hints := make([]string, 0)
	for _, match := range matches {
		if len(match) > 1 && match[1] != "" {
			hints = append(hints, match[1])
		} else if len(match) > 2 && match[2] != "" {
			hints = append(hints, match[2])
		}
	}
	return hints
}

func (m NoteView) renderNote() string {
	title, err1 := m.mdRenderer.Render(m.title)
	if err1 != nil {
		title = m.title + "\n\n"
	}

	var body string
	var err2 error

	switch m.viewState {
	case StateContent:
		body, err2 = m.mdRenderer.Render(m.content)
	case StateHints:
		if len(m.hints) == 0 {
			body = "No hints available."
		} else {
			// Format hints as a list
			hintsList := ""
			for _, h := range m.hints {
				hintsList += "- " + h + "\n"
			}
			body, err2 = m.mdRenderer.Render(hintsList)
		}
	case StateTitleOnly:
		body = "'a' to view hints, 'd' to view content."
	}

	if err2 != nil {
		return title + "\nError rendering content."
	}

	return title + body
}
