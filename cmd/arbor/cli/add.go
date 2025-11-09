package cli

import (
	"fmt"

	"github.com/matiasmartin00/arbor/internal/add"
	"github.com/spf13/cobra"
)

func NewAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add <files...>",
		Short:   "Add files to the staging area",
		Args:    cobra.MinimumNArgs(1),
		PreRunE: preRunErr,
		RunE: func(cmd *cobra.Command, args []string) error {

			added, err := add.Add(repoPath, args)
			if err != nil {
				return err
			}

			if len(added) == 0 {
				fmt.Printf("Nothing pending!\n")
			}

			for _, ad := range added {
				fmt.Printf("Added %s with hash %s\n", ad.Path, ad.Hash)
			}

			return nil
		},
	}

	return cmd
}
