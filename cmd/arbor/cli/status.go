package cli

import (
	"fmt"

	"github.com/matiasmartin00/arbor/internal/status"
	"github.com/spf13/cobra"
)

func NewStatusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "status",
		Short:   "Show working tree status",
		Args:    cobra.NoArgs,
		PreRunE: preRunErr,
		RunE: func(cmd *cobra.Command, args []string) error {
			status, err := status.Status(repoPath)
			if err != nil {
				return err
			}

			if len(status.ToBeCommitted) == 0 {
				fmt.Println("No changes to be committed.")
			} else {
				fmt.Println("Changes to be commited: ")
				for _, l := range status.ToBeCommitted {
					fmt.Printf("  %s\n", l)
				}
			}
			fmt.Printf("\n\n")

			if len(status.NotStaged) == 0 {
				fmt.Println("No changes not staged for commit.")
			} else {
				fmt.Println("Changes not staged for commit: ")
				for _, l := range status.NotStaged {
					fmt.Printf("  %s\n", l)
				}
			}
			fmt.Printf("\n\n")

			if len(status.Untracked) == 0 {
				fmt.Println("No untracked files.")
			} else {
				fmt.Println("Untracked files: ")
				for _, u := range status.Untracked {
					fmt.Printf("  %s\n", u)
				}
			}

			return nil
		},
	}

	return cmd
}
