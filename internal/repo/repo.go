package repo

import (
	"fmt"

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

func EnsureRepo(path string) error {
	repoDir := utils.GetRepoDir(path)
	if !utils.Exists(repoDir) {
		return fmt.Errorf("not a valid arbor repository (or any of the parent directories): .arbor")
	}
	return nil
}
