package cli

import (
	"fmt"

	"github.com/matiasmartin00/arbor/internal/commit"
	"github.com/spf13/cobra"
)

func NewCommitCommand() *cobra.Command {
	var message string
	cmd := &cobra.Command{
		Use:     "commit -m <message>",
		Short:   "Create a new commit",
		PreRunE: preRunErr,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(message) == 0 {
				return fmt.Errorf("commit message required: arbor commit -m <message>")
			}

			commitHash, err := commit.Commit(repoPath, message)
			if err != nil {
				return err
			}

			fmt.Println("Committed with hash:", commitHash)
			return nil
		},
	}

	cmd.Flags().StringVarP(&message, "message", "m", "", "Commit message")

	return cmd
}
