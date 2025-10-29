package cli

import (
	"fmt"

	"github.com/matiasmartin00/arbor/internal/checkout"
	"github.com/spf13/cobra"
)

func NewCheckoutCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "checkout <commit-hash|branch-name>",
		Short:   "Checkout a commit or branch",
		Args:    cobra.ExactArgs(1),
		PreRunE: preRunErr,
		RunE: func(cmd *cobra.Command, args []string) error {

			if err := checkout.Checkout(repoPath, args[0]); err != nil {
				return err
			}

			fmt.Println("Checked out to", args[0])
			return nil
		},
	}

	return cmd
}
