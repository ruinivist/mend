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
	// for view layer
	expanded bool // makes sense only for folder nodes
}

func (n *FsNode) FileName() string {
	return filepath.Base(n.path)
}

// ==================== FsNode definition ====================
type FsTree struct {
	root     *FsNode
	selected int
	lines    []*FsNode // flattened view of nodes
}

func (t *FsTree) selectedNode() *FsNode {
	if t.selected <= 0 || t.selected >= len(t.lines) {
		return nil
	}
	return t.lines[t.selected]
}

func NewFsTree(rootPath string) *FsTree {
	root := &FsNode{
		nodeType: FolderNode,
		path:     rootPath,
		children: make([]*FsNode, 0),
		expanded: true,
	}
	flatTree := make([]*FsNode, 0)
	walkFileSystemAndBuildTree(rootPath, root, &flatTree)

	if len(root.children) > 0 {
		return &FsTree{
			root:     root,
			selected: 1, // first child
			lines:    flatTree,
		}
	}
	return &FsTree{
		root:  root,
		lines: flatTree,
	}
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
	return nil
}

func (t *FsTree) move(delta int) error {
	if t.selectedNode() == nil {
		return errors.New("no node is currently selected")
	}
	if delta != -1 && delta != 1 {
		return errors.New("delta must be either -1 (up) or 1 (down)")
	}

	newIndex := t.selected + delta
	if newIndex < 1 || newIndex >= len(t.lines) {
		return nil // noop
	}
	t.selected = newIndex
	return nil
}

func (t *FsTree) MoveUp() error   { return t.move(-1) }
func (t *FsTree) MoveDown() error { return t.move(1) }

func (t *FsTree) SelectNodeAtLine(line int) error {
	if line < 0 || line >= len(t.lines) {
		return errors.New("line number out of bounds")
	}
	t.selected = line
	return nil
}

func (t *FsTree) ToggleSelectedExpand() error {
	if t.selectedNode() == nil {
		return errors.New("no node is currently selected")
	}
	if t.selectedNode().nodeType != FolderNode {
		return errors.New("only folder nodes can be expanded or collapsed")
	}

	t.ToggleExpand(t.selectedNode())
	return nil
}

func (t *FsTree) GetSelectedContent() (string, error) {
	if t.selectedNode() == nil {
		return "", errors.New("no node is currently selected")
	}

	if t.selectedNode().nodeType != FileNode {
		return "", nil
	}

	contentBytes, err := os.ReadFile(t.selectedNode().path)
	if err != nil {
		return "", err
	}
	return string(contentBytes), nil
}

func (t *FsTree) Render(hoverLine int) string {
	builder := &strings.Builder{}
	lineCounter := 0
	t.renderNode(t.root, 0, builder, hoverLine, &lineCounter)
	return builder.String()
}

func (t *FsTree) renderNode(node *FsNode, depth int, builder *strings.Builder, hoverLine int, currentLine *int) {
	if node == nil {
		return
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
	fileName := node.FileName()
	isSelected := node == t.selectedNode()
	isHovered := hoverLine >= 0 && *currentLine == hoverLine

	if isSelected {
		fileName = lipgloss.NewStyle().Foreground(styles.Highlight).Bold(true).Render(fileName)
	} else if isHovered {
		fileName = lipgloss.NewStyle().Foreground(styles.HoverHighlight).Render(fileName)
	}

	line := icon + " " + fileName + "\n"
	builder.WriteString(line)
	*currentLine++

	if node.expanded {
		for _, child := range node.children {
			t.renderNode(child, depth+1, builder, hoverLine, currentLine)
		}
	}
}
