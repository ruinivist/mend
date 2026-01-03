/*
this pakckage basically handles notes which are the individual files for spaced
repetition in mend
- ruinivist, 3Jan26
*/

package notes

import (
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	path    string
	title   string   // the first line is always # title
	content string   // strip the first line from file rest is content
	hints   []string //a hint is anything that matches (hint: <text>) in content
	err     error
	loading bool
}

// ================== messages ===================
type contentFetchedMsg struct {
	title   string
	content string
	hints   []string
	err     error
}

func NewNote(path string) Model {
	return Model{
		path:    path,
		loading: true,
	}
}

func (m Model) Init() tea.Cmd {
	return fetchContent(m.path)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case contentFetchedMsg:
		m.loading = false
		m.title = msg.title
		m.content = msg.content
		m.hints = msg.hints
		m.err = msg.err
		return m, nil
	}
	return m, nil
}

func (m Model) View() string {
	if m.loading {
		return "Loading note..."
	}

	if m.err != nil {
		return "Error: " + m.err.Error()
	}

	var sb strings.Builder
	sb.WriteString(m.title)
	sb.WriteString("\n\n")
	sb.WriteString(m.content)

	if len(m.hints) > 0 {
		sb.WriteString("\n\nHints:\n")
		for _, hint := range m.hints {
			sb.WriteString("- ")
			sb.WriteString(hint)
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

func fetchContent(path string) tea.Cmd {
	return func() tea.Msg {
		data, err := os.ReadFile(path)
		if err != nil {
			return contentFetchedMsg{err: err}
		}

		content := string(data)
		lines := strings.Split(content, "\n")

		var title string

		// tit;e
		if len(lines) > 0 && strings.HasPrefix(lines[0], "# ") {
			title = strings.TrimPrefix(lines[0], "# ")
			content = strings.Join(lines[1:], "\n")
		}

		// hints
		hints := extractHints(content)
		content = strings.TrimSpace(content)

		return contentFetchedMsg{
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
