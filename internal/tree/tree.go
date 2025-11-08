package tree

import (
	"bufio"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/matiasmartin00/arbor/internal/index"
	"github.com/matiasmartin00/arbor/internal/object"
	"github.com/matiasmartin00/arbor/internal/refs"
)

// writeTree builds a recursive tree objects from the index and returns the root tree hash.
// tree format: each line is "blob <hash> <path>" for blobs, and "tree <hash> <path>" for subtrees.
func WriteTree(repoPath string) (object.ObjectHash, error) {
	idx, err := index.Load(repoPath)
	if err != nil {
		return nil, err
	}

	// build a map of path -> hash
	entries := make(map[string]object.ObjectHash, len(idx))
	for p, h := range idx {
		entries[filepath.ToSlash(p)] = h
	}

	return writeTreeRecursive(repoPath, entries, "")
}

func writeTreeRecursive(repoPath string, entries map[string]object.ObjectHash, prefix string) (object.ObjectHash, error) {
	files := make(map[string]object.ObjectHash)
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
			return nil, err
		}

		line := fmt.Sprintf("tree %s %s\n", subtreeHash, name)
		content = append(content, []byte(line)...)
	}

	return object.WriteTree(repoPath, content)
}

// fillPathMapFromTree fills pathMap with entries path->blobHash using the tree recursively.
// paths returned use OS-specific separators.
func FillPathMapFromTree(repoPath string, treeHash object.ObjectHash, basePath string, pathMap map[string]object.ObjectHash) error {
	data, err := object.ReadTree(repoPath, treeHash)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, " ", 3)
		if len(parts) != 3 {
			continue
		}

		typ := parts[0]
		hash, err := object.NewObjectHash(parts[1])
		if err != nil {
			return err
		}

		path := parts[2]

		if typ == "blob" {
			fullPath := filepath.FromSlash(filepath.Join(basePath, path))
			pathMap[fullPath] = hash
			continue
		}

		if typ == "tree" {
			// recurse into subtree
			subBasePath := filepath.Join(basePath, path)
			if err := FillPathMapFromTree(repoPath, hash, subBasePath, pathMap); err != nil {
				return err
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func GetHeadTreeMap(repoPath string) (map[string]object.ObjectHash, error) {
	m := map[string]object.ObjectHash{}

	commitHash, err := refs.GetRefHash(repoPath)
	if err != nil {
		return nil, err
	}

	// if commit hash is nil, no head yet
	if commitHash == nil {
		return m, nil
	}

	commit, err := object.ReadCommit(repoPath, commitHash)
	if err != nil {
		return nil, err
	}

	treeHash := commit.TreeHash()

	if err = FillPathMapFromTree(repoPath, treeHash, "", m); err != nil {
		return nil, err
	}

	return m, nil
}
