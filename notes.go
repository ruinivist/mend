/*
this pakckage basically handles notes which are the individual files for spaced
repetition in mend
- ruinivist, 3Jan26
*/

package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
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
	path                string
	rawContent          string // full content for editing
	sections            []Section
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
type loadNote struct {
	path string
}

type loadedNote struct {
	rawContent string
	sections   []Section
	err        error
}

func NewNoteView() *NoteView {
	mdRenderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)
	ta := textarea.New()
	ta.Focus()
	ta.Prompt = ""
	ta.ShowLineNumbers = false

	return &NoteView{
		loading:    false,
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
		m.vp.Height = msg.Height - 1
		m.textarea.SetWidth(msg.Width)
		m.textarea.SetHeight(msg.Height)
		return m, nil

	case tea.KeyMsg:
		if m.isEditing {
			switch msg.String() {
			case "esc", "ctrl+q":
				m.isEditing = false
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

	case loadNote:
		if m.path == msg.path {
			return m, nil //noop
		}
		m.path = msg.path
		m.isEditing = false
		m.loading = true
		m.currentSectionIndex = 0
		return m, fetchContent(msg.path)

	case loadedNote:
		m.loading = false
		m.rawContent = msg.rawContent
		m.sections = msg.sections
		m.err = msg.err
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
	if m.path == "" {
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
			return loadedNote{err: err}
		}

		rawContent := string(data)
		sections := parseSections(data)

		return loadedNote{
			rawContent: rawContent,
			sections:   sections,
		}
	}
}

func parseSections(source []byte) []Section {
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
				hints := extractHints(contents)
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
		hints := extractHints(contents)
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
	if m.path == "" {
		return "" // no note is loaded, don't need to bother with anything
	}
	var titleText, contentText string
	if len(m.sections) > 0 {
		titleText = m.sections[m.currentSectionIndex].Title
		contentText = m.sections[m.currentSectionIndex].Content
	} else {
		contentText = m.rawContent
	}

	title, err1 := m.mdRenderer.Render(titleText)
	if err1 != nil {
		title = titleText + "\n\n"
	}

	var body string
	var err2 error

	switch m.viewState {
	case StateContent:
		body, err2 = m.mdRenderer.Render(contentText)
	case StateHints:
		currentHints := []string{}
		if m.currentSectionIndex < len(m.sections) {
			currentHints = m.sections[m.currentSectionIndex].Hints
		}

		if len(currentHints) == 0 {
			body = "No hints available."
		} else {
			// Format hints as a list
			hintsList := ""
			for _, h := range currentHints {
				hintsList += "- " + h + "\n"
			}
			body, err2 = m.mdRenderer.Render(hintsList)
		}
	case StateTitleOnly:
		body = "" // TOOD: there was a message here but I removed it, maybe remove enum entry as well
		// or improve this part ui
	}

	if err2 != nil {
		return title + "\nError rendering content."
	}

	return title + body
}
