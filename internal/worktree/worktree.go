package worktree

import (
	"path/filepath"

	"github.com/matiasmartin00/arbor/internal/index"
	"github.com/matiasmartin00/arbor/internal/object"
	"github.com/matiasmartin00/arbor/internal/repo"
	"github.com/matiasmartin00/arbor/internal/utils"
)

func RestoreCommitWorktree(repoPath string, commitHash object.ObjectHash) error {
	// ensure repository is clean
	if err := repo.EnsureCleanWorktree(repoPath); err != nil {
		return err
	}

	// read commit object
	commit, err := object.ReadCommit(repoPath, commitHash)
	if err != nil {
		return err
	}

	tree, err := object.ReadTree(repoPath, commit.TreeHash())
	if err != nil {
		return err
	}

	// apply tree to working directory
	err = applyTree(repoPath, tree)
	if err != nil {
		return err
	}

	// build map[path]hash for entire in the tree
	treeMap := make(map[string]object.ObjectHash)
	tree.FillPathMap(treeMap)

	// remove tracked that are in the index but not in the tree
	idx, err := index.Load(repoPath)
	if err != nil {
		return err
	}

	for path := range idx {
		if _, ok := treeMap[path]; !ok {
			// file is in index but not in tree, remove it
			if err := utils.RemoveFile(path); err != nil {
				return err
			}
			delete(idx, path)
		}
	}

	// update index to match tree
	for path, hash := range treeMap {
		idx.AddEntry(path, hash)
	}

	if err := idx.Save(repoPath); err != nil {
		return err
	}

	return nil
}

// applyTree writes files and directories from the given tree object into basePath. (relative to repo root)
func applyTree(repoPath string, tree object.Tree) error {
	for _, bl := range tree.Blobs() {
		// ensure directory exists, the bl.File contains the full path with /
		dir := filepath.Dir(bl.File)
		if dir != "." {
			if err := utils.CreateDir(dir); err != nil {
				return err
			}
		}

		// read blob data
		blob, err := object.ReadBlob(repoPath, bl.Hash)
		if err != nil {
			return err
		}

		// write file
		if err := utils.WriteFile(bl.File, blob.Data()); err != nil {
			return err
		}
	}

	for _, st := range tree.SubTrees() {
		// create directory and recurse
		if err := utils.CreateDir(st.Basepath()); err != nil {
			return err
		}

		if err := applyTree(repoPath, st); err != nil {
			return err
		}
	}

	return nil
}
