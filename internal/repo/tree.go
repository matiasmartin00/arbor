package repo

import (
	"fmt"
	"sort"

	"github.com/matiasmartin00/arbor/internal/index"
	"github.com/matiasmartin00/arbor/internal/object"
)

func writeTree(repoPath string) (string, error) {
	idx, err := index.Load(repoPath)
	if err != nil {
		return "", err
	}

	paths := make([]string, 0, len(idx))
	for p := range idx {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	var content []byte
	for _, p := range paths {
		h := idx[p]
		line := fmt.Sprintf("blob %s %s\n", h, p)
		content = append(content, []byte(line)...)
	}

	return object.WriteTree(repoPath, content)
}
