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
	FolderBlue     = lipgloss.Color("#5FAFFF") // Blue for folder names in search
	FileGreen      = lipgloss.Color("#98C379") // Green for file icons
)

// ==================== icons (nerd fonts) ====================
var (
	VerticalLine   = "│"
	ArrowDownIcon  = "⌄"
	ArrowRightIcon = "›"
	// need nerd fonts to render correctly, how I got them? https://fontawesome.com/v4/icon/folder has a unicode
	FolderIcon = "\uf07b"
	FileIcon   = "\uf0f6"
)
