package object

import (
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	headerTree      = "tree"
	headerParent    = "parent"
	headerAuthor    = "author"
	headerCommitter = "committer"
)

type Commit interface {
	ParentHash() ObjectHash
	TreeHash() ObjectHash
	Author() string
	Email() string
	Timestamp() time.Time
	Message() string
}

type commit struct {
	hash               ObjectHash
	tree               ObjectHash
	parent             ObjectHash
	author             string
	authorEmail        string
	authorTimestamp    time.Time
	committer          string
	committerEmail     string
	committerTimestamp time.Time
	message            string
	raw                []byte
}

func (c *commit) TreeHash() ObjectHash {
	return c.tree
}

func (c *commit) ParentHash() ObjectHash {
	return c.parent
}

func (c *commit) Author() string {
	if len(c.author) == 0 {
		return "unknown"
	}
	return c.author
}

func (c *commit) Message() string {
	return c.message
}

func (c *commit) Email() string {
	if len(c.authorEmail) == 0 {
		return "unknown"
	}
	return c.authorEmail
}

func (c *commit) Timestamp() time.Time {
	return c.authorTimestamp
}

func ReadCommit(repoPath string, hash ObjectHash) (Commit, error) {
	data, objType, err := readObject(repoPath, hash)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("hash not found %s", hash)
		}
		return nil, err
	}

	if objType != CommitType {
		return nil, fmt.Errorf("object %s is not a commit", hash)
	}

	headers, msg, err := parseCommitContent(data)
	if err != nil {
		return nil, err
	}

	tree, err := NewObjectHash(headers[headerTree])
	if err != nil {
		return nil, fmt.Errorf("invalid commit object (%s) no tree found. err %v", hash, err)
	}

	authorLine := headers[headerAuthor]
	author, authorEmail, authorTimestamp := parseAuthorCommitterLine(authorLine)
	committerLine := headers[headerCommitter]
	committer, committerEmail, committerTimestamp := parseAuthorCommitterLine(committerLine)

	parent, _ := NewObjectHash(headers[headerParent])

	return &commit{
		tree:               tree,
		parent:             parent,
		author:             author,
		authorEmail:        authorEmail,
		authorTimestamp:    authorTimestamp,
		committer:          committer,
		committerEmail:     committerEmail,
		committerTimestamp: committerTimestamp,
		message:            msg,
		raw:                data,
	}, nil
}

func parseAuthorCommitterLine(line string) (string, string, time.Time) {
	// author|commiter line format: "name <email> timestamp +0000"
	if len(line) == 0 {
		return "", "", time.Time{}
	}

	parts := strings.SplitN(line, " <", 2)
	if len(parts) != 2 {
		return "", "", time.Time{}
	}
	name := parts[0]

	rest := strings.SplitN(parts[1], "> ", 2)
	if len(rest) != 2 {
		return "", "", time.Time{}
	}
	email := rest[0]

	timeParts := strings.SplitN(rest[1], " ", 2)
	if len(timeParts) == 0 {
		return name, email, time.Time{}
	}

	epoch, err := parseInt64(timeParts[0])
	if err != nil {
		return name, email, time.Time{}
	}

	timestamp := time.Unix(epoch, 0).UTC()
	return name, email, timestamp
}

func WriteCommit(repoPath string, treeHash ObjectHash, parentHash ObjectHash, message string) (ObjectHash, error) {
	return writeObject(repoPath, buildCommitContent(treeHash, parentHash, message), CommitType)
}

func buildCommitContent(treeHash ObjectHash, parentHash ObjectHash, message string) []byte {
	// author
	user := os.Getenv("USER")
	if len(user) == 0 {
		user = "anonymous"
	}
	email := fmt.Sprintf("%s@localhost", user)
	ts := time.Now().Unix()

	// commit content
	data := fmt.Sprintf("%s %s\n", headerTree, treeHash)
	if parentHash != nil {
		data += fmt.Sprintf("%s %s\n", headerParent, parentHash)
	}
	data += fmt.Sprintf("%s %s <%s> %d +0000\n", headerAuthor, user, email, ts)
	data += fmt.Sprintf("%s %s <%s> %d +0000\n\n", headerCommitter, user, email, ts)
	data += message + "\n"

	return []byte(data)
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

func parseInt64(s string) (int64, error) {
	var v int64
	_, err := fmt.Sscanf(s, "%d", &v)
	return v, err
}
