package filesystem

import (
	"errors"
	"os"
)

func CreateFile(path string, content []byte) error {
	if path == "" {
		return errors.New("file path cannot be empty")
	}

	if _, err := os.Stat(path); err == nil {
		return errors.New("file already exists")
	}

	return os.WriteFile(path, content, 0644)
}

func CreateFolder(path string) error {
	if path == "" {
		return errors.New("folder path cannot be empty")
	}

	if _, err := os.Stat(path); err == nil {
		return errors.New("folder already exists")
	}

	return os.Mkdir(path, 0755)
}

func DeletePath(path string) error {
	if path == "" {
		return errors.New("path cannot be empty")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return errors.New("path does not exist")
	}

	return os.RemoveAll(path)
}
