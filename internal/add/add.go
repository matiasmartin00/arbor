package add

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/matiasmartin00/arbor/internal/index"
	"github.com/matiasmartin00/arbor/internal/object"
)

type AddResult struct {
	Hash object.ObjectHash
	Path string
}

// add paths one or more files to the index
// returns thepath->blobHash of the added files
func Add(repoPath string, inputs []string) ([]AddResult, error) {
	added := make(map[string]object.ObjectHash)
	idx, err := index.Load(repoPath)
	if err != nil {
		return nil, err
	}

	collectFile := func(filePath string) error {
		// ignore .arbor directory
		relToRepo, err := filepath.Rel(repoPath, filePath)
		if err == nil &&
			!strings.HasPrefix(relToRepo, "..") &&
			(relToRepo == ".arbor" || strings.HasPrefix(relToRepo, ".arbor"+string(os.PathSeparator))) {
			return nil
		}

		// just add regular files
		info, err := os.Stat(filePath)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		hash, err := object.WriteBlob(repoPath, filePath)
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(repoPath, filePath)
		if err != nil || strings.HasPrefix(relPath, "..") {
			cwd, _ := os.Getwd()
			relPath, _ = filepath.Rel(cwd, filePath)
		}

		relPath = filepath.ToSlash(relPath) // use slash as separator in the index

		curIdxEntry, ok := idx[relPath]
		// if not exists in index, it is a new file
		if !ok {
			idx.AddEntry(relPath, hash)
			added[relPath] = hash
			return nil
		}

		// if it is the same, then it don't have changes
		if curIdxEntry.Hash.Equals(hash) {
			return nil
		}

		// exists but with changes
		idx.AddEntry(relPath, hash)
		added[relPath] = hash

		return nil
	}

	for _, in := range inputs {
		// handle globs
		if strings.ContainsAny(in, "*?[]") {
			matches, err := filepath.Glob(in)
			if err != nil {
				return nil, err
			}

			for _, match := range matches {
				// recursively add files in directories
				info, err := os.Stat(match)
				if err != nil {
					return nil, err
				}

				if info.IsDir() {
					err := filepath.WalkDir(match, func(path string, d os.DirEntry, err error) error {
						if err != nil {
							return err
						}
						if !d.IsDir() {
							return collectFile(path)
						}
						return collectFile(path)
					})
					if err != nil {
						return nil, err
					}
					continue
				}

				if err := collectFile(match); err != nil {
					return nil, err
				}
			}
			continue
		}

		// if not a glob, check if it exists
		info, err := os.Stat(in)
		if err != nil {
			alt := filepath.Join(repoPath, in)
			info, err = os.Stat(alt)
			if err == nil {
				in = alt
			} else {
				return nil, err
			}
		}

		if info.IsDir() {
			err := filepath.WalkDir(in, func(path string, d os.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if !d.IsDir() {
					return collectFile(path)
				}
				if d.Name() == ".arbor" {
					return filepath.SkipDir
				}
				return nil
			})
			if err != nil {
				return nil, err
			}
			continue
		}

		if err := collectFile(in); err != nil {
			return nil, err
		}
	}

	if err := idx.Save(repoPath); err != nil {
		return nil, err
	}

	return toAddResult(added), nil
}

func toAddResult(added map[string]object.ObjectHash) []AddResult {
	r := make([]AddResult, 0, len(added))
	for p, h := range added {
		r = append(r, AddResult{
			Hash: h,
			Path: p,
		})
	}
	return r
}
