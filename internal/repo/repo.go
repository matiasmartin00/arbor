package repo

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/matiasmartin00/arbor/internal/index"
	"github.com/matiasmartin00/arbor/internal/object"
	"github.com/matiasmartin00/arbor/internal/utils"
)

func Init(path string) error {
	fmt.Println("Initializing repository at", path)
	dirs := []string{
		utils.GetObjectsDir(path),
		utils.GetRefsDir(path),
	}
	for _, dir := range dirs {
		if err := utils.CreateDir(dir); err != nil {
			return err
		}
	}
	headPath := utils.GetHeadPath(path)
	return utils.WriteFile(headPath, []byte("refs/heads/main"))
}

func Add(repoPath, filePath string) (string, error) {
	absRepo, _ := filepath.Abs(repoPath)

	// read working directory file content
	content, err := utils.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	// write object
	hash, err := object.WriteBlob(absRepo, content)
	if err != nil {
		return "", err
	}

	// update index
	idx, err := index.Load(absRepo)
	if err != nil {
		return "", err
	}
	relPath, err := filepath.Rel(absRepo, filePath)
	if err != nil {
		relPath = filePath
	}

	// if the file is outside the repo, use their relative path from the pwd
	if relPath == "." || relPath == "" {
		relPath = filepath.Base(filePath)
	}
	idx[relPath] = hash
	if err := idx.Save(absRepo); err != nil {
		return "", err
	}

	return hash, nil
}

func EnsureRepo(path string) error {
	repoDir := utils.GetRepoDir(path)
	if !utils.Exists(repoDir) {
		return fmt.Errorf("not a valid arbor repository (or any of the parent directories): .arbor")
	}
	return nil
}

func writeTree(repoPath string) (string, error) {
	idx, err := index.Load(repoPath)
	if err != nil {
		return "", err
	}

	paths := make([]string, 0, len(idx))
	for p := range idx {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	var content []byte
	for _, p := range paths {
		h := idx[p]
		line := fmt.Sprintf("blob %s %s\n", h, p)
		content = append(content, []byte(line)...)
	}

	return object.WriteTree(repoPath, content)
}
