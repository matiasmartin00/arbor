package cli

import (
	"github.com/matiasmartin00/arbor/internal/merge"
	"github.com/spf13/cobra"
)

func NewMergeCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "merge <branch>",
		Short:   "Merge a branch into the current branch",
		Args:    cobra.ExactArgs(1),
		PreRunE: preRunErr,
		RunE: func(c *cobra.Command, args []string) error {
			branchName := args[0]
			if err := merge.Merge(".", branchName); err != nil {
				return err
			}

			return nil
		},
	}
}
