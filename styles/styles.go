/*
everything related to theming goes here
a theme is any set of colors or styles that are used in several
places or are expected to be globally consistent/modified
- ruinivist, 30Dec25
*/

package styles

import "github.com/charmbracelet/lipgloss"

// ==================== colors ====================
var (
	Primary    = lipgloss.Color("#7D56F4")
	Secondary  = lipgloss.Color("#FF5C8F")
	Background = lipgloss.Color("#1E1E2E")
	Highlight  = lipgloss.Color("#FFA500") // Orange highlight for selected node
	// TODO: to extend as needed, right now this is just a
	// placeholder
)

// ==================== lipgloss styles ====================
var (
	DividerStyle = lipgloss.NewStyle().
		Width(1).
		Background(Primary)
)

// ==================== icons ====================
var (
	VerticalLine   = "│"
	ArrowDownIcon  = "⌄"
	ArrowRightIcon = "›"
)
