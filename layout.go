package main

// Layout constants to avoid magic numbers
const (
	minFsTreeWidth   = 5
	minNoteViewWidth = 10
	dividerWidth     = 1
	dragHitArea      = 2 // +/- chars around divider
)

// calculateLayout computes the widths for the file tree and note view
// based on the terminal dimensions and the desired tree width.
// It returns the constrained tree width and the resulting note view width.
func calculateLayout(totalWidth, requestedTreeWidth int) (treeWidth, noteWidth int) {
	treeWidth = requestedTreeWidth
	
	// Default initial width if not set
	if treeWidth == 0 {
		treeWidth = totalWidth / 4
	}

	// Calculate maximum allowed width for the tree
	// Total width - divider - minimum note view width
	maxTreeWidth := totalWidth - dividerWidth - minNoteViewWidth

	// Apply constraints
	if treeWidth > maxTreeWidth {
		treeWidth = maxTreeWidth
	}
	if treeWidth < minFsTreeWidth {
		treeWidth = minFsTreeWidth
	}

	noteWidth = totalWidth - treeWidth - dividerWidth
	
	// Safety check to ensure noteWidth is never negative
	if noteWidth < 0 {
		noteWidth = 0
	}

	return treeWidth, noteWidth
}

// isHoveringDivider checks if the mouse cursor is within the interaction area of the divider
func isHoveringDivider(mouseX, dividerPos int) bool {
	return mouseX >= dividerPos-dragHitArea && mouseX <= dividerPos+dragHitArea
}
