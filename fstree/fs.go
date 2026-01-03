/*
This file has the raw file io operations that are used by FsTree
or elsewhere in the app. Nowhere else should there be a direct
fileopen

- ruinivist, 30Dec25
*/

package fstree

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

	files := make([]os.DirEntry, 0)
	folders := make([]os.DirEntry, 0)

	for _, entry := range entries {
		// dot folders and files skipped
		if len(entry.Name()) > 0 && entry.Name()[0] == '.' {
			continue
		}

		if entry.IsDir() {
			folders = append(folders, entry)
		} else {
			files = append(files, entry)
		}
	}

	for _, file := range files {
		newNode := &FsNode{
			nodeType: FileNode,
			path:     filepath.Join(rootPath, file.Name()),
			children: make([]*FsNode, 0),
			parent:   node,
			expanded: false,
		}
		node.children = append(node.children, newNode)
	}

	for _, folder := range folders {
		newNode := &FsNode{
			nodeType: FolderNode,
			path:     filepath.Join(rootPath, folder.Name()),
			children: make([]*FsNode, 0),
			parent:   node,
			expanded: true, // all expanded by default
		}
		node.children = append(node.children, newNode)
		walkFileSystemAndBuildTree(newNode.path, newNode)
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
