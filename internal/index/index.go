package index

import (
	"encoding/json"
	"os"

	"github.com/matiasmartin00/arbor/internal/object"
	"github.com/matiasmartin00/arbor/internal/utils"
)

type indexEntry struct {
	Hash     object.ObjectHash
	IsBinary bool
}

type Index map[string]indexEntry

func Load(repoPath string) (Index, error) {
	indexPath := utils.GetIndexPath(repoPath)
	data, err := utils.ReadFile(indexPath)

	if err != nil {
		if os.IsNotExist(err) {
			return Index{}, nil
		}
		return nil, err
	}

	var idx Index
	if err := json.Unmarshal(data, &idx); err != nil {
		return nil, err
	}

	return idx, err
}

func (idx Index) Save(repoPath string) error {
	indexPath := utils.GetIndexPath(repoPath)
	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return err
	}

	return utils.WriteFile(indexPath, data)
}

func (idx Index) AddEntry(path string, hash object.ObjectHash) {
	blob, _ := object.ReadBlob(".", hash)
	idx[path] = indexEntry{
		Hash:     hash,
		IsBinary: utils.IsBinary(blob.Data()),
	}
}

type indexEntryJSON struct {
	Hash     string `json:"hash"`
	IsBinary bool   `json:"is_binary"`
}

func (e indexEntry) MarshalJSON() ([]byte, error) {
	var hs string
	if e.Hash != nil {
		hs = e.Hash.String()
	}

	j := indexEntryJSON{
		Hash:     hs,
		IsBinary: e.IsBinary,
	}

	return json.MarshalIndent(j, "", "  ")
}

func (e *indexEntry) UnmarshalJSON(b []byte) error {
	var j indexEntryJSON
	if err := json.Unmarshal(b, &j); err != nil {
		return err
	}

	if j.Hash == "" {
		e.Hash = nil
	} else {
		concrete, err := object.NewObjectHash(j.Hash)
		if err != nil {
			return err
		}
		e.Hash = concrete
	}

	e.IsBinary = j.IsBinary
	return nil
}
