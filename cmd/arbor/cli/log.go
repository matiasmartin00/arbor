package cli

import (
	"github.com/matiasmartin00/arbor/internal/log"
	"github.com/spf13/cobra"
)

func NewLogCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "log",
		Short:   "Show commit logs",
		Args:    cobra.NoArgs,
		PreRunE: preRunErr,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := log.Log(repoPath); err != nil {
				return err
			}

			return nil
		},
	}

	return cmd
}
