/*
search ui model
*/
package search

import (
	"strings"

	"mend/internal/search"
	"mend/styles"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SearchView struct {
	input         textinput.Model
	engine        *search.SearchEngine
	results       []search.SearchResult
	selectedIndex int
	width         int
	height        int
	active        bool
}

func NewSearchView(engine *search.SearchEngine) *SearchView {
	ti := textinput.New()
	ti.Placeholder = "Search files..."
	ti.CharLimit = 256
	ti.Width = 50 // TODO: make this dynamic, Note that there is a context len on the engine as well
	ti.Focus()

	return &SearchView{
		input:         ti,
		engine:        engine,
		results:       make([]search.SearchResult, 0),
		selectedIndex: 0,
	}
}

// SearchSelectMsg is sent when user selects a search result
type SearchSelectMsg struct {
	Path     string
	IsFolder bool
}

// SearchCancelMsg is sent when user cancels search
type SearchCancelMsg struct{}

func (v *SearchView) Init() tea.Cmd {
	return textinput.Blink
}

func (v *SearchView) IsActive() bool {
	return v.active
}

func (v *SearchView) Activate() tea.Cmd {
	v.active = true
	v.input.SetValue("")
	v.results = nil
	v.selectedIndex = 0
	v.input.Focus()
	return textinput.Blink
}

func (v *SearchView) Deactivate() {
	v.active = false
	v.input.Blur()
}

func (v *SearchView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height
		v.input.Width = msg.Width - 4
		return v, nil

	case tea.KeyMsg:
		// operational events
		switch msg.String() {
		case "esc", "q", "ctrl+c", "ctrl+d", "ctrl+q":
			v.Deactivate()
			return v, func() tea.Msg { return SearchCancelMsg{} }

		case "enter":
			if len(v.results) > 0 && v.selectedIndex < len(v.results) {
				result := v.results[v.selectedIndex]
				v.Deactivate()
				return v, func() tea.Msg {
					return SearchSelectMsg{Path: result.Path, IsFolder: result.IsFolder}
				}
			}
			return v, nil

		case "up":
			if v.selectedIndex > 0 {
				v.selectedIndex--
			}
			return v, nil

		case "down":
			if v.selectedIndex < len(v.results)-1 {
				v.selectedIndex++
			}
			return v, nil
		}

		// this is now for text input
		var cmd tea.Cmd
		oldValue := v.input.Value()
		v.input, cmd = v.input.Update(msg)

		// If query changed, re-search
		if v.input.Value() != oldValue {
			v.results = v.engine.Search(v.input.Value())
			v.selectedIndex = 0
		}

		return v, cmd
	}

	return v, nil
}

func (v *SearchView) View() string {
	if !v.active {
		return ""
	}

	if v.engine.IsIndexing() {
		return " Still indexing..."
	}

	noResultsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true).
		Padding(0, 2)

	if v.input.Value() == "" {
		return noResultsStyle.Render("Type a query to search")
	}

	var b strings.Builder

	// Header with input
	inputStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Width(v.width - 4)

	b.WriteString(inputStyle.Render(v.input.View()))
	b.WriteString("\n")

	// Results
	if len(v.results) == 0 {
		b.WriteString(noResultsStyle.Render("No results found"))
	} else {
		maxVisible := max(1, v.height-6) // 6 is kinda arbitrary, I just do a good enough offset for text box

		// Calculate visible range, this handling next is bit buggy
		// also TODO: i don't highlight if there are more results at the end
		startIdx := 0
		if v.selectedIndex >= maxVisible {
			startIdx = v.selectedIndex - maxVisible + 1
		}
		endIdx := startIdx + maxVisible
		if endIdx > len(v.results) {
			endIdx = len(v.results)
		}

		for i := startIdx; i < endIdx; i++ {
			result := v.results[i]
			isSelected := i == v.selectedIndex

			// Result line style
			lineStyle := lipgloss.NewStyle().Padding(0, 2)
			if isSelected {
				lineStyle = lineStyle.
					Background(lipgloss.Color("237")). // TODO: move to styles, I need to unify stules too
					Foreground(styles.Highlight).
					Bold(true)
			}

			// Build the display line with icon
			var icon string
			var path string
			if result.IsFolder {
				if isSelected {
					icon = styles.FolderIcon
				} else {
					icon = lipgloss.NewStyle().Foreground(styles.FolderBlue).Render(styles.FolderIcon)
				}
				path = result.RelativePath + "/"
			} else {
				if isSelected {
					icon = styles.FileIcon
				} else {
					icon = lipgloss.NewStyle().Foreground(styles.FileGreen).Render(styles.FileIcon)
				}
				path = result.RelativePath
				if result.Snippet != "" && !isSelected {
					snippetStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
					path += snippetStyle.Render("   " + result.Snippet)
				} else if result.Snippet != "" {
					path += "   " + result.Snippet
				}
			}

			line := icon + " " + path
			b.WriteString(lineStyle.Render(line))
			b.WriteString("\n")
		}
	}

	// Full screen container
	containerStyle := lipgloss.NewStyle().
		Width(v.width).
		Height(v.height).
		Padding(1, 2)

	return containerStyle.Render(b.String())
}
