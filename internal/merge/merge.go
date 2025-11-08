package merge

import (
	"bufio"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/matiasmartin00/arbor/internal/add"
	"github.com/matiasmartin00/arbor/internal/branch"
	"github.com/matiasmartin00/arbor/internal/commit"
	"github.com/matiasmartin00/arbor/internal/object"
	"github.com/matiasmartin00/arbor/internal/refs"
	"github.com/matiasmartin00/arbor/internal/tree"
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

	baseTree := map[string]object.ObjectHash{}
	headTree := map[string]object.ObjectHash{}
	targetTree := map[string]object.ObjectHash{}

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

	if err := tree.FillPathMapFromTree(repoPath, baseCommit.TreeHash(), "", baseTree); err != nil {
		return err
	}

	if err := tree.FillPathMapFromTree(repoPath, headCommit.TreeHash(), "", headTree); err != nil {
		return err
	}

	if err := tree.FillPathMapFromTree(repoPath, targetCommit.TreeHash(), "", targetTree); err != nil {
		return err
	}

	conflicts := []string{}
	merged := map[string]object.ObjectHash{}

	// union all paths
	allPaths := map[string]struct{}{}
	for p := range baseTree {
		allPaths[p] = struct{}{}
	}
	for p := range headTree {
		allPaths[p] = struct{}{}
	}
	for p := range targetTree {
		allPaths[p] = struct{}{}
	}

	for path := range allPaths {
		base := baseTree[path]
		head := headTree[path]
		target := targetTree[path]

		switch {
		case head.Equals(target):
			merged[path] = head
		case base.Equals(head):
			merged[path] = target // changed only in target
		case base.Equals(target):
			merged[path] = head // changed only in head
		default:
			// conficlt
			headLines, _ := readBlobContent(repoPath, head)
			targetLines, _ := readBlobContent(repoPath, target)
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
		data, err := object.ReadBlob(repoPath, hash)
		if err != nil {
			return err
		}
		if err := utils.CreateDir(filepath.Dir(p)); err != nil {
			return err
		}
		if err := utils.WriteFile(p, []byte(data)); err != nil {
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

// it is duplicated with diff, TODO: REFACTOR THIS.
// readBlobContent returns []string lines for a blob hash or file path.
// if src looks like a blob hash (40 hex) it reads object; otherwise it treats src as filesystem path.
func readBlobContent(repoPath string, hash object.ObjectHash) ([]string, error) {
	blob, err := object.ReadBlob(repoPath, hash)
	if err != nil {
		return nil, err
	}

	return splitLines(blob)
}

func splitLines(data []byte) ([]string, error) {
	var out []string
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		out = append(out, scanner.Text())
	}

	if scanner.Err() != nil {
		return []string{}, scanner.Err()
	}

	return out, nil
}
