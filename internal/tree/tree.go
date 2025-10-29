package tree

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/matiasmartin00/arbor/internal/index"
	"github.com/matiasmartin00/arbor/internal/object"
)

// writeTree builds a recursive tree objects from the index and returns the root tree hash.
// tree format: each line is "blob <hash> <path>" for blobs, and "tree <hash> <path>" for subtrees.
func WriteTree(repoPath string) (string, error) {
	idx, err := index.Load(repoPath)
	if err != nil {
		return "", err
	}

	// build a map of path -> hash
	entries := make(map[string]string, len(idx))
	for p, h := range idx {
		entries[filepath.ToSlash(p)] = h
	}

	return writeTreeRecursive(repoPath, entries, "")
}

func writeTreeRecursive(repoPath string, entries map[string]string, prefix string) (string, error) {
	files := make(map[string]string)
	subdirsSet := make(map[string]struct{})

	for p, h := range entries {
		if !strings.HasPrefix(p, prefix) {
			continue
		}

		rest := strings.TrimPrefix(p, prefix)
		if len(rest) == 0 {
			continue
		}

		parts := strings.SplitN(rest, "/", 2)
		if len(parts) == 1 {
			// it's a file
			files[parts[0]] = h
		} else {
			// it's a subdirectory
			subdirsSet[parts[0]] = struct{}{}
		}
	}

	// prepare deterministic order
	names := make([]string, 0, len(files)+len(subdirsSet))
	for f := range files {
		names = append(names, f)
	}
	for d := range subdirsSet {
		names = append(names, d)
	}
	// sort names
	sort.Strings(names)

	var content []byte
	for _, name := range names {
		if h, ok := files[name]; ok {
			// file
			line := fmt.Sprintf("blob %s %s\n", h, name)
			content = append(content, []byte(line)...)
			continue
		}

		// subdirectory
		subtreeHash, err := writeTreeRecursive(repoPath, entries, prefix+name+"/")
		if err != nil {
			return "", err
		}
		line := fmt.Sprintf("tree %s %s\n", subtreeHash, name)
		content = append(content, []byte(line)...)
	}

	return object.WriteTree(repoPath, content)
}
