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

func createFile(path string, content []byte) error {
	if path == "" {
		return errors.New("file path cannot be empty")
	}

	if _, err := os.Stat(path); err == nil {
		return errors.New("file already exists")
	}

	return os.WriteFile(path, content, 0644)
}

func createFolder(path string) error {
	if path == "" {
		return errors.New("folder path cannot be empty")
	}

	if _, err := os.Stat(path); err == nil {
		return errors.New("folder already exists")
	}

	return os.Mkdir(path, 0755)
}

func deleteFile(path string) error {
	if path == "" {
		return errors.New("file path cannot be empty")
	}

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.New("file does not exist")
		}
		return err
	}

	if info.IsDir() {
		return errors.New("path is a directory, not a file")
	}

	return os.Remove(path)
}

func deleteFolder(path string) error {
	if path == "" {
		return errors.New("folder path cannot be empty")
	}

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.New("folder does not exist")
		}
		return err
	}

	if !info.IsDir() {
		return errors.New("path is a file, not a directory")
	}

	return os.RemoveAll(path)
}
