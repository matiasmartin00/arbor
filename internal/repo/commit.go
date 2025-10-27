package repo

import (
	"fmt"
	"os"
	"time"

	"github.com/matiasmartin00/arbor/internal/object"
)

func Commit(repoPath, message string) (string, error) {
	// write tree
	treeHash, err := writeTree(repoPath)

	if err != nil {
		return "", err
	}

	// get parent
	parentHash, err := getRefHash(repoPath)
	if err != nil {
		return "", err
	}

	// author
	user := os.Getenv("USER")
	if len(user) == 0 {
		user = "anonymous"
	}
	email := fmt.Sprintf("%s@localhost", user)
	ts := time.Now().Unix()

	// commit content
	commit := fmt.Sprintf("tree %s\n", treeHash)
	if parentHash != "" {
		commit += fmt.Sprintf("parent %s\n", parentHash)
	}
	commit += fmt.Sprintf("author %s <%s> %d +0000\n", user, email, ts)
	commit += fmt.Sprintf("committer %s <%s> %d +0000\n\n", user, email, ts)
	commit += message + "\n"

	// write commit object
	commitHash, err := object.WriteCommit(repoPath, []byte(commit))
	if err != nil {
		return "", err
	}

	// update ref
	if err := updateRef(repoPath, commitHash); err != nil {
		return "", err
	}

	return commitHash, nil
}
