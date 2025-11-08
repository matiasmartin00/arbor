package diff

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/matiasmartin00/arbor/internal/index"
	"github.com/matiasmartin00/arbor/internal/object"
	"github.com/matiasmartin00/arbor/internal/tree"
	"github.com/matiasmartin00/arbor/internal/utils"
)

const (
	eqLine  = " "
	rmLine  = "-"
	addLine = "+"
)

func readFileContent(repoPath, path string) ([]string, error) {
	data, err := utils.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	return splitLines(data)
}

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

// unifiedDiff produces a simple unified diff between a and b.
// it uses LCS to compute inserts/deletes. Context lines are not collapsed.
func unifiedDiff(aLines, bLines []string) []string {
	n, m := len(aLines), len(bLines)
	dp := make([][]int, n+1)

	for i := 0; i <= n; i++ {
		dp[i] = make([]int, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			if aLines[i] == bLines[j] {
				dp[i][j] = dp[i+1][j+1] + 1
				continue
			}

			if dp[i+1][j] >= dp[i][j+1] {
				dp[i][j] = dp[i+1][j]
				continue
			}

			dp[i][j] = dp[i][j+1]
		}
	}

	out := []string{}
	i, j := 0, 0
	for i < n || j < m {

		if i < n && j < m && aLines[i] == bLines[j] {
			line := fmt.Sprintf("%s%s", eqLine, aLines[i])
			out = append(out, line)
			i++
			j++
			continue
		}

		if j == m || (i < n && dp[i+1][j] >= dp[i][j+1]) {
			line := fmt.Sprintf("%s%s", rmLine, aLines[i])
			out = append(out, line)
			i++
			continue
		}

		if i == n || (j < m && dp[i][j+1] >= dp[i+1][j]) {
			line := fmt.Sprintf("%s%s", addLine, bLines[j])
			out = append(out, line)
			j++
			continue
		}
	}

	return out
}

// DiffWorktreeVsIndex diffs the working copy vs the index and prints to stdout.
func DiffWorktreeVsIndex(repoPath string, paths []string) error {
	idx, err := index.Load(repoPath)
	if err != nil {
		return err
	}

	targets := pathsToTargetMap(paths)

	for p, blobHash := range idx {
		// if paths filter given, skip others
		if len(targets) > 0 {
			if _, ok := targets[p]; !ok {
				continue
			}
		}

		// read index content via blob
		indexLines, err := readBlobContent(repoPath, blobHash)
		if err != nil {
			return err
		}

		// read worktree content from file
		workPath := filepath.FromSlash(p)
		workLines, err := readFileContent(repoPath, workPath)
		if err != nil {
			return err
		}
		if equalLines(indexLines, workLines) {
			continue
		}
		fmt.Printf("diff -- %s (workdir vs index)\n", p)
		for _, l := range unifiedDiff(indexLines, workLines) {
			fmt.Println(l)
		}
		fmt.Println()
	}

	return nil
}

// DiffIndexVsHead diffs the index vs HEAD tree (staged changes).
func DiffIndexVsHead(repoPath string, paths []string) error {
	idx, err := index.Load(repoPath)
	if err != nil {
		return err
	}

	headMap, err := tree.GetHeadTreeMap(repoPath)
	if err != nil {
		return err
	}

	targets := pathsToTargetMap(paths)

	// union of keys from idx and headMap
	seen := map[string]struct{}{}
	for p := range idx {
		seen[p] = struct{}{}
	}

	for p := range headMap {
		seen[p] = struct{}{}
	}

	for p := range seen {
		if len(targets) > 0 {
			if _, ok := targets[p]; !ok {
				continue
			}
		}

		ih := idx[p]
		hh := headMap[p]

		var idxLines, headLines []string

		if ih != nil {
			idxLines, err = readBlobContent(repoPath, ih)
			if err != nil {
				return err
			}
		} else {
			idxLines = []string{}
		}

		if hh != nil {
			headLines, err = readBlobContent(repoPath, hh)
			if err != nil {
				return err
			}
		} else {
			headLines = []string{}
		}

		if equalLines(idxLines, headLines) {
			continue
		}

		fmt.Printf("diff -- %s (index vs HEAD)\n", p)
		for _, l := range unifiedDiff(headLines, idxLines) {
			fmt.Println(l)
		}

		fmt.Println()
	}

	return nil
}

// diffCommits diff two commits by comparings their trees
func DiffCommits(repoPath, commitA, commitB string, paths []string) error {
	mapA := map[string]object.ObjectHash{}
	mapB := map[string]object.ObjectHash{}

	if len(commitA) > 0 {
		hash, err := object.NewObjectHash(commitA)
		if err != nil {
			return err
		}

		commit, err := object.ReadCommit(repoPath, hash)
		if err != nil {
			return err
		}

		if err := tree.FillPathMapFromTree(repoPath, commit.TreeHash(), "", mapA); err != nil {
			return err
		}
	}

	if len(commitB) > 0 {
		hash, err := object.NewObjectHash(commitB)
		if err != nil {
			return err
		}

		commit, err := object.ReadCommit(repoPath, hash)
		if err != nil {
			return err
		}

		if err := tree.FillPathMapFromTree(repoPath, commit.TreeHash(), "", mapB); err != nil {
			return err
		}
	}

	// union of keys
	targets := pathsToTargetMap(paths)

	seen := map[string]struct{}{}
	for p := range mapA {
		seen[p] = struct{}{}
	}

	for p := range mapA {
		seen[p] = struct{}{}
	}

	for p := range seen {
		if len(targets) > 0 {
			if _, ok := targets[p]; !ok {
				continue
			}
		}

		aHash := mapA[p]
		bHash := mapB[p]
		var aLines, bLines []string
		if aHash != nil {
			blob, err := readBlobContent(repoPath, aHash)
			if err != nil {
				return nil
			}

			aLines = blob
		} else {
			aLines = []string{}
		}

		if bHash != nil {
			blob, err := readBlobContent(repoPath, bHash)
			if err != nil {
				return nil
			}

			bLines = blob
		} else {
			bLines = []string{}
		}

		if equalLines(aLines, bLines) {
			continue
		}

		fmt.Printf("diff -- %s (commit %s -> %s)\n", p, commitA, commitB)
		for _, l := range unifiedDiff(aLines, bLines) {
			fmt.Println(l)
		}
		fmt.Println()
	}

	return nil
}

func equalLines(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func pathsToTargetMap(paths []string) map[string]struct{} {
	targets := map[string]struct{}{}
	for _, p := range paths {
		targets[filepath.ToSlash(p)] = struct{}{}
	}
	return targets
}
