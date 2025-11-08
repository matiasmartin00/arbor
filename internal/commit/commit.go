package commit

import (
	"github.com/matiasmartin00/arbor/internal/object"
	"github.com/matiasmartin00/arbor/internal/refs"
	"github.com/matiasmartin00/arbor/internal/tree"
)

const (
	HeaderTree      = "tree"
	HeaderParent    = "parent"
	HeaderAuthor    = "author"
	HeaderCommitter = "committer"
)

func Commit(repoPath, message string) (object.ObjectHash, error) {
	// write tree
	treeHash, err := tree.WriteTree(repoPath)

	if err != nil {
		return nil, err
	}

	// get parent
	parentHash, err := refs.GetRefHash(repoPath)
	if err != nil {
		return nil, err
	}

	// write commit object
	commitHash, err := object.WriteCommit(repoPath, treeHash, parentHash, message)
	if err != nil {
		return nil, err
	}

	// update ref
	if err := refs.UpdateRef(repoPath, commitHash); err != nil {
		return nil, err
	}

	return commitHash, nil
}
