package checkout

import (
	"github.com/matiasmartin00/arbor/internal/refs"
	"github.com/matiasmartin00/arbor/internal/worktree"
)

func Checkout(repoPath, commitHashOrRef string) error {
	var commitHash string
	if refs.ExistsRef(".", commitHashOrRef) {
		hash, err := refs.GetRefHashByName(repoPath, commitHashOrRef)
		if err != nil {
			return err
		}

		commitHash = hash
	}

	err := worktree.RestoreCommitWorktree(repoPath, commitHash)
	if err != nil {
		return err
	}

	return refs.UpdateHEAD(repoPath, commitHashOrRef)
}
