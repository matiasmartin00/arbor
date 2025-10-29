package refs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/matiasmartin00/arbor/internal/utils"
)

const (
	headFile = "HEAD"
	refsDir  = "refs/heads"
)

func readHEAD(repoPath string) (string, error) {
	headPath := getHeadPath(repoPath)
	data, err := utils.ReadFile(headPath)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func GetHEAD(repoPath string) (string, error) {
	return readHEAD(repoPath)
}

// GetRefHash returns the commit hash that HEAD points to
func GetRefHash(repoPath string) (string, error) {
	head, err := readHEAD(repoPath)
	if err != nil {
		return "", err
	}

	return getRefHash(repoPath, head)
}

// GetRefHashByName returns the commit hash for a given ref name
func GetRefHashByName(repoPath, ref string) (string, error) {
	if NotExistsRef(repoPath, ref) {
		return "", fmt.Errorf("ref %s does not exist", ref)
	}

	head := filepath.Join(refsDir, ref)
	return getRefHash(repoPath, head)
}

// getRefHash assumes head is like "refs/heads/main"
func getRefHash(repoPath, head string) (string, error) {
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
	return utils.WriteFile(headPath, []byte(filepath.Join(refsDir, ref)+"\n"))
}

func IsRef(head string) bool {
	return strings.HasPrefix(head, "refs/")
}

func ExistsRef(repoPath, ref string) bool {
	refPath := filepath.Join(utils.GetRepoDir(repoPath), refsDir, ref)
	return utils.Exists(refPath)
}

func NotExistsRef(repoPath, ref string) bool {
	return !ExistsRef(repoPath, ref)
}

func CreateRef(repoPath, ref, hash string) error {
	refPath := filepath.Join(utils.GetRepoDir(repoPath), refsDir, ref)
	return utils.WriteFile(refPath, []byte(hash+"\n"))
}

func getHeadPath(path string) string {
	return filepath.Join(utils.GetRepoDir(path), headFile)
}
