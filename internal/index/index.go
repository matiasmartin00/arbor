package index

import (
	"encoding/json"
	"os"

	"github.com/matiasmartin00/arbor/internal/utils"
)

type Index map[string]string

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

	return idx, nil
}

func (idx Index) Save(repoPath string) error {
	indexPath := utils.GetIndexPath(repoPath)
	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return err
	}

	return utils.WriteFile(indexPath, data)
}
