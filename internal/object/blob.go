package object

import (
	"bufio"
	"fmt"
	"strings"
)

type Blob interface {
	Hash() ObjectHash
	SplitLines() ([]string, error)
	Data() []byte
}

type blob struct {
	hash ObjectHash
	data []byte
}

func NewHashBlob(data []byte) (ObjectHash, error) {
	return hashObject(data, BlobType)
}

func (b *blob) Hash() ObjectHash {
	return b.hash
}

func (b *blob) Data() []byte {
	return b.data
}

func (b *blob) SplitLines() ([]string, error) {
	return SplitLines(b.data)
}

func WriteBlob(repoPath string, data []byte) (ObjectHash, error) {
	return writeObject(repoPath, data, BlobType)
}

func ReadBlob(repoPath string, hash ObjectHash) (Blob, error) {
	data, objType, err := readObject(repoPath, hash)
	if err != nil {
		return nil, err
	}

	if objType != BlobType {
		return nil, fmt.Errorf("object %s is not a blob", hash)
	}
	return &blob{
		hash: hash,
		data: data,
	}, nil
}

func SplitLines(data []byte) ([]string, error) {
	var out []string
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		out = append(out, scanner.Text())
	}

	if scanner.Err() != nil {
		return []string{}, scanner.Err()
	}

	return out, nil
}
