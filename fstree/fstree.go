/*
This file contains the implementation of an in-memory tree of files and folders that the app used.
It is initialised at startup and used for state changes in UI and later persisted to disk.

- ruinivist, 30Dec25
*/

package fstree

import (
	"errors"
	"mend/styles"
	"mend/utils"
	"os"
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

// ==================== Message types ====================
type Msg interface{}

type MsgMoveUp struct{}
type MsgMoveDown struct{}
type MsgToggleExpand struct{}
type MsgLeftClickLine struct {
	Line int
}
type MsgCreateNode struct {
	Parent   *FsNode
	Name     string
	NodeType FsNodeType
}
type MsgDeleteNode struct {
	Node *FsNode
}
type MsgHover struct {
	Line int
}

// ==================== FsNode definition ====================
// a single node, Fs => deals with file system related info mostly
// a name for example can be content derived as well
type FsNode struct {
	nodeType FsNodeType
	path     string
	children []*FsNode
	parent   *FsNode // for fast traversal up the tree
	expanded bool    // makes sense only for folder nodes
	line     int
	// in the flattree
	prev *FsNode
	next *FsNode
}

func (n *FsNode) FileName() string {
	return filepath.Base(n.path)
}

// ==================== FsNode definition ====================
type FsTree struct {
	root         *FsNode
	lines        map[int]*FsNode // flattened view of nodes for easy line access, map so that I can handle blank padding
	selectedNode *FsNode
	hoveredNode  *FsNode
}

// ==================== Bubble Tea Interface Implementation ====================
func (t *FsTree) Init() tea.Cmd {
	return nil
}

func (t *FsTree) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case MsgMoveUp:
		_ = t.MoveUp()
	case MsgMoveDown:
		_ = t.MoveDown()
	case MsgToggleExpand:
		_ = t.ToggleSelectedExpand()
	case MsgLeftClickLine:
		nodeAtLine := t.lines[m.Line]
		if nodeAtLine != nil {
			t.selectedNode = nodeAtLine
			if nodeAtLine.nodeType == FolderNode {
				_ = t.ToggleExpand(nodeAtLine)
			}
		}
	case MsgCreateNode:
		_ = t.CreateNode(m.Parent, m.Name, m.NodeType)
	case MsgDeleteNode:
		_ = t.DeleteNode(m.Node)
	case MsgHover:
		t.hoveredNode = t.lines[m.Line]
	}
	return t, nil
}

func (t *FsTree) View() string {
	builder := &strings.Builder{}
	t.renderNode(t.root, 0, builder)
	return builder.String()
}

// ==================== FsTree helper methods ====================

func NewFsTree(rootPath string) *FsTree {
	root := &FsNode{
		nodeType: FolderNode,
		path:     rootPath,
		children: make([]*FsNode, 0),
		expanded: true,
	}
	// root is at -1, rest all are 0 indexed, REM: this walk func does no handle spaces right now
	// but since I use a map, it's easy to add
	walkFileSystemAndBuildTree(rootPath, root)

	var tree *FsTree
	if len(root.children) > 0 {
		tree = &FsTree{
			root:         root,
			selectedNode: root.children[0],
		}
	} else {
		tree = &FsTree{
			root: root,
		}
	}
	tree.buildLines()
	return tree
}

func (t *FsTree) CreateNode(parent *FsNode, name string, nodeType FsNodeType) error {
	if parent == nil {
		return errors.New("parent node cannot be nil")
	}
	if parent.nodeType != FolderNode {
		return errors.New("can only add nodes to folder nodes")
	}
	if name == "" {
		return errors.New("node name cannot be empty")
	}

	newNode := &FsNode{
		nodeType: nodeType,
		path:     filepath.Join(parent.path, name),
		children: make([]*FsNode, 0),
		parent:   parent,
		expanded: true,
	}

	parent.children = append(parent.children, newNode)
	t.buildLines()
	return nil
}

func (t *FsTree) DeleteNode(node *FsNode) error {
	if node == nil {
		return errors.New("node to delete cannot be nil")
	}
	parent := node.parent
	if parent == nil {
		return errors.New("node to delete must have a parent")
	}

	parent.children = utils.RemoveFromSlice(parent.children, node)
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

func (t *FsTree) GetSelectedContent() (string, error) {
	if t.selectedNode == nil {
		return "", errors.New("no node is currently selected")
	}

	if t.selectedNode.nodeType != FileNode {
		return "", nil
	}

	contentBytes, err := os.ReadFile(t.selectedNode.path)
	if err != nil {
		return "", err
	}
	return string(contentBytes), nil
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
		icon = lipgloss.NewStyle().Faint(true).Render(prevIndent + icon + indent)
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
	flatTree = flatTree[1:] // skip root

	for i, node := range flatTree {
		if i > 0 {
			node.prev = flatTree[i-1]
		}
		if i < len(flatTree)-1 {
			node.next = flatTree[i+1]
		}
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
