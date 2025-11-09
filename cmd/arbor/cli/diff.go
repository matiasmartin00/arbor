package cli

import (
	"fmt"

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
		RunE: func(c *cobra.Command, args []string) error {

			// if commits present -> diff commits
			if len(args) >= 2 {
				diffResult, err := diff.DiffCommits(repoPath, args[0], args[1], paths)
				if err != nil {
					return err
				}
				printResult(fmt.Sprintf("commit %s -> %s", args[0], args[1]), diffResult)
				return nil
			}

			// staged mode
			if staged {
				diffResults, err := diff.DiffIndexVsHead(repoPath, paths)
				if err != nil {
					return err
				}
				printResult("index vs HEAD", diffResults)
				return nil
			}

			// default: workdir vs index
			diffResult, err := diff.DiffWorktreeVsIndex(repoPath, paths)
			if err != nil {
				return err
			}

			printResult("workdir vs index", diffResult)

			return nil
		},
	}

	cmd.Flags().BoolVarP(&staged, "staged", "s", false, "Show diff between index and HEAD (staged changes)")
	cmd.Flags().StringSliceVarP(&paths, "paths", "p", []string{}, "You can pass paths to limit to specific files")
	return cmd
}

func printResult(difference string, diffResult []diff.DiffResult) {
	for _, dr := range diffResult {
		fmt.Printf("diff -- a/%s b/%s (%s)\n", dr.File, dr.File, difference)
		if dr.AHash != nil && dr.BHash != nil {
			fmt.Printf("index -- %s vs %s\n", dr.AHash, dr.BHash)
		} else if dr.AHash != nil {
			fmt.Printf("index -- %s\n", dr.AHash)
		} else {
			fmt.Printf("index -- %s\n", dr.BHash)
		}

		if dr.Lines == nil {
			fmt.Printf("Binary file\n")
			fmt.Printf("\n\n")
			continue
		}

		for _, ld := range dr.Lines {
			fmt.Printf("%s%s\n", ld.Result, ld.ResultLine)
		}
		fmt.Printf("\n\n")
	}
}
