package main

import (
	"fmt"
	fs "mend/fstree"
	"os"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
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
	width          int
	terminalWidth  int
	terminalHeight int
	tree           *fs.FsTree
	// spinner needs to be state as I need to update the spinner on
	// each tick in update func
	spinner  spinner.Model
	loading  bool
	viewport viewport.Model
}

// initial model state
func createModel() model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return model{
		width:          30, // char count
		terminalWidth:  0,
		terminalHeight: 0,
		tree:           nil,
		spinner:        s,
		loading:        true,
		viewport:       viewport.New(0, 0),
	}
}

// =================== bubbletea ui fns ===================
// these need to be on the "model" ( duck typing "implements" interface )

type treeLoadedMsg struct {
	tree *fs.FsTree
}

func loadTreeCmd() tea.Msg {

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting cwd:", err)
		os.Exit(1)
	}
	return treeLoadedMsg{tree: fs.NewFsTree(cwd)}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, loadTreeCmd)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.terminalWidth = msg.Width
		m.terminalHeight = msg.Height
		m.viewport.Width = msg.Width - m.width - 1
		m.viewport.Height = msg.Height
		return m, nil
	case treeLoadedMsg:
		m.tree = msg.tree
		m.loading = false
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case tea.MouseMsg:
		if m.tree != nil {
			// track hover for all mouse events within file tree
			if msg.X < m.width {
				m.tree.Update(fs.MsgHover{Line: msg.Y})
			} else {
				m.tree.Update(fs.MsgHover{Line: -1})
			}

			if msg.Button == tea.MouseButtonLeft {
				// within file tree
				if msg.X < m.width {
					m.tree.Update(fs.MsgSelectAtLine{Line: msg.Y})
					content, err := m.tree.GetSelectedContent()
					if err == nil {
						m.viewport.SetContent(content)
					}
				}
			}
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.tree != nil {
				m.tree.Update(fs.MsgMoveUp{})
			}
		case "down", "j":
			if m.tree != nil {
				m.tree.Update(fs.MsgMoveDown{})
			}
		case "enter", " ":
			if m.tree != nil {
				m.tree.Update(fs.MsgToggleExpand{})
			}
		case "l", "L":
			if m.tree != nil {
				content, err := m.tree.GetSelectedContent()
				if err != nil {
					fmt.Println("Error getting selected content:", err)
				} else {
					m.viewport.SetContent(content)
				}
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.loading {
		return fmt.Sprintf("%s Loading files...", m.spinner.View())
	}

	return m.tree.View()
}

// =================== bubbletea ui fns ===================

func main() {
	p := tea.NewProgram(
		createModel(),
		tea.WithAltScreen(), // full screen tui
		tea.WithMouseAllMotion(),
	)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
