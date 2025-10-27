package repo

import (
	"path/filepath"

	"github.com/matiasmartin00/arbor/internal/index"
	"github.com/matiasmartin00/arbor/internal/object"
	"github.com/matiasmartin00/arbor/internal/utils"
)

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
