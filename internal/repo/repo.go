package repo

import (
	"fmt"
	"os"
	"path/filepath"
)

const repoDir = ".arbor"
const headFile = "HEAD"

var dirs = []string{
	"objects",
	"refs/heads",
}

func Init(path string) error {
	fmt.Println("Initializing repository at", path)
	basePath := filepath.Join(path, repoDir)
	for _, dir := range dirs {
		if err := createDir(filepath.Join(basePath, dir)); err != nil {
			return err
		}
	}
	headPath := filepath.Join(basePath, headFile)
	return os.WriteFile(headPath, []byte("refs/heads/main"), 0644)
}

func createDir(path string) error {
	return os.MkdirAll(path, 0755)
}
