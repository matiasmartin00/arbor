package object

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/matiasmartin00/arbor/internal/utils"
)

// HashObject takes the data and its type (e.g., "blob", "tree", "commit")
// and returns the SHA-1 hash of the object as a hexadecimal string.
func hashObject(data []byte, objType string) string {
	storeData := createObject(data, objType)
	h := sha1.Sum(storeData)
	return hex.EncodeToString(h[:])
}

func HashBlob(data []byte) string {
	return hashObject(data, "blob")
}

func WriteBlob(repoPath string, data []byte) (string, error) {
	return writeObject(repoPath, data, "blob")
}

func WriteTree(repoPath string, data []byte) (string, error) {
	return writeObject(repoPath, data, "tree")
}

func WriteCommit(repoPath string, data []byte) (string, error) {
	return writeObject(repoPath, data, "commit")
}

func writeObject(repoPath string, data []byte, objType string) (string, error) {
	hash := hashObject(data, objType)
	dir := utils.GetObjectsDir(repoPath)
	objDir := filepath.Join(dir, hash[:2])
	if err := utils.CreateDir(objDir); err != nil {
		return "", err
	}

	file := filepath.Join(objDir, hash[2:])
	if utils.Exists(file) {
		return hash, nil
	}

	content := createObject(data, objType)

	if err := utils.WriteFile(file, content); err != nil {
		return "", err
	}
	return hash, nil
}

func ReadCommit(repoPath, hash string) ([]byte, error) {
	data, objType, err := ReadObject(repoPath, hash)
	if err != nil {
		return nil, err
	}

	if objType != "commit" {
		return nil, fmt.Errorf("object %s is not a commit", hash)
	}

	return data, nil
}

func ReadBlob(repoPath, hash string) ([]byte, error) {
	data, objType, err := ReadObject(repoPath, hash)
	if err != nil {
		return nil, err
	}

	if objType != "blob" {
		return nil, fmt.Errorf("object %s is not a blob", hash)
	}
	return data, nil
}

func ReadTree(repoPath, hash string) ([]byte, error) {
	data, objType, err := ReadObject(repoPath, hash)
	if err != nil {
		return nil, err
	}

	if objType != "tree" {
		return nil, fmt.Errorf("object %s is not a tree", hash)
	}
	return data, nil
}

func ReadObject(repoPath, hash string) ([]byte, string, error) {
	dir := utils.GetObjectsDir(repoPath)
	objDir := filepath.Join(dir, hash[:2])
	file := filepath.Join(objDir, hash[2:])

	content, err := os.ReadFile(file)
	if err != nil {
		return nil, "", err
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
		return nil, "", fmt.Errorf("invalid object format")
	}

	var size int
	var objType string
	if _, err := fmt.Sscanf(string(content[:zero]), "%s %d", &objType, &size); err != nil {
		return nil, "", err
	}

	data := content[zero+1:]
	if len(data) != size {
		return nil, "", fmt.Errorf("invalid object size")
	}

	return data, objType, nil
}

func createObject(data []byte, objType string) []byte {
	header := fmt.Sprintf("%s %d\x00", objType, len(data))
	return append([]byte(header), data...)
}
