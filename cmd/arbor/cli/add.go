package cli

import (
	"fmt"

	"github.com/matiasmartin00/arbor/internal/add"
	"github.com/spf13/cobra"
)

func NewAddCommand() *cobra.Command {
	var stageDeleted bool
	cmd := &cobra.Command{
		Use:     "add <files...>",
		Short:   "Add files to the staging area",
		Args:    cobra.MinimumNArgs(1),
		PreRunE: preRunErr,
		RunE: func(cmd *cobra.Command, args []string) error {

			added, err := add.Add(repoPath, stageDeleted, args)
			if err != nil {
				return err
			}

			if len(added) == 0 {
				fmt.Printf("Nothing pending!\n")
			}

			for _, ad := range added {
				if ad.IsDeleted {
					fmt.Printf("Removed %s\n", ad.Path)
					continue
				}
				fmt.Printf("Added %s with hash %s\n", ad.Path, ad.Hash)
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&stageDeleted, "deletions", "d", false, "Stage deletions")
	return cmd
}
