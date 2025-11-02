package commit

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/matiasmartin00/arbor/internal/object"
	"github.com/matiasmartin00/arbor/internal/refs"
	"github.com/matiasmartin00/arbor/internal/tree"
)

const (
	HeaderTree      = "tree"
	HeaderParent    = "parent"
	HeaderAuthor    = "author"
	HeaderCommitter = "committer"
)

func Commit(repoPath, message string) (string, error) {
	// write tree
	treeHash, err := tree.WriteTree(repoPath)

	if err != nil {
		return "", err
	}

	// get parent
	parentHash, err := refs.GetRefHash(repoPath)
	if err != nil {
		return "", err
	}

	// commit content
	content := buildCommitContent(treeHash, parentHash, message)
	// write commit object
	commitHash, err := object.WriteCommit(repoPath, []byte(content))
	if err != nil {
		return "", err
	}

	// update ref
	if err := refs.UpdateRef(repoPath, commitHash); err != nil {
		return "", err
	}

	return commitHash, nil
}

func buildCommitContent(treeHash, parentHash, message string) string {
	// author
	user := os.Getenv("USER")
	if len(user) == 0 {
		user = "anonymous"
	}
	email := fmt.Sprintf("%s@localhost", user)
	ts := time.Now().Unix()

	// commit content
	commit := fmt.Sprintf("%s %s\n", HeaderTree, treeHash)
	if parentHash != "" {
		commit += fmt.Sprintf("%s %s\n", HeaderParent, parentHash)
	}
	commit += fmt.Sprintf("%s %s <%s> %d +0000\n", HeaderAuthor, user, email, ts)
	commit += fmt.Sprintf("%s %s <%s> %d +0000\n\n", HeaderCommitter, user, email, ts)
	commit += message + "\n"

	return commit
}

// resturns headers, message, error
func GetCommitContent(repoPath, commitHash string) (map[string]string, string, error) {
	data, err := object.ReadCommit(repoPath, commitHash)
	if err != nil {
		return nil, "", err
	}

	return parseCommitContent(data)
}

func parseCommitContent(data []byte) (map[string]string, string, error) {
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
