package main

// layout constants
const (
	minFsTreeWidth     = 5
	minNoteViewWidth   = 10
	dividerWidth       = 1
	dragHitArea        = 2 // +/- chars around divider
	fsTreeStartOffset  = 0 // there's bug when this is not 0 and the terminal height is small; I don't want to bother
	statusBarHeight    = 3
	fsTreeWidthPercent = 15
)

// calculateLayout computes the widths for the file tree and note view
// based on the terminal dimensions and the desired tree width.
func calculateLayout(totalWidth, requestedTreeWidth int) (treeWidth, noteWidth int) {
	treeWidth = requestedTreeWidth

	if treeWidth == 0 {
		treeWidth = totalWidth * fsTreeWidthPercent / 100
	}

	maxTreeWidth := totalWidth - dividerWidth - minNoteViewWidth

	// constraints
	if treeWidth > maxTreeWidth {
		treeWidth = maxTreeWidth
	}
	if treeWidth < minFsTreeWidth {
		treeWidth = minFsTreeWidth
	}

	noteWidth = max(0, totalWidth-treeWidth-dividerWidth)

	return treeWidth, noteWidth
}

// isHoveringDivider checks if the mouse cursor is within the interaction area of the divider
func isHoveringDivider(mouseX, dividerPos int) bool {
	return mouseX >= dividerPos-dragHitArea && mouseX <= dividerPos+dragHitArea
}
