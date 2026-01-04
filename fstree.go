/*
This file contains the implementation of an in-memory tree of files and folders that the app used.
It is initialised at startup and used for state changes in UI and later persisted to disk.

- ruinivist, 30Dec25
*/

package main

import (
	"errors"
	"mend/styles"
	"mend/utils"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// enums for node types
type FsNodeType int

const (
	FileNode FsNodeType = iota
	FolderNode
)

// ==================== FsNode definition ====================
// a single node, Fs => deals with file system related info mostly
// a name for example can be content derived as well
type FsNode struct {
	nodeType FsNodeType
	path     string
	children []*FsNode
	parent   *FsNode // for fast traversal up the tree
	expanded bool    // makes sense only for folder nodes
	// these are populated by buildLines for fast access
	line int
	prev *FsNode
	next *FsNode
}

func (n *FsNode) FileName() string {
	name := filepath.Base(n.path)
	if n.nodeType == FileNode {
		return strings.TrimSuffix(name, ".md")
	}
	return name
}

// ================== messages ===================
type nodeSelected struct {
	path string
}

type FsActionType int

const (
	ActionNewFile FsActionType = iota
	ActionNewFolder
	ActionNewRoot
)

type RequestInputMsg struct {
	Action FsActionType
}

type PerformActionMsg struct {
	Action FsActionType
	Name   string
}

// ==================== FsNode definition ====================
type FsTree struct {
	root         *FsNode
	lines        map[int]*FsNode // flattened view of nodes for easy line access, map so that I can handle blank padding
	selectedNode *FsNode
	hoveredNode  *FsNode
	errMsg       string
	height       int
	width        int
	viewStart    int
	viewEnd      int
	totalLines   int
	oldSelected  *FsNode
	startOffset  int
}

// ==================== Bubble Tea Interface Implementation ====================
func (t *FsTree) Init() tea.Cmd {
	return nil
}

func (t *FsTree) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case PerformActionMsg:
		err := t.PerformAction(m.Action, m.Name)
		if err != nil {
			t.errMsg = err.Error()
		}
	case tea.WindowSizeMsg:
		t.width = m.Width
		t.height = m.Height
	case tea.KeyMsg:
		t.errMsg = ""
		switch m.String() {
		case "w", "up":
			_ = t.MoveUp()
		case "s", "down":
			_ = t.MoveDown()
		case "e", "space":
			_ = t.ToggleSelectedExpand()
		case "n": // new file
			return t, func() tea.Msg { return RequestInputMsg{Action: ActionNewFile} }
		case "N": // new folder
			return t, func() tea.Msg { return RequestInputMsg{Action: ActionNewFolder} }
		case "C": // new root node
			return t, func() tea.Msg { return RequestInputMsg{Action: ActionNewRoot} }
		case "delete": // delete node
			err := t.DeleteNode(t.selectedNode)
			if err != nil {
				t.errMsg = err.Error()
			}
		}

	case tea.MouseMsg:
		if m.Action == tea.MouseActionPress {
			switch m.Button {
			case tea.MouseButtonWheelUp:
				_ = t.MoveUp()
			case tea.MouseButtonWheelDown:
				_ = t.MoveDown()
			}
		}

		m.Y += t.viewStart - t.startOffset // adjust for viewport
		if m.X >= t.width {
			break
		}
		t.errMsg = ""
		t.hoveredNode = t.lines[m.Y] // hover

		// click
		if m.Button == tea.MouseButtonLeft && m.Action == tea.MouseActionPress {
			nodeAtLine := t.lines[m.Y]
			if nodeAtLine != nil {
				t.selectedNode = nodeAtLine
				if nodeAtLine.nodeType == FolderNode {
					_ = t.ToggleExpand(nodeAtLine)
				}
			}
		}
	}

	t.viewStart, t.viewEnd = t.getViewportBounds()

	if t.oldSelected != t.selectedNode && t.selectedNode.nodeType == FileNode {
		t.oldSelected = t.selectedNode
		return t, func() tea.Msg {
			return nodeSelected{path: t.selectedNode.path}
		}
	}
	return t, nil
}

