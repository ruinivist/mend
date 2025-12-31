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
	expanded bool
}

func (n *FsNode) FileName() string {
	return filepath.Base(n.path)
}

// ==================== FsNode definition ====================
type FsTree struct {
	root *FsNode
}

func NewFsTree(rootPath string) *FsTree {
	root := &FsNode{
		nodeType: FolderNode,
		path:     rootPath,
		children: make([]*FsNode, 0),
		expanded: true,
	}
	walkFileSystemAndBuildTree(rootPath, root)

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

	line := icon + " " + node.FileName() + "\n"
	builder.WriteString(line)

	if node.expanded {
		for _, child := range node.children {
			t.renderNode(child, depth+1, builder)
		}
	}
}
