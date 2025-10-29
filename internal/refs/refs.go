package refs

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/matiasmartin00/arbor/internal/utils"
)

const headFile = "HEAD"
const refsDir = "refs/heads"

func readHEAD(repoPath string) (string, error) {
	headPath := getHeadPath(repoPath)
	data, err := utils.ReadFile(headPath)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func GetRefHash(repoPath string) (string, error) {
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

func UpdateRef(repoPath, hash string) error {
	head, err := readHEAD(repoPath)
	if err != nil {
		return err
	}

	refPath := filepath.Join(utils.GetRepoDir(repoPath), head)

	return utils.WriteFile(refPath, []byte(hash+"\n"))
}

func UpdateHEAD(repoPath, ref string) error {
	headPath := getHeadPath(repoPath)
	return utils.WriteFile(headPath, []byte(ref+"\n"))
}

func IsRef(head string) bool {
	return strings.HasPrefix(head, "refs/")
}

func ExistsRef(repoPath, ref string) bool {
	refPath := filepath.Join(utils.GetRepoDir(repoPath), refsDir, ref)
	return utils.Exists(refPath)
}

func CreateRef(repoPath, ref, hash string) error {
	refPath := filepath.Join(utils.GetRepoDir(repoPath), refsDir, ref)
	return utils.WriteFile(refPath, []byte(hash+"\n"))
}

func getHeadPath(path string) string {
	return filepath.Join(utils.GetRepoDir(path), headFile)
}