func (t *FsTree) PerformAction(action FsActionType, name string) error {
	switch action {
	case ActionNewFile:
		return t.CreateNode(t.selectedNode, name, FileNode)
	case ActionNewFolder:
		return t.CreateNode(t.selectedNode, name, FolderNode)
	case ActionNewRoot:
		return t.CreateNode(t.root, name, FolderNode)
	}
	return nil
}

func (t *FsTree) getViewportBounds() (startLine, endLine int) {
	if t.selectedNode == nil {
		return 0, 0 // doesn't amtter in this case
	}
	selectedLine := t.selectedNode.line
	halfHeight := t.height / 2

	startLine = max(0, selectedLine-halfHeight)

	endLine = startLine + t.height
	if endLine > t.totalLines {
		endLine = t.totalLines
		startLine = endLine - t.height
		if startLine < 0 {
			startLine = 0
		}
	}

	return startLine, endLine
}

func (t *FsTree) View() string {
	if t.errMsg != "" {
		return t.errMsg
	}

	if len(t.root.children) == 0 {
		return "no files/folders\nPress C to create"
	}

	builder := &strings.Builder{}
	t.renderNode(t.root, 0, builder)
	rendered := builder.String()

	lines := strings.Split(rendered, "\n")

	clampedLines := lines[t.viewStart:t.viewEnd]
	rendered = strings.Join(clampedLines, "\n")

	rendered = lipgloss.NewStyle().
		Width(t.width).
		Render(rendered)

	return rendered
}

// ==================== FsTree helper methods ====================

func NewFsTree(rootPath string, startOffset int) *FsTree {
	root := &FsNode{
		nodeType: FolderNode,
		path:     rootPath,
		children: make([]*FsNode, 0),
		expanded: true,
	}
	walkFileSystemAndBuildTree(rootPath, root)

	tree := &FsTree{
		root:        root,
		startOffset: startOffset,
	}
	if len(root.children) > 0 {
		tree.selectedNode = root.children[0]
	}
	tree.buildLines()
	return tree
}

func (t *FsTree) DeleteNode(node *FsNode) error {
	if node == nil {
		return errors.New("node to delete cannot be nil")
	}
	parent := node.parent
	if parent == nil {
		return errors.New("node to delete must have a parent")
	}

	// materialise
	if err := deletePath(node.path); err != nil {
		return err
	}

	t.selectedNode = node.prev // cannot be next as subfolder/file deletion
	parent.children = utils.RemoveFromSlice(parent.children, node)

	if t.selectedNode == nil {
		// can only happen if first root level node is deleted
		if len(t.root.children) > 0 {
			t.selectedNode = t.root.children[0]
		}
	}

	t.buildLines()
	return nil
}

func (t *FsTree) ToggleExpand(node *FsNode) error {
	if node == nil {
		return errors.New("node cannot be nil")
	}
	if node.nodeType != FolderNode {
		return errors.New("only folder nodes can be expanded or collapsed")
	}

	node.expanded = !node.expanded
	t.buildLines()
	return nil
}

func (t *FsTree) move(delta int) error {
	selected := t.selectedNode
	if selected == nil {
		return errors.New("no node is currently selected")
	}
	if delta != -1 && delta != 1 {
		return errors.New("delta must be either -1 (up) or 1 (down)")
	}

	var next *FsNode
	if delta == 1 {
		next = selected.next
	} else {
		next = selected.prev
	}
	if next != nil {
		t.selectedNode = next
	}
	return nil
}

func (t *FsTree) MoveUp() error   { return t.move(-1) }
func (t *FsTree) MoveDown() error { return t.move(1) }

func (t *FsTree) ToggleSelectedExpand() error {
	if t.selectedNode == nil {
		return errors.New("no node is currently selected")
	}
	if t.selectedNode.nodeType != FolderNode {
		return errors.New("only folder nodes can be expanded or collapsed")
	}

	t.ToggleExpand(t.selectedNode)
	return nil
}

