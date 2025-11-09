package status

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/matiasmartin00/arbor/internal/index"
	"github.com/matiasmartin00/arbor/internal/object"
	"github.com/matiasmartin00/arbor/internal/tree"
	"github.com/matiasmartin00/arbor/internal/utils"
)

type StatusDetail struct {
	ToBeCommitted []string
	NotStaged     []string
	Untracked     []string
}

func fileBlobHash(path string) (object.ObjectHash, error) {
	data, err := utils.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return object.NewHashBlob(data)
}

func Status(repoPath string) (StatusDetail, error) {

	idx, err := index.Load(repoPath)
	if err != nil {
		return StatusDetail{}, err
	}

	// headMap files
	headMap, err := tree.GetHeadTreeMap(repoPath)
	if err != nil {
		return StatusDetail{}, err
	}

	// changes to be committed: index vs head tree
	toBeCommitted := []string{}
	for p, ih := range idx {
		hh, ok := headMap[p]
		if !ok {
			toBeCommitted = append(toBeCommitted, fmt.Sprintf("new file: %s", p))
			continue
		}

		if ih.NotEquals(hh) {
			toBeCommitted = append(toBeCommitted, fmt.Sprintf("modified: %s", p))
		}
	}

	// dectect deleted staged, (in the head but not in the index)
	for p := range headMap {
		if _, ok := idx[p]; !ok {
			toBeCommitted = append(toBeCommitted, fmt.Sprintf("deleted: %s", p))
		}
	}

	// changes not staged for commit: workdir vs index
	notStaged := []string{}
	for p, ih := range idx {
		if _, err := os.Stat(p); err != nil {
			if os.IsNotExist(err) {
				notStaged = append(notStaged, fmt.Sprintf("deleted: %s"))
				continue
			}
			return StatusDetail{}, err
		}

		curHash, err := fileBlobHash(p)
		if err != nil {
			return StatusDetail{}, err
		}

		if curHash.NotEquals(ih) {
			notStaged = append(notStaged, fmt.Sprintf("modified: %s", p))
		}
	}

	// untrucked
	untracked := []string{}
	err = filepath.WalkDir(repoPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		// ignorar .arbor
		if d.IsDir() && utils.IsRepoDir(d.Name()) {
			return filepath.SkipDir
		}

		if d.IsDir() {
			return nil
		}

		rel := filepath.ToSlash(path)

		if _, ok := idx[rel]; ok {
			return nil
		}

		untracked = append(untracked, rel)
		return nil
	})

	if err != nil {
		return StatusDetail{}, err
	}

	return StatusDetail{
		ToBeCommitted: toBeCommitted,
		NotStaged:     notStaged,
		Untracked:     untracked,
	}, nil
}
