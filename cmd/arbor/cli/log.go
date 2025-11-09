package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/matiasmartin00/arbor/internal/log"
	"github.com/spf13/cobra"
)

func NewLogCommand() *cobra.Command {
	var from string
	var limit int
	cmd := &cobra.Command{
		Use:     "log [--from <commitHash>] [--limit <number>]",
		Short:   "Show commit logs",
		Args:    cobra.NoArgs,
		PreRunE: preRunErr,
		RunE: func(cmd *cobra.Command, args []string) error {
			logResult, err := log.Log(repoPath, from, limit)
			if err != nil {
				return err
			}

			for _, l := range logResult.Logs {
				fmt.Printf("commit %s\n", l.Hash)
				if len(l.Author) > 0 {
					fmt.Printf("Author: %s <%s>\n", l.Author, l.Email)
				}

				fmt.Printf("Date:   %s\n\n", l.Date.Format(time.RFC1123))

				if len(l.Message) > 0 {
					fmt.Printf("    %s\n\n", strings.ReplaceAll(l.Message, "\n", "\n    "))
				}
			}

			if logResult.NextCommit != nil {
				fmt.Printf("Next commit: %s", logResult.NextCommit)
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 5, "You can set a limit to restrict the number of commits")
	cmd.Flags().StringVar(&from, "from", "", "You can pass commitHash to start from the same one")
	return cmd
}
