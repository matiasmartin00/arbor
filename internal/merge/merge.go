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

type mergeType int

const (
	fastForward mergeType = iota
	threeWay
)

type MergeDetail struct {
	OriginBranch string
	TargetBranch string
	Type         mergeType
	Conflicts    []string
	Merged       []string
	CommitHash   object.ObjectHash
}

func (mt mergeType) String() string {
	types := []string{"Fast-Forward", "3-Way"}
	if mt < 0 || int(mt) >= len(types) {
		return ""
	}
	return types[int(mt)]
}

func (md MergeDetail) IsFastForward() bool {
	return md.Type == fastForward
}

func Merge(repoPath, branchName string) (MergeDetail, error) {
	currentBranch, err := branch.GetCurrentBranch(repoPath)
	if err != nil {
		return MergeDetail{}, err
	}

	if branchName == currentBranch {
		return MergeDetail{}, fmt.Errorf("cannot merge a branch into itself")
	}

	headHash, err := refs.GetRefHash(repoPath)
	if err != nil {
		return MergeDetail{}, err
	}

	targetHash, err := refs.GetRefHashByName(repoPath, branchName)
	if err != nil {
		return MergeDetail{}, err
	}

	ff, err := tryFastForwardMerge(repoPath, headHash, targetHash)
	if err != nil {
		return MergeDetail{}, err
	}

	if ff {
		return MergeDetail{
			OriginBranch: branchName,
			TargetBranch: currentBranch,
			Type:         fastForward,
			CommitHash:   targetHash,
		}, err
	}

	return threeWayMerge(repoPath, branchName, currentBranch, headHash, targetHash)

}

// otherwise, 3-way merge TODO: Implement rollback
func threeWayMerge(repoPath, branchName, currentBranch string, headHash, targetHash object.ObjectHash) (MergeDetail, error) {

	baseHash, err := findCommonAncestor(repoPath, headHash, targetHash)
	if err != nil {
		return MergeDetail{}, err
	}

	baseTreePathMap, err := buildTreePathMap(repoPath, baseHash)
	if err != nil {
		return MergeDetail{}, err
	}

	headTreePathMap, err := buildTreePathMap(repoPath, headHash)
	if err != nil {
		return MergeDetail{}, err
	}

	targetTreePathMap, err := buildTreePathMap(repoPath, targetHash)
	if err != nil {
		return MergeDetail{}, err
	}

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
			writeConflictFile(repoPath, path, branchName, head, target)
			conflicts = append(conflicts, path)
		}
	}

	// write merged files
	err = writeMergedFiles(repoPath, merged)
	if err != nil {
		return MergeDetail{}, nil
	}

	// add to stage area merged and conflict files
	add.Add(repoPath, []string{"."})

	mergedFiles := make([]string, 0, len(merged))
	for k, _ := range merged {
		mergedFiles = append(mergedFiles, k)
	}

	// if we have conflicts don't auto commit
	if len(conflicts) > 0 {
		return MergeDetail{
			Type:      threeWay,
			Conflicts: conflicts,
			Merged:    mergedFiles,
		}, nil
	}

	// auto commit merge
	msg := fmt.Sprintf("Merge branch '%s' into '%s'", branchName, currentBranch)
	mergeHash, err := commit.Commit(repoPath, msg)
	if err != nil {
		return MergeDetail{}, err
	}
	return MergeDetail{
		OriginBranch: branchName,
		TargetBranch: currentBranch,
		Type:         threeWay,
		Conflicts:    conflicts,
		Merged:       mergedFiles,
		CommitHash:   mergeHash,
	}, nil
}

// fast-forward TODO: impl rollback
func tryFastForwardMerge(repoPath string, headHash, targetHash object.ObjectHash) (bool, error) {

	// detect fast-forward
	ff, err := isAncestorCommit(repoPath, headHash, targetHash)
	if err != nil {
		return false, err
	}

	if !ff {
		return false, nil
	}

	if err := worktree.RestoreCommitWorktree(repoPath, targetHash); err != nil {
		return false, err
	}

	if err := refs.UpdateRef(repoPath, targetHash); err != nil {
		return false, err
	}

	return true, nil
}

func writeMergedFiles(repoPath string, merged map[string]object.ObjectHash) error {
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

	return nil
}

func writeConflictFile(repoPath, path, branchName string, head, target object.ObjectHash) error {

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

	return nil
}

func buildTreePathMap(repoPath string, commitHash object.ObjectHash) (map[string]object.ObjectHash, error) {
	treePathMap := map[string]object.ObjectHash{}

	commit, err := object.ReadCommit(repoPath, commitHash)
	if err != nil {
		return nil, err
	}

	tree, err := object.ReadTree(repoPath, commit.TreeHash())
	if err != nil {
		return nil, err
	}

	tree.FillPathMap(treePathMap)

	return treePathMap, nil
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
