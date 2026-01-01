package main

import (
	"fmt"
	"mend/compositor"
	fs "mend/fstree"
	"mend/styles"
	"os"
	"time"

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
	spinner    spinner.Model
	loading    bool
	viewport   viewport.Model
	hoverLine  int // track which line is being hovered
	showDialog bool
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
		hoverLine:      -1, // no hover initially
		showDialog:     false,
	}
}

// =================== bubbletea ui fns ===================
// these need to be on the "model" ( duck typing "implements" interface )

type treeLoadedMsg struct {
	tree *fs.FsTree
}

func loadTreeCmd() tea.Msg {
	time.Sleep(1 * time.Second) // delay sim

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
		// track hover for all mouse events within file tree
		if msg.X < m.width {
			m.hoverLine = msg.Y
		} else {
			m.hoverLine = -1
		}

		if msg.Button == tea.MouseButtonLeft {
			// within file tree
			if msg.X < m.width {
				if err := m.tree.SelectNodeAtLine(msg.Y); err == nil {
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
		case "d", "D":
			m.showDialog = !m.showDialog
			return m, nil
		}

		switch msg.String() {
		case "up", "k":
			if err := m.tree.MoveUp(); err != nil {
				fmt.Println("Error moving up:", err)
			}
		case "down", "j":
			if err := m.tree.MoveDown(); err != nil {
				fmt.Println("Error moving down:", err)
			}
		case "enter", " ":
			if err := m.tree.ToggleSelectedExpand(); err != nil {
				fmt.Println("Error toggling expand:", err)
			}
		case "l", "L":
			content, err := m.tree.GetSelectedContent()
			if err != nil {
				fmt.Println("Error getting selected content:", err)
			} else {
				m.viewport.SetContent(content)
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	var baseUI string
	if m.loading {
		baseUI = fmt.Sprintf("%s Loading files...", m.spinner.View())
	} else {
		left := lipgloss.NewStyle().
			Width(m.width).
			Render(m.tree.Render(m.hoverLine))

		right := lipgloss.NewStyle().
			Width(m.terminalWidth - m.width - 1).
			Height(m.terminalHeight).
			Render(m.viewport.View())

		divider := lipgloss.NewStyle().
			Width(1).
			Height(m.terminalHeight).
			Background(styles.Primary).
			Render("")

		baseUI = lipgloss.JoinHorizontal(lipgloss.Top, left, divider, right)
	}

	// no compositing needed, optimised return
	if !m.showDialog {
		return baseUI
	}

	// manual compositing
	grid := compositor.NewGrid(m.terminalWidth, m.terminalHeight)
	grid.Write(0, 0, baseUI)
	dialog := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Primary).
		Background(lipgloss.Color("#222")).
		Foreground(lipgloss.Color("#FFF")).
		Padding(1, 2).
		Align(lipgloss.Center).
		Render("this is a sample dialog\npress d to close")

	grid.Write(2, 2, dialog)
	return grid.Render()
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
