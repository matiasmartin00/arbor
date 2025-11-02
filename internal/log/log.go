package log

import (
	"fmt"
	"strings"
	"time"

	"github.com/matiasmartin00/arbor/internal/commit"
	"github.com/matiasmartin00/arbor/internal/refs"
)

func Log(repoPath string) error {
	hash, err := refs.GetRefHash(repoPath)
	if err != nil {
		return err
	}

	if len(hash) == 0 {
		println("No commits yet.")
		return nil
	}

	for len(hash) > 0 {
		hmap, msg, err := commit.GetCommitContent(repoPath, hash)
		if err != nil {
			return err
		}

		authorLine, ok := hmap["author"]
		if !ok {
			authorLine = "unknown"
		}
		var ts int64
		var nameEmail string

		if len(authorLine) > 0 {
			// author line format: "name <email> timestamp +0000"
			parts := strings.Fields(authorLine)
			if len(parts) >= 2 {
				// timestamp is the penultimate part numeric token; scan from the end
				for i := len(parts) - 2; i >= 0; i-- {
					if t, err := parseInt64(parts[i]); err == nil {
						ts = t
						nameEmail = strings.Join(parts[:i], " ")
						break
					}
				}
			} else {
				nameEmail = authorLine
			}
		}

		fmt.Printf("commit %s\n", hash)
		if len(nameEmail) > 0 {
			if ts != 0 {
				fmt.Printf("Author: %s\n", nameEmail)
				fmt.Printf("Date:   %s\n\n", time.Unix(ts, 0).UTC().Format(time.RFC1123))
			} else {
				fmt.Printf("Author: %s\n\n", nameEmail)
			}
		}

		if len(msg) > 0 {
			fmt.Printf("    %s\n\n", strings.ReplaceAll(msg, "\n", "\n    "))
		}

		parentHash, ok := hmap["parent"]
		if ok {
			hash = parentHash
		} else {
			hash = ""
		}

	}

	return nil
}

func parseInt64(s string) (int64, error) {
	var v int64
	_, err := fmt.Sscanf(s, "%d", &v)
	return v, err
}
