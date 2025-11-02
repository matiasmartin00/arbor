package merge

import (
	"bufio"
	"fmt"
	"os"
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

		fmt.Printf("Fast-forward merge.")
		return nil
	}

	// otherwise, 3-way merge
	baseHash, err := findCommonAncestor(repoPath, headHash, targetHash)
	if err != nil {
		return err
	}

	baseTree := map[string]string{}
	headTree := map[string]string{}
	targetTree := map[string]string{}

	baseTreeHash, err := tree.GetTreeHashFromCommitHash(repoPath, baseHash)
	if err != nil {
		return err
	}

	headTreeHash, err := tree.GetTreeHashFromCommitHash(repoPath, headHash)
	if err != nil {
		return err
	}

	targetTreeHash, err := tree.GetTreeHashFromCommitHash(repoPath, targetHash)
	if err != nil {
		return err
	}

	if err := tree.FillPathMapFromTree(repoPath, baseTreeHash, "", baseTree); err != nil {
		return err
	}

	if err := tree.FillPathMapFromTree(repoPath, headTreeHash, "", headTree); err != nil {
		return err
	}

	if err := tree.FillPathMapFromTree(repoPath, targetTreeHash, "", targetTree); err != nil {
		return err
	}

	conflicts := []string{}
	merged := map[string]string{}

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
		case head == target:
			merged[path] = head
		case base == head:
			merged[path] = target // changed only in target
		case base == target:
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

func isAncestorCommit(repoPath, maybeAncestor, targetHash string) (bool, error) {
	if len(maybeAncestor) == 0 || len(targetHash) == 0 {
		return false, nil
	}

	toVisit := []string{targetHash}
	seen := map[string]struct{}{}
	for len(toVisit) > 0 {
		c := toVisit[0]
		toVisit = toVisit[1:]
		if c == maybeAncestor {
			return true, nil
		}

		if _, ok := seen[c]; ok {
			continue
		}

		seen[c] = struct{}{}
		data, err := object.ReadCommit(repoPath, c)
		if err != nil {
			continue
		}

		for _, l := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(l, "parent ") {
				toVisit = append(toVisit, strings.TrimPrefix(l, "parent "))
			}
		}
	}
	return false, nil
}

func findCommonAncestor(repoPath, a, b string) (string, error) {
	if len(a) == 0 || len(b) == 0 {
		return "", nil
	}

	ancA := allAncestors(repoPath, a)
	ancB := allAncestors(repoPath, b)

	for k := range ancA {
		if _, ok := ancB[k]; ok {
			return k, nil
		}
	}

	return "", nil
}

func allAncestors(repoPath, start string) map[string]struct{} {
	m := map[string]struct{}{}
	queue := []string{start}
	for len(queue) > 0 {
		c := queue[0]
		queue = queue[1:]
		if _, ok := m[c]; ok {
			continue
		}

		m[c] = struct{}{}
		data, err := object.ReadCommit(repoPath, c)
		if err != nil {
			continue
		}

		for _, l := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(l, "parent ") {
				queue = append(queue, strings.TrimPrefix(l, "parent "))
			}
		}
	}

	return m
}

// it is duplicated with diff, TODO: REFACTOR THIS.
// readBlobContent returns []string lines for a blob hash or file path.
// if src looks like a blob hash (40 hex) it reads object; otherwise it treats src as filesystem path.
func readBlobContent(repoPath, pathOrHash string) ([]string, error) {

	// heuristic: if src length==40 and object exists, read it
	if len(pathOrHash) == 40 {
		blob, err := object.ReadBlob(repoPath, pathOrHash)
		if err != nil {
			return nil, err
		}

		return splitLines(blob)
	}

	data, err := utils.ReadFile(pathOrHash)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	return splitLines(data)
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
