package checkout

import (
	"bufio"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/matiasmartin00/arbor/internal/index"
	"github.com/matiasmartin00/arbor/internal/object"
	"github.com/matiasmartin00/arbor/internal/refs"
	"github.com/matiasmartin00/arbor/internal/repo"
	"github.com/matiasmartin00/arbor/internal/utils"
)

func Checkout(repoPath, commitHash string) error {
	if len(commitHash) < 4 {
		return fmt.Errorf("commit hash too short")
	}

	// ensure repository is clean
	if err := repo.EnsureCleanWorktree(repoPath); err != nil {
		return err
	}

	// read commit object
	data, err := object.ReadCommit(repoPath, commitHash)
	if err != nil {
		return err
	}

	// parse commit to get tree hash
	parts := strings.SplitN(string(data), "\n\n", 2)
	headers := strings.Split(parts[0], "\n")
	var treeHash string
	for _, header := range headers {
		if strings.HasPrefix(header, "tree ") {
			treeHash = strings.TrimSpace(strings.TrimPrefix(header, "tree "))
			break
		}
	}

	if len(treeHash) == 0 {
		return fmt.Errorf("invalid commit object: no tree found")
	}

	// apply tree to working directory
	err = applyTree(repoPath, treeHash, "")
	if err != nil {
		return err
	}

	// build map[path]hash for entire in the tree
	treeMap := make(map[string]string)
	if err := fillPathMapFromTree(repoPath, treeHash, "", treeMap); err != nil {
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

	// update HEAD to point to this cpmmit (detached HEAD behavior)
	// i will update the current ref with the commit hash
	if err := refs.UpdateRef(repoPath, commitHash); err != nil {
		return err
	}

	return nil
}

// applyTree writes files and directories from the given tree object into basePath. (relative to repo root)
func applyTree(repoPath, treeHash, basePath string) error {
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
		hash := parts[1]
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

// fillPathMapFromTree fills pathMap with entries path->blobHash using the tree recursively.
// paths returned use OS-specific separators.
func fillPathMapFromTree(repoPath, treeHash, basePath string, pathMap map[string]string) error {
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
		hash := parts[1]
		path := parts[2]

		if typ == "blob" {
			fullPath := filepath.FromSlash(filepath.Join(basePath, path))
			pathMap[fullPath] = hash
			continue
		}

		if typ == "tree" {
			// recurse into subtree
			subBasePath := filepath.Join(basePath, path)
			if err := fillPathMapFromTree(repoPath, hash, subBasePath, pathMap); err != nil {
				return err
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
