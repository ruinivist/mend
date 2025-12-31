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
	selected *FsNode // currently selected node
}

func NewFsTree(rootPath string) *FsTree {
	root := &FsNode{
		nodeType: FolderNode,
		path:     rootPath,
		children: make([]*FsNode, 0),
		expanded: true,
	}
	walkFileSystemAndBuildTree(rootPath, root)

	if len(root.children) > 0 {
		return &FsTree{
			root:     root,
			selected: root.children[0],
		}
	}
	return &FsTree{
		root: root,
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

// At this point I think there HAS to be a simpler way to do this
// but since I'm going to refactor it some day let's leave this
// as it is
func (t *FsTree) move(delta int) error {
	if t.selected == nil {
		return errors.New("no node is currently selected")
	}
	if delta != -1 && delta != 1 {
		return errors.New("delta must be either -1 (up) or 1 (down)")
	}

	// base case: if it's a folder and expanded and going down
	if delta == 1 && t.selected.nodeType == FolderNode && t.selected.expanded && len(t.selected.children) > 0 {
		t.selected = t.selected.children[0]
		return nil
	}
	// base case: if it's the first element of a folder and going up
	if delta == -1 && t.selected.parent != t.root {
		parent := t.selected.parent
		if len(parent.children) > 0 && parent.children[0] == t.selected {
			t.selected = parent
			return nil
		}
	}

	// same level sibling move
	// the idea is really simple for now:
	// - go up the parent and find the sibling in the direction
	//   - if no sibling go up again
	// - if no parent stop
	siblingFor := t.selected
	for siblingFor.parent != nil {
		parent := siblingFor.parent
		// O(n) for now, though easy to obtimize with index tracking
		nextIdx := -1
		for idx, child := range parent.children {
			if child == siblingFor {
				nextIdx = idx + delta
				break
			}
		}

		if nextIdx >= 0 && nextIdx < len(parent.children) {
			// found case
			t.selected = parent.children[nextIdx]
			break
		}
		// not found case, go up again
		siblingFor = parent
	}

	// case: folder and I'm going up, I want to go as deep and last
	if delta == -1 && t.selected.nodeType == FolderNode {
		current := t.selected
		for {
			if len(current.children) == 0 || !current.expanded {
				break
			}
			current = current.children[len(current.children)-1]
		}
		t.selected = current
	}
	return nil
}

func (t *FsTree) MoveUp() error   { return t.move(-1) }
func (t *FsTree) MoveDown() error { return t.move(1) }

func (t *FsTree) ToggleSelectedExpand() error {
	if t.selected == nil {
		return errors.New("no node is currently selected")
	}
	if t.selected.nodeType != FolderNode {
		return errors.New("only folder nodes can be expanded or collapsed")
	}

	t.ToggleExpand(t.selected)
	return nil
}

func (t *FsTree) GetSelectedContent() (string, error) {
	if t.selected == nil {
		return "", errors.New("no node is currently selected")
	}

	if t.selected.nodeType != FileNode {
		return "", nil
	}

	contentBytes, err := os.ReadFile(t.selected.path)
	if err != nil {
		return "", err
	}
	return string(contentBytes), nil
}

func (t *FsTree) Render() string {
	builder := &strings.Builder{}
	t.renderNode(t.root, 0, builder)
	return builder.String()
}

func (t *FsTree) renderNode(node *FsNode, depth int, builder *strings.Builder) {
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

	// Apply highlight if selected
	fileName := node.FileName()
	if node == t.selected {
		fileName = lipgloss.NewStyle().Foreground(styles.Highlight).Bold(true).Render(fileName)
	}

	line := icon + " " + fileName + "\n"
	builder.WriteString(line)

	if node.expanded {
		for _, child := range node.children {
			t.renderNode(child, depth+1, builder)
		}
	}
}
