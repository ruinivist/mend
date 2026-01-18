package filesystem

import (
	"os"
	"path/filepath"
	"testing"
)

// tests creating a file in a temp dir
func TestCreateFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mend_fs_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// test valid creation
	filePath := filepath.Join(tmpDir, "test.md")
	err = CreateFile(filePath, []byte("content"))
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("file was not created")
	}

	// test creation failure (exists)
	err = CreateFile(filePath, []byte("content"))
	if err == nil {
		t.Error("expected error for existing file, got nil")
	}

	// test empty path
	err = CreateFile("", []byte{})
	if err == nil {
		t.Error("expected error for empty path, got nil")
	}
}

// tests creating a folder
func TestCreateFolder(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mend_fs_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	folderPath := filepath.Join(tmpDir, "subfolder")
	err = CreateFolder(folderPath)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	info, err := os.Stat(folderPath)
	if os.IsNotExist(err) {
		t.Error("folder was not created")
	}
	if !info.IsDir() {
		t.Error("created path is not a directory")
	}

	// test creation failure (exists)
	err = CreateFolder(folderPath)
	if err == nil {
		t.Error("expected error for existing folder, got nil")
	}

	// test empty path
	err = CreateFolder("")
	if err == nil {
		t.Error("expected error for empty path, got nil")
	}
}

// tests recursive deletion
func TestDeletePath(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mend_fs_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	subDir := filepath.Join(tmpDir, "subdir")
	os.Mkdir(subDir, 0755)
	fileInSubDir := filepath.Join(subDir, "file.txt")
	os.WriteFile(fileInSubDir, []byte("data"), 0644)

	err = DeletePath(subDir)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if _, err := os.Stat(subDir); !os.IsNotExist(err) {
		t.Error("directory was not deleted")
	}

	// test delete non-existent
	err = DeletePath(filepath.Join(tmpDir, "nonexistent"))
	if err == nil {
		t.Error("expected error for nonexistent path, got nil")
	}

	// test empty path
	err = DeletePath("")
	if err == nil {
		t.Error("expected error for empty path, got nil")
	}
}
