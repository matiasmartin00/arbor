package repo

import (
	"fmt"
	"os"

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

// EnsureRepo checks if the given path is a valid arbor repository.
func EnsureRepo(path string) error {
	repoDir := utils.GetRepoDir(path)
	if !utils.Exists(repoDir) {
		return fmt.Errorf("not a valid arbor repository (or any of the parent directories): .arbor")
	}
	return nil
}

// ensureCleanWorktree checks that for every entry in the index the working file matches the indexed blob hash.
// If a file is missing or modified (workdir != index) it returns an error.
func EnsureCleanWorktree(repoPath string) error {
	index, err := index.Load(repoPath)
	if err != nil {
		return err
	}

	for p, h := range index {
		data, err := utils.ReadFile(p)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("uncommitted changes: file %s is missing (not committed or staged)", p)
			}
			return err
		}

		curHash := object.HashBlob(data)
		if curHash != h {
			return fmt.Errorf("uncommitted changes: file %s has been modified (not committed or staged)", p)
		}
	}

	return nil
}
