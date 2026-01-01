/*
another test that is entirely ae ayee generated
- ruinivist, 30Dec25
*/

package fstree

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWalkFileSystemAndBuildTree(t *testing.T) {
	t.Run("build tree from real directory", func(t *testing.T) {
		// Create temporary directory structure
		tmpDir := t.TempDir()

		// Create test structure:
		// tmpDir/
		//   file1.txt
		//   file2.go
		//   folder1/
		//     nested.txt
		//   folder2/
		os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("content"), 0644)
		os.WriteFile(filepath.Join(tmpDir, "file2.go"), []byte("code"), 0644)
		os.Mkdir(filepath.Join(tmpDir, "folder1"), 0755)
		os.WriteFile(filepath.Join(tmpDir, "folder1", "nested.txt"), []byte("nested"), 0644)
		os.Mkdir(filepath.Join(tmpDir, "folder2"), 0755)

		// Build tree
		root := &FsNode{
			nodeType: FolderNode,
			path:     tmpDir,
			children: make([]*FsNode, 0),
			expanded: true,
		}

		err := walkFileSystemAndBuildTree(tmpDir, root)
		if err != nil {
			t.Fatalf("walkFileSystemAndBuildTree() error = %v", err)
		}

		// Verify structure
		if len(root.children) != 4 {
			t.Fatalf("root children count = %d, want 4", len(root.children))
		}

		// Check children exist and have correct types
		fileCount := 0
		folderCount := 0
		var folder1 *FsNode

		for _, child := range root.children {
			if child.parent != root {
				t.Errorf("child %q parent pointer incorrect", child.FileName())
			}

			switch child.nodeType {
			case FileNode:
				fileCount++
			case FolderNode:
				folderCount++
				if child.FileName() == "folder1" {
					folder1 = child
				}
			}
		}

		if fileCount != 2 {
			t.Errorf("file count = %d, want 2", fileCount)
		}
		if folderCount != 2 {
			t.Errorf("folder count = %d, want 2", folderCount)
		}

		// Verify nested structure
		if folder1 == nil {
			t.Fatal("folder1 not found")
		}
		if len(folder1.children) != 1 {
			t.Errorf("folder1 children count = %d, want 1", len(folder1.children))
		}
		if folder1.children[0].FileName() != "nested.txt" {
			t.Errorf("nested file name = %q, want %q", folder1.children[0].FileName(), "nested.txt")
		}
		if folder1.children[0].nodeType != FileNode {
			t.Errorf("nested node type = %v, want %v", folder1.children[0].nodeType, FileNode)
		}
	})

	t.Run("error when node is nil", func(t *testing.T) {
		tmpDir := t.TempDir()
		err := walkFileSystemAndBuildTree(tmpDir, nil)
		if err == nil {
			t.Fatal("walkFileSystemAndBuildTree() with nil node should return error")
		}
		if err.Error() != "node cannot be nil" {
			t.Errorf("error message = %q", err.Error())
		}
	})

	t.Run("error when node already has children", func(t *testing.T) {
		tmpDir := t.TempDir()
		root := &FsNode{
			nodeType: FolderNode,
			path:     tmpDir,
			children: []*FsNode{{nodeType: FileNode, path: "dummy"}},
			expanded: true,
		}

		err := walkFileSystemAndBuildTree(tmpDir, root)
		if err == nil {
			t.Fatal("walkFileSystemAndBuildTree() with existing children should return error")
		}
		if err.Error() != "node already has children" {
			t.Errorf("error message = %q", err.Error())
		}
	})

	t.Run("error when path does not exist", func(t *testing.T) {
		root := &FsNode{
			nodeType: FolderNode,
			path:     "/nonexistent/path",
			children: make([]*FsNode, 0),
			expanded: true,
		}

		err := walkFileSystemAndBuildTree("/nonexistent/path", root)
		if err == nil {
			t.Fatal("walkFileSystemAndBuildTree() with nonexistent path should return error")
		}
	})

	t.Run("empty directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		root := &FsNode{
			nodeType: FolderNode,
			path:     tmpDir,
			children: make([]*FsNode, 0),
			expanded: true,
		}

		err := walkFileSystemAndBuildTree(tmpDir, root)
		if err != nil {
			t.Fatalf("walkFileSystemAndBuildTree() error = %v", err)
		}

		if len(root.children) != 0 {
			t.Errorf("empty directory should have 0 children, got %d", len(root.children))
		}
	})
}

func TestCreateFile(t *testing.T) {
	t.Run("create file successfully", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "test.txt")
		content := []byte("test content")

		err := createFile(filePath, content)
		if err != nil {
			t.Fatalf("createFile() error = %v", err)
		}

		data, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("failed to read created file: %v", err)
		}
		if string(data) != string(content) {
			t.Errorf("file content = %q, want %q", string(data), string(content))
		}
	})

	t.Run("create empty file", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "empty.txt")

		err := createFile(filePath, nil)
		if err != nil {
			t.Fatalf("createFile() error = %v", err)
		}

		info, err := os.Stat(filePath)
		if err != nil {
			t.Fatalf("failed to stat file: %v", err)
		}
		if info.Size() != 0 {
			t.Errorf("file size = %d, want 0", info.Size())
		}
	})

	t.Run("error when path is empty", func(t *testing.T) {
		err := createFile("", []byte("content"))
		if err == nil {
			t.Fatal("createFile() with empty path should return error")
		}
		if err.Error() != "file path cannot be empty" {
			t.Errorf("error message = %q", err.Error())
		}
	})

	t.Run("error when file already exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "existing.txt")
		os.WriteFile(filePath, []byte("existing"), 0644)

		err := createFile(filePath, []byte("new"))
		if err == nil {
			t.Fatal("createFile() on existing file should return error")
		}
		if err.Error() != "file already exists" {
			t.Errorf("error message = %q", err.Error())
		}
	})
}

