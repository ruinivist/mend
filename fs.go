/*
This file has the raw file io operations that are used by FsTree
or elsewhere in the app. Nowhere else should there be a direct
fileopen

- ruinivist, 30Dec25
*/

package main

import (
	"errors"
	"os"
	"path/filepath"
)

func walkFileSystemAndBuildTree(rootPath string, node *FsNode) error {
	if node == nil {
		return errors.New("node cannot be nil")
	}
	if len(node.children) > 0 {
		return errors.New("node already has children")
	}

	entries, err := os.ReadDir(rootPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		childPath := filepath.Join(rootPath, entry.Name())

		var nodeType FsNodeType
		if entry.IsDir() {
			nodeType = FolderNode
		} else {
			nodeType = FileNode
		}

		childNode := &FsNode{
			nodeType: nodeType,
			path:     childPath,
			children: make([]*FsNode, 0),
			parent:   node,
			expanded: false,
		}

		node.children = append(node.children, childNode)

		// Recursively walk subdirectories
		if entry.IsDir() {
			if err := walkFileSystemAndBuildTree(childPath, childNode); err != nil {
				return err
			}
		}
	}

	return nil
}
