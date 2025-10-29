package cli

import (
	"github.com/matiasmartin00/arbor/internal/repo"
	"github.com/spf13/cobra"
)

func NewInitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new Arbor repository",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return repo.Init(repoPath)
		},
	}

	return cmd
}
