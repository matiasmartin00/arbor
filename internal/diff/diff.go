package diff

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/matiasmartin00/arbor/internal/index"
	"github.com/matiasmartin00/arbor/internal/object"
	"github.com/matiasmartin00/arbor/internal/tree"
	"github.com/matiasmartin00/arbor/internal/utils"
)

type LineResult int

const (
	EqLine LineResult = iota
	RemovedLine
	AddedLine
)

type LineData struct {
	ALine      string
	BLine      string
	ResultLine string
	Result     LineResult
}

func (ld LineResult) String() string {
	types := []string{" ", "-", "+"}
	if ld < 0 || int(ld) >= len(types) {
		return ""
	}
	return types[int(ld)]
}

type DiffResult struct {
	File  string
	AHash object.ObjectHash
	BHash object.ObjectHash
	Lines []LineData
}

func readFileContent(repoPath, path string) ([]string, error) {
	data, err := utils.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	return object.SplitLines(data)
}

// DiffWorktreeVsIndex diffs the working copy vs the index and return diff result
func DiffWorktreeVsIndex(repoPath string, paths []string) ([]DiffResult, error) {
	idx, err := index.Load(repoPath)
	if err != nil {
		return nil, err
	}

	targets := pathsToTargetMap(paths)

	diffResult := []DiffResult{}
	for p, ie := range idx {
		// if paths filter given, skip others
		if len(targets) > 0 {
			if _, ok := targets[p]; !ok {
				continue
			}
		}

		// if file is a binary don't compare lines, just checks hashes
		if ie.IsBinary {
			workPath := filepath.FromSlash(p)
			data, err := utils.ReadFile(workPath)
			if err != nil {
				return nil, err
			}
			workHash, err := object.NewHashBlob(data)

			if workHash.Equals(ie.Hash) {
				continue
			}

			diffResult = append(diffResult, DiffResult{
				File:  p,
				AHash: ie.Hash,
				BHash: workHash,
				Lines: nil,
			})

			continue
		}

		// read index content via blob
		indexBlob, err := object.ReadBlob(repoPath, ie.Hash)
		if err != nil {
			return nil, err
		}

		indexLines, err := indexBlob.SplitLines()
		if err != nil {
			return nil, err
		}

		// read worktree content from file
		workPath := filepath.FromSlash(p)
		workLines, err := readFileContent(repoPath, workPath)
		if err != nil {
			return nil, err
		}
		if equalLines(indexLines, workLines) {
			continue
		}

		diffResult = append(diffResult, DiffResult{
			File:  p,
			AHash: ie.Hash,
			BHash: nil,
			Lines: unifiedDiff(indexLines, workLines),
		})
	}

	return diffResult, nil
}

