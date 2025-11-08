package object

import (
	"fmt"
	"strings"
)

type objectHash struct {
	val string
}

type ObjectHash interface {
	String() string
	Short(int) string
	Equals(ObjectHash) bool
	NotEquals(ObjectHash) bool
	Dir() string
	File() string
}

func NewObjectHashFromBytes(value []byte) (ObjectHash, error) {
	return NewObjectHash(string(value))
}

func NewObjectHash(value string) (ObjectHash, error) {
	v := strings.TrimSpace(value)
	if len(v) == 0 || notIsHex(v) {
		return nil, fmt.Errorf("invalid hash value")
	}
	return &objectHash{
		val: v,
	}, nil
}

func notIsHex(s string) bool {
	if len(s) == 0 {
		return true
	}
	for _, r := range s {
		if !('0' <= r && r <= '9' || 'a' <= r && r <= 'f') {
			return true
		}
	}
	return false
}

func (h objectHash) String() string {
	return h.val
}

func (h objectHash) Short(n int) string {
	s := h.String()
	if n <= 0 || len(s) <= n {
		return s
	}
	return s[:n]
}

func (h objectHash) Equals(oh ObjectHash) bool {
	if oh == nil {
		return false
	}
	return strings.Compare(h.String(), oh.String()) == 0
}

func (h objectHash) NotEquals(oh ObjectHash) bool {
	return !h.Equals(oh)
}

func (h objectHash) Dir() string {
	s := h.String()
	return s[:2]
}

func (h objectHash) File() string {
	s := h.String()
	return s[2:]
}
