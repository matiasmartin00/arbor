package cli

import (
	"fmt"

	"github.com/matiasmartin00/arbor/internal/repo"
	"github.com/spf13/cobra"
)

func NewInitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new Arbor repository",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Initializing repository at", repoPath)
			return repo.Init(repoPath)
		},
	}

	return cmd
}
