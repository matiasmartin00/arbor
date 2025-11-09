package log

import (
	"fmt"
	"time"

	"github.com/matiasmartin00/arbor/internal/object"
	"github.com/matiasmartin00/arbor/internal/refs"
)

type LogCommit struct {
	Hash    object.ObjectHash
	Author  string
	Email   string
	Date    time.Time
	Message string
}

type LogResult struct {
	Logs       []LogCommit
	NextCommit object.ObjectHash
}

func Log(repoPath string, commitHashFrom string, limit int) (LogResult, error) {
	hash, err := calculateHash(repoPath, commitHashFrom)
	if err != nil {
		return LogResult{}, err
	}

	if hash == nil {
		println("No commits yet.")
		return LogResult{}, nil
	}

	if limit <= 0 && limit > 100 {
		return LogResult{}, fmt.Errorf("invalid limit should be gte 1 and lte 100")
	}

	logs := make([]LogCommit, 0, limit)
	var nextCommit object.ObjectHash
	count := 0
	for hash != nil {
		commit, err := object.ReadCommit(repoPath, hash)
		if err != nil {
			return LogResult{}, err
		}

		if count == limit {
			nextCommit = hash
			break
		}

		logs = append(logs, LogCommit{
			Hash:    hash,
			Author:  commit.Author(),
			Email:   commit.Email(),
			Date:    commit.Timestamp(),
			Message: commit.Message(),
		})

		hash = commit.ParentHash()
		count++
	}

	return LogResult{
		Logs:       logs,
		NextCommit: nextCommit,
	}, nil
}

func parseInt64(s string) (int64, error) {
	var v int64
	_, err := fmt.Sscanf(s, "%d", &v)
	return v, err
}

func calculateHash(repoPath, commitHash string) (object.ObjectHash, error) {
	if len(commitHash) > 0 {
		return object.NewObjectHash(commitHash)
	}
	return refs.GetRefHash(repoPath)
}
