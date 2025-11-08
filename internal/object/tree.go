package object

import (
	"bufio"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

type Tree interface {
	Hash() ObjectHash
	FillPathMap(map[string]ObjectHash)
	Blobs() []blobLine
	SubTrees() []Tree
	Basepath() string
}

type tree struct {
	hash     ObjectHash
	basepath string
	blobs    []blobLine
	trees    []Tree
}

type blobLine struct {
	Hash ObjectHash
	File string
}

func (t *tree) Hash() ObjectHash {
	return t.hash
}

func (t *tree) Blobs() []blobLine {
	return t.blobs
}

func (t *tree) SubTrees() []Tree {
	return t.trees
}

func (t *tree) Basepath() string {
	return t.basepath
}

func (t *tree) FillPathMap(pathMap map[string]ObjectHash) {
	for _, bl := range t.blobs {
		pathMap[bl.File] = bl.Hash
	}

	for _, st := range t.trees {
		st.FillPathMap(pathMap)
	}
}

func ReadTree(repoPath string, hash ObjectHash) (Tree, error) {
	return readRecursiveTree(repoPath, hash, "")
}

// writeTree builds a recursive tree objects from the index and returns the root tree hash.
// tree format: each line is "blob <hash> <path>" for blobs, and "tree <hash> <path>" for subtrees.
func WriteTree(repoPath string, entries map[string]ObjectHash) (ObjectHash, error) {
	return writeRecursiveTree(repoPath, entries, "")
}

func writeRecursiveTree(repoPath string, entries map[string]ObjectHash, basepath string) (ObjectHash, error) {
	files := make(map[string]ObjectHash)
	subdirsSet := make(map[string]struct{})

	for p, h := range entries {
		if !strings.HasPrefix(p, basepath) {
			continue
		}

		rest := strings.TrimPrefix(p, basepath)
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
			line := fmt.Sprintf("%s %s %s\n", BlobType, h, name)
			content = append(content, []byte(line)...)
			continue
		}

		// subdirectory
		subtreeHash, err := writeRecursiveTree(repoPath, entries, basepath+name+"/")
		if err != nil {
			return nil, err
		}

		line := fmt.Sprintf("%s %s %s\n", TreeType, subtreeHash, name)
		content = append(content, []byte(line)...)
	}

	return writeObject(repoPath, content, TreeType)
}

func readRecursiveTree(repoPath string, hash ObjectHash, basepath string) (Tree, error) {
	data, err := readTreeData(repoPath, hash)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	tree := &tree{
		hash:     hash,
		blobs:    []blobLine{},
		trees:    []Tree{},
		basepath: basepath,
	}

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, " ", 3)
		if len(parts) != 3 {
			continue
		}

		typ := parts[0]
		hash, err := NewObjectHash(parts[1])
		if err != nil {
			return nil, err
		}

		path := parts[2]

		if typ == "blob" {
			fullPath := filepath.FromSlash(filepath.Join(basepath, path))
			tree.blobs = append(tree.blobs, blobLine{
				Hash: hash,
				File: fullPath,
			})
			continue
		}

		if typ == "tree" {
			// recurse into subtree
			subBasepath := filepath.Join(basepath, path)
			subTree, err := readRecursiveTree(repoPath, hash, subBasepath)
			if err != nil {
				return nil, err
			}
			tree.trees = append(tree.trees, subTree)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return tree, nil
}

func readTreeData(repoPath string, hash ObjectHash) ([]byte, error) {
	data, objType, err := readObject(repoPath, hash)
	if err != nil {
		return nil, err
	}

	if objType != TreeType {
		return nil, fmt.Errorf("object %s is not a tree", hash)
	}

	return data, nil
}
