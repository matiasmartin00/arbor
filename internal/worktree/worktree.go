package worktree

import (
	"bufio"
	"path/filepath"
	"strings"

	"github.com/matiasmartin00/arbor/internal/index"
	"github.com/matiasmartin00/arbor/internal/object"
	"github.com/matiasmartin00/arbor/internal/repo"
	"github.com/matiasmartin00/arbor/internal/tree"
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

	treeHash := commit.TreeHash()

	// apply tree to working directory
	err = applyTree(repoPath, treeHash, "")
	if err != nil {
		return err
	}

	// build map[path]hash for entire in the tree
	treeMap := make(map[string]object.ObjectHash)
	if err := tree.FillPathMapFromTree(repoPath, treeHash, "", treeMap); err != nil {
		return err
	}

	// remove tracked that are in the index but not in the tree
	index, err := index.Load(repoPath)
	if err != nil {
		return err
	}

	for path := range index {
		if _, ok := treeMap[path]; !ok {
			// file is in index but not in tree, remove it
			if err := utils.RemoveFile(path); err != nil {
				return err
			}
			delete(index, path)
		}
	}

	// update index to match tree
	for path, hash := range treeMap {
		index[path] = hash
	}

	if err := index.Save(repoPath); err != nil {
		return err
	}

	return nil
}

// applyTree writes files and directories from the given tree object into basePath. (relative to repo root)
func applyTree(repoPath string, treeHash object.ObjectHash, basePath string) error {
	data, err := object.ReadTree(repoPath, treeHash)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, " ", 3)
		if len(parts) != 3 {
			continue
		}

		typ := parts[0]
		hash, err := object.NewObjectHash(parts[1])
		if err != nil {
			return err
		}

		path := parts[2]

		targetPath := filepath.FromSlash(filepath.Join(basePath, path))
		if typ == "blob" {
			// ensure directory exists
			dir := filepath.Dir(targetPath)
			if dir != "." {
				if err := utils.CreateDir(dir); err != nil {
					return err
				}
			}

			// read blob data
			blobData, err := object.ReadBlob(repoPath, hash)
			if err != nil {
				return err
			}

			// write file
			if err := utils.WriteFile(targetPath, blobData); err != nil {
				return err
			}

			continue
		}

		if typ == "tree" {
			// create directory and recurse
			if err := utils.CreateDir(targetPath); err != nil {
				return err
			}

			if err := applyTree(repoPath, hash, targetPath); err != nil {
				return err
			}
		}
	}
	return nil
}