func (t *FsTree) renderNode(node *FsNode, depth int, builder *strings.Builder) {
	if node == nil {
		return
	}

	folderInRoot := false
	if node.nodeType == FolderNode && depth == 1 && node.prev != nil {
		folderInRoot = true
	}

	indent := strings.Repeat(" ", depth)
	prevIndent := strings.Repeat(" ", max(depth-1, 0))

	var icon string
	switch node.nodeType {
	case FolderNode:
		if node.expanded {
			icon = indent + styles.ArrowDownIcon
		} else {
			icon = indent + styles.ArrowRightIcon
		}
	case FileNode:
		icon = styles.VerticalLine
		icon = lipgloss.NewStyle().Faint(true).Render(prevIndent + icon + " ")
	}

	// highlight if selected or hovered
	// note: the logic of lines cache needs to match render
	fileName := node.FileName()
	isSelected := node == t.selectedNode
	isHovered := node == t.hoveredNode

	if isSelected {
		fileName = lipgloss.NewStyle().Foreground(styles.Highlight).Bold(true).Render(fileName)
	} else if isHovered {
		fileName = lipgloss.NewStyle().Foreground(styles.HoverHighlight).Render(fileName)
	}

	line := icon + " " + fileName + "\n"

	if depth > 0 {
		if folderInRoot {
			line = "\n" + line
		}
		builder.WriteString(line)
	}

	if node.expanded {
		for _, child := range node.children {
			t.renderNode(child, depth+1, builder)
		}
	}
}

// builds a cache of line num to rendered node in view
func (t *FsTree) buildLines() {
	t.lines = make(map[int]*FsNode)
	line := -1
	flatTree := make([]*FsNode, 0)
	t.buildLinesRec(t.root, 0, &line, &flatTree) // root has -1 depth and -1 index, as it's not meant to be rendered
	// everything is a child of root
	flatTree[0] = nil // basically a merge of skip root and padding ends with nil
	flatTree = append(flatTree, nil)
	t.totalLines = line

	for i := 1; i < len(flatTree)-1; i++ {
		flatTree[i].prev = flatTree[i-1]
		flatTree[i].next = flatTree[i+1]
	}
}

// Deprecated: isn't meant to be used directly
func (t *FsTree) buildLinesRec(node *FsNode, depth int, currentLine *int, flatTree *[]*FsNode) {
	if node == nil {
		return
	}

	if node.nodeType == FolderNode && depth == 1 && *currentLine != 0 {
		(*currentLine)++
	}
	t.lines[*currentLine] = node
	node.line = *currentLine
	*flatTree = append(*flatTree, node)
	(*currentLine)++

	if node.expanded {
		for _, child := range node.children {
			t.buildLinesRec(child, depth+1, currentLine, flatTree)
		}
	}
}

func (t *FsTree) CreateNode(folder *FsNode, name string, nodeType FsNodeType) error {
	if folder == nil {
		return errors.New("folder node cannot be nil")
	}

	if folder.nodeType == FileNode {
		folder = folder.parent
	}

	if folder.nodeType != FolderNode {
		return errors.New("parent node must be a folder. this should not be allowed by ui")
	} else if folder.nodeType == FolderNode && !folder.expanded {
		folder.expanded = true
	}

	if name == "" {
		return errors.New("node name cannot be empty")
	}

	if nodeType == FileNode && !strings.HasSuffix(name, ".md") {
		name += ".md"
	}

	path := filepath.Join(folder.path, name)
	// materialise it first
	switch nodeType {
	case FileNode:
		err := createFile(path, []byte{})
		if err != nil {
			return err
		}
	case FolderNode:
		err := createFolder(path)
		if err != nil {
			return err
		}
	}

	expanded := false
	if nodeType == FolderNode {
		expanded = true
	}
	newNode := &FsNode{
		nodeType: nodeType,
		path:     path,
		children: make([]*FsNode, 0),
		parent:   folder,
		expanded: expanded,
	}
	// files are first of children, folder last of children
	if nodeType == FileNode {
		folder.children = append([]*FsNode{newNode}, folder.children...)
	} else {
		folder.children = append(folder.children, newNode)
	}
	t.selectedNode = newNode
	t.buildLines()
	return nil
}
