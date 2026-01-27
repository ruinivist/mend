package main

import (
	"fmt"
	"os"
	"os/exec"

	"mend/internal/ui/fstree"
	"mend/internal/ui/note"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

/*
Quick notes to self:
... (comments preserved)
*/

type model struct {
	width             int
	terminalWidth     int
	terminalHeight    int
	fsTreeWidth       int
	noteViewWidth     int
	tree              *fstree.FsTree
	rootPath          string // path to load the tree from
	loading           bool
	noteView          *note.NoteView
	isDragging        bool
	isHoveringDivider bool
	contentHeight     int
	showStatusBar     bool
	showSidebar       bool
	// input handling
	textInput     textinput.Model
	inputMode     bool
	pendingAction fstree.FsActionType
}

func NewModel(rootPath string) *model {
	ti := textinput.New()
	ti.CharLimit = 156
	ti.Width = 30

	return &model{
		rootPath:      rootPath,
		loading:       true,
		noteView:      note.NewNoteView(),
		showStatusBar: false,
		showSidebar:   true,
		textInput:     ti,
	}
}

// =================== bubbletea ui fns ===================
// these need to be on the "model" ( duck typing "implements" interface )

type treeLoadedMsg struct {
	tree *fstree.FsTree
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
		return treeLoadedMsg{tree: fstree.NewFsTree(targetPath, fsTreeStartOffset)}
	}
}

func (m *model) layout(width, height int) {
	m.terminalWidth = width
	m.terminalHeight = height
	minW := 0
	if m.tree != nil {
		minW = m.tree.ContentWidth()
	}
	if !m.showSidebar {
		m.fsTreeWidth = 0
		m.noteViewWidth = width
	} else {
		m.fsTreeWidth, m.noteViewWidth = getUpdatedWindowSizes(width, m.fsTreeWidth, minW)
	}

	h := height
	if m.showStatusBar || m.inputMode {
		h -= statusBarHeight
	}
	m.contentHeight = max(0, h)
}

