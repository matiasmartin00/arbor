package log

import (
	"fmt"
	"strings"
	"time"

	"github.com/matiasmartin00/arbor/internal/object"
	"github.com/matiasmartin00/arbor/internal/refs"
)

func Log(repoPath string) error {
	hash, err := refs.GetRefHash(repoPath)
	if err != nil {
		return err
	}

	if hash == nil {
		println("No commits yet.")
		return nil
	}

	for hash != nil {
		commit, err := object.ReadCommit(repoPath, hash)
		if err != nil {
			return err
		}

		fmt.Printf("commit %s\n", hash)
		if len(commit.Author()) > 0 {
			fmt.Printf("Author: %s <%s>\n", commit.Author(), commit.Email())
		}

		fmt.Printf("Date:   %s\n\n", commit.Timestamp().Format(time.RFC1123))

		if len(commit.Message()) > 0 {
			fmt.Printf("    %s\n\n", strings.ReplaceAll(commit.Message(), "\n", "\n    "))
		}

		hash = commit.ParentHash()

	}

	return nil
}

func parseInt64(s string) (int64, error) {
	var v int64
	_, err := fmt.Sscanf(s, "%d", &v)
	return v, err
}
