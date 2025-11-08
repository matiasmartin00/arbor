package merge

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/matiasmartin00/arbor/internal/add"
	"github.com/matiasmartin00/arbor/internal/branch"
	"github.com/matiasmartin00/arbor/internal/commit"
	"github.com/matiasmartin00/arbor/internal/object"
	"github.com/matiasmartin00/arbor/internal/refs"
	"github.com/matiasmartin00/arbor/internal/utils"
	"github.com/matiasmartin00/arbor/internal/worktree"
)

func Merge(repoPath, branchName string) error {
	currentBranch, err := branch.GetCurrentBranch(repoPath)
	if err != nil {
		return err
	}

	if branchName == currentBranch {
		return fmt.Errorf("cannot merge a branch into itself")
	}

	headHash, err := refs.GetRefHash(repoPath)
	if err != nil {
		return err
	}

	targetHash, err := refs.GetRefHashByName(repoPath, branchName)
	if err != nil {
		return err
	}

	// detect fast-forward
	ff, err := isAncestorCommit(repoPath, headHash, targetHash)
	if err != nil {
		return err
	}

	// fast-forward
	if ff {

		if err := worktree.RestoreCommitWorktree(repoPath, targetHash); err != nil {
			return err
		}

		if err := refs.UpdateRef(repoPath, targetHash); err != nil {
			return err
		}

		fmt.Printf("Fast-forward merge.\n")
		return nil
	}

	// otherwise, 3-way merge
	baseHash, err := findCommonAncestor(repoPath, headHash, targetHash)
	if err != nil {
		return err
	}

	baseTreePathMap := map[string]object.ObjectHash{}
	headTreePathMap := map[string]object.ObjectHash{}
	targetTreePathMap := map[string]object.ObjectHash{}

	baseCommit, err := object.ReadCommit(repoPath, baseHash)
	if err != nil {
		return err
	}

	headCommit, err := object.ReadCommit(repoPath, headHash)
	if err != nil {
		return err
	}

	targetCommit, err := object.ReadCommit(repoPath, targetHash)
	if err != nil {
		return err
	}

	baseTree, err := object.ReadTree(repoPath, baseCommit.TreeHash())
	if err != nil {
		return err
	}
	baseTree.FillPathMap(baseTreePathMap)

	headTree, err := object.ReadTree(repoPath, headCommit.TreeHash())
	if err != nil {
		return err
	}
	headTree.FillPathMap(headTreePathMap)

	targetTree, err := object.ReadTree(repoPath, targetCommit.TreeHash())
	if err != nil {
		return err
	}
	targetTree.FillPathMap(targetTreePathMap)

	conflicts := []string{}
	merged := map[string]object.ObjectHash{}

	// union all paths
	allPaths := map[string]struct{}{}
	for p := range baseTreePathMap {
		allPaths[p] = struct{}{}
	}
	for p := range headTreePathMap {
		allPaths[p] = struct{}{}
	}
	for p := range targetTreePathMap {
		allPaths[p] = struct{}{}
	}

	for path := range allPaths {
		base := baseTreePathMap[path]
		head := headTreePathMap[path]
		target := targetTreePathMap[path]

		switch {
		case head.Equals(target):
			merged[path] = head
		case base.Equals(head):
			merged[path] = target // changed only in target
		case base.Equals(target):
			merged[path] = head // changed only in head
		default:
			// conficlt
			headBlob, err := object.ReadBlob(repoPath, head)
			if err != nil {
				return err
			}
			headLines, err := headBlob.SplitLines()
			if err != nil {
				return err
			}

			targetBlob, err := object.ReadBlob(repoPath, target)
			if err != nil {
				return err
			}
			targetLines, err := targetBlob.SplitLines()
			if err != nil {
				return err
			}

			content := "<<<<<<< HEAD\n" + strings.Join(headLines, "\n") + "\n=======\n" + strings.Join(targetLines, "\n") + "\n>>>>>>> " + branchName + "\n"
			tmpFile := filepath.Join(repoPath, path)
			if err := utils.CreateDir(filepath.Dir(tmpFile)); err != nil {
				return err
			}
			if err := utils.WriteFile(tmpFile, []byte(content)); err != nil {
				return err
			}
			conflicts = append(conflicts, path)
		}
	}

	for p, hash := range merged {
		blob, err := object.ReadBlob(repoPath, hash)
		if err != nil {
			return err
		}
		if err := utils.CreateDir(filepath.Dir(p)); err != nil {
			return err
		}
		if err := utils.WriteFile(p, blob.Data()); err != nil {
			return err
		}
	}

	add.Add(repoPath, []string{"."})

	if len(conflicts) > 0 {
		fmt.Println("Merge completed with conflicts:")
		for _, c := range conflicts {
			fmt.Println(" -", c)
		}
		fmt.Println("Resolve conflicts and run `arbor commit ...` to finalize merge")
		return nil
	}

	// auto commit merge
	msg := fmt.Sprintf("Merge branch '%s' into '%s'", branchName, currentBranch)
	mergeHash, err := commit.Commit(repoPath, msg)
	if err != nil {
		return err
	}
	fmt.Println("Merge commit created:", mergeHash)
	return nil
}

func isAncestorCommit(repoPath string, maybeAncestor, targetHash object.ObjectHash) (bool, error) {
	if maybeAncestor == nil || targetHash == nil {
		return false, nil
	}

	toVisit := []object.ObjectHash{targetHash}
	seen := map[object.ObjectHash]struct{}{}
	for len(toVisit) > 0 {
		c := toVisit[0]
		toVisit = toVisit[1:]
		if maybeAncestor.Equals(c) {
			return true, nil
		}

		if _, ok := seen[c]; ok {
			continue
		}

		seen[c] = struct{}{}
		commit, err := object.ReadCommit(repoPath, object.ObjectHash(c))
		if err != nil {
			continue
		}

		if commit.ParentHash() != nil {
			toVisit = append(toVisit, commit.ParentHash())
		}
	}
	return false, nil
}

func findCommonAncestor(repoPath string, a, b object.ObjectHash) (object.ObjectHash, error) {
	if a == nil || b == nil {
		return nil, nil
	}

	ancA := allAncestors(repoPath, a)
	ancB := allAncestors(repoPath, b)

	for k, v := range ancA {
		if _, ok := ancB[k]; ok {
			return v, nil
		}
	}

	return nil, fmt.Errorf("ancestor not found")
}

func allAncestors(repoPath string, start object.ObjectHash) map[string]object.ObjectHash {
	m := make(map[string]object.ObjectHash)
	queue := []object.ObjectHash{start}
	for len(queue) > 0 {
		c := queue[0]
		queue = queue[1:]
		if _, ok := m[c.String()]; ok {
			continue
		}

		m[c.String()] = c
		data, err := object.ReadCommit(repoPath, c)
		if err != nil {
			continue
		}

		if data.ParentHash() != nil {
			queue = append(queue, data.ParentHash())
		}
	}

	return m
}