func (m *model) resizeChildren() tea.Cmd {
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
	return tea.Batch(cmds...)
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
		return m, m.resizeChildren()

	case treeLoadedMsg:
		m.tree = msg.tree
		m.loading = false
		m.fsTreeWidth = m.tree.ContentWidth()
		m.fsTreeWidth, m.noteViewWidth = getUpdatedWindowSizes(m.terminalWidth, m.fsTreeWidth, m.tree.ContentWidth())
		_, cmd := m.tree.Update(tea.WindowSizeMsg{
			Height: m.contentHeight,
			Width:  m.fsTreeWidth,
		})
		return m, cmd

	case fstree.NodeSelectedMsg:
		// Forward node selection to noteView
		_, cmd := m.noteView.Update(note.LoadNoteMsg{Path: msg.Path})
		return m, cmd

	case note.LoadNoteMsg:
		_, cmd := m.noteView.Update(msg)
		if msg.Force {
			return m, tea.Batch(cmd, tea.EnableMouseAllMotion)
		}
		return m, cmd

	case note.LoadedNote:
		// Forward loaded note to noteView
		_, cmd := m.noteView.Update(msg)
		return m, cmd

	case fstree.PerformActionMsg:
		if m.tree != nil {
			_, cmd := m.tree.Update(msg)
			return m, cmd
		}

	case fstree.ContentSizeChangeMsg:
		// layout update needed, sent when a new note is created
		m.fsTreeWidth, m.noteViewWidth = getUpdatedWindowSizes(m.terminalWidth, m.tree.ContentWidth(), m.tree.ContentWidth())
		// TODO: this can be simplified a lot
		return m, m.resizeChildren() // batches two updates

	case fstree.RequestInputMsg:
		m.inputMode = true
		m.pendingAction = msg.Action
		m.textInput.Focus()
		m.textInput.SetValue("")
		switch msg.Action {
		case fstree.ActionNewFile:
			m.textInput.Placeholder = "New File Name"
		case fstree.ActionNewFolder:
			m.textInput.Placeholder = "New Folder Name"
		case fstree.ActionNewRoot:
			m.textInput.Placeholder = "New Root Folder Name"
		}
		m.layout(m.terminalWidth, m.terminalHeight) // recalc layout for status bar area
		return m, m.resizeChildren()

	case tea.KeyMsg:
		if m.inputMode {
			switch msg.String() {
			case "enter":
				val := m.textInput.Value()
				m.inputMode = false
				m.textInput.Blur()
				m.layout(m.terminalWidth, m.terminalHeight)

				cmds := []tea.Cmd{m.resizeChildren()}
				if val != "" {
					cmds = append(cmds, func() tea.Msg {
						return fstree.PerformActionMsg{
							Action: m.pendingAction,
							Name:   val,
						}
					})
				}
				return m, tea.Batch(cmds...)
			case "esc":
				m.inputMode = false
				m.textInput.Blur()
				m.layout(m.terminalWidth, m.terminalHeight)
				return m, m.resizeChildren()
			}
			var cmd tea.Cmd
			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd
		}

		// If editing, forward all keys to noteView and ignore global bindings
		if m.noteView.IsEditing() {
			_, cmd := m.noteView.Update(msg)
			return m, cmd
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "ctrl+b":
			m.showSidebar = !m.showSidebar
			m.layout(m.terminalWidth, m.terminalHeight)
			return m, m.resizeChildren()
		case "i":
			m.showStatusBar = !m.showStatusBar
			m.layout(m.terminalWidth, m.terminalHeight)
			return m, m.resizeChildren()
		case "o":
			if m.tree != nil && m.tree.SelectedNode != nil && m.tree.SelectedNode.Type == fstree.FileNode {
				c := exec.Command("micro", m.tree.SelectedNode.Path)
				c.Stdin = os.Stdin
				c.Stdout = os.Stdout
				c.Stderr = os.Stderr
				return m, tea.ExecProcess(c, func(err error) tea.Msg {
					// mouse loses focus so this is neededd
					return note.LoadNoteMsg{Path: m.tree.SelectedNode.Path, Force: true}
				})
			}
		case "delete":
			// Forward delete to fstree if focused (implied focus on tree for now when not editing)
			if m.tree != nil {
				_, cmd := m.tree.Update(msg)
				return m, cmd
			}
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
				m.fsTreeWidth, m.noteViewWidth = getUpdatedWindowSizes(m.terminalWidth, msg.X, m.tree.ContentWidth())

				// Update children with new sizes
				return m, m.resizeChildren()
			}
		}

		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			if isHoveringDivider(msg.X, m.fsTreeWidth) {
				m.isDragging = true
				return m, nil
			}
		}

		// Forward mouse input to children if not dragging
		var cmds []tea.Cmd
		if !m.isDragging {
			if m.tree != nil && msg.X < m.fsTreeWidth {
				_, cmd := m.tree.Update(msg)
				cmds = append(cmds, cmd)
			}
			if msg.X > m.fsTreeWidth {
				_, cmd := m.noteView.Update(msg)
				cmds = append(cmds, cmd)
			}
		}
		return m, tea.Batch(cmds...)
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

	var full string
	if m.showSidebar {
		full = lipgloss.JoinHorizontal(
			lipgloss.Top,
			tree,
			divider,
			notes,
		)
	} else {
		full = notes
	}

	if !m.showStatusBar && !m.inputMode {
		return full
	}

	statusContent := ""
	if m.inputMode {
		statusContent = m.textInput.View()
	} else if m.tree != nil && m.tree.ErrMsg != "" {
		statusContent = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(m.tree.ErrMsg)
	}

	statusBar := lipgloss.NewStyle().
		Width(m.terminalWidth - 2). // Subtract borders
		Height(1).
		Border(lipgloss.NormalBorder()).
		Render(statusContent)

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
