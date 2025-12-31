/*
This file contains the implementation of an in-memory tree of files and folders that the app used.
It is initialised at startup and used for state changes in UI and later persisted to disk.

- ruinivist, 30Dec25
*/

package main

import (
	"errors"
	"path/filepath"
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

// ==================== FsTree definition ====================
type FsTree interface {
	CreateNode(parent *FsNode, name string, nodeType FsNodeType) error
	DeleteNode(node *FsNode) error
	ToggleExpand(node *FsNode) error
}

// reminder: a bit redundant but I this above shows what funcs are there at a glace
// so I still like to use this pattern ( 30Dec25)
type FsTreeImpl struct {
	root *FsNode
}

func NewFsTree(rootPath string) *FsTreeImpl {
	root := &FsNode{
		nodeType: FolderNode,
		path:     rootPath,
		children: make([]*FsNode, 0),
		expanded: true,
	}
	walkFileSystemAndBuildTree(rootPath, root)

	return &FsTreeImpl{
		root: root,
	}
}

func (t *FsTreeImpl) CreateNode(parent *FsNode, name string, nodeType FsNodeType) error {
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
		expanded: false,
	}

	parent.children = append(parent.children, newNode)
	return nil
}

func (t *FsTreeImpl) DeleteNode(node *FsNode) error {
	if node == nil {
		return errors.New("node to delete cannot be nil")
	}
	parent := node.parent
	if parent == nil {
		return errors.New("node to delete must have a parent")
	}

	parent.children = RemoveFromSlice(parent.children, node)
	return nil
}

// ToggleExpand toggles the expanded state of a node
func (t *FsTreeImpl) ToggleExpand(node *FsNode) error {
	if node == nil {
		return errors.New("node cannot be nil")
	}
	if node.nodeType != FolderNode {
		return errors.New("only folder nodes can be expanded or collapsed")
	}

	node.expanded = !node.expanded
	return nil
}

// ==================== FsTree definition ====================
