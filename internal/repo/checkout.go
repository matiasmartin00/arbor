package repo

import (
	"bufio"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/matiasmartin00/arbor/internal/index"
	"github.com/matiasmartin00/arbor/internal/object"
	"github.com/matiasmartin00/arbor/internal/utils"
)

func Checkout(repoPath, commitHash string) error {
	if len(commitHash) < 4 {
		return fmt.Errorf("commit hash too short")
	}

	// ensure repository is clean
	if err := ensureCleanWorktree(repoPath); err != nil {
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

	// read tree object (our tree format is: "blob <hash> <path>\n")
	treeData, err := object.ReadTree(repoPath, treeHash)
	if err != nil {
		return err
	}

	// build map[path]hash for entire in the tree
	treeMap := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(string(treeData)))
	for scanner.Scan() {
		line := scanner.Text()
		// expecting format: "blob <hash> <path>"
		parts := strings.SplitN(line, " ", 3)
		if len(parts) != 3 {
			return fmt.Errorf("invalid tree entry: %s", line)
		}
		typ := parts[0]
		hash := parts[1]
		path := parts[2]

		if typ != "blob" {
			continue // skip non-blob entries for simplicity
		}

		treeMap[filepath.FromSlash(path)] = hash
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	// apply tree to working directory
	for path, hash := range treeMap {
		// ensure directory exists
		dir := filepath.Dir(path)
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

		// overwrite working file
		if err := utils.WriteFile(path, blobData); err != nil {
			return err
		}
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
	if err := updateRef(repoPath, commitHash); err != nil {
		return err
	}

	return nil
}
