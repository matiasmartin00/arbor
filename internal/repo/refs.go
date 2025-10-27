package repo

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/matiasmartin00/arbor/internal/utils"
)

func readHEAD(repoPath string) (string, error) {
	headPath := utils.GetHeadPath(repoPath)
	data, err := utils.ReadFile(headPath)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func getRefHash(repoPath string) (string, error) {
	head, err := readHEAD(repoPath)
	if err != nil {
		return "", err
	}

	// assume it's a ref like "refs/heads/main"
	refPath := filepath.Join(utils.GetRepoDir(repoPath), head)
	data, err := utils.ReadFile(refPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	return strings.TrimSpace(string(data)), nil
}

func updateRef(repoPath, hash string) error {
	head, err := readHEAD(repoPath)
	if err != nil {
		return err
	}

	refPath := filepath.Join(utils.GetRepoDir(repoPath), head)

	return utils.WriteFile(refPath, []byte(hash+"\n"))
}
