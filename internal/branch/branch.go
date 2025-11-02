package branch

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/matiasmartin00/arbor/internal/refs"
	"github.com/matiasmartin00/arbor/internal/utils"
)

var (
	errInvalidBranchName = fmt.Errorf("invalid branch name")
	errBranchExists      = fmt.Errorf("branch already exists")
	errNoCommits         = fmt.Errorf("no commits yet")
)

func CreateBranch(repoPath, name string) error {
	if strings.Contains(name, "/") {
		return errInvalidBranchName
	}

	// get current commit hash
	hash, err := refs.GetRefHash(repoPath)
	if err != nil {
		return err
	}

	if len(hash) == 0 {
		return errNoCommits
	}

	// check if branch exists
	if refs.ExistsRef(repoPath, name) {
		return errBranchExists
	}

	// create ref
	if err := refs.CreateRef(repoPath, name, hash); err != nil {
		return err
	}

	return nil
}

// listBranches returns a list of branch names and mark the current one with '*'
func ListBranches(repoPath string) ([]string, error) {
	refsDir := utils.GetRefsDir(repoPath)

	var branches []string
	err := filepath.WalkDir(refsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}

		if d.IsDir() {
			return nil
		}

		rel, _ := filepath.Rel(refsDir, path)
		branches = append(branches, filepath.ToSlash(rel))
		return nil
	})

	if err != nil {
		return nil, err
	}

	current, err := GetCurrentBranch(repoPath)
	if err != nil {
		return nil, err
	}

	for i, b := range branches {
		if b == current {
			branches[i] = "* " + b
		}
	}

	return branches, nil
}

func GetCurrentBranch(repoPath string) (string, error) {
	headRaw, err := refs.GetHEAD(repoPath)
	if err != nil {
		return "", err
	}

	if !refs.IsRef(headRaw) {
		return "", nil
	}

	return strings.TrimPrefix(headRaw, "refs/heads/"), nil
}
