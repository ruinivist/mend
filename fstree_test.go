package main

import (
	"os"
	"path/filepath"
	"testing"
)

// tests creating a new fstree
func TestNewFsTree(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mend_fstree_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// setup structure:
	// root/
	//   folder1/
	//     file1.md
	//   file2.md

	os.Mkdir(filepath.Join(tmpDir, "folder1"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "folder1", "file1.md"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.md"), []byte(""), 0644)

	tree := NewFsTree(tmpDir, 0)

	if tree == nil {
		t.Fatal("expected tree to be created, got nil")
	}
	if tree.root.path != tmpDir {
		t.Errorf("expected root path %s, got %s", tmpDir, tree.root.path)
	}
	// root children: folder1, file2.md
	if len(tree.root.children) != 2 {
		t.Errorf("expected 2 children, got %d", len(tree.root.children))
	}
}

// tests creating nodes via tree
func TestTreeCreateNode(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mend_fstree_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	tree := NewFsTree(tmpDir, 0)
	// create folder
	err = tree.CreateNode(tree.root, "new_folder", FolderNode)
	if err != nil {
		t.Errorf("expected no error creating folder, got %v", err)
	}
	// check fs
	if _, err := os.Stat(filepath.Join(tmpDir, "new_folder")); os.IsNotExist(err) {
		t.Error("folder not created on fs")
	}
	// check tree
	if len(tree.root.children) != 1 || tree.root.children[0].nodeType != FolderNode {
		t.Error("tree block not updated with new folder")
	}

	// create file in new folder
	newFolder := tree.root.children[0]
	err = tree.CreateNode(newFolder, "new_file", FileNode)
	if err != nil {
		t.Errorf("expected no error creating file, got %v", err)
	}

	// check fs (.md appended automatically)
	if _, err := os.Stat(filepath.Join(tmpDir, "new_folder", "new_file.md")); os.IsNotExist(err) {
		t.Error("file not created on fs")
	}
	// check tree
	if len(newFolder.children) != 1 || newFolder.children[0].nodeType != FileNode {
		t.Error("tree node not updated with new file")
	}
}

// tests deleting nodes
func TestTreeDeleteNode(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mend_fstree_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// setup: root/file.md
	os.WriteFile(filepath.Join(tmpDir, "file.md"), []byte(""), 0644)

	tree := NewFsTree(tmpDir, 0)
	targetNode := tree.root.children[0]

	err = tree.DeleteNode(targetNode)
	if err != nil {
		t.Errorf("expected no error deleting node, got %v", err)
	}

	// check fs
	if _, err := os.Stat(filepath.Join(tmpDir, "file.md")); !os.IsNotExist(err) {
		t.Error("file not deleted from fs")
	}
	// check tree
	if len(tree.root.children) != 0 {
		t.Error("tree node not updated after deletion")
	}
}

// tests toggling expand
func TestTreeToggleExpand(t *testing.T) {
	root := &FsNode{
		nodeType: FolderNode,
		expanded: true,
		children: []*FsNode{}, // Initialize children
	}

	node := &FsNode{
		nodeType: FolderNode,
		expanded: true,
		parent:   root,
	}
	root.children = append(root.children, node)

	// mocked tree for this test
	tree := &FsTree{
		root: root,
	}
	// Initial buildLines to setup state
	tree.buildLines()

	err := tree.ToggleExpand(node)
	if err != nil {
		t.Error(err)
	}
	if node.expanded {
		t.Error("expected node to be collapsed")
	}

	err = tree.ToggleExpand(node)
	if err != nil {
		t.Error(err)
	}
	if !node.expanded {
		t.Error("expected node to be expanded")
	}
}