// DiffIndexVsHead diffs the index vs HEAD tree (staged changes).
func DiffIndexVsHead(repoPath string, paths []string) ([]DiffResult, error) {
	idx, err := index.Load(repoPath)
	if err != nil {
		return nil, err
	}

	headMap, err := tree.GetHeadTreeMap(repoPath)
	if err != nil {
		return nil, err
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

	diffResult := []DiffResult{}
	for p := range seen {
		if len(targets) > 0 {
			if _, ok := targets[p]; !ok {
				continue
			}
		}

		ih := idx[p]
		hh := headMap[p]

		var idxLines, headLines []string

		// if file is a binary don't compare lines, just checks hashes
		if ih.IsBinary {
			if ih.Hash.Equals(hh) {
				continue
			}

			diffResult = append(diffResult, DiffResult{
				File:  p,
				AHash: hh,
				BHash: ih.Hash,
				Lines: nil,
			})
			continue
		}

		if ih.Hash != nil {
			idxBlob, err := object.ReadBlob(repoPath, ih.Hash)
			if err != nil {
				return nil, err
			}
			idxLines, err = idxBlob.SplitLines()
			if err != nil {
				return nil, err
			}
		} else {
			idxLines = []string{}
		}

		if hh != nil {
			headBlob, err := object.ReadBlob(repoPath, hh)
			if err != nil {
				return nil, err
			}
			headLines, err = headBlob.SplitLines()
			if err != nil {
				return nil, err
			}
		} else {
			headLines = []string{}
		}

		if equalLines(idxLines, headLines) {
			continue
		}

		diffResult = append(diffResult, DiffResult{
			File:  p,
			AHash: hh,
			BHash: ih.Hash,
			Lines: unifiedDiff(headLines, idxLines),
		})
	}

	return diffResult, nil
}

// diffCommits diff two commits by comparings their trees
func DiffCommits(repoPath, commitA, commitB string, paths []string) ([]DiffResult, error) {
	mapA, err := makeTreePathMap(repoPath, commitA)
	if err != nil {
		return nil, err
	}
	mapB, err := makeTreePathMap(repoPath, commitB)
	if err != nil {
		return nil, err
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

	diffResult := []DiffResult{}
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
			blob, err := object.ReadBlob(repoPath, aHash)
			if err != nil {
				return nil, err
			}

			if utils.IsBinary(blob.Data()) {
				if aHash.Equals(bHash) {
					continue
				}

				diffResult = append(diffResult, DiffResult{
					File:  p,
					AHash: aHash,
					BHash: bHash,
					Lines: nil,
				})
				continue
			}

			aLines, err = blob.SplitLines()
			if err != nil {
				return nil, err
			}
		} else {
			aLines = []string{}
		}

		if bHash != nil {
			blob, err := object.ReadBlob(repoPath, bHash)
			if err != nil {
				return nil, err
			}

			if utils.IsBinary(blob.Data()) {
				if bHash.Equals(aHash) {
					continue
				}

				diffResult = append(diffResult, DiffResult{
					File:  p,
					AHash: aHash,
					BHash: bHash,
					Lines: nil,
				})
				continue
			}

			bLines, err = blob.SplitLines()
			if err != nil {
				return nil, err
			}
		} else {
			bLines = []string{}
		}

		if equalLines(aLines, bLines) {
			continue
		}

		diffResult = append(diffResult, DiffResult{
			File:  p,
			AHash: aHash,
			BHash: bHash,
			Lines: unifiedDiff(aLines, bLines),
		})
	}

	return diffResult, nil
}

func makeTreePathMap(repoPath, commitHash string) (map[string]object.ObjectHash, error) {
	m := map[string]object.ObjectHash{}
	if len(commitHash) == 0 {
		return nil, fmt.Errorf("invalid commit '%s'", commitHash)
	}

	hash, err := object.NewObjectHash(commitHash)
	if err != nil {
		return nil, err
	}

	commit, err := object.ReadCommit(repoPath, hash)
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

// unifiedDiff produces a simple unified diff between a and b.
// it uses LCS to compute inserts/deletes. Context lines are not collapsed.
func unifiedDiff(aLines, bLines []string) []LineData {
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

	out := []LineData{}
	i, j := 0, 0
	for i < n || j < m {

		if i < n && j < m && aLines[i] == bLines[j] {
			out = append(out, LineData{
				ALine:      aLines[i],
				BLine:      bLines[j],
				ResultLine: aLines[i],
				Result:     EqLine,
			})
			i++
			j++
			continue
		}

		if j == m || (i < n && dp[i+1][j] >= dp[i][j+1]) {
			var bLine string
			if j < m {
				bLine = bLines[j]
			}
			out = append(out, LineData{
				ALine:      aLines[i],
				BLine:      bLine,
				ResultLine: aLines[i],
				Result:     RemovedLine,
			})
			i++
			continue
		}

		if i == n || (j < m && dp[i][j+1] >= dp[i+1][j]) {
			var aLine string
			if i < n {
				aLine = aLines[i]
			}
			out = append(out, LineData{
				ALine:      aLine,
				BLine:      bLines[j],
				ResultLine: bLines[j],
				Result:     AddedLine,
			})
			j++
			continue
		}
	}

	return out
}
