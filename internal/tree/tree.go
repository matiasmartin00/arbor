package tree

import (
	"path/filepath"

	"github.com/matiasmartin00/arbor/internal/index"
	"github.com/matiasmartin00/arbor/internal/object"
	"github.com/matiasmartin00/arbor/internal/refs"
)

func WriteTree(repoPath string) (object.ObjectHash, error) {
	idx, err := index.Load(repoPath)
	if err != nil {
		return nil, err
	}

	// build a map of path -> hash
	entries := make(map[string]object.ObjectHash, len(idx))
	for p, ie := range idx {
		entries[filepath.ToSlash(p)] = ie.Hash
	}

	return object.WriteTree(repoPath, entries)
}

func GetHeadTreeMap(repoPath string) (map[string]object.ObjectHash, error) {
	m := map[string]object.ObjectHash{}

	commitHash, err := refs.GetRefHash(repoPath)
	if err != nil {
		return nil, err
	}

	// if commit hash is nil, no head yet
	if commitHash == nil {
		return m, nil
	}

	commit, err := object.ReadCommit(repoPath, commitHash)
	if err != nil {
		return nil, err
	}

	tree, err := object.ReadTree(repoPath, commit.TreeHash())
	if err != nil {
		return nil, err
	}

	tree.FillPathMap(m)

	return m, nil
}
