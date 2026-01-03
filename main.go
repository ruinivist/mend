package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

/*
Quick notes to self:

# How bubbletea works?
Bubbletea is a
Init ( once for sync/async opts both )
-> Update ( model ) ( againc an be sync/async )
-> View loop
init creates initial model ( model is the global ui state and just a struct )
update modifies model based on an immediate or deferred compute
view is purely for rendering based on model

notes:
- model struct next -> global ui model
- createModel is the initial sync model state population
I anyways need to create the model struct even if blank
- Init func is for bubbletea to call once at start so both are needed
- all widths and heights are character count based
- async updates are via cmds (tea.Cmd ) that are no arg funcs that return
a tea.Msg ( that is basically a struct and hence has the data needed )
this data in msg when returned to the update func is used to update the model
*/

type model struct {
	width             int
	terminalWidth     int
	terminalHeight    int
	fsTreeWidth       int
	noteViewWidth     int
	tree              *FsTree
	rootPath          string // path to load the tree from
	loading           bool
	noteView          *NoteView
	isDragging        bool
	isHoveringDivider bool
	contentHeight     int
	showStatusBar     bool
}

func NewModel(rootPath string) *model {
	return &model{
		rootPath:      rootPath,
		loading:       true,
		noteView:      NewNoteView(),
		showStatusBar: true,
	}
}

// =================== bubbletea ui fns ===================
// these need to be on the "model" ( duck typing "implements" interface )

type treeLoadedMsg struct {
	tree *FsTree
}

func (m *model) loadTreeCmd(path string) tea.Cmd {
	return func() tea.Msg {
		var targetPath string
		if path == "" {
			cwd, err := os.Getwd()
			if err != nil {
				fmt.Println("Error getting cwd:", err)
				os.Exit(1)
			}
			targetPath = cwd
		} else {
			targetPath = path
		}
		return treeLoadedMsg{tree: NewFsTree(targetPath, fsTreeStartOffset)}
	}
}

func (m *model) layout(width, height int) {
	m.terminalWidth = width
	m.terminalHeight = height
	m.fsTreeWidth, m.noteViewWidth = calculateLayout(width, m.fsTreeWidth)

	h := height
	if m.showStatusBar {
		h -= statusBarHeight
	}
	m.contentHeight = max(0, h)
}

func (m *model) Init() tea.Cmd {
	return m.loadTreeCmd(m.rootPath)
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	// DEV TIP: ALWAYS RETURN IN EACH BRANCH
	// FALLTHROUGHS ARE BAD
	case tea.WindowSizeMsg:
		m.layout(msg.Width, msg.Height)

		var cmds []tea.Cmd
		if m.tree != nil {
			_, cmd := m.tree.Update(tea.WindowSizeMsg{
				Width:  m.fsTreeWidth,
				Height: m.contentHeight,
			})
			cmds = append(cmds, cmd)
		}
		_, cmd := m.noteView.Update(tea.WindowSizeMsg{
			Width:  m.noteViewWidth,
			Height: m.contentHeight,
		})
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)

	case treeLoadedMsg:
		m.tree = msg.tree
		m.loading = false
		_, cmd := m.tree.Update(tea.WindowSizeMsg{
			Height: m.contentHeight,
			Width:  m.fsTreeWidth,
		})
		return m, cmd

	case nodeSelected:
		// Forward node selection to noteView
		_, cmd := m.noteView.Update(loadNote{path: msg.path})
		return m, cmd

	case loadedNote:
		// Forward loaded note to noteView
		_, cmd := m.noteView.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "i":
			m.showStatusBar = !m.showStatusBar
			m.layout(m.terminalWidth, m.terminalHeight)

			var cmds []tea.Cmd
			if m.tree != nil {
				_, cmd := m.tree.Update(tea.WindowSizeMsg{
					Width:  m.fsTreeWidth,
					Height: m.contentHeight,
				})
				cmds = append(cmds, cmd)
			}
			_, cmd := m.noteView.Update(tea.WindowSizeMsg{
				Width:  m.noteViewWidth,
				Height: m.contentHeight,
			})
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		}

		var cmds []tea.Cmd

		// Forward keyboard input to tree
		if m.tree != nil {
			_, cmd := m.tree.Update(msg)
			cmds = append(cmds, cmd)
		}

		// Forward keyboard input to noteView
		_, cmd := m.noteView.Update(msg)
		cmds = append(cmds, cmd)

		return m, tea.Batch(cmds...)

	case tea.MouseMsg:
		if msg.Action == tea.MouseActionRelease {
			m.isDragging = false
		}

		if msg.Action == tea.MouseActionMotion {
			m.isHoveringDivider = isHoveringDivider(msg.X, m.fsTreeWidth)

			if m.isDragging {
				m.fsTreeWidth, m.noteViewWidth = calculateLayout(m.terminalWidth, msg.X)

				// Update children with new sizes
				var cmds []tea.Cmd
				if m.tree != nil {
					_, cmd := m.tree.Update(tea.WindowSizeMsg{
						Width:  m.fsTreeWidth,
						Height: m.contentHeight,
					})
					cmds = append(cmds, cmd)
				}
				_, cmd := m.noteView.Update(tea.WindowSizeMsg{
					Width:  m.noteViewWidth,
					Height: m.contentHeight,
				})
				cmds = append(cmds, cmd)
				return m, tea.Batch(cmds...)
			}
		}

		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			if isHoveringDivider(msg.X, m.fsTreeWidth) {
				m.isDragging = true
				return m, nil
			}
		}

		// Forward mouse input to tree only if not dragging
		var cmd tea.Cmd
		if m.tree != nil && !m.isDragging {
			_, cmd = m.tree.Update(msg)
		}
		return m, cmd
	}

	return m, nil
}

func (m model) View() string {
	if m.loading {
		return "Loading files..."
	}

	tree := m.tree.View()
	tree = lipgloss.NewStyle().
		Height(m.contentHeight).
		Width(m.fsTreeWidth).
		Align(lipgloss.Left).
		PaddingTop(fsTreeStartOffset).
		PaddingRight(1).
		Render(tree)

	var dividerChar string
	if m.isDragging || m.isHoveringDivider {
		dividerChar = "█"
	} else {
		dividerChar = "│"
	}

	// Repeat the character vertically to fill height
	dividerLines := make([]string, m.contentHeight)
	for i := range dividerLines {
		dividerLines[i] = dividerChar
	}
	divider := lipgloss.JoinVertical(lipgloss.Left, dividerLines...)

	// Ensure divider has the correct height style applied (though JoinVertical does most of it)
	divider = lipgloss.NewStyle().
		Height(m.contentHeight).
		Render(divider)

	notes := m.noteView.View()

	full := lipgloss.JoinHorizontal(
		lipgloss.Top,
		tree,
		divider,
		notes,
	)

	if !m.showStatusBar {
		return full
	}

	statusBar := lipgloss.NewStyle().
		Width(m.terminalWidth - 2). // Subtract borders
		Height(1).
		Border(lipgloss.NormalBorder()).
		Render("") // Placeholder text to ensure it's visible, can be empty string

	return lipgloss.JoinVertical(
		lipgloss.Left,
		full,
		statusBar,
	)
}

// =================== bubbletea ui fns ===================

func main() {
	var rootPath string
	if len(os.Args) > 1 {
		rootPath = os.Args[1]
	}
	// if rootPath is empty, createModel will use cwd
	p := tea.NewProgram(
		NewModel(rootPath),
		tea.WithAltScreen(), // full screen tui
		tea.WithMouseAllMotion(),
	)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
