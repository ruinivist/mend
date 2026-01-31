/*
this pakckage basically handles notes which are the individual files for spaced
repetition in mend
*/

package note

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	glStyles "github.com/charmbracelet/glamour/styles"
	"github.com/charmbracelet/lipgloss"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

type Section struct {
	Title   string
	Content string
	Hints   []string
}

type ViewState int

const (
	StateTitleOnly ViewState = iota
	StateContent
	StateHints
)

type NoteView struct {
	// the actual content
	Path                string
	rawContent          string    // full content for editing
	sections            []Section // number of BLOCKS (heading separated)
	currentSectionIndex int
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
type LoadNoteMsg struct {
	Path  string
	Force bool
}

type LoadedNote struct {
	RawContent string
	Sections   []Section
	Err        error
}

func newMdRenderer() *glamour.TermRenderer {
	// styling in glamour can be better, I would rather have a fluent style api here
	// https://github.com/charmbracelet/glamour/issues/294
	mdStyleConfig := glStyles.TokyoNightStyleConfig
	var margin uint = 4
	mdStyleConfig.Document.Margin = &margin
	mdStyleConfig.Document.BlockPrefix = ""
	mdStyleConfig.Document.BlockSuffix = ""
	mdRenderer, _ := glamour.NewTermRenderer(
		glamour.WithStyles(mdStyleConfig),
		glamour.WithWordWrap(80),
	)
	return mdRenderer
}

func newTextArea() textarea.Model {
	ta := textarea.New()
	ta.Focus()
	ta.Prompt = ""
	ta.ShowLineNumbers = false
	return ta
}

func NewNoteView() *NoteView {
	return &NoteView{
		loading:    false,
		mdRenderer: newMdRenderer(),
		vp:         viewport.New(0, 0),
		viewState:  StateTitleOnly,
		textarea:   newTextArea(),
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
		m.vp.Height = msg.Height - 1
		m.textarea.SetWidth(msg.Width)
		m.textarea.SetHeight(msg.Height)
		return m, nil

	case tea.KeyMsg:
		if m.isEditing {
			switch msg.String() {
			case "esc", "ctrl+q":
				m.isEditing = false
				return m, saveContent(m.Path, m.textarea.Value())
			}
			var cmd tea.Cmd
			m.textarea, cmd = m.textarea.Update(msg)
			return m, cmd
		}

		switch msg.String() {
		case "enter":
			if m.Path != "" && !m.loading {
				m.isEditing = true
				m.textarea.SetValue(m.rawContent)
				m.textarea.Focus()
				return m, textarea.Blink
			}
		case " ":
			switch m.viewState {
			case StateTitleOnly:
				m.viewState = StateHints
			case StateHints:
				m.viewState = StateContent
			case StateContent:
				m.viewState = StateTitleOnly
			}
			m.vp.SetContent(m.renderNote())
		case "pgup":
			m.vp.PageUp()
			return m, nil
		case "pgdown":
			m.vp.PageDown()
			return m, nil
		case "left", "a":
			if m.currentSectionIndex > 0 {
				m.currentSectionIndex--
				m.vp.SetContent(m.renderNote())
				m.vp.GotoTop()
			}
			return m, nil
		case "right", "d":
			if m.currentSectionIndex < len(m.sections)-1 {
				m.currentSectionIndex++
				m.vp.SetContent(m.renderNote())
				m.vp.GotoTop()
			}
			return m, nil
		}
		var cmd tea.Cmd
		m.vp, cmd = m.vp.Update(msg)
		return m, cmd

	case LoadNoteMsg:
		if m.Path == msg.Path && !msg.Force {
			return m, nil //noop
		}
		m.Path = msg.Path
		m.isEditing = false
		m.loading = true
		m.currentSectionIndex = 0
		return m, fetchContent(msg.Path)

	case LoadedNote:
		m.loading = false
		m.rawContent = msg.RawContent
		m.sections = msg.Sections
		m.err = msg.Err
		m.currentSectionIndex = 0
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
		return "loading..."
	}
	if m.Path == "" {
		return ""
	}

	if m.err != nil {
		return "Error: " + m.err.Error()
	}

	if m.isEditing {
		return m.textarea.View()
	}

	page := m.currentSectionIndex + 1
	total := len(m.sections)
	var footer string
	if total == 0 {
		footer = "No sections"
	} else {
		footer = fmt.Sprintf("%d/%d", page, total)
	}
	footer = lipgloss.NewStyle().Width(m.vp.Width).Align(lipgloss.Right).Render(footer)

	return m.vp.View() + "\n" + footer
}

func fetchContent(path string) tea.Cmd {
	return func() tea.Msg {
		data, err := os.ReadFile(path)
		if err != nil {
			return LoadedNote{Err: err}
		}

		rawContent := string(data)
		sections := ParseSections(data)

		return LoadedNote{
			RawContent: rawContent,
			Sections:   sections,
		}
	}
}

func ParseSections(source []byte) []Section {
	md := goldmark.New()
	reader := text.NewReader(source)
	doc := md.Parser().Parse(reader)

	title := "no title"
	sections := make([]Section, 0)
	lastPos := 0

	for child := doc.FirstChild(); child != nil; child = child.NextSibling() {
		if child.Kind() == ast.KindHeading {
			if child.Lines().Len() <= 0 {
				continue
			}

			level := child.(*ast.Heading).Level
			headingStart := child.Lines().At(0).Start - level - 1
			headingEnd := child.Lines().At(child.Lines().Len() - 1).Stop

			contentEnd := headingStart
			if lastPos < contentEnd {
				// you have a heading and a content to accumulte over
				contentsRaw := source[lastPos:contentEnd]
				contents := strings.TrimSpace(string(contentsRaw))
				hints := ExtractHints(contents)
				sections = append(sections, Section{
					Title:   title,
					Content: contents,
					Hints:   hints,
				})
				lastPos = headingEnd
			}
			// for the next heading
			title = strings.TrimSpace(string(source[headingStart:headingEnd]))
			lastPos = headingEnd
		}
	}
	// last section
	if lastPos < len(source) {
		contentsRaw := source[lastPos:]
		contents := strings.TrimSpace(string(contentsRaw))
		hints := ExtractHints(contents)
		sections = append(sections, Section{
			Title:   title,
			Content: contents,
			Hints:   hints,
		})
	}

	return sections
}

func saveContent(path, content string) tea.Cmd {
	return func() tea.Msg {
		err := os.WriteFile(path, []byte(content), 0644)
		if err != nil {
			return LoadedNote{Err: err}
		}
		return fetchContent(path)()
	}
}

func ExtractHints(content string) []string {
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
	if m.Path == "" {
		return ""
	}

	section := Section{Content: m.rawContent} // default section if no sections are present
	if m.currentSectionIndex < len(m.sections) {
		section = m.sections[m.currentSectionIndex]
	}

	title, err := m.mdRenderer.Render(section.Title)
	if err != nil {
		title = section.Title + "\n\n"
	}

	var body string
	isListStart := false

	switch m.viewState {
	case StateTitleOnly:
		// no body
	case StateContent:
		body, err = m.mdRenderer.Render(section.Content)
		isListStart = strings.HasPrefix(section.Content, "-") || strings.HasPrefix(section.Content, "*")
	case StateHints:
		if len(section.Hints) == 0 {
			body, err = m.mdRenderer.Render("\nNo hints available.")
		} else {
			hintsList := ""
			for _, h := range section.Hints {
				hintsList += "- " + h + "\n"
			}
			body, err = m.mdRenderer.Render(hintsList)
			isListStart = true
		}
	}

	if err != nil {
		return title + "\nError rendering content."
	}

	if !isListStart {
		title += "\n"
	}

	return "\n\n" + title + body
}
