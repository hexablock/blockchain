package bcpb

import (
	"bytes"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/hexablock/blockchain/hasher"
)

// Digest is a hash with the hash function used. The format is as follows
// algo ':' hash
type Digest []byte

// NewDigest creates a new digest with the given algorithm identifier and hash
func NewDigest(algo string, hash []byte) Digest {
	p := []byte(algo + ":")
	return Digest(append(p, hash...))
}

// NewZeroDigest returns a zero digest for the given hash function
func NewZeroDigest(h hasher.Hasher) Digest {
	return NewDigest(h.Name(), make([]byte, h.Size()))
}

// Copy does a byte copy of the digest returning a new copy
func (digest Digest) Copy() Digest {
	clone := make([]byte, len(digest))
	copy(clone, digest)
	return Digest(clone)
}

// Equal returns true if both digests are the same
func (digest Digest) Equal(d Digest) bool {
	return bytes.Compare(digest, d) == 0
}

// Algorithm returns the hasing algorithm used
func (digest Digest) Algorithm() string {
	s := bytes.IndexRune(digest, ':')
	return string(digest[:s])
}

// Hash returns the hash bytes of the digest
func (digest Digest) Hash() []byte {
	s := bytes.IndexRune(digest, ':')
	return digest[s+1:]
}

func (digest Digest) String() string {
	s := bytes.IndexRune(digest, ':')
	return string(digest[:s+1]) + hex.EncodeToString(digest[s+1:])
}

// ParseDigest parses a digest string into a digest
func ParseDigest(str string) (Digest, error) {
	i := strings.Index(str, ":")
	if i < 1 {
		return nil, errors.New("invalid digest")
	}

	sh, err := hex.DecodeString(str[i+1:])
	if err != nil {
		return nil, err
	}

	return NewDigest(str[:i], sh), nil
}
