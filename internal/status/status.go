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

func fileBlobHash(path string) (string, error) {
	data, err := utils.ReadFile(path)
	if err != nil {
		return "", err
	}
	return object.HashBlob(data), nil
}

func Status(repoPath string) error {

	idx, err := index.Load(repoPath)
	if err != nil {
		return err
	}

	// headMap files
	headMap, err := tree.GetHeadTreeMap(repoPath)
	if err != nil {
		return err
	}

	// changes to be committed: index vs head tree
	toBeCommitted := []string{}
	for p, ih := range idx {
		hh, ok := headMap[p]
		if !ok {
			toBeCommitted = append(toBeCommitted, fmt.Sprintf("new file: %s", p))
			continue
		}

		if ih != hh {
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
			return err
		}

		curHash, err := fileBlobHash(p)
		if err != nil {
			return err
		}

		if curHash != ih {
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
		return err
	}

	if len(toBeCommitted) == 0 {
		fmt.Println("No changes to be committed.")
	} else {
		fmt.Println("Changes to be commited: ")
		for _, l := range toBeCommitted {
			fmt.Printf("  %s\n", l)
		}
	}
	fmt.Printf("\n\n")

	if len(notStaged) == 0 {
		fmt.Println("No changes not staged for commit.")
	} else {
		fmt.Println("Changes not staged for commit: ")
		for _, l := range notStaged {
			fmt.Printf("  %s\n", l)
		}
	}
	fmt.Printf("\n\n")

	if len(untracked) == 0 {
		fmt.Println("No untracked files.")
	} else {
		fmt.Println("Untracked files: ")
		for _, u := range untracked {
			fmt.Printf("  %s\n", u)
		}
	}

	return nil
}
