package cli

import (
	"fmt"
	"os"

	"github.com/matiasmartin00/arbor/internal/diff"
	"github.com/spf13/cobra"
)

func NewDiffCommand() *cobra.Command {
	var staged bool
	var paths []string
	cmd := &cobra.Command{
		Use:   "diff [<commitA> <commitB>] [--paths paths...]",
		Short: "Show changes between commits, index and working tree",
		Long: `Three primary modes:
						- arbor diff             : working tree vs index (unstaged)
						- arbor diff --staged   : index vs HEAD (staged)
						- arbor diff <A> <B>    : diff between two commits
						You can pass paths with flag --paths to limit to specific files.
					`,
		Args:    cobra.ArbitraryArgs,
		PreRunE: preRunErr,
		Run: func(c *cobra.Command, args []string) {

			// if commits present -> diff commits
			if len(args) >= 2 {
				if err := diff.DiffCommits(repoPath, args[0], args[1], paths); err != nil {
					fmt.Println("Error:", err)
					os.Exit(1)
				}
				return
			}

			// staged mode
			if staged {
				if err := diff.DiffIndexVsHead(repoPath, paths); err != nil {
					fmt.Println("Error:", err)
					os.Exit(1)
				}
				return
			}

			// default: workdir vs index
			if err := diff.DiffWorktreeVsIndex(repoPath, paths); err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().BoolVar(&staged, "staged", false, "Show diff between index and HEAD (staged changes)")
	cmd.Flags().StringSliceVar(&paths, "paths", []string{}, "You can pass paths to limit to specific files")
	return cmd
}
