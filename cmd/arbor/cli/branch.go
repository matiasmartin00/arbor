package cli

import (
	"fmt"

	"github.com/matiasmartin00/arbor/internal/branch"
	"github.com/spf13/cobra"
)

func NewBranchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "branch",
		Short: "Branch operations",
	}

	createCmd := &cobra.Command{
		Use:     "create <branch-name>",
		Short:   "Create a new branch",
		Args:    cobra.ExactArgs(1),
		PreRunE: preRunErr,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := branch.CreateBranch(repoPath, args[0]); err != nil {
				return err
			}

			fmt.Printf("Created branch %s\n", args[0])
			return nil
		},
	}

	listCmd := &cobra.Command{
		Use:     "list",
		Short:   "List all branches",
		Args:    cobra.NoArgs,
		PreRunE: preRunErr,
		RunE: func(cmd *cobra.Command, args []string) error {
			branches, err := branch.ListBranches(repoPath)
			if err != nil {
				fmt.Println("Error listing branches:", err)
				return err
			}

			fmt.Println("Branches:")
			for _, b := range branches {
				if b.IsActive {
					fmt.Printf(" * %s\n", b.Name)
					continue
				}
				fmt.Printf("   %s\n", b.Name)
			}

			return nil
		},
	}

	cmd.AddCommand(createCmd, listCmd)
	return cmd
}
