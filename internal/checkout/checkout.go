package checkout

import (
	"github.com/matiasmartin00/arbor/internal/object"
	"github.com/matiasmartin00/arbor/internal/refs"
	"github.com/matiasmartin00/arbor/internal/worktree"
)

func Checkout(repoPath, commitHashOrRef string) error {
	if refs.ExistsRef(repoPath, commitHashOrRef) {
		hash, err := refs.GetRefHashByName(repoPath, commitHashOrRef)
		if err != nil {
			return err
		}

		if err := worktree.RestoreCommitWorktree(repoPath, hash); err != nil {
			return err
		}

		return refs.UpdateHEAD(repoPath, commitHashOrRef)
	}

	hash, err := object.NewObjectHash(commitHashOrRef)
	if err != nil {
		return err
	}

	return worktree.RestoreCommitWorktree(repoPath, hash)
}
