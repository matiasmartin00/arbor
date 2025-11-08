package object

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/matiasmartin00/arbor/internal/utils"
)

type ObjectType int

const (
	BlobType ObjectType = iota
	TreeType
	CommitType
)

func (ot ObjectType) String() string {
	types := []string{"blob", "tree", "commit"}
	if int(ot) < 0 || int(ot) >= len(types) {
		return ""
	}
	return types[int(ot)]
}

func parseObjectType(s string) ObjectType {
	switch s {
	case "blob":
		return BlobType
	case "tree":
		return TreeType
	case "commit":
		return CommitType
	default:
		return -1
	}
}

// HashObject takes the data and its type (e.g., "blob", "tree", "commit")
// and returns the SHA-1 hash of the object as a hexadecimal string.
func hashObject(data []byte, objType ObjectType) (ObjectHash, error) {
	storeData := createObject(data, objType)
	h := sha1.Sum(storeData)
	return NewObjectHash(hex.EncodeToString(h[:]))
}

func WriteTree(repoPath string, data []byte) (ObjectHash, error) {
	return writeObject(repoPath, data, TreeType)
}

func writeObject(repoPath string, data []byte, objType ObjectType) (ObjectHash, error) {
	hash, err := hashObject(data, objType)
	if err != nil {
		return nil, err
	}

	dir := utils.GetObjectsDir(repoPath)
	objDir := filepath.Join(dir, hash.Dir())
	if err := utils.CreateDir(objDir); err != nil {
		return nil, err
	}

	file := filepath.Join(objDir, hash.File())
	if utils.Exists(file) {
		return hash, nil
	}

	content := createObject(data, objType)

	if err := utils.WriteFile(file, content); err != nil {
		return nil, err
	}
	return hash, nil
}

func ReadTree(repoPath string, hash ObjectHash) ([]byte, error) {
	data, objType, err := ReadObject(repoPath, hash)
	if err != nil {
		return nil, err
	}

	if objType != TreeType {
		return nil, fmt.Errorf("object %s is not a tree", hash)
	}
	return data, nil
}

func ReadObject(repoPath string, hash ObjectHash) ([]byte, ObjectType, error) {
	dir := utils.GetObjectsDir(repoPath)
	objDir := filepath.Join(dir, hash.Dir())
	file := filepath.Join(objDir, hash.File())

	content, err := os.ReadFile(file)
	if err != nil {
		return nil, -1, err
	}

	// header: "<type> <size>\x00"
	zero := -1
	for i := 0; i < len(content); i++ {
		if content[i] == 0 {
			zero = i
			break
		}
	}

	if zero == -1 {
		return nil, -1, fmt.Errorf("invalid object format")
	}

	var size int
	var objType string
	if _, err := fmt.Sscanf(string(content[:zero]), "%s %d", &objType, &size); err != nil {
		return nil, -1, err
	}

	data := content[zero+1:]
	if len(data) != size {
		return nil, -1, fmt.Errorf("invalid object size")
	}

	return data, parseObjectType(objType), nil
}

func createObject(data []byte, objType ObjectType) []byte {
	header := fmt.Sprintf("%s %d\x00", objType, len(data))
	return append([]byte(header), data...)
}
