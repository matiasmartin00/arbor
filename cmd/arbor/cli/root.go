package cli

import (
	"github.com/matiasmartin00/arbor/internal/repo"
	"github.com/spf13/cobra"
)

const version = "0.1.0"
const repoPath = "."

var preRunErr = func(cmd *cobra.Command, args []string) error {
	if err := repo.EnsureRepo(repoPath); err != nil {
		return err
	}
	return nil
}

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "arbor",
		Short: "Arbor is a simple version control system",
	}

	cmd.AddCommand(
		NewInitCommand(),
		NewAddCommand(),
		NewCommitCommand(),
		NewLogCommand(),
		NewCheckoutCommand(),
		NewBranchCommand(),
		NewStatusCommand(),
	)

	return cmd
}
