package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

const repoDir = ".arbor"
const objectsDir = "objects"
const indexDir = "index"
const refsDir = "refs/heads"

func IsRepoDir(name string) bool {
	return repoDir == name
}

func CreateDir(path string) error {
	return os.MkdirAll(path, 0755)
}

func GetRepoDir(path string) string {
	return filepath.Join(path, repoDir)
}

func GetObjectsDir(path string) string {
	return filepath.Join(GetRepoDir(path), objectsDir)
}

func GetIndexPath(path string) string {
	return filepath.Join(GetRepoDir(path), indexDir)
}

func GetRefsDir(path string) string {
	return filepath.Join(GetRepoDir(path), refsDir)
}

func Exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func WriteFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0644)
}

func ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func RemoveFile(path string) error {
	if _, err := os.Stat(path); err == nil {
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("remove %s: %w", path, err)
		}
	}
	return nil
}
