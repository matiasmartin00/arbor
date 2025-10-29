package repo

import (
	"fmt"
	"strings"
	"time"

	"github.com/matiasmartin00/arbor/internal/object"
)

func parseCommit(data []byte) (map[string]string, string, error) {
	s := string(data)
	parts := strings.SplitN(s, "\n\n", 2)
	if len(parts) != 2 {
		return nil, "", nil
	}

	headers := strings.Split(parts[0], "\n")
	hmap := make(map[string]string)
	for _, h := range headers {
		if len(h) == 0 {
			continue
		}

		kv := strings.SplitN(h, " ", 2)
		if len(kv) != 2 {
			continue
		}

		hmap[kv[0]] = kv[1]
	}

	msg := ""

	if len(parts) == 2 {
		msg = strings.TrimSpace(parts[1])
	}

	return hmap, msg, nil
}

func Log(repoPath string) error {
	hash, err := getRefHash(repoPath)
	if err != nil {
		return err
	}

	if len(hash) == 0 {
		println("No commits yet.")
		return nil
	}

	for len(hash) > 0 {
		data, err := object.ReadCommit(repoPath, hash)
		if err != nil {
			return err
		}

		hmap, msg, err := parseCommit(data)
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
