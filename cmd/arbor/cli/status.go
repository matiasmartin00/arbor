package cli

import (
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
			if err := status.Status(repoPath); err != nil {
				return err
			}

			return nil
		},
	}

	return cmd
}
