/*
This file is almost entirely ae ayee generated. That's one of the
few good uses of ae ayee in my opinion

- ruinivist, 30Dec25
*/

package fstree

import (
	"path/filepath"
	"testing"
)

func TestFsNode_FileName(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "simple filename",
			path:     "/home/user/test.txt",
			expected: "test.txt",
		},
		{
			name:     "folder name",
			path:     "/home/user/folder",
			expected: "folder",
		},
		{
			name:     "root path",
			path:     "/",
			expected: "/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &FsNode{path: tt.path}
			got := node.FileName()
			if got != tt.expected {
				t.Errorf("FileName() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestNewFsTree(t *testing.T) {
	rootPath := "/home/test"
	tree := NewFsTree(rootPath)

	if tree == nil {
		t.Fatal("NewFsTree() returned nil")
	}
	if tree.root == nil {
		t.Fatal("tree.root is nil")
	}
	if tree.root.nodeType != FolderNode {
		t.Errorf("root nodeType = %v, want %v", tree.root.nodeType, FolderNode)
	}
	if tree.root.path != rootPath {
		t.Errorf("root path = %q, want %q", tree.root.path, rootPath)
	}
	if !tree.root.expanded {
		t.Error("root should be expanded by default")
	}
	if tree.root.children == nil {
		t.Error("root children slice should be initialized")
	}
}

func TestFsTreeImpl_CreateNode(t *testing.T) {
	t.Run("create file node successfully", func(t *testing.T) {
		tree := NewFsTree("/home/test")
		err := tree.CreateNode(tree.root, "file.txt", FileNode)
		if err != nil {
			t.Fatalf("CreateNode() error = %v, want nil", err)
		}
		if len(tree.root.children) != 1 {
			t.Fatalf("children count = %d, want 1", len(tree.root.children))
		}

		child := tree.root.children[0]
		if child.nodeType != FileNode {
			t.Errorf("child nodeType = %v, want %v", child.nodeType, FileNode)
		}
		expectedPath := filepath.Join("/home/test", "file.txt")
		if child.path != expectedPath {
			t.Errorf("child path = %q, want %q", child.path, expectedPath)
		}
		if child.parent != tree.root {
			t.Error("child parent should point to root")
		}
		if child.expanded {
			t.Error("new node should not be expanded by default")
		}
	})

	t.Run("create folder node successfully", func(t *testing.T) {
		tree := NewFsTree("/home/test")
		err := tree.CreateNode(tree.root, "subfolder", FolderNode)
		if err != nil {
			t.Fatalf("CreateNode() error = %v, want nil", err)
		}
		if len(tree.root.children) != 1 {
			t.Fatalf("children count = %d, want 1", len(tree.root.children))
		}

		child := tree.root.children[0]
		if child.nodeType != FolderNode {
			t.Errorf("child nodeType = %v, want %v", child.nodeType, FolderNode)
		}
	})

	t.Run("error when parent is nil", func(t *testing.T) {
		tree := NewFsTree("/home/test")
		err := tree.CreateNode(nil, "file.txt", FileNode)
		if err == nil {
			t.Fatal("CreateNode() with nil parent should return error")
		}
		if err.Error() != "parent node cannot be nil" {
			t.Errorf("error message = %q, want %q", err.Error(), "parent node cannot be nil")
		}
	})

	t.Run("error when parent is file node", func(t *testing.T) {
		tree := NewFsTree("/home/test")
		fileNode := &FsNode{
			nodeType: FileNode,
			path:     "/home/test/file.txt",
			parent:   tree.root,
		}
		err := tree.CreateNode(fileNode, "child.txt", FileNode)
		if err == nil {
			t.Fatal("CreateNode() with file parent should return error")
		}
		if err.Error() != "can only add nodes to folder nodes" {
			t.Errorf("error message = %q, want %q", err.Error(), "can only add nodes to folder nodes")
		}
	})

	t.Run("error when name is empty", func(t *testing.T) {
		tree := NewFsTree("/home/test")
		err := tree.CreateNode(tree.root, "", FileNode)
		if err == nil {
			t.Fatal("CreateNode() with empty name should return error")
		}
		if err.Error() != "node name cannot be empty" {
			t.Errorf("error message = %q, want %q", err.Error(), "node name cannot be empty")
		}
	})

	t.Run("create nested nodes", func(t *testing.T) {
		tree := NewFsTree("/home/test")
		err := tree.CreateNode(tree.root, "folder1", FolderNode)
		if err != nil {
			t.Fatalf("CreateNode() error = %v", err)
		}
		folder1 := tree.root.children[0]

		err = tree.CreateNode(folder1, "file.txt", FileNode)
		if err != nil {
			t.Fatalf("CreateNode() error = %v", err)
		}

		if len(folder1.children) != 1 {
			t.Fatalf("folder1 children count = %d, want 1", len(folder1.children))
		}
		expectedPath := filepath.Join("/home/test", "folder1", "file.txt")
		if folder1.children[0].path != expectedPath {
			t.Errorf("nested file path = %q, want %q", folder1.children[0].path, expectedPath)
		}
	})
}

func TestFsTreeImpl_DeleteNode(t *testing.T) {
	t.Run("delete node successfully", func(t *testing.T) {
		tree := NewFsTree("/home/test")
		tree.CreateNode(tree.root, "file1.txt", FileNode)
		tree.CreateNode(tree.root, "file2.txt", FileNode)

		if len(tree.root.children) != 2 {
			t.Fatalf("initial children count = %d, want 2", len(tree.root.children))
		}

		nodeToDelete := tree.root.children[0]
		err := tree.DeleteNode(nodeToDelete)
		if err != nil {
			t.Fatalf("DeleteNode() error = %v, want nil", err)
		}

		if len(tree.root.children) != 1 {
			t.Errorf("children count after delete = %d, want 1", len(tree.root.children))
		}
	})

	t.Run("error when node is nil", func(t *testing.T) {
		tree := NewFsTree("/home/test")
		err := tree.DeleteNode(nil)
		if err == nil {
			t.Fatal("DeleteNode() with nil node should return error")
		}
		if err.Error() != "node to delete cannot be nil" {
			t.Errorf("error message = %q, want %q", err.Error(), "node to delete cannot be nil")
		}
	})

	t.Run("error when node has no parent", func(t *testing.T) {
		tree := NewFsTree("/home/test")
		orphanNode := &FsNode{
			nodeType: FileNode,
			path:     "/orphan.txt",
			parent:   nil,
		}
		err := tree.DeleteNode(orphanNode)
		if err == nil {
			t.Fatal("DeleteNode() with no parent should return error")
		}
		if err.Error() != "node to delete must have a parent" {
			t.Errorf("error message = %q, want %q", err.Error(), "node to delete must have a parent")
		}
	})
}

func TestFsTreeImpl_ToggleExpand(t *testing.T) {
	t.Run("toggle folder expansion", func(t *testing.T) {
		tree := NewFsTree("/home/test")
		tree.CreateNode(tree.root, "folder1", FolderNode)
		folder := tree.root.children[0]

		// Initially not expanded
		if folder.expanded {
			t.Error("new folder should not be expanded")
		}

		// Toggle to expanded
		err := tree.ToggleExpand(folder)
		if err != nil {
			t.Fatalf("ToggleExpand() error = %v", err)
		}
		if !folder.expanded {
			t.Error("folder should be expanded after toggle")
		}

		// Toggle back to collapsed
		err = tree.ToggleExpand(folder)
		if err != nil {
			t.Fatalf("ToggleExpand() error = %v", err)
		}
		if folder.expanded {
			t.Error("folder should be collapsed after second toggle")
		}
	})

	t.Run("error when node is nil", func(t *testing.T) {
		tree := NewFsTree("/home/test")
		err := tree.ToggleExpand(nil)
		if err == nil {
			t.Fatal("ToggleExpand() with nil node should return error")
		}
		if err.Error() != "node cannot be nil" {
			t.Errorf("error message = %q, want %q", err.Error(), "node cannot be nil")
		}
	})

	t.Run("error when node is file", func(t *testing.T) {
		tree := NewFsTree("/home/test")
		tree.CreateNode(tree.root, "file.txt", FileNode)
		fileNode := tree.root.children[0]

		err := tree.ToggleExpand(fileNode)
		if err == nil {
			t.Fatal("ToggleExpand() on file node should return error")
		}
		if err.Error() != "only folder nodes can be expanded or collapsed" {
			t.Errorf("error message = %q, want %q", err.Error(), "only folder nodes can be expanded or collapsed")
		}
	})
}
