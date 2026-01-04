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
	Primary        = lipgloss.Color("#7D56F4")
	Highlight      = lipgloss.Color("#89DDFF")
	HoverHighlight = lipgloss.Color("#91B4D5")
	// TODO: to extend as needed, right now this is just a
	// placeholder
)

// ==================== icons ====================
var (
	VerticalLine   = "│"
	ArrowDownIcon  = "⌄"
	ArrowRightIcon = "›"
)
