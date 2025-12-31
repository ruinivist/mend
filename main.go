package main

import (
	"fmt"
	"mend/styles"
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
*/

type model struct {
	width int
	quit  bool
}

// initial model state
func createModel() model {
	return model{
		width: 20, // char count
	}
}

// =================== bubbletea ui fns ===================
// these need to be on the "model" ( duck typing "implements" interface )
func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quit = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	left := lipgloss.NewStyle().
		Width(m.width).
		Render("section1")

	right := lipgloss.NewStyle().
		Render("section2")

	divider := lipgloss.NewStyle().
		Width(1).
		Height(5).
		Background(styles.Primary).
		Render("") // everything is a string in tui

	return lipgloss.JoinHorizontal(lipgloss.Top, left, divider, right)
}

// =================== bubbletea ui fns ===================

func main() {
	p := tea.NewProgram(createModel(), tea.WithAltScreen() /* <- full screen tui */)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
