package branch

import (
	"fmt"
	"strings"

	"github.com/matiasmartin00/arbor/internal/refs"
)

var (
	errInvalidBranchName = fmt.Errorf("invalid branch name")
	errBranchExists      = fmt.Errorf("branch already exists")
	errNoCommits         = fmt.Errorf("no commits yet")
)

func CreateBranch(repoPath, name string) (string, error) {
	if strings.Contains(name, "/") {
		return "", errInvalidBranchName
	}

	// get current commit hash
	hash, err := refs.GetRefHash(repoPath)
	if err != nil {
		return "", err
	}

	if len(hash) == 0 {
		return "", errNoCommits
	}

	// check if branch exists
	if refs.ExistsRef(repoPath, name) {
		return "", errBranchExists
	}

	// create ref
	if err := refs.CreateRef(repoPath, name, hash); err != nil {
		return "", err
	}

	return name, nil
}