func TestCreateFolder(t *testing.T) {
	t.Run("create folder successfully", func(t *testing.T) {
		tmpDir := t.TempDir()
		folderPath := filepath.Join(tmpDir, "testfolder")

		err := createFolder(folderPath)
		if err != nil {
			t.Fatalf("createFolder() error = %v", err)
		}

		info, err := os.Stat(folderPath)
		if err != nil {
			t.Fatalf("failed to stat folder: %v", err)
		}
		if !info.IsDir() {
			t.Error("created path is not a directory")
		}
	})

	t.Run("error when path is empty", func(t *testing.T) {
		err := createFolder("")
		if err == nil {
			t.Fatal("createFolder() with empty path should return error")
		}
		if err.Error() != "folder path cannot be empty" {
			t.Errorf("error message = %q", err.Error())
		}
	})

	t.Run("error when folder already exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		folderPath := filepath.Join(tmpDir, "existing")
		os.Mkdir(folderPath, 0755)

		err := createFolder(folderPath)
		if err == nil {
			t.Fatal("createFolder() on existing folder should return error")
		}
		if err.Error() != "folder already exists" {
			t.Errorf("error message = %q", err.Error())
		}
	})
}

func TestDeleteFile(t *testing.T) {
	t.Run("delete file successfully", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "delete.txt")
		os.WriteFile(filePath, []byte("content"), 0644)

		err := deleteFile(filePath)
		if err != nil {
			t.Fatalf("deleteFile() error = %v", err)
		}

		if _, err := os.Stat(filePath); !os.IsNotExist(err) {
			t.Error("file still exists after deletion")
		}
	})

	t.Run("error when path is empty", func(t *testing.T) {
		err := deleteFile("")
		if err == nil {
			t.Fatal("deleteFile() with empty path should return error")
		}
		if err.Error() != "file path cannot be empty" {
			t.Errorf("error message = %q", err.Error())
		}
	})

	t.Run("error when file does not exist", func(t *testing.T) {
		err := deleteFile("/nonexistent/file.txt")
		if err == nil {
			t.Fatal("deleteFile() on nonexistent file should return error")
		}
		if err.Error() != "file does not exist" {
			t.Errorf("error message = %q", err.Error())
		}
	})

	t.Run("error when path is a directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		folderPath := filepath.Join(tmpDir, "folder")
		os.Mkdir(folderPath, 0755)

		err := deleteFile(folderPath)
		if err == nil {
			t.Fatal("deleteFile() on directory should return error")
		}
		if err.Error() != "path is a directory, not a file" {
			t.Errorf("error message = %q", err.Error())
		}
	})
}

func TestDeleteFolder(t *testing.T) {
	t.Run("delete empty folder successfully", func(t *testing.T) {
		tmpDir := t.TempDir()
		folderPath := filepath.Join(tmpDir, "emptyfolder")
		os.Mkdir(folderPath, 0755)

		err := deleteFolder(folderPath)
		if err != nil {
			t.Fatalf("deleteFolder() error = %v", err)
		}

		if _, err := os.Stat(folderPath); !os.IsNotExist(err) {
			t.Error("folder still exists after deletion")
		}
	})

	t.Run("delete folder with contents", func(t *testing.T) {
		tmpDir := t.TempDir()
		folderPath := filepath.Join(tmpDir, "withcontents")
		os.Mkdir(folderPath, 0755)
		os.WriteFile(filepath.Join(folderPath, "file.txt"), []byte("content"), 0644)
		os.Mkdir(filepath.Join(folderPath, "subfolder"), 0755)

		err := deleteFolder(folderPath)
		if err != nil {
			t.Fatalf("deleteFolder() error = %v", err)
		}

		if _, err := os.Stat(folderPath); !os.IsNotExist(err) {
			t.Error("folder still exists after deletion")
		}
	})

	t.Run("error when path is empty", func(t *testing.T) {
		err := deleteFolder("")
		if err == nil {
			t.Fatal("deleteFolder() with empty path should return error")
		}
		if err.Error() != "folder path cannot be empty" {
			t.Errorf("error message = %q", err.Error())
		}
	})

	t.Run("error when folder does not exist", func(t *testing.T) {
		err := deleteFolder("/nonexistent/folder")
		if err == nil {
			t.Fatal("deleteFolder() on nonexistent folder should return error")
		}
		if err.Error() != "folder does not exist" {
			t.Errorf("error message = %q", err.Error())
		}
	})

	t.Run("error when path is a file", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "file.txt")
		os.WriteFile(filePath, []byte("content"), 0644)

		err := deleteFolder(filePath)
		if err == nil {
			t.Fatal("deleteFolder() on file should return error")
		}
		if err.Error() != "path is a file, not a directory" {
			t.Errorf("error message = %q", err.Error())
		}
	})
}
