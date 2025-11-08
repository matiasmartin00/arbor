package index

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/matiasmartin00/arbor/internal/object"
	"github.com/matiasmartin00/arbor/internal/utils"
)

type Index map[string]object.ObjectHash

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

func (idx Index) MarshalJSON() ([]byte, error) {
	raw := make(map[string]string, len(idx))
	for k, v := range idx {
		if v == nil {
			continue
		}
		raw[k] = v.String()
	}
	return json.Marshal(raw)
}

func (idx *Index) UnmarshalJSON(b []byte) error {
	var raw map[string]string
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}

	res := make(Index, len(raw))
	for k, s := range raw {
		oh, err := object.NewObjectHash(s)
		if err != nil {
			return fmt.Errorf("invalid hash for key %q: %w", k, err)
		}
		res[k] = oh
	}
	*idx = res
	return nil
}
