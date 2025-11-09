package cli

import (
	"fmt"

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
			mergeDetail, err := merge.Merge(repoPath, branchName)
			if err != nil {
				return err
			}

			if mergeDetail.IsFastForward() {
				fmt.Printf("%s merged.\n\n", mergeDetail.Type)
				fmt.Printf("Current commit: %s\n", mergeDetail.CommitHash)
				return nil
			}

			// if we have conflicts don't auto commit
			if len(mergeDetail.Conflicts) > 0 {
				fmt.Printf("Merge %s completed with conflicts:\n", mergeDetail.Type)
				for _, c := range mergeDetail.Conflicts {
					fmt.Printf(" -%s\n", c)
				}
				fmt.Printf("\n\n    Resolve conflicts and run `arbor commit ...` to finalize merge\n\n")
				return nil
			}

			// auto commit merge
			fmt.Printf("%s merged.\n\n", mergeDetail.Type)
			fmt.Printf("Merge commit created: %s\n", mergeDetail.CommitHash)
			return nil
		},
	}
}
